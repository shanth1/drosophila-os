import * as THREE from 'three';

export function createServerRacks() {
  const root = new THREE.Group();
  const chassis = new THREE.MeshStandardMaterial({ color: '#49463e', metalness: 0.35, roughness: 0.65 });
  const panel = new THREE.MeshStandardMaterial({ color: '#797265', metalness: 0.4, roughness: 0.55 });
  const dark = new THREE.MeshStandardMaterial({ color: '#242720', roughness: 0.8 });
  const steel = new THREE.MeshStandardMaterial({ color: '#b0a28a', metalness: 0.65, roughness: 0.45 });
  const wood = new THREE.MeshStandardMaterial({ color: '#a88c68', roughness: 0.8 });
  const indicators: { material: THREE.MeshStandardMaterial; phase: number; speed: number }[] = [];
  const loadBars: { mesh: THREE.Mesh; phase: number; axis: 'x' | 'y' }[] = [];

  function box(parent: THREE.Object3D, width: number, height: number, depth: number, x: number, y: number, z: number, material: THREE.Material) {
    const mesh = new THREE.Mesh(new THREE.BoxGeometry(width, height, depth), material);
    mesh.position.set(x, y, z);
    mesh.castShadow = true;
    mesh.receiveShadow = true;
    parent.add(mesh);
    return mesh;
  }

  function glow(color: string, intensity = 2) {
    return new THREE.MeshStandardMaterial({ color, emissive: color, emissiveIntensity: intensity, roughness: 0.3 });
  }

  for (const side of [-1, 1]) {
    const bank = new THREE.Group();
    bank.position.set(side * 10.8, 0, 0);
    // Front panels face inward, toward the fly and the open center of the room.
    bank.rotation.y = side * Math.PI / 2;
    root.add(bank);
    const warm = glow('#e8b875', 0.65);
    const screen = glow('#91aa85', 0.45);
    const storage = side === 1;

    for (let rackIndex = 0; rackIndex < 2; rackIndex++) {
      const rack = new THREE.Group();
      rack.position.x = (rackIndex - 0.5) * 1.95;
      if (storage) rack.scale.y = rackIndex === 0 ? 0.72 : 0.82;
      bank.add(rack);
      box(rack, 1.8, 0.2, 1.55, 0, 0.1, 0, chassis);
      box(rack, 1.8, 0.16, 1.55, 0, 4.95, 0, chassis);
      box(rack, 1.65, 4.7, 0.12, 0, 2.55, 0.7, dark);
      for (const x of [-0.84, 0.84]) {
        box(rack, 0.12, 4.75, 1.5, x, 2.55, 0, chassis);
        box(rack, 0.018, 4.45, 0.035, x, 2.55, -0.77, warm);
      }
      box(rack, 1.5, 0.3, 0.12, 0, 4.65, -0.7, panel);
      // Small luminous header bars distinguish compute and storage cabinets.
      for (let bar = 0; bar < (storage ? 5 : 3); bar++) {
        box(rack, 0.15, 0.045, 0.02, -0.45 + bar * 0.2, 4.65, -0.775, warm);
      }
      for (let row = 0; row < 9; row++) {
        const y = 0.48 + row * 0.44;
        box(rack, 1.5, 0.37, 1.25, 0, y, 0, panel);
        if (!storage) {
          for (let vent = 0; vent < 9; vent++) {
            box(rack, 0.055, 0.19, 0.025, -0.55 + vent * 0.1, y, -0.64, dark);
          }
          for (const x of [-0.68, 0.68]) box(rack, 0.035, 0.2, 0.06, x, y, -0.68, steel);
        } else {
          for (let drive = 0; drive < 4; drive++) {
            const x = -0.54 + drive * 0.36;
            box(rack, 0.3, 0.25, 0.04, x, y, -0.65, dark);
            box(rack, 0.2, 0.025, 0.03, x, y - 0.07, -0.685, steel);
          }
        }
        for (let led = 0; led < 3; led++) {
          const material = glow(led === 2 ? '#e8ae59' : '#9acb7a', 1);
          const x = !storage ? 0.42 + led * 0.085 : -0.54 + led * 0.36;
          box(rack, 0.035, 0.045, 0.025, x, y + 0.07, -0.69, material);
          indicators.push({ material, phase: row * 1.73 + led * 2.4 + rackIndex + side, speed: 2.5 + (row % 4) * 1.3 + led });
        }
        box(rack, 0.5, 0.035, 0.025, -0.27, y - 0.12, -0.69, dark);
        const bar = box(rack, 0.48, 0.018, 0.025, -0.27, y - 0.12, -0.71, screen);
        loadBars.push({ mesh: bar, phase: row * 0.7 + rackIndex * 2 + side, axis: 'x' });
      }
    }

    // A stacked cylinder makes the neighboring database unit recognizable.
    if (storage) {
      const database = new THREE.Group();
      database.position.set(2.85, 0, 0);
      bank.add(database);
      for (let tier = 0; tier < 4; tier++) {
      const drum = new THREE.Mesh(new THREE.CylinderGeometry(0.72, 0.72, 0.78, 48), panel);
      drum.position.y = 0.55 + tier * 0.84;
      drum.castShadow = true;
      drum.receiveShadow = true;
      database.add(drum);
      const ring = new THREE.Mesh(new THREE.TorusGeometry(0.725, 0.015, 8, 48), warm);
      ring.rotation.x = Math.PI / 2;
      ring.position.y = 0.22 + tier * 0.84;
      database.add(ring);
      box(database, 0.3, 0.1, 0.04, 0, drum.position.y, -0.73, dark);
      const material = glow('#9acb7a', 1);
      box(database, 0.18, 0.025, 0.025, 0, drum.position.y, -0.76, material);
      indicators.push({ material, phase: tier * 2.1, speed: 3 + tier });
      }
    }

    // Compute has an integrated service console; storage has a furnished desk.
    const desk = new THREE.Group();
    desk.position.set(storage ? 0 : -0.975, 0, storage ? -2.1 : -0.95);
    bank.add(desk);
    if (storage) {
      box(desk, 3.4, 0.12, 1.2, 0, 1.35, 0, wood);
      for (const x of [-1.48, 1.48]) box(desk, 0.1, 1.3, 0.85, x, 0.65, 0, chassis);
    } else {
      box(desk, 1.45, 0.08, 0.85, 0, 1.35, -0.15, chassis);
      for (const x of [-0.66, 0.66]) box(desk, 0.045, 0.08, 0.7, x, 1.27, 0, steel);
    }
    const monitorWidth = storage ? 2.1 : 1.2;
    for (const x of [0]) {
      box(desk, 0.45, 0.045, 0.28, x, 1.44, 0.15, chassis);
      box(desk, 0.07, 0.3, 0.07, x, 1.58, 0.24, steel);
      box(desk, monitorWidth, 0.76, 0.09, x, 1.98, 0.24, chassis);
      box(desk, monitorWidth - 0.12, 0.64, 0.015, x, 1.98, 0.185, dark);
      for (let line = 0; line < 6; line++) {
        box(desk, 0.24 + (line % 3) * 0.14, 0.014, 0.01, x - (storage ? 0.55 : 0.18), 2.19 - line * 0.065, 0.17, screen);
      }
      for (let column = 0; column < 5; column++) {
        const bar = box(desk, 0.055, 0.3, 0.012, x + 0.12 + column * 0.07, 1.96, 0.165, screen);
        loadBars.push({ mesh: bar, phase: column * 0.9 + x + side, axis: 'y' });
      }
    }
    box(desk, 1.1, 0.045, 0.35, 0, 1.44, -0.32, chassis);
    for (let row = 0; row < 3; row++) {
      for (let key = 0; key < 12; key++) {
        box(desk, 0.065, 0.012, 0.065, -0.46 + key * 0.083, 1.47, -0.43 + row * 0.085, panel);
      }
    }
    if (storage) {
      box(desk, 0.14, 0.06, 0.22, 0.85, 1.45, -0.32, panel);
      box(desk, 0.4, 0.9, 0.75, 1.12, 0.87, 0.08, chassis);
      box(desk, 0.25, 0.55, 0.025, 1.12, 0.92, -0.31, dark);
      box(desk, 0.03, 0.03, 0.025, 1.12, 1.25, -0.33, warm);

      // Spare drives, notebook, and a shaded task light soften the workstation.
      box(desk, 0.42, 0.05, 0.35, -1.25, 1.44, 0.1, wood);
      for (let drive = 0; drive < 3; drive++) {
        box(desk, 0.09, 0.26, 0.26, -1.38 + drive * 0.13, 1.59, 0.1, steel);
        box(desk, 0.06, 0.09, 0.01, -1.38 + drive * 0.13, 1.62, -0.035, panel);
      }
      box(desk, 0.35, 0.035, 0.27, -0.9, 1.44, -0.36, dark);
      box(desk, 0.32, 0.018, 0.25, -0.9, 1.465, -0.36, steel);
      box(desk, 0.23, 0.045, 0.23, 1.38, 1.44, 0.22, chassis);
      box(desk, 0.035, 0.52, 0.035, 1.38, 1.7, 0.22, steel);
      const shade = new THREE.Mesh(new THREE.ConeGeometry(0.2, 0.22, 24, 1, true), panel);
      shade.position.set(1.38, 2.01, 0.22);
      shade.material = new THREE.MeshStandardMaterial({ color: '#797265', roughness: 0.65, side: THREE.DoubleSide });
      desk.add(shade);
      box(desk, 0.19, 0.018, 0.19, 1.38, 1.91, 0.22, warm);
      const taskLight = new THREE.PointLight('#ffd6a0', 0.8, 2);
      taskLight.position.set(1.38, 1.88, 0.22);
      desk.add(taskLight);

      const cabinet = new THREE.Group();
      cabinet.position.set(-2.7, 0, -1.9);
      bank.add(cabinet);
      box(cabinet, 1.05, 1.18, 1, 0, 0.59, 0, chassis);
      box(cabinet, 1.12, 0.08, 1.05, 0, 1.22, 0, wood);
      for (let drawer = 0; drawer < 3; drawer++) {
        box(cabinet, 0.93, 0.3, 0.04, 0, 0.24 + drawer * 0.35, -0.52, panel);
        box(cabinet, 0.24, 0.025, 0.055, 0, 0.29 + drawer * 0.35, -0.56, steel);
      }
    } else {
      // Separate UPS and patch panel keep the compute side vertical and technical.
      box(bank, 1.1, 1.3, 1.05, 2.65, 0.65, 0, chassis);
      box(bank, 0.85, 1.1, 0.04, 2.65, 0.65, -0.55, panel);
      box(bank, 0.48, 0.2, 0.025, 2.65, 1.02, -0.58, dark);
      box(bank, 0.3, 0.035, 0.02, 2.65, 1.02, -0.6, screen);
      for (let vent = 0; vent < 6; vent++) box(bank, 0.65, 0.025, 0.02, 2.65, 0.28 + vent * 0.075, -0.58, dark);
      box(bank, 0.24, 4.1, 0.25, 2.18, 2.4, 0.35, dark);
      for (let rung = 0; rung < 12; rung++) box(bank, 0.28, 0.045, 0.3, 2.18, 0.65 + rung * 0.32, 0.35, steel);
      box(bank, 1.5, 0.32, 0.12, 0.975, 4.23, -0.73, dark);
      for (let port = 0; port < 8; port++) {
        const x = 0.4 + port * 0.16;
        box(bank, 0.1, 0.1, 0.03, x, 4.25, -0.81, steel);
        const cable = new THREE.CatmullRomCurve3([
          new THREE.Vector3(x, 4.25, -0.84),
          new THREE.Vector3(x, 4.02 - port * 0.035, -1.02),
          new THREE.Vector3(1.85, 3.85 - port * 0.035, -0.9),
          new THREE.Vector3(2.18, 3.6, 0.35),
        ]);
        const wire = new THREE.Mesh(new THREE.TubeGeometry(cable, 16, 0.012, 5, false), port % 2 ? dark : wood);
        bank.add(wire);
      }
    }
    const light = new THREE.PointLight('#e8bc82', 3, 6, 2);
    light.position.set(0, 2.4, -1.1);
    bank.add(light);
  }

  return {
    root,
    update(time: number) {
      for (const indicator of indicators) {
        const signal = Math.sin(time * indicator.speed + indicator.phase);
        indicator.material.emissiveIntensity = signal > 0.35 ? 1.8 : 0.15;
      }
      // Decorative utilization displays, independent of host telemetry.
      for (const bar of loadBars) {
        const load = 0.2 + 0.8 * (0.5 + 0.5 * Math.sin(time * 0.7 + bar.phase));
        bar.mesh.scale[bar.axis] = load;
      }
    },
  };
}
