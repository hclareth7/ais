export interface SelectionAnchor {
  anchorText: string;
  prefixContext: string;
  suffixContext: string;
}

export function captureSelection(viewerEl: HTMLElement): SelectionAnchor | null {
  const sel = window.getSelection();
  if (!sel || sel.isCollapsed || !sel.rangeCount) return null;

  const range = sel.getRangeAt(0);
  if (!viewerEl.contains(range.commonAncestorContainer)) return null;

  const text = sel.toString();
  if (!text.trim()) return null;

  // Reject selections inside code blocks
  const ancestor = range.commonAncestorContainer;
  const node = ancestor.nodeType === Node.TEXT_NODE ? ancestor.parentElement : ancestor as HTMLElement;
  if (node?.closest('pre, code')) return null;

  // Get context
  const container = range.startContainer;
  const fullText = container.textContent || '';
  const startOffset = range.startOffset;
  const endContainer = range.endContainer;
  const endText = endContainer.textContent || '';
  const endOffset = range.endOffset;

  const prefixStart = Math.max(0, startOffset - 30);
  const prefixContext = fullText.slice(prefixStart, startOffset);
  const suffixContext = endText.slice(endOffset, endOffset + 30);

  return { anchorText: text, prefixContext, suffixContext };
}
