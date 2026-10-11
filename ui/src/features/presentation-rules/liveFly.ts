import type { FlyState } from '../../entities/fly/state.ts';
import type { HostFeed } from '../../shared/api/client.ts';

// Artistic choices live here. Neither HTTP failure nor a spike is relabeled as
// an anomaly diagnosis. The alert gesture reflects a freshly received output.
export function resolveFlyState(manual: FlyState, feed: HostFeed, now: number): FlyState {
  if (manual.mode === 'manual') return manual;
  const state = feed.state;
  if (feed.connection !== 'connected' || !state || state.status !== 'running') return { mode: 'live', behavior: 'idle', activity: 0 };
  const activity = state.brain.neurons ? Math.min(1, state.brain.spikes / state.brain.neurons) : 0;
  const recentOutput = feed.lastOutputReceivedAt !== null && now - feed.lastOutputReceivedAt >= 0 && now - feed.lastOutputReceivedAt < 1500;
  return { mode: 'live', behavior: recentOutput ? 'alarmed' : activity > 0.05 ? 'working' : 'idle', activity: recentOutput ? Math.max(0.8, activity) : Math.max(0.05, activity) };
}
