import { useEffect, useRef, useState } from 'react';
import { alarmSirenStrength, alarmSirenThreshold } from '../presentation-rules/alarm';
import type { FlyState } from '../../entities/fly/state';

type PresentationAudio = { context: AudioContext; tone: OscillatorNode; modulation: OscillatorNode; depth: GainNode; filter: BiquadFilterNode; gain: GainNode; master: GainNode };

export function SoundControls({ state }: { state: FlyState }) {
  const audio = useRef<PresentationAudio | null>(null);
  const [enabled, setEnabled] = useState(false);
  const [volume, setVolume] = useState(0.2);
  const [error, setError] = useState('');
  const lastSip = useRef(0);
  const lastKeyPress = useRef(0);

  useEffect(() => {
    const siren = audio.current;
    if (!siren) return;
    const parameter = siren.gain.gain;
    parameter.cancelScheduledValues(siren.context.currentTime);
    const strength = state.behavior === 'alarmed' ? alarmSirenStrength(state.activity) : 0;
    parameter.setTargetAtTime(strength, siren.context.currentTime, 0.08);
    siren.master.gain.cancelScheduledValues(siren.context.currentTime);
    siren.master.gain.setTargetAtTime(enabled ? volume * 0.18 : 0, siren.context.currentTime, 0.04);
  }, [enabled, state.behavior, state.activity, volume]);

  useEffect(() => {
    if (typeof BroadcastChannel === 'undefined') return;
    const channel = new BroadcastChannel('drosophila.presentation-sound.v1');
    channel.onmessage = (event: MessageEvent<unknown>) => {
      const message = event.data;
      if (!enabled || !audio.current || !message || typeof message !== 'object') return;
      const signal = message as Record<string, unknown>;
      if (typeof signal.timestamp !== 'number' || !Number.isFinite(signal.timestamp)) return;
      const age = Date.now() - signal.timestamp;
      if (age < 0 || age > 500) return;
      if (signal.type === 'coffee.sip' && state.behavior === 'coffee') {
        if (signal.timestamp - lastSip.current < 1000) return;
        lastSip.current = signal.timestamp;
        playSip(audio.current);
      } else if (signal.type === 'working.keypress' && state.behavior === 'working') {
        if (signal.timestamp - lastKeyPress.current < 25) return;
        lastKeyPress.current = signal.timestamp;
        playKeyPress(audio.current);
      }
    };
    return () => channel.close();
  }, [enabled, state.behavior]);

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
    siren.master.disconnect();
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
        const master = context.createGain();
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
        master.gain.value = 0;
        modulation.connect(depth);
        depth.connect(tone.frequency);
        tone.connect(filter);
        filter.connect(gain);
        gain.connect(master);
        master.connect(context.destination);
        audio.current = { context, tone, modulation, depth, filter, gain, master };
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

  return <div className="sound-controls">
    <button type="button" aria-pressed={enabled} onClick={() => void toggle()}>{enabled ? 'Mute sound' : 'Enable sound'}</button>
    <label className="slider">Volume <strong>{Math.round(volume * 100)}%</strong>
      <input aria-label="Sound volume" type="range" min="0" max="1" step="0.01" value={volume} onChange={event => setVolume(Number(event.target.value))} />
    </label>
    <small>Siren activates at {Math.round(alarmSirenThreshold * 100)}% activity when sound is enabled.</small>
    {error && <small role="alert">{error}</small>}
  </div>;
}

function playSip({ context, master }: PresentationAudio) {
  if (context.state !== 'running') return;
  const duration = 0.42;
  const buffer = context.createBuffer(1, Math.ceil(context.sampleRate * duration), context.sampleRate);
  const samples = buffer.getChannelData(0);
  for (let index = 0; index < samples.length; index++) samples[index] = Math.random() * 2 - 1;
  const noise = context.createBufferSource();
  noise.buffer = buffer;
  const filter = context.createBiquadFilter();
  filter.type = 'bandpass';
  filter.Q.value = 2.5;
  const envelope = context.createGain();
  const gulp = context.createOscillator();
  gulp.type = 'sine';
  const gulpGain = context.createGain();
  const start = context.currentTime;
  filter.frequency.setValueAtTime(650, start);
  filter.frequency.exponentialRampToValueAtTime(1600, start + 0.2);
  filter.frequency.exponentialRampToValueAtTime(450, start + duration);
  envelope.gain.setValueAtTime(0, start);
  envelope.gain.linearRampToValueAtTime(0.65, start + 0.05);
  envelope.gain.linearRampToValueAtTime(0.3, start + 0.18);
  envelope.gain.linearRampToValueAtTime(0.5, start + 0.25);
  envelope.gain.linearRampToValueAtTime(0, start + duration);
  gulp.frequency.setValueAtTime(190, start);
  gulp.frequency.exponentialRampToValueAtTime(75, start + duration);
  gulpGain.gain.setValueAtTime(0, start);
  gulpGain.gain.linearRampToValueAtTime(0.2, start + 0.22);
  gulpGain.gain.linearRampToValueAtTime(0, start + duration);
  noise.connect(filter);
  filter.connect(envelope);
  envelope.connect(master);
  gulp.connect(gulpGain);
  gulpGain.connect(master);
  noise.onended = () => {
    noise.disconnect();
    filter.disconnect();
    envelope.disconnect();
    gulp.disconnect();
    gulpGain.disconnect();
  };
  noise.start(start);
  gulp.start(start);
  gulp.stop(start + duration);
}

function playKeyPress({ context, master }: PresentationAudio) {
  if (context.state !== 'running') return;
  const duration = 0.09;
  const buffer = context.createBuffer(1, Math.ceil(context.sampleRate * duration), context.sampleRate);
  const samples = buffer.getChannelData(0);
  const bodyFrequency = 260 + Math.random() * 140;
  const switchFrequency = 1200 + Math.random() * 700;
  const releaseTime = 0.025 + Math.random() * 0.025;
  const strength = 0.65 + Math.random() * 0.35;
  let previousNoise = 0;
  for (let index = 0; index < samples.length; index++) {
    const time = index / context.sampleRate;
    const noise = Math.random() * 2 - 1;
    const attack = 1 - Math.exp(-time / 0.0003);
    const click = (noise - previousNoise) * 0.2 * Math.exp(-time / 0.002);
    const body = Math.sin(time * bodyFrequency * Math.PI * 2) * 0.55 * Math.exp(-time / 0.012);
    const switchClick = Math.sin(time * switchFrequency * Math.PI * 2) * 0.16 * Math.exp(-time / 0.004);
    const release = time >= releaseTime ? noise * 0.17 * Math.exp(-(time - releaseTime) / 0.003) : 0;
    // Brief switch impact, damped case resonance, and a quieter key return.
    samples[index] = ((click + body + switchClick) * attack + release) * strength;
    previousNoise = noise;
  }
  const click = context.createBufferSource();
  click.buffer = buffer;
  click.connect(master);
  click.onended = () => click.disconnect();
  click.start();
}
