import assert from 'node:assert/strict';
import test from 'node:test';
import * as THREE from 'three';
import { createFly } from './model.ts';
import { createKeyboard } from './keyboard.ts';
import { smokingPose } from './smokingBreak.ts';
import type { FlyState } from './state.ts';

test('coffee signals one sound per sip without restarting when activity changes', () => {
  let sips = 0;
  const fly = createFly(() => sips++);
  const state: FlyState = { mode: 'manual', behavior: 'coffee', activity: 0.5 };
  try {
    fly.update(0, state);
    for (let frame = 1; frame <= 50; frame++) fly.update(frame / 10, state);
    assert.equal(sips, 0, 'the mug rests between sips');
    for (let frame = 51; frame <= 90; frame++) {
      fly.update(frame / 10, { ...state, activity: frame % 2 });
    }
    assert.equal(sips, 1, 'slider changes do not restart or duplicate a sip');
    for (let frame = 91; frame <= 180; frame++) fly.update(frame / 10, state);
    assert.equal(sips, 2);
    fly.update(18.1, { ...state, behavior: 'idle' });
    fly.update(18.2, state);
    assert.equal(sips, 2, 'entering coffee begins with a rest, not a replayed sip');
  } finally {
    dispose(fly.root);
  }
});

test('typing has irregular gaps, silent pauses, and no replay after a stalled frame', () => {
  let seed = 42;
  const random = () => {
    seed = (Math.imul(seed, 1664525) + 1013904223) >>> 0;
    return seed / 0x100000000;
  };
  const keyboard = createKeyboard(random);
  let clicks = 0;
  let clock = 0;
  const presses: number[] = [];
  const click = () => { clicks++; presses.push(clock); };
  try {
    let silentFrames = 0;
    for (let frame = 0; frame < 9600; frame++) {
      clock = frame / 80;
      const before = clicks;
      if (keyboard.update(clock, click).resting) {
        assert.equal(clicks, before, 'no sound during a pause');
        silentFrames++;
      }
    }
    assert.ok(clicks > 35);
    assert.ok(silentFrames > 0);
    const gaps = presses.slice(1).map((press, index) => press - presses[index]);
    assert.ok(Math.min(...gaps) < 1.5, 'fast runs are present');
    assert.ok(Math.max(...gaps) > 7, 'occasional longer pauses are present');
    assert.ok(new Set(gaps.map(gap => Math.round(gap * 10))).size > 10, 'the rhythm is not periodic');
    const beforeReset = clicks;
    keyboard.reset();
    clock = 0;
    keyboard.update(0, click);
    assert.equal(clicks, beforeReset, 'reset does not replay a keypress');
    clock = 1000;
    keyboard.update(clock, click);
    assert.equal(clicks, beforeReset, 'a stalled frame does not burst queued clicks');
    clock = 1000.25;
    keyboard.update(clock, click);
    keyboard.update(clock, click);
    assert.equal(clicks, beforeReset + 1, 'only one click for the new stroke');
  } finally {
    dispose(keyboard.root);
  }
});

test('a smoking break rests between draws and exhales only after lowering the cigarette', () => {
  let drawSeen = false;
  let exhaleSeen = false;
  let restingFrames = 0;
  for (let frame = 0; frame < 1200; frame++) {
    const pose = smokingPose(frame / 100);
    if (pose.draw > 0) {
      drawSeen = true;
      assert.equal(pose.exhale, 0, 'no exhale while the cigarette is at the mouth');
    }
    if (pose.exhale > 0) {
      assert.ok(drawSeen, 'a draw precedes exhalation');
      assert.equal(pose.draw, 0);
      exhaleSeen = true;
    }
    if (pose.draw === 0 && pose.exhale === 0) restingFrames++;
  }
  assert.ok(drawSeen && exhaleSeen);
  assert.ok(restingFrames > 600, 'most of the break is spent resting');
  assert.deepEqual(smokingPose(12), smokingPose(0), 'the cycle returns smoothly to rest');
});

function dispose(root: THREE.Object3D) {
  const geometries = new Set<THREE.BufferGeometry>();
  const materials = new Set<THREE.Material>();
  root.traverse(object => {
    if (!(object instanceof THREE.Mesh)) return;
    geometries.add(object.geometry);
    for (const material of Array.isArray(object.material) ? object.material : [object.material]) materials.add(material);
  });
  geometries.forEach(geometry => geometry.dispose());
  materials.forEach(material => material.dispose());
}
