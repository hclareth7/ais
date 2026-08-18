import { writable, derived, get } from 'svelte/store';

// StreamState matches the 6-state machine defined in spec/ux.md.
export type StreamState = 'idle' | 'prompting' | 'streaming' | 'complete' | 'cancelled' | 'error';

export interface StreamChunk {
  text: string;
  done: boolean;
  totalTokens?: number;
}

export interface StreamError {
  code: 'network' | 'auth' | 'rate_limit' | 'cancelled' | 'api';
  message: string;
}

export interface StreamSession {
  tabId: string;
  state: StreamState;
  content: string;
  error?: StreamError;
  totalTokens?: number;
}

// Core stores
export const activeStream = writable<StreamSession | null>(null);

// Derived convenience stores
export const streamState = derived(activeStream, ($s) => $s?.state ?? 'idle');
export const streamActive = derived(activeStream, ($s) => $s?.state === 'streaming');

export function startStreamSession(tabId: string): void {
  activeStream.set({
    tabId,
    state: 'streaming',
    content: '',
  });
}

export function setPrompting(): void {
  activeStream.set({
    tabId: '',
    state: 'prompting',
    content: '',
  });
}

export function appendStreamContent(text: string): void {
  activeStream.update(s => {
    if (!s) return s;
    return { ...s, content: s.content + text };
  });
}

export function completeStream(totalTokens: number): void {
  activeStream.update(s => {
    if (!s) return s;
    return { ...s, state: 'complete', totalTokens };
  });
}

export function cancelStreamState(): void {
  activeStream.update(s => {
    if (!s) return s;
    return { ...s, state: 'cancelled' };
  });
}

export function setStreamError(error: StreamError): void {
  activeStream.update(s => {
    if (!s) return s;
    return { ...s, state: 'error', error };
  });
}

export function clearStream(): void {
  activeStream.set(null);
}
