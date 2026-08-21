import { writable } from 'svelte/store';

export interface SearchScrollTarget {
  query: string;
  filePath: string;
  lineNumber: number;
}

export const searchScrollTarget = writable<SearchScrollTarget | null>(null);
export const inFileSearchOpen = writable(false);
