import * as THREE from 'three';

export function createAnalysisProps() {
  const root = new THREE.Group();
  const magnifier = new THREE.Group();
  const report = new THREE.Group();
  root.add(magnifier, report);
  const frameMaterial = new THREE.MeshStandardMaterial({ color: '#c79e59', metalness: 0.45, roughness: 0.35 });
  const gripMaterial = new THREE.MeshStandardMaterial({ color: '#365851', roughness: 0.55 });
  const paperMaterial = new THREE.MeshStandardMaterial({ color: '#f2e7ce', roughness: 0.85 });
  const inkMaterial = new THREE.MeshStandardMaterial({ color: '#62877c', roughness: 0.8 });
  const graphMaterial = new THREE.MeshStandardMaterial({ color: '#d58a43', roughness: 0.65 });
  const frameGeometry = new THREE.TorusGeometry(0.28, 0.027, 12, 48);
  for (const z of [-0.025, 0.025]) {
    const frame = new THREE.Mesh(frameGeometry, frameMaterial);
    frame.position.z = z;
    frame.castShadow = true;
    magnifier.add(frame);
  }
  const lensGeometry = new THREE.SphereGeometry(1, 48, 32);
  lensGeometry.scale(0.257, 0.257, 0.075);
  const lens = new THREE.Mesh(lensGeometry, new THREE.MeshPhysicalMaterial({
    color: '#ffffff', transmission: 1, thickness: 0.15, ior: 1.52,
    roughness: 0.025, clearcoat: 1, clearcoatRoughness: 0.05,
  }));
  // Convex surface normals and optical thickness refract the scene behind the lens.
  magnifier.add(lens);
  const neck = new THREE.Mesh(new THREE.CylinderGeometry(0.028, 0.028, 0.1, 12), frameMaterial);
  neck.position.y = -0.32;
  magnifier.add(neck);
  const handle = new THREE.Mesh(new THREE.CylinderGeometry(0.043, 0.052, 0.31, 16), gripMaterial);
  handle.position.y = -0.5;
  handle.castShadow = true;
  magnifier.add(handle);
  for (const y of [-0.37, -0.63]) {
    const band = new THREE.Mesh(new THREE.TorusGeometry(0.046, 0.008, 6, 16), frameMaterial);
    band.rotation.x = Math.PI / 2;
    band.position.y = y;
    magnifier.add(band);
  }
  const backing = new THREE.Mesh(new THREE.BoxGeometry(0.64, 0.5, 0.035), gripMaterial);
  backing.castShadow = true;
  report.add(backing);
  const paper = new THREE.Mesh(new THREE.BoxGeometry(0.58, 0.43, 0.006), paperMaterial);
  paper.position.z = 0.022;
  report.add(paper);
  const clip = new THREE.Mesh(new THREE.BoxGeometry(0.16, 0.055, 0.025), frameMaterial);
  clip.position.set(0, 0.22, 0.025);
  report.add(clip);
  for (let row = 0; row < 3; row++) {
    const line = new THREE.Mesh(new THREE.BoxGeometry(0.46, 0.005, 0.003), inkMaterial);
    line.position.set(0, -0.12 + row * 0.085, 0.028);
    report.add(line);
  }
  const title = new THREE.Mesh(new THREE.BoxGeometry(0.24, 0.015, 0.003), inkMaterial);
  title.position.set(-0.1, 0.14, 0.028);
  report.add(title);
  // A decorative chart, independent of host measurements or neural topology.
  const graph = new THREE.Mesh(new THREE.TubeGeometry(new THREE.CatmullRomCurve3([
    new THREE.Vector3(-0.22, -0.11, 0.032), new THREE.Vector3(-0.13, -0.06, 0.032),
    new THREE.Vector3(-0.04, -0.09, 0.032), new THREE.Vector3(0.05, 0.04, 0.032),
    new THREE.Vector3(0.14, 0.01, 0.032), new THREE.Vector3(0.23, 0.09, 0.032),
  ]), 32, 0.008, 6, false), graphMaterial);
  report.add(graph);
  const magnifierGrip = new THREE.Vector3();
  const reportGrip = new THREE.Vector3();

  return {
    root,
    update(clock: number) {
      const scan = Math.sin(clock * 0.8);
      const lean = Math.sin(clock * 0.45);
      magnifier.position.set(0.29 + scan * 0.06, 1.28 + Math.sin(clock * 0.65) * 0.035, -1.34);
      magnifier.rotation.set(-0.08, -scan * 0.08, -0.1 + lean * 0.06);
      report.position.set(-0.22, 0.65 + lean * 0.025, -1.23);
      report.rotation.set(-0.95, 0.12 + scan * 0.04, -0.06);
      magnifierGrip.set(0, -0.5, 0).applyEuler(magnifier.rotation).add(magnifier.position);
      reportGrip.set(-0.27, -0.18, 0.022).applyEuler(report.rotation).add(report.position);
      return { magnifierGrip, reportGrip, scan, lean };
    },
  };
}
