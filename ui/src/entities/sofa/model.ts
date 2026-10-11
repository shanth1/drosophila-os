import * as THREE from 'three';
import { RoundedBoxGeometry } from 'three/examples/jsm/geometries/RoundedBoxGeometry.js';

export function createSofa() {
  const root = new THREE.Group();
  root.position.set(0.85, 0, 9.7);
  root.rotation.y = 0.16;
  root.scale.setScalar(1.15);

  const size = 128;
  const fibers = new Uint8Array(size * size * 4);
  let seed = 419;
  for (let index = 0; index < size * size; index++) {
    seed = (Math.imul(seed, 1664525) + 1013904223) >>> 0;
    const x = index % size;
    const y = Math.floor(index / size);
    const loop = Math.sin(x * Math.PI / 3) * Math.sin(y * Math.PI / 3);
    const value = Math.round(130 + loop * 35 + (seed >>> 27));
    fibers.set([value, value, value, 255], index * 4);
  }
  const boucle = new THREE.DataTexture(fibers, size, size);
  boucle.wrapS = boucle.wrapT = THREE.RepeatWrapping;
  boucle.repeat.set(5, 5);
  boucle.magFilter = THREE.LinearFilter;
  boucle.minFilter = THREE.LinearMipmapLinearFilter;
  boucle.generateMipmaps = true;
  boucle.needsUpdate = true;

  const upholstery = new THREE.MeshStandardMaterial({ color: '#d4c7b3', roughness: 1, bumpMap: boucle, bumpScale: 0.025 });
  const piping = new THREE.MeshStandardMaterial({ color: '#b5a58d', roughness: 1 });
  const sage = new THREE.MeshStandardMaterial({ color: '#798674', roughness: 1, bumpMap: boucle, bumpScale: 0.02 });
  const clay = new THREE.MeshStandardMaterial({ color: '#ae8067', roughness: 1, bumpMap: boucle, bumpScale: 0.02 });
  const feet = new THREE.MeshStandardMaterial({ color: '#493b30', roughness: 0.8 });

  function cushion(width: number, height: number, depth: number, radius: number, material: THREE.Material, x: number, y: number, z: number) {
    const object = new THREE.Mesh(new RoundedBoxGeometry(width, height, depth, 5, radius), material);
    object.position.set(x, y, z);
    object.castShadow = true;
    object.receiveShadow = true;
    root.add(object);
    return object;
  }

  for (const x of [-2.05, 2.05]) {
    for (const z of [-0.57, 0.57]) cushion(0.16, 0.22, 0.16, 0.035, feet, x, 0.11, z);
  }
  cushion(5.2, 0.5, 1.85, 0.24, upholstery, 0, 0.43, 0);
  const back = cushion(4.95, 1.05, 0.55, 0.25, upholstery, 0, 1.12, 0.68);
  back.rotation.x = -0.1;
  for (const x of [-2.34, 2.34]) {
    cushion(0.58, 0.85, 1.78, 0.28, upholstery, x, 0.85, -0.04);
  }
  for (const x of [-1.34, 0, 1.34]) {
    cushion(1.33, 0.05, 1.43, 0.024, piping, x, 0.695, -0.15);
    cushion(1.32, 0.29, 1.42, 0.14, upholstery, x, 0.83, -0.15);
  }
  const leftPillow = cushion(0.78, 0.72, 0.26, 0.125, sage, -1.55, 1.18, 0.23);
  leftPillow.rotation.set(-0.2, 0.1, -0.17);
  const rightPillow = cushion(0.68, 0.65, 0.27, 0.125, clay, 1.57, 1.15, 0.2);
  rightPillow.rotation.set(-0.24, -0.12, 0.2);
  return root;
}
