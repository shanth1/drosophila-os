import { initialFlyState, isFlyState } from '../../entities/fly/state';
import type { FlyState } from '../../entities/fly/state';

// This channel carries presentation controls only, never host commands.
export function connectLab(onState: (state: FlyState) => void, respondsToRequests: boolean) {
  let current = { ...initialFlyState };
  const channel = typeof BroadcastChannel !== 'undefined'
    ? new BroadcastChannel('drosophila.presentation.v1') : null;
  channel?.addEventListener('message', (event: MessageEvent<unknown>) => {
    const message = event.data;
    if (!message || typeof message !== 'object') return;
    const envelope = message as Record<string, unknown>;
    if (envelope.type === 'request' && respondsToRequests) {
      channel?.postMessage({ type: 'state', state: current });
    } else if (envelope.type === 'state' && isFlyState(envelope.state)) {
      current = envelope.state;
      onState(current);
    }
  });
  channel?.postMessage({ type: 'request' });
  return {
    supported: channel !== null,
    send(state: FlyState) {
      current = state;
      onState(state);
      channel?.postMessage({ type: 'state', state });
    },
    close() { channel?.close(); },
  };
}
