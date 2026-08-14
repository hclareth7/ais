<script lang="ts">
  import { onMount } from 'svelte';
  import Sidebar from './lib/components/Sidebar.svelte';
  import TabBar from './lib/components/TabBar.svelte';
  import MarkdownViewer from './lib/components/MarkdownViewer.svelte';
  import TocPanel from './lib/components/TocPanel.svelte';
  import ControlStrip from './lib/components/ControlStrip.svelte';
  import SettingsPanel from './lib/components/SettingsPanel.svelte';
  import CommandPalette from './lib/components/CommandPalette.svelte';
  import { loadFileTree, rootPath, type FileNode } from './lib/stores/files';
  import { openTab, closeTab, nextTab, prevTab, activeTabId, activeTab, updateTabContent, tabs } from './lib/stores/tabs';
  import { loadSettings } from './lib/stores/settings';
  import { readFile } from './lib/stores/files';
  import { zoomLevel, readingWidth, focusMode, zoomIn, zoomOut, resetZoom, toggleFocusMode, toggleCommandPalette, commandPaletteOpen, changeOpacity, tocVisible, toggleToc, settingsOpen } from './lib/stores/ui';
  import { get } from 'svelte/store';

  function windowMinimise() {
    import('../wailsjs/runtime/runtime').then(r => r.WindowMinimise()).catch(() => {});
  }

  function windowToggleMaximise() {
    import('../wailsjs/runtime/runtime').then(r => r.WindowToggleMaximise()).catch(() => {});
  }

  function windowQuit() {
    import('../wailsjs/runtime/runtime').then(r => r.Quit()).catch(() => {});
  }

  let navVisible = $state(false);

  function handleFileClick(node: FileNode) {
    if (!node.isDir) {
      openTab(node.path, node.name);
      navVisible = false;
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

      if (e.ctrlKey && e.shiftKey && (e.key === '=' || e.key === '+')) {
        e.preventDefault();
        changeOpacity(5);
        return;
      }

      if (e.ctrlKey && e.shiftKey && e.key === '-') {
        e.preventDefault();
        changeOpacity(-5);
        return;
      }

      if (e.ctrlKey && e.shiftKey && (e.key === 'O' || e.key === 'o')) {
        e.preventDefault();
        toggleToc();
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
          settingsOpen.set(false);
        }
        return;
      }
    }

    document.addEventListener('keydown', handleKeydown);

    function handleClickOutsideSettings(e: MouseEvent) {
      const target = e.target as HTMLElement;
      if (!target.closest('.settings') && !target.closest('.cb[title="Settings"]')) {
        settingsOpen.set(false);
      }
    }

    document.addEventListener('click', handleClickOutsideSettings);

    return () => {
      document.removeEventListener('keydown', handleKeydown);
      document.removeEventListener('click', handleClickOutsideSettings);
    };
  });
</script>

<div id="srAnnounce" class="sr-only" aria-live="polite" aria-atomic="true" role="status"></div>

<div
  class="reader"
  class:focus={$focusMode}
  role="application"
  aria-label="ais document reader"
>
  <div class="titlebar" style="--wails-draggable: drag;">
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
    <div class="tb-right" style="--wails-draggable: no-drag;">
      <button class="wb" title="Minimize" aria-label="Minimize window" onclick={windowMinimise}>
        <svg viewBox="0 0 14 14"><line x1="2" y1="7" x2="12" y2="7"/></svg>
      </button>
      <button class="wb" title="Maximize" aria-label="Maximize window" onclick={windowToggleMaximise}>
        <svg viewBox="0 0 14 14"><rect x="2" y="2" width="10" height="10" rx="1.5"/></svg>
      </button>
      <button class="wb cl" title="Close" aria-label="Close window" onclick={windowQuit}>
        <svg viewBox="0 0 14 14"><line x1="3" y1="3" x2="11" y2="11"/><line x1="11" y1="3" x2="3" y2="11"/></svg>
      </button>
    </div>
  </div>

  <TabBar />

  <div class="content">
    <div class="nav-trigger" role="presentation"></div>
    <Sidebar onFileClick={handleFileClick} visible={navVisible} />

    <MarkdownViewer zoomLevel={$zoomLevel} readingWidth={$readingWidth} />

    <div class="toc-trigger" role="presentation"></div>
    <TocPanel />

    <div class="edge edge-left"></div>
    <div class="edge edge-right"></div>

    <div class="bottom-trigger" role="presentation"></div>
    <ControlStrip />
    <SettingsPanel />
  </div>
</div>

<CommandPalette />

<style>
  .reader.focus {
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
