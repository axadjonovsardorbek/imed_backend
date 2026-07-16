/**
 * Production smoke test — minimal yuk, faqat read.
 * Deploy yoki monitoring uchun. ~1 min.
 *
 *   ALLOW_PROD=1 \
 *   BASE_URL=https://prod.imedteam.uz \
 *   ACCESS_TOKEN=eyJ... \
 *   COURSE_ID=... LESSON_ID=... \
 *   k6 run k6/prod-smoke.js
 */
import { sleep } from 'k6';
import { resolveToken } from './lib/auth.js';
import { smokeThresholds } from './lib/config.js';
import { assertProdAllowed } from './lib/prod-guard.js';
import { fullUserSession } from './lib/scenarios.js';

export const options = {
  vus: 1,
  iterations: 3,
  thresholds: {
    ...smokeThresholds,
    http_req_failed: ['rate<0.02'],
  },
  tags: { env: 'prod', test: 'smoke' },
};

export function setup() {
  assertProdAllowed();
  const token = resolveToken();
  if (!token) {
    throw new Error('ACCESS_TOKEN yoki USER_PHONE+USER_PASSWORD kerak');
  }
  return { token };
}

export default function (data) {
  fullUserSession(data.token);
  sleep(2);
}
