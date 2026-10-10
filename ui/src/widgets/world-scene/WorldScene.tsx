import { useEffect, useRef } from 'react';
import * as THREE from 'three';
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js';
import { EffectComposer } from 'three/examples/jsm/postprocessing/EffectComposer.js';
import { RenderPass } from 'three/examples/jsm/postprocessing/RenderPass.js';
import { UnrealBloomPass } from 'three/examples/jsm/postprocessing/UnrealBloomPass.js';
import { OutputPass } from 'three/examples/jsm/postprocessing/OutputPass.js';
import { createFly } from '../../entities/fly/model';
import { createPlants } from '../../entities/plants/model';
import { createDecorativeLighting } from '../../entities/lighting/model';
import { createBackdrop } from '../../entities/room/model';
import { createCityBackdrop } from '../../entities/city-backdrop/model';
import { createSofa } from '../../entities/sofa/model';
import { createRug } from '../../entities/room/rug';
import { createAtmosphericHaze } from '../../entities/atmosphere/model';
import { createServerRacks } from '../../entities/server-racks/model';
import type { FlyState } from '../../entities/fly/state';
import { alarmSirenStrength } from '../../features/presentation-rules/alarm';

export function WorldScene({ state }: { state: FlyState }) {
  const container = useRef<HTMLDivElement>(null);
  const current = useRef(state);
  current.current = state;
  useEffect(() => {
    const element = container.current!;
    let renderer: THREE.WebGLRenderer;
    try { renderer = new THREE.WebGLRenderer({ antialias: true }); }
    catch { element.textContent = 'WebGL is unavailable. Try a browser with hardware acceleration.'; return; }
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
    renderer.shadowMap.enabled = true;
    renderer.shadowMap.type = THREE.PCFSoftShadowMap;
    renderer.toneMapping = THREE.ACESFilmicToneMapping;
    element.appendChild(renderer.domElement);
    const scene = new THREE.Scene();
    scene.background = new THREE.Color('#48443e');
    scene.fog = new THREE.Fog('#48443e', 18, 42);
    const camera = new THREE.PerspectiveCamera(40, 1, 0.1, 60);
    camera.position.set(5, 4, -7);
    const controls = new OrbitControls(camera, renderer.domElement);
    controls.target.set(0, 0.8, 0);
    controls.enableDamping = true;
    controls.minDistance = 3;
    // Keep the orbit camera inside the inward-facing room wall.
    controls.maxDistance = 12;
    controls.maxPolarAngle = Math.PI / 2 - 0.03;
    scene.add(new THREE.HemisphereLight('#fff0da', '#645b4b', 2));
    const key = new THREE.DirectionalLight('#ffe9c9', 3);
    key.position.set(3, 7, -4);
    key.castShadow = true;
    scene.add(key);
    const accent = new THREE.PointLight('#ffc078', 8, 12);
    accent.position.set(-3, 3, 2);
    scene.add(accent);
    const siren = new THREE.Group();
    scene.add(siren);
    for (const side of [-1, 1]) {
      const beam = new THREE.SpotLight('#ff3026', 0, 14, Math.PI / 5, 0.65, 1.5);
      beam.position.set(0, 4, 0);
      beam.target.position.set(side * 5, 0, 0);
      siren.add(beam, beam.target);
    }
    const alarmMaterial = new THREE.MeshBasicMaterial({ color: '#ff382b', transparent: true, opacity: 0, depthWrite: false, side: THREE.DoubleSide });
    const alarmRing = new THREE.Mesh(new THREE.RingGeometry(2.1, 2.25, 80), alarmMaterial);
    alarmRing.rotation.x = -Math.PI / 2;
    alarmRing.position.y = 0.005;
    scene.add(alarmRing);
    const normalBackground = new THREE.Color('#48443e');
    const warningColor = new THREE.Color('#ffcd38');
    const dangerColor = new THREE.Color('#ff3026');
    const alarmColor = new THREE.Color();
    const warningBackground = new THREE.Color('#30291a');
    const dangerBackground = new THREE.Color('#36151d');
    const alarmBackground = new THREE.Color();
    const floor = new THREE.Mesh(new THREE.CircleGeometry(14, 64), new THREE.MeshStandardMaterial({ color: '#343a36', roughness: 0.9 }));
    floor.rotation.x = -Math.PI / 2;
    floor.position.y = -0.02;
    floor.receiveShadow = true;
    scene.add(floor, new THREE.GridHelper(20, 40, '#35524c', '#233831'));
    scene.add(createPlants());
    scene.add(createBackdrop());
    scene.add(createSofa());
    const cityBackdrop = createCityBackdrop();
    scene.add(cityBackdrop.root);
    scene.add(createRug());
    scene.add(createDecorativeLighting());
    scene.add(createAtmosphericHaze());
    const serverRacks = createServerRacks();
    scene.add(serverRacks.root);
    const soundChannel = typeof BroadcastChannel !== 'undefined' ? new BroadcastChannel('drosophila.presentation-sound.v1') : null;
    const fly = createFly(
      () => soundChannel?.postMessage({ type: 'coffee.sip', timestamp: Date.now() }),
      () => soundChannel?.postMessage({ type: 'working.keypress', timestamp: Date.now() }),
    );
    scene.add(fly.root);
    const composer = new EffectComposer(renderer);
    const bloom = new UnrealBloomPass(new THREE.Vector2(1, 1), 0.12, 0.55, 1.2);
    const output = new OutputPass();
    composer.addPass(new RenderPass(scene, camera));
    composer.addPass(bloom);
    composer.addPass(output);
    const observer = new ResizeObserver(() => {
      const { width, height } = element.getBoundingClientRect();
      if (!width || !height) return;
      renderer.setSize(width, height);
      composer.setSize(width, height);
      camera.aspect = width / height;
      camera.updateProjectionMatrix();
    });
    observer.observe(element);
    renderer.setAnimationLoop((milliseconds) => {
      fly.update(milliseconds / 1000, current.current);
      const time = milliseconds / 1000;
      serverRacks.update(time);
      const alarmed = current.current.behavior === 'alarmed';
      const activity = current.current.activity;
      const sirenStrength = alarmed ? alarmSirenStrength(activity) : 0;
      const pulse = 0.5 + 0.5 * Math.sin(time * Math.PI * 3);
      alarmColor.copy(warningColor).lerp(dangerColor, Math.min(1, activity / 0.8));
      alarmBackground.copy(warningBackground).lerp(dangerBackground, Math.min(1, activity / 0.8));
      if (alarmed) accent.color.copy(alarmColor);
      else accent.color.set('#ffc078');
      accent.intensity = alarmed ? 20 + activity * 12 + pulse * (6 + activity * 16) : 8;
      siren.rotation.y = time * 4;
      siren.children.forEach(object => {
        if (object instanceof THREE.SpotLight) {
          object.intensity = sirenStrength * 80;
          object.color.copy(alarmColor);
        }
      });
      alarmMaterial.color.copy(alarmColor);
      alarmMaterial.opacity = alarmed ? 0.2 + pulse * (0.15 + activity * 0.2) : 0;
      alarmRing.scale.setScalar(1 + (alarmed ? pulse * 0.06 : 0));
      (scene.background as THREE.Color).copy(normalBackground).lerp(alarmBackground, alarmed ? 0.45 + pulse * 0.25 : 0);
      (scene.fog as THREE.Fog).color.copy(scene.background as THREE.Color);
      controls.update();
      composer.render();
    });
    return () => {
      observer.disconnect();
      soundChannel?.close();
      renderer.setAnimationLoop(null);
      controls.dispose();
      bloom.dispose();
      output.dispose();
      composer.dispose();
      const geometries = new Set<THREE.BufferGeometry>();
      const materials = new Set<THREE.Material>();
      scene.traverse((object) => {
        if (object instanceof THREE.InstancedMesh) object.dispose();
        if (object instanceof THREE.DirectionalLight || object instanceof THREE.SpotLight || object instanceof THREE.PointLight) {
          object.shadow.dispose();
        }
        if (object instanceof THREE.Mesh || object instanceof THREE.LineSegments) {
          geometries.add(object.geometry);
          (Array.isArray(object.material) ? object.material : [object.material]).forEach(material => materials.add(material));
        }
      });
      geometries.forEach(geometry => geometry.dispose());
      const textures = new Set<THREE.Texture>();
      materials.forEach(material => {
        if (material instanceof THREE.MeshStandardMaterial && material.bumpMap) {
          textures.add(material.bumpMap);
        }
      });
      textures.forEach(texture => texture.dispose());
      materials.forEach(material => material.dispose());
      renderer.dispose();
      renderer.domElement.remove();
    };
  }, []);
  return <div className="scene" ref={container} aria-label="Interactive three-dimensional fly. Drag to orbit, scroll to zoom." />;
}
