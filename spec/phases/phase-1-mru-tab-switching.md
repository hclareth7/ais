# Phase 1: MRU Tab Switching

**Status:** COMPLETE (v1.2.0, PR #19)
**Branch:** `feat/mru-tab-switching`

## Feature

Replicate Alt+Tab (OS) behavior for Ctrl+Tab in ais. Quick press switches between the 2 most recent tabs. Holding Ctrl shows an overlay to cycle through all tabs in MRU (Most Recently Used) order. Ctrl+Shift+Tab cycles in reverse.

## Tasks (all complete)

1. Add `mruOrder` store and `pushToMruFront()` in `tabs.ts`
2. Create `MruOverlay.svelte` — glass surface overlay with tab list
3. Wire Ctrl+Tab / Ctrl+Shift+Tab in `App.svelte` with 150ms overlay delay
4. Add `capture: true` + `e.code === 'Tab'` fallback for WebKitGTK (Linux)
5. Keyup handler confirms selection on Ctrl release, Escape cancels

## Files changed

- `frontend/src/lib/stores/tabs.ts`
- `frontend/src/lib/components/MruOverlay.svelte` (new)
- `frontend/src/App.svelte`
