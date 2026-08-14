<script lang="ts">
  import { onMount } from 'svelte';
  import Sidebar from './lib/components/Sidebar.svelte';
  import TabBar from './lib/components/TabBar.svelte';
  import MarkdownViewer from './lib/components/MarkdownViewer.svelte';
  import { loadFileTree, type FileNode } from './lib/stores/files';
  import { openTab, closeTab, nextTab, prevTab, activeTabId, updateTabContent, tabs } from './lib/stores/tabs';
  import { loadSettings } from './lib/stores/settings';
  import { readFile } from './lib/stores/files';
  import { get } from 'svelte/store';

  let sidebarVisible = $state(true);

  function handleFileClick(node: FileNode) {
    if (!node.isDir) {
      openTab(node.path, node.name);
    }
  }

  onMount(async () => {
    await loadSettings();
    await loadFileTree();

    try {
      const { EventsOn } = await import('../wailsjs/runtime/runtime');
      EventsOn('file:changed', async (changedPath: string) => {
        const currentTabs = get(tabs);
        for (const tab of currentTabs) {
          if (changedPath === tab.path) {
            const content = await readFile(tab.path);
            updateTabContent(tab.path, content);
          }
        }
      });
    } catch (err) {
      console.warn('Wails runtime not available (dev mode):', err);
    }

    function handleKeydown(e: KeyboardEvent) {
      if (e.ctrlKey && e.key === 'b') {
        e.preventDefault();
        sidebarVisible = !sidebarVisible;
        return;
      }

      if (e.ctrlKey && e.key === 'w') {
        e.preventDefault();
        const id = get(activeTabId);
        if (id) closeTab(id);
        return;
      }

      if (e.ctrlKey && (e.key === 'Tab' || e.key === 'PageDown' || e.key === 'PageUp')) {
        e.preventDefault();
        if (e.shiftKey || e.key === 'PageUp') {
          prevTab();
        } else {
          nextTab();
        }
        return;
      }
    }

    document.addEventListener('keydown', handleKeydown);

    return () => {
      document.removeEventListener('keydown', handleKeydown);
    };
  });
</script>

<div
  class="app-shell"
  style:grid-template-columns={sidebarVisible ? 'var(--sidebar-width) 1fr' : '0 1fr'}
>
  {#if sidebarVisible}
    <Sidebar onFileClick={handleFileClick} />
  {/if}
  <TabBar />
  <MarkdownViewer />
</div>

<style>
  .app-shell {
    height: 100vh;
    display: grid;
    grid-template-rows: var(--tab-height) 1fr;
    grid-template-columns: var(--sidebar-width) 1fr;
    grid-template-areas:
      "sidebar tabs"
      "sidebar viewer";
    transition: grid-template-columns 200ms ease-out;
  }
</style>
