import * as THREE from 'three';

export function createBackdrop() {
  const root = new THREE.Group();
  const radius = 13.5;
  const openingAngle = 1.0;
  const windowBottom = 0.06;
  const windowTop = 8.85;
  const windowHeight = windowTop - windowBottom;
  const windowCenter = (windowTop + windowBottom) / 2;
  const wallMaterial = new THREE.MeshStandardMaterial({ color: '#24231f', roughness: 1, fog: false, side: THREE.BackSide });
  const addWall = (height: number, y: number, start: number, length: number) => {
    const wall = new THREE.Mesh(new THREE.CylinderGeometry(radius, radius, height, 160, 1, true, start, length), wallMaterial);
    wall.position.y = y;
    wall.receiveShadow = true;
    root.add(wall);
  };
  // Leave a curved panoramic opening on the far side of the room.
  addWall(32, 15.98, openingAngle / 2, Math.PI * 2 - openingAngle);
  addWall(windowBottom + 0.02, (windowBottom - 0.02) / 2, -openingAngle / 2, openingAngle);
  addWall(31.98 - windowTop, (31.98 + windowTop) / 2, -openingAngle / 2, openingAngle);

  const frameMaterial = new THREE.MeshStandardMaterial({ color: '#151c21', metalness: 0.65, roughness: 0.32 });
  for (const y of [windowBottom, windowTop]) {
    const frame = new THREE.Mesh(new THREE.CylinderGeometry(radius - 0.04, radius - 0.04, 0.16, 64, 1, true, -openingAngle / 2, openingAngle), frameMaterial);
    frame.material.side = THREE.DoubleSide;
    frame.position.y = y;
    root.add(frame);
  }
  for (let index = 0; index <= 4; index++) {
    const angle = -openingAngle / 2 + openingAngle * index / 4;
    const frame = new THREE.Mesh(new THREE.BoxGeometry(0.1, windowHeight, 0.2), frameMaterial);
    frame.position.set(Math.sin(angle) * radius, windowCenter, Math.cos(angle) * radius);
    frame.rotation.y = angle;
    root.add(frame);
  }
  const glass = new THREE.Mesh(
    new THREE.CylinderGeometry(radius, radius, windowHeight, 64, 1, true, -openingAngle / 2, openingAngle),
    new THREE.MeshBasicMaterial({ color: '#9ac5df', transparent: true, opacity: 0.08, depthWrite: false, side: THREE.DoubleSide, fog: false }),
  );
  glass.position.y = windowCenter;
  root.add(glass);
  const cove = new THREE.Mesh(
    new THREE.CylinderGeometry(radius - 0.08, radius - 0.08, 0.075, 160, 1, true),
    new THREE.MeshBasicMaterial({ color: new THREE.Color('#ffb34f').multiplyScalar(1.35), side: THREE.BackSide, toneMapped: false, fog: false }),
  );
  cove.position.y = windowTop + 0.1;
  root.add(cove);
  return root;
}
