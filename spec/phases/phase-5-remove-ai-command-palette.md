# Phase 5: Remove AI Tab from Command Palette

**Status:** PENDING
**Branch:** `feat/remove-ai-palette`
**Release:** v1.5.0

## Feature

Remove the AI prompt/chat tab from the Command Palette (Ctrl+K). The AI tab does not add sufficient value to justify its presence in the command palette. All other AI features remain intact: translation (T key, Ctrl+Space, TranslationPopover, Settings > Translation), streaming, model selection in Settings, API key management.

## Scope Boundary

**REMOVE:** The "AI" category tab in the Command Palette, the prompt textarea, model pill selector, and all associated state/functions within CommandPalette.svelte.

**DO NOT TOUCH:**
- Translation feature (T key, Ctrl+Space, TranslateText binding, TranslationPopover, Settings > Translation)
- LLM streaming (StartStream, CancelStream, llm:chunk/done/error events)
- AI settings in SettingsPanel (API key, model selector, provider)
- Stream stores (stream.ts)

---

## Tasks

### Task 5.1: Remove AI category from categories array

**Files:** `frontend/src/lib/components/CommandPalette.svelte`

Remove the `{ key: 'ai', label: 'AI' }` entry from the `categories` array (line 291).

**Acceptance:**
- [ ] "AI" tab no longer appears in the Command Palette tab bar
- [ ] All, Docs, Search, Commands tabs remain
- [ ] Tab switching still works

---

### Task 5.2: Remove AI-related state variables

**Files:** `frontend/src/lib/components/CommandPalette.svelte`

Remove these state declarations:
- `aiPrompt` (line 16)
- `selectedModel` (line 17)
- `hasApiKey` (line 18)
- `currentProvider` (line 19)
- `promptEl` (line 12 — the textarea ref)
- `modelOptions` array (lines 25-29)

Remove the `hasApiKey` check in the `$effect` that runs on open (lines 106-114) — keep the rest of the effect (query reset, selectedIdx reset, category reset, searchResults reset).

Remove the `$effect` that focuses the prompt textarea when switching to AI tab (lines 117-123 — the `setPrompting()` call and `promptEl?.focus()`).

**Acceptance:**
- [ ] No AI-related state variables remain
- [ ] Opening the palette still works (query resets, focus on input)
- [ ] No TypeScript/build errors

---

### Task 5.3: Remove AI prompt UI template

**Files:** `frontend/src/lib/components/CommandPalette.svelte`

Remove the AI template block:
- The `{#if activeCategory === 'ai'}` block in the input section (lines 307-316) — the readonly input placeholder
- The `{#if activeCategory === 'ai'}` block in the results section (lines 339-355) — the textarea, model pill, provider display
- The AI-specific footer section (lines 423-425) — "send", "newline", "close" hints

Remove the `submitAIPrompt` function (lines 174-188).
Remove the `cycleModel` function (lines 191-195).

**Acceptance:**
- [ ] No AI prompt textarea in Command Palette
- [ ] No model pill or provider display
- [ ] Footer shows correct hints for remaining categories
- [ ] Command palette opens and works with docs, search, commands tabs

---

### Task 5.4: Remove AI-related CSS

**Files:** `frontend/src/lib/components/CommandPalette.svelte`

Remove unused CSS classes:
- `.pal-ai` (line 630)
- `.pal-ai-textarea` (line 634)
- `.pal-ai-textarea::placeholder` (line 650)
- `.pal-ai-bar` (line 654)
- `.pal-ai-provider` (line 660)
- `.pal-model-pill` (line 665)
- `.pal-model-pill:hover` (line 675)
- `.pal-ai-nokey` (line 682)
- `.pal-ai-dot` (line 689)
- `.pal-ai-dot.warning` (line 696)
- `.pal-ai-msg` (line 700)
- `.pal-ai-hint` (line 706)

**Acceptance:**
- [ ] No unused AI-related CSS classes remain
- [ ] No visual regressions in remaining palette features

---

### Task 5.5: Clean up unused imports

**Files:** `frontend/src/lib/components/CommandPalette.svelte`

Remove imports that are now unused:
- `openStreamTab` from `'../stores/tabs'` (only used by `submitAIPrompt`)
- `setPrompting`, `startStreamSession`, `clearStream` from `'../stores/stream'` — BUT check: `clearStream` is used in `close()` function (line 172). Only remove `setPrompting` and `startStreamSession` if they are solely used for the AI tab.
- Keep `clearStream` if it is still used in the `close()` function.

Update the `close()` function: remove the `if (activeCategory !== 'ai') { clearStream(); }` conditional — now `clearStream()` is always called, or remove it entirely if the stream is no longer started from the palette.

Since streams are no longer started from the palette, `clearStream` in `close()` is a no-op safety call. It can remain for defensive cleanup or be removed.

**Acceptance:**
- [ ] No unused imports
- [ ] `close()` function still works (no runtime errors)
- [ ] `npm run check` passes (no TypeScript errors)

---

## File Summary

| File | Change |
|------|--------|
| `frontend/src/lib/components/CommandPalette.svelte` | Remove AI tab, state, template, CSS, and clean up imports |

## Verification

1. Open Command Palette (Ctrl+K) -> 4 tabs: All, Docs, Search, Commands (no AI)
2. Search for a document -> works
3. Execute a command -> works
4. Cross-file search (Ctrl+Shift+F) -> works
5. Translation (T key on selection) -> still works (untouched)
6. AI settings in Settings panel -> still works (untouched)
7. `npm run check` -> passes
8. `wails build` -> succeeds
