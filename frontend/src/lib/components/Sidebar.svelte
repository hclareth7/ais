<script lang="ts">
  import { fileTree, rootPath } from '../stores/files';
  import type { FileNode } from '../stores/files';
  import FileTree from './FileTree.svelte';
  import ThemeToggle from './ThemeToggle.svelte';

  let { onFileClick }: {
    onFileClick: (node: FileNode) => void;
  } = $props();

  let searchQuery = $state('');

  function getFolderName(path: string): string {
    if (!path) return 'ais';
    const parts = path.split('/');
    return parts[parts.length - 1] || 'ais';
  }
</script>

<aside class="sidebar" aria-label="File explorer">
  <div class="sidebar-header">
    <span class="folder-name" title={$rootPath}>
      {getFolderName($rootPath)}
    </span>
  </div>

  <div class="search-container">
    <input
      type="text"
      class="search-input"
      placeholder="Search files..."
      bind:value={searchQuery}
      aria-label="Search files"
    />
    {#if searchQuery}
      <button
        class="search-clear"
        onclick={() => searchQuery = ''}
        aria-label="Clear search"
      >&times;</button>
    {/if}
  </div>

  <div class="tree-container">
    <FileTree
      tree={$fileTree}
      {onFileClick}
      {searchQuery}
    />
  </div>

  <div class="sidebar-footer">
    <ThemeToggle />
  </div>
</aside>

<style>
  .sidebar {
    grid-area: sidebar;
    background: var(--bg-secondary);
    border-right: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    height: 100vh;
  }

  .sidebar-header {
    padding: 12px 16px;
    border-bottom: 1px solid var(--border);
    flex-shrink: 0;
  }

  .folder-name {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    display: block;
  }

  .search-container {
    padding: 8px 12px;
    flex-shrink: 0;
    position: relative;
  }

  .search-input {
    width: 100%;
    padding: 6px 28px 6px 10px;
    border: 1px solid var(--border);
    border-radius: 6px;
    background: var(--bg-primary);
    color: var(--text-primary);
    font-size: 13px;
    outline: none;
    box-sizing: border-box;
  }

  .search-input::placeholder {
    color: var(--text-muted);
  }

  .search-input:focus {
    border-color: var(--accent);
  }

  .search-clear {
    position: absolute;
    right: 18px;
    top: 50%;
    transform: translateY(-50%);
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    font-size: 16px;
    padding: 0 4px;
    line-height: 1;
  }

  .search-clear:hover {
    color: var(--text-primary);
  }

  .tree-container {
    flex: 1;
    overflow-y: auto;
    overflow-x: hidden;
  }

  .sidebar-footer {
    padding: 8px 12px;
    border-top: 1px solid var(--border);
    flex-shrink: 0;
  }
</style>
