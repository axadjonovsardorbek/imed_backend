/**
 * Production load test — ehtiyotkor, faqat GET read endpointlar.
 * Kechki/soat past vaqtda ishga tushiring. Max ~15 parallel user.
 *
 *   ALLOW_PROD=1 \
 *   BASE_URL=https://prod.imedteam.uz \
 *   ACCESS_TOKEN=eyJ... \
 *   COURSE_ID=... LESSON_ID=... \
 *   k6 run k6/prod-load.js
 */
import { sleep } from 'k6';
import { resolveToken } from './lib/auth.js';
import { loadThresholds } from './lib/config.js';
import { assertProdAllowed } from './lib/prod-guard.js';
import { fullUserSession } from './lib/scenarios.js';

const maxVUs = parseInt(__ENV.PROD_MAX_VUS || '15', 10);

export const options = {
  scenarios: {
    prod_read: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: Math.min(5, maxVUs) },
        { duration: '3m', target: maxVUs },
        { duration: '1m', target: 0 },
      ],
      gracefulRampDown: '30s',
    },
  },
  thresholds: {
    ...loadThresholds,
    http_req_failed: ['rate<0.03'],
    http_req_duration: ['p(95)<5000'],
  },
  tags: { env: 'prod', test: 'load' },
};

export function setup() {
  assertProdAllowed();
  const token = resolveToken();
  if (!token) {
    throw new Error('ACCESS_TOKEN kerak — prod da login stress qilmang');
  }
  console.log(`Prod load: BASE_URL=${__ENV.BASE_URL}, max VUs=${maxVUs}`);
  return { token };
}

export default function (data) {
  fullUserSession(data.token);
  sleep(Math.random() * 3 + 1);
}
