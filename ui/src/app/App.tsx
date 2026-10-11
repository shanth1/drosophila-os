import { useEffect, useRef, useState } from 'react';
import { initialFlyState } from '../entities/fly/state';
import { connectLab } from '../features/debug-controls/channel';
import { Dashboard } from '../pages/dashboard/Dashboard';
import { Lab } from '../pages/lab/Lab';
import { Brain } from '../pages/brain/Brain';
import { useHostConnection } from '../features/host-connection/useHostConnection';
import { resolveFlyState } from '../features/presentation-rules/liveFly';

export function App() {
  const [state, setState] = useState(initialFlyState);
  const feed = useHostConnection();
  const [now, setNow] = useState(Date.now);
  const connection = useRef<ReturnType<typeof connectLab> | null>(null);
  useEffect(() => {
    connection.current = connectLab(setState, window.location.pathname === '/');
    return () => { connection.current?.close(); connection.current = null; };
  }, []);
  useEffect(() => {
    const timer = setInterval(() => setNow(Date.now()), 200);
    return () => clearInterval(timer);
  }, []);
  const fly = resolveFlyState(state, feed, now);
  const path = window.location.pathname;
  return <><header><a className="brand" href="/">DROSOPHILA<span>.OS</span></a><nav aria-label="Main navigation">
    {[['/', 'World'], ['/brain', 'Brain'], ['/lab', 'Lab']].map(([href, label]) => <a key={href} href={href} aria-current={path === href ? 'page' : undefined}>{label}</a>)}
  </nav><span className="build-label">HOST TELEMETRY / 01</span></header>
    {path === '/' ? <Dashboard state={fly} feed={feed} now={now} /> : path === '/lab' ? <Lab state={fly} supported={typeof BroadcastChannel !== 'undefined'} send={value => connection.current?.send(value)} /> : path === '/brain' ? <Brain feed={feed} /> : <main className="page"><h1>Page not found</h1><a href="/">Return to world</a></main>}
  </>;
}
