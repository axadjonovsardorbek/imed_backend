/**
 * Og'ir endpointlarga fokus — lesson detail + rating + course list.
 * DB va cache limitini aniqlash uchun.
 *
 *   k6 run k6/load-heavy.js
 */
import { sleep } from 'k6';
import { resolveToken } from './lib/auth.js';
import { loadThresholds } from './lib/config.js';
import { heavyEndpointsOnly } from './lib/scenarios.js';

export const options = {
  scenarios: {
    heavy: {
      executor: 'constant-vus',
      vus: 50,
      duration: '5m',
    },
  },
  thresholds: {
    ...loadThresholds,
    http_req_duration: ['p(95)<5000', 'p(99)<12000'],
  },
};

export function setup() {
  const token = resolveToken();
  if (!token) {
    throw new Error('Token olinmadi');
  }
  if (!__ENV.LESSON_ID) {
    console.warn('LESSON_ID berilmagan — lesson_detail test o\'tkazib yuboriladi');
  }
  return { token };
}

export default function (data) {
  heavyEndpointsOnly(data.token);
  sleep(0.2);
}
