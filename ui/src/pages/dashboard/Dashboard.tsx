import type { FlyState } from '../../entities/fly/state';
import { WorldScene } from '../../widgets/world-scene/WorldScene';

export function Dashboard({ state }: { state: FlyState }) {
  return <main className="dashboard">
    <WorldScene state={state} />
    <section className="scene-caption"><span className="eyebrow">ONE SMALL FLY. A WORLD TO WATCH.</span>
      <h1>Your resident operator.</h1><p>Original procedural model · Drag to orbit · Scroll to zoom</p>
    </section>
    <aside className="status-card"><span className="status-dot" />{state.mode === 'manual' ? 'MANUAL PRESENTATION' : 'VISUAL PREVIEW'}
      <h2>{state.behavior}</h2><p>Activity {Math.round(state.activity * 100)}%</p><small>No backend telemetry connected.</small>
      <a href="/lab" target="_blank" rel="noreferrer">Open laboratory ↗</a>
    </aside>
  </main>;
}
