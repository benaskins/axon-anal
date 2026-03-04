<script>
  import { page } from '$app/stores';
  import { fetchApi } from '$lib/api.js';

  let runId = $derived($page.params.runId);
  let evalRun = $state(null);
  let loading = $state(true);
  let error = $state(null);
  let openScenarios = $state({});

  $effect(() => {
    loadEval(runId);
  });

  async function loadEval(id) {
    loading = true;
    error = null;
    try {
      const rows = await fetchApi(`/api/evals/${id}`);
      evalRun = rows[0] || null;
    } catch (e) {
      // API not available yet
      error = e.message;
    } finally {
      loading = false;
    }
  }

  function toggle(index) {
    openScenarios = { ...openScenarios, [index]: !openScenarios[index] };
  }

  function scenarioBadge(scenario) {
    const passed = scenario.results.filter(r => r.pass).length;
    const total = scenario.results.length;
    if (passed === total) return { text: 'All passed', cls: 'badge-pass' };
    if (passed === 0) return { text: 'All failed', cls: 'badge-fail' };
    return { text: `${passed}/${total}`, cls: 'badge-mixed' };
  }

  function overallStats(run) {
    if (!run?.scenarios) return { scenarios: 0, passed: 0, failed: 0, rate: 0 };
    let passed = 0, failed = 0;
    for (const s of run.scenarios) {
      for (const r of s.results) {
        if (r.pass) passed++; else failed++;
      }
    }
    const total = passed + failed;
    return {
      scenarios: run.scenarios.length,
      passed,
      failed,
      rate: total ? Math.round((passed / total) * 100) : 0
    };
  }

  function rateColor(rate) {
    if (rate >= 80) return 'var(--chart-4)';
    if (rate >= 50) return 'var(--chart-6)';
    return 'var(--chart-5)';
  }

  function fmtDuration(ms) {
    if (ms == null) return '—';
    if (ms < 1000) return `${Math.round(ms)}ms`;
    return `${(ms / 1000).toFixed(1)}s`;
  }
</script>

<div class="page">
  <header>
    <a href="/evals" class="back">&larr; evals</a>
    <h1>Eval Run</h1>
  </header>

  {#if loading}
    <p class="status">Loading...</p>
  {:else if error}
    <p class="status error">{error}</p>
  {:else if !evalRun}
    <p class="status">Run not found</p>
  {:else}
    {@const stats = overallStats(evalRun)}

    <div class="run-meta">
      <span class="run-id">{evalRun.run_id}</span>
      {#if evalRun.plan}
        <span class="run-plan">{evalRun.plan}</span>
      {/if}
      {#if evalRun.timestamp}
        <span class="run-time">{new Date(evalRun.timestamp).toLocaleString()}</span>
      {/if}
    </div>

    <div class="stats-grid">
      <div class="stat-card">
        <span class="stat-label">Scenarios</span>
        <span class="stat-value accent">{stats.scenarios}</span>
      </div>
      <div class="stat-card">
        <span class="stat-label">Passed</span>
        <span class="stat-value pass">{stats.passed}</span>
      </div>
      <div class="stat-card">
        <span class="stat-label">Failed</span>
        <span class="stat-value fail">{stats.failed}</span>
      </div>
      <div class="stat-card">
        <span class="stat-label">Pass Rate</span>
        <span class="stat-value" style="color: {rateColor(stats.rate)}">{stats.rate}%</span>
        <div class="progress-bar">
          <div class="progress-fill" style="width: {stats.rate}%; background: {rateColor(stats.rate)}"></div>
        </div>
      </div>
    </div>

    <div class="scenarios">
      {#each evalRun.scenarios as scenario, i}
        {@const badge = scenarioBadge(scenario)}
        <div class="scenario">
          <button class="scenario-header" onclick={() => toggle(i)}>
            <span class="scenario-name">{scenario.scenario}</span>
            <div class="scenario-header-right">
              {#if scenario.duration_ms}
                <span class="scenario-duration">{fmtDuration(scenario.duration_ms)}</span>
              {/if}
              <span class="badge {badge.cls}">{badge.text}</span>
              <span class="chevron" class:open={openScenarios[i]}>&#9654;</span>
            </div>
          </button>

          {#if openScenarios[i]}
            <div class="scenario-body">
              <div class="criteria-list">
                {#each scenario.results as criterion}
                  <div class="criterion-row">
                    <span class="criterion-icon" class:pass={criterion.pass} class:fail={!criterion.pass}>
                      {criterion.pass ? '✓' : '✗'}
                    </span>
                    <span class="criterion-type">{criterion.criterion}</span>
                    <span class="criterion-reason">{criterion.reason || ''}</span>
                    {#if criterion.score != null && criterion.score !== 1 && criterion.score !== 0}
                      <span class="criterion-score">{(criterion.score * 100).toFixed(0)}%</span>
                    {/if}
                  </div>
                {/each}
              </div>

              {#if scenario.response}
                <div class="response-section">
                  <div class="response-label">Agent Response</div>
                  <div class="response-text">{scenario.response}</div>
                </div>
              {/if}

              {#if scenario.tools_used?.length}
                <div class="tools-section">
                  <span class="response-label">Tools Used</span>
                  <div class="tool-tags">
                    {#each scenario.tools_used as tool}
                      <span class="tool-tag">{tool}</span>
                    {/each}
                  </div>
                </div>
              {/if}
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .page {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  header {
    display: flex;
    align-items: baseline;
    gap: 1rem;
  }

  .back {
    font-size: 0.75rem;
    color: var(--text-muted);
  }

  h1 {
    font-size: 1.5rem;
    font-weight: 600;
    color: var(--accent);
  }

  .status {
    color: var(--text-muted);
    text-align: center;
    padding: 2rem;
  }

  .status.error { color: var(--chart-5); }

  .run-meta {
    display: flex;
    align-items: center;
    gap: 1rem;
    flex-wrap: wrap;
  }

  .run-id {
    font-family: var(--font-mono);
    font-size: 0.8125rem;
    color: var(--text-secondary);
  }

  .run-plan {
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--accent);
    background: var(--accent-subtle);
    padding: 0.125rem 0.5rem;
    border-radius: 3px;
  }

  .run-time {
    font-size: 0.75rem;
    color: var(--text-muted);
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 0.75rem;
  }

  .stat-card {
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: 6px;
    padding: 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .stat-label {
    font-size: 0.625rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-muted);
  }

  .stat-value {
    font-size: 1.75rem;
    font-weight: 700;
    font-family: var(--font-mono);
    color: var(--text-primary);
  }

  .stat-value.accent { color: var(--accent); }
  .stat-value.pass { color: var(--chart-4); }
  .stat-value.fail { color: var(--chart-5); }

  .progress-bar {
    height: 4px;
    background: var(--bg-tertiary);
    border-radius: 2px;
    overflow: hidden;
    margin-top: 0.5rem;
  }

  .progress-fill {
    height: 100%;
    border-radius: 2px;
  }

  .scenarios {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .scenario {
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: 6px;
    overflow: hidden;
  }

  .scenario-header {
    width: 100%;
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.875rem 1rem;
    background: none;
    border: none;
    color: var(--text-primary);
    font: inherit;
    cursor: pointer;
    text-align: left;
  }

  .scenario-header:hover {
    background: var(--bg-tertiary);
  }

  .scenario-name {
    font-size: 0.9375rem;
    font-weight: 600;
  }

  .scenario-header-right {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .scenario-duration {
    font-family: var(--font-mono);
    font-size: 0.75rem;
    color: var(--text-muted);
  }

  .badge {
    font-size: 0.6875rem;
    font-weight: 600;
    padding: 0.1875rem 0.5rem;
    border-radius: 10px;
  }

  .badge-pass {
    background: rgba(97, 201, 138, 0.15);
    color: var(--chart-4);
  }

  .badge-fail {
    background: rgba(201, 97, 97, 0.15);
    color: var(--chart-5);
  }

  .badge-mixed {
    background: rgba(201, 194, 97, 0.15);
    color: var(--chart-6);
  }

  .chevron {
    font-size: 0.625rem;
    color: var(--text-muted);
    transition: transform 0.15s;
  }

  .chevron.open {
    transform: rotate(90deg);
  }

  .scenario-body {
    padding: 0 1rem 1rem;
    border-top: 1px solid var(--border);
  }

  .criteria-list {
    padding-top: 0.75rem;
  }

  .criterion-row {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.375rem 0;
    border-bottom: 1px solid var(--bg-tertiary);
  }

  .criterion-row:last-child {
    border-bottom: none;
  }

  .criterion-icon {
    width: 1.25rem;
    text-align: center;
    font-size: 0.8125rem;
    font-weight: 600;
  }

  .criterion-icon.pass { color: var(--chart-4); }
  .criterion-icon.fail { color: var(--chart-5); }

  .criterion-type {
    font-family: var(--font-mono);
    font-size: 0.6875rem;
    color: var(--text-muted);
    background: var(--bg-tertiary);
    padding: 0.125rem 0.5rem;
    border-radius: 3px;
    min-width: 7rem;
    text-align: center;
  }

  .criterion-reason {
    font-size: 0.8125rem;
    color: var(--text-secondary);
    flex: 1;
  }

  .criterion-score {
    font-family: var(--font-mono);
    font-size: 0.75rem;
    color: var(--accent);
  }

  .response-section {
    margin-top: 0.75rem;
  }

  .response-label {
    font-size: 0.625rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-muted);
    margin-bottom: 0.375rem;
  }

  .response-text {
    font-size: 0.8125rem;
    color: var(--text-secondary);
    background: var(--bg-primary);
    border: 1px solid var(--border);
    border-radius: 4px;
    padding: 0.75rem;
    line-height: 1.6;
    max-height: 10rem;
    overflow-y: auto;
  }

  .tools-section {
    margin-top: 0.75rem;
  }

  .tool-tags {
    display: flex;
    gap: 0.375rem;
    flex-wrap: wrap;
    margin-top: 0.25rem;
  }

  .tool-tag {
    font-family: var(--font-mono);
    font-size: 0.6875rem;
    color: var(--accent);
    background: var(--accent-subtle);
    padding: 0.125rem 0.5rem;
    border-radius: 3px;
  }
</style>
