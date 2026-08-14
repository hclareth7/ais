import { writable, get } from 'svelte/store';

export type ThemeMode = 'light' | 'dark' | 'system';

export const theme = writable<ThemeMode>('system');

export async function loadSettings(): Promise<void> {
  try {
    const { GetTheme } = await import('../../../wailsjs/go/main/App');
    const t = await GetTheme();
    if (t === 'light' || t === 'dark' || t === 'system') {
      theme.set(t);
    }
    applyTheme(get(theme));
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
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    html.classList.toggle('dark', prefersDark);
  } else {
    html.classList.toggle('dark', mode === 'dark');
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
