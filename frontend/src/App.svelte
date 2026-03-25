<script>
  import Rules from './tabs/Rules.svelte';
  import Integrations from './tabs/Integrations.svelte';
  import Diagnostics from './tabs/Diagnostics.svelte';

  let tab = $state('rules');
  let status = $state(null);

  $effect(() => {
    fetch('/api/status')
      .then(r => r.json())
      .then(d => { status = d; })
      .catch(() => { status = { error: 'Backend unreachable' }; });
  });

  const tabs = [
    { id: 'rules', label: 'Rules' },
    { id: 'integrations', label: 'Integrations' },
    { id: 'diagnostics', label: 'Diagnostics' },
  ];
</script>

<div class="app">
  <header>
    <div class="brand">
      <h1>Synapse</h1>
      {#if status && !status.error}
        <span class="badge">{status.rules} rules · {status.categories} categories</span>
      {/if}
    </div>
    <nav>
      {#each tabs as t}
        <button
          class="tab-btn"
          class:active={tab === t.id}
          onclick={() => tab = t.id}
        >{t.label}</button>
      {/each}
    </nav>
  </header>

  <main>
    {#if tab === 'rules'}
      <Rules />
    {:else if tab === 'integrations'}
      <Integrations />
    {:else if tab === 'diagnostics'}
      <Diagnostics />
    {/if}
  </main>
</div>

<style>
  :global(*, *::before, *::after) { box-sizing: border-box; }
  :global(body) {
    margin: 0;
    font-family: system-ui, -apple-system, sans-serif;
    background: #0f1117;
    color: #e2e8f0;
  }

  .app { min-height: 100vh; display: flex; flex-direction: column; }

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.75rem 1.5rem;
    background: #1a1d27;
    border-bottom: 1px solid #2d3148;
    gap: 1rem;
  }

  .brand { display: flex; align-items: center; gap: 0.5rem; }
  .brand h1 { margin: 0; font-size: 1.2rem; font-weight: 700; color: #a78bfa; }
.badge {
    font-size: 0.75rem;
    padding: 0.2rem 0.5rem;
    background: #2d3148;
    border-radius: 999px;
    color: #94a3b8;
  }

  nav { display: flex; gap: 0.25rem; }

  .tab-btn {
    background: none;
    border: none;
    color: #94a3b8;
    padding: 0.4rem 0.9rem;
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.9rem;
    transition: background 0.15s, color 0.15s;
  }
  .tab-btn:hover { background: #2d3148; color: #e2e8f0; }
  .tab-btn.active { background: #4c3d8f; color: #e2e8f0; font-weight: 600; }

  main { flex: 1; padding: 1.5rem; max-width: 1200px; width: 100%; margin: 0 auto; }
</style>
