<script>
  let integrations = $state([]);
  let loading = $state(true);
  let error = $state(null);

  $effect(() => {
    fetch('/api/integrations')
      .then(r => r.json())
      .then(d => {
        integrations = d.integrations ?? [];
        loading = false;
      })
      .catch(e => {
        error = e.message;
        loading = false;
      });
  });
</script>

<div class="integrations-view">
  <h2>Integrations</h2>
  <p class="subtitle">Third-party tools Synapse can read data from.</p>

  {#if loading}
    <p class="muted">Loading integrations...</p>
  {:else if error}
    <p class="error">Error: {error}</p>
  {:else if integrations.length === 0}
    <p class="muted">No integration manifests loaded.</p>
  {:else}
    <div class="list">
      {#each integrations as itg}
        <div class="item">
          <div class="item-left">
            <span class="item-name">{itg.name}</span>
            <span class="item-desc">{itg.description ?? ''}</span>
          </div>
          <div class="item-right">
            <span class="status-dot" class:green={itg.available} title={itg.available ? 'Binary found' : 'Not installed'}></span>
            <span class="status-label">{itg.available ? 'Installed' : 'Not found'}</span>
            {#if itg.data_found > 0}
              <span class="data-badge">{itg.data_found} data file{itg.data_found !== 1 ? 's' : ''}</span>
            {/if}
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .integrations-view { display: flex; flex-direction: column; gap: 1rem; }
  h2 { margin: 0; font-size: 1.3rem; }
  .subtitle { margin: 0; color: #94a3b8; font-size: 0.9rem; }
  .muted { color: #64748b; }
  .error { color: #f87171; }

  .list { display: flex; flex-direction: column; gap: 0.5rem; }

  .item {
    background: #1a1d27;
    border: 1px solid #2d3148;
    border-radius: 8px;
    padding: 0.9rem 1.1rem;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
  }
  .item:hover { border-color: #4c3d8f; }

  .item-left { display: flex; flex-direction: column; gap: 0.2rem; }
  .item-name { font-weight: 600; }
  .item-desc { font-size: 0.82rem; color: #94a3b8; }

  .item-right { display: flex; align-items: center; gap: 0.6rem; flex-shrink: 0; }

  .status-dot {
    width: 8px; height: 8px;
    border-radius: 50%;
    background: #ef4444;
    flex-shrink: 0;
  }
  .status-dot.green { background: #10b981; }
  .status-label { font-size: 0.85rem; color: #94a3b8; }

  .data-badge {
    font-size: 0.75rem;
    padding: 0.15rem 0.5rem;
    background: #2d3148;
    border-radius: 999px;
    color: #94a3b8;
  }
</style>
