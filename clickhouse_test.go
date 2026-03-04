package anal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClickHouse_Exec(t *testing.T) {
	var receivedQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		receivedQuery = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ch := NewClickHouse(server.URL)
	err := ch.Exec(context.Background(), "CREATE TABLE test (id UInt32) ENGINE = MergeTree() ORDER BY id", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(receivedQuery, "CREATE TABLE test") {
		t.Errorf("expected query to be sent, got: %s", receivedQuery)
	}
}

func TestClickHouse_Exec_WithParams(t *testing.T) {
	var receivedURL string
	var receivedQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedURL = r.URL.String()
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		receivedQuery = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ch := NewClickHouse(server.URL)
	params := map[string]string{
		"slug":   "helper",
		"run_id": "run-123",
	}
	err := ch.Exec(context.Background(), "INSERT INTO test VALUES ({slug:String}, {run_id:String})", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(receivedURL, "param_slug=helper") {
		t.Errorf("expected param_slug in URL, got: %s", receivedURL)
	}
	if !strings.Contains(receivedURL, "param_run_id=run-123") {
		t.Errorf("expected param_run_id in URL, got: %s", receivedURL)
	}
	if !strings.Contains(receivedQuery, "{slug:String}") {
		t.Errorf("expected placeholder in query body, got: %s", receivedQuery)
	}
}

func TestClickHouse_Exec_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("DB error"))
	}))
	defer server.Close()

	ch := NewClickHouse(server.URL)
	err := ch.Exec(context.Background(), "BAD QUERY", nil)
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	if !strings.Contains(err.Error(), "DB error") {
		t.Errorf("expected error body, got: %v", err)
	}
}

func TestClickHouse_Query(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":42}` + "\n" + `{"count":7}` + "\n"))
	}))
	defer server.Close()

	ch := NewClickHouse(server.URL)
	body, err := ch.Query(context.Background(), "SELECT count() as count FROM test FORMAT JSONEachRow", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(string(body), `"count":42`) {
		t.Errorf("expected query results, got: %s", string(body))
	}
}

func TestClickHouse_Query_WithParams(t *testing.T) {
	var receivedURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedURL = r.URL.String()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1}` + "\n"))
	}))
	defer server.Close()

	ch := NewClickHouse(server.URL)
	params := map[string]string{"slug": "helper"}
	_, err := ch.Query(context.Background(), "SELECT count() FROM test WHERE slug = {slug:String} FORMAT JSONEachRow", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(receivedURL, "param_slug=helper") {
		t.Errorf("expected param_slug in URL, got: %s", receivedURL)
	}
}

func TestClickHouse_InitSchema(t *testing.T) {
	var queries []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		queries = append(queries, string(body))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ch := NewClickHouse(server.URL)
	err := ch.InitSchema(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should create all event tables
	tableNames := []string{"events_message", "events_tool_invocation", "events_conversation", "events_memory", "events_relationship", "events_consolidation"}
	for _, name := range tableNames {
		found := false
		for _, q := range queries {
			if strings.Contains(q, name) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected schema to create table %s", name)
		}
	}
}
