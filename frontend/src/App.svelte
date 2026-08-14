<script lang="ts">
  import { onMount } from 'svelte';
  import Sidebar from './lib/components/Sidebar.svelte';
  import TabBar from './lib/components/TabBar.svelte';
  import MarkdownViewer from './lib/components/MarkdownViewer.svelte';
  import ControlStrip from './lib/components/ControlStrip.svelte';
  import CommandPalette from './lib/components/CommandPalette.svelte';
  import { loadFileTree, rootPath, type FileNode } from './lib/stores/files';
  import { openTab, closeTab, nextTab, prevTab, activeTabId, activeTab, updateTabContent, tabs } from './lib/stores/tabs';
  import { loadSettings } from './lib/stores/settings';
  import { readFile } from './lib/stores/files';
  import { zoomLevel, readingWidth, focusMode, zoomIn, zoomOut, resetZoom, toggleFocusMode, toggleCommandPalette, commandPaletteOpen } from './lib/stores/ui';
  import { get } from 'svelte/store';

  let navVisible = $state(false);

  function handleFileClick(node: FileNode) {
    if (!node.isDir) {
      openTab(node.path, node.name);
    }
  }

  function getFolderName(path: string): string {
    if (!path) return 'ais';
    const parts = path.split('/');
    return parts[parts.length - 1] || 'ais';
  }

  onMount(async () => {
    await loadSettings();
    await loadFileTree();

    const starsEl = document.getElementById('stars');
    if (starsEl) {
      for (let i = 0; i < 80; i++) {
        const star = document.createElement('div');
        star.className = 'star';
        star.style.left = `${(i * 17 + 31) % 100}%`;
        star.style.top = `${(i * 23 + 7) % 100}%`;
        star.style.opacity = `${0.2 + ((i * 13) % 5) * 0.1}`;
        starsEl.appendChild(star);
      }
    }

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
      if (e.ctrlKey && e.key === 'k') {
        e.preventDefault();
        toggleCommandPalette();
        return;
      }

      if (get(commandPaletteOpen)) return;

      if (e.ctrlKey && e.key === 'b') {
        e.preventDefault();
        navVisible = !navVisible;
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

      if (e.ctrlKey && (e.key === '=' || e.key === '+')) {
        e.preventDefault();
        zoomIn();
        return;
      }

      if (e.ctrlKey && e.key === '-') {
        e.preventDefault();
        zoomOut();
        return;
      }

      if (e.ctrlKey && e.key === '0') {
        e.preventDefault();
        resetZoom();
        return;
      }

      if (e.key === 'F11') {
        e.preventDefault();
        toggleFocusMode();
        return;
      }

      if (e.key === 'Escape') {
        if (get(focusMode)) {
          toggleFocusMode();
        } else {
          navVisible = false;
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

<div id="srAnnounce" class="sr-only" aria-live="polite" aria-atomic="true" role="status"></div>

<div class="background" aria-hidden="true">
  <div class="bg-sky"></div>
  <div class="bg-mountains">
    <div class="mountain m1"></div>
    <div class="mountain m2"></div>
    <div class="mountain m3"></div>
  </div>
  <div class="bg-fog"></div>
  <div class="bg-stars" id="stars"></div>
</div>

<div
  class="reader"
  class:focus={$focusMode}
  style="width: 82vw; height: 90vh;"
  role="application"
  aria-label="ais document reader"
>
  <div class="titlebar">
    <div class="tb-left">
      <span class="logo">ais</span>
      <div class="breadcrumb">
        <span>{getFolderName($rootPath)}</span>
        {#if $activeTab}
          <span class="sep">/</span>
          <span>{$activeTab.name}</span>
        {/if}
      </div>
    </div>
  </div>

  <TabBar />

  <div class="content">
    <div class="nav-trigger" role="presentation"></div>
    <Sidebar onFileClick={handleFileClick} visible={navVisible} />

    <MarkdownViewer zoomLevel={$zoomLevel} readingWidth={$readingWidth} />

    <div class="toc-trigger" role="presentation"></div>

    <div class="edge edge-left"></div>
    <div class="edge edge-right"></div>

    <div class="bottom-trigger" role="presentation"></div>
    <ControlStrip />
  </div>
</div>

<CommandPalette />

<style>
  .reader.focus {
    width: 100vw !important;
    height: 100vh !important;
    border-radius: 0;
    border-color: transparent;
  }

  .reader.focus :global(.titlebar) {
    opacity: 0;
    height: 0;
    padding: 0;
    overflow: hidden;
  }

  .reader.focus :global(.tabbar) {
    opacity: 0;
    height: 0;
    padding: 0;
  }
</style>
