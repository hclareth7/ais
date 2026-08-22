<script lang="ts">
  import { commandPaletteOpen, commandPaletteCategory } from '../stores/ui';
  import { fileTree, type FileNode } from '../stores/files';
  import { openTab, activeTab } from '../stores/tabs';
  import { setTheme } from '../stores/settings';
  import { searchScrollTarget } from '../stores/search';
  import { get } from 'svelte/store';

  let query = $state('');
  let inputEl: HTMLInputElement | undefined = $state();
  let selectedIdx = $state(0);
  let activeCategory = $state('all');

  // Search state
  let searchResults = $state<Array<{filePath: string, lineNumber: number, matchOffset: number, context: string}>>([]);
  let searchLoading = $state(false);
  let searchDebounceTimer: ReturnType<typeof setTimeout> | null = null;

  type Result = {
    type: 'file' | 'command';
    label: string;
    detail: string;
    action: () => void;
  };

  function flattenTree(node: FileNode | null, results: Result[] = []): Result[] {
    if (!node) return results;
    if (!node.isDir && node.name.endsWith('.md')) {
      results.push({
        type: 'file',
        label: node.name,
        detail: node.path,
        action: () => {
          openTab(node.path, node.name);
          close();
        }
      });
    }
    if (node.children) {
      for (const child of node.children) {
        flattenTree(child, results);
      }
    }
    return results;
  }

  async function copyDocument() {
    const tab = get(activeTab);
    if (!tab) return;
    try {
      const { ClipboardSetText } = await import('../../../wailsjs/runtime/runtime');
      await ClipboardSetText(tab.content);
    } catch {
      await navigator.clipboard.writeText(tab.content);
    }
  }

  const commands: Result[] = [
    { type: 'command', label: 'Copy document', detail: 'Copy markdown source to clipboard (Ctrl+Shift+C)', action: () => { copyDocument(); close(); } },
    { type: 'command', label: 'Theme: Dark', detail: 'Switch to dark mode', action: () => { setTheme('dark'); close(); } },
    { type: 'command', label: 'Theme: Light', detail: 'Switch to light mode', action: () => { setTheme('light'); close(); } },
    { type: 'command', label: 'Theme: System', detail: 'Follow system preference', action: () => { setTheme('system'); close(); } },
  ];

  let files = $derived(flattenTree($fileTree));

  let results = $derived.by(() => {
    const lowerQuery = query.toLowerCase();
    let items: Result[] = [];

    if (activeCategory === 'all' || activeCategory === 'docs') {
      items.push(...files.filter(f => f.label.toLowerCase().includes(lowerQuery) || f.detail.toLowerCase().includes(lowerQuery)));
    }
    if (activeCategory === 'all' || activeCategory === 'commands') {
      items.push(...commands.filter(c => c.label.toLowerCase().includes(lowerQuery) || c.detail.toLowerCase().includes(lowerQuery)));
    }

    return items.slice(0, 12);
  });

  let wasOpen = $state(false);
  $effect(() => {
    const isOpen = $commandPaletteOpen;
    if (isOpen && !wasOpen) {
      const initialCat = get(commandPaletteCategory);
      query = '';
      selectedIdx = 0;
      activeCategory = initialCat ?? 'all';
      commandPaletteCategory.set(null);
      searchResults = [];
      searchLoading = false;
      requestAnimationFrame(() => inputEl?.focus());
    }
    wasOpen = isOpen;
  });

  // Debounced search when in search mode
  $effect(() => {
    if (activeCategory !== 'search') return;
    const q = query;

    if (searchDebounceTimer) {
      clearTimeout(searchDebounceTimer);
    }

    if (!q.trim()) {
      searchResults = [];
      searchLoading = false;
      return;
    }

    searchLoading = true;
    searchDebounceTimer = setTimeout(async () => {
      try {
        const App = await import('../../../wailsjs/go/main/App');
        const results = await App.SearchFiles(q);
        // Only update if still on the search tab and query hasn't changed
        if (activeCategory === 'search') {
          searchResults = results ?? [];
        }
      } catch (err) {
        console.error('Search failed:', err);
        searchResults = [];
      } finally {
        searchLoading = false;
      }
    }, 300);
  });

  $effect(() => {
    void results;
    selectedIdx = 0;
  });

  function close() {
    if (searchDebounceTimer) {
      clearTimeout(searchDebounceTimer);
      searchDebounceTimer = null;
    }
    commandPaletteOpen.set(false);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      e.preventDefault();
      close();
      return;
    }

    if (activeCategory === 'search') {
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        selectedIdx = Math.min(selectedIdx + 1, searchResults.length - 1);
        return;
      }
      if (e.key === 'ArrowUp') {
        e.preventDefault();
        selectedIdx = Math.max(selectedIdx - 1, 0);
        return;
      }
      if (e.key === 'Enter' && searchResults[selectedIdx]) {
        e.preventDefault();
        const sr = searchResults[selectedIdx];
        const filename = sr.filePath.split('/').pop() ?? sr.filePath;
        searchScrollTarget.set({ query: query, filePath: sr.filePath, lineNumber: sr.lineNumber });
        openTab(sr.filePath, filename);
        close();
        return;
      }
      return;
    }

    if (e.key === 'ArrowDown') {
      e.preventDefault();
      selectedIdx = Math.min(selectedIdx + 1, results.length - 1);
      return;
    }

    if (e.key === 'ArrowUp') {
      e.preventDefault();
      selectedIdx = Math.max(selectedIdx - 1, 0);
      return;
    }

    if (e.key === 'Enter' && results[selectedIdx]) {
      e.preventDefault();
      results[selectedIdx].action();
      return;
    }
  }

  function handleOverlayClick(e: MouseEvent) {
    if (e.target === e.currentTarget) close();
  }

  function highlightContext(context: string, q: string): string {
    if (!q) return escapeHtmlStr(context);
    const escaped = escapeHtmlStr(context);
    const lowerEscaped = escaped.toLowerCase();
    const lowerQ = q.toLowerCase();
    const parts: string[] = [];
    let cursor = 0;
    while (cursor < escaped.length) {
      const idx = lowerEscaped.indexOf(lowerQ, cursor);
      if (idx < 0) {
        parts.push(escaped.slice(cursor));
        break;
      }
      parts.push(escaped.slice(cursor, idx));
      parts.push('<strong>' + escaped.slice(idx, idx + lowerQ.length) + '</strong>');
      cursor = idx + lowerQ.length;
    }
    return parts.join('');
  }

  function escapeHtmlStr(s: string): string {
    return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&#39;');
  }

  const categories = [
    { key: 'all', label: 'All' },
    { key: 'docs', label: 'Docs' },
    { key: 'search', label: 'Search' },
    { key: 'commands', label: 'Commands' },
  ];
</script>

{#if $commandPaletteOpen}
  <div
    class="pal-overlay"
    onclick={handleOverlayClick}
    onkeydown={handleKeydown}
    role="dialog"
    tabindex="-1"
    aria-label="Command palette"
    aria-modal="true"
  >
    <div class="pal">
      <div class="pal-input">
        <svg viewBox="0 0 24 24"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
        <input
          bind:this={inputEl}
          bind:value={query}
          placeholder={activeCategory === 'search' ? 'Search in files...' : 'Search docs, commands...'}
          autocomplete="off"
          spellcheck="false"
        />
        <span class="pal-kbd">esc</span>
      </div>

      <div class="pal-tabs">
        {#each categories as cat (cat.key)}
          <button
            class="pal-tab"
            class:on={activeCategory === cat.key}
            onclick={() => { activeCategory = cat.key; }}
          >{cat.label}</button>
        {/each}
      </div>

      {#if activeCategory === 'search'}
        <div class="pal-results" role="listbox">
          {#if searchLoading}
            <div class="pr-empty">Searching...</div>
          {:else}
            {#each searchResults as sr, i (sr.filePath + ':' + sr.lineNumber + ':' + sr.matchOffset)}
              {@const filename = sr.filePath.split('/').pop() ?? sr.filePath}
              <button
                class="pr"
                class:sel={i === selectedIdx}
                role="option"
                aria-selected={i === selectedIdx}
                onclick={() => {
                  searchScrollTarget.set({ query: query, filePath: sr.filePath, lineNumber: sr.lineNumber });
                  openTab(sr.filePath, filename);
                  close();
                }}
                onmouseenter={() => selectedIdx = i}
              >
                <svg viewBox="0 0 24 24">
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/>
                </svg>
                <div class="pr-info">
                  <div class="pr-t">{filename}<span class="pr-line">:{sr.lineNumber}</span></div>
                  <div class="pr-p pr-context">{@html highlightContext(sr.context, query)}</div>
                </div>
              </button>
            {/each}
            {#if searchResults.length === 0 && query.trim()}
              <div class="pr-empty">No matches found</div>
            {:else if !query.trim()}
              <div class="pr-empty">Type to search in files</div>
            {/if}
          {/if}
        </div>
      {:else}
        <div class="pal-results" role="listbox">
          {#each results as result, i (result.detail + result.label)}
            <button
              class="pr"
              class:sel={i === selectedIdx}
              role="option"
              aria-selected={i === selectedIdx}
              onclick={() => result.action()}
              onmouseenter={() => selectedIdx = i}
            >
              <svg viewBox="0 0 24 24">
                {#if result.type === 'file'}
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/>
                {:else}
                  <polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/>
                {/if}
              </svg>
              <div class="pr-info">
                <div class="pr-t">{result.label}</div>
                <div class="pr-p">{result.detail}</div>
              </div>
            </button>
          {/each}
          {#if results.length === 0}
            <div class="pr-empty">No results found</div>
          {/if}
        </div>
      {/if}

      <div class="pal-footer">
        {#if activeCategory === 'search'}
          <span><kbd>&uarr;&darr;</kbd> navigate</span>
          <span><kbd>&crarr;</kbd> open</span>
          <span><kbd>esc</kbd> close</span>
        {:else}
          <span><kbd>&uarr;&darr;</kbd> navigate</span>
          <span><kbd>&crarr;</kbd> select</span>
          <span><kbd>esc</kbd> close</span>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  .pal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    z-index: 100;
    display: flex;
    align-items: flex-start;
    justify-content: center;
    padding-top: 18vh;
  }

  .pal {
    width: 520px;
    background: var(--surface-solid);
    border: 1px solid var(--border);
    border-radius: 16px;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.4);
    overflow: hidden;
  }

  .pal-input {
    padding: 14px 18px;
    border-bottom: 1px solid var(--border);
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .pal-input svg {
    width: 18px;
    height: 18px;
    stroke: var(--text-tertiary);
    stroke-width: 1.5;
    stroke-linecap: round;
    fill: none;
    flex-shrink: 0;
  }

  .pal-input input {
    background: 0;
    border: 0;
    outline: 0;
    color: var(--text-primary);
    font-size: 15px;
    font-family: inherit;
    width: 100%;
  }

  .pal-input input::placeholder {
    color: var(--text-tertiary);
  }

  .pal-kbd {
    font-size: 11px;
    font-family: 'JetBrains Mono', monospace;
    color: var(--text-ghost);
    padding: 2px 7px;
    background: var(--hover-bg);
    border: 1px solid var(--border);
    border-radius: 5px;
    white-space: nowrap;
  }

  .pal-tabs {
    display: flex;
    padding: 0 12px;
    border-bottom: 1px solid var(--border);
  }

  .pal-tab {
    padding: 9px 13px;
    font-size: 12px;
    color: var(--text-tertiary);
    cursor: pointer;
    border-bottom: 2px solid transparent;
    transition: color 0.12s, border-color 0.12s;
    background: 0;
    border-top: 0;
    border-left: 0;
    border-right: 0;
    font-family: inherit;
  }

  .pal-tab:hover {
    color: var(--text-secondary);
  }

  .pal-tab.on {
    color: var(--accent-text);
    border-bottom-color: var(--accent-solid);
  }

  .pal-results {
    padding: 6px;
    max-height: 280px;
    overflow-y: auto;
  }

  .pr {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 9px 10px;
    border-radius: 10px;
    cursor: pointer;
    transition: background 0.1s;
    border: none;
    background: none;
    width: 100%;
    text-align: left;
    font-family: inherit;
  }

  .pr:hover, .pr.sel {
    background: var(--hover-bg);
  }

  .pr svg {
    width: 16px;
    height: 16px;
    stroke: var(--text-tertiary);
    stroke-width: 1.5;
    stroke-linecap: round;
    fill: none;
    flex-shrink: 0;
  }

  .pr-info {
    flex: 1;
    overflow: hidden;
  }

  .pr-t {
    font-size: 14px;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .pr-p {
    font-size: 11px;
    color: var(--text-tertiary);
    margin-top: 1px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .pr-line {
    font-size: 11px;
    color: var(--text-tertiary);
    font-family: 'JetBrains Mono', monospace;
    margin-left: 4px;
  }

  .pr-context :global(strong) {
    color: var(--text-primary);
    font-weight: 600;
  }

  .pr-empty {
    padding: 20px;
    text-align: center;
    font-size: 13px;
    color: var(--text-tertiary);
  }

  .pal-footer {
    padding: 8px 16px;
    border-top: 1px solid var(--border);
    display: flex;
    gap: 14px;
    font-size: 11px;
    color: var(--text-ghost);
  }

  .pal-footer kbd {
    font-family: 'JetBrains Mono', monospace;
    padding: 1px 4px;
    background: var(--hover-bg);
    border: 1px solid var(--border);
    border-radius: 3px;
    font-size: 10px;
    color: var(--text-tertiary);
  }

</style>
