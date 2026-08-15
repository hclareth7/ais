<script lang="ts">
  import { activeTab } from '../stores/tabs';
  import { tocVisible } from '../stores/ui';

  let { visible = false }: { visible?: boolean } = $props();

  interface TocEntry {
    text: string;
    level: number;
    id: string;
  }

  let entries = $derived.by(() => {
    const content = $activeTab?.content;
    if (!content) return [];
    const result: TocEntry[] = [];
    const lines = content.split('\n');
    for (const line of lines) {
      const match = line.match(/^(#{1,6})\s+(.+)/);
      if (match) {
        const level = match[1].length;
        const text = match[2].replace(/[#*`\[\]]/g, '').trim();
        const id = text.toLowerCase().replace(/[^\w]+/g, '-').replace(/(^-|-$)/g, '');
        if (text) result.push({ text, level, id });
      }
    }
    return result;
  });

  let activeSection = $state('');

  function scrollToHeading(entry: TocEntry) {
    const doc = document.querySelector('.doc');
    if (!doc) return;
    const headings = doc.querySelectorAll('h1, h2, h3, h4, h5, h6');
    for (const h of headings) {
      if (h.textContent?.trim() === entry.text) {
        h.scrollIntoView({ behavior: 'smooth', block: 'start' });
        activeSection = entry.id;
        break;
      }
    }
  }

  function handleKeydown(e: KeyboardEvent, entry: TocEntry) {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      scrollToHeading(entry);
    }
  }
</script>

{#if $activeTab && entries.length > 0}
  <aside
    class="toc-panel"
    class:vis={visible || $tocVisible}
    aria-label="Table of contents, {entries.length} headings"
    role="navigation"
  >
    <div class="toc-title">Outline</div>
    {#each entries as entry (entry.id + entry.text)}
      <div
        class="ti"
        class:sub={entry.level >= 3}
        class:on={activeSection === entry.id}
        role="link"
        tabindex="0"
        onclick={() => scrollToHeading(entry)}
        onkeydown={(e) => handleKeydown(e, entry)}
      >
        {entry.text}
      </div>
    {/each}
  </aside>
{/if}

<style>
  .toc-panel {
    position: absolute;
    right: 14px;
    top: 14px;
    bottom: 14px;
    width: 220px;
    background: var(--surface-solid);
    backdrop-filter: blur(30px);
    -webkit-backdrop-filter: blur(30px);
    border-radius: 20px;
    border: 1px solid var(--nav-border);
    padding: 18px 14px;
    z-index: 25;
    opacity: 0;
    transform: translateX(10px);
    pointer-events: none;
    transition: opacity 0.2s, transform 0.2s;
    overflow-y: auto;
    user-select: none;
  }

  @supports (backdrop-filter: blur(1px)) or (-webkit-backdrop-filter: blur(1px)) {
    .toc-panel {
      background: var(--nav-bg);
    }
  }

  .toc-panel:hover,
  .toc-panel.vis {
    opacity: 1;
    transform: translateX(0);
    pointer-events: auto;
    transition: opacity 0.12s, transform 0.12s;
  }

  .toc-title {
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--text-tertiary);
    padding: 0 6px 10px;
  }

  .ti {
    padding: 5px 6px;
    font-size: 13px;
    color: var(--text-tertiary);
    cursor: pointer;
    border-radius: 6px;
    transition: color 0.12s, background 0.12s;
    line-height: 1.4;
  }

  .ti:hover {
    color: var(--text-primary);
    background: var(--hover-bg);
  }

  .ti.on {
    color: var(--accent-text);
  }

  .ti.sub {
    padding-left: 22px;
    font-size: 12px;
  }
</style>
