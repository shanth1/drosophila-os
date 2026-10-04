import { behaviors, initialFlyState } from '../../entities/fly/state';
import type { FlyState } from '../../entities/fly/state';

export function Lab({ state, send, supported }: { state: FlyState; send: (state: FlyState) => void; supported: boolean }) {
  return <main className="page"><span className="eyebrow">PRESENTATION LABORATORY</span><h1>Give the fly a direction.</h1>
    <p>Open the dashboard in another tab on the same origin. These controls affect graphics only, never the brain or host.</p>
    {!supported && <p role="alert">BroadcastChannel is unavailable in this browser.</p>}
    <section className="panel"><h2>Behavior preview</h2><div className="behavior-buttons">{behaviors.map(behavior =>
      <button key={behavior} aria-pressed={state.behavior === behavior} onClick={() => send({ ...state, mode: 'manual', behavior })}>{behavior}</button>
    )}</div><label className="slider">Activity <strong>{Math.round(state.activity * 100)}%</strong>
      <input type="range" min="0" max="1" step="0.01" value={state.activity} onChange={event => send({ ...state, mode: 'manual', activity: Number(event.target.value) })} />
    </label><button onClick={() => send({ ...initialFlyState })}>Reset to live preview</button>
    <p className="muted">Current mode: {state.mode}. Coffee shows a cup; break turns the fly away. Full action sequences and event injection are later milestones.</p></section>
    <a href="/" target="_blank" rel="noreferrer">Open dashboard ↗</a>
  </main>;
}
