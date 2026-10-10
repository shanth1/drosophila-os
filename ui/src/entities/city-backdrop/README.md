# City backdrop

`createCityBackdrop()` owns static skyline geometry, its sky, and an independent
`airspace` group for future moving objects. Room openings and indoor lighting stay
in their own entities; the world widget composes them.

- `setTheme(theme)` changes the art palette without rebuilding geometry.
- `setState(state)` applies transient sky brightness and window glow.
- A future day/night presentation controller should own its clock and map time
  to a theme and render state. The entity must not subscribe to host telemetry or
  manage React state. The current scene uses a fixed evening appearance.
- Future aircraft should own their geometry and animation in a separate entity;
  the world widget can attach their roots to `airspace` and update them in its
  existing animation loop.
- Static facade details use instancing. Geometry and materials participate in
  the world scene's existing disposal traversal. New textures or render targets
  require explicit cleanup when introduced.
