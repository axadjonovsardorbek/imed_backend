/**
 * Smoke test — deploy dan keyin tez tekshirish.
 *
 * Ishga tushirish:
 *   ACCESS_TOKEN=eyJ... BASE_URL=http://localhost:8004 k6 run k6/smoke.js
 *
 * yoki login bilan:
 *   BASE_URL=http://localhost:8004 USER_PHONE=998901234567 USER_PASSWORD=secret k6 run k6/smoke.js
 */
import { sleep } from 'k6';
import { resolveToken } from './lib/auth.js';
import { smokeThresholds } from './lib/config.js';
import { fullUserSession } from './lib/scenarios.js';

export const options = {
  vus: 1,
  duration: '30s',
  thresholds: smokeThresholds,
};

export function setup() {
  const token = resolveToken();
  if (!token) {
    throw new Error('Token olinmadi — ACCESS_TOKEN yoki login ma\'lumotlarini tekshiring');
  }
  return { token };
}

export default function (data) {
  fullUserSession(data.token);
  sleep(1);
}
