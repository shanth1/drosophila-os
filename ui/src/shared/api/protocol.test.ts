import { readFileSync } from 'node:fs';
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { parseMessage, ProtocolError, StreamCursor, StreamGapError } from './protocol.ts';

const fixture = JSON.parse(readFileSync(new URL('../../../../testdata/telemetry-v1.json', import.meta.url), 'utf8'));

test('the shared Go wire fixture is accepted and additive fields remain compatible', () => {
  const message = parseMessage({ ...fixture, futureField: true, payload: { ...fixture.payload, futureState: {} } });
  assert.equal(message.type, 'snapshot');
  if (message.type !== 'snapshot') assert.fail('Expected snapshot');
  assert.equal(message.payload.brain.neurons, 3);
  assert.equal(message.payload.observations[0].name, 'http.response');
});

test('unsafe counts, malformed known measurements, and incompatible versions are rejected', () => {
  assert.throws(() => parseMessage({ ...fixture, apiVersion: 2 }), ProtocolError);
  assert.throws(() => parseMessage({ ...fixture, sequence: Number.MAX_SAFE_INTEGER + 1 }), ProtocolError);
  assert.throws(() => parseMessage({ ...fixture, payload: { ...fixture.payload, brain: { ...fixture.payload.brain, spikes: -1 } } }), ProtocolError);
  assert.throws(() => parseMessage({ ...fixture, payload: { ...fixture.payload, observations: [{ ...fixture.payload.observations[0], data: { statusCode: 200, elapsedMs: 'slow' } }] } }), ProtocolError);
});

test('unknown messages advance ordering; gaps and restarts require snapshots', () => {
  const cursor = new StreamCursor();
  assert.equal(cursor.accept(parseMessage(fixture)), true);
  const future = parseMessage({ ...fixture, type: 'future.event', sequence: 8, payload: {} });
  assert.equal(future.type, 'unknown');
  assert.equal(cursor.accept(future), true);
  assert.equal(cursor.accept(future), false);
  assert.throws(() => cursor.accept(parseMessage({ ...fixture, type: 'state', sequence: 10 })), StreamGapError);
  assert.throws(() => cursor.accept(parseMessage({ ...fixture, type: 'state', runId: 'new-host', sequence: 9 })), StreamGapError);
  assert.equal(cursor.accept(parseMessage({ ...fixture, runId: 'new-host', sequence: 0 })), true);
  assert.equal(cursor.accept(parseMessage({ ...fixture, type: 'state', runId: 'new-host', sequence: 1 })), true);
});
