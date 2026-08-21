<script lang="ts">
  import { renderMarkdown } from '../markdown/renderer';
  import { activeTab, saveScrollPos, openTab } from '../stores/tabs';
  import { activeStream, streamState, type StreamError } from '../stores/stream';
  import { settingsOpen } from '../stores/ui';
  import { applyHighlights, clearHighlightMarks, type HighlightData } from '../highlights/renderer';
  import { captureSelection } from '../highlights/selection';
  import { highlightsForFile, loadHighlightsForFile, addHighlight, removeHighlight, lastUsedColor } from '../stores/highlights';
  import { searchScrollTarget, inFileSearchOpen } from '../stores/search';
  import { get } from 'svelte/store';
  import WelcomeScreen from './WelcomeScreen.svelte';
  import ImageLightbox from './ImageLightbox.svelte';
  import InFileSearch from './InFileSearch.svelte';

  let { zoomLevel = 100, readingWidth = 720 }: {
    zoomLevel?: number;
    readingWidth?: number;
  } = $props();

  let viewerEl: HTMLElement | undefined = $state();
  let renderedHtml = $derived($activeTab ? renderMarkdown($activeTab.content, $activeTab.type === 'file' ? $activeTab.path : undefined) : '');

  let previousTabId: string | null = $state(null);

  // Streaming state
  let isStreamTab = $derived($activeTab?.type === 'stream');
  let isStreaming = $derived(isStreamTab && $activeTab?.streamActive === true);
  let userScrolledUp = $state(false);
  let showResumePill = $derived(isStreaming && userScrolledUp);
  let caretFading = $state(false);

  // Lightbox state
  let lightboxSrc = $state('');
  let lightboxAlt = $state('');
  let lightboxOpen = $state(false);

  // Quick action bar state
  let quickActionPos = $state<{x: number, y: number} | null>(null);
  let showQuickAction = $state(false);
  let cachedSelection: {anchorText: string, prefixContext: string, suffixContext: string} | null = null;

  // Search highlight state
  let activeSearchQuery = $state<string | null>(null);
  let searchHighlightTimer: ReturnType<typeof setTimeout> | null = null;

  // In-file search state
  let inFileSearchVisible = $state(false);
  let inFileSearchQuery = $state('');
  let inFileMatchCount = $state(0);
  let inFileCurrentMatch = $state(0);

  // Stream error/cancelled state for current tab
  let currentStreamState = $derived.by(() => {
    const stream = $activeStream;
    if (!stream || !$activeTab || stream.tabId !== $activeTab.id) return null;
    return stream;
  });

  function escapeHtml(text: string): string {
    return text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
  }

  $effect(() => {
    const currentId = $activeTab?.id ?? null;
    if (previousTabId && previousTabId !== currentId && viewerEl) {
      saveScrollPos(previousTabId, viewerEl.scrollTop);
    }
    previousTabId = currentId;
  });

  // Restore scroll for file tabs; reset for stream tabs
  $effect(() => {
    if ($activeTab && viewerEl) {
      if ($activeTab.type !== 'stream') {
        requestAnimationFrame(() => {
          if (viewerEl) {
            viewerEl.scrollTop = $activeTab?.scrollPos ?? 0;
          }
        });
      }
    }
  });

  // Auto-scroll during streaming
  $effect(() => {
    // Access content to trigger reactivity
    const _content = $activeTab?.content;
    if (isStreaming && viewerEl && !userScrolledUp) {
      requestAnimationFrame(() => {
        if (viewerEl) {
          viewerEl.scrollTop = viewerEl.scrollHeight;
        }
      });
    }
  });

  // Reset userScrolledUp when a new stream starts
  $effect(() => {
    if (isStreaming) {
      userScrolledUp = false;
    }
  });

  // Caret fade-out on stream completion
  $effect(() => {
    if (currentStreamState && (currentStreamState.state === 'complete' || currentStreamState.state === 'cancelled')) {
      caretFading = true;
      setTimeout(() => { caretFading = false; }, 200);
    }
  });

  // Image error fallback (Task 2.3)
  $effect(() => {
    const _html = renderedHtml;
    if (!viewerEl) return;
    requestAnimationFrame(() => {
      if (!viewerEl) return;
      const imgs = viewerEl.querySelectorAll('img:not([data-error-handled])');
      imgs.forEach(img => {
        img.setAttribute('data-error-handled', 'true');
        (img as HTMLImageElement).onerror = () => {
          const alt = (img as HTMLImageElement).alt || 'Image unavailable';
          const placeholder = document.createElement('div');
          placeholder.className = 'md-image-error';
          placeholder.innerHTML = `<svg viewBox="0 0 24 24" width="32" height="32"><rect x="3" y="3" width="18" height="18" rx="2" fill="none" stroke="currentColor" stroke-width="1.5"/><circle cx="8.5" cy="8.5" r="1.5" fill="currentColor"/><path d="M21 15l-5-5L5 21" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg><span>${escapeHtml(alt)}</span>`;
          img.replaceWith(placeholder);
        };
      });
    });
  });

  // Load highlights for current file tab (Task 3.6a)
  $effect(() => {
    const tab = $activeTab;
    if (tab?.type === 'file' && tab.path) {
      loadHighlightsForFile(tab.path);
    }
  });

  // Apply highlights after render (Task 3.6b)
  $effect(() => {
    const highlights = $highlightsForFile;
    const _html = renderedHtml;
    if (!viewerEl || !$activeTab || $activeTab.type !== 'file') return;
    requestAnimationFrame(() => {
      if (viewerEl) applyHighlights(viewerEl, highlights);
    });
  });

  // Listen for text selection changes (Task 3.6c)
  $effect(() => {
    document.addEventListener('selectionchange', handleSelectionChange);
    return () => document.removeEventListener('selectionchange', handleSelectionChange);
  });

  // Search: scroll to match and highlight all occurrences (from Command Palette cross-file search)
  $effect(() => {
    const target = $searchScrollTarget;
    if (!target || !viewerEl || !$activeTab) return;
    if ($activeTab.path !== target.filePath) return;

    activeSearchQuery = target.query;
    searchScrollTarget.set(null);

    requestAnimationFrame(() => {
      if (!viewerEl) return;
      applySearchHighlights(viewerEl, target.query);

      const firstMark = viewerEl.querySelector('.search-match');
      if (firstMark) {
        firstMark.scrollIntoView({ behavior: 'smooth', block: 'center' });
      }
    });
  });

  // Clear search highlights on tab change
  $effect(() => {
    const currentId = $activeTab?.id ?? null;
    if (currentId !== previousTabId) {
      clearSearchHighlights();
      activeSearchQuery = null;
      closeInFileSearch();
    }
  });

  // Listen for Ctrl+F toggle from store
  $effect(() => {
    const open = $inFileSearchOpen;
    if (open && !inFileSearchVisible) {
      inFileSearchVisible = true;
      inFileSearchOpen.set(false);
    }
  });

  function applySearchHighlights(container: HTMLElement, query: string) {
    if (!query) return;
    const lowerQuery = query.toLowerCase();
    const walker = document.createTreeWalker(container, NodeFilter.SHOW_TEXT, {
      acceptNode(node) {
        // Skip nodes inside pre, code, or existing mark elements
        const parent = node.parentElement;
        if (parent?.closest('pre, code, mark')) return NodeFilter.FILTER_REJECT;
        if (node.textContent && node.textContent.toLowerCase().includes(lowerQuery)) {
          return NodeFilter.FILTER_ACCEPT;
        }
        return NodeFilter.FILTER_SKIP;
      }
    });

    const matchNodes: Text[] = [];
    while (walker.nextNode()) {
      matchNodes.push(walker.currentNode as Text);
    }

    for (const textNode of matchNodes) {
      const text = textNode.textContent ?? '';
      const lowerText = text.toLowerCase();
      const fragments: (string | HTMLElement)[] = [];
      let cursor = 0;

      while (cursor < text.length) {
        const idx = lowerText.indexOf(lowerQuery, cursor);
        if (idx < 0) {
          fragments.push(text.slice(cursor));
          break;
        }
        if (idx > cursor) {
          fragments.push(text.slice(cursor, idx));
        }
        const mark = document.createElement('mark');
        mark.className = 'search-match';
        mark.textContent = text.slice(idx, idx + query.length);
        fragments.push(mark);
        cursor = idx + query.length;
      }

      if (fragments.length > 1 || (fragments.length === 1 && fragments[0] instanceof HTMLElement)) {
        const parent = textNode.parentNode;
        if (!parent) continue;
        for (const frag of fragments) {
          if (typeof frag === 'string') {
            parent.insertBefore(document.createTextNode(frag), textNode);
          } else {
            parent.insertBefore(frag, textNode);
          }
        }
        parent.removeChild(textNode);
      }
    }
  }

  function clearSearchHighlights() {
    if (!viewerEl) return;
    const marks = viewerEl.querySelectorAll('mark.search-match');
    marks.forEach(mark => {
      const parent = mark.parentNode;
      if (!parent) return;
      const text = document.createTextNode(mark.textContent ?? '');
      parent.replaceChild(text, mark);
      // Normalize to merge adjacent text nodes
      parent.normalize();
    });
    if (searchHighlightTimer) {
      clearTimeout(searchHighlightTimer);
      searchHighlightTimer = null;
    }
  }

  function handleInFileSearch(query: string) {
    inFileSearchQuery = query;
    clearSearchHighlights();
    activeSearchQuery = null;

    if (!query || !viewerEl) {
      inFileMatchCount = 0;
      inFileCurrentMatch = 0;
      return;
    }

    activeSearchQuery = query;
    applySearchHighlights(viewerEl, query);

    const marks = viewerEl.querySelectorAll('.search-match');
    inFileMatchCount = marks.length;
    if (marks.length > 0) {
      inFileCurrentMatch = 1;
      marks[0].classList.add('search-match-current');
      marks[0].scrollIntoView({ behavior: 'smooth', block: 'center' });
    } else {
      inFileCurrentMatch = 0;
    }
  }

  function goToNextMatch() {
    if (!viewerEl || inFileMatchCount === 0) return;
    const marks = viewerEl.querySelectorAll('.search-match');
    marks.forEach(m => m.classList.remove('search-match-current'));
    inFileCurrentMatch = (inFileCurrentMatch % marks.length) + 1;
    const current = marks[inFileCurrentMatch - 1];
    current.classList.add('search-match-current');
    current.scrollIntoView({ behavior: 'smooth', block: 'center' });
  }

  function goToPrevMatch() {
    if (!viewerEl || inFileMatchCount === 0) return;
    const marks = viewerEl.querySelectorAll('.search-match');
    marks.forEach(m => m.classList.remove('search-match-current'));
    inFileCurrentMatch = inFileCurrentMatch <= 1 ? marks.length : inFileCurrentMatch - 1;
    const current = marks[inFileCurrentMatch - 1];
    current.classList.add('search-match-current');
    current.scrollIntoView({ behavior: 'smooth', block: 'center' });
  }

  function closeInFileSearch() {
    inFileSearchVisible = false;
    inFileSearchQuery = '';
    inFileMatchCount = 0;
    inFileCurrentMatch = 0;
    clearSearchHighlights();
    activeSearchQuery = null;
  }

  function handleSelectionChange() {
    const sel = window.getSelection();
    if (!sel || sel.isCollapsed || !viewerEl?.contains(sel.anchorNode)) {
      showQuickAction = false;
      cachedSelection = null;
      return;
    }
    if (isStreaming) { showQuickAction = false; cachedSelection = null; return; }
    if ($activeTab?.type !== 'file') { showQuickAction = false; cachedSelection = null; return; }

    const node = sel.anchorNode?.nodeType === Node.TEXT_NODE ? sel.anchorNode.parentElement : sel.anchorNode as HTMLElement;
    if (node?.closest('pre, code')) { showQuickAction = false; cachedSelection = null; return; }

    const captured = captureSelection(viewerEl!);
    if (captured) cachedSelection = captured;

    const range = sel.getRangeAt(0);
    const rect = range.getBoundingClientRect();
    quickActionPos = { x: rect.left + rect.width / 2, y: rect.top - 8 };
    showQuickAction = true;
  }

  async function createHighlight(color: string) {
    if (!viewerEl || !$activeTab || $activeTab.type !== 'file') return;
    const sel = cachedSelection;
    if (!sel) return;

    lastUsedColor.set(color);
    const hl: HighlightData = {
      id: `hl-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
      filePath: $activeTab.path,
      anchorText: sel.anchorText,
      prefixContext: sel.prefixContext,
      suffixContext: sel.suffixContext,
      color,
      createdAt: new Date().toISOString(),
    };
    cachedSelection = null;
    await addHighlight(hl);
    window.getSelection()?.removeAllRanges();
    showQuickAction = false;
  }

  function handleMarkClick(e: MouseEvent) {
    const mark = (e.target as HTMLElement).closest('mark[data-highlight]');
    if (!mark || !$activeTab?.path) return;
    const hlId = mark.getAttribute('data-highlight-id');
    if (hlId) {
      removeHighlight($activeTab.path, hlId);
    }
  }

  function handleScroll() {
    if (!viewerEl || !isStreaming) return;
    const atBottom = viewerEl.scrollHeight - viewerEl.scrollTop - viewerEl.clientHeight < 50;
    userScrolledUp = !atBottom;
  }

  function resumeFollowing() {
    userScrolledUp = false;
    if (viewerEl) {
      viewerEl.scrollTop = viewerEl.scrollHeight;
    }
  }

  async function handleCodeCopy(btn: HTMLElement) {
    const pre = btn.closest('pre.code-block');
    if (!pre) return;
    const code = pre.querySelector('code');
    if (!code) return;
    const text = code.textContent ?? '';
    try {
      const { ClipboardSetText } = await import('../../../wailsjs/runtime/runtime');
      await ClipboardSetText(text);
    } catch {
      await navigator.clipboard.writeText(text);
    }
    btn.classList.add('copied');
    btn.innerHTML = '<svg viewBox="0 0 20 20"><polyline points="4 10 8 14 16 5"/></svg>';
    setTimeout(() => {
      btn.classList.remove('copied');
      btn.innerHTML = '<svg viewBox="0 0 20 20"><rect x="6" y="6" width="10" height="12" rx="1.5"/><path d="M4 14V4a1.5 1.5 0 011.5-1.5H13"/></svg>';
    }, 1500);
  }

  function handleContentClick(e: MouseEvent) {
    const target = e.target as HTMLElement;

    // Handle "Open Settings" link in stream errors
    if (target.closest('.stream-error-action')) {
      e.preventDefault();
      settingsOpen.set(true);
      return;
    }

    const copyBtn = target.closest('.code-copy-btn') as HTMLElement | null;
    if (copyBtn) {
      e.preventDefault();
      e.stopPropagation();
      handleCodeCopy(copyBtn);
      return;
    }

    // Image lightbox (Task 2.7)
    const img = target.closest('img') as HTMLImageElement | null;
    if (img && !img.closest('.md-image-error')) {
      lightboxSrc = img.src;
      lightboxAlt = img.alt || '';
      lightboxOpen = true;
      return;
    }

    // Link interception (Tasks 1.2-1.5)
    const link = target.closest('a') as HTMLAnchorElement | null;
    if (link) {
      e.preventDefault();
      const href = link.getAttribute('href');
      if (!href) return;
      if (href.startsWith('http://') || href.startsWith('https://')) {
        // External: open in browser via Go binding
        import('../../../wailsjs/go/main/App').then((app: any) => {
          app.OpenExternalURL(href);
        }).catch((err: any) => console.error('Failed to open URL:', err));
      } else if (href.startsWith('#')) {
        // Anchor: scroll to heading
        const id = href.slice(1);
        const el = viewerEl?.querySelector(`[id="${CSS.escape(id)}"]`);
        if (el) el.scrollIntoView({ behavior: 'smooth' });
      } else if (/\.(?:md|markdown)(?:#|$)/i.test(href)) {
        // Local markdown: open as tab
        const currentPath = $activeTab?.path ?? '';
        const currentDir = currentPath.includes('/') ? currentPath.substring(0, currentPath.lastIndexOf('/')) : '';
        const resolved = currentDir ? `${currentDir}/${href.split('#')[0]}` : href.split('#')[0];
        // Normalize path (remove ../ segments)
        const parts = resolved.split('/');
        const normalized: string[] = [];
        for (const p of parts) {
          if (p === '..') normalized.pop();
          else if (p !== '.') normalized.push(p);
        }
        const finalPath = normalized.join('/');
        const filename = finalPath.split('/').pop() ?? finalPath;
        openTab(finalPath, filename);
      }
      return;
    }

    // Highlight mark click (Task 3.6d)
    const mark = target.closest('mark[data-highlight]') as HTMLElement | null;
    if (mark) {
      handleMarkClick(e);
      return;
    }

    const heading = target.closest('h1, h2, h3, h4, h5, h6');
    if (!heading || !viewerEl) return;

    const headingLevel = parseInt(heading.tagName[1]);
    let sibling = heading.nextElementSibling;
    const elementsToToggle: HTMLElement[] = [];

    while (sibling) {
      if (sibling.matches('h1, h2, h3, h4, h5, h6')) {
        const siblingLevel = parseInt(sibling.tagName[1]);
        if (siblingLevel <= headingLevel) break;
      }
      elementsToToggle.push(sibling as HTMLElement);
      sibling = sibling.nextElementSibling;
    }

    const isCollapsed = elementsToToggle.some(el => el.style.display === 'none');
    elementsToToggle.forEach(el => {
      el.style.display = isCollapsed ? '' : 'none';
    });

    heading.classList.toggle('collapsed', !isCollapsed);
  }

  function getErrorDisplay(error: StreamError): { dotClass: string; message: string; hint: string; showSettings: boolean } {
    switch (error.code) {
      case 'auth':
        return { dotClass: '', message: 'API key is invalid.', hint: 'Check your key in Settings.', showSettings: true };
      case 'rate_limit':
        return { dotClass: '', message: 'Rate limit reached.', hint: 'Try again in a moment.', showSettings: false };
      case 'network':
        return { dotClass: '', message: 'Connection lost.', hint: 'Check your network and try again.', showSettings: false };
      default:
        return { dotClass: '', message: 'Something went wrong.', hint: error.message, showSettings: false };
    }
  }
</script>

{#if $activeTab}
  <InFileSearch
    visible={inFileSearchVisible}
    matchCount={inFileMatchCount}
    currentMatch={inFileCurrentMatch}
    onsearch={handleInFileSearch}
    onnext={goToNextMatch}
    onprev={goToPrevMatch}
    onclose={closeInFileSearch}
  />
  <div
    class="doc"
    class:streaming={isStreaming}
    bind:this={viewerEl}
    role="article"
    aria-label={isStreamTab ? `AI response: ${$activeTab.name}` : $activeTab.name}
    aria-busy={isStreaming}
    aria-live={isStreamTab ? 'polite' : undefined}
    onscroll={handleScroll}
  >
    <div
      class="doc-inner"
      role="presentation"
      onclick={handleContentClick}
      onkeydown={() => {}}
      style="max-width: {readingWidth}px; transform: scale({zoomLevel / 100}); transform-origin: top center;"
    >
      {@html renderedHtml}

      {#if isStreaming || caretFading}
        <span class="stream-caret" class:fade-out={caretFading}></span>
      {/if}

      {#if currentStreamState?.state === 'cancelled'}
        <div class="stream-stopped">Stopped</div>
      {/if}

      {#if currentStreamState?.state === 'error' && currentStreamState.error}
        {@const errDisplay = getErrorDisplay(currentStreamState.error)}
        <div class="stream-error">
          <span class="stream-error-dot" class:warning={currentStreamState.error.code === 'auth'}></span>
          <div class="stream-error-text">
            <div>{errDisplay.message}</div>
            <div style="color: var(--text-tertiary); margin-top: 2px;">{errDisplay.hint}</div>
            {#if errDisplay.showSettings}
              <button class="stream-error-action" style="margin-top: 4px;">Open Settings</button>
            {/if}
          </div>
        </div>
      {/if}

      {#if isStreamTab && !$activeTab.content && !isStreaming && !currentStreamState?.error && currentStreamState?.state !== 'cancelled'}
        <div class="stream-error">
          <span class="stream-error-dot warning"></span>
          <div class="stream-error-text">No content received.</div>
        </div>
      {/if}
    </div>
  </div>

  {#if showResumePill}
    <button
      class="resume-pill"
      aria-label="Resume auto-scroll"
      onclick={resumeFollowing}
    >
      Resume following
    </button>
  {/if}

  <ImageLightbox src={lightboxSrc} alt={lightboxAlt} open={lightboxOpen} onclose={() => lightboxOpen = false} />

  {#if showQuickAction && quickActionPos}
    <div
      class="quick-action"
      style="left: {quickActionPos.x}px; top: {quickActionPos.y}px;"
      role="toolbar"
      aria-label="Highlight text"
    >
      {#each ['yellow', 'green', 'blue', 'pink', 'purple', 'orange'] as color}
        <button
          class="qa-dot"
          style="background: var(--hl-{color}-dot);"
          title="Highlight {color}"
          aria-label="Highlight {color}"
          onmousedown={(e) => { e.preventDefault(); createHighlight(color); }}
        ></button>
      {/each}
    </div>
  {/if}
{:else}
  <WelcomeScreen />
{/if}

<style>
  .doc {
    flex: 1;
    overflow-y: auto;
    overflow-x: hidden;
    display: flex;
    justify-content: center;
    scrollbar-width: thin;
    scrollbar-color: var(--scrollbar-thumb) transparent;
    user-select: text;
    position: relative;
  }

  /* Streaming document border glow (Design.md) */
  .doc.streaming {
    box-shadow: inset 0 0 0 1px var(--stream-glow);
    animation: stream-pulse 2s ease-in-out infinite;
  }

  @media (prefers-reduced-motion: reduce) {
    .doc.streaming {
      animation: none;
      box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--stream-glow) 35%, transparent);
    }
  }

  .doc-inner {
    width: 100%;
    max-width: var(--reading-max-width);
    padding: 36px 32px 140px;
    line-height: 1.75;
    color: var(--text-primary);
    font-size: 16px;
    transition: max-width 0.2s;
  }

  /* -- Headings -- */
  .doc-inner :global(h1) {
    font-size: 36px;
    font-weight: 700;
    letter-spacing: -0.03em;
    line-height: 1.2;
    margin-bottom: 8px;
    margin-top: 48px;
    cursor: pointer;
    transition: color 0.12s;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .doc-inner :global(h1:first-child) {
    margin-top: 0;
  }

  .doc-inner :global(h1:hover) {
    color: var(--accent-text);
  }

  .doc-inner :global(h2) {
    font-size: 24px;
    font-weight: 600;
    letter-spacing: -0.02em;
    margin-top: 48px;
    margin-bottom: 16px;
    cursor: pointer;
    transition: color 0.12s;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .doc-inner :global(h2:hover) {
    color: var(--accent-text);
  }

  .doc-inner :global(h3) {
    font-size: 18px;
    font-weight: 600;
    margin-top: 32px;
    margin-bottom: 12px;
    cursor: pointer;
    transition: color 0.12s;
  }

  .doc-inner :global(h3:hover) {
    color: var(--accent-text);
  }

  .doc-inner :global(h4) {
    font-size: 16px;
    font-weight: 600;
    margin-top: 24px;
    margin-bottom: 12px;
    cursor: pointer;
    transition: color 0.12s;
  }

  .doc-inner :global(h5),
  .doc-inner :global(h6) {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-secondary);
    margin-top: 20px;
    margin-bottom: 8px;
    cursor: pointer;
  }

  .doc-inner :global(h1.collapsed::after),
  .doc-inner :global(h2.collapsed::after),
  .doc-inner :global(h3.collapsed::after),
  .doc-inner :global(h4.collapsed::after),
  .doc-inner :global(h5.collapsed::after),
  .doc-inner :global(h6.collapsed::after) {
    content: ' \25B6';
    font-size: 12px;
    color: var(--text-tertiary);
    margin-left: 8px;
  }

  /* -- Body -- */
  .doc-inner :global(p) {
    margin-bottom: 16px;
    color: var(--text-secondary);
    font-size: 16px;
    line-height: 1.75;
  }

  .doc-inner :global(ul),
  .doc-inner :global(ol) {
    margin: 0 0 16px;
    padding-left: 24px;
    color: var(--text-secondary);
  }

  .doc-inner :global(li) {
    margin-bottom: 6px;
    line-height: 1.65;
  }

  .doc-inner :global(li ul),
  .doc-inner :global(li ol) {
    margin-top: 4px;
    margin-bottom: 0;
  }

  .doc-inner :global(blockquote) {
    margin: 0 0 20px;
    padding: 16px;
    background: var(--hover-bg);
    border-left: 3px solid var(--accent-solid);
    border-radius: 0 6px 6px 0;
  }

  .doc-inner :global(blockquote p:last-child) {
    margin-bottom: 0;
  }

  /* -- Code -- */
  .doc-inner :global(pre.code-block) {
    margin: 16px 0 24px;
    padding: 0;
    background: var(--code-bg);
    border: 1px solid var(--code-border);
    border-radius: 12px;
    overflow: hidden;
    position: relative;
  }

  .doc-inner :global(.code-lang) {
    position: absolute;
    top: 8px;
    right: 12px;
    font-size: 11px;
    color: var(--text-tertiary);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    pointer-events: none;
  }

  .doc-inner :global(.code-copy-btn) {
    position: absolute;
    top: 6px;
    right: 6px;
    width: 28px;
    height: 28px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: none;
    background: var(--hover-bg);
    border-radius: 6px;
    cursor: pointer;
    opacity: 0;
    transition: opacity 0.15s, background 0.12s;
    z-index: 2;
    padding: 0;
  }

  .doc-inner :global(pre.code-block:hover .code-copy-btn) {
    opacity: 1;
  }

  .doc-inner :global(.code-copy-btn:hover) {
    background: var(--active-bg);
  }

  .doc-inner :global(.code-copy-btn svg) {
    width: 14px;
    height: 14px;
    stroke: var(--text-secondary);
    stroke-width: 1.5;
    stroke-linecap: round;
    stroke-linejoin: round;
    fill: none;
  }

  .doc-inner :global(.code-copy-btn.copied svg) {
    stroke: var(--accent-text);
  }

  .doc-inner :global(pre.code-block:hover .code-lang) {
    opacity: 0;
  }

  .doc-inner :global(pre.code-block code) {
    display: block;
    padding: 14px 18px;
    overflow-x: auto;
    font-family: 'JetBrains Mono', 'Fira Code', 'Cascadia Code', 'SF Mono', Consolas, monospace;
    font-size: 14px;
    line-height: 1.6;
    background: none;
    border-radius: 0;
    border: none;
    color: var(--text-secondary);
  }

  .doc-inner :global(code) {
    font-family: 'JetBrains Mono', 'Fira Code', 'Cascadia Code', 'SF Mono', Consolas, monospace;
    font-size: 14px;
    background: var(--code-bg);
    border: 1px solid var(--code-border);
    padding: 2px 6px;
    border-radius: 4px;
  }

  /* -- Misc -- */
  .doc-inner :global(hr) {
    border: none;
    height: 1px;
    background: var(--border);
    margin: 24px 0;
  }

  .doc-inner :global(table) {
    width: 100%;
    border-collapse: collapse;
    margin: 0 0 20px;
  }

  .doc-inner :global(th),
  .doc-inner :global(td) {
    padding: 12px 16px;
    border: 1px solid var(--border);
    text-align: left;
  }

  .doc-inner :global(th) {
    background: var(--hover-bg);
    font-weight: 600;
  }

  /* -- Images (Task 2.4) -- */
  .doc-inner :global(img) {
    max-width: 100%;
    border-radius: 12px;
    margin: 24px auto;
    display: block;
    border: 1px solid var(--border);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    cursor: zoom-in;
    transition: border-color 0.12s;
  }

  .doc-inner :global(img:hover) {
    border-color: var(--border-hover);
  }

  .doc-inner :global(.md-image-error) {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 32px;
    margin: 24px auto;
    background: var(--img-error-bg);
    border: 1px solid var(--border);
    border-radius: 12px;
    color: var(--img-error-text);
    font-size: 13px;
    max-width: 400px;
  }

  .doc-inner :global(.md-image-error svg) {
    opacity: 0.4;
  }

  /* -- Links -- */
  .doc-inner :global(a) {
    color: var(--accent-text);
    text-decoration: none;
  }

  .doc-inner :global(a:hover) {
    text-decoration: underline;
  }

  /* External link indicator (Task 1.6) */
  .doc-inner :global(a[data-external])::after {
    content: '\2197';
    font-size: 0.7em;
    opacity: 0.35;
    color: var(--text-tertiary);
    margin-left: 2px;
  }

  .doc-inner :global(strong) {
    font-weight: 700;
    color: var(--text-primary);
  }

  .doc-inner :global(del) {
    text-decoration: line-through;
    color: var(--text-tertiary);
  }

  /* -- Syntax Highlighting -- */
  .doc-inner :global(.hljs-keyword),
  .doc-inner :global(.hljs-selector-tag) {
    color: var(--code-kw);
  }

  .doc-inner :global(.hljs-string),
  .doc-inner :global(.hljs-regexp) {
    color: var(--code-str);
  }

  .doc-inner :global(.hljs-comment),
  .doc-inner :global(.hljs-doctag) {
    color: var(--code-cm);
    font-style: italic;
  }

  .doc-inner :global(.hljs-title),
  .doc-inner :global(.hljs-title.function_) {
    color: var(--code-fn);
  }

  .doc-inner :global(.hljs-number),
  .doc-inner :global(.hljs-literal) {
    color: var(--code-num);
  }

  .doc-inner :global(.hljs-type),
  .doc-inner :global(.hljs-built_in),
  .doc-inner :global(.hljs-title.class_) {
    color: var(--code-type);
  }

  .doc-inner :global(.hljs-variable),
  .doc-inner :global(.hljs-attr) {
    color: var(--text-secondary);
  }

  .doc-inner :global(.hljs-operator) {
    color: var(--code-kw);
  }

  .doc-inner :global(.hljs-punctuation) {
    color: var(--text-secondary);
  }

  .doc-inner :global(.hljs-meta),
  .doc-inner :global(.hljs-symbol) {
    color: var(--code-kw);
  }

  .doc-inner :global(.hljs-params) {
    color: var(--text-secondary);
  }

  /* -- Quick Action Bar (Task 3.6e) -- */
  .quick-action {
    position: fixed;
    transform: translate(-50%, -100%);
    display: flex;
    gap: 6px;
    padding: 6px 10px;
    background: var(--surface-elevated);
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    border: 1px solid var(--border);
    border-radius: 20px;
    z-index: 50;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
    animation: qa-in 120ms ease-out;
  }

  .qa-dot {
    width: 14px;
    height: 14px;
    border-radius: 50%;
    border: 1.5px solid rgba(255, 255, 255, 0.2);
    cursor: pointer;
    transition: transform 0.12s;
    padding: 0;
  }

  .qa-dot:hover {
    transform: scale(1.3);
  }

  @keyframes qa-in {
    from { opacity: 0; transform: translate(-50%, -100%) scale(0.9); }
    to { opacity: 1; transform: translate(-50%, -100%) scale(1); }
  }
</style>
