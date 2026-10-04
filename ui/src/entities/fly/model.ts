import * as THREE from 'three';
import type { FlyState } from './state';

export function createFly() {
  const root = new THREE.Group();
  const body = new THREE.Group();
  root.add(body);
  const shell = new THREE.MeshStandardMaterial({ color: '#766047', roughness: 0.65 });
  const dark = new THREE.MeshStandardMaterial({ color: '#272b29', roughness: 0.5 });
  const eye = new THREE.MeshStandardMaterial({ color: '#b84230', roughness: 0.35 });
  const wingMaterial = new THREE.MeshPhysicalMaterial({ color: '#c8e6df', transparent: true, opacity: 0.45, roughness: 0.25, side: THREE.DoubleSide });
  const sphere = new THREE.SphereGeometry(1, 24, 16);
  function ellipsoid(parent: THREE.Object3D, material: THREE.Material, position: number[], scale: number[]) {
    const mesh = new THREE.Mesh(sphere, material);
    mesh.position.set(position[0], position[1], position[2]);
    mesh.scale.set(scale[0], scale[1], scale[2]);
    mesh.castShadow = true;
    parent.add(mesh);
    return mesh;
  }
  ellipsoid(body, shell, [0, 1.15, 0], [0.48, 0.46, 0.65]);
  ellipsoid(body, shell, [0, 1.1, 0.95], [0.48, 0.4, 0.83]);
  for (let i = 0; i < 4; i++) {
    const band = new THREE.Mesh(new THREE.TorusGeometry(0.4 - i * 0.035, 0.025, 8, 32), dark);
    band.position.set(0, 1.1, 0.65 + i * 0.27);
    band.scale.y = 0.83;
    body.add(band);
  }
  const head = new THREE.Group();
  head.position.set(0, 1.28, -0.72);
  body.add(head);
  ellipsoid(head, shell, [0, 0, 0], [0.4, 0.38, 0.35]);
  for (const side of [-1, 1]) {
    ellipsoid(head, eye, [side * 0.28, 0.04, -0.12], [0.24, 0.3, 0.24]);
    ellipsoid(head, dark, [side * 0.13, 0.29, -0.28], [0.04, 0.17, 0.04]).rotation.x = -0.5;
  }
  const legs: THREE.Group[] = [];
  for (const side of [-1, 1]) {
    for (let index = 0; index < 3; index++) {
      const leg = new THREE.Group();
      leg.position.set(side * 0.3, 1.05, -0.35 + index * 0.45);
      body.add(leg);
      const points = [new THREE.Vector3(), new THREE.Vector3(side * 0.65, -0.4, (index - 1) * 0.3), new THREE.Vector3(side * 0.85, -1.02, (index - 1) * 0.48)];
      leg.add(new THREE.Mesh(new THREE.TubeGeometry(new THREE.CatmullRomCurve3(points), 12, 0.035, 6, false), dark));
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
