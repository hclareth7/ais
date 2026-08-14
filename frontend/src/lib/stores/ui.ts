import { writable } from 'svelte/store';

const ZOOM_LEVELS = [50, 75, 90, 100, 110, 125, 150, 175, 200];

export const zoomLevel = writable(100);
export const readingWidth = writable(720);
export const focusMode = writable(false);
export const commandPaletteOpen = writable(false);

export function zoomIn(): void {
  zoomLevel.update(z => {
    const idx = ZOOM_LEVELS.indexOf(z);
    return idx < ZOOM_LEVELS.length - 1 ? ZOOM_LEVELS[idx + 1] : z;
  });
}

export function zoomOut(): void {
  zoomLevel.update(z => {
    const idx = ZOOM_LEVELS.indexOf(z);
    return idx > 0 ? ZOOM_LEVELS[idx - 1] : z;
  });
}

export function resetZoom(): void {
  zoomLevel.set(100);
}

export function toggleFocusMode(): void {
  focusMode.update(f => !f);
}

export function toggleCommandPalette(): void {
  commandPaletteOpen.update(o => !o);
}
