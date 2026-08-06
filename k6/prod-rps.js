/**
 * Production RPS / capacity check — oldingi 15 VU dan kuchliroq,
 * lekin 500–1500 VU spike EMAS (prod xavfsizligi).
 *
 * Faqat GET. Default max 80 VU. Oshirish: PROD_MAX_VUS=100
 *
 *   ALLOW_PROD=1 \
 *   BASE_URL=https://prod.imedteam.uz \
 *   ACCESS_TOKEN=eyJ... \
 *   COURSE_ID=... LESSON_ID=... \
 *   PROD_MAX_VUS=80 \
 *   k6 run k6/prod-rps.js
 */
import { sleep } from 'k6';
import { resolveToken } from './lib/auth.js';
import { assertProdAllowed } from './lib/prod-guard.js';
import { mobileHomeFlow, mobileLearningFlow } from './lib/scenarios.js';

const maxVUs = Math.min(parseInt(__ENV.PROD_MAX_VUS || '80', 10), 150);
const mid = Math.max(Math.floor(maxVUs / 2), 10);

export const options = {
  scenarios: {
    prod_rps: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '20s', target: mid },
        { duration: '40s', target: maxVUs },
        { duration: '20s', target: 0 },
      ],
      gracefulRampDown: '15s',
    },
  },
  thresholds: {
    // Prod spike da 1% qattiq; realistik chegaralar
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<5000'],
    checks: ['rate>0.90'],
  },
  tags: { env: 'prod', test: 'rps' },
};

export function setup() {
  assertProdAllowed();
  const token = resolveToken();
  if (!token) {
    throw new Error('ACCESS_TOKEN kerak');
  }
  console.log(`Prod RPS test: BASE_URL=${__ENV.BASE_URL}, maxVUs=${maxVUs}`);
  return { token };
}

export default function (data) {
  // Yengilroq flow → RPS aniqroq chiqadi (to'liq 11 endpoint emas)
  mobileHomeFlow(data.token);
  mobileLearningFlow(data.token);
  sleep(0.15);
}
