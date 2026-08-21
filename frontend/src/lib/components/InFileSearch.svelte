<script lang="ts">
  interface Props {
    visible: boolean;
    matchCount: number;
    currentMatch: number;
    onsearch: (query: string) => void;
    onnext: () => void;
    onprev: () => void;
    onclose: () => void;
  }

  let { visible, matchCount, currentMatch, onsearch, onnext, onprev, onclose }: Props = $props();

  let inputEl: HTMLInputElement | undefined = $state();
  let query = $state('');

  $effect(() => {
    if (visible) {
      requestAnimationFrame(() => inputEl?.focus());
    } else {
      query = '';
    }
  });

  function handleInput() {
    onsearch(query);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      e.preventDefault();
      e.stopPropagation();
      onclose();
      return;
    }
    if (e.key === 'Enter') {
      e.preventDefault();
      if (e.shiftKey) {
        onprev();
      } else {
        onnext();
      }
    }
  }
</script>

{#if visible}
  <div class="in-file-search" role="search" aria-label="Find in document">
    <input
      bind:this={inputEl}
      bind:value={query}
      oninput={handleInput}
      onkeydown={handleKeydown}
      placeholder="Find in document..."
      autocomplete="off"
      spellcheck="false"
      aria-label="Search text"
    />
    <span class="ifs-count" aria-live="polite">
      {#if query && matchCount > 0}
        {currentMatch} of {matchCount}
      {:else if query}
        No matches
      {/if}
    </span>
    <button class="ifs-btn" title="Previous (Shift+Enter)" aria-label="Previous match" onclick={onprev} disabled={matchCount === 0}>
      <svg viewBox="0 0 16 16"><polyline points="12 10 8 6 4 10"/></svg>
    </button>
    <button class="ifs-btn" title="Next (Enter)" aria-label="Next match" onclick={onnext} disabled={matchCount === 0}>
      <svg viewBox="0 0 16 16"><polyline points="4 6 8 10 12 6"/></svg>
    </button>
    <button class="ifs-btn ifs-close" title="Close (Esc)" aria-label="Close search" onclick={onclose}>
      <svg viewBox="0 0 16 16"><line x1="4" y1="4" x2="12" y2="12"/><line x1="12" y1="4" x2="4" y2="12"/></svg>
    </button>
  </div>
{/if}

<style>
  .in-file-search {
    position: absolute;
    top: 8px;
    right: 16px;
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 6px 10px;
    background: var(--surface-elevated);
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    border: 1px solid var(--border);
    border-radius: 12px;
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.2);
    z-index: 40;
    animation: ifs-in 120ms ease-out;
  }

  .in-file-search input {
    background: none;
    border: none;
    outline: none;
    color: var(--text-primary);
    font-size: 13px;
    font-family: inherit;
    width: 200px;
    padding: 4px 6px;
  }

  .in-file-search input::placeholder {
    color: var(--text-tertiary);
  }

  .ifs-count {
    font-size: 11px;
    color: var(--text-tertiary);
    font-family: 'JetBrains Mono', monospace;
    white-space: nowrap;
    min-width: 56px;
    text-align: center;
  }

  .ifs-btn {
    width: 24px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: none;
    background: transparent;
    border-radius: 6px;
    cursor: pointer;
    padding: 0;
    transition: background 0.1s;
  }

  .ifs-btn:hover:not(:disabled) {
    background: var(--hover-bg);
  }

  .ifs-btn:disabled {
    opacity: 0.3;
    cursor: default;
  }

  .ifs-btn svg {
    width: 14px;
    height: 14px;
    stroke: var(--text-secondary);
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
    fill: none;
  }

  .ifs-close svg {
    width: 12px;
    height: 12px;
  }

  @keyframes ifs-in {
    from { opacity: 0; transform: translateY(-4px); }
    to { opacity: 1; transform: translateY(0); }
  }

  @media (prefers-reduced-motion: reduce) {
    .in-file-search { animation: none; }
  }
</style>
