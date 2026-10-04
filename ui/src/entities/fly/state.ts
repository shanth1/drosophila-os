export const behaviors = ['idle', 'working', 'alarmed', 'coffee', 'break'] as const;
export type Behavior = typeof behaviors[number];
export type FlyState = { mode: 'live' | 'manual'; behavior: Behavior; activity: number };
export const initialFlyState: FlyState = { mode: 'live', behavior: 'idle', activity: 0.15 };

export function isFlyState(value: unknown): value is FlyState {
  if (!value || typeof value !== 'object') return false;
  const state = value as Record<string, unknown>;
  return (state.mode === 'live' || state.mode === 'manual') &&
    behaviors.includes(state.behavior as Behavior) && typeof state.activity === 'number' &&
    Number.isFinite(state.activity) && state.activity >= 0 && state.activity <= 1;
}
