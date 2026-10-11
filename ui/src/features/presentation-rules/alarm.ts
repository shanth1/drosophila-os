export const alarmSirenThreshold = 0.6;

export function alarmSirenStrength(activity: number): number {
  if (activity < alarmSirenThreshold) return 0;
  return 0.35 + 0.65 * (activity - alarmSirenThreshold) / (1 - alarmSirenThreshold);
}
