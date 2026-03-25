<script>
  let rules = $state([]);
  let categories = $state([]);
  let loading = $state(true);
  let error = $state(null);
  let search = $state('');
  let typeFilter = $state('');
  let searchResults = $state(null);
  let searching = $state(false);

  $effect(() => {
    Promise.all([
      fetch('/api/rules').then(r => r.json()),
      fetch('/api/rules/categories').then(r => r.json()),
    ])
      .then(([rulesData, catsData]) => {
        rules = rulesData.rules ?? [];
        categories = catsData.categories ?? [];
        loading = false;
      })
      .catch(e => {
        error = e.message;
        loading = false;
      });
  });

  async function doSearch() {
    if (!search.trim()) {
      searchResults = null;
      return;
    }
    searching = true;
    try {
      const r = await fetch(`/api/rules/search?q=${encodeURIComponent(search)}`);
      const d = await r.json();
      searchResults = d.results ?? [];
    } catch {
      searchResults = [];
    } finally {
      searching = false;
    }
  }

  const displayed = $derived(
    searchResults !== null
      ? searchResults
      : typeFilter
        ? rules.filter(r => r.type === typeFilter)
        : rules
  );

  const typeColors = {
    rule: '#6366f1',
    skill: '#10b981',
    agent: '#f59e0b',
    command: '#ec4899',
    context: '#3b82f6',
  };
</script>

<div class="rules-view">
  <div class="toolbar">
    <div class="search-row">
      <input
        type="text"
        placeholder="Search rules..."
        bind:value={search}
        onkeydown={e => e.key === 'Enter' && doSearch()}
        class="search-input"
      />
      <button onclick={doSearch} class="btn-primary" disabled={searching}>
        {searching ? '...' : 'Search'}
      </button>
      {#if searchResults !== null}
        <button onclick={() => { search = ''; searchResults = null; }} class="btn-ghost">
          Clear
        </button>
      {/if}
    </div>
    <div class="filters">
      <span class="filter-label">Filter:</span>
      {#each ['', 'rule', 'skill', 'agent', 'command', 'context'] as t}
        <button
          class="chip"
          class:active={typeFilter === t}
          onclick={() => typeFilter = t}
        >{t || 'All'}</button>
      {/each}
    </div>
  </div>

  {#if loading}
    <p class="muted">Loading rules...</p>
  {:else if error}
    <p class="error">Error: {error}</p>
  {:else}
    <p class="count">{displayed.length} rule{displayed.length !== 1 ? 's' : ''}</p>
    <div class="grid">
      {#each displayed as rule}
        <div class="card">
          <div class="card-header">
            <span class="name">{rule.name}</span>
            <span class="type-badge" style="background: {typeColors[rule.type] ?? '#555'}">
              {rule.type}
            </span>
          </div>
          {#if rule.description}
            <p class="desc">{rule.description}</p>
          {/if}
          {#if rule.keywords?.length}
            <div class="keywords">
              {#each rule.keywords.slice(0, 6) as kw}
                <span class="keyword">{kw}</span>
              {/each}
            </div>
          {/if}
          {#if rule.always_apply}
            <span class="always-badge">always active</span>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .rules-view { display: flex; flex-direction: column; gap: 1rem; }

  .toolbar { display: flex; flex-direction: column; gap: 0.75rem; }
  .search-row { display: flex; gap: 0.5rem; align-items: center; }
  .search-input {
    flex: 1;
    padding: 0.5rem 0.75rem;
    background: #1a1d27;
    border: 1px solid #2d3148;
    border-radius: 6px;
    color: #e2e8f0;
    font-size: 0.9rem;
    outline: none;
  }
  .search-input:focus { border-color: #6366f1; }

  .btn-primary {
    padding: 0.5rem 1rem;
    background: #6366f1;
    color: white;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.9rem;
  }
  .btn-primary:disabled { opacity: 0.5; cursor: default; }
  .btn-ghost {
    padding: 0.5rem 0.75rem;
    background: none;
    border: 1px solid #2d3148;
    border-radius: 6px;
    color: #94a3b8;
    cursor: pointer;
    font-size: 0.85rem;
  }

  .filters { display: flex; align-items: center; gap: 0.4rem; flex-wrap: wrap; }
  .filter-label { font-size: 0.8rem; color: #64748b; }
  .chip {
    padding: 0.2rem 0.6rem;
    border-radius: 999px;
    border: 1px solid #2d3148;
    background: none;
    color: #94a3b8;
    font-size: 0.8rem;
    cursor: pointer;
    transition: all 0.15s;
  }
  .chip:hover { border-color: #6366f1; color: #e2e8f0; }
  .chip.active { background: #6366f1; border-color: #6366f1; color: white; }

  .count { font-size: 0.85rem; color: #64748b; margin: 0; }
  .muted { color: #64748b; }
  .error { color: #f87171; }

  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 0.75rem;
  }

  .card {
    background: #1a1d27;
    border: 1px solid #2d3148;
    border-radius: 8px;
    padding: 0.9rem;
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
  }
  .card:hover { border-color: #4c3d8f; }

  .card-header { display: flex; align-items: center; justify-content: space-between; gap: 0.5rem; }
  .name { font-weight: 600; font-size: 0.95rem; }
  .type-badge {
    font-size: 0.7rem;
    padding: 0.15rem 0.5rem;
    border-radius: 999px;
    color: white;
    font-weight: 500;
    flex-shrink: 0;
  }

  .desc { font-size: 0.82rem; color: #94a3b8; margin: 0; line-height: 1.4; }

  .keywords { display: flex; flex-wrap: wrap; gap: 0.3rem; }
  .keyword {
    font-size: 0.72rem;
    padding: 0.1rem 0.4rem;
    background: #2d3148;
    border-radius: 4px;
    color: #94a3b8;
  }

  .always-badge {
    font-size: 0.72rem;
    color: #10b981;
    border: 1px solid #10b981;
    padding: 0.1rem 0.4rem;
    border-radius: 4px;
    align-self: flex-start;
  }
</style>
