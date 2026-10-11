import * as THREE from 'three';

function ease(value: number) {
  const clamped = Math.max(0, Math.min(1, value));
  return clamped * clamped * (3 - 2 * clamped);
}

export function smokingPose(clock: number) {
  const phase = clock % 12;
  const draw = ease((phase - 3.4) / 0.8) * (1 - ease((phase - 4.8) / 0.9));
  const exhale = phase > 5.8 && phase < 7.3 ? Math.sin((phase - 5.8) / 1.5 * Math.PI) : 0;
  return { draw, exhale };
}

export function createSmokingBreak() {
  const root = new THREE.Group();
  const cigarette = new THREE.Group();
  root.add(cigarette);
  const paper = new THREE.MeshStandardMaterial({ color: '#eee5d4', roughness: 0.9 });
  const filter = new THREE.MeshStandardMaterial({ color: '#c48b4e', roughness: 0.8 });
  const ash = new THREE.MeshStandardMaterial({ color: '#6e6961', roughness: 1 });
  const ember = new THREE.MeshStandardMaterial({ color: '#ea642e', emissive: '#ff4c18', emissiveIntensity: 0.5, roughness: 0.8 });
  function cylinder(material: THREE.Material, radius: number, length: number, z: number) {
    const mesh = new THREE.Mesh(new THREE.CylinderGeometry(radius, radius, length, 16), material);
    mesh.rotation.x = Math.PI / 2;
    mesh.position.z = z;
    mesh.castShadow = true;
    cigarette.add(mesh);
  }
  cylinder(filter, 0.033, 0.13, 0.025);
  cylinder(paper, 0.03, 0.23, -0.155);
  cylinder(ember, 0.031, 0.018, -0.279);
  cylinder(ash, 0.032, 0.035, -0.305);
  for (const z of [-0.005, 0.025, 0.055]) {
    const band = new THREE.Mesh(new THREE.TorusGeometry(0.033, 0.002, 5, 16), paper);
    band.position.z = z;
    cigarette.add(band);
  }
  const tip = new THREE.Vector3();
  const grip = new THREE.Vector3();
  const wisps = Array.from({ length: 4 }, (_, index) => {
    const points = Array.from({ length: 20 }, (_, step) => {
      const progress = step / 19;
      return new THREE.Vector3(Math.sin(progress * Math.PI * 3 + index) * progress * 0.06, progress * 0.48, Math.cos(progress * Math.PI * 2) * progress * 0.025);
    });
    const material = new THREE.MeshBasicMaterial({ color: '#bcc5c1', transparent: true, opacity: 0, depthWrite: false });
    const mesh = new THREE.Mesh(new THREE.TubeGeometry(new THREE.CatmullRomCurve3(points), 28, 0.009, 5, false), material);
    root.add(mesh);
    return mesh;
  });
  const puffGeometry = new THREE.SphereGeometry(1, 12, 8);
  const puffs = Array.from({ length: 12 }, () => {
    const material = new THREE.MeshBasicMaterial({ color: '#c6cfcb', transparent: true, opacity: 0, depthWrite: false });
    const mesh = new THREE.Mesh(puffGeometry, material);
    mesh.visible = false;
    root.add(mesh);
    return { mesh, age: Infinity, origin: new THREE.Vector3() };
  });
  let nextPuff = 0;
  let emissionTime = 0;
  let smokeTime = 0;

  return {
    root,
    update(clock: number, delta: number, mouth: THREE.Vector3) {
      const { draw, exhale } = smokingPose(clock);
      smokeTime += delta;
      cigarette.position.set(0.65 - draw * 0.59, 0.64 + draw * 0.37, -0.98 - draw * 0.22);
      cigarette.rotation.set(0.2 * (1 - draw), -0.2 * (1 - draw), -0.12 * (1 - draw));
      grip.copy(cigarette.position);
      tip.set(0, 0, -0.315).applyEuler(cigarette.rotation).add(cigarette.position);
      ember.emissiveIntensity = 0.5 + draw * 2.5;
      wisps.forEach((mesh, index) => {
        const progress = (smokeTime * 0.3 + index / wisps.length) % 1;
        mesh.position.copy(tip);
        mesh.position.y += progress * 0.5;
        mesh.position.x += Math.sin(smokeTime * 0.7 + index) * progress * 0.07;
        mesh.scale.setScalar(0.5 + progress * 0.8);
        mesh.rotation.y = index * 1.7 + smokeTime * 0.2;
        mesh.material.opacity = Math.sin(progress * Math.PI) * (0.14 + draw * 0.06);
      });
      emissionTime += delta;
      if (exhale > 0.15 && emissionTime >= 0.13) {
        const puff = puffs[nextPuff];
        puff.age = 0;
        puff.origin.copy(mouth);
        puff.origin.z -= 0.04;
        nextPuff = (nextPuff + 1) % puffs.length;
        emissionTime = 0;
      }
      puffs.forEach((puff, index) => {
        puff.age += delta;
        puff.mesh.visible = puff.age < 2.8;
        if (!puff.mesh.visible) return;
        const age = puff.age;
        const size = 0.035 + age * 0.12;
        puff.mesh.position.copy(puff.origin);
        puff.mesh.position.z -= age * 0.5;
        puff.mesh.position.y += age * 0.18;
        puff.mesh.position.x += Math.sin(age * 2 + index) * age * 0.05;
        puff.mesh.scale.set(size * 1.25, size, size * 1.1);
        puff.mesh.material.opacity = (1 - age / 2.8) ** 2 * 0.13;
      });
      return grip;
    },
    reset() {
      smokeTime = emissionTime = nextPuff = 0;
      puffs.forEach(puff => { puff.age = Infinity; puff.mesh.visible = false; });
    },
  };
}
