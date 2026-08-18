<script lang="ts">
  import { zoomLevel, zoomIn, zoomOut, resetZoom, readingWidth, toggleFocusMode, focusMode, toggleSettings } from '../stores/ui';
  import { theme, setTheme, type ThemeMode } from '../stores/settings';
  import { activeTab, setStreamActive } from '../stores/tabs';
  import { streamActive, cancelStreamState, activeStream } from '../stores/stream';
  import { get } from 'svelte/store';

  let copied = $state(false);
  let copyTimer: ReturnType<typeof setTimeout> | null = $state(null);

  function cycleTheme() {
    const modes: ThemeMode[] = ['system', 'light', 'dark'];
    const current = $theme;
    const idx = modes.indexOf(current);
    const next = modes[(idx + 1) % modes.length];
    setTheme(next);
  }

  function narrowWidth() {
    readingWidth.update(w => Math.max(600, w - 40));
  }

  function widenWidth() {
    readingWidth.update(w => Math.min(1000, w + 40));
  }

  async function handleStopStream() {
    try {
      const App = await import('../../../wailsjs/go/main/App');
      await App.CancelStream();
    } catch { /* ignore */ }
    cancelStreamState();
    const stream = get(activeStream);
    if (stream) {
      setStreamActive(stream.tabId, false);
    }
  }

  async function copyDocument() {
    const tab = $activeTab;
    if (!tab) return;
    try {
      const { ClipboardSetText } = await import('../../../wailsjs/runtime/runtime');
      await ClipboardSetText(tab.content);
    } catch {
      // Fallback for dev mode
      await navigator.clipboard.writeText(tab.content);
    }
    copied = true;
    if (copyTimer) clearTimeout(copyTimer);
    copyTimer = setTimeout(() => { copied = false; }, 1500);
  }
</script>

<div class="controls" role="toolbar" aria-label="Document controls">
  <button class="cb" onclick={zoomOut} aria-label="Zoom out" title="Zoom out (Ctrl+-)">
    <svg viewBox="0 0 20 20"><circle cx="8.5" cy="8.5" r="5.5"/><line x1="13" y1="13" x2="17" y2="17"/><line x1="6" y1="8.5" x2="11" y2="8.5"/></svg>
  </button>

  <button class="cb-val" onclick={resetZoom} title="Reset zoom (Ctrl+0)" aria-label="Current zoom {$zoomLevel}%">
    {$zoomLevel}%
  </button>

  <button class="cb" onclick={zoomIn} aria-label="Zoom in" title="Zoom in (Ctrl+=)">
    <svg viewBox="0 0 20 20"><circle cx="8.5" cy="8.5" r="5.5"/><line x1="13" y1="13" x2="17" y2="17"/><line x1="6" y1="8.5" x2="11" y2="8.5"/><line x1="8.5" y1="6" x2="8.5" y2="11"/></svg>
  </button>

  {#if $streamActive}
    <div class="cb-div" aria-hidden="true"></div>
    <button class="cb cb-stop" title="Stop receiving response (Esc)" aria-label="Stop streaming" onclick={handleStopStream}>
      <svg viewBox="0 0 20 20"><rect x="5" y="5" width="10" height="10" rx="1.5" fill="currentColor" stroke="none"/></svg>
    </button>
  {/if}

  <div class="cb-div" aria-hidden="true"></div>

  <button class="cb" onclick={narrowWidth} aria-label="Decrease reading width" title="Narrower">
    <svg viewBox="0 0 20 20"><line x1="10" y1="3" x2="10" y2="17"/><polyline points="3,7 3,5 17,5 17,7"/><polyline points="3,13 3,15 17,15 17,13"/><polyline points="5,8.5 7,10 5,11.5"/><polyline points="15,8.5 13,10 15,11.5"/></svg>
  </button>

  <span class="cb-val" aria-label="Current reading width">{$readingWidth}</span>

  <button class="cb" onclick={widenWidth} aria-label="Increase reading width" title="Wider">
    <svg viewBox="0 0 20 20"><line x1="10" y1="3" x2="10" y2="17"/><polyline points="3,7 3,5 17,5 17,7"/><polyline points="3,13 3,15 17,15 17,13"/><polyline points="5,8.5 3,10 5,11.5"/><polyline points="15,8.5 17,10 15,11.5"/></svg>
  </button>

  <div class="cb-div" aria-hidden="true"></div>

  <button class="cb" class:on={$focusMode} onclick={toggleFocusMode} aria-label="Toggle focus mode" aria-pressed={$focusMode} title="Focus mode (F11)">
    <svg viewBox="0 0 20 20"><path d="M3 7V4a1 1 0 011-1h3"/><path d="M13 3h3a1 1 0 011 1v3"/><path d="M17 13v3a1 1 0 01-1 1h-3"/><path d="M7 17H4a1 1 0 01-1-1v-3"/></svg>
  </button>

  <button class="cb" onclick={cycleTheme} aria-label="Toggle theme" title="Theme: {$theme}">
    <svg viewBox="0 0 20 20"><circle cx="10" cy="10" r="7"/><path d="M10 3a7 7 0 000 14" fill="var(--icon-stroke)" opacity=".15"/></svg>
  </button>

  <button class="cb" onclick={copyDocument} disabled={!$activeTab} aria-label="Copy document to clipboard" title="Copy document">
    {#if copied}
      <svg viewBox="0 0 20 20"><polyline points="4 10 8 14 16 5"/></svg>
    {:else}
      <svg viewBox="0 0 20 20"><rect x="6" y="6" width="10" height="12" rx="1.5"/><path d="M4 14V4a1.5 1.5 0 011.5-1.5H13"/></svg>
    {/if}
  </button>

  <div class="cb-div" aria-hidden="true"></div>

  <button class="cb" onclick={toggleSettings} aria-label="Open settings panel" title="Settings">
    <svg viewBox="0 0 20 20"><circle cx="10" cy="10" r="2.5"/><path d="M10 2v2m0 12v2M3.5 5l1.5 1m10 8l1.5 1M2 10h2m12 0h2M3.5 15l1.5-1m10-8l1.5-1"/></svg>
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
  .controls:hover {
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

  .cb:hover:not(:disabled) {
    background: var(--hover-bg);
  }

  .cb:disabled {
    opacity: 0.35;
    cursor: default;
  }

  .cb.on {
    background: var(--active-bg);
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

  .cb.on svg {
    stroke: var(--icon-active);
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

  /* Stop button (Design.md) */
  .cb-stop {
    color: var(--icon-stroke);
    transition: background 0.12s, color 0.12s;
  }

  .cb-stop:hover {
    background: var(--stream-stop-bg-dim);
    color: var(--stream-stop-bg);
  }

  .cb-stop svg {
    stroke: none;
    fill: currentColor;
  }
</style>
