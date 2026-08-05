import http from 'k6/http';
import { check, sleep } from 'k6';
import {
  BASE_URL,
  COURSE_ID,
  LESSON_ID,
  defaultHeaders,
} from './config.js';

function get(path, token, name) {
  const res = http.get(`${BASE_URL}${path}`, {
    headers: defaultHeaders(token),
    tags: { name },
  });
  check(res, {
    [`${name} status 200`]: (r) => r.status === 200,
  });
  return res;
}

/** Ilova ochilganda va asosiy navigatsiya */
export function mobileHomeFlow(token) {
  get('/mobile/banner', token, 'banner_list');
  get('/mobile/course', token, 'course_list');
  get('/mobile/subject', token, 'subject_list');
  get('/mobile/notification/user', token, 'notifications');
  get('/mobile/user/get/profile', token, 'profile');
}

/** Kurs ichida o'rganish */
export function mobileLearningFlow(token) {
  get('/mobile/course/permission', token, 'my_courses');

  if (COURSE_ID) {
    get(`/mobile/course/${COURSE_ID}`, token, 'course_detail');
  }

  if (LESSON_ID) {
    get(`/mobile/lesson/${LESSON_ID}`, token, 'lesson_detail');
  }
}

/** Reyting va activity */
export function mobileEngagementFlow(token) {
  get('/mobile/user/rating?type=total&limit=10', token, 'rating');
  get('/mobile/user/activity?type=week', token, 'activity_stats');
  get('/mobile/tariff', token, 'tariff_list');
}

/** To'liq foydalanuvchi sessiyasi (80% read trafik modeli) */
export function fullUserSession(token) {
  mobileHomeFlow(token);
  sleep(0.3);
  mobileLearningFlow(token);
  sleep(0.2);
  mobileEngagementFlow(token);
}

/** Faqat eng og'ir endpointlar */
export function heavyEndpointsOnly(token) {
  if (LESSON_ID) {
    get(`/mobile/lesson/${LESSON_ID}`, token, 'lesson_detail');
  }
  get('/mobile/user/rating?type=total&limit=10', token, 'rating');
  get('/mobile/course', token, 'course_list');
}
