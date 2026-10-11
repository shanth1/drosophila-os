import { useEffect, useState } from 'react';
import { connectHost, initialHostFeed } from '../../shared/api/client';

export function useHostConnection() {
  const [feed, setFeed] = useState(initialHostFeed);
  useEffect(() => connectHost(setFeed), []);
  return feed;
}
