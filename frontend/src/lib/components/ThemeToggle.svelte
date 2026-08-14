<script lang="ts">
  import { theme, setTheme, type ThemeMode } from '../stores/settings';

  function cycleTheme() {
    const modes: ThemeMode[] = ['system', 'light', 'dark'];
    const current = $theme;
    const idx = modes.indexOf(current);
    const next = modes[(idx + 1) % modes.length];
    setTheme(next);
  }

  function getIcon(mode: ThemeMode): string {
    switch (mode) {
      case 'light': return '☀';
      case 'dark': return '●';
      case 'system': return '◐';
    }
  }

  function getLabel(mode: ThemeMode): string {
    switch (mode) {
      case 'light': return 'Light';
      case 'dark': return 'Dark';
      case 'system': return 'System';
    }
  }
</script>

<button
  class="theme-toggle"
  onclick={cycleTheme}
  aria-label="Toggle theme: {getLabel($theme)}"
  title="Theme: {getLabel($theme)}"
>
  <span class="theme-icon">{getIcon($theme)}</span>
  <span class="theme-label">{getLabel($theme)}</span>
</button>

<style>
  .theme-toggle {
    display: flex;
    align-items: center;
    gap: 9px;
    padding: 7px 10px;
    font-size: 14px;
    color: var(--text-secondary);
    border-radius: 10px;
    cursor: pointer;
    transition: background 0.12s, color 0.12s;
    border: none;
    background: none;
    font-family: inherit;
    width: 100%;
  }

  .theme-toggle:hover {
    background: var(--hover-bg);
    color: var(--text-primary);
  }

  .theme-toggle:focus-visible {
    outline: 2px solid var(--border-focus);
    outline-offset: 4px;
  }

  .theme-icon {
    font-size: 16px;
    line-height: 1;
    width: 16px;
    text-align: center;
  }

  .theme-label {
    font-size: 13px;
  }
</style>
