<script lang="ts">
  interface Props {
    visible: boolean;
    text: string;
    language: string;
    loading: boolean;
    position: { x: number; y: number };
    onclose: () => void;
  }

  let { visible, text, language, loading, position, onclose }: Props = $props();

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      e.preventDefault();
      e.stopPropagation();
      onclose();
    }
  }

  $effect(() => {
    if (visible) {
      document.addEventListener('keydown', handleKeydown, true);
      return () => document.removeEventListener('keydown', handleKeydown, true);
    }
  });
</script>

{#if visible}
  <div
    class="translation-popover"
    style="left: {position.x}px; top: {position.y}px;"
    role="tooltip"
    aria-live="polite"
  >
    <span class="tp-lang">{language.toUpperCase()}</span>
    {#if loading}
      <span class="tp-loading">...</span>
    {:else}
      <span class="tp-text">{text}</span>
    {/if}
  </div>
{/if}

<style>
  .translation-popover {
    position: fixed;
    transform: translateX(-50%);
    max-width: 400px;
    padding: 8px 12px;
    background: var(--surface-elevated);
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    border: 1px solid var(--border);
    border-radius: 12px;
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.25);
    z-index: 51;
    display: flex;
    align-items: flex-start;
    gap: 8px;
    animation: tp-in 100ms ease-out;
  }

  .tp-lang {
    font-size: 10px;
    font-family: 'JetBrains Mono', monospace;
    font-weight: 600;
    color: var(--accent-text);
    background: var(--accent-dim);
    padding: 2px 6px;
    border-radius: 4px;
    white-space: nowrap;
    flex-shrink: 0;
    margin-top: 1px;
  }

  .tp-text {
    font-size: 14px;
    color: var(--text-secondary);
    line-height: 1.5;
    word-wrap: break-word;
    overflow-wrap: break-word;
  }

  .tp-loading {
    font-size: 14px;
    color: var(--text-tertiary);
    animation: tp-pulse 1s ease-in-out infinite;
  }

  @keyframes tp-in {
    from { opacity: 0; transform: translateX(-50%) translateY(-4px); }
    to { opacity: 1; transform: translateX(-50%) translateY(0); }
  }

  @keyframes tp-pulse {
    0%, 100% { opacity: 0.4; }
    50% { opacity: 1; }
  }

  @media (prefers-reduced-motion: reduce) {
    .translation-popover { animation: none; }
    .tp-loading { animation: none; }
  }
</style>
