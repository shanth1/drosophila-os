import type { HostFeed } from '../../shared/api/client';

export function Brain({ feed }: { feed: HostFeed }) {
  const brain = feed.state?.brain;
  return <main className="page"><span className="eyebrow">NEURAL OBSERVATORY</span><h1>Activity, not anatomy.</h1>
    <p>The connectome uses CSR to store connections. It does not contain anatomical coordinates.</p>
    <section className="panel"><h2>Host: {feed.connection}</h2>
      {brain ? <dl className="telemetry-values"><dt>Neurons</dt><dd>{brain.neurons}</dd><dt>Synapses</dt><dd>{brain.synapses}</dd><dt>Tick</dt><dd>{brain.tick}</dd><dt>Spikes on sampled tick</dt><dd>{brain.spikes}</dd><dt>Total spikes</dt><dd>{brain.totalSpikes}</dd><dt>Output events</dt><dd>{brain.outputEvents}</dd></dl> : <p>Waiting for a host snapshot.</p>}
      {feed.error && <p>{feed.error}</p>}
      <p className="muted">{feed.connection !== 'connected' ? 'Displayed values are last known, not live. ' : ''}Counts come from the real engine. Raster plots, selected voltages, and local connectivity belong to later milestones.</p></section>
    {feed.state && <section className="panel"><h2>Modules</h2><ul>{feed.state.modules.map(module => <li key={module.id}>{module.id} · {module.kind} · {module.status}<p className="muted">{module.capabilities.join(', ')}</p></li>)}</ul></section>}
  </main>;
}
