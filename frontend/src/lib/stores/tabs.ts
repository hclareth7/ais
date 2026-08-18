import { writable, derived, get } from 'svelte/store';
import { readFile } from './files';

export interface Tab {
  id: string;
  name: string;
  path: string;
  content: string;
  scrollPos: number;
  type: 'file' | 'stream';
  streamActive?: boolean;
}

export const tabs = writable<Tab[]>([]);
export const activeTabId = writable<string | null>(null);

export const activeTab = derived(
  [tabs, activeTabId],
  ([$tabs, $activeTabId]) => $tabs.find(t => t.id === $activeTabId) ?? null
);

export async function openTab(path: string, name: string): Promise<void> {
  const currentTabs = get(tabs);
  const existing = currentTabs.find(t => t.path === path);

  if (existing) {
    activeTabId.set(existing.id);
    return;
  }

  const content = await readFile(path);
  const id = path;

  const newTab: Tab = { id, name, path, content, scrollPos: 0, type: 'file' };

  tabs.update(t => {
    if (t.length >= 20) {
      t = t.slice(1);
    }
    return [...t, newTab];
  });
  activeTabId.set(id);
}

export function closeTab(id: string): void {
  tabs.update(t => {
    const idx = t.findIndex(tab => tab.id === id);
    if (idx === -1) return t;

    const newTabs = t.filter(tab => tab.id !== id);

    const currentActive = get(activeTabId);
    if (currentActive === id) {
      if (newTabs.length === 0) {
        activeTabId.set(null);
      } else if (idx < newTabs.length) {
        activeTabId.set(newTabs[idx].id);
      } else {
        activeTabId.set(newTabs[newTabs.length - 1].id);
      }
    }

    return newTabs;
  });
}

export function nextTab(): void {
  const currentTabs = get(tabs);
  const currentId = get(activeTabId);
  if (currentTabs.length <= 1) return;

  const idx = currentTabs.findIndex(t => t.id === currentId);
  const nextIdx = (idx + 1) % currentTabs.length;
  activeTabId.set(currentTabs[nextIdx].id);
}

export function prevTab(): void {
  const currentTabs = get(tabs);
  const currentId = get(activeTabId);
  if (currentTabs.length <= 1) return;

  const idx = currentTabs.findIndex(t => t.id === currentId);
  const prevIdx = (idx - 1 + currentTabs.length) % currentTabs.length;
  activeTabId.set(currentTabs[prevIdx].id);
}

export function updateTabContent(path: string, content: string): void {
  tabs.update(t => t.map(tab =>
    tab.path === path ? { ...tab, content } : tab
  ));
}

export function saveScrollPos(id: string, pos: number): void {
  tabs.update(t => t.map(tab =>
    tab.id === id ? { ...tab, scrollPos: pos } : tab
  ));
}

// Opens a new stream tab for an AI response and returns its ID.
export function openStreamTab(prompt: string): string {
  const id = `stream-${Date.now()}`;
  const name = `AI: ${prompt.slice(0, 30)}${prompt.length > 30 ? '...' : ''}`;
  const newTab: Tab = {
    id,
    name,
    path: '',
    content: '',
    scrollPos: 0,
    type: 'stream',
    streamActive: true,
  };

  tabs.update(t => {
    if (t.length >= 20) {
      t = t.slice(1);
    }
    return [...t, newTab];
  });
  activeTabId.set(id);
  return id;
}

// Updates the streamActive flag on a tab.
export function setStreamActive(id: string, active: boolean): void {
  tabs.update(t => t.map(tab =>
    tab.id === id ? { ...tab, streamActive: active } : tab
  ));
}

// Updates tab content by tab ID (used for stream tabs where path is empty).
export function updateTabContentById(id: string, content: string): void {
  tabs.update(t => t.map(tab =>
    tab.id === id ? { ...tab, content } : tab
  ));
}
