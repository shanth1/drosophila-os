import * as THREE from 'three';
import { eveningCityTheme, initialCityState, type CityBackdropState, type CityBackdropTheme } from './state';

export function createCityBackdrop() {
  const root = new THREE.Group();
  const architecture = new THREE.Group();
  // Future moving objects belong here, independently of static architecture.
  const airspace = new THREE.Group();
  root.add(architecture, airspace);
  const material = () => new THREE.MeshBasicMaterial({ fog: false });
  const skyMaterial = material();
  skyMaterial.side = THREE.BackSide;
  root.add(new THREE.Mesh(new THREE.SphereGeometry(45, 48, 24), skyMaterial));
  const glass = material();
  const distantGlass = material();
  const bands = material();
  const windows = material();
  const crowns = material();
  const box = new THREE.BoxGeometry(1, 1, 1);
  const panel = new THREE.PlaneGeometry(0.1, 0.19);
  const spireGeometry = new THREE.ConeGeometry(1, 1, 4);

  function tower(x: number, z: number, height: number, width: number, twist: number, distant: boolean) {
    const floors = Math.floor(height / 0.34);
    for (let floor = 0; floor < floors; floor++) {
      const progress = floor / floors;
      const taper = 1 - progress * 0.28;
      const level = new THREE.Group();
      level.position.set(x + Math.sin(progress * Math.PI) * twist * 0.4, -1.5 + floor * 0.34, z);
      level.rotation.y = progress * twist;
      architecture.add(level);
      const facade = new THREE.Mesh(box, distant ? distantGlass : glass);
      facade.scale.set(width * taper, 0.34, width * 0.78 * taper);
      level.add(facade);
      const band = new THREE.Mesh(box, floor === floors - 1 ? crowns : bands);
      band.scale.set(width * taper + 0.025, 0.025, width * 0.78 * taper + 0.025);
      band.position.y = 0.16;
      level.add(band);
      for (const side of [-1, 1]) {
        for (let column = 0; column < 5; column++) {
          if ((floor * 7 + column * 3 + Math.round(x * 10)) % 6 === 0) continue;
          const light = new THREE.Mesh(panel, windows);
          light.position.set((column - 2) * width * taper / 6, 0, side * (width * 0.39 * taper + 0.006));
          light.rotation.y = side === -1 ? Math.PI : 0;
          level.add(light);
        }
      }
    }
    if (twist === 0) {
      const spire = new THREE.Mesh(spireGeometry, glass);
      spire.scale.set(width * 0.36, 1.8, width * 0.36);
      spire.position.set(x, floors * 0.34 - 0.6, z);
      spire.rotation.y = Math.PI / 4;
      architecture.add(spire);
    }
  }

  // Deliberately composed office silhouettes: twisting, tapered, and paired towers.
  for (let index = 0; index < 9; index++) {
    tower((index - 4) * 3.8, 29, 6 + (index * 7 % 9), 2.1, index % 3 === 0 ? 0.8 : 0, true);
  }
  tower(-9, 21, 9.8, 2.5, 0, false);
  tower(-5.5, 22, 12.2, 2.6, 1.15, false);
  tower(-1.6, 24, 10.7, 2.1, 0.25, false);
  tower(1.2, 24, 12.8, 2.1, 0.25, false);
  tower(5.2, 22, 11.4, 2.7, -0.9, false);
  tower(9.1, 23, 9.1, 2.2, 0, false);

  // Batch repeated facade details so the skyline costs only a few draw calls.
  architecture.updateMatrixWorld(true);
  const batches = new Map<THREE.BufferGeometry, Map<THREE.Material, THREE.Matrix4[]>>();
  architecture.traverse(object => {
    if (!(object instanceof THREE.Mesh)) return;
    const material = object.material as THREE.Material;
    let materials = batches.get(object.geometry);
    if (!materials) batches.set(object.geometry, materials = new Map());
    let transforms = materials.get(material);
    if (!transforms) materials.set(material, transforms = []);
    transforms.push(object.matrixWorld.clone());
  });
  architecture.clear();
  batches.forEach((materials, geometry) => materials.forEach((transforms, material) => {
    const instances = new THREE.InstancedMesh(geometry, material, transforms.length);
    transforms.forEach((transform, index) => instances.setMatrixAt(index, transform));
    instances.instanceMatrix.needsUpdate = true;
    architecture.add(instances);
  }));

  let theme: CityBackdropTheme = { ...eveningCityTheme };
  let state: CityBackdropState = { ...initialCityState };
  function applyAppearance() {
    skyMaterial.color.set(theme.sky).multiplyScalar(state.skyBrightness);
    glass.color.set(theme.glass);
    distantGlass.color.set(theme.distantGlass);
    bands.color.set(theme.floorBands);
    windows.color.set(theme.windows).multiplyScalar(state.windowGlow);
    crowns.color.set(theme.crown).multiplyScalar(state.windowGlow);
  }
  applyAppearance();
  return {
    root,
    airspace,
    setState(next: Readonly<CityBackdropState>) {
      state = { ...next };
      applyAppearance();
    },
    setTheme(next: Readonly<CityBackdropTheme>) {
      theme = { ...next };
      applyAppearance();
    },
  };
}
