import { writable, get } from 'svelte/store';

const ZOOM_LEVELS = [50, 75, 90, 100, 110, 125, 150, 175, 200];
const RADIUS_OPTIONS = [20, 28, 36, 48];

export const zoomLevel = writable(100);
export const readingWidth = writable(1000);
export const focusMode = writable(false);
export const commandPaletteOpen = writable(false);
export const opacity = writable(100);
export const readerRadius = writable(20);
export const backgroundMode = writable<'gradient' | 'solid' | 'frost'>('gradient');
export const settingsOpen = writable(false);
export const tocVisible = writable(false);

export { RADIUS_OPTIONS };

let persistTimer: ReturnType<typeof setTimeout> | null = null;

function persistUISettings(): void {
  if (persistTimer) clearTimeout(persistTimer);
  persistTimer = setTimeout(async () => {
    try {
      const App: any = await import('../../../wailsjs/go/main/App');
      await App.SaveUISettings({
        zoomLevel: get(zoomLevel),
        opacity: get(opacity),
        readingWidth: get(readingWidth),
        readerRadius: get(readerRadius),
        backgroundMode: get(backgroundMode),
      });
    } catch (err) {
      console.error('Failed to save UI settings:', err);
    }
  }, 500);
}

export function zoomIn(): void {
  zoomLevel.update(z => {
    const idx = ZOOM_LEVELS.indexOf(z);
    return idx < ZOOM_LEVELS.length - 1 ? ZOOM_LEVELS[idx + 1] : z;
  });
  persistUISettings();
}

export function zoomOut(): void {
  zoomLevel.update(z => {
    const idx = ZOOM_LEVELS.indexOf(z);
    return idx > 0 ? ZOOM_LEVELS[idx - 1] : z;
  });
  persistUISettings();
}

export function resetZoom(): void {
  zoomLevel.set(100);
  persistUISettings();
}

export function toggleFocusMode(): void {
  focusMode.update(f => !f);
}

export function toggleCommandPalette(): void {
  commandPaletteOpen.update(o => !o);
}

export function changeOpacity(delta: number): void {
  opacity.update(o => Math.max(40, Math.min(100, o + delta)));
  applyOpacity();
  persistUISettings();
}

export function setOpacity(value: number): void {
  opacity.set(Math.max(40, Math.min(100, value)));
  applyOpacity();
  persistUISettings();
}

function applyOpacity(): void {
  const o = get(opacity);
  document.documentElement.style.setProperty('--surface-opacity', String(o / 100));
}

export function setReaderRadius(px: number): void {
  readerRadius.set(px);
  document.documentElement.style.setProperty('--reader-radius', `${px}px`);
  if (document.documentElement.classList.contains('macos')) {
    import('../../../wailsjs/go/main/App').then(m => m.SetCornerRadius(px)).catch(() => {});
  }
  persistUISettings();
}

export function setBackgroundMode(mode: 'gradient' | 'solid' | 'frost'): void {
  backgroundMode.set(mode);
  persistUISettings();
}

export function setReadingWidth(value: number): void {
  readingWidth.set(value);
  persistUISettings();
}

export function toggleSettings(): void {
  settingsOpen.update(s => !s);
}

export function toggleToc(): void {
  tocVisible.update(t => !t);
}
