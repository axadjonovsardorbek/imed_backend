/**
 * Production test xavfsizligi — ALLOW_PROD=1 bo'lmasa test to'xtaydi.
 */
export function assertProdAllowed() {
  if (__ENV.ALLOW_PROD !== '1') {
    throw new Error(
      'Production test uchun ALLOW_PROD=1 kerak. ' +
        'Masalan: ALLOW_PROD=1 BASE_URL=https://prod.imedteam.uz k6 run k6/prod-smoke.js'
    );
  }

  const url = (__ENV.BASE_URL || '').toLowerCase();
  if (!url || url.includes('localhost') || url.includes('127.0.0.1')) {
    throw new Error(
      'prod-smoke/prod-load uchun haqiqiy prod BASE_URL kerak (localhost emas). ' +
        'Masalan: BASE_URL=https://prod.imedteam.uz'
    );
  }
}
