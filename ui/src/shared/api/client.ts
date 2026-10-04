import { parseMessage, ProtocolError, StreamCursor, StreamGapError } from './protocol.ts';
import type { HostState, Message } from './protocol.ts';

export type HostFeed = {
  connection: 'connecting' | 'connected' | 'disconnected' | 'incompatible';
  state: HostState | null;
  runId: string | null;
  sequence: number;
  error: string | null;
  lastReceivedAt: number | null;
  lastOutputReceivedAt: number | null;
};
export const initialHostFeed: HostFeed = { connection: 'connecting', state: null, runId: null, sequence: 0, error: null, lastReceivedAt: null, lastOutputReceivedAt: null };

// The optional base URL also permits testing the real client with Node's native
// fetch/WebSocket. No React or Three.js dependencies belong in this transport.
export function connectHost(onChange: (feed: HostFeed) => void, baseURL = window.location.href) {
  let feed = { ...initialHostFeed };
  let stopped = false;
  let generation = 0;
  let retryDelay = 500;
  let retryTimer: ReturnType<typeof setTimeout> | undefined;
  let socket: WebSocket | null = null;
  let controller: AbortController | null = null;
  let lastMessageAt = Date.now();
  function emit(patch: Partial<HostFeed>) { feed = { ...feed, ...patch }; onChange(feed); }
  function clearAttempt() {
    controller?.abort(); controller = null;
    if (socket) {
      socket.onclose = null; socket.onerror = null; socket.onmessage = null;
      socket.close(); socket = null;
    }
  }
  function fail(epoch: number, error: unknown) {
    if (stopped || epoch !== generation) return;
    generation++;
    clearAttempt();
    const reason = error instanceof Error ? error.message : 'Host connection lost';
    if (error instanceof ProtocolError) {
      emit({ connection: 'incompatible', error: reason, lastOutputReceivedAt: null });
      return;
    }
    emit({ connection: 'disconnected', error: reason, lastOutputReceivedAt: null });
    retryTimer = setTimeout(start, retryDelay);
    retryDelay = Math.min(retryDelay * 2, 10000);
  }
  function apply(message: Message, connected: boolean) {
    const patch: Partial<HostFeed> = { runId: message.runId, sequence: message.sequence, lastReceivedAt: Date.now() };
    if (message.type === 'snapshot' || message.type === 'state') patch.state = message.payload;
    if (message.type === 'snapshot') patch.lastOutputReceivedAt = null;
    if (message.type === 'event' && message.payload.name === 'neural.output.spike' && message.payload.schemaVersion === 1) patch.lastOutputReceivedAt = Date.now();
    if (connected) { patch.connection = 'connected'; patch.error = null; retryDelay = 500; }
    emit(patch);
  }
  async function start() {
    if (stopped) return;
    const epoch = ++generation;
    const cursor = new StreamCursor();
    const abort = new AbortController();
    controller = abort;
    emit({ connection: 'connecting', error: null });
    const timeout = setTimeout(() => abort.abort(), 5000);
    try {
      const response = await fetch(new URL('/api/v1/snapshot', baseURL), { signal: abort.signal, cache: 'no-store' });
      if (!response.ok) throw new Error(`Host snapshot returned HTTP ${response.status}`);
      const snapshot = parseMessage(await response.json());
      if (stopped || epoch !== generation) return;
      if (snapshot.type !== 'snapshot') throw new ProtocolError('Expected an HTTP snapshot');
      cursor.accept(snapshot);
      apply(snapshot, false);
      const url = new URL('/api/v1/stream', baseURL);
      url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';
      const connection = new WebSocket(url);
      socket = connection;
      lastMessageAt = Date.now();
      let first = true;
      connection.onmessage = event => {
        if (stopped || epoch !== generation) return;
        try {
          let value: unknown;
          try { value = JSON.parse(String(event.data)); }
          catch { throw new ProtocolError('Invalid host JSON'); }
          const message = parseMessage(value);
          if (first && message.type !== 'snapshot') throw new StreamGapError('Expected a stream snapshot');
          first = false;
          lastMessageAt = Date.now();
          if (cursor.accept(message)) apply(message, true);
        } catch (error) { fail(epoch, error); }
      };
      connection.onclose = () => fail(epoch, new Error('Host connection lost; reconnecting'));
      connection.onerror = () => fail(epoch, new Error('Host stream unavailable; reconnecting'));
    } catch (error) { fail(epoch, error); }
    finally { clearTimeout(timeout); }
  }
  const watchdog = setInterval(() => {
    const starting = feed.connection === 'connected' && feed.state?.status === 'starting';
    if (socket && !starting && Date.now() - lastMessageAt > 6000) fail(generation, new Error('Host updates are stale; reconnecting'));
  }, 1000);
  void start();
  return () => {
    stopped = true; generation++;
    clearTimeout(retryTimer); clearInterval(watchdog); clearAttempt();
  };
}
