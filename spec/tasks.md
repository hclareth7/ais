# Tasks — LLM Streaming

All tasks trace to requirements in `spec/requirements.md`. Tasks are ordered by dependency — each task builds on the previous.

---

## Task 1: internal/llm package — Go backend

**Traces to:** FR-1, FR-2, FR-4, FR-5, NFR-2

**Files to create:**
- `internal/llm/client.go`
- `internal/llm/types.go`
- `internal/llm/keystore.go`

### 1.1 Types (types.go)

Define the data types for streaming communication.

```
StreamRequest {
    Prompt string
    Model  string
}

StreamChunk {
    Text        string
    Done        bool
    TotalTokens int
}

StreamError {
    Code    string
    Message string
}
```

**Acceptance criteria:**
- All types have JSON tags
- StreamChunk is the payload for `llm:chunk` events
- StreamError distinguishes network, auth, rate-limit, and cancellation errors
- Types are in a separate file from client logic for clean imports

---

### 1.2 Keystore (keystore.go)

Implement API key storage with the resolution chain: env var > OS keychain > credentials file.

**Acceptance criteria:**
- `GetAPIKey() (string, error)` — resolves key from the chain in priority order
- `SetAPIKey(key string) error` — stores in OS keychain, falls back to `~/.config/ais/credentials.json` with `0600` permissions
- `HasAPIKey() bool` — returns true if any source in the chain has a key
- `DeleteAPIKey() error` — removes from all storage locations
- Resolution order: `ANTHROPIC_API_KEY` env > OS keychain (go-keyring, service: "ais", user: "api-key") > `~/.config/ais/credentials.json`
- Credentials file format: `{"anthropic_api_key": "sk-..."}` with `0600` permissions
- Key is NEVER logged, NEVER returned to frontend

**New dependency:** `github.com/zalando/go-keyring`

---

### 1.3 Client (client.go)

Implement the streaming HTTP client using `anthropic-sdk-go`.

**Acceptance criteria:**
- `NewClient(apiKey string) *Client` — creates client with API key
- `Stream(ctx context.Context, req StreamRequest, emit func(StreamChunk)) error` — streams response, calls emit callback for each batched chunk
- Uses `context.WithCancel` for cancellation support
- Batches SSE events at 50ms intervals before calling emit
- Emits a final `StreamChunk{Done: true, TotalTokens: N}` on completion
- Returns `StreamError` on failure with appropriate error code
- Default model: `claude-sonnet-5`
- Max tokens: 4096

**New dependency:** `github.com/anthropics/anthropic-sdk-go`

---

## Task 2: Wails bindings and App integration

**Traces to:** FR-1, FR-3, FR-6, FR-7

**Files to modify:**
- `app.go` — add new methods and fields

### 2.1 App struct extension

Add fields for LLM state management.

```go
type App struct {
    // ... existing fields ...
    llmClient    *llm.Client
    streamCancel context.CancelFunc  // nil when no stream active
    streamMu     sync.Mutex          // guards streamCancel
}
```

**Acceptance criteria:**
- `streamCancel` is `nil` when no stream is active
- `streamMu` protects concurrent access to `streamCancel`
- LLM client is initialized lazily on first `StartStream` call (not at startup)

---

### 2.2 Bound methods

| Method | Signature | Behavior |
|--------|-----------|----------|
| `StartStream` | `(prompt string) error` | Resolves API key, creates client if needed, starts goroutine that streams and emits events. Returns error if no API key or stream already active |
| `CancelStream` | `() error` | Calls `streamCancel()`, sets it to nil. Returns error if no active stream |
| `SetAPIKey` | `(key string) error` | Delegates to `keystore.SetAPIKey`. Invalidates cached client |
| `HasAPIKey` | `() bool` | Delegates to `keystore.HasAPIKey` |
| `GetAvailableModels` | `() []string` | Returns `["claude-haiku-4-5", "claude-sonnet-5", "claude-opus-5"]` |

**Acceptance criteria:**
- `StartStream` returns error if a stream is already active (single-stream constraint)
- `StartStream` emits `llm:chunk`, `llm:done`, `llm:error` events via `wailsRuntime.EventsEmit`
- `CancelStream` is safe to call when no stream is active (returns descriptive error, no panic)
- `SetAPIKey` invalidates any cached `llmClient` so the next stream uses the new key
- `HasAPIKey` never exposes the key value — only boolean status

---

## Task 3: Stream store and Tab extension (frontend)

**Traces to:** FR-1, FR-2, FR-3, FR-4, FR-8

**Files to create:**
- `frontend/src/lib/stores/stream.ts`

**Files to modify:**
- `frontend/src/lib/stores/tabs.ts` — extend Tab interface

### 3.1 Extended Tab interface

```typescript
export interface Tab {
    id: string;
    name: string;
    path: string;
    content: string;
    scrollPos: number;
    type: 'file' | 'stream';           // NEW — default 'file'
    streamActive?: boolean;             // NEW — true while streaming
}
```

**Acceptance criteria:**
- All existing tabs default to `type: 'file'`
- Stream tabs use `type: 'stream'` with a generated ID (not a file path)
- `openTab` continues to work for file tabs unchanged
- New `openStreamTab(prompt: string): string` function returns the tab ID

---

### 3.2 Stream store (stream.ts)

The stream store uses a state enum to match the 6-state machine defined in `spec/ux.md`:

```typescript
type StreamState = 'idle' | 'prompting' | 'streaming' | 'complete' | 'cancelled' | 'error';

streamContent: writable<string>('')              // Accumulated markdown text
streamState: writable<StreamState>('idle')       // Current state in the lifecycle
streamError: writable<StreamError | null>(null)  // Error details if state === 'error'
```

**Acceptance criteria:**
- `startStream(prompt: string)` — transitions to `'streaming'`, calls `StartStream`, opens stream tab, registers event listeners
- `cancelStream()` — transitions to `'cancelled'`, calls `CancelStream`
- Event listeners: `llm:chunk` appends text to `streamContent`, `llm:done` transitions to `'complete'`, `llm:error` transitions to `'error'` and sets `streamError`
- `streamContent` resets to `''` on each new stream
- Cleanup: event listeners are removed on terminal states (`complete`, `cancelled`, `error`)
- Derived convenience: `streamActive` derived as `streamState === 'streaming'`
- The `'prompting'` state is set when the Command Palette enters prompt input mode (before submission)

---

## Task 4: Incremental rendering with morphdom

**Traces to:** FR-2, NFR-1

**Files to modify:**
- `frontend/src/lib/components/MarkdownViewer.svelte`

**New dependency:** `morphdom` (npm)

### 4.1 morphdom integration

**Acceptance criteria:**
- When `tab.type === 'stream'` and `tab.streamActive === true`, the viewer uses morphdom to diff and patch the DOM instead of replacing innerHTML
- morphdom receives the full rendered HTML on each chunk and applies minimal mutations
- Re-render of 50KB content completes in < 5ms (NFR-1)
- When `streamActive === false`, the viewer reverts to standard innerHTML assignment (no morphdom overhead for static content)

### 4.2 Auto-scroll and "Resume following" pill

Per Design.md, scroll behavior during streaming:

**Acceptance criteria:**
- If the user is at the bottom (within 50px of scroll end), new content auto-scrolls into view
- If the user scrolls up manually, the scroll position locks — new content arrives below without disturbing the viewport
- When scroll is locked (user scrolled up), a "Resume following" glass pill appears at the bottom of the viewer
- Clicking the pill or scrolling back to the bottom re-enables auto-follow
- The pill has ARIA label "Resume following new content" and is keyboard-accessible
- The pill uses the glass aesthetic: translucent background, rounded, subtle border

### 4.3 Streaming caret

Per Design.md, a blinking caret marks the insertion point during streaming:

**Acceptance criteria:**
- A caret element (2px wide, 1.2em tall) appears at the end of the last rendered content
- The caret blinks with a step-end animation at 1s interval
- On stream completion or cancellation, the caret fades out over 200ms
- The caret is managed by morphdom — it must not interfere with DOM diffing (excluded from diff via `onBeforeElUpdated` skip)
- CSS uses `--stream-caret` custom property for color

### 4.4 Code block streaming accent

Per Design.md, open code blocks during streaming show a visual accent:

**Acceptance criteria:**
- While a fenced code block is open (opening fence received, closing fence not yet received), the code block has a 2px left border using `--stream-glow` at 0.3 opacity
- When the closing fence arrives and the block is complete, the accent fades out over 120ms
- Detection of open/incomplete code blocks requires tracking fence state in the accumulated markdown

### 4.5 Streaming visual indicators

**Acceptance criteria:**
- Active stream: CSS class `streaming` on the viewer container adds a subtle border pulse (`1px accent glow`, 2s ease-in-out, opacity 0.15-0.55, per Design.md)
- Completed: class removed, border returns to default (200ms fade-out on tab indicator)
- Cancelled: "Stopped" status label rendered below content (per Design.md)
- Error: brief flash class, then inline error text with 6px colored dot prefix (per Design.md, 80ms fade-in)
- `aria-busy="true"` set during active stream
- `aria-live="polite"` on the content container
- All animations disabled under `prefers-reduced-motion: reduce`

---

## Task 5: Command Palette and ControlStrip integration

**Traces to:** FR-1, FR-3, FR-8

**Files to modify:**
- `frontend/src/lib/components/CommandPalette.svelte`
- `frontend/src/lib/components/ControlStrip.svelte`

### 5.1 Command Palette — "Ask AI" command

Per Design.md, the Command Palette gains a 4th category tab ("AI") with a dedicated prompt mode:

**Acceptance criteria:**
- A 4th "AI" category tab appears in the Command Palette tab strip
- Selecting "AI" transitions the palette to a prompt textarea mode (replaces the results list)
- The prompt textarea supports multi-line input (`Shift+Enter` for newlines, `Enter` to submit)
- An inline model selector pill shows the currently selected model (e.g., "claude-sonnet-5"), clickable to cycle models
- Submitting the prompt calls `startStream(prompt)` from the stream store and closes the palette
- If no API key is configured, show a message directing to Settings
- Command is keyboard accessible

### 5.2 ControlStrip — Stop button

Per Design.md, the stop button is a 34px round button with a filled square icon:

**Acceptance criteria:**
- When `streamActive === true`, the ControlStrip shows a 34px round stop button (after the zoom group)
- The button displays a filled square icon (stop symbol)
- The button calls `cancelStream()`
- Hover state uses danger-dim color treatment
- The button disappears when streaming ends (complete, error, or cancel)
- Button has `aria-label="Stop streaming"`
- Keyboard accessible (Tab navigable)
- `Escape` key cancels the active stream, following the priority chain: Command Palette > active stream > navigation panel (per Design.md)

---

## Task 6: Settings panel — API key and model selector

**Traces to:** FR-6, FR-7

**Files to modify:**
- `frontend/src/lib/components/SettingsPanel.svelte`

### 6.1 API key configuration

**Acceptance criteria:**
- New section in SettingsPanel: "AI Configuration"
- API key input: masked by default (`type="password"`), toggle to reveal
- Save button calls `SetAPIKey(key)` via Wails binding
- Status indicator shows "Configured" (green) or "Not configured" (neutral) via `HasAPIKey()`
- Clear button to remove the API key
- Help text: "Your API key is stored securely in your OS keychain. It is never saved to the config file."

### 6.2 Model selector

Per Design.md, the model selector uses a segmented control (Haiku | Sonnet | Opus):

**Acceptance criteria:**
- Segmented control with three options: Haiku, Sonnet, Opus (per Design.md)
- Selected model persists in config (new `selectedModel` field in Config)
- Default model: `claude-sonnet-5` (resolved: quality-first default per Design.md)
- Each segment shows the model tier name; full model ID used internally
- 6px status dot next to the section header (success when key configured, warning when not, per Design.md)

---

## Task 7: Error handling and edge cases

**Traces to:** FR-5, NFR-4

**Files to modify:**
- `internal/llm/client.go`
- `frontend/src/lib/stores/stream.ts`
- `frontend/src/lib/components/MarkdownViewer.svelte`

### 7.1 Backend error classification

**Acceptance criteria:**
- Network errors (DNS, timeout, connection refused) map to `StreamError{Code: "network"}`
- Auth errors (401) map to `StreamError{Code: "auth"}`
- Rate limit (429) maps to `StreamError{Code: "rate_limit"}` with retry-after if available
- Cancellation maps to `StreamError{Code: "cancelled"}`
- Other API errors map to `StreamError{Code: "api", Message: ...}`

### 7.2 Frontend error display

Per Design.md, errors use inline text with a 6px colored dot prefix. Cancellation shows a "Stopped" label.

**Acceptance criteria:**
- Error renders as inline text below content with a 6px colored dot prefix (per Design.md, 80ms fade-in)
- Four error variants per Design.md: no API key (warning dot + "Open Settings" link), network failure, rate limit, API error
- Auth errors suggest checking the API key in Settings with a clickable "Open Settings" link
- Rate limit errors show retry-after time if available
- Network errors suggest checking internet connectivity
- Cancelled streams show a "Stopped" status label below content (not an error — distinct visual treatment per Design.md)
- Error notice follows the glass aesthetic (translucent background, soft border)

### 7.3 Edge cases

**Acceptance criteria:**
- Closing a stream tab while streaming calls `CancelStream()` automatically
- Switching away from a stream tab does not stop the stream — content accumulates in background
- Application shutdown during active stream cancels cleanly (no goroutine leak)
- Empty response (zero chunks) shows a "No content received" notice
- Very long responses (>100KB) continue to render without degradation

---

## Task 8: Tests

**Traces to:** All requirements

**Files to create:**
- `internal/llm/client_test.go`
- `internal/llm/keystore_test.go`

### 8.1 Client tests (client_test.go)

**Acceptance criteria:**
- SSE mock server using `httptest` that emits a sequence of `content_block_delta` events
- Test: successful stream receives all chunks in order
- Test: cancellation mid-stream stops chunk delivery
- Test: server error (500) returns appropriate `StreamError`
- Test: auth error (401) returns `StreamError{Code: "auth"}`
- Test: rate limit (429) returns `StreamError{Code: "rate_limit"}`
- Test: empty response (no content events) returns gracefully
- Test: chunk batching — rapid events are batched at 50ms intervals

### 8.2 Keystore tests (keystore_test.go)

**Acceptance criteria:**
- Test: env var takes priority over keychain and file
- Test: keychain takes priority over file (may require mock or build tag for CI)
- Test: file fallback works when keychain unavailable
- Test: credentials file has `0600` permissions
- Test: `HasAPIKey()` returns correct boolean for each storage state
- Test: `DeleteAPIKey()` removes from all locations

---

## Dependency Graph

```
Task 1 (internal/llm)
    |
    v
Task 2 (Wails bindings)
    |
    +---> Task 3 (stream store + Tab extension)
    |         |
    |         v
    |     Task 4 (morphdom rendering)
    |         |
    |         v
    |     Task 5 (CommandPalette + ControlStrip)
    |
    +---> Task 6 (Settings panel)
    |
    v
Task 7 (Error handling) — depends on Tasks 1-6
    |
    v
Task 8 (Tests) — depends on Tasks 1-7
```

Tasks 3-5 and Task 6 can proceed in parallel once Task 2 is complete.

---

## Effort Estimates

| Task | Effort | Rationale |
|------|--------|-----------|
| Task 1: internal/llm | Medium | New package, SDK integration, SSE batching logic |
| Task 2: Wails bindings | Small | Thin layer delegating to internal/llm |
| Task 3: Stream store + Tab | Small | Store pattern is established; extending Tab is mechanical |
| Task 4: morphdom rendering | Large | DOM diffing, caret management, auto-scroll + resume pill, code block accent, scroll preservation |
| Task 5: CommandPalette + ControlStrip | Small | Adding commands and conditional button to existing components |
| Task 6: Settings panel | Small | Form inputs with Wails binding calls |
| Task 7: Error handling | Medium | Cross-cutting concern touching backend and frontend |
| Task 8: Tests | Medium | SSE mock server, keystore mocking, table-driven tests |

---

## Resolved Questions

### OQ-1: Default model — RESOLVED: Sonnet

**Decision:** `claude-sonnet-5` is the default model. The first-stream experience defines user trust in the feature. Sonnet delivers the quality that makes the user return. Users who want lower cost can switch to Haiku in the model selector. All specs updated to reflect this decision.

---

### OQ-2: Session persistence — RESOLVED: Deferred to v2

**Decision:** Stream tabs are in-memory only for v1. Content is lost on application close. Users can copy content before closing. Session persistence is a v2 addition once the core streaming experience is stable. ux.md updated by Steve to reflect the v1/v2 split.
