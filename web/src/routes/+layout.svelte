<script>
  import '../app.css';

  let { children } = $props();
</script>

<nav class="topbar">
  <a href="/" class="logo">Agent Analytics</a>
  <div class="nav-links">
    <a href="/evals" class="nav-link">Evals</a>
  </div>
</nav>

<main class="content">
  {@render children()}
</main>

<style>
  .topbar {
    display: flex;
    align-items: center;
    padding: 0 1.5rem;
    height: 48px;
    background: var(--bg-secondary);
    border-bottom: 1px solid var(--border);
  }

  .logo {
    font-size: 0.875rem;
    font-weight: 600;
    letter-spacing: 0.05em;
    text-transform: uppercase;
    color: var(--accent);
  }

  .nav-links {
    display: flex;
    gap: 1rem;
    margin-left: 2rem;
  }

  .nav-link {
    font-size: 0.8125rem;
    color: var(--text-muted);
  }

  .nav-link:hover {
    color: var(--accent);
  }

  .content {
    max-width: 1200px;
    margin: 0 auto;
    padding: 2rem 1.5rem;
  }
</style>
