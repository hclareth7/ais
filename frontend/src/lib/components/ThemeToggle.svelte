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
    gap: 8px;
    padding: 6px 12px;
    border: 1px solid var(--border);
    border-radius: 6px;
    background: var(--bg-primary);
    color: var(--text-secondary);
    cursor: pointer;
    font-size: 14px;
    transition: background-color 150ms ease, color 150ms ease;
  }

  .theme-toggle:hover {
    background: var(--bg-elevated);
    color: var(--text-primary);
  }

  .theme-toggle:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }

  .theme-icon {
    font-size: 16px;
    line-height: 1;
  }

  .theme-label {
    font-size: 12px;
  }
</style>
