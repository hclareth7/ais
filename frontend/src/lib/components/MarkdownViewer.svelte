<script lang="ts">
  import { renderMarkdown } from '../markdown/renderer';
  import { activeTab, saveScrollPos } from '../stores/tabs';
  import WelcomeScreen from './WelcomeScreen.svelte';

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

<div class="viewer-area" style="grid-area: viewer;">
  {#if $activeTab}
    <div
      class="viewer-scroll"
      bind:this={viewerEl}
      role="article"
      aria-label={$activeTab.name}
    >
      <div
        class="markdown-body"
        role="presentation"
        onclick={handleContentClick}
        onkeydown={() => {}}
      >
        {@html renderedHtml}
      </div>
    </div>
  {:else}
    <WelcomeScreen />
  {/if}
</div>

<style>
  .viewer-area {
    overflow: hidden;
    background: var(--bg-primary);
  }

  .viewer-scroll {
    height: 100%;
    overflow-y: auto;
    overflow-x: hidden;
  }

  .markdown-body {
    max-width: 720px;
    margin: 0 auto;
    padding: 40px 32px 48px;
    color: var(--text-primary);
    font-size: 16px;
    line-height: 1.75;
  }

  .markdown-body :global(h1) {
    font-size: 30px;
    line-height: 40px;
    font-weight: 700;
    margin-top: 48px;
    margin-bottom: 24px;
    padding-bottom: 8px;
    border-bottom: 1px solid var(--border);
    cursor: pointer;
  }

  .markdown-body :global(h1:first-child) {
    font-size: 36px;
    line-height: 44px;
    margin-top: 0;
  }

  .markdown-body :global(h2) {
    font-size: 24px;
    line-height: 32px;
    font-weight: 700;
    margin-top: 40px;
    margin-bottom: 20px;
    padding-bottom: 6px;
    border-bottom: 1px solid var(--border);
    cursor: pointer;
  }

  .markdown-body :global(h3) {
    font-size: 20px;
    line-height: 28px;
    font-weight: 600;
    margin-top: 32px;
    margin-bottom: 16px;
    cursor: pointer;
  }

  .markdown-body :global(h4) {
    font-size: 18px;
    line-height: 28px;
    font-weight: 600;
    margin-top: 24px;
    margin-bottom: 12px;
    cursor: pointer;
  }

  .markdown-body :global(h5),
  .markdown-body :global(h6) {
    font-size: 16px;
    line-height: 24px;
    font-weight: 600;
    color: var(--text-secondary);
    margin-top: 20px;
    margin-bottom: 8px;
    cursor: pointer;
  }

  .markdown-body :global(h1.collapsed::after),
  .markdown-body :global(h2.collapsed::after),
  .markdown-body :global(h3.collapsed::after),
  .markdown-body :global(h4.collapsed::after),
  .markdown-body :global(h5.collapsed::after),
  .markdown-body :global(h6.collapsed::after) {
    content: ' ▶';
    font-size: 12px;
    color: var(--text-muted);
    margin-left: 8px;
  }

  .markdown-body :global(p) {
    margin: 0 0 20px;
  }

  .markdown-body :global(ul),
  .markdown-body :global(ol) {
    margin: 0 0 20px;
    padding-left: 24px;
  }

  .markdown-body :global(li) {
    margin-bottom: 4px;
  }

  .markdown-body :global(li ul),
  .markdown-body :global(li ol) {
    margin-top: 4px;
    margin-bottom: 0;
  }

  .markdown-body :global(blockquote) {
    margin: 0 0 20px;
    padding: 16px;
    background: var(--bg-inset);
    border-left: 3px solid var(--accent);
    border-radius: 0 6px 6px 0;
  }

  .markdown-body :global(blockquote p:last-child) {
    margin-bottom: 0;
  }

  .markdown-body :global(pre.code-block) {
    margin: 0 0 20px;
    padding: 0;
    background: var(--bg-code);
    border-radius: 6px;
    overflow: hidden;
    position: relative;
  }

  .markdown-body :global(.code-lang) {
    position: absolute;
    top: 8px;
    right: 12px;
    font-size: 12px;
    color: var(--text-muted);
    pointer-events: none;
  }

  .markdown-body :global(pre.code-block code) {
    display: block;
    padding: 16px;
    overflow-x: auto;
    font-family: 'JetBrains Mono', 'Fira Code', 'Cascadia Code', 'SF Mono', Consolas, monospace;
    font-size: 14px;
    line-height: 22px;
    background: none;
    border-radius: 0;
  }

  .markdown-body :global(code) {
    background: var(--bg-code);
    padding: 2px 6px;
    border-radius: 4px;
    font-family: 'JetBrains Mono', 'Fira Code', 'Cascadia Code', 'SF Mono', Consolas, monospace;
    font-size: 14px;
  }

  .markdown-body :global(hr) {
    border: none;
    border-top: 1px solid var(--border);
    margin: 32px 0;
  }

  .markdown-body :global(table) {
    width: 100%;
    border-collapse: collapse;
    margin: 0 0 20px;
  }

  .markdown-body :global(th),
  .markdown-body :global(td) {
    padding: 12px 16px;
    border: 1px solid var(--border);
    text-align: left;
  }

  .markdown-body :global(th) {
    background: var(--bg-inset);
    font-weight: 600;
  }

  .markdown-body :global(img) {
    max-width: 100%;
    border-radius: 6px;
    margin: 8px auto 20px;
    display: block;
  }

  .markdown-body :global(a) {
    color: var(--text-link);
    text-decoration: none;
  }

  .markdown-body :global(a:hover) {
    text-decoration: underline;
  }

  .markdown-body :global(strong) {
    font-weight: 700;
  }

  .markdown-body :global(del) {
    text-decoration: line-through;
    color: var(--text-muted);
  }

  /* Syntax highlighting */
  .markdown-body :global(.hljs-keyword),
  .markdown-body :global(.hljs-selector-tag) {
    color: var(--hl-keyword);
  }

  .markdown-body :global(.hljs-string),
  .markdown-body :global(.hljs-regexp) {
    color: var(--hl-string);
  }

  .markdown-body :global(.hljs-comment),
  .markdown-body :global(.hljs-doctag) {
    color: var(--hl-comment);
    font-style: italic;
  }

  .markdown-body :global(.hljs-title),
  .markdown-body :global(.hljs-title.function_) {
    color: var(--hl-function);
  }

  .markdown-body :global(.hljs-number),
  .markdown-body :global(.hljs-literal) {
    color: var(--hl-number);
  }

  .markdown-body :global(.hljs-type),
  .markdown-body :global(.hljs-built_in),
  .markdown-body :global(.hljs-title.class_) {
    color: var(--hl-type);
  }

  .markdown-body :global(.hljs-variable),
  .markdown-body :global(.hljs-attr) {
    color: var(--hl-variable);
  }

  .markdown-body :global(.hljs-operator) {
    color: var(--hl-operator);
  }

  .markdown-body :global(.hljs-punctuation) {
    color: var(--hl-punctuation);
  }

  .markdown-body :global(.hljs-meta),
  .markdown-body :global(.hljs-symbol) {
    color: var(--hl-keyword);
  }

  .markdown-body :global(.hljs-params) {
    color: var(--hl-variable);
  }
</style>
