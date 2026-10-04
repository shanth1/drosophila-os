import { readFileSync } from 'node:fs';
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { parseMessage } from '../../shared/api/protocol.ts';
import { initialHostFeed } from '../../shared/api/client.ts';
import { initialFlyState } from '../../entities/fly/state.ts';
import { resolveFlyState } from './liveFly.ts';

const message = parseMessage(JSON.parse(readFileSync(new URL('../../../../testdata/telemetry-v1.json', import.meta.url), 'utf8')));
if (message.type !== 'snapshot') throw new Error('Expected fixture snapshot');
const connected = { ...initialHostFeed, connection: 'connected' as const, state: message.payload };

test('a failed HTTP observation alone is not presented as a brain output', () => {
  const feed = { ...connected, state: { ...connected.state, observations: [{ ...connected.state.observations[0], data: { statusCode: 503, elapsedMs: 0 } }] } };
  assert.equal(resolveFlyState(initialFlyState, feed, 1000).behavior, 'idle');
});

test('fresh output gestures expire and disconnection clears live activity', () => {
  const feed = { ...connected, lastOutputReceivedAt: 1000 };
  assert.equal(resolveFlyState(initialFlyState, feed, 1100).behavior, 'alarmed');
  assert.equal(resolveFlyState(initialFlyState, feed, 2600).behavior, 'idle');
  assert.equal(resolveFlyState(initialFlyState, { ...feed, connection: 'disconnected' }, 1100).activity, 0);
  // A recovered snapshot may contain historical output counts, never a fresh gesture.
  const recovered = { ...connected, state: { ...connected.state, brain: { ...connected.state.brain, outputEvents: 50 } } };
  assert.equal(resolveFlyState(initialFlyState, recovered, 1100).behavior, 'idle');
});

test('manual behavior survives live updates and connection loss', () => {
  const manual = { mode: 'manual' as const, behavior: 'coffee' as const, activity: 0.3 };
  assert.deepEqual(resolveFlyState(manual, { ...connected, lastOutputReceivedAt: 1000 }, 1100), manual);
  assert.deepEqual(resolveFlyState(manual, initialHostFeed, 1100), manual);
});
