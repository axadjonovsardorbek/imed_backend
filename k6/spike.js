/**
 * Spike test — qisqa vaqt ichida kutilmagan trafik.
 * Masalan: push notification yuborilganda hamma ilovani ochadi.
 *
 *   k6 run k6/spike.js
 */
import { sleep } from 'k6';
import { resolveToken } from './lib/auth.js';
import { spikeThresholds } from './lib/config.js';
import { mobileHomeFlow } from './lib/scenarios.js';

export const options = {
  scenarios: {
    spike: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 20 },
        { duration: '1m', target: 200 },
        { duration: '30s', target: 20 },
        { duration: '1m', target: 0 },
      },
      gracefulRampDown: '20s',
    },
  },
  thresholds: spikeThresholds,
};

export function setup() {
  const token = resolveToken();
  if (!token) {
    throw new Error('Token olinmadi');
  }
  return { token };
}

export default function (data) {
  mobileHomeFlow(data.token);
  sleep(0.1);
}
