import * as THREE from 'three';
import type { FlyState } from './state';

export function createFly() {
  const root = new THREE.Group();
  const body = new THREE.Group();
  root.add(body);
  const shell = new THREE.MeshStandardMaterial({ color: '#766047', roughness: 0.65 });
  const dark = new THREE.MeshStandardMaterial({ color: '#272b29', roughness: 0.5 });
  const eye = new THREE.MeshPhysicalMaterial({
    color: '#c7503e',
    roughness: 0.32,
    clearcoat: 0.8,
    clearcoatRoughness: 0.24,
  });
  const face = new THREE.MeshStandardMaterial({ color: '#aa8960', roughness: 0.7 });
  const wingMaterial = new THREE.MeshPhysicalMaterial({ color: '#c8e6df', transparent: true, opacity: 0.38, roughness: 0.25, depthWrite: false, side: THREE.DoubleSide });
  const veinMaterial = new THREE.MeshStandardMaterial({ color: '#83aaa0', transparent: true, opacity: 0.55, roughness: 0.65 });
  const sphere = new THREE.SphereGeometry(1, 24, 16);
  function ellipsoid(parent: THREE.Object3D, material: THREE.Material, position: number[], scale: number[]) {
    const mesh = new THREE.Mesh(sphere, material);
    mesh.position.set(position[0], position[1], position[2]);
    mesh.scale.set(scale[0], scale[1], scale[2]);
    mesh.castShadow = true;
    parent.add(mesh);
    return mesh;
  }
  function tube(parent: THREE.Object3D, material: THREE.Material, points: number[][], radius: number) {
    const curve = new THREE.CatmullRomCurve3(points.map(point => new THREE.Vector3(point[0], point[1], point[2])));
    const mesh = new THREE.Mesh(new THREE.TubeGeometry(curve, 16, radius, 5, false), material);
    mesh.castShadow = material === dark;
    parent.add(mesh);
    return mesh;
  }
  ellipsoid(body, shell, [0, 1.15, 0], [0.48, 0.46, 0.65]);
  ellipsoid(body, face, [0, 1.51, 0.08], [0.34, 0.13, 0.45]);
  ellipsoid(body, shell, [0, 1.1, 0.95], [0.48, 0.4, 0.83]);
  for (let i = 0; i < 4; i++) {
    const z = 0.65 + i * 0.29;
    const radius = 0.48 * Math.sqrt(1 - ((z - 0.95) / 0.83) ** 2);
    const band = new THREE.Mesh(new THREE.TorusGeometry(radius, 0.022, 8, 32), dark);
    band.position.set(0, 1.1, z);
    band.scale.y = 0.4 / 0.48;
    body.add(band);
  }
  const head = new THREE.Group();
  head.position.set(0, 1.28, -0.72);
  body.add(head);
  ellipsoid(head, shell, [0, 0, 0], [0.4, 0.38, 0.35]);
  ellipsoid(head, face, [0, -0.13, -0.29], [0.2, 0.2, 0.12]);
  ellipsoid(head, dark, [0, -0.27, -0.36], [0.07, 0.1, 0.08]);
  for (const side of [-1, 1]) {
    const compoundEye = ellipsoid(head, eye, [side * 0.29, 0.05, -0.15], [0.27, 0.33, 0.26]);
    compoundEye.rotation.z = -side * 0.12;
    ellipsoid(head, face, [side * 0.12, 0.23, -0.33], [0.065, 0.08, 0.065]);
    ellipsoid(head, dark, [side * 0.15, 0.32, -0.38], [0.055, 0.11, 0.045]).rotation.z = -side * 0.3;
    tube(head, dark, [[side * 0.17, 0.39, -0.38], [side * 0.26, 0.52, -0.4], [side * 0.36, 0.56, -0.42]], 0.012);
    for (let i = 0; i < 3; i++) {
      tube(head, dark, [[side * (0.22 + i * 0.04), 0.47 + i * 0.03, -0.4], [side * (0.22 + i * 0.05), 0.55 + i * 0.03, -0.44]], 0.006);
      tube(body, dark, [[side * 0.3, 1.48, -0.22 + i * 0.23], [side * 0.43, 1.68, -0.27 + i * 0.23]], 0.012);
    }
  }
  const legs: THREE.Group[] = [];
  for (const side of [-1, 1]) {
    for (let index = 0; index < 3; index++) {
      const leg = new THREE.Group();
      leg.position.set(side * 0.3, 1.05, -0.35 + index * 0.45);
      body.add(leg);
      const knee = [side * 0.56, -0.35, (index - 1) * 0.3];
      const ankle = [side * 0.8, -0.94, (index - 1) * 0.48];
      tube(leg, dark, [[0, 0, 0], [side * 0.3, -0.12, (index - 1) * 0.16], knee], 0.043);
      ellipsoid(leg, shell, knee, [0.062, 0.062, 0.062]);
      tube(leg, dark, [knee, [side * 0.73, -0.65, (index - 1) * 0.4], ankle], 0.029);
      tube(leg, dark, [ankle, [side * 0.86, -1.01, (index - 1) * 0.48 - 0.06], [side * 0.94, -1.02, (index - 1) * 0.48 - 0.15]], 0.02);
      legs.push(leg);
    }
  }
  const wings: THREE.Group[] = [];
  for (const side of [-1, 1]) {
    const pivot = new THREE.Group();
    pivot.position.set(side * 0.24, 1.5, 0.05);
    body.add(pivot);
    const wing = ellipsoid(pivot, wingMaterial, [side * 0.72, 0, 0.64], [0.63, 0.018, 1.05]);
    wing.rotation.y = side * 0.5;
    wing.castShadow = false;
    const veins = new THREE.Group();
    veins.position.copy(wing.position);
    veins.rotation.copy(wing.rotation);
    pivot.add(veins);
    tube(veins, veinMaterial, [[0, 0.022, -0.96], [side * 0.22, 0.022, -0.42], [side * 0.25, 0.022, 0.28], [side * 0.12, 0.022, 0.9]], 0.009);
    tube(veins, veinMaterial, [[0, 0.022, -0.96], [-side * 0.15, 0.022, -0.25], [-side * 0.22, 0.022, 0.4], [0, 0.022, 0.98]], 0.008);
    for (const z of [-0.3, 0.25, 0.65]) {
      const width = 0.63 * Math.sqrt(1 - (z / 1.05) ** 2);
      tube(veins, veinMaterial, [[-side * width * 0.8, 0.022, z - 0.12], [0, 0.022, z], [side * width * 0.85, 0.022, z + 0.08]], 0.006);
    }
    wings.push(pivot);
  }
  const cup = new THREE.Group();
  const ceramic = new THREE.MeshStandardMaterial({ color: '#f2e8cf' });
  cup.add(new THREE.Mesh(new THREE.CylinderGeometry(0.13, 0.1, 0.23, 20), ceramic));
  const handle = new THREE.Mesh(new THREE.TorusGeometry(0.08, 0.025, 8, 16), ceramic);
  handle.position.x = 0.14;
  cup.add(handle);
  cup.position.set(0.45, 0.65, -0.95);
  body.add(cup);

  return {
    root,
    update(time: number, state: FlyState) {
      const busy = state.behavior === 'working' || state.behavior === 'alarmed';
      body.position.y = Math.sin(time * 2) * 0.025;
      head.rotation.y = Math.sin(time * (busy ? 4 : 0.7)) * (busy ? 0.2 : 0.08);
      head.rotation.x = state.behavior === 'coffee' ? 0.18 : 0;
      wings.forEach((wing, index) => {
        wing.rotation.z = (index === 0 ? -1 : 1) * (0.08 + Math.sin(time * (busy ? 35 : 4)) * state.activity * 0.3);
      });
      legs.forEach((leg, index) => { leg.rotation.x = busy ? Math.sin(time * 8 + index) * 0.08 : 0; });
      cup.visible = state.behavior === 'coffee';
      cup.position.y = 0.75 + Math.sin(time * 2) * 0.08;
      root.rotation.y = state.behavior === 'break' ? -0.45 : 0;
    },
  };
}
