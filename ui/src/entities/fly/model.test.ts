import assert from 'node:assert/strict';
import test from 'node:test';
import * as THREE from 'three';
import { createFly } from './model.ts';
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
    const geometries = new Set<THREE.BufferGeometry>();
    const materials = new Set<THREE.Material>();
    fly.root.traverse(object => {
      if (!(object instanceof THREE.Mesh)) return;
      geometries.add(object.geometry);
      for (const material of Array.isArray(object.material) ? object.material : [object.material]) materials.add(material);
    });
    geometries.forEach(geometry => geometry.dispose());
    materials.forEach(material => material.dispose());
  }
});
