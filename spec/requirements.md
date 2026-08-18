# Requirements — LLM Streaming

## Problem Statement

ais is a markdown reading surface. Users currently view static markdown files. The next evolution is to generate markdown content directly within ais by streaming responses from Claude (Anthropic API). The streamed content must render in real-time with the same quality as a static document — no flicker, no layout shift, no degraded reading experience.

This is the first network-connected feature in ais. Every prior feature operates entirely offline against the local filesystem.

---

## Stakeholders

| Stakeholder | Interest |
|-------------|----------|
| End user | Stream AI-generated markdown into the reading surface with minimal latency and high rendering quality |
| Developer | Clean internal/llm package with testable boundaries, no coupling to Wails runtime internals |
| Security reviewer | API key handling that never leaks credentials to config.json or frontend logs |

---

## Functional Requirements

### FR-1: Stream Initiation

The user can initiate an LLM stream from the Command Palette via an "Ask AI" command. The user provides a text prompt. The application opens a new stream tab and begins streaming the response.

**Acceptance criteria:**
- Command Palette includes an "Ask AI" action
- Prompt input accepts free-form text
- A new tab opens with `type: 'stream'` before the first chunk arrives
- The tab title uses the format "AI: {first 30 characters}..." (per Design.md)

---

### FR-2: Real-Time Streaming

The application streams Claude API responses and renders markdown incrementally as chunks arrive.

**Acceptance criteria:**
- Chunks arrive via Wails event `llm:chunk` at batched 50ms intervals
- The MarkdownViewer renders accumulated content progressively
- DOM updates use morphdom for incremental diffing — no full re-render per chunk
- Content already rendered does not re-flow when new content arrives

---

### FR-3: Stream Cancellation

The user can cancel an active stream at any time.

**Acceptance criteria:**
- ControlStrip shows a stop button when a stream is active
- Clicking the stop button calls `CancelStream()` and terminates the HTTP connection
- Partial content remains visible in the tab after cancellation
- The tab transitions from `streamActive: true` to `streamActive: false`

---

### FR-4: Stream Completion

When the stream finishes, the application transitions the tab to a completed state.

**Acceptance criteria:**
- `llm:done` event fires with final token count
- The stream tab's `streamActive` flag becomes `false`
- The streaming border indicator stops
- The tab content is final and behaves identically to a static file tab

---

### FR-5: Error Handling

Stream errors are surfaced to the user without disrupting the reading experience.

**Acceptance criteria:**
- `llm:error` event carries an error message
- The error appears as an inline notice in the stream tab content area
- The tab transitions to inactive state
- Partial content (if any) remains visible
- Network errors, API errors (rate limit, auth failure), and cancellation are distinguishable

---

### FR-6: API Key Management

The user can configure their Anthropic API key via the Settings panel.

**Acceptance criteria:**
- SettingsPanel includes an API key input field (masked by default, toggleable visibility)
- `SetAPIKey(key)` stores the key in OS keychain (via go-keyring) with fallback to `~/.config/ais/credentials.json` (permissions `0600`)
- `HasAPIKey()` returns whether a key is configured
- The API key is NEVER stored in `config.json`
- The API key is NEVER sent to the frontend — only a boolean "configured" status

---

### FR-7: Model Selection

The user can select which Claude model to use for streaming.

**Acceptance criteria:**
- SettingsPanel includes a model selector dropdown
- `GetAvailableModels()` returns the supported model list
- Default model: `claude-sonnet-5` (quality-first default per Design.md; users can downshift to Haiku)
- Available models: `claude-haiku-4-5`, `claude-sonnet-5`, `claude-opus-5`
- Selected model persists in config

---

### FR-8: Streaming Visual Indicators

The UI communicates streaming state through ambient visual cues aligned with the Ambient Intuition Design philosophy.

**Acceptance criteria:**
- Active stream: subtle pulse on document border (1px accent glow, per Design.md)
- Streaming caret: 2px wide, 1.2em tall, step-end blink at 1s, fades out 200ms on completion (per Design.md)
- Code block streaming state: left border accent (2px, `--stream-glow` at 0.3 opacity) while fence is open, fades 120ms on close (per Design.md)
- Auto-scroll follows bottom; pauses on manual scroll-up; "Resume following" glass pill appears at bottom, resumes on click or scroll-to-bottom (per Design.md)
- Cancelled stream: "Stopped" status label below content (per Design.md)
- Completed stream: border returns to default
- Error: border flashes once, inline text with colored dot prefix (per Design.md)
- No spinners, no progress bars, no loading text
- Escape key priority chain: Command Palette > active stream > navigation panel (per Design.md)
- Screen reader: `aria-busy="true"` during stream, `aria-live="polite"` for new content

---

## Non-Functional Requirements

### NFR-1: Rendering Performance

| Metric | Target |
|--------|--------|
| Re-render latency for 50KB accumulated content | < 5ms |
| Event rate from backend to frontend | <= 20 events/second (50ms batch interval) |
| Time from chunk arrival to DOM update | < 16ms (single frame at 60fps) |
| Memory overhead per stream tab | < 10MB for documents up to 100KB |

---

### NFR-2: Network Efficiency

| Metric | Target |
|--------|--------|
| Backend chunk batching interval | 50ms |
| Connection protocol | HTTPS only (enforced by Anthropic API) |
| Cancellation latency | < 100ms from user action to HTTP connection close |

---

### NFR-3: Security

| Requirement | Detail |
|-------------|--------|
| API key storage | OS keychain preferred, fallback file at `0600` permissions |
| API key in config.json | NEVER — config.json is `0644` (world-readable) |
| API key in frontend | NEVER — only boolean status crosses the Wails bridge |
| Network consent | User is informed that prompts are sent to Anthropic's API before first use |
| Transport | HTTPS only |

---

### NFR-4: Reliability

| Requirement | Detail |
|-------------|--------|
| Partial content on error | Preserved and visible |
| Partial content on cancel | Preserved and visible |
| Concurrent streams | Not supported — one active stream at a time. Second attempt returns error |
| Backend crash isolation | LLM errors do not crash the application or affect file viewing |

---

### NFR-5: Accessibility

| Requirement | Detail |
|-------------|--------|
| Streaming content | `aria-live="polite"` on the content region |
| Stream active | `aria-busy="true"` on the viewer |
| Screen reader announcement on start | "Receiving content" |
| Screen reader announcement on complete | "Content complete, {n} sections" |
| Stop button | Keyboard accessible, labeled "Stop streaming" |
| Reduced motion | Streaming border glow disabled when `prefers-reduced-motion: reduce` |

---

## Constraints

| Constraint | Rationale |
|------------|-----------|
| Single active stream | Simplifies state management; multiple concurrent streams add complexity with low user value |
| No conversation history | v1 is single-shot prompt-to-response; multi-turn conversation is a future feature |
| No prompt editing after send | Once a stream starts, the prompt is immutable |
| morphdom for DOM diffing | Design.md mandates incremental DOM mutations for streaming; morphdom is the lightest option |
| anthropic-sdk-go for API client | Official SDK; avoids hand-rolling SSE parsing |
| go-keyring for OS keychain | Cross-platform keychain access (macOS Keychain, Linux Secret Service, Windows Credential Manager) |

---

## Out of Scope

- Multi-turn conversation (chat history)
- Prompt templates or system prompts
- File context injection (sending file contents as context)
- Token usage tracking or cost estimation UI
- Session persistence to disk (Design.md describes this as future)
- Side-by-side comparison of streams
- Custom API endpoints or self-hosted models
- Image generation or non-text modalities
