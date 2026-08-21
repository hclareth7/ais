<script lang="ts">
  interface Props {
    tabs: Array<{ id: string; name: string }>;
    selectedIndex: number;
    visible: boolean;
  }

  let { tabs, selectedIndex, visible }: Props = $props();
</script>

{#if visible && tabs.length > 0}
  <div class="mru-overlay" role="presentation">
    <div class="mru-panel">
      <div class="mru-title">Switch Tab</div>
      <div role="listbox" aria-label="Recent tabs">
        {#each tabs as tab, i (tab.id)}
          <div
            class="mru-entry"
            class:selected={i === selectedIndex}
            role="option"
            aria-selected={i === selectedIndex}
          >
            <svg viewBox="0 0 24 24" class="mru-icon" aria-hidden="true">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
              <polyline points="14 2 14 8 20 8"/>
            </svg>
            <span class="mru-name">{tab.name}</span>
          </div>
        {/each}
      </div>
    </div>
  </div>
{/if}

<style>
  .mru-overlay {
    position: fixed;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(0, 0, 0, 0.3);
    z-index: 100;
    animation: mru-fade-in 120ms ease-out;
  }

  .mru-panel {
    min-width: 280px;
    max-width: 420px;
    padding: 8px;
    background: var(--surface-elevated);
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    border: 1px solid var(--border);
    border-radius: 16px;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.4);
  }

  .mru-title {
    font-size: 12px;
    color: var(--text-tertiary);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    padding: 8px 16px 4px;
  }

  .mru-entry {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 16px;
    border-radius: 10px;
    font-size: 14px;
    color: var(--text-primary);
    background: transparent;
    transition: background 0.08s;
  }

  .mru-entry.selected {
    background: var(--active-bg);
    color: var(--accent-text);
  }

  .mru-icon {
    width: 16px;
    height: 16px;
    stroke: currentColor;
    stroke-width: 1.5;
    stroke-linecap: round;
    fill: none;
    flex-shrink: 0;
    opacity: 0.7;
  }

  .mru-entry.selected .mru-icon {
    opacity: 1;
  }

  .mru-name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  @keyframes mru-fade-in {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  @media (prefers-reduced-motion: reduce) {
    .mru-overlay {
      animation: none;
    }
  }
</style>
