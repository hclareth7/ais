<script lang="ts">
  import type { FileNode } from '../stores/files';
  import FileTreeNode from './FileTreeNode.svelte';

  let { tree, onFileClick, searchQuery = '' }: {
    tree: FileNode | null;
    onFileClick: (node: FileNode) => void;
    searchQuery?: string;
  } = $props();
</script>

<div class="file-tree" role="tree" aria-label="File tree">
  {#if tree && tree.children}
    {#each tree.children as child (child.path)}
      <FileTreeNode
        node={child}
        depth={0}
        {onFileClick}
        {searchQuery}
      />
    {/each}
  {:else if tree && !tree.children}
    <div class="empty-tree">
      No markdown files found
    </div>
  {/if}
</div>

<style>
  .file-tree {
    padding: 4px;
  }

  .empty-tree {
    padding: 16px;
    color: var(--text-muted);
    font-size: 14px;
    text-align: center;
  }
</style>
