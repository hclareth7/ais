# Phase 3: Translation on Highlight

**Status:** PENDING
**Branch:** `feat/translation`
**Release:** v1.4.0

## Feature

When text is selected in a document and the user presses `T`, the selected text is translated to the configured default language using AI (Claude API). Pressing `T` again cycles to the next language in the configured list. Translation must be as fast as possible — no thinking, no streaming, just a direct synchronous API call returning the translated text.

A keyboard shortcut (Ctrl+Space) allows quick switching between the 2 most recently used target languages, similar to Win+Space for OS language switching.

## UX Flow

1. User selects text → Quick Action Bar appears (existing: color dots for highlights)
2. User presses `T` → translation popover appears below the Quick Action Bar showing translated text in the default language
3. User presses `T` again → popover updates with the next language in the configured list
4. User presses `T` again → cycles to the next language, and so on (wraps around)
5. User clicks away or presses Escape → popover closes
6. `Ctrl+Space` → switches the default target language between the 2 most recently used languages (shown briefly as a toast/indicator)

## Configuration

- `translationLanguages`: ordered list of language codes, e.g. `["es", "en", "fr", "pt", "de"]`
- `translationDefaultIndex`: index into the list for the current default (0-based)
- Default: `["es", "en"]` with index `0` (Spanish)

---

## Tasks

### Task 3.1: Add translation config fields

**Files:** `internal/config/config.go`, `internal/config/defaults.go`

Add to `Config` struct:
- `TranslationLanguages []string` (`json:"translationLanguages"`)
- `TranslationDefaultIndex int` (`json:"translationDefaultIndex"`)

Add defaults in `DefaultConfig()`:
- `TranslationLanguages: []string{"es", "en"}`
- `TranslationDefaultIndex: 0`

Add zero-value guards in `Load()` (if `TranslationLanguages` is empty, set from defaults).

**Acceptance:**
- [ ] Config loads/saves with new fields
- [ ] Default is `["es", "en"]` with index 0
- [ ] `go test ./internal/...` passes

---

### Task 3.2: Add `TranslateText` method to LLM client

**Files:** `internal/llm/client.go`

Add a non-streaming `Translate(ctx, text, targetLang string) (string, error)` method to `Client`:
- Uses `anthropic.MessageNew` (NOT streaming) for minimal latency
- System prompt: `"Translate the following text to {language}. Return only the translation, nothing else."`
- User message: the selected text
- Model: Haiku (fastest, cheapest) — hardcoded, not user-selectable for translations
- `MaxTokens`: 1024 (translations are short)
- Returns the translated text string directly
- Returns `*StreamError` on failure (reuses existing error classification)

**Acceptance:**
- [ ] Method compiles and returns translated text
- [ ] Uses Haiku model regardless of user's selected model
- [ ] No streaming, no thinking — single request/response
- [ ] Unit test with mock server

---

### Task 3.3: Add `TranslateText` Wails binding

**Files:** `app.go`, `frontend/wailsjs/go/main/App.js`, `App.d.ts`

Add to `App`:
```go
func (a *App) TranslateText(text string, targetLang string) (string, error)
```

- **Security: validate `targetLang` against `config.TranslationLanguages`** — reject any language code not in the configured list to prevent prompt injection via crafted language strings
- Validates API key exists (returns error if not)
- Validates text is non-empty and under 5000 chars (reasonable translation limit)
- Creates an LLM client (Haiku, non-streaming)
- Calls `client.Translate(ctx, text, targetLang)`
- Returns the translated text

Add Wails bindings manually in `App.js` and `App.d.ts`.

**Acceptance:**
- [ ] Binding callable from frontend
- [ ] Returns error if `targetLang` is not in configured languages list
- [ ] Returns error if no API key configured
- [ ] Returns error if text is empty or exceeds 5000 chars
- [ ] Returns translated text on success

---

### Task 3.4: Add `TranslationPopover.svelte` component

**Files:** `frontend/src/lib/components/TranslationPopover.svelte` (new)

Floating popover that shows translated text:
- Props: `visible`, `text`, `language`, `loading`, `position: {x, y}`, `onclose`
- Glass surface styling (same as Quick Action Bar: `backdrop-filter`, `var(--surface-elevated)`)
- Shows language label (e.g. "ES") as a small pill/badge
- Shows translated text (max-width 400px, word-wrap)
- Loading state: subtle shimmer or "..." placeholder
- `position: fixed`, positioned below the Quick Action Bar
- `z-index: 51` (above Quick Action Bar at 50)
- Escape closes, click outside closes
- `aria-live="polite"` for screen readers

**Acceptance:**
- [ ] Renders translated text with language badge
- [ ] Shows loading state while API call is in progress
- [ ] Glass surface styling matching design system
- [ ] Accessible (ARIA, keyboard dismissible)

---

### Task 3.5: Wire `T` key to translate selected text

**Files:** `frontend/src/lib/components/MarkdownViewer.svelte`

Add keydown handler for `T` key when text is selected (`showQuickAction === true`):
- On first press: call `TranslateText(selectedText, defaultLanguage)`, show `TranslationPopover` in loading state, then update with result
- On subsequent presses: advance to next language in `translationLanguages` list (wrapping), call translate again
- Track `currentTranslationIndex` to cycle through languages
- Reset index when selection changes or popover closes
- Cache translations per language to avoid redundant API calls for the same text
- Use `cachedSelection.anchorText` for the translation source text

**Acceptance:**
- [ ] First `T` press shows translation popover with default language
- [ ] Each subsequent `T` press cycles to next language
- [ ] Translations are cached per language for current selection
- [ ] New selection resets the cycle
- [ ] Loading state shown during API call

---

### Task 3.6: Add language configuration UI in Settings

**Files:** `frontend/src/lib/components/SettingsPanel.svelte`

Add "Translation" section to Settings panel:
- Show ordered list of languages with drag-to-reorder or up/down buttons
- Add language input (text field + add button)
- Remove button (X) on each language
- Highlight the current default language
- Save to config via `UpdateConfig`
- Language labels: show both code and full name (e.g. "ES — Español")

**Acceptance:**
- [ ] User can add/remove/reorder languages
- [ ] Default language is visually indicated
- [ ] Changes persist to config
- [ ] At least 1 language must remain (prevent empty list)

---

### Task 3.7: Add `Ctrl+Space` language switch shortcut

**Files:** `frontend/src/App.svelte`, `frontend/src/lib/stores/ui.ts`

Add `Ctrl+Space` handler in `App.svelte`:
- Switches `translationDefaultIndex` between current and previously used language
- Track `previousTranslationIndex` in a store or local state
- Show a brief toast/indicator (e.g. "ES → EN" for 1.5 seconds) — reuse existing toast pattern or create minimal one
- Update config via `UpdateConfig`
- Does NOT conflict with OS shortcuts (Win+Space is OS-level, Ctrl+Space is app-level)

**Acceptance:**
- [ ] Ctrl+Space toggles between 2 most recent languages
- [ ] Brief visual indicator shows the switch
- [ ] Config is updated
- [ ] Does not conflict with OS shortcuts

---

### Task 3.8: Update Wails binding types

**Files:** `frontend/wailsjs/go/models.ts`

Update `config.Config` class to include:
- `translationLanguages: string[]`
- `translationDefaultIndex: number`

**Acceptance:**
- [ ] TypeScript types match Go struct
- [ ] Frontend can read/write translation config

---

## File summary

| File | Change |
|------|--------|
| `internal/config/config.go` | Add translation config fields + load guards |
| `internal/config/defaults.go` | Add translation defaults |
| `internal/llm/client.go` | Add `Translate()` non-streaming method |
| `app.go` | Add `TranslateText` Wails binding |
| `frontend/src/lib/components/TranslationPopover.svelte` | NEW — translation popover |
| `frontend/src/lib/components/MarkdownViewer.svelte` | Wire `T` key, show popover, cycle languages |
| `frontend/src/lib/components/SettingsPanel.svelte` | Add language configuration UI |
| `frontend/src/App.svelte` | Add `Ctrl+Space` handler |
| `frontend/src/lib/stores/ui.ts` | Add translation language state/store |
| `frontend/wailsjs/go/main/App.js` | Add `TranslateText` binding |
| `frontend/wailsjs/go/main/App.d.ts` | Add `TranslateText` type |
| `frontend/wailsjs/go/models.ts` | Update Config type |

## Verification

1. Select text, press `T` → translation popover appears with default language
2. Press `T` again → cycles to next language
3. `Ctrl+Space` → switches default language, shows brief indicator
4. Settings → add/remove/reorder languages, changes persist
5. No API key → error message in popover
6. Build on all 3 platforms, test all shortcuts
