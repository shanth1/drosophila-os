import * as THREE from 'three';

export function createPlants() {
  const root = new THREE.Group();
  const clays = ['#c3ac90', '#bda08a', '#d0bba0'].map(color =>
    new THREE.MeshStandardMaterial({ color, roughness: 0.95 }),
  );
  const soil = new THREE.MeshStandardMaterial({ color: '#30271f', roughness: 1 });
  const stem = new THREE.MeshStandardMaterial({ color: '#587447', roughness: 0.8 });
  const vein = new THREE.MeshStandardMaterial({ color: '#8caa63', roughness: 0.8 });
  const greens = ['#385f3c', '#507944', '#64894e'].map(color =>
    new THREE.MeshStandardMaterial({ color, roughness: 0.68, side: THREE.DoubleSide, shadowSide: THREE.DoubleSide }),
  );

  function mesh(parent: THREE.Object3D, geometry: THREE.BufferGeometry, material: THREE.Material) {
    const object = new THREE.Mesh(geometry, material);
    object.castShadow = true;
    object.receiveShadow = true;
    parent.add(object);
    return object;
  }

  function branch(parent: THREE.Object3D, points: THREE.Vector3[], radius: number, material: THREE.Material) {
    const curve = new THREE.CatmullRomCurve3(points);
    mesh(parent, new THREE.TubeGeometry(curve, 12, radius, 6, false), material);
    return curve;
  }

  function pot(material: THREE.Material, radius: number, height: number, taper = 0.8, belly = 0.94) {
    const plant = new THREE.Group();
    root.add(plant);
    const profile = [
      new THREE.Vector2(radius * (taper - 0.07), 0),
      new THREE.Vector2(radius * taper, 0.06),
      new THREE.Vector2(radius * belly, height * 0.55),
      new THREE.Vector2(radius, height - 0.04),
      new THREE.Vector2(radius, height),
      new THREE.Vector2(radius - 0.055, height),
      new THREE.Vector2(radius - 0.055, height - 0.12),
    ];
    mesh(plant, new THREE.LatheGeometry(profile, 40), material);
    const earth = mesh(plant, new THREE.CircleGeometry(radius - 0.06, 32), soil);
    earth.rotation.x = -Math.PI / 2;
    earth.position.y = height - 0.09;
    return plant;
  }

  // A folded, tapering blade with a raised midrib and a gently drooping tip.
  function leaf(parent: THREE.Object3D, origin: THREE.Vector3, length: number, width: number, yaw: number, tilt: number, color: number) {
    const blade = new THREE.Group();
    blade.position.copy(origin);
    blade.rotation.set(tilt, yaw, 0, 'YXZ');
    parent.add(blade);
    const vertices: number[] = [];
    const indices: number[] = [];
    const segments = 12;
    for (let i = 0; i <= segments; i++) {
      const t = i / segments;
      const halfWidth = width * Math.pow(Math.sin(Math.PI * t), 0.85) / 2;
      const arch = length * (0.2 * Math.sin(Math.PI * t) - 0.18 * t * t);
      for (const side of [-1, 0, 1]) {
        vertices.push(side * halfWidth, arch - Math.abs(side) * halfWidth * 0.22, -length * t);
      }
      if (i < segments) {
        for (let side = 0; side < 2; side++) {
          const a = i * 3 + side;
          indices.push(a, a + 3, a + 1, a + 1, a + 3, a + 4);
        }
      }
    }
    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute('position', new THREE.Float32BufferAttribute(vertices, 3));
    geometry.setIndex(indices);
    geometry.computeVertexNormals();
    mesh(blade, geometry, greens[color % greens.length]);
    const midrib = [0, 0.3, 0.65, 0.95].map(t => new THREE.Vector3(
      0, length * (0.2 * Math.sin(Math.PI * t) - 0.18 * t * t) + 0.008, -length * t,
    ));
    branch(blade, midrib, 0.006, vein);
  }

  function palm(seed: number, fronds: number, baseHeight: number, spread: number, potVariant: number) {
    // Stable variation keeps the decor consistent across remounts and reloads.
    function random() {
      seed = (Math.imul(seed, 1664525) + 1013904223) >>> 0;
      return seed / 4294967296;
    }
    const plant = pot(clays[potVariant], 0.54, 0.72, potVariant === 0 ? 0.8 : 0.88, potVariant === 0 ? 0.94 : 1.02);
    for (let i = 0; i < fronds; i++) {
      const yaw = i * 2.4 + (random() - 0.5) * 0.7;
      const height = baseHeight + (random() - 0.5) * 0.95;
      const reach = spread * (0.75 + random() * 0.45);
      const droop = 0.45 + random() * 0.5;
      const direction = new THREE.Vector3(Math.sin(yaw), 0, Math.cos(yaw));
      const base = new THREE.Vector3((random() - 0.5) * 0.4, 0.63, (random() - 0.5) * 0.4);
      const curve = branch(plant, [
        base,
        direction.clone().multiplyScalar(reach * 0.15).setY(height - 0.7),
        direction.clone().multiplyScalar(reach * 0.52).setY(height),
        direction.clone().multiplyScalar(reach).setY(height - droop),
      ], 0.02, stem);
      const pairs = 11 + Math.floor(random() * 3);
      for (let j = 0; j < pairs; j++) {
        const t = 0.36 + j / pairs * 0.61;
        const length = (0.58 * Math.sin((j + 2) / (pairs + 3) * Math.PI) + 0.1) * (0.85 + random() * 0.3);
        for (const side of [-1, 1]) {
          const origin = curve.getPoint(t + (random() - 0.5) * 0.015);
          leaf(plant, origin, length * (0.9 + random() * 0.2), 0.08 + random() * 0.035,
            -yaw + Math.PI + side * (0.85 + random() * 0.35),
            -0.12 - j * 0.03 - random() * 0.15, i + j);
        }
      }
    }
    return plant;
  }

  function monstera() {
    const plant = pot(clays[2], 0.49, 0.62, 0.85, 1.05);
    for (let i = 0; i < 7; i++) {
      const yaw = i * 2.4 + Math.sin(i * 3.7) * 0.2;
      const height = 1.25 + (i % 3) * 0.37;
      const reach = 0.95 + (i % 3) * 0.16;
      const origin = new THREE.Vector3(Math.sin(yaw) * reach, height, Math.cos(yaw) * reach);
      branch(plant, [new THREE.Vector3(0, 0.53, 0), origin.clone().setY(height * 0.72), origin], 0.022, stem);
      const shape = new THREE.Shape();
      shape.moveTo(0, 0);
      // Deep edge cuts form separate lobes while preserving the central blade.
      for (const side of [1, -1]) {
        const points = [
          [0.12, 0.02], [0.36, -0.08], [0.49, 0.08], [0.46, 0.25],
          [0.12, 0.3], [0.48, 0.34], [0.46, 0.49], [0.1, 0.51],
          [0.41, 0.57], [0.35, 0.71], [0.08, 0.7], [0.29, 0.78],
          [0.18, 0.92], [0, 1.06],
        ];
        if (side === -1) points.reverse();
        for (const [x, y] of points) shape.lineTo(side * x, y);
      }
      shape.closePath();
      for (const side of [-1, 1]) {
        for (const y of [0.18, 0.41]) {
          const hole = new THREE.Path();
          hole.absellipse(side * 0.22, y, 0.035, 0.065, 0, Math.PI * 2, true, side * 0.3);
          shape.holes.push(hole);
        }
      }
      const geometry = new THREE.ShapeGeometry(shape, 12);
      const positions = geometry.getAttribute('position');
      for (let vertex = 0; vertex < positions.count; vertex++) {
        const x = positions.getX(vertex);
        const y = positions.getY(vertex);
        positions.setZ(vertex, 0.13 * Math.sin(y * Math.PI) - x * x * 0.35);
      }
      geometry.computeVertexNormals();
      const blade = new THREE.Group();
      blade.position.copy(origin);
      blade.rotation.set(0.8 + (i % 3) * 0.16, yaw, Math.sin(i * 1.7) * 0.2, 'YXZ');
      blade.scale.setScalar(0.95 + (i % 3) * 0.14);
      plant.add(blade);
      mesh(blade, geometry, greens[i % greens.length]);
      branch(blade, [new THREE.Vector3(0, 0, 0.008), new THREE.Vector3(0, 0.5, 0.138), new THREE.Vector3(0, 1.04, -0.008)], 0.009, vein);
    }
    return plant;
  }

  const tallPalm = palm(73, 9, 3.35, 1.8, 0);
  tallPalm.position.set(-7.8, 0, 4.5);
  const spreadingPalm = palm(218, 11, 2.95, 2.2, 1);
  spreadingPalm.position.set(4.8, 0, 7.4);
  const broadMonstera = monstera();
  broadMonstera.position.set(-3.8, 0, 8.6);
  const consolePalm = palm(341, 8, 2.5, 1.4, 1);
  consolePalm.position.set(-8.6, 0, -6.2);
  const consoleMonstera = monstera();
  consoleMonstera.position.set(8.6, 0, -6.2);
  consoleMonstera.rotation.y = 0.7;
  consoleMonstera.scale.setScalar(0.85);
  return root;
}
