package look

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type execCall struct {
	query  string
	params map[string]string
}

type mockClickHouse struct {
	execCalls []execCall
}

func (m *mockClickHouse) Exec(ctx context.Context, query string, params map[string]string) error {
	m.execCalls = append(m.execCalls, execCall{query: query, params: params})
	return nil
}

func TestIngestHandler_MessageEvent(t *testing.T) {
	ch := &mockClickHouse{}
	handler := &ingestHandler{db: ch}

	success := true
	events := []Event{
		{
			Type:             "message",
			Timestamp:        time.Date(2026, 3, 4, 14, 0, 0, 0, time.UTC),
			ConversationID:   "conv-1",
			AgentSlug:        "helper",
			UserID:           "user1",
			Role:             "assistant",
			PromptTokens:     1200,
			CompletionTokens: 450,
			DurationMs:       3200,
		},
		{
			Type:           "tool_invocation",
			Timestamp:      time.Date(2026, 3, 4, 14, 0, 1, 0, time.UTC),
			ConversationID: "conv-1",
			AgentSlug:      "helper",
			UserID:         "user1",
			ToolName:       "web_search",
			Success:        &success,
			DurationMs:     850,
		},
	}

	body, _ := json.Marshal(events)
	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d: %s", w.Code, w.Body.String())
	}

	if len(ch.execCalls) != 2 {
		t.Fatalf("expected 2 exec calls, got %d", len(ch.execCalls))
	}

	if !strings.Contains(ch.execCalls[0].query, "events_message") {
		t.Errorf("expected insert into events_message, got: %s", ch.execCalls[0].query)
	}
	if ch.execCalls[0].params["agent_slug"] != "helper" {
		t.Errorf("expected agent_slug=helper, got: %s", ch.execCalls[0].params["agent_slug"])
	}
	if ch.execCalls[0].params["conversation_id"] != "conv-1" {
		t.Errorf("expected conversation_id=conv-1, got: %s", ch.execCalls[0].params["conversation_id"])
	}

	if !strings.Contains(ch.execCalls[1].query, "events_tool_invocation") {
		t.Errorf("expected insert into events_tool_invocation, got: %s", ch.execCalls[1].query)
	}
	if ch.execCalls[1].params["tool_name"] != "web_search" {
		t.Errorf("expected tool_name=web_search, got: %s", ch.execCalls[1].params["tool_name"])
	}
}

func TestIngestHandler_AllEventTypes(t *testing.T) {
	ch := &mockClickHouse{}
	handler := &ingestHandler{db: ch}

	success := true
	events := []Event{
		{Type: "message", Timestamp: time.Now(), AgentSlug: "bot", Role: "user"},
		{Type: "tool_invocation", Timestamp: time.Now(), AgentSlug: "bot", ToolName: "search", Success: &success},
		{Type: "conversation_started", Timestamp: time.Now(), AgentSlug: "bot", EventName: "started"},
		{Type: "memory_extracted", Timestamp: time.Now(), AgentSlug: "bot", MemoryType: "episodic", Importance: 0.8},
		{Type: "relationship_snapshot", Timestamp: time.Now(), AgentSlug: "bot", Trust: 0.7},
		{Type: "consolidation_completed", Timestamp: time.Now(), AgentSlug: "bot", PatternsFound: 3},
	}

	body, _ := json.Marshal(events)
	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", w.Code)
	}

	if len(ch.execCalls) != 6 {
		t.Fatalf("expected 6 exec calls, got %d", len(ch.execCalls))
	}

	expectedTables := []string{
		"events_message",
		"events_tool_invocation",
		"events_conversation",
		"events_memory",
		"events_relationship",
		"events_consolidation",
	}
	for i, table := range expectedTables {
		if !strings.Contains(ch.execCalls[i].query, table) {
			t.Errorf("call %d: expected %s, got: %s", i, table, ch.execCalls[i].query)
		}
		if ch.execCalls[i].params == nil {
			t.Errorf("call %d: expected params to be non-nil", i)
		}
	}
}

func TestIngestHandler_EmptyBatch(t *testing.T) {
	ch := &mockClickHouse{}
	handler := &ingestHandler{db: ch}

	body, _ := json.Marshal([]Event{})
	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", w.Code)
	}
	if len(ch.execCalls) != 0 {
		t.Errorf("expected no exec calls for empty batch")
	}
}

func TestIngestHandler_InvalidBody(t *testing.T) {
	ch := &mockClickHouse{}
	handler := &ingestHandler{db: ch}

	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestIngestHandler_RunEvents(t *testing.T) {
	ch := &mockClickHouse{}
	handler := &ingestHandler{db: ch}

	events := []Event{
		{
			Type:        "run_started",
			Timestamp:   time.Date(2026, 3, 4, 15, 30, 0, 0, time.UTC),
			RunID:       "run-20260304-153000",
			AgentSlug:   "xagent",
			UserID:      "user1",
			Description: "smoke test",
		},
		{
			Type:        "run_completed",
			Timestamp:   time.Date(2026, 3, 4, 15, 35, 0, 0, time.UTC),
			RunID:       "run-20260304-153000",
			AgentSlug:   "xagent",
			UserID:      "user1",
			Description: "smoke test",
		},
	}

	body, _ := json.Marshal(events)
	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", w.Code)
	}

	if len(ch.execCalls) != 2 {
		t.Fatalf("expected 2 exec calls, got %d", len(ch.execCalls))
	}

	for i, call := range ch.execCalls {
		if !strings.Contains(call.query, "events_run") {
			t.Errorf("call %d: expected events_run, got: %s", i, call.query)
		}
		if call.params["run_id"] != "run-20260304-153000" {
			t.Errorf("call %d: expected run_id=run-20260304-153000, got: %s", i, call.params["run_id"])
		}
	}

	if ch.execCalls[0].params["event"] != "started" {
		t.Errorf("expected 'started' event, got: %s", ch.execCalls[0].params["event"])
	}
	if ch.execCalls[1].params["event"] != "completed" {
		t.Errorf("expected 'completed' event, got: %s", ch.execCalls[1].params["event"])
	}
}

func TestIngestHandler_RunID_OnRegularEvents(t *testing.T) {
	ch := &mockClickHouse{}
	handler := &ingestHandler{db: ch}

	events := []Event{
		{
			Type:           "message",
			Timestamp:      time.Now(),
			AgentSlug:      "xagent",
			UserID:         "user1",
			ConversationID: "conv-1",
			Role:           "user",
			RunID:          "run-20260304-153000",
		},
	}

	body, _ := json.Marshal(events)
	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", w.Code)
	}

	if len(ch.execCalls) != 1 {
		t.Fatalf("expected 1 exec call, got %d", len(ch.execCalls))
	}

	if !strings.Contains(ch.execCalls[0].query, "run_id") {
		t.Errorf("expected run_id column in insert, got: %s", ch.execCalls[0].query)
	}
	if ch.execCalls[0].params["run_id"] != "run-20260304-153000" {
		t.Errorf("expected run_id param, got: %s", ch.execCalls[0].params["run_id"])
	}
}

func TestIngestHandler_EvalResultEvent(t *testing.T) {
	ch := &mockClickHouse{}
	handler := &ingestHandler{db: ch}

	events := []Event{
		{
			Type:       "eval_result",
			Timestamp:  time.Date(2026, 3, 4, 17, 0, 0, 0, time.UTC),
			RunID:      "run-20260304-170000",
			AgentSlug:  "xagent",
			UserID:     "user1",
			Scenario:   "greeting",
			Response:   "Hello there!",
			DurationMs: 2847,
			ToolsUsed:  json.RawMessage(`["check_weather"]`),
			Passed:     1,
			Failed:     2,
			Total:      3,
			Criteria:   json.RawMessage(`[{"criterion":"min_length","pass":true,"score":1,"reason":"ok"}]`),
		},
	}

	body, _ := json.Marshal(events)
	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", w.Code)
	}

	if len(ch.execCalls) != 1 {
		t.Fatalf("expected 1 exec call, got %d", len(ch.execCalls))
	}

	call := ch.execCalls[0]
	if !strings.Contains(call.query, "events_eval") {
		t.Errorf("expected events_eval table, got: %s", call.query)
	}
	if call.params["scenario"] != "greeting" {
		t.Errorf("expected scenario=greeting, got: %s", call.params["scenario"])
	}
	if call.params["run_id"] != "run-20260304-170000" {
		t.Errorf("expected run_id in params, got: %s", call.params["run_id"])
	}
}

func TestIngestHandler_ConversationEndedEvent(t *testing.T) {
	ch := &mockClickHouse{}
	handler := &ingestHandler{db: ch}

	events := []Event{
		{
			Type:           "conversation_ended",
			Timestamp:      time.Date(2026, 3, 4, 14, 30, 0, 0, time.UTC),
			ConversationID: "conv-1",
			AgentSlug:      "helper",
			UserID:         "user1",
		},
	}

	body, _ := json.Marshal(events)
	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d: %s", w.Code, w.Body.String())
	}

	if len(ch.execCalls) != 1 {
		t.Fatalf("expected 1 exec call, got %d", len(ch.execCalls))
	}

	call := ch.execCalls[0]
	if !strings.Contains(call.query, "events_conversation") {
		t.Errorf("expected events_conversation table, got: %s", call.query)
	}
	if call.params["event"] != "conversation_ended" {
		t.Errorf("expected event=conversation_ended, got: %s", call.params["event"])
	}
}

func TestIngestHandler_ToolInvocationNilSuccess(t *testing.T) {
	ch := &mockClickHouse{}
	handler := &ingestHandler{db: ch}

	events := []Event{
		{
			Type:           "tool_invocation",
			Timestamp:      time.Date(2026, 3, 4, 14, 0, 1, 0, time.UTC),
			ConversationID: "conv-1",
			AgentSlug:      "helper",
			UserID:         "user1",
			ToolName:       "web_search",
			Success:        nil,
			DurationMs:     500,
		},
	}

	body, _ := json.Marshal(events)
	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d: %s", w.Code, w.Body.String())
	}

	if len(ch.execCalls) != 1 {
		t.Fatalf("expected 1 exec call, got %d", len(ch.execCalls))
	}

	call := ch.execCalls[0]
	if !strings.Contains(call.query, "events_tool_invocation") {
		t.Errorf("expected events_tool_invocation table, got: %s", call.query)
	}
	if call.params["success"] != "false" {
		t.Errorf("expected success=false for nil Success, got: %s", call.params["success"])
	}
}

func TestIngestHandler_UnknownEventType(t *testing.T) {
	ch := &mockClickHouse{}
	handler := &ingestHandler{db: ch}

	events := []Event{{Type: "unknown_type", Timestamp: time.Now()}}
	body, _ := json.Marshal(events)
	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Should still return 202 — skip unknown types
	if w.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", w.Code)
	}
	if len(ch.execCalls) != 0 {
		t.Errorf("expected no exec calls for unknown event type")
	}
}

func TestIngestHandler_SQLInjectionRegression(t *testing.T) {
	ch := &mockClickHouse{}
	handler := &ingestHandler{db: ch}

	maliciousSlug := "' OR 1=1; --"
	events := []Event{
		{
			Type:           "message",
			Timestamp:      time.Now(),
			AgentSlug:      maliciousSlug,
			UserID:         "user1",
			ConversationID: "conv-1",
			Role:           "user",
		},
	}

	body, _ := json.Marshal(events)
	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", w.Code)
	}

	if len(ch.execCalls) != 1 {
		t.Fatalf("expected 1 exec call, got %d", len(ch.execCalls))
	}

	call := ch.execCalls[0]

	// The query must use placeholders, NOT interpolated values
	if strings.Contains(call.query, maliciousSlug) {
		t.Errorf("SQL injection: malicious input found in query string: %s", call.query)
	}
	if !strings.Contains(call.query, "{agent_slug:String}") {
		t.Errorf("expected parameterized placeholder in query, got: %s", call.query)
	}

	// The malicious value should be in the params map (passed safely via URL params)
	if call.params["agent_slug"] != maliciousSlug {
		t.Errorf("expected malicious value in params map, got: %s", call.params["agent_slug"])
	}
}
