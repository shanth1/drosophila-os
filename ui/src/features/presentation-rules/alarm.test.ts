import assert from 'node:assert/strict';
import test from 'node:test';
import { alarmSirenStrength, alarmSirenThreshold } from './alarm.ts';

test('the beacon and siren engage only at high activity and intensify together', () => {
  assert.equal(alarmSirenStrength(0), 0);
  assert.equal(alarmSirenStrength(0.5), 0);
  assert.equal(alarmSirenStrength(alarmSirenThreshold - 0.01), 0);
  assert.equal(alarmSirenStrength(alarmSirenThreshold), 0.35);
  assert.ok(alarmSirenStrength(0.8) > alarmSirenStrength(alarmSirenThreshold));
  assert.equal(alarmSirenStrength(1), 1);
});
