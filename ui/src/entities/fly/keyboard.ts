import * as THREE from 'three';

export function createKeyboard(random: () => number = Math.random) {
  const root = new THREE.Group();
  root.position.set(0, 0.38, -1.36);
  const casing = new THREE.MeshStandardMaterial({ color: '#344b53', roughness: 0.45 });
  const keyMaterial = new THREE.MeshStandardMaterial({ color: '#dae4d8', roughness: 0.55 });
  const accent = new THREE.MeshStandardMaterial({ color: '#e6ac56', roughness: 0.5 });
  const legendMaterial = new THREE.MeshStandardMaterial({ color: '#657b78', roughness: 0.7 });
  const keyGeometry = new THREE.BoxGeometry(0.112, 0.055, 0.105);
  const base = new THREE.Mesh(new THREE.BoxGeometry(1.3, 0.09, 0.64), casing);
  base.castShadow = true;
  base.receiveShadow = true;
  root.add(base);
  const keys: THREE.Mesh[] = [];
  for (let row = 0; row < 3; row++) {
    for (let column = 0; column < 8; column++) {
      const key = new THREE.Mesh(keyGeometry, row === 0 && column === 0 ? accent : keyMaterial);
      key.position.set(-0.49 + column * 0.14, 0.07, -0.19 + row * 0.135);
      key.castShadow = true;
      root.add(key);
      keys.push(key);
      const legend = new THREE.Mesh(new THREE.BoxGeometry(0.032, 0.002, 0.008), legendMaterial);
      legend.position.set(0, 0.028, -0.012);
      key.add(legend);
    }
  }
  const space = new THREE.Mesh(new THREE.BoxGeometry(0.55, 0.055, 0.09), accent);
  space.position.set(0, 0.07, 0.215);
  root.add(space);
  const lightMaterial = new THREE.MeshStandardMaterial({ color: '#82e7c9', emissive: '#35b995', emissiveIntensity: 0.6 });
  for (let index = 0; index < 3; index++) {
    const indicator = new THREE.Mesh(new THREE.SphereGeometry(0.015, 10, 8), lightMaterial);
    indicator.position.set(0.53 - index * 0.055, 0.049, 0.22);
    root.add(indicator);
  }
  for (const side of [-1, 1]) {
    const stand = new THREE.Mesh(new THREE.BoxGeometry(0.065, 0.33, 0.46), casing);
    stand.position.set(side * 0.52, -0.2, 0);
    stand.castShadow = true;
    root.add(stand);
  }
  const cable = new THREE.Mesh(new THREE.TubeGeometry(new THREE.CatmullRomCurve3([
    new THREE.Vector3(0.57, -0.02, -0.31), new THREE.Vector3(0.8, -0.1, -0.5),
    new THREE.Vector3(0.91, -0.34, -0.4), new THREE.Vector3(1.1, -0.35, -0.15),
  ]), 20, 0.018, 6, false), casing);
  root.add(cable);
  const left = new THREE.Vector3();
  const right = new THREE.Vector3();
  const leftTarget = new THREE.Vector3();
  const rightTarget = new THREE.Vector3();
  let leftKey = keys[9];
  let rightKey = keys[14];
  let activeKey = leftKey;
  let activeSide = 0;
  let strokeStart = -Infinity;
  let nextStroke = 0.2;
  let sounded = false;
  let previousClock: number | null = null;

  return {
    root,
    update(clock: number, onKeyPress?: () => void) {
      if (clock >= nextStroke) {
        activeSide = random() < 0.7 ? 1 - activeSide : activeSide;
        const row = Math.floor(random() * 3);
        const column = activeSide * 4 + Math.floor(random() * 4);
        activeKey = random() < 0.08 ? space : keys[row * 8 + column];
        if (activeSide === 0) leftKey = activeKey;
        else rightKey = activeKey;
        strokeStart = clock;
        sounded = false;
        const pause = random();
        const interval = pause < 0.06 ? 7 + random() * 8 : pause < 0.25 ? 1.8 + random() * 2.2 : 0.65 + random() * 0.85;
        // Schedule from now, without replaying missed presses after a stalled frame.
        nextStroke = clock + interval;
      }
      const elapsed = clock - strokeStart;
      const resting = elapsed > 0.9;
      const stroke = elapsed < 0.5 ? Math.sin(elapsed / 0.5 * Math.PI) ** 2 : 0;
      keys.forEach(key => { key.position.y = 0.07; });
      space.position.y = 0.07;
      activeKey.position.y -= stroke * 0.023;
      leftTarget.copy(leftKey.position).add(root.position);
      rightTarget.copy(rightKey.position).add(root.position);
      const smoothing = previousClock === null ? 1 : 1 - Math.exp(-Math.max(0, clock - previousClock) * 14);
      previousClock = clock;
      left.lerp(leftTarget, smoothing);
      right.lerp(rightTarget, smoothing);
      // Keep taps exact while smoothing lateral travel between keys.
      left.y = leftTarget.y + 0.145 - (activeSide === 0 ? stroke * 0.08 : 0);
      right.y = rightTarget.y + 0.145 - (activeSide === 1 ? stroke * 0.08 : 0);
      if (!sounded && elapsed >= 0.22 && elapsed < 0.5) {
        onKeyPress?.();
        sounded = true;
      }
      lightMaterial.emissiveIntensity = 0.5 + stroke * 0.8;
      return { left, right, resting };
    },
    reset() { strokeStart = -Infinity; nextStroke = 0.2; sounded = false; previousClock = null; },
  };
}
