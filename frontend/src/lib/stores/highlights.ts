import { writable, get } from 'svelte/store';
import type { HighlightData } from '../highlights/renderer';

const cache = new Map<string, HighlightData[]>();
export const highlightsForFile = writable<HighlightData[]>([]);
export const lastUsedColor = writable<string>('yellow');

export async function loadHighlightsForFile(filePath: string): Promise<void> {
  if (cache.has(filePath)) {
    highlightsForFile.set(cache.get(filePath)!);
    return;
  }
  try {
    const App: any = await import('../../../wailsjs/go/main/App');
    const highlights = await App.GetHighlights(filePath);
    const data = highlights || [];
    cache.set(filePath, data);
    highlightsForFile.set(data);
  } catch {
    highlightsForFile.set([]);
  }
}

export async function addHighlight(hl: HighlightData): Promise<void> {
  try {
    const App: any = await import('../../../wailsjs/go/main/App');
    await App.AddHighlight(hl);
    cache.delete(hl.filePath);
    await loadHighlightsForFile(hl.filePath);
  } catch (err) {
    console.error('Failed to add highlight:', err);
  }
}

export async function removeHighlight(filePath: string, highlightId: string): Promise<void> {
  try {
    const App: any = await import('../../../wailsjs/go/main/App');
    await App.RemoveHighlight(filePath, highlightId);
    cache.delete(filePath);
    await loadHighlightsForFile(filePath);
  } catch (err) {
    console.error('Failed to remove highlight:', err);
  }
}
