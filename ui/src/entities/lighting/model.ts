import * as THREE from 'three';

export function createDecorativeLighting() {
  const root = new THREE.Group();
  const metal = new THREE.MeshStandardMaterial({ color: '#514638', metalness: 0.65, roughness: 0.45 });
  const stone = new THREE.MeshStandardMaterial({ color: '#b8a58a', roughness: 0.9 });
  const diffuser = new THREE.MeshBasicMaterial({
    color: new THREE.Color('#ffb34f').multiplyScalar(1.35), toneMapped: false, fog: false,
  });
  const warmWhite = '#ffad58';
  const haloMaterial = new THREE.ShaderMaterial({
    uniforms: { color: { value: new THREE.Color('#ffad58') } },
    vertexShader: `
      varying vec2 haloUV;
      void main() {
        haloUV = uv;
        vec4 center = modelViewMatrix * vec4(0.0, 0.0, 0.0, 1.0);
        vec2 size = vec2(length(modelMatrix[0].xyz), length(modelMatrix[1].xyz));
        center.xy += position.xy * size;
        gl_Position = projectionMatrix * center;
      }
    `,
    fragmentShader: `
      uniform vec3 color;
      varying vec2 haloUV;
      void main() {
        float radius = length((haloUV - 0.5) * 2.0);
        float alpha = pow(max(0.0, 1.0 - radius), 2.1) * 0.13;
        gl_FragColor = vec4(color, alpha);
        #include <colorspace_fragment>
      }
    `,
    transparent: true,
    blending: THREE.AdditiveBlending,
    depthWrite: false,
    toneMapped: false,
  });
  const beamHazeMaterial = new THREE.ShaderMaterial({
    uniforms: { color: { value: new THREE.Color(warmWhite) } },
    vertexShader: `
      varying vec2 hazeUV;
      varying vec3 viewNormal;
      varying vec3 viewPosition;
      void main() {
        hazeUV = uv;
        viewNormal = normalize(normalMatrix * normal);
        vec4 positionInView = modelViewMatrix * vec4(position, 1.0);
        viewPosition = positionInView.xyz;
        gl_Position = projectionMatrix * positionInView;
      }
    `,
    fragmentShader: `
      uniform vec3 color;
      varying vec2 hazeUV;
      varying vec3 viewNormal;
      varying vec3 viewPosition;
      void main() {
        float facing = abs(dot(normalize(viewNormal), normalize(-viewPosition)));
        float fade = pow(1.0 - hazeUV.y, 1.6);
        gl_FragColor = vec4(color, 0.016 * fade * pow(facing, 0.6));
        #include <colorspace_fragment>
      }
    `,
    transparent: true,
    blending: THREE.AdditiveBlending,
    depthWrite: false,
    side: THREE.DoubleSide,
    forceSinglePass: true,
    toneMapped: false,
  });

  function halo(parent: THREE.Object3D, height: number, width: number, haloHeight = width) {
    const aura = mesh(parent, new THREE.PlaneGeometry(1, 1), haloMaterial, false);
    aura.position.y = height;
    aura.scale.set(width, haloHeight, 1);
    aura.frustumCulled = false;
  }

  function mesh(parent: THREE.Object3D, geometry: THREE.BufferGeometry, material: THREE.Material, castsShadow = true) {
    const object = new THREE.Mesh(geometry, material);
    object.castShadow = castsShadow;
    object.receiveShadow = castsShadow;
    parent.add(object);
    return object;
  }

  function base(parent: THREE.Object3D, radius: number, height: number) {
    const foot = mesh(parent, new THREE.CylinderGeometry(radius, radius * 1.04, height, 32), stone);
    foot.position.y = height / 2;
  }

  function glow(parent: THREE.Object3D, height: number, intensity: number, distance: number) {
    const light = new THREE.PointLight(warmWhite, intensity, distance, 2);
    light.position.y = height;
    parent.add(light);
  }

  function floorCylinder(x: number, z: number, height: number, radius: number) {
    const lamp = new THREE.Group();
    lamp.position.set(x, 0, z);
    root.add(lamp);
    base(lamp, radius + 0.04, 0.1);
    const shade = mesh(lamp, new THREE.CylinderGeometry(radius, radius, height, 32), diffuser, false);
    shade.position.y = height / 2 + 0.1;
    const cap = mesh(lamp, new THREE.CylinderGeometry(radius, radius, 0.035, 32), stone);
    cap.position.y = height + 0.12;
    glow(lamp, height * 0.25 + 0.1, 6, 4.5);
    glow(lamp, height * 0.75 + 0.1, 6, 4.5);
    halo(lamp, height / 2 + 0.1, radius * 8, height * 2.6);
  }

  function arcLamp() {
    const lamp = new THREE.Group();
    lamp.position.set(6.7, 0, 3.6);
    lamp.rotation.y = 0.45;
    lamp.scale.setScalar(1.2);
    root.add(lamp);
    base(lamp, 0.43, 0.16);
    const curve = new THREE.CatmullRomCurve3([
      new THREE.Vector3(0, 0.15, 0), new THREE.Vector3(0, 2.6, 0),
      new THREE.Vector3(-0.35, 3.65, 0), new THREE.Vector3(-1.45, 4.05, 0),
      new THREE.Vector3(-2.3, 3.6, 0),
    ]);
    mesh(lamp, new THREE.TubeGeometry(curve, 48, 0.035, 8, false), metal);
    const pendant = new THREE.Group();
    pendant.position.set(-2.3, 3.35, 0);
    lamp.add(pendant);
    const collar = mesh(pendant, new THREE.CylinderGeometry(0.09, 0.09, 0.18, 20), metal);
    collar.position.y = 0.5;
    mesh(pendant, new THREE.SphereGeometry(0.58, 32, 24), diffuser, false);
    glow(pendant, 0, 16, 6);
    halo(pendant, 0, 3.5);
  }

  function gardenSpot(position: THREE.Vector3, target: THREE.Vector3, intensity: number) {
    const fixture = new THREE.Group();
    fixture.position.set(position.x, 0, position.z);
    root.add(fixture);
    base(fixture, 0.19, 0.08);
    const head = new THREE.Group();
    head.position.y = position.y;
    fixture.add(head);
    // Aim upward through the foliage so its silhouette lands on the rear wall.
    head.lookAt(target);
    // The emitter sits inside the housing; exclude it from its own shadow map.
    const housing = mesh(head, new THREE.CylinderGeometry(0.105, 0.12, 0.24, 20), metal, false);
    housing.rotation.x = Math.PI / 2;
    const lens = mesh(head, new THREE.CircleGeometry(0.085, 20), diffuser, false);
    lens.position.z = 0.125;
    const lensGlow = new THREE.Group();
    lensGlow.position.z = 0.15;
    head.add(lensGlow);
    halo(lensGlow, 0, 0.65);
    const beam = new THREE.SpotLight(warmWhite, intensity, 20, Math.PI / 7, 0.45, 2);
    beam.position.copy(position);
    beam.target.position.copy(target);
    beam.castShadow = true;
    beam.shadow.mapSize.set(2048, 2048);
    beam.shadow.camera.near = 0.18;
    beam.shadow.camera.far = 20;
    beam.shadow.bias = -0.00005;
    beam.shadow.normalBias = 0;
    // Static decor shadows depend on the light, never on the orbit camera.
    beam.shadow.autoUpdate = false;
    beam.shadow.needsUpdate = true;
    const length = 11;
    const hazeGeometry = new THREE.CylinderGeometry(Math.tan(beam.angle) * length, 0.035, length, 40, 1, true);
    hazeGeometry.translate(0, length / 2, 0);
    hazeGeometry.rotateX(Math.PI / 2);
    mesh(head, hazeGeometry, beamHazeMaterial, false);
    root.add(beam, beam.target);
  }

  floorCylinder(-6.1, 5.7, 0.95, 0.3);
  arcLamp();
  gardenSpot(new THREE.Vector3(-7.3, 0.24, 2.7), new THREE.Vector3(-7.8, 2, 4.5), 260);
  gardenSpot(new THREE.Vector3(5.3, 0.24, 5.6), new THREE.Vector3(4.8, 1.9, 7.4), 260);
  return root;
}
