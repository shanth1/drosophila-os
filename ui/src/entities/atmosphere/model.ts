import * as THREE from 'three';

export function createAtmosphericHaze() {
  const root = new THREE.Group();
  const geometry = new THREE.PlaneGeometry(1, 1);
  const material = new THREE.ShaderMaterial({
    uniforms: { color: { value: new THREE.Color('#b4a28a') } },
    vertexShader: `
      varying vec2 hazeUV;
      varying vec3 hazeOrigin;
      void main() {
        hazeUV = uv;
        hazeOrigin = modelMatrix[3].xyz;
        vec4 center = modelViewMatrix * vec4(0.0, 0.0, 0.0, 1.0);
        vec2 size = vec2(length(modelMatrix[0].xyz), length(modelMatrix[1].xyz));
        center.xy += position.xy * size;
        gl_Position = projectionMatrix * center;
      }
    `,
    fragmentShader: `
      uniform vec3 color;
      varying vec2 hazeUV;
      varying vec3 hazeOrigin;
      void main() {
        vec2 p = (hazeUV - 0.5) * 2.0;
        float falloff = pow(max(0.0, 1.0 - dot(p, p)), 2.0);
        float wisps = 0.7 + 0.15 * sin(p.x * 5.0 + p.y * 3.0 + hazeOrigin.x)
          + 0.15 * sin(p.y * 8.0 - p.x * 2.0 + hazeOrigin.z);
        gl_FragColor = vec4(color, falloff * wisps * 0.065);
        #include <colorspace_fragment>
      }
    `,
    transparent: true,
    depthWrite: false,
    toneMapped: false,
  });
  // Separate world-space layers preserve parallax and normal depth occlusion.
  const layers = [
    { position: [-2.2, 1.1, 1.8], size: [6, 2.4] },
    { position: [-5.5, 1.7, 4.5], size: [6, 3.4] },
    { position: [4.8, 2.3, 5.7], size: [5, 4] },
    { position: [-0.8, 2.1, 8.2], size: [9, 4.2] },
  ];
  for (const { position, size } of layers) {
    const layer = new THREE.Mesh(geometry, material);
    layer.position.set(position[0], position[1], position[2]);
    layer.scale.set(size[0], size[1], 1);
    layer.frustumCulled = false;
    root.add(layer);
  }
  return root;
}
