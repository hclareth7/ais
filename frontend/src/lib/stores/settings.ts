import { writable, get } from 'svelte/store';

export type ThemeMode = 'light' | 'dark' | 'system';

export const theme = writable<ThemeMode>('system');

export async function loadSettings(): Promise<void> {
  try {
    const { GetTheme, GetConfig } = await import('../../../wailsjs/go/main/App');
    const t = await GetTheme();
    if (t === 'light' || t === 'dark' || t === 'system') {
      theme.set(t);
    }
    applyTheme(get(theme));

    // Load UI settings from config
    try {
      const cfg: any = await GetConfig();
      const { opacity, zoomLevel, readingWidth, readerRadius, backgroundMode } = await import('./ui');
      if (cfg.opacity) opacity.set(cfg.opacity);
      if (cfg.zoomLevel) zoomLevel.set(cfg.zoomLevel);
      if (cfg.readingWidth) readingWidth.set(cfg.readingWidth);
      if (cfg.readerRadius) {
        readerRadius.set(cfg.readerRadius);
        document.documentElement.style.setProperty('--reader-radius', `${cfg.readerRadius}px`);
      }
      if (cfg.backgroundMode) backgroundMode.set(cfg.backgroundMode);
      // Apply opacity
      if (cfg.opacity) {
        document.documentElement.style.setProperty('--surface-opacity', String(cfg.opacity / 100));
      }
    } catch (err) {
      console.error('Failed to load UI settings:', err);
    }
  } catch (err) {
    console.error('Failed to load settings:', err);
    applyTheme('system');
  }
}

export async function setTheme(mode: ThemeMode): Promise<void> {
  theme.set(mode);
  applyTheme(mode);
  try {
    const { SetTheme } = await import('../../../wailsjs/go/main/App');
    await SetTheme(mode);
  } catch (err) {
    console.error('Failed to save theme:', err);
  }
}

export function applyTheme(mode: ThemeMode): void {
  const html = document.documentElement;

  if (mode === 'system') {
    const prefersLight = !window.matchMedia('(prefers-color-scheme: dark)').matches;
    html.classList.toggle('light', prefersLight);
  } else {
    html.classList.toggle('light', mode === 'light');
  }
}

if (typeof window !== 'undefined') {
  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
    const currentTheme = get(theme);
    if (currentTheme === 'system') {
      applyTheme('system');
    }
  });
}
