import * as THREE from 'three';

export function createCoffeeCup() {
  const root = new THREE.Group();
  const ceramic = new THREE.MeshPhysicalMaterial({ color: '#e7a34b', roughness: 0.3, clearcoat: 0.65 });
  const cream = new THREE.MeshStandardMaterial({ color: '#ffebc8', roughness: 0.55 });
  const coffee = new THREE.MeshPhysicalMaterial({ color: '#422017', roughness: 0.24, clearcoat: 0.8 });
  const profile = [
    [0, -0.18], [0.15, -0.18], [0.18, -0.15], [0.22, 0.15],
    [0.22, 0.18], [0.195, 0.18], [0.19, 0.14], [0.155, -0.12], [0, -0.12],
  ].map(([radius, height]) => new THREE.Vector2(radius, height));
  const shell = new THREE.Mesh(new THREE.LatheGeometry(profile, 40), ceramic);
  shell.castShadow = true;
  root.add(shell);
  const rim = new THREE.Mesh(new THREE.TorusGeometry(0.207, 0.014, 10, 40), cream);
  rim.rotation.x = Math.PI / 2;
  rim.position.y = 0.176;
  root.add(rim);
  const handle = new THREE.Mesh(new THREE.TorusGeometry(0.115, 0.034, 12, 32), ceramic);
  handle.position.set(0.25, 0.01, 0);
  handle.scale.y = 1.15;
  handle.castShadow = true;
  root.add(handle);
  const drink = new THREE.Mesh(new THREE.CircleGeometry(0.19, 40), coffee);
  drink.rotation.x = -Math.PI / 2;
  drink.position.y = 0.135;
  root.add(drink);
  const crema = new THREE.Mesh(new THREE.RingGeometry(0.16, 0.188, 40), cream);
  crema.rotation.x = -Math.PI / 2;
  crema.position.y = 0.138;
  root.add(crema);
  const emblem = new THREE.Mesh(new THREE.SphereGeometry(1, 16, 12), cream);
  emblem.position.set(0, -0.015, -0.205);
  emblem.scale.set(0.075, 0.085, 0.015);
  root.add(emblem);

  const steam = Array.from({ length: 5 }, (_, index) => {
    const points = Array.from({ length: 18 }, (_, step) => {
      const progress = step / 17;
      return new THREE.Vector3(Math.sin(progress * Math.PI * 3 + index) * 0.045 * progress, progress * 0.45, 0);
    });
    const material = new THREE.MeshBasicMaterial({ color: '#fcebd3', transparent: true, opacity: 0, depthWrite: false });
    const wisp = new THREE.Mesh(new THREE.TubeGeometry(new THREE.CatmullRomCurve3(points), 24, 0.012, 6, false), material);
    root.add(wisp);
    return wisp;
  });

  return {
    root,
    update(time: number, sip: number) {
      root.position.set(0.52 - sip * 0.4, 0.56 + sip * 0.23, -1.13 - sip * 0.25);
      root.rotation.x = -sip * 0.08;
      steam.forEach((wisp, index) => {
        const progress = (time * 0.32 + index / steam.length) % 1;
        const angle = index * 2.4;
        wisp.position.set(Math.cos(angle) * 0.08 + Math.sin(time + index) * progress * 0.06, 0.19 + progress * 0.45, Math.sin(angle) * 0.08);
        wisp.scale.setScalar(0.55 + progress * 0.65);
        wisp.rotation.y = angle + Math.sin(time * 0.7 + index) * 0.4;
        wisp.material.opacity = Math.sin(progress * Math.PI) * 0.22;
      });
    },
  };
}
