import * as THREE from 'three';

export function createBackdrop() {
  const wall = new THREE.Mesh(
    new THREE.CylinderGeometry(13.5, 13.5, 32, 160, 1, true),
    new THREE.MeshStandardMaterial({ color: '#24231f', roughness: 1, fog: false, side: THREE.BackSide }),
  );
  // A continuous inward-facing wall keeps side edges and the top out of view.
  wall.position.y = 15.98;
  wall.receiveShadow = true;
  return wall;
}
