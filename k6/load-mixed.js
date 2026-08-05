/**
 * Aralash trafik: 85% read, 15% write (activity + notification read).
 *
 *   k6 run k6/load-mixed.js
 */
import http from 'k6/http';
import { check, sleep } from 'k6';
import { resolveToken } from './lib/auth.js';
import { BASE_URL, defaultHeaders, loadThresholds } from './lib/config.js';
import { fullUserSession } from './lib/scenarios.js';

export const options = {
  scenarios: {
    mixed: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2m', target: 40 },
        { duration: '5m', target: 40 },
        { duration: '1m', target: 0 },
      ],
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

function postActivity(token) {
  const res = http.post(
    `${BASE_URL}/mobile/user/activity`,
    JSON.stringify({ activity: 1 }),
    { headers: defaultHeaders(token), tags: { name: 'activity_write' } }
  );
  check(res, {
    'activity status 200': (r) => r.status === 200,
  });
}

export default function (data) {
  const roll = Math.random();

  if (roll < 0.85) {
    fullUserSession(data.token);
  } else if (roll < 0.95) {
    postActivity(data.token);
  } else {
    const res = http.get(`${BASE_URL}/mobile/user/rating?type=day&limit=20`, {
      headers: defaultHeaders(data.token),
      tags: { name: 'rating_heavy' },
    });
    check(res, { 'rating status 200': (r) => r.status === 200 });
  }

  sleep(Math.random() * 1.5 + 0.3);
}
