<script lang="ts">
  import { fileTree, rootPath } from '../stores/files';
  import type { FileNode } from '../stores/files';
  import FileTree from './FileTree.svelte';
  import ThemeToggle from './ThemeToggle.svelte';

  let { onFileClick, visible = false }: {
    onFileClick: (node: FileNode) => void;
    visible?: boolean;
  } = $props();

  let searchQuery = $state('');
</script>

<nav class="nav-panel" class:vis={visible} aria-label="File navigation">
  <div class="nav-search">
    <svg class="search-icon" viewBox="0 0 24 24">
      <circle cx="11" cy="11" r="8" /><line x1="21" y1="21" x2="16.65" y2="16.65" />
    </svg>
    <input
      type="text"
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

  <div class="nav-label">Files</div>

  <div class="tree-container">
    <FileTree
      tree={$fileTree}
      {onFileClick}
      {searchQuery}
    />
  </div>

  <div class="nav-bottom">
    <ThemeToggle />
  </div>
</nav>

<style>
  .nav-search {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 7px 10px;
    background: var(--hover-bg);
    border-radius: 10px;
    margin-bottom: 6px;
    position: relative;
  }

  .search-icon {
    width: 16px;
    height: 16px;
    stroke: var(--text-tertiary);
    stroke-width: 1.5;
    stroke-linecap: round;
    fill: none;
    flex-shrink: 0;
  }

  .nav-search input {
    background: 0;
    border: 0;
    outline: 0;
    color: var(--text-primary);
    font-size: 13px;
    font-family: inherit;
    width: 100%;
  }

  .nav-search input::placeholder {
    color: var(--text-tertiary);
  }

  .search-clear {
    background: none;
    border: none;
    color: var(--text-tertiary);
    cursor: pointer;
    font-size: 16px;
    padding: 0 4px;
    line-height: 1;
  }

  .search-clear:hover {
    color: var(--text-primary);
  }

  .nav-label {
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--text-tertiary);
    padding: 14px 10px 6px;
  }

  .tree-container {
    flex: 1;
    overflow-y: auto;
    overflow-x: hidden;
  }

  .nav-bottom {
    margin-top: auto;
    padding-top: 10px;
    border-top: 1px solid var(--border);
  }
</style>
