<script>
  import { fetchApi } from '$lib/api.js';

  let evals = $state([]);
  let loading = $state(true);
  let error = $state(null);

  $effect(() => {
    loadEvals();
  });

  async function loadEvals() {
    loading = true;
    error = null;
    try {
      evals = await fetchApi('/api/evals');
    } catch (e) {
      // API not available yet — show empty state, not error
      evals = [];
    } finally {
      loading = false;
    }
  }

  function passRate(ev) {
    if (!ev.total) return 0;
    return Math.round((ev.passed / ev.total) * 100);
  }

  function passColor(rate) {
    if (rate >= 80) return 'var(--chart-4)';
    if (rate >= 50) return 'var(--chart-6)';
    return 'var(--chart-5)';
  }
</script>

<div class="page">
  <header>
    <a href="/" class="back">&larr; home</a>
    <h1>Eval Runs</h1>
  </header>

  {#if loading}
    <p class="status">Loading...</p>
  {:else if error}
    <p class="status error">{error}</p>
  {:else if evals.length === 0}
    <div class="empty-state">
      <p class="empty-title">No eval runs yet</p>
      <p class="empty-hint">Run <code>lamina eval plans/smoke.yaml</code> to generate results</p>
    </div>
  {:else}
    <div class="eval-list">
      {#each evals as ev}
        <a href="/evals/{ev.run_id}" class="eval-card">
          <div class="eval-header">
            <span class="eval-plan">{ev.plan || 'unknown'}</span>
            <span class="eval-time">{new Date(ev.timestamp).toLocaleString()}</span>
          </div>
          <div class="eval-id">{ev.run_id}</div>
          <div class="eval-stats">
            <div class="eval-stat">
              <span class="stat-label">Scenarios</span>
              <span class="stat-value">{ev.scenarios}</span>
            </div>
            <div class="eval-stat">
              <span class="stat-label">Passed</span>
              <span class="stat-value pass">{ev.passed}</span>
            </div>
            <div class="eval-stat">
              <span class="stat-label">Failed</span>
              <span class="stat-value fail">{ev.failed}</span>
            </div>
            <div class="eval-stat">
              <span class="stat-label">Pass Rate</span>
              <span class="stat-value" style="color: {passColor(passRate(ev))}">{passRate(ev)}%</span>
            </div>
          </div>
          <div class="progress-bar">
            <div class="progress-fill" style="width: {passRate(ev)}%; background: {passColor(passRate(ev))}"></div>
          </div>
        </a>
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

  .empty-state {
    text-align: center;
    padding: 4rem 2rem;
  }

  .empty-title {
    font-size: 1.125rem;
    color: var(--text-secondary);
    margin-bottom: 0.5rem;
  }

  .empty-hint {
    font-size: 0.8125rem;
    color: var(--text-muted);
  }

  .empty-hint code {
    font-family: var(--font-mono);
    font-size: 0.75rem;
    color: var(--accent);
    background: var(--bg-tertiary);
    padding: 0.125rem 0.375rem;
    border-radius: 3px;
  }

  .eval-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .eval-card {
    display: block;
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: 6px;
    padding: 1rem;
    text-decoration: none;
    transition: border-color 0.15s;
  }

  .eval-card:hover {
    border-color: var(--accent);
  }

  .eval-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.25rem;
  }

  .eval-plan {
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--accent);
  }

  .eval-time {
    font-size: 0.7rem;
    color: var(--text-muted);
  }

  .eval-id {
    font-family: var(--font-mono);
    font-size: 0.75rem;
    color: var(--text-muted);
    margin-bottom: 0.75rem;
  }

  .eval-stats {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 0.5rem;
    margin-bottom: 0.75rem;
  }

  .eval-stat {
    display: flex;
    flex-direction: column;
    gap: 0.125rem;
  }

  .stat-label {
    font-size: 0.625rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-muted);
  }

  .stat-value {
    font-size: 1.25rem;
    font-weight: 600;
    font-family: var(--font-mono);
    color: var(--text-primary);
  }

  .stat-value.pass { color: var(--chart-4); }
  .stat-value.fail { color: var(--chart-5); }

  .progress-bar {
    height: 4px;
    background: var(--bg-tertiary);
    border-radius: 2px;
    overflow: hidden;
  }

  .progress-fill {
    height: 100%;
    border-radius: 2px;
    transition: width 0.3s;
  }
</style>
