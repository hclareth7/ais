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
      class="nv"
      class:is-dir={node.isDir}
      class:on={isActive}
      onclick={toggle}
      role="treeitem"
      aria-selected={isActive}
      aria-expanded={node.isDir ? expanded : undefined}
      title={node.path}
    >
      {#if node.isDir}
        <svg class="node-icon" viewBox="0 0 24 24">
          {#if expanded}
            <polyline points="6 9 12 15 18 9" />
          {:else}
            <polyline points="9 6 15 12 9 18" />
          {/if}
        </svg>
      {:else}
        <svg class="node-icon" viewBox="0 0 24 24">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
          <polyline points="14 2 14 8 20 8" />
        </svg>
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

  .nv {
    display: flex;
    align-items: center;
    gap: 9px;
    width: 100%;
    padding: 7px 10px;
    font-size: 14px;
    color: var(--text-secondary);
    border-radius: 10px;
    cursor: pointer;
    transition: background 0.12s, color 0.12s;
    border: none;
    background: none;
    text-align: left;
    font-family: inherit;
  }

  .nv:hover {
    background: var(--hover-bg);
    color: var(--text-primary);
  }

  .nv:hover :global(svg) {
    stroke: var(--text-primary);
  }

  .nv.on {
    background: var(--active-bg);
    color: var(--accent-text);
  }

  .nv.on :global(svg) {
    stroke: var(--icon-active);
  }

  .nv.is-dir {
    font-weight: 500;
  }

  .nv:focus-visible {
    outline: 2px solid var(--border-focus);
    outline-offset: 4px;
  }

  .node-icon {
    width: 16px;
    height: 16px;
    stroke: var(--icon-stroke);
    stroke-width: 1.5;
    stroke-linecap: round;
    stroke-linejoin: round;
    fill: none;
    flex-shrink: 0;
    transition: stroke 0.12s;
  }

  .node-name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
