// Umumiy sozlamalar — muhit o'zgaruvchilari orqali boshqariladi.
// Local: BASE_URL=http://localhost:8004
// Prod:  BASE_URL=https://prod.imedteam.uz ALLOW_PROD=1

export const PROD_BASE_URL = 'https://prod.imedteam.uz';

export const BASE_URL = __ENV.BASE_URL || 'http://localhost:8004';

export const ACCESS_TOKEN = __ENV.ACCESS_TOKEN || '';
export const USER_PHONE = __ENV.USER_PHONE || '';
export const USER_PASSWORD = __ENV.USER_PASSWORD || '';

// Test ma'lumotlari (staging/prod dan haqiqiy UUID lar)
export const COURSE_ID = __ENV.COURSE_ID || '';
export const LESSON_ID = __ENV.LESSON_ID || '';

export const LANG = __ENV.LANG || 'uz';

export const defaultHeaders = (token) => ({
  'Content-Type': 'application/json',
  Accept: 'application/json',
  'Accept-Language': LANG,
  ...(token ? { Authorization: `Bearer ${token}` } : {}),
});

export const smokeThresholds = {
  http_req_failed: ['rate<0.01'],
  http_req_duration: ['p(95)<2000'],
};

export const loadThresholds = {
  http_req_failed: ['rate<0.05'],
  http_req_duration: ['p(95)<3000', 'p(99)<8000'],
  checks: ['rate>0.95'],
};

export const spikeThresholds = {
  http_req_failed: ['rate<0.15'],
  http_req_duration: ['p(95)<10000'],
};
