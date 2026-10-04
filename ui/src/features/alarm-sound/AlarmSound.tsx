import { useEffect, useRef, useState } from 'react';
import { alarmSirenStrength, alarmSirenThreshold } from '../presentation-rules/alarm';

type Siren = { context: AudioContext; tone: OscillatorNode; modulation: OscillatorNode; depth: GainNode; filter: BiquadFilterNode; gain: GainNode };

export function AlarmSound({ alarmed, activity }: { alarmed: boolean; activity: number }) {
  const audio = useRef<Siren | null>(null);
  const [enabled, setEnabled] = useState(false);
  const [volume, setVolume] = useState(0.2);
  const [error, setError] = useState('');

  useEffect(() => {
    const siren = audio.current;
    if (!siren) return;
    const parameter = siren.gain.gain;
    parameter.cancelScheduledValues(siren.context.currentTime);
    const strength = alarmed ? alarmSirenStrength(activity) : 0;
    parameter.setTargetAtTime(enabled ? volume * 0.18 * strength : 0, siren.context.currentTime, 0.08);
  }, [enabled, alarmed, activity, volume]);

  useEffect(() => () => {
    const siren = audio.current;
    if (!siren) return;
    siren.tone.stop();
    siren.modulation.stop();
    siren.tone.disconnect();
    siren.modulation.disconnect();
    siren.depth.disconnect();
    siren.filter.disconnect();
    siren.gain.disconnect();
    void siren.context.close();
    audio.current = null;
  }, []);

  async function toggle() {
    if (enabled) { setEnabled(false); return; }
    try {
      if (!audio.current) {
        const context = new AudioContext();
        const tone = context.createOscillator();
        const modulation = context.createOscillator();
        const depth = context.createGain();
        const gain = context.createGain();
        const filter = context.createBiquadFilter();
        // Alternating 380/660 Hz with filtered harmonics gives a two-tone alarm.
        tone.type = 'sawtooth';
        tone.frequency.value = 520;
        modulation.type = 'square';
        modulation.frequency.value = 1.25;
        depth.gain.value = 140;
        filter.type = 'lowpass';
        filter.frequency.value = 1800;
        filter.Q.value = 0.5;
        gain.gain.value = 0;
        modulation.connect(depth);
        depth.connect(tone.frequency);
        tone.connect(filter);
        filter.connect(gain);
        gain.connect(context.destination);
        audio.current = { context, tone, modulation, depth, filter, gain };
        tone.start();
        modulation.start();
      }
      const siren = audio.current;
      await siren.context.resume();
      if (audio.current !== siren) return;
      setError('');
      setEnabled(true);
    } catch {
      setError('Audio is unavailable. Try enabling sound again.');
    }
  }

  return <div className="alarm-sound">
    <button type="button" aria-pressed={enabled} onClick={() => void toggle()}>{enabled ? 'Mute siren' : 'Enable siren sound'}</button>
    <label className="slider">Siren volume <strong>{Math.round(volume * 100)}%</strong>
      <input aria-label="Siren volume" type="range" min="0" max="1" step="0.01" value={volume} onChange={event => setVolume(Number(event.target.value))} />
    </label>
    <small>Siren activates at {Math.round(alarmSirenThreshold * 100)}% activity when sound is enabled.</small>
    {error && <small role="alert">{error}</small>}
  </div>;
}
