export class ProtocolError extends Error {}
export class StreamGapError extends Error {}

export type HostState = {
  mode: string;
  status: string;
  brain: { neurons: number; synapses: number; tick: number; spikes: number; totalSpikes: number; outputEvents: number; lastOutputAt: string | null };
  modules: { id: string; kind: string; status: string; capabilities: string[] }[];
  entities: { id: string; kind: string; label: string }[];
  observations: { source: string; subject: string; name: string; schemaVersion: number; observedAt: string; data: unknown }[];
};
export type HostEvent = { source: string; subject: string; name: string; schemaVersion: number; data: unknown };
type Envelope = { apiVersion: 1; runId: string; sequence: number; timestamp: string };
export type Message = Envelope & (
  | { type: 'snapshot' | 'state'; payload: HostState }
  | { type: 'event'; payload: HostEvent }
  | { type: 'unknown'; wireType: string; payload: unknown }
);

function object(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== 'object' || Array.isArray(value)) throw new ProtocolError('Expected an object');
  return value as Record<string, unknown>;
}
function text(value: unknown): string {
  if (typeof value !== 'string') throw new ProtocolError('Expected a string');
  return value;
}
function count(value: unknown): number {
  if (typeof value !== 'number' || !Number.isSafeInteger(value) || value < 0) throw new ProtocolError('Expected a nonnegative safe integer');
  return value;
}
function timestamp(value: unknown): string {
  const result = text(value);
  if (!Number.isFinite(Date.parse(result))) throw new ProtocolError('Invalid timestamp');
  return result;
}
function list(value: unknown): unknown[] {
  if (!Array.isArray(value)) throw new ProtocolError('Expected an array');
  return value;
}
function state(value: unknown): HostState {
  const data = object(value);
  const brain = object(data.brain);
  return {
    mode: text(data.mode), status: text(data.status),
    brain: {
      neurons: count(brain.neurons), synapses: count(brain.synapses), tick: count(brain.tick),
      spikes: count(brain.spikes), totalSpikes: count(brain.totalSpikes), outputEvents: count(brain.outputEvents),
      lastOutputAt: brain.lastOutputAt === null ? null : timestamp(brain.lastOutputAt),
    },
    modules: list(data.modules).map(value => {
      const item = object(value);
      return { id: text(item.id), kind: text(item.kind), status: text(item.status), capabilities: list(item.capabilities).map(text) };
    }),
    entities: list(data.entities).map(value => {
      const item = object(value);
      return { id: text(item.id), kind: text(item.kind), label: text(item.label) };
    }),
    observations: list(data.observations).map(value => {
      const item = object(value);
      const observation = { source: text(item.source), subject: text(item.subject), name: text(item.name), schemaVersion: count(item.schemaVersion), observedAt: timestamp(item.observedAt), data: item.data };
      if (observation.name === 'http.response' && observation.schemaVersion === 1) parseHTTPResponse(observation.data);
      return observation;
    }),
  };
}

export function parseHTTPResponse(value: unknown): { statusCode: number; elapsedMs: number; error?: string } {
  const data = object(value);
  const statusCode = count(data.statusCode);
  if (statusCode > 599) throw new ProtocolError('Invalid HTTP status');
  return { statusCode, elapsedMs: count(data.elapsedMs), error: data.error === undefined ? undefined : text(data.error) };
}

export function parseMessage(value: unknown): Message {
  const data = object(value);
  if (data.apiVersion !== 1) throw new ProtocolError('Unsupported host API version');
  const envelope: Envelope = { apiVersion: 1, runId: text(data.runId), sequence: count(data.sequence), timestamp: timestamp(data.timestamp) };
  if (!envelope.runId) throw new ProtocolError('Missing host run ID');
  const type = text(data.type);
  if (type === 'snapshot' || type === 'state') return { ...envelope, type, payload: state(data.payload) };
  if (type === 'event') {
    const event = object(data.payload);
    const payload: HostEvent = { source: text(event.source), subject: text(event.subject), name: text(event.name), schemaVersion: count(event.schemaVersion), data: event.data };
    if (payload.name === 'neural.output.spike' && payload.schemaVersion === 1) {
      const spike = object(payload.data);
      count(spike.tick); count(spike.neuron);
    }
    return { ...envelope, type, payload };
  }
  return { ...envelope, type: 'unknown', wireType: type, payload: data.payload };
}

// Snapshots establish a new baseline; unknown messages still advance the cursor.
export class StreamCursor {
  private runId: string | null = null;
  private sequence = 0;
  accept(message: Message): boolean {
    if (message.type === 'snapshot') {
      this.runId = message.runId;
      this.sequence = message.sequence;
      return true;
    }
    if (message.runId !== this.runId) throw new StreamGapError('Host changed without a snapshot');
    if (message.sequence <= this.sequence) return false;
    if (message.sequence !== this.sequence + 1) throw new StreamGapError('Stream sequence gap');
    this.sequence = message.sequence;
    return true;
  }
}
