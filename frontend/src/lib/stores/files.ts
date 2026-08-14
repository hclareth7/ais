import { writable } from 'svelte/store';

export interface FileNode {
  name: string;
  path: string;
  isDir: boolean;
  children?: FileNode[];
}

export const fileTree = writable<FileNode | null>(null);
export const rootPath = writable<string>('');

export async function loadFileTree(): Promise<void> {
  try {
    const { GetFileTree, GetRootPath } = await import('../../../wailsjs/go/main/App');
    const tree = await GetFileTree();
    fileTree.set(tree);
    const path = await GetRootPath();
    rootPath.set(path);
  } catch (err) {
    console.error('Failed to load file tree:', err);
  }
}

export async function readFile(path: string): Promise<string> {
  try {
    const { ReadFile } = await import('../../../wailsjs/go/main/App');
    return await ReadFile(path);
  } catch (err) {
    console.error('Failed to read file:', err);
    return `Error reading file: ${err}`;
  }
}
