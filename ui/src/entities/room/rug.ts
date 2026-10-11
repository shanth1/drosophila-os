import * as THREE from 'three'

export function createRug() {
  const outline = new THREE.CatmullRomCurve3(
    [
      new THREE.Vector3(-3.9, -0.6, 0),
      new THREE.Vector3(-3.3, -1.9, 0),
      new THREE.Vector3(-1.8, -2.6, 0),
      new THREE.Vector3(-0.4, -2.1, 0),
      new THREE.Vector3(1.2, -2.2, 0),
      new THREE.Vector3(3.5, -1.2, 0),
      new THREE.Vector3(3.8, 0.2, 0),
      new THREE.Vector3(2.8, 1.3, 0),
      new THREE.Vector3(1.5, 2.7, 0),
      new THREE.Vector3(-0.1, 2.9, 0),
      new THREE.Vector3(-0.9, 1.9, 0),
      new THREE.Vector3(-2.9, 1.6, 0),
    ],
    true,
    'centripetal'
  )
  const shape = new THREE.Shape(
    outline.getPoints(160).map(point => new THREE.Vector2(point.x, point.y))
  )

  // Dense staggered strands provide both wool color variation and raised pile.
  const size = 256
  const fibers = new Uint8Array(size * size * 4)
  let seed = 73
  for (let index = 0; index < size * size; index++) {
    seed = (Math.imul(seed, 1664525) + 1013904223) >>> 0
    const x = index % size
    const y = Math.floor(index / size)
    const strandRow = (y + (x % 3) * 2) % 8
    const strand = Math.sin((Math.PI * strandRow) / 8)
    const ridge = x % 3 === 1 ? 62 * strand : -32 * strand
    const tuft = 18 * Math.sin((x * Math.PI) / 16) * Math.sin((y * Math.PI) / 16)
    const value = Math.round(168 + ridge + tuft + (seed >>> 28))
    fibers.set([value, value, value, 255], index * 4)
  }
  const pile = new THREE.DataTexture(fibers, size, size)
  pile.wrapS = pile.wrapT = THREE.RepeatWrapping
  pile.repeat.set(1.5, 1.5)
  pile.magFilter = THREE.LinearFilter
  pile.minFilter = THREE.LinearMipmapLinearFilter
  pile.generateMipmaps = true
  pile.needsUpdate = true

  const geometry = new THREE.ExtrudeGeometry(shape, {
    depth: 0.018,
    bevelEnabled: true,
    bevelThickness: 0.006,
    bevelSize: 0.035,
    bevelSegments: 3,
    steps: 1,
  })
  const binding = new THREE.Mesh(
    geometry,
    new THREE.MeshStandardMaterial({
      color: '#493d30',
      roughness: 1,
    })
  )
  const surface = new THREE.Mesh(
    geometry,
    new THREE.MeshStandardMaterial({
      color: '#80664d',
      roughness: 1,
      map: pile,
      bumpMap: pile,
      bumpScale: 0.09,
    })
  )
  surface.scale.set(0.975, 0.975, 1)
  surface.position.z = 0.006
  binding.receiveShadow = true
  surface.receiveShadow = true
  const rug = new THREE.Group()
  rug.add(binding, surface)
  rug.scale.set(1.5, 1.5, 1)
  rug.rotation.set(-Math.PI / 2, 0, 0)
  rug.position.set(1, -0.019, 8.2)
  return rug
}
