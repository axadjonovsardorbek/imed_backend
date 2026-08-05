/**
 * Asosiy load test — mobile read trafik (realistic).
 *
 *   BASE_URL=https://api.example.com \
 *   ACCESS_TOKEN=eyJ... \
 *   COURSE_ID=uuid LESSON_ID=uuid \
 *   k6 run k6/load-read.js
 */
import { sleep } from 'k6';
import { resolveToken } from './lib/auth.js';
import { loadThresholds } from './lib/config.js';
import { fullUserSession } from './lib/scenarios.js';

export const options = {
  scenarios: {
    ramp_up: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 30 },
        { duration: '3m', target: 30 },
        { duration: '1m', target: 80 },
        { duration: '3m', target: 80 },
        { duration: '1m', target: 0 },
      ],
      gracefulRampDown: '30s',
    },
  },
  thresholds: loadThresholds,
};

export function setup() {
  const token = resolveToken();
  if (!token) {
    throw new Error('Token olinmadi');
  }
  return { token };
}

export default function (data) {
  fullUserSession(data.token);
  sleep(Math.random() * 2 + 0.5);
}
