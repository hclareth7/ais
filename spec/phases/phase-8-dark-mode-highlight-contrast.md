# Phase 8: Improve Highlight Color Contrast in Dark Mode

**Status:** PENDING
**Branch:** `feat/dark-mode-highlight-contrast`
**Release:** v1.5.0

## Feature

Improve the highlight/selection color contrast in dark mode. The current selection color (`--accent-dim: rgba(120, 160, 255, 0.12)`) and search match colors are too subtle against the dark background (#0B0D10), making selected text hard to read.

## Analysis

Current dark mode values with issues:

| Token | Current Value | Problem |
|-------|--------------|---------|
| `::selection` background | `var(--accent-dim)` = `rgba(120, 160, 255, 0.12)` | Nearly invisible on #0B0D10 |
| `--search-match-bg` | `rgba(255, 200, 50, 0.35)` | Acceptable but could be stronger |
| `--search-match-current-bg` | `rgba(255, 160, 0, 0.6)` | Good contrast |
| `--search-match-text` | `#FFE082` | Adequate |

The `::selection` pseudo-element is the most impactful issue. The 12% opacity blue is almost indistinguishable from the background. Highlight mark colors (`--hl-*-bg`) at 16% opacity are also quite faint.

## Design Constraints

- Must maintain the "zero presence" / "cognitive calm" philosophy from spec/Design.md
- Colors should be noticeable but not aggressive
- Light mode values should remain unchanged (they already have good contrast)
- WCAG AA requires 3:1 contrast ratio for UI components (selections are not strictly covered by WCAG, but readability matters)

---

## Tasks

### Task 8.1: Improve `::selection` color in dark mode

**Files:** `frontend/src/style.css`

Update the `::selection` pseudo-element to use a dedicated selection token instead of `--accent-dim`:

Add a new token in `:root` (dark mode):
```css
--selection-bg: rgba(120, 160, 255, 0.25);
```

Add the same token in `.light`:
```css
--selection-bg: rgba(60, 100, 220, 0.18);
```

Update the `::selection` rule:
```css
::selection {
  background-color: var(--selection-bg);
  color: var(--text-primary);
}
```

The opacity increase from 0.12 to 0.25 in dark mode doubles the visibility while remaining subtle. For light mode, 0.18 (up from the current `--accent-dim` 0.08) provides a similar improvement.

**Acceptance:**
- [ ] Selected text is clearly visible in dark mode
- [ ] Selected text is clearly visible in light mode
- [ ] Selection color is not jarring or overly bright
- [ ] Maintains design system aesthetic (subtle, calm)

---

### Task 8.2: Improve search match highlight colors in dark mode

**Files:** `frontend/src/style.css`

Update dark mode search match tokens in `:root`:
```css
--search-match-bg: rgba(255, 200, 50, 0.40);
--search-match-current-bg: rgba(255, 160, 0, 0.65);
```

These are minor bumps (0.35 -> 0.40, 0.6 -> 0.65) to improve visibility without making the highlights harsh.

**Acceptance:**
- [ ] Search matches are clearly visible in dark mode
- [ ] Current match is distinguishable from other matches
- [ ] Light mode search match colors unchanged

---

### Task 8.3: Improve highlight mark colors in dark mode

**Files:** `frontend/src/style.css`

Increase opacity of highlight mark backgrounds in dark mode (`:root` section) from 0.16 to 0.24:

```css
--hl-yellow-bg: rgba(250, 204, 21, 0.24);
--hl-green-bg: rgba(34, 197, 94, 0.24);
--hl-blue-bg: rgba(96, 165, 250, 0.24);
--hl-pink-bg: rgba(244, 114, 182, 0.24);
--hl-purple-bg: rgba(168, 85, 247, 0.24);
--hl-orange-bg: rgba(251, 146, 60, 0.24);
```

This brings dark mode highlight opacity closer to light mode (0.22-0.24) for visual consistency across themes.

**Acceptance:**
- [ ] Highlighted text is clearly visible in dark mode
- [ ] Each highlight color is distinguishable from the others
- [ ] Light mode highlight colors unchanged
- [ ] Highlights are noticeable but not distracting (cognitive calm)

---

### Task 8.4: Verify contrast with in-file search (Ctrl+F)

**Files:** No code changes (verification only)

Manual verification that the updated search match tokens work correctly with both:
- Cross-file search (Ctrl+Shift+F -> Command Palette search tab)
- In-file search (Ctrl+F -> InFileSearch component)

Both use the `.search-match` and `.search-match-current` CSS classes from style.css.

**Acceptance:**
- [ ] Ctrl+F search matches visible in dark mode
- [ ] Ctrl+Shift+F search results scroll to visible highlights
- [ ] Current match clearly distinguishable from other matches
- [ ] No color clashes between search highlights and text highlights

---

## File Summary

| File | Change |
|------|--------|
| `frontend/src/style.css` | Add `--selection-bg` token, update `::selection`, bump search match and highlight mark opacities in dark mode |

## Color Values Summary

| Token | Before (Dark) | After (Dark) | Light (unchanged) |
|-------|---------------|--------------|-------------------|
| Selection bg | `rgba(120,160,255,0.12)` via `--accent-dim` | `rgba(120,160,255,0.25)` via `--selection-bg` | `rgba(60,100,220,0.18)` via `--selection-bg` |
| Search match bg | `rgba(255,200,50,0.35)` | `rgba(255,200,50,0.40)` | `rgba(255,180,0,0.3)` |
| Search current bg | `rgba(255,160,0,0.6)` | `rgba(255,160,0,0.65)` | `rgba(255,140,0,0.5)` |
| Highlight marks | `0.16` opacity | `0.24` opacity | `0.22-0.24` |

## Verification

1. Dark mode: select text in a document -> selection clearly visible (blue tint)
2. Dark mode: Ctrl+F search -> matches highlighted in visible yellow
3. Dark mode: highlight text with each color -> all 6 colors distinguishable
4. Light mode: repeat all tests -> no regressions
5. System theme toggle -> correct colors in both modes
6. Compare before/after screenshots to confirm improvement without aggression
