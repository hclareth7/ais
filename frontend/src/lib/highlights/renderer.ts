export interface HighlightData {
  id: string;
  filePath: string;
  anchorText: string;
  prefixContext: string;
  suffixContext: string;
  color: string;
  createdAt: string;
}

export function clearHighlightMarks(viewerEl: HTMLElement): void {
  const marks = viewerEl.querySelectorAll('mark[data-highlight]');
  marks.forEach(mark => {
    const parent = mark.parentNode;
    if (!parent) return;
    while (mark.firstChild) {
      parent.insertBefore(mark.firstChild, mark);
    }
    parent.removeChild(mark);
    parent.normalize();
  });
}

export function applyHighlights(viewerEl: HTMLElement, highlights: HighlightData[]): void {
  clearHighlightMarks(viewerEl);
  if (!highlights.length) return;

  for (const hl of highlights) {
    const walker = document.createTreeWalker(viewerEl, NodeFilter.SHOW_TEXT, {
      acceptNode: (node) => {
        const parent = node.parentElement;
        if (parent?.closest('pre, code, mark[data-highlight]')) return NodeFilter.FILTER_REJECT;
        return NodeFilter.FILTER_ACCEPT;
      }
    });

    let textNode: Text | null;
    while ((textNode = walker.nextNode() as Text | null)) {
      const content = textNode.textContent || '';
      const idx = content.indexOf(hl.anchorText);
      if (idx === -1) continue;

      // Verify context match
      if (hl.prefixContext) {
        const before = content.slice(Math.max(0, idx - hl.prefixContext.length), idx);
        if (!before.endsWith(hl.prefixContext.slice(-Math.min(before.length, hl.prefixContext.length)))) continue;
      }

      const range = document.createRange();
      range.setStart(textNode, idx);
      range.setEnd(textNode, idx + hl.anchorText.length);

      const mark = document.createElement('mark');
      mark.setAttribute('data-highlight', hl.color);
      mark.setAttribute('data-highlight-id', hl.id);
      mark.setAttribute('tabindex', '0');
      range.surroundContents(mark);
      break;
    }
  }
}
