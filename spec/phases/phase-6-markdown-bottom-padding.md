# Phase 6: Add Bottom Padding to Markdown Content

**Status:** PENDING
**Branch:** `feat/markdown-bottom-padding`
**Release:** v1.5.0

## Feature

Add approximately 100px of extra padding-bottom to the `.doc` container in MarkdownViewer.svelte so content does not visually touch the OS taskbar edge when scrolled to the bottom. This ensures the last content block has comfortable breathing room from the bottom of the viewport.

## Analysis

The `.doc-inner` element already has `padding: 36px 32px 140px`. However, the ControlStrip floats at the bottom of the viewport (fixed position, `bottom: 0`) and can obscure content when scrolled to the very end. The `.doc` container (the scrollable parent) has no bottom padding of its own.

The fix adds `padding-bottom` to `.doc` itself. Since `.doc` uses `display: flex` with `overflow-y: auto`, padding on the scroll container is respected by browsers for scroll extent calculation.

---

## Tasks

### Task 6.1: Add padding-bottom to `.doc` container

**Files:** `frontend/src/lib/components/MarkdownViewer.svelte`

In the `<style>` section, add `padding-bottom: 100px;` to the `.doc` class:

```css
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
  padding-bottom: 100px;
}
```

**Acceptance:**
- [ ] `.doc` has `padding-bottom: 100px`
- [ ] Content no longer visually touches the bottom viewport edge
- [ ] Scrolling to the bottom reveals ~100px of empty space below content
- [ ] Auto-scroll during streaming still works (scrolls to bottom correctly)
- [ ] No horizontal layout shift
- [ ] `wails build` succeeds

---

## File Summary

| File | Change |
|------|--------|
| `frontend/src/lib/components/MarkdownViewer.svelte` | Add `padding-bottom: 100px` to `.doc` class |

## Verification

1. Open a long markdown file
2. Scroll to the very bottom -> ~100px of space below the last content element
3. ControlStrip hovers above the padding area (not overlapping content)
4. Start an AI stream -> auto-scroll still follows content to the bottom
5. Test in both light and dark modes
