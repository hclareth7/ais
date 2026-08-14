<script lang="ts">
  import type { FileNode } from '../stores/files';
  import { activeTabId } from '../stores/tabs';
  import FileTreeNode from './FileTreeNode.svelte';

  let { node, depth = 0, onFileClick, searchQuery = '' }: {
    node: FileNode;
    depth?: number;
    onFileClick: (node: FileNode) => void;
    searchQuery?: string;
  } = $props();

  // Intentionally capture initial depth — tree nodes don't change depth after mount
  let expanded = $state((() => depth < 2)());

  function toggle() {
    if (node.isDir) {
      expanded = !expanded;
    } else {
      onFileClick(node);
    }
  }

  function matchesSearch(n: FileNode, query: string): boolean {
    if (!query) return true;
    const lowerQuery = query.toLowerCase();
    if (n.name.toLowerCase().includes(lowerQuery)) return true;
    if (n.isDir && n.children) {
      return n.children.some(child => matchesSearch(child, query));
    }
    return false;
  }

  let filteredChildren = $derived(
    node.children?.filter(child => matchesSearch(child, searchQuery)) ?? []
  );

  let isActive = $derived($activeTabId === node.path);
</script>

{#if matchesSearch(node, searchQuery)}
  <div class="tree-node" style="padding-left: {depth * 16}px">
    <button
      class="tree-item"
      class:is-dir={node.isDir}
      class:is-active={isActive}
      onclick={toggle}
      role="treeitem"
      aria-selected={isActive}
      aria-expanded={node.isDir ? expanded : undefined}
      title={node.path}
    >
      {#if node.isDir}
        <span class="chevron">{expanded ? '▼' : '▶'}</span>
      {:else}
        <span class="file-icon">•</span>
      {/if}
      <span class="node-name">{node.name}</span>
    </button>
  </div>

  {#if node.isDir && expanded}
    {#each filteredChildren as child (child.path)}
      <FileTreeNode
        node={child}
        depth={depth + 1}
        {onFileClick}
        {searchQuery}
      />
    {/each}
  {/if}
{/if}

<style>
  .tree-node {
    user-select: none;
  }

  .tree-item {
    display: flex;
    align-items: center;
    gap: 4px;
    width: 100%;
    padding: 4px 8px;
    border: none;
    background: none;
    color: var(--text-secondary);
    font-size: 14px;
    line-height: 20px;
    cursor: pointer;
    text-align: left;
    border-radius: 4px;
  }

  .tree-item:hover {
    background: var(--bg-inset);
    color: var(--text-primary);
  }

  .tree-item.is-active {
    background: var(--accent-subtle);
    color: var(--text-primary);
  }

  .tree-item.is-dir {
    font-weight: 500;
  }

  .tree-item:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: -2px;
  }

  .chevron {
    font-size: 10px;
    width: 16px;
    text-align: center;
    flex-shrink: 0;
    transition: transform 150ms ease;
    color: var(--text-muted);
  }

  .file-icon {
    font-size: 10px;
    width: 16px;
    text-align: center;
    flex-shrink: 0;
    color: var(--text-muted);
  }

  .node-name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
