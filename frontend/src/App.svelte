<script lang="ts">
  import { onMount } from 'svelte';
  import Sidebar from './lib/components/Sidebar.svelte';
  import TabBar from './lib/components/TabBar.svelte';
  import MarkdownViewer from './lib/components/MarkdownViewer.svelte';
  import TocPanel from './lib/components/TocPanel.svelte';
  import ControlStrip from './lib/components/ControlStrip.svelte';
  import SettingsPanel from './lib/components/SettingsPanel.svelte';
  import CommandPalette from './lib/components/CommandPalette.svelte';
  import MruOverlay from './lib/components/MruOverlay.svelte';
  import { loadFileTree, rootPath, type FileNode } from './lib/stores/files';
  import { openTab, closeTab, nextTab, prevTab, activeTabId, activeTab, updateTabContent, updateTabContentById, setStreamActive, openStreamTab, tabs, mruOrder } from './lib/stores/tabs';
  import { loadSettings } from './lib/stores/settings';
  import { readFile } from './lib/stores/files';
  import { zoomLevel, readingWidth, focusMode, zoomIn, zoomOut, resetZoom, toggleFocusMode, toggleCommandPalette, commandPaletteOpen, changeOpacity, tocVisible, toggleToc, settingsOpen, commandPaletteCategory, editMode, toggleEditMode } from './lib/stores/ui';
  import { inFileSearchOpen } from './lib/stores/search';
  import { activeStream, appendStreamContent, completeStream, cancelStreamState, setStreamError, streamState, startStreamSession } from './lib/stores/stream';
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
  let mruCycling = $state(false);
  let mruShowOverlay = $state(false);
  let mruCycleIndex = $state(0);
  let mruTabList: Array<{id: string; name: string}> = $state([]);
  let mruOverlayTimer: ReturnType<typeof setTimeout> | null = null;
  let pipeTabId: string | null = null;
  let langToast = $state('');
  let langToastTimer: ReturnType<typeof setTimeout> | null = null;

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
    if (navigator.platform.startsWith('Mac')) {
      document.documentElement.classList.add('macos');
    }

    await loadSettings();
    await loadFileTree();

    try {
      const App: any = await import('../wailsjs/go/main/App');
      const initialFile = await App.GetInitialFile();
      if (initialFile) {
        openTab(initialFile, initialFile.split('/').pop() ?? initialFile);
      }
    } catch {}

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
      EventsOn('file:created', async (createdPath: string) => {
        const filename = createdPath.split('/').pop() ?? createdPath;
        openTab(createdPath, filename);
        await loadFileTree();
      });

      EventsOn('llm:chunk', (chunk: { text: string; done: boolean; totalTokens: number }) => {
        const stream = get(activeStream);
        if (!stream) return;
        appendStreamContent(chunk.text);
        const updated = get(activeStream);
        if (updated) {
          updateTabContentById(updated.tabId, updated.content);
        }
      });

      EventsOn('llm:done', (chunk: { text: string; done: boolean; totalTokens: number }) => {
        const stream = get(activeStream);
        if (!stream) return;
        if (chunk.text) {
          appendStreamContent(chunk.text);
        }
        completeStream(chunk.totalTokens);
        const updated = get(activeStream);
        if (updated) {
          updateTabContentById(updated.tabId, updated.content);
          setStreamActive(updated.tabId, false);
        }
      });

      EventsOn('llm:error', (error: { code: string; message: string }) => {
        setStreamError({ code: error.code as any, message: error.message });
        const stream = get(activeStream);
        if (stream) {
          setStreamActive(stream.tabId, false);
        }
      });

      EventsOn('pipe:data', (text: string) => {
        if (!pipeTabId) {
          pipeTabId = openStreamTab('Pipe Input');
          startStreamSession(pipeTabId);
        }
        appendStreamContent(text);
        const stream = get(activeStream);
        if (stream) {
          updateTabContentById(stream.tabId, stream.content);
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

      if (e.ctrlKey && e.shiftKey && (e.key === 'F' || e.key === 'f')) {
        e.preventDefault();
        commandPaletteCategory.set('search');
        commandPaletteOpen.set(true);
        return;
      }

      if (e.ctrlKey && (e.key === 'f' || e.key === 'F') && !e.shiftKey) {
        e.preventDefault();
        inFileSearchOpen.set(true);
        return;
      }

      if (e.ctrlKey && e.key === ' ') {
        e.preventDefault();
        import('../wailsjs/go/main/App').then(async (App) => {
          const cfg = await App.GetConfig();
          const langs = cfg.translationLanguages;
          if (!langs || langs.length < 2) return;
          const nextIdx = (cfg.translationDefaultIndex + 1) % langs.length;
          cfg.translationDefaultIndex = nextIdx;
          await App.UpdateConfig(cfg);
          if (langToastTimer) clearTimeout(langToastTimer);
          langToast = langs[nextIdx].toUpperCase();
          langToastTimer = setTimeout(() => { langToast = ''; }, 1500);
        }).catch(() => {});
        return;
      }

      if (e.ctrlKey && (e.key === 'o' || e.key === 'O') && !e.shiftKey) {
        e.preventDefault();
        import('../wailsjs/go/main/App').then(async (App) => {
          const relPath = await App.OpenFile();
          if (relPath) {
            await loadFileTree();
            const filename = relPath.split('/').pop() ?? relPath;
            openTab(relPath, filename);
          }
        }).catch(() => {});
        return;
      }

      if (e.ctrlKey && (e.key === 'e' || e.key === 'E') && !e.shiftKey) {
        e.preventDefault();
        const tab = get(activeTab);
        if (tab && !tab.isStream) {
          toggleEditMode();
        }
        return;
      }

      if (e.ctrlKey && e.shiftKey && (e.key === 'C' || e.key === 'c')) {
        e.preventDefault();
        const tab = get(activeTab);
        if (tab) {
          import('../wailsjs/runtime/runtime')
            .then(r => r.ClipboardSetText(tab.content))
            .catch(() => navigator.clipboard.writeText(tab.content));
        }
        return;
      }

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

      if (e.ctrlKey && (e.key === 'Tab' || e.code === 'Tab')) {
        e.preventDefault();
        if (!mruCycling) {
          const order = get(mruOrder);
          const currentTabs = get(tabs);
          mruTabList = order
            .map(id => {
              const tab = currentTabs.find(t => t.id === id);
              return tab ? { id: tab.id, name: tab.name } : null;
            })
            .filter((t): t is { id: string; name: string } => t !== null);
          if (mruTabList.length < 2) return;
          mruCycling = true;
          mruShowOverlay = false;
          mruCycleIndex = e.shiftKey ? mruTabList.length - 1 : 1;
          if (mruOverlayTimer) clearTimeout(mruOverlayTimer);
          mruOverlayTimer = setTimeout(() => { mruShowOverlay = true; }, 150);
        } else {
          mruShowOverlay = true;
          if (mruOverlayTimer) { clearTimeout(mruOverlayTimer); mruOverlayTimer = null; }
          if (e.shiftKey) {
            mruCycleIndex = (mruCycleIndex - 1 + mruTabList.length) % mruTabList.length;
          } else {
            mruCycleIndex = (mruCycleIndex + 1) % mruTabList.length;
          }
        }
        return;
      }

      if (e.ctrlKey && (e.key === 'PageDown' || e.key === 'PageUp')) {
        e.preventDefault();
        if (e.key === 'PageUp') {
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
        if (mruCycling) {
          if (mruOverlayTimer) { clearTimeout(mruOverlayTimer); mruOverlayTimer = null; }
          mruCycling = false;
          mruShowOverlay = false;
          mruCycleIndex = 0;
          mruTabList = [];
          return;
        }
        // Priority chain: focus mode > active stream > nav panel
        if (get(focusMode)) {
          toggleFocusMode();
        } else if (get(streamState) === 'streaming') {
          import('../wailsjs/go/main/App').then(app => app.CancelStream()).catch(() => {});
          cancelStreamState();
          const stream = get(activeStream);
          if (stream) {
            setStreamActive(stream.tabId, false);
          }
        } else {
          navVisible = false;
          settingsOpen.set(false);
        }
        return;
      }
    }

    document.addEventListener('keydown', handleKeydown, true);

    function handleKeyup(e: KeyboardEvent) {
      if (e.key === 'Control' && mruCycling) {
        if (mruOverlayTimer) { clearTimeout(mruOverlayTimer); mruOverlayTimer = null; }
        const selectedTab = mruTabList[mruCycleIndex];
        if (selectedTab) {
          activeTabId.set(selectedTab.id);
        }
        mruCycling = false;
        mruShowOverlay = false;
        mruCycleIndex = 0;
        mruTabList = [];
      }
    }
    document.addEventListener('keyup', handleKeyup);

    function handleClickOutsideSettings(e: MouseEvent) {
      const target = e.target as HTMLElement;
      if (!target.closest('.settings') && !target.closest('.cb[title="Settings"]')) {
        settingsOpen.set(false);
      }
    }

    document.addEventListener('click', handleClickOutsideSettings);

    return () => {
      document.removeEventListener('keydown', handleKeydown, true);
      document.removeEventListener('keyup', handleKeyup);
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

<MruOverlay tabs={mruTabList} selectedIndex={mruCycleIndex} visible={mruShowOverlay} />

{#if langToast}
  <div class="lang-toast" aria-live="polite">
    <span class="lang-toast-label">Translation</span>
    <span class="lang-toast-lang">{langToast}</span>
  </div>
{/if}

<style>
  .lang-toast {
    position: fixed;
    bottom: 80px;
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 16px;
    background: var(--surface-elevated);
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    border: 1px solid var(--border);
    border-radius: 12px;
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.25);
    z-index: 60;
    animation: lang-toast-in 150ms ease-out;
    pointer-events: none;
  }

  .lang-toast-label {
    font-size: 12px;
    color: var(--text-tertiary);
  }

  .lang-toast-lang {
    font-size: 14px;
    font-family: 'JetBrains Mono', monospace;
    font-weight: 700;
    color: var(--accent-text);
    background: var(--accent-dim);
    padding: 2px 8px;
    border-radius: 6px;
  }

  @keyframes lang-toast-in {
    from { opacity: 0; transform: translateX(-50%) translateY(8px); }
    to { opacity: 1; transform: translateX(-50%) translateY(0); }
  }

  @media (prefers-reduced-motion: reduce) {
    .lang-toast { animation: none; }
  }
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
