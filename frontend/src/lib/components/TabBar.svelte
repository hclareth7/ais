<script lang="ts">
  import { tabs, activeTabId, closeTab } from '../stores/tabs';

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

{#if $tabs.length > 0}
  <div class="tab-bar" role="tablist" aria-label="Open files">
    {#each $tabs as tab (tab.id)}
      <div
        class="tab"
        class:active={$activeTabId === tab.id}
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
          class="tab-close"
          tabindex="-1"
          aria-label="Close {tab.name}"
          onclick={(e) => handleCloseClick(e, tab.id)}
        >&times;</button>
      </div>
    {/each}
  </div>
{/if}

<style>
  .tab-bar {
    grid-area: tabs;
    display: flex;
    align-items: stretch;
    background: var(--bg-secondary);
    border-bottom: 1px solid var(--border);
    overflow-x: auto;
    overflow-y: hidden;
    scrollbar-width: none;
  }

  .tab-bar::-webkit-scrollbar {
    display: none;
  }

  .tab {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 0 16px;
    border: none;
    background: none;
    color: var(--text-secondary);
    font-size: 13px;
    cursor: pointer;
    white-space: nowrap;
    flex-shrink: 0;
    border-bottom: 2px solid transparent;
    height: 100%;
    box-sizing: border-box;
  }

  .tab:hover {
    background: var(--bg-inset);
    color: var(--text-primary);
  }

  .tab.active {
    color: var(--text-primary);
    background: var(--accent-subtle);
    border-bottom-color: var(--accent);
  }

  .tab:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: -2px;
  }

  .tab-name {
    max-width: 160px;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .tab-close {
    font-size: 16px;
    line-height: 1;
    opacity: 0;
    color: var(--text-muted);
    padding: 0 4px;
    border: none;
    background: none;
    border-radius: 4px;
    cursor: pointer;
    transition: opacity 100ms ease;
  }

  .tab:hover .tab-close,
  .tab.active .tab-close {
    opacity: 1;
  }

  .tab-close:hover {
    color: var(--text-primary);
    background: var(--bg-inset);
  }
</style>
