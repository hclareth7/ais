<script lang="ts">
  import { tabs, activeTabId, closeTab } from '../stores/tabs';

  let tabCount = $derived($tabs.length);

  function handleTabClick(id: string) {
    activeTabId.set(id);
  }

  function handleMiddleClick(e: MouseEvent, id: string) {
    if (e.button === 1) {
      e.preventDefault();
      closeTab(id);
    }
  }

  function handleCloseClick(e: MouseEvent, id: string) {
    e.stopPropagation();
    closeTab(id);
  }

  function handleTabKeydown(e: KeyboardEvent, id: string) {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      handleTabClick(id);
    }
  }
</script>

{#if tabCount > 0}
  <div class="tabbar" class:single={tabCount <= 1} role="tablist" aria-label="Open files">
    {#each $tabs as tab (tab.id)}
      <div
        class="tab"
        class:on={$activeTabId === tab.id}
        role="tab"
        tabindex="0"
        aria-selected={$activeTabId === tab.id}
        onclick={() => handleTabClick(tab.id)}
        onkeydown={(e) => handleTabKeydown(e, tab.id)}
        onauxclick={(e) => handleMiddleClick(e, tab.id)}
        title={tab.path}
      >
        <span class="tab-name">{tab.name}</span>
        <button
          class="tc"
          tabindex="-1"
          aria-label="Close {tab.name}"
          onclick={(e) => handleCloseClick(e, tab.id)}
        >
          <svg viewBox="0 0 24 24"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
        </button>
      </div>
    {/each}
  </div>
{/if}

<style>
  .tabbar {
    display: flex;
    align-items: center;
    gap: 3px;
    padding: 0 22px 6px;
    flex-shrink: 0;
    transition: opacity 0.12s, height 0.12s, padding 0.12s;
    overflow: hidden;
    user-select: none;
  }

  .tabbar.single {
    opacity: 0;
    height: 0;
    padding: 0;
    pointer-events: none;
  }

  .tab {
    display: flex;
    align-items: center;
    gap: 7px;
    padding: 5px 14px;
    font-size: 13px;
    color: var(--text-tertiary);
    cursor: pointer;
    border-radius: 8px;
    white-space: nowrap;
    transition: color 0.12s, background 0.12s;
    border: none;
    background: none;
    position: relative;
  }

  .tab:hover {
    color: var(--text-secondary);
    background: var(--hover-bg);
  }

  .tab.on {
    color: var(--text-primary);
    background: var(--accent-dim);
  }

  .tab:focus-visible {
    outline: 2px solid var(--border-focus);
    outline-offset: 4px;
  }

  .tab-name {
    max-width: 160px;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .tc {
    width: 16px;
    height: 16px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 4px;
    opacity: 0;
    transition: opacity 0.12s;
    border: none;
    background: none;
    cursor: pointer;
    padding: 0;
  }

  .tc svg {
    width: 10px;
    height: 10px;
    stroke: var(--text-tertiary);
    stroke-width: 2;
    stroke-linecap: round;
    fill: none;
  }

  .tab:hover .tc {
    opacity: 0.5;
  }

  .tc:hover {
    opacity: 1 !important;
    background: var(--hover-bg);
  }
</style>
