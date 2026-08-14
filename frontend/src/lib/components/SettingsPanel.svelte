<script lang="ts">
  import {
    settingsOpen,
    opacity, setOpacity,
    readerRadius, setReaderRadius, RADIUS_OPTIONS,
    backgroundMode, setBackgroundMode,
    readingWidth
  } from '../stores/ui';
  import { theme, setTheme, type ThemeMode } from '../stores/settings';

  const themes: { key: ThemeMode; label: string }[] = [
    { key: 'light', label: 'Light' },
    { key: 'dark', label: 'Dark' },
    { key: 'system', label: 'System' },
  ];

  const bgModes: { key: 'gradient' | 'solid' | 'frost'; label: string }[] = [
    { key: 'gradient', label: 'Gradient' },
    { key: 'solid', label: 'Solid' },
    { key: 'frost', label: 'Frost' },
  ];

  function handleWidthInput(e: Event) {
    const target = e.target as HTMLInputElement;
    readingWidth.set(parseInt(target.value));
  }

  function handleOpacityInput(e: Event) {
    const target = e.target as HTMLInputElement;
    setOpacity(parseInt(target.value));
  }
</script>

{#if $settingsOpen}
  <div class="settings" role="dialog" aria-label="Appearance settings">
    <div class="sp-title">
      <svg viewBox="0 0 20 20"><circle cx="10" cy="10" r="2.5"/><path d="M10 2v2m0 12v2M3.5 5l1.5 1m10 8l1.5 1M2 10h2m12 0h2M3.5 15l1.5-1m10-8l1.5-1"/></svg>
      Appearance
    </div>

    <div class="sr">
      <span class="sr-label">
        <svg viewBox="0 0 20 20"><circle cx="10" cy="10" r="7"/><path d="M10 3a7 7 0 000 14" fill="var(--icon-stroke)" opacity=".15"/></svg>
        Theme
      </span>
      <div class="theme-sel" role="radiogroup" aria-label="Theme selection">
        {#each themes as t (t.key)}
          <button
            class="th-opt"
            class:on={$theme === t.key}
            role="radio"
            aria-checked={$theme === t.key}
            onclick={() => setTheme(t.key)}
          >{t.label}</button>
        {/each}
      </div>
    </div>

    <div class="sr">
      <span class="sr-label">
        <svg viewBox="0 0 20 20"><rect x="3" y="5" width="14" height="10" rx="1"/><line x1="7" y1="5" x2="7" y2="15" opacity=".3"/><line x1="13" y1="5" x2="13" y2="15" opacity=".3"/></svg>
        Reading Width
      </span>
      <input
        type="range"
        class="w-slider"
        min="600"
        max="1000"
        step="20"
        value={$readingWidth}
        oninput={handleWidthInput}
        aria-label="Reading width"
      />
    </div>

    <div class="sr">
      <span class="sr-label">
        <svg viewBox="0 0 20 20"><circle cx="10" cy="10" r="7"/><circle cx="10" cy="10" r="4" opacity=".3"/></svg>
        Opacity
      </span>
      <div class="sr-val-group">
        <input
          type="range"
          class="w-slider"
          min="40"
          max="100"
          step="5"
          value={$opacity}
          oninput={handleOpacityInput}
          aria-label="Surface opacity"
        />
        <span class="cb-val">{$opacity}%</span>
      </div>
    </div>

    <div class="sr">
      <span class="sr-label">
        <svg viewBox="0 0 20 20"><rect x="3" y="3" width="14" height="14" rx="4"/></svg>
        Window Radius
      </span>
      <div class="radius-sel" role="radiogroup" aria-label="Window corner radius">
        {#each RADIUS_OPTIONS as r (r)}
          <button
            class="rad-opt"
            class:on={$readerRadius === r}
            role="radio"
            aria-checked={$readerRadius === r}
            onclick={() => setReaderRadius(r)}
          >{r}</button>
        {/each}
      </div>
    </div>

    <div class="sr">
      <span class="sr-label">
        <svg viewBox="0 0 20 20"><path d="M2 14l4-5 3 3 4-6 5 8"/><rect x="2" y="2" width="16" height="16" rx="2"/></svg>
        Background
      </span>
      <div class="bg-sel" role="radiogroup" aria-label="Background mode">
        {#each bgModes as bg (bg.key)}
          <button
            class="bg-opt"
            class:on={$backgroundMode === bg.key}
            role="radio"
            aria-checked={$backgroundMode === bg.key}
            onclick={() => setBackgroundMode(bg.key)}
          >{bg.label}</button>
        {/each}
      </div>
    </div>
  </div>
{/if}

<style>
  .settings {
    position: absolute;
    bottom: 60px;
    left: 50%;
    transform: translateX(-50%);
    background: var(--surface-elevated);
    backdrop-filter: blur(40px);
    -webkit-backdrop-filter: blur(40px);
    border: 1px solid var(--border);
    border-radius: 18px;
    padding: 20px;
    width: 360px;
    z-index: 35;
    box-shadow: 0 12px 40px rgba(0, 0, 0, 0.3);
    user-select: none;
  }

  .sp-title {
    font-size: 13px;
    font-weight: 600;
    margin-bottom: 16px;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .sp-title svg {
    width: 16px;
    height: 16px;
    stroke: var(--icon-stroke);
    stroke-width: 1.5;
    stroke-linecap: round;
    stroke-linejoin: round;
    fill: none;
  }

  .sr {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 9px 0;
    border-bottom: 1px solid var(--border);
  }

  .sr:last-child {
    border-bottom: 0;
  }

  .sr-label {
    font-size: 13px;
    color: var(--text-secondary);
    display: flex;
    align-items: center;
    gap: 7px;
  }

  .sr-label svg {
    width: 15px;
    height: 15px;
    stroke: var(--text-tertiary);
    stroke-width: 1.5;
    stroke-linecap: round;
    stroke-linejoin: round;
    fill: none;
  }

  .sr-val-group {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .cb-val {
    font-size: 11px;
    font-family: 'JetBrains Mono', monospace;
    color: var(--text-secondary);
    min-width: 32px;
    text-align: right;
  }

  .theme-sel {
    display: flex;
    gap: 4px;
    background: var(--hover-bg);
    border-radius: 9px;
    padding: 3px;
  }

  .th-opt {
    padding: 5px 12px;
    font-size: 12px;
    color: var(--text-tertiary);
    border-radius: 7px;
    cursor: pointer;
    border: none;
    background: 0;
    font-family: inherit;
    transition: background 0.12s, color 0.12s;
  }

  .th-opt:hover {
    color: var(--text-secondary);
  }

  .th-opt.on {
    background: var(--surface-solid);
    color: var(--text-primary);
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.15);
  }

  .radius-sel {
    display: flex;
    gap: 3px;
    background: var(--hover-bg);
    border-radius: 9px;
    padding: 3px;
  }

  .rad-opt {
    padding: 4px 8px;
    font-size: 11px;
    color: var(--text-tertiary);
    border-radius: 6px;
    cursor: pointer;
    border: none;
    background: 0;
    font-family: 'JetBrains Mono', monospace;
    transition: background 0.12s, color 0.12s;
  }

  .rad-opt:hover {
    color: var(--text-secondary);
  }

  .rad-opt.on {
    background: var(--surface-solid);
    color: var(--text-primary);
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.15);
  }

  .bg-sel {
    display: flex;
    gap: 3px;
    background: var(--hover-bg);
    border-radius: 9px;
    padding: 3px;
  }

  .bg-opt {
    padding: 4px 8px;
    font-size: 11px;
    color: var(--text-tertiary);
    border-radius: 6px;
    cursor: pointer;
    border: none;
    background: 0;
    font-family: inherit;
    transition: background 0.12s, color 0.12s;
  }

  .bg-opt:hover {
    color: var(--text-secondary);
  }

  .bg-opt.on {
    background: var(--surface-solid);
    color: var(--text-primary);
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.15);
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
