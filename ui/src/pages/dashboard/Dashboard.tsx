import type { FlyState } from '../../entities/fly/state';
import { WorldScene } from '../../widgets/world-scene/WorldScene';
import type { HostFeed } from '../../shared/api/client';
import { parseHTTPResponse } from '../../shared/api/protocol';

export function Dashboard({ state, feed, now }: { state: FlyState; feed: HostFeed; now: number }) {
  const observation = feed.state?.observations.find(item => item.name === 'http.response' && item.schemaVersion === 1);
  const response = observation ? parseHTTPResponse(observation.data) : null;
  const observationStale = observation ? now - Date.parse(observation.observedAt) > 3000 : true;
  const connected = feed.connection === 'connected';
  return <main className="dashboard">
    <WorldScene state={state} />
    <section className="scene-caption"><span className="eyebrow">ONE SMALL FLY. A WORLD TO WATCH.</span>
      <h1>Your resident operator.</h1><p>Original procedural model · Drag to orbit · Scroll to zoom</p>
    </section>
    <aside className="status-card"><span className={`status-dot ${connected ? '' : 'offline'}`} />{state.mode === 'manual' ? 'MANUAL PRESENTATION' : 'LIVE PRESENTATION'}
      <h2>{state.behavior}</h2><p>Visual activity {Math.round(state.activity * 100)}%</p>
      <small aria-live="polite">Host: {feed.connection}{feed.state ? ` · ${connected ? '' : 'last known '}${feed.state.mode} / ${feed.state.status}` : ''}</small>
      {feed.error && <small className="connection-error">{feed.error}</small>}
      {!connected && feed.state && <small>Counters below are last known.</small>}
      {feed.state && <dl className="telemetry-values"><dt>Tick</dt><dd>{feed.state.brain.tick}</dd><dt>Spikes / tick</dt><dd>{feed.state.brain.spikes}</dd><dt>Output events</dt><dd>{feed.state.brain.outputEvents}</dd></dl>}
      {response && <div className="observation"><small>{!connected || observationStale ? 'LAST KNOWN HTTP RESPONSE' : 'OBSERVED HTTP RESPONSE'}</small>
        <p>HTTP {response.statusCode || 'unavailable'} · {response.elapsedMs} ms</p>
        {response.error && <small className="connection-error">{response.error}</small>}
      </div>}
      <small>Alert gesture = output spike, not an anomaly diagnosis.</small>
      <a href="/lab" target="_blank" rel="noreferrer">Open laboratory ↗</a>
    </aside>
  </main>;
}
