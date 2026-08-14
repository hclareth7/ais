<script lang="ts">
  import { zoomLevel, zoomIn, zoomOut, resetZoom, readingWidth, toggleFocusMode } from '../stores/ui';
  import { theme, setTheme, type ThemeMode } from '../stores/settings';

  function cycleTheme() {
    const modes: ThemeMode[] = ['system', 'light', 'dark'];
    const current = $theme;
    const idx = modes.indexOf(current);
    const next = modes[(idx + 1) % modes.length];
    setTheme(next);
  }

  function handleWidthChange(e: Event) {
    const target = e.target as HTMLInputElement;
    readingWidth.set(parseInt(target.value));
  }
</script>

<div class="controls">
  <button class="cb" onclick={zoomOut} aria-label="Zoom out" title="Zoom out (Ctrl+-)">
    <svg viewBox="0 0 24 24"><line x1="5" y1="12" x2="19" y2="12"/></svg>
  </button>

  <button class="cb-val" onclick={resetZoom} title="Reset zoom (Ctrl+0)">
    {$zoomLevel}%
  </button>

  <button class="cb" onclick={zoomIn} aria-label="Zoom in" title="Zoom in (Ctrl+=)">
    <svg viewBox="0 0 24 24"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
  </button>

  <div class="cb-div"></div>

  <input
    type="range"
    class="w-slider"
    min="500"
    max="1000"
    step="20"
    value={$readingWidth}
    oninput={handleWidthChange}
    aria-label="Reading width"
    title="Reading width: {$readingWidth}px"
  />

  <div class="cb-div"></div>

  <button class="cb" onclick={cycleTheme} aria-label="Toggle theme" title="Theme: {$theme}">
    <svg viewBox="0 0 24 24">
      {#if $theme === 'dark'}
        <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
      {:else if $theme === 'light'}
        <circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/>
      {:else}
        <circle cx="12" cy="12" r="9"/><path d="M12 3a9 9 0 0 0 0 18V3z"/>
      {/if}
    </svg>
  </button>

  <button class="cb" onclick={toggleFocusMode} aria-label="Toggle focus mode" title="Focus mode (F11)">
    <svg viewBox="0 0 24 24"><polyline points="15 3 21 3 21 9"/><polyline points="9 21 3 21 3 15"/><line x1="21" y1="3" x2="14" y2="10"/><line x1="3" y1="21" x2="10" y2="14"/></svg>
  </button>
</div>

<style>
  .controls {
    position: absolute;
    bottom: 16px;
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    align-items: center;
    gap: 3px;
    background: var(--surface-elevated);
    backdrop-filter: blur(30px);
    -webkit-backdrop-filter: blur(30px);
    border: 1px solid var(--border);
    border-radius: 28px;
    padding: 5px 6px;
    z-index: 30;
    opacity: 0;
    transition: opacity 0.25s;
    pointer-events: none;
    user-select: none;
  }

  :global(.bottom-trigger:hover ~ .controls),
  .controls:hover,
  :global(.reader.focus) .controls {
    opacity: 1;
    pointer-events: auto;
  }

  .cb {
    width: 34px;
    height: 34px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: none;
    background: 0;
    cursor: pointer;
    border-radius: 50%;
    transition: background 0.12s;
  }

  .cb:hover {
    background: var(--hover-bg);
  }

  .cb svg {
    width: 17px;
    height: 17px;
    stroke: var(--icon-stroke);
    stroke-width: 1.5;
    stroke-linecap: round;
    stroke-linejoin: round;
    fill: none;
    transition: stroke 0.12s;
  }

  .cb:hover svg {
    stroke: var(--text-primary);
  }

  .cb:focus-visible {
    outline: 2px solid var(--border-focus);
    outline-offset: 4px;
  }

  .cb-div {
    width: 1px;
    height: 18px;
    background: var(--border);
    margin: 0 2px;
  }

  .cb-val {
    font-size: 11px;
    font-family: 'JetBrains Mono', monospace;
    color: var(--text-secondary);
    min-width: 36px;
    text-align: center;
    cursor: pointer;
    border: none;
    background: none;
    border-radius: 6px;
    padding: 4px;
    transition: background 0.12s;
  }

  .cb-val:hover {
    background: var(--hover-bg);
  }

  .w-slider {
    width: 80px;
    height: 4px;
    -webkit-appearance: none;
    appearance: none;
    background: var(--border);
    border-radius: 2px;
    outline: 0;
    cursor: pointer;
  }

  .w-slider::-webkit-slider-thumb {
    -webkit-appearance: none;
    width: 14px;
    height: 14px;
    background: var(--accent-solid);
    border-radius: 50%;
    cursor: pointer;
    box-shadow: 0 1px 6px rgba(0, 0, 0, 0.3);
  }

  .w-slider::-moz-range-thumb {
    width: 14px;
    height: 14px;
    background: var(--accent-solid);
    border-radius: 50%;
    cursor: pointer;
    border: 0;
    box-shadow: 0 1px 6px rgba(0, 0, 0, 0.3);
  }
</style>
