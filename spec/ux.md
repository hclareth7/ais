# UX — ais LLM Streaming

## Overview

This document defines the user experience for the LLM streaming feature in ais. Every interaction described here serves the core philosophy: the content is the protagonist. Streaming AI responses must feel like reading a document that is being written for you — not like using a chatbot.

---

## User Personas

### The Engineer

A software engineer who reads technical documentation daily. Uses ais to browse project docs, runbooks, and architecture decision records. When a question arises while reading, they want to ask AI without leaving the reading surface. Context switching is the enemy.

### The Researcher

A technical writer or analyst who consumes long-form content. They want AI to generate summaries, explanations, or alternative perspectives on the material they are reading. They value accuracy and the ability to review the full response as a document — not fragmented messages.

### The Learner

Someone new to a codebase or technology. They use ais to read documentation and ask clarifying questions. They want the AI response to feel approachable and look like the rest of the content — not like a foreign interface bolted onto the reader.

---

## User Journey — Ask AI

### Phase 1: Intent

The user has a question. They are reading a document in ais and want to ask Claude something.

Trigger: `Ctrl + K` (opens Command Palette).

Mental model: "I want to ask something."

The Command Palette is the universal entry point. The user already knows it — it is how they search files, switch themes, and run commands. Adding AI here means zero new interface to learn.

---

### Phase 2: Prompt

The Command Palette opens. The user sees four category tabs:

```text
All | Docs | Commands | AI
```

The user clicks the **AI** tab (or presses `Tab` to cycle categories until AI is selected).

The search input transforms:

```text
Placeholder: "Search docs, commands..."  →  "Ask Claude..."
```

The results area below is replaced by an open text area for the prompt.

The user types their question. Multi-line is supported via `Shift + Enter`.

A small model selector pill appears in the bottom-right of the prompt area:

```text
claude-sonnet-5 v
```

The user can click the pill to cycle models, or ignore it (Sonnet is the default).

---

### Phase 3: Submit

The user presses `Enter`.

What happens next — in order:

1. The Command Palette closes (120ms fade-out)
2. A new tab appears in the tab bar with the label `AI: {prompt preview}...`
3. The tab has a subtle pulsing bottom border (the streaming indicator)
4. The viewer area clears and prepares for incoming content
5. A thin blinking caret appears at the top of the viewer area
6. Streaming begins

The entire transition from submit to first token should feel instant. No loading screen. No intermediate state. The user pressed Enter; the answer is arriving.

---

### Phase 4: Streaming

Content appears paragraph by paragraph in the viewer.

The experience is identical to reading a regular markdown document, except:

- A blinking caret marks the end of the stream
- The tab has a subtle pulsing indicator
- The ControlStrip shows a stop button

The viewer auto-scrolls to follow new content.

If the user scrolls up to re-read an earlier section, auto-scroll pauses. A "Resume following" pill appears at the bottom. Clicking it (or scrolling back to the bottom) resumes auto-scroll.

Code blocks render as plain monospace text while the fence is open. When the closing fence arrives, syntax highlighting applies in a single pass. A thin accent border on the left edge of the code block indicates it is still receiving content.

Headings populate the Table of Contents as they arrive. If the TOC panel is open, entries appear in place.

---

### Phase 5: Complete

Streaming ends.

What happens — in order:

1. The blinking caret fades out (200ms)
2. The tab pulsing indicator fades out (200ms)
3. The stop button fades out of the ControlStrip (120ms)
4. The tab becomes a regular tab — indistinguishable from a file tab

The response is now a document. The user can scroll, collapse sections, copy code blocks, open the TOC, zoom in, adjust reading width — every tool available for file documents works identically on AI responses.

The session is saved locally. Closing and reopening ais restores the tab with its content and scroll position.

---

### Phase 6: Copy / Use

The user interacts with the response as a document:

- Select and copy text
- Click headings to collapse sections
- Use the TOC to navigate
- Open another document in a new tab for comparison
- Ask another question via `Ctrl + K` > AI tab

---

## Interaction States

### State Machine

```text
idle → prompting → streaming → complete
                       ↓            ↑
                    cancelled ──────┘
                       ↓
                     error
```

### State: Idle

No active streaming. No streaming indicators visible. The application behaves exactly as it does without the LLM feature.

**Visual:**
- No streaming indicators
- No stop button in ControlStrip
- Command Palette AI tab available

---

### State: Prompting

The user has opened the Command Palette and selected the AI tab. They are composing their prompt.

**Visual:**
- Command Palette overlay visible
- AI tab active
- Prompt textarea focused
- Model selector pill visible

**Transitions:**
- `Enter` → Streaming (submit prompt)
- `Escape` → Idle (close palette, discard prompt)

---

### State: Streaming

Content is actively arriving from the API.

**Visual:**
- Stream tab active with pulsing indicator
- Blinking caret at end of content
- Stop button visible in ControlStrip
- Auto-scroll active (unless paused by user)
- "Resume following" pill visible (if auto-scroll paused)

**Transitions:**
- Stream completes naturally → Complete
- `Escape` (with no overlay open) → Cancelled
- Stop button clicked → Cancelled
- Network error → Error
- API error → Error

---

### State: Complete

Streaming has finished successfully.

**Visual:**
- Tab is a regular tab (no indicator)
- No caret
- No stop button
- Content is a complete markdown document

**Transitions:**
- `Ctrl + K` > AI tab → Prompting (new question)
- Close tab → Idle

---

### State: Cancelled

The user stopped the stream before it completed.

**Visual:**
- Tab is a regular tab (no indicator)
- No caret
- No stop button
- Content shows everything received up to cancellation
- "Stopped" label below the last content, separated by a thin border

**Transitions:**
- Same as Complete

---

### State: Error

An error occurred during streaming.

**Visual:**
- Tab is a regular tab (no indicator)
- No caret
- No stop button
- Content shows everything received before the error (may be empty)
- Error message inline below the content (or as the only content if error was immediate)
- Error messages are described in spec/Design.md under "Error States"

**Transitions:**
- Same as Complete

---

## Keyboard-Only Flow

Every step of the streaming experience is fully operable without a mouse.

### Full keyboard sequence

```text
1. Ctrl + K                    Open Command Palette
2. Tab (or Shift+Tab)          Cycle category tabs until AI is focused
3. Enter                       Activate AI tab
4. Type prompt                 Compose the question
5. Enter                       Submit prompt
6. (streaming begins)
7. Escape                      Stop streaming (optional)
8. Ctrl + W                    Close stream tab (when done)
```

### Keyboard reference — streaming context

| Key | Context | Action |
|-----|---------|--------|
| `Ctrl + K` | Any | Open Command Palette |
| `Enter` | AI prompt textarea | Submit prompt |
| `Shift + Enter` | AI prompt textarea | Insert newline |
| `Escape` | Streaming active, no overlay | Stop streaming |
| `Escape` | Command Palette open | Close palette (takes priority) |
| `Ctrl + W` | Stream tab active | Close stream tab |
| `Ctrl + Tab` | Any | Next tab |
| `Ctrl + Shift + Tab` | Any | Previous tab |
| `Home` | Streaming, viewer focused | Jump to top, pause auto-scroll |
| `End` | Streaming, viewer focused | Jump to bottom, resume auto-scroll |
| `Page Up` | Streaming, viewer focused | Scroll up, pause auto-scroll |

### Focus management

When the user submits a prompt:

```text
Focus moves from Command Palette input
    ↓
Command Palette closes
    ↓
Focus moves to the viewer area
    ↓
The viewer is focusable for keyboard scrolling
```

When streaming completes:

```text
Focus remains on the viewer
    ↓
The user can read, scroll, or press Ctrl+K again
```

No focus traps. No unexpected focus shifts. Focus follows the reading flow.

---

## Accessibility

### ARIA Attributes

The viewer area during streaming:

```html
<main
  role="article"
  aria-label="AI response"
  aria-busy="true"
  aria-live="polite"
>
  <!-- rendered content -->
</main>
```

When streaming completes, `aria-busy` changes to `false`.

The streaming tab:

```html
<div
  role="tab"
  aria-selected="true"
  aria-label="AI response: {prompt preview}, streaming"
>
```

When streaming completes, `aria-label` removes the ", streaming" suffix.

The stop button:

```html
<button
  aria-label="Stop streaming"
  title="Stop receiving response (Esc)"
>
```

The "Resume following" pill:

```html
<button
  aria-label="Resume auto-scroll"
>
  Resume following
</button>
```

---

### Screen Reader Announcements

Announcements use an `aria-live="assertive"` visually-hidden region for critical transitions and `aria-live="polite"` for content updates.

| Event | Region | Announcement |
|-------|--------|-------------|
| Prompt submitted | assertive | "Sending prompt to Claude" |
| First content arrives | polite | "Receiving response" |
| Stream complete | assertive | "Response complete, {n} sections" |
| Stream cancelled | assertive | "Response stopped" |
| Error occurred | assertive | "Error: {description}" |
| Auto-scroll paused | polite | "Auto-scroll paused" |
| Auto-scroll resumed | polite | "Auto-scroll resumed" |

Content paragraphs are NOT individually announced. The `aria-live="polite"` on the main viewer provides progressive awareness without overwhelming the screen reader with every paragraph.

---

### Reduced Motion

When `prefers-reduced-motion: reduce` is active:

- Tab streaming indicator: static opacity (0.35), no pulse animation
- Caret: static, no blink
- Auto-scroll: instant jump (no smooth scroll)
- Resume pill: instant appear/disappear (no fade)
- Error messages: instant appear (no fade)
- All 120ms/200ms transitions become 0ms

The experience is fully functional. Only animation is removed. No content or functionality is lost.

---

### Contrast

All streaming-specific elements must pass WCAG AA:

| Element | Foreground | Background | Min ratio |
|---------|-----------|------------|-----------|
| Error text | --text-secondary | --surface-solid | 4.5:1 |
| Stopped label | --text-ghost | --surface-solid | 3.0:1 (large text equivalent — decorative) |
| Resume pill text | --text-secondary | --surface-elevated | 4.5:1 |
| Caret | --stream-caret | --surface-solid | Not applicable (decorative) |
| Error dot | --danger | --surface-solid | Not applicable (decorative, redundant with text) |

---

## Error Handling UX

### Principle

Errors should feel like part of the document.

They should inform, not alarm.

They should guide, not blame.

### Error: No API Key

**When:** User submits a prompt without a configured API key.

**What the user sees:**
A new tab opens. Instead of streaming content, a quiet message appears at the top of the viewer area:

```
(warning dot) No API key configured.
Add your API key in Settings to start a conversation.
Open Settings
```

"Open Settings" is a clickable link that opens the SettingsPanel scrolled to the AI section.

**Why this approach:** The user chose to ask AI. Blocking them with a modal before they can even try would feel hostile. Instead, we open the tab (matching their expectation) and guide them gently to the configuration step.

---

### Error: Network Failure

**When:** The connection drops mid-stream.

**What the user sees:**
Content received before the failure is preserved. Below the last paragraph, an error line appears:

```
(error dot) Connection lost -- content may be incomplete.
```

No retry button. The user can close the tab and try again when their network recovers. The partial content remains available.

---

### Error: Rate Limit

**When:** The API returns HTTP 429.

**What the user sees:**

```
(error dot) Rate limit reached.
Try again in a moment.
```

No countdown. No automatic retry. The user retries when they choose.

---

### Error: Invalid API Key

**When:** The API returns HTTP 401.

**What the user sees:**

```
(error dot) API key is invalid.
Check your key in Settings.
Open Settings
```

The SettingsPanel AI section status dot changes to `--danger` (red).

---

## Session Persistence

### v1 Scope

In v1, stream tabs live for the duration of the session only.

When ais closes, stream tab content is lost. The user can copy content before closing.

This is acceptable because:

- The primary value is the immediate answer, not the archive
- The user can copy any section or the full response
- It keeps v1 shippable without a persistence layer for generated content

### v2 Target

AI responses persist exactly like file documents.

When ais closes and reopens:

```text
Restore:
  - Stream tabs with their content
  - Scroll position per stream tab
  - Tab order (stream tabs among file tabs)
  - Prompt metadata (model, timestamp)
```

Active streams do not resume after restart. If a stream was in progress when ais closed, the tab reopens with whatever content was received, in the Cancelled state with the "Stopped" label.

Sessions are stored locally per workspace. No cloud sync. No external service. The user's conversations remain on their machine.

---

## Mental Model Summary

The user should think of "Ask AI" as:

> "I open a new document that Claude writes for me."

Not as:

> "I open a chat window."

The stream tab IS a document. It behaves like a document. It looks like a document. It persists like a document. The only difference is that it was generated on demand instead of read from disk.

This mental model unification is the core UX achievement of the streaming feature. The user learns zero new interaction patterns. Everything they know about navigating, reading, and managing documents in ais applies without modification.
