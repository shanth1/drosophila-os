import * as THREE from 'three';

export function createControlConsole() {
  const root = new THREE.Group();
  root.scale.y = 0.45;
  const body = new THREE.MeshStandardMaterial({ color: '#555148', metalness: 0.3, roughness: 0.65 });
  const panel = new THREE.MeshStandardMaterial({ color: '#a49a85', metalness: 0.25, roughness: 0.6 });
  const trim = new THREE.MeshStandardMaterial({ color: '#423e36', metalness: 0.45, roughness: 0.5 });
  const dark = new THREE.MeshStandardMaterial({ color: '#252a25', roughness: 0.65 });
  const amber = new THREE.MeshStandardMaterial({ color: '#d7ad71', emissive: '#d7ad71', emissiveIntensity: 0.55 });
  const green = new THREE.MeshStandardMaterial({ color: '#91ae7b', emissive: '#91ae7b', emissiveIntensity: 0.4 });
  const red = new THREE.MeshStandardMaterial({ color: '#a85b45', roughness: 0.55 });
  const indicators: { material: THREE.MeshStandardMaterial; phase: number }[] = [];

  function box(parent: THREE.Object3D, w: number, h: number, d: number, x: number, y: number, z: number, material: THREE.Material) {
    const mesh = new THREE.Mesh(new THREE.BoxGeometry(w, h, d), material);
    mesh.position.set(x, y, z);
    mesh.castShadow = true;
    mesh.receiveShadow = true;
    parent.add(mesh);
    return mesh;
  }

  // A continuous horseshoe cabinet surrounds the operator's open working area.
  const radius = 10.7;
  const halfArc = 0.55;
  function arc(inner: number, outer: number, height: number, y: number, material: THREE.Material) {
    const shape = new THREE.Shape();
    for (let step = 0; step <= 64; step++) {
      const angle = -halfArc + step / 64 * halfArc * 2;
      const x = Math.sin(angle) * outer;
      const z = Math.cos(angle) * outer;
      if (step === 0) shape.moveTo(x, z);
      else shape.lineTo(x, z);
    }
    for (let step = 64; step >= 0; step--) {
      const angle = -halfArc + step / 64 * halfArc * 2;
      shape.lineTo(Math.sin(angle) * inner, Math.cos(angle) * inner);
    }
    shape.closePath();
    const mesh = new THREE.Mesh(new THREE.ExtrudeGeometry(shape, { depth: height, bevelEnabled: false, steps: 1 }), material);
    mesh.rotation.x = -Math.PI / 2;
    mesh.position.y = y;
    mesh.castShadow = true;
    mesh.receiveShadow = true;
    root.add(mesh);
  }
  arc(radius - 0.95, radius + 0.95, 1.55, 0.12, body);
  arc(radius - 1, radius + 1, 0.12, 0, trim);
  arc(radius - 1, radius - 0.97, 0.025, 1.6, amber);

  for (let index = 0; index < 9; index++) {
    const angle = -halfArc + (index + 0.5) / 9 * halfArc * 2;
    const station = new THREE.Group();
    station.position.set(Math.sin(angle) * radius, 1.68, -Math.cos(angle) * radius);
    station.rotation.y = -angle;
    root.add(station);
    const deck = new THREE.Group();
    deck.rotation.x = 0.2;
    station.add(deck);
    box(deck, 1.27, 0.16, 1.95, 0, 0, 0, panel);
    box(deck, 1.27, 0.16, 0.08, 0, 0.015, 0.94, trim);
    // Blank instrument windows reserve room for future module readings.
    box(deck, 1.1, 0.025, 0.46, 0, 0.1, -0.48, trim);
    box(deck, 0.99, 0.012, 0.35, 0, 0.118, -0.48, dark);
    for (let key = 0; key < 3; key++) {
      const material = (key === 2 ? amber : green).clone();
      box(deck, 0.09, 0.045, 0.09, -0.32 + key * 0.32, 0.11, -0.08, material);
      indicators.push({ material, phase: index * 1.7 + key * 2.3 });
      box(deck, 0.085, 0.05, 0.1, -0.32 + key * 0.32, 0.11, 0.12, key === 2 && index % 4 === 0 ? red : trim);
    }
    const dial = new THREE.Mesh(new THREE.CylinderGeometry(0.055, 0.065, 0.06, 16), trim);
    dial.position.set(-0.12, 0.115, 0.31);
    deck.add(dial);
    box(deck, 0.012, 0.009, 0.042, -0.12, 0.148, 0.31, panel);
    box(deck, 0.08, 0.018, 0.09, 0.14, 0.1, 0.31, dark);
    const toggle = box(deck, 0.018, 0.09, 0.018, 0.14, 0.15, 0.31, panel);
    toggle.rotation.x = index % 2 ? 0.3 : -0.3;
    if (index % 3 === 0) {
      // Compact keyboard-like controls echo the administration workstations.
      for (let row = 0; row < 3; row++) {
        for (let key = 0; key < 8; key++) {
          box(deck, 0.085, 0.035, 0.08, -0.4 + key * 0.115, 0.11, 0.48 + row * 0.12, key === 7 ? trim : body);
        }
      }
    } else if (index % 3 === 1) {
      for (let slider = 0; slider < 4; slider++) {
        const x = -0.39 + slider * 0.26;
        box(deck, 0.035, 0.015, 0.38, x, 0.095, 0.65, dark);
        box(deck, 0.13, 0.065, 0.075, x, 0.13, 0.53 + (slider % 3) * 0.11, trim);
        for (let tick = 0; tick < 4; tick++) box(deck, 0.045, 0.008, 0.012, x + 0.07, 0.1, 0.5 + tick * 0.1, body);
      }
    } else {
      for (const x of [-0.28, 0.28]) {
        const gauge = new THREE.Mesh(new THREE.CylinderGeometry(0.18, 0.18, 0.025, 32), trim);
        gauge.position.set(x, 0.11, 0.63);
        deck.add(gauge);
        const face = new THREE.Mesh(new THREE.CylinderGeometry(0.15, 0.15, 0.012, 32), panel);
        face.position.set(x, 0.13, 0.63);
        deck.add(face);
        const needle = box(deck, 0.012, 0.01, 0.12, x, 0.142, 0.6, red);
        needle.rotation.y = index * 0.4;
      }
    }
    box(station, 1.08, 0.52, 0.045, 0, -0.58, 0.98, trim);
    for (let vent = 0; vent < 5; vent++) {
      box(station, 0.8, 0.025, 0.015, 0, -0.75 + vent * 0.08, 1.01, dark);
    }
  }

  // The blank display follows the room's inward-facing cylindrical wall.
  const monitor = new THREE.Group();
  const monitorHeightScale = 1.75;
  const monitorBottom = 4.55 - 4.125 / 2;
  monitor.scale.y = monitorHeightScale;
  monitor.position.y = monitorBottom * (1 - monitorHeightScale);
  root.add(monitor);
  const screenArc = 1.025;
  const screenStart = Math.PI - screenArc / 2;
  function curvedPanel(radius: number, height: number, y: number, material: THREE.MeshStandardMaterial) {
    material.side = THREE.BackSide;
    const mesh = new THREE.Mesh(new THREE.CylinderGeometry(radius, radius, height, 128, 1, true, screenStart, screenArc), material);
    mesh.position.y = y;
    monitor.add(mesh);
    return mesh;
  }
  curvedPanel(13.35, 4.125, 4.55, body.clone());
  curvedPanel(13.28, 3.95, 4.55, trim.clone());
  const displayMaterial = new THREE.MeshStandardMaterial({ color: '#18211d', emissive: '#718675', emissiveIntensity: 0.08, roughness: 0.35 });
  const display = curvedPanel(13.2, 3.8, 4.55, displayMaterial);
  for (const angle of [screenStart, screenStart + screenArc]) {
    const edge = box(monitor, 0.15, 4.125, 0.18, Math.sin(angle) * 13.23, 4.55, Math.cos(angle) * 13.23, trim);
    edge.rotation.y = angle;
  }
  curvedPanel(13.17, 0.025, 2.55, amber.clone());

  return {
    root,
    display,
    update(time: number) {
      for (const indicator of indicators) {
        indicator.material.emissiveIntensity = Math.sin(time * 2.7 + indicator.phase) > 0.25 ? 1.3 : 0.15;
      }
    },
  };
}
