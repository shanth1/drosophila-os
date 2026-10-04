import { behaviors, initialFlyState } from '../../entities/fly/state';
import type { FlyState } from '../../entities/fly/state';
import { SoundControls } from '../../features/presentation-sound/SoundControls';

export function Lab({ state, send, supported }: { state: FlyState; send: (state: FlyState) => void; supported: boolean }) {
  return <main className="page"><span className="eyebrow">PRESENTATION LABORATORY</span><h1>Give the fly a direction.</h1>
    <p>Open the dashboard in another tab on the same origin. These controls affect graphics only, never the brain or host.</p>
    {!supported && <p role="alert">BroadcastChannel is unavailable in this browser.</p>}
    <section className="panel"><h2>Behavior preview</h2><div className="behavior-buttons">{behaviors.map(behavior =>
      <button key={behavior} aria-pressed={state.mode === 'manual' && state.behavior === behavior} onClick={() => send({ ...state, mode: 'manual', behavior })}>{behavior}</button>
    )}</div><label className="slider">Activity <strong>{Math.round(state.activity * 100)}%</strong>
      <input type="range" min="0" max="1" step="0.01" value={state.activity} onChange={event => send({ ...state, mode: 'manual', activity: Number(event.target.value) })} />
    </label><button onClick={() => send({ ...initialFlyState })}>Return to live telemetry</button>
    <p className="muted">Alarmed stays restless even at zero activity. Higher activity intensifies panic and shifts lighting from yellow to red; at 60%, the rotating beacon and enabled siren join in.</p>
    <p className="muted">Analyzing examines a decorative chart through a magnifying glass, with slow scanning and thoughtful head tilts. Activity adjusts inspection speed smoothly.</p>
    <p className="muted">Current mode: {state.mode}. Manual mode overrides host-driven animation; telemetry continues updating. Working types in fast, irregular runs with natural gaps; activity adjusts typing speed. Coffee rests a steaming mug beside the fly, gestures with the free limb, and takes occasional sips; activity gently adjusts the pace. Break turns the fly away. Full action sequences and event injection are later milestones.</p></section>
    <section className="panel"><h2>Sound</h2>
      <SoundControls state={state} />
      <p className="muted">Controls all presentation sounds: alarm siren, coffee sips, and keyboard clicks. Sound plays from this laboratory tab; keep it open with the dashboard on the same origin. Sip and typing sounds follow the actual scene animation.</p>
    </section>
    <a href="/" target="_blank" rel="noreferrer">Open dashboard ↗</a>
  </main>;
}
