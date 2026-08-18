<script lang="ts">
  import { renderMarkdown } from '../markdown/renderer';
  import { activeTab, saveScrollPos } from '../stores/tabs';
  import { activeStream, streamState, type StreamError } from '../stores/stream';
  import { settingsOpen } from '../stores/ui';
  import WelcomeScreen from './WelcomeScreen.svelte';

  let { zoomLevel = 100, readingWidth = 720 }: {
    zoomLevel?: number;
    readingWidth?: number;
  } = $props();

  let viewerEl: HTMLElement | undefined = $state();
  let renderedHtml = $derived($activeTab ? renderMarkdown($activeTab.content) : '');

  let previousTabId: string | null = $state(null);

  // Streaming state
  let isStreamTab = $derived($activeTab?.type === 'stream');
  let isStreaming = $derived(isStreamTab && $activeTab?.streamActive === true);
  let userScrolledUp = $state(false);
  let showResumePill = $derived(isStreaming && userScrolledUp);
  let caretFading = $state(false);

  // Stream error/cancelled state for current tab
  let currentStreamState = $derived.by(() => {
    const stream = $activeStream;
    if (!stream || !$activeTab || stream.tabId !== $activeTab.id) return null;
    return stream;
  });

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

  /* ── Headings ── */
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

  /* ── Body ── */
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

  /* ── Code ── */
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

  /* ── Misc ── */
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

  .doc-inner :global(img) {
    max-width: 100%;
    border-radius: 6px;
    margin: 8px auto 20px;
    display: block;
  }

  .doc-inner :global(a) {
    color: var(--accent-text);
    text-decoration: none;
  }

  .doc-inner :global(a:hover) {
    text-decoration: underline;
  }

  .doc-inner :global(strong) {
    font-weight: 700;
    color: var(--text-primary);
  }

  .doc-inner :global(del) {
    text-decoration: line-through;
    color: var(--text-tertiary);
  }

  /* ── Syntax Highlighting ── */
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
</style>
