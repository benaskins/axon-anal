<script>
  // Agent list will be built in Step 11.
  // For now, a simple landing with a link to navigate to an agent.
  let slug = $state('');
</script>

<div class="home">
  <h1>Agent Analytics</h1>
  <p class="subtitle">Per-agent observability dashboard</p>

  <div class="nav-form">
    <label for="slug">Agent slug</label>
    <div class="input-row">
      <input
        id="slug"
        type="text"
        bind:value={slug}
        placeholder="e.g. helper"
      />
      <a
        href="/agents/{slug}"
        class="btn"
        class:disabled={!slug.trim()}
      >
        View Dashboard
      </a>
    </div>
  </div>
</div>

<style>
  .home {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding-top: 4rem;
    gap: 0.5rem;
  }

  h1 {
    font-size: 1.5rem;
    font-weight: 600;
    color: var(--accent);
  }

  .subtitle {
    color: var(--text-muted);
    font-size: 0.875rem;
  }

  .nav-form {
    margin-top: 2rem;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  label {
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-muted);
  }

  .input-row {
    display: flex;
    gap: 0.5rem;
  }

  input {
    padding: 0.5rem 0.75rem;
    background: var(--bg-tertiary);
    border: 1px solid var(--border);
    border-radius: 4px;
    color: var(--text-primary);
    font-size: 0.875rem;
    outline: none;
  }

  input:focus {
    border-color: var(--border-focus);
  }

  .btn {
    padding: 0.5rem 1rem;
    background: var(--accent);
    color: var(--bg-primary);
    border-radius: 4px;
    font-size: 0.875rem;
    font-weight: 500;
    text-decoration: none;
  }

  .btn:hover {
    background: var(--accent-hover);
    color: var(--bg-primary);
  }

  .btn.disabled {
    opacity: 0.4;
    pointer-events: none;
  }
</style>
