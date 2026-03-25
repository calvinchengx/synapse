<script>
  let data = $state(null);
  let loading = $state(true);
  let error = $state(null);

  async function runDoctor() {
    loading = true;
    error = null;
    try {
      const r = await fetch('/api/doctor');
      data = await r.json();
    } catch(e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  $effect(() => { runDoctor(); });
</script>

<div class="diagnostics-view">
  <div class="header-row">
    <h2>Diagnostics</h2>
    <button onclick={runDoctor} class="btn-refresh" disabled={loading}>
      {loading ? 'Checking...' : 'Re-run'}
    </button>
  </div>

  {#if loading}
    <p class="muted">Running checks...</p>
  {:else if error}
    <p class="error">Error: {error}</p>
  {:else if data}
    <div class="overall" class:ok={data.ok} class:fail={!data.ok}>
      <span class="overall-icon">{data.ok ? '✓' : '✗'}</span>
      <span>{data.ok ? 'All checks passed' : 'Some checks failed'}</span>
    </div>
    <div class="checks">
      {#each (data.checks ?? []) as check}
        <div class="check-item">
          <span class="check-icon" class:pass={check.ok} class:fail={!check.ok}>
            {check.ok ? '✓' : '✗'}
          </span>
          <div class="check-info">
            <span class="check-name">{check.name}</span>
            {#if check.detail}
              <span class="check-detail">{check.detail}</span>
            {/if}
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .diagnostics-view { display: flex; flex-direction: column; gap: 1rem; }
  .header-row { display: flex; align-items: center; justify-content: space-between; }
  h2 { margin: 0; font-size: 1.3rem; }

  .btn-refresh {
    padding: 0.4rem 0.9rem;
    background: #2d3148;
    border: none;
    border-radius: 6px;
    color: #94a3b8;
    cursor: pointer;
    font-size: 0.85rem;
  }
  .btn-refresh:hover { background: #4c3d8f; color: white; }
  .btn-refresh:disabled { opacity: 0.5; cursor: default; }

  .muted { color: #64748b; }
  .error { color: #f87171; }

  .overall {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    padding: 0.75rem 1rem;
    border-radius: 8px;
    font-weight: 600;
  }
  .overall.ok { background: #052e16; color: #4ade80; border: 1px solid #166534; }
  .overall.fail { background: #2d0505; color: #f87171; border: 1px solid #7f1d1d; }
  .overall-icon { font-size: 1.1rem; }

  .checks { display: flex; flex-direction: column; gap: 0.5rem; }

  .check-item {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
    background: #1a1d27;
    border: 1px solid #2d3148;
    border-radius: 8px;
    padding: 0.75rem 1rem;
  }

  .check-icon {
    font-size: 1rem;
    flex-shrink: 0;
    margin-top: 0.05rem;
  }
  .check-icon.pass { color: #4ade80; }
  .check-icon.fail { color: #f87171; }

  .check-info { display: flex; flex-direction: column; gap: 0.15rem; }
  .check-name { font-weight: 500; font-size: 0.9rem; }
  .check-detail { font-size: 0.8rem; color: #64748b; }
</style>
