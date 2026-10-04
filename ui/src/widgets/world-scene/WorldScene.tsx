import { useEffect, useRef } from 'react';
import * as THREE from 'three';
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js';
import { createFly } from '../../entities/fly/model';
import type { FlyState } from '../../entities/fly/state';

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
    element.appendChild(renderer.domElement);
    const scene = new THREE.Scene();
    scene.background = new THREE.Color('#10171c');
    scene.fog = new THREE.Fog('#10171c', 12, 28);
    const camera = new THREE.PerspectiveCamera(40, 1, 0.1, 60);
    camera.position.set(5, 4, -7);
    const controls = new OrbitControls(camera, renderer.domElement);
    controls.target.set(0, 0.8, 0);
    controls.enableDamping = true;
    controls.minDistance = 3;
    controls.maxDistance = 15;
    controls.maxPolarAngle = Math.PI / 2 - 0.03;
    scene.add(new THREE.HemisphereLight('#d8f5ec', '#343326', 2.4));
    const key = new THREE.DirectionalLight('#ffe9c6', 4);
    key.position.set(3, 7, -4);
    key.castShadow = true;
    scene.add(key);
    const accent = new THREE.PointLight('#6cd9b7', 16, 12);
    accent.position.set(-3, 3, 2);
    scene.add(accent);
    const floor = new THREE.Mesh(new THREE.CircleGeometry(12, 64), new THREE.MeshStandardMaterial({ color: '#172329', roughness: 0.9 }));
    floor.rotation.x = -Math.PI / 2;
    floor.position.y = -0.02;
    floor.receiveShadow = true;
    scene.add(floor, new THREE.GridHelper(20, 40, '#35524c', '#233831'));
    const fly = createFly();
    scene.add(fly.root);
    const observer = new ResizeObserver(() => {
      const { width, height } = element.getBoundingClientRect();
      if (!width || !height) return;
      renderer.setSize(width, height);
      camera.aspect = width / height;
      camera.updateProjectionMatrix();
    });
    observer.observe(element);
    renderer.setAnimationLoop((milliseconds) => {
      fly.update(milliseconds / 1000, current.current);
      accent.color.set(current.current.behavior === 'alarmed' ? '#ff4438' : '#6cd9b7');
      controls.update();
      renderer.render(scene, camera);
    });
    return () => {
      observer.disconnect();
      renderer.setAnimationLoop(null);
      controls.dispose();
      const geometries = new Set<THREE.BufferGeometry>();
      const materials = new Set<THREE.Material>();
      scene.traverse((object) => {
        if (object instanceof THREE.Mesh || object instanceof THREE.LineSegments) {
          geometries.add(object.geometry);
          (Array.isArray(object.material) ? object.material : [object.material]).forEach(material => materials.add(material));
        }
      });
      geometries.forEach(geometry => geometry.dispose());
      materials.forEach(material => material.dispose());
      renderer.dispose();
      renderer.domElement.remove();
    };
  }, []);
  return <div className="scene" ref={container} aria-label="Interactive three-dimensional fly. Drag to orbit, scroll to zoom." />;
}
