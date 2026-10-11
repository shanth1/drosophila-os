import * as THREE from 'three';

export function createOcclusionFade(objects: THREE.Object3D[]) {
  const sources = new Set<THREE.Material>();
  const entries = objects.map(object => {
    const materials = new Map<THREE.Material, THREE.Material>();
    object.traverse(child => {
      if (!(child instanceof THREE.Mesh)) return;
      const clone = (source: THREE.Material) => {
        let material = materials.get(source);
        if (!material) {
          material = source.clone();
          materials.set(source, material);
          sources.add(source);
        }
        return material;
      };
      child.material = Array.isArray(child.material) ? child.material.map(clone) : clone(child.material);
    });
    return { bounds: new THREE.Box3().setFromObject(object).expandByScalar(0.4), materials, opacity: 1 };
  });
  const direction = new THREE.Vector3();
  const hit = new THREE.Vector3();
  const ray = new THREE.Ray();
  let previousTime: number | undefined;

  return {
    update(camera: THREE.Camera, target: THREE.Vector3, time: number) {
      const delta = previousTime === undefined ? 0 : Math.min(0.1, Math.max(0, time - previousTime));
      previousTime = time;
      const distance = camera.position.distanceTo(target);
      ray.set(camera.position, direction.subVectors(target, camera.position).normalize());
      for (const entry of entries) {
        const intersection = ray.intersectBox(entry.bounds, hit);
        const blocked = entry.bounds.containsPoint(camera.position)
          || (intersection !== null && camera.position.distanceTo(hit) < distance - 0.5);
        entry.opacity = THREE.MathUtils.damp(entry.opacity, blocked ? 0.12 : 1, 10, delta);
        const faded = entry.opacity < 0.995;
        for (const [source, material] of entry.materials) {
          material.opacity = source.opacity * (faded ? entry.opacity : 1);
          const transparent = source.transparent || faded;
          if (material.transparent !== transparent) {
            material.transparent = transparent;
            material.needsUpdate = true;
          }
          material.depthWrite = source.depthWrite && !faded;
          // Preserve animated indicator brightness on the isolated fade materials.
          if (source instanceof THREE.MeshStandardMaterial && material instanceof THREE.MeshStandardMaterial) {
            material.emissiveIntensity = source.emissiveIntensity;
          }
        }
      }
    },
    dispose() {
      sources.forEach(material => material.dispose());
    },
  };
}
