<script lang="ts">
  import { renderMarkdown } from '../markdown/renderer';
  import { activeTab, saveScrollPos } from '../stores/tabs';
  import WelcomeScreen from './WelcomeScreen.svelte';

  let { zoomLevel = 100, readingWidth = 720 }: {
    zoomLevel?: number;
    readingWidth?: number;
  } = $props();

  let viewerEl: HTMLElement | undefined = $state();
  let renderedHtml = $derived($activeTab ? renderMarkdown($activeTab.content) : '');

  let previousTabId: string | null = $state(null);

  $effect(() => {
    const currentId = $activeTab?.id ?? null;
    if (previousTabId && previousTabId !== currentId && viewerEl) {
      saveScrollPos(previousTabId, viewerEl.scrollTop);
    }
    previousTabId = currentId;
  });

  $effect(() => {
    if ($activeTab && viewerEl) {
      requestAnimationFrame(() => {
        if (viewerEl) {
          viewerEl.scrollTop = $activeTab?.scrollPos ?? 0;
        }
      });
    }
  });

  function handleContentClick(e: MouseEvent) {
    const target = e.target as HTMLElement;
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
</script>

{#if $activeTab}
  <div
    class="doc"
    bind:this={viewerEl}
    role="article"
    aria-label={$activeTab.name}
  >
    <div
      class="doc-inner"
      role="presentation"
      onclick={handleContentClick}
      onkeydown={() => {}}
      style="max-width: {readingWidth}px; transform: scale({zoomLevel / 100}); transform-origin: top center;"
    >
      {@html renderedHtml}
    </div>
  </div>
{:else}
  <WelcomeScreen />
{/if}

<style>
  .doc {
    flex: 1;
    overflow-y: auto;
    overflow-x: hidden;
    padding: 36px 0 80px;
    display: flex;
    justify-content: center;
    scrollbar-width: thin;
    scrollbar-color: var(--scrollbar-thumb) transparent;
    user-select: text;
  }

  .doc-inner {
    width: 100%;
    max-width: var(--reading-max-width);
    padding: 0 32px;
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
