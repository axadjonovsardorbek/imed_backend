import http from 'k6/http';
import { check } from 'k6';
import { BASE_URL, USER_PHONE, USER_PASSWORD, defaultHeaders } from './config.js';

/**
 * setup() da bir marta chaqiriladi.
 * ACCESS_TOKEN berilgan bo'lsa login qilmaydi.
 */
export function resolveToken() {
  if (__ENV.ACCESS_TOKEN) {
    return __ENV.ACCESS_TOKEN;
  }

  if (!USER_PHONE || !USER_PASSWORD) {
    console.warn(
      'ACCESS_TOKEN yoki USER_PHONE+USER_PASSWORD kerak. ' +
        'Masalan: ACCESS_TOKEN=eyJ... k6 run k6/smoke.js'
    );
    return '';
  }

  const res = http.post(
    `${BASE_URL}/mobile/auth/user/login`,
    JSON.stringify({
      phone_number: USER_PHONE,
      password: USER_PASSWORD,
    }),
    { headers: defaultHeaders() }
  );

  const ok = check(res, {
    'login status 200': (r) => r.status === 200,
  });

  if (!ok) {
    console.error(`login failed: ${res.status} ${res.body}`);
    return '';
  }

  const body = res.json();
  return body.access_token || '';
}
