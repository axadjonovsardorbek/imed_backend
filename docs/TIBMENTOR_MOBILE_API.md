# TibMentor: mobil ilova uchun API o'zgarishlari

Bu hujjat TibMentor yangilanishida (TZ T-01…T-17) backendda nima o'zgarganini va mobil ilova qaysi ekranda nima chaqirishini tushuntiradi.

- Barcha so'rovlar `Authorization: Bearer <access_token>` bilan yuboriladi.
- Til `Accept-Language` sarlavhasi orqali beriladi: `uz` (standart), `ru`, `en`.
- To'liq sxema Swagger'da: `/api/swagger/index.html`.

## Qisqacha: nima o'zgardi

| TZ | Ekran | Backend |
|----|-------|---------|
| T-05 | — | 24 soat kirmagan foydalanuvchiga avtomatik push keladi (`type: inactive_reminder`) |
| T-06 | Kurs, dars | Faqat kontenti bor dars va modullar qaytadi, darsda `content_types` bor, `audios` va `quizzes` qo'shildi |
| T-07…T-10 | Test | Ilova ichidagi test: start, answer, finish |
| T-11, T-12 | Premium | Bitta oylik Premium, Click orqali to'lov |
| T-14 | Reyting | `by=test` parametri qo'shildi, `me` doim keladi |
| T-15 | Profil | `is_premium`, `premium_until`, `is_ustoz`, test tarixi |
| T-17 | — | Kontent qo'shilganda avtomatik push kelmaydi, faqat admin yuboradi |
| — | Dars, test | Yopiq darsga kirilsa `403` qaytadi |

Yangi endpointlar avtomatik ravishda foydalanuvchining "oxirgi faollik" vaqtini yangilaydi. Ilova alohida hech narsa yuborishi shart emas.

---

## 1. Kurs sahifasi (T-06)

`GET /mobile/course/:id` avvalgidek ishlaydi, lekin endi:

- Kontenti yo'q darslar va darssiz modullar **qaytmaydi**. Bo'sh bo'lim ko'rsatmang.
- Har bir darsda `content_types` bor: shu darsda nima borligi.
- `has_access` endi `true` bo'ladi, agar foydalanuvchi kursni sotib olgan **yoki** premium faol **yoki** ustoz bo'lsa.
- `tests` soni endi ilova ichidagi testlar soni (eski PDF testlar hisoblanmaydi).

```json
{
  "has_access": true,
  "tests": 3,
  "modules": [
    {
      "id": "…",
      "name": "1-modul",
      "lessons": [
        {
          "id": "…",
          "name": "Yurak anatomiyasi",
          "is_public": true,
          "is_completed": false,
          "order_num": 1,
          "content_types": ["video", "document", "test"]
        }
      ]
    }
  ]
}
```

`content_types` qiymatlari: `video`, `audio`, `document`, `test`. Faqat shu ro'yxatdagi bo'limlarni (tab/ikonka) ko'rsating.

## 2. Dars sahifasi (T-06)

`GET /mobile/lesson/:id`

```json
{
  "id": "…",
  "name": "Yurak anatomiyasi",
  "is_completed": false,
  "content_types": ["video", "test"],
  "videos":    [{ "id": "…", "name": "…", "url": "…", "video_url": "…", "type": "video", "order_num": 1 }],
  "audios":    [],
  "documents": [],
  "quizzes": [
    { "id": "test-uuid", "title": "1-test", "time_limit_sec": 600, "question_count": 20, "best_percent": 85.5 }
  ],
  "tests": []
}
```

- `audios`: yangi. Format `videos`/`documents` bilan bir xil. Hozircha admin audio yuklamagan, ro'yxat bo'sh keladi.
- `quizzes`: ilova ichidagi testlar. `best_percent` foydalanuvchining shu testdagi eng yaxshi natijasi (hali ishlamagan bo'lsa `null`).
- `tests`: eski PDF testlar uchun edi. Endi **doim bo'sh**, eski ilova versiyalari buzilmasligi uchun qoldirilgan. Yangi ilovada ishlatmang.

**Ruxsat yo'q bo'lsa:** `403 {"message": "no access to this lesson"}`. Bu holatda Premium sotib olish oynasini ko'rsating (5-bo'lim).

## 3. Test ishlash (T-07…T-10)

### 3.1 Testni olish

`GET /mobile/test/:id`

```json
{
  "id": "test-uuid",
  "lesson_id": "…",
  "title": "1-test",
  "time_limit_sec": 600,
  "best_percent": null,
  "questions": [
    {
      "id": "question-uuid",
      "text": "Yurak nechta kameradan iborat?",
      "image_url": null,
      "options": [
        { "id": "opt-1", "text": "2" },
        { "id": "opt-2", "text": "3" },
        { "id": "opt-3", "text": "4" }
      ]
    }
  ]
}
```

To'g'ri javob bu yerda **kelmaydi**. `time_limit_sec` `null` bo'lsa, vaqt cheklanmagan.

### 3.2 Boshlash

`POST /mobile/test/:id/start` (body yo'q)

```json
{ "attempt_id": "attempt-uuid", "started_at": "2026-10-26T13:00:00+05:00", "time_limit_sec": 600 }
```

Har chaqiruv **yangi urinish** ochadi. "Testni qayta ishlash" tugmasi ham shu endpointni chaqiradi.

### 3.3 Javob berish (T-09)

`POST /mobile/test/attempt/:attempt_id/answer`

```json
{ "question_id": "question-uuid", "option_id": "opt-2" }
```

Javob:

```json
{ "is_correct": false, "correct_option_id": "opt-3" }
```

- `is_correct: true` bo'lsa, tanlangan variantni yashil qiling.
- `false` bo'lsa, tanlanganini qizil, `correct_option_id` ni yashil qiling.
- Bitta savolga faqat bitta javob hisoblanadi. Qayta yuborilsa birinchi javob natijasi qaytadi.

Xatolar:

| Kod | `message` | Ma'nosi |
|-----|-----------|---------|
| 404 | `attempt not found` | Urinish yo'q yoki boshqa foydalanuvchiniki |
| 409 | `test time is up` | Vaqt tugagan, darhol `finish` chaqiring |
| 409 | `attempt already finished` | Urinish yakunlangan |
| 409 | `option does not belong to question` | Noto'g'ri `option_id` |

### 3.4 Yakunlash (T-10)

`POST /mobile/test/attempt/:attempt_id/finish` (body yo'q)

```json
{
  "attempt_id": "…",
  "test_id": "…",
  "correct_count": 17,
  "total_count": 20,
  "percent": 85,
  "next_lesson_id": "lesson-uuid"
}
```

- Natijani foizda ko'rsating (`percent`).
- **"Keyingi mavzuga o'tish"**: `next_lesson_id` ga o'ting. Bu shu kursdagi keyingi dars (keyingi modulda bo'lishi ham mumkin). `null` bo'lsa, bu kursning oxirgi darsi, tugmani yashiring.
- **"Testni qayta ishlash"**: yana `POST /mobile/test/:id/start`.
- Javob berilmagan savollar xato hisoblanadi.
- Vaqt tugaganda ham `finish` ni chaqiring.
- Takror chaqirilsa, o'sha natija qaytadi.

### Oqim

```
GET  /mobile/lesson/:id              → quizzes[].id
GET  /mobile/test/:id                → savollar
POST /mobile/test/:id/start          → attempt_id
POST /mobile/test/attempt/:aid/answer   (har savol uchun)
POST /mobile/test/attempt/:aid/finish   → percent, next_lesson_id
```

## 4. Profil (T-15)

`GET /mobile/user/get/profile` ga yangi maydonlar qo'shildi:

```json
{
  "id": "…",
  "name": "…",
  "is_premium": true,
  "premium_until": "2026-11-26T13:00:00+05:00",
  "is_ustoz": false,
  "first_login_at": "…",
  "last_seen_at": "…"
}
```

- `is_ustoz: true` bo'lsa, "Ustoz" belgisini ko'rsating.
- `is_premium` va `premium_until` bilan Premium holati va tugash sanasini ko'rsating.

### Test natijalari tarixi

`GET /mobile/user/test-history?page=1&limit=10`

```json
{
  "total": 12, "page": 1, "limit": 10, "total_pages": 2, "has_next": true, "has_previous": false,
  "data": [
    {
      "attempt_id": "…", "test_id": "…", "test_title": "1-test", "lesson_id": "…",
      "correct_count": 17, "total_count": 20, "percent": 85, "finished_at": "…"
    }
  ]
}
```

## 5. Premium (T-11, T-12)

Alohida 1, 3, 6, 12 oylik tariflar endi ishlatilmaydi. Tarif tanlash ekranini olib tashlang, uning o'rniga **PREMIUM** tugmasi qo'yiladi.

### 5.1 Narx va holat

`GET /mobile/premium`

```json
{ "price": 50000, "duration_months": 1, "is_premium": false, "premium_until": null, "is_ustoz": false }
```

- `price` so'mda.
- `price: 0` bo'lsa, admin hali narx qo'ymagan. Sotib olish tugmasini yashiring.

### 5.2 Sotib olish

`POST /mobile/premium/order`

```json
{ "promocode_id": "…" }
```

Body ixtiyoriy, promokodsiz bo'sh yuborish mumkin.

```json
{ "message": "order created successfully", "order_id": "…", "url": "https://my.click.uz/services/pay?..." }
```

- `url` ni Click'da oching.
- To'lov o'tgach, `GET /mobile/premium` ni qayta chaqirib holatni yangilang.
- Push ham keladi: `payment_success`.
- Premium faol paytda yana to'lansa, muddat **uzaytiriladi** (tugash sanasiga +1 oy).
- Premium **barcha kurslarni** ochadi.
- Promokod premium narxiga qo'llanadi. Faqat "barcha kurslar" turidagi promokod premiumga ishlaydi.

**Promokodni tekshirish:** `POST /mobile/promocode/check`, body'da `type: "premium"`. Bu holda `course_id` va `tariff_id` kerak emas.

```json
{ "type": "premium", "promocode": "BAHOR10" }
```

Javob avvalgidek: `is_valid`, `total_amount`, `discount_amount` va boshqalar. Promokod tekshiruvdan o'tsa, uning `promocode_id` qiymatini `POST /mobile/premium/order` ga yuboring.

Xato: `400 {"message": "premium price is not configured"}`, narx qo'yilmagan.

## 6. Reyting (T-14)

`GET /mobile/user/rating?type=total&by=test&limit=10`

| Parametr | Qiymatlar | Standart |
|----------|-----------|----------|
| `type` | `day`, `week`, `month`, `year`, `total` | `total` |
| `by` | `activity` (avvalgi reyting), `test` (test natijalari) | `activity` |

```json
{
  "type": "total",
  "by": "test",
  "limit": 10,
  "items": [
    { "user_id": "…", "name": "Ali", "image_url": null, "score": 92.5, "tests_count": 14, "activity": 0, "is_ustoz": true, "rank": 1, "is_me": false }
  ],
  "me": { "user_id": "…", "name": "Siz", "score": 61, "tests_count": 3, "rank": 48, "is_me": true }
}
```

- `by=test`: `score` har test bo'yicha eng yaxshi natijalarning **o'rtacha foizi**.
- `by=activity`: `activity` avvalgidek, `score` ham shu qiymatga teng.
- `me` doim keladi. Foydalanuvchi top ro'yxatda bo'lmasa ham o'z o'rnini ko'radi.

## 7. Push xabarlari

| `data.type` | Qachon | Nima qilish |
|-------------|--------|-------------|
| `inactive_reminder` | 24 soat kirilmaganda (T-05) | Ilovani bosh sahifada ochish |
| `admin_notification` | Admin qo'lda yuborganda | Avvalgidek |
| `payment_success`, `payment_failed` | Kurs yoki premium to'lovidan keyin | Premium uchun `course_id` bo'sh keladi |

`new_course`, `lesson_added`, `source_added`, `*_updated` pushlari endi **kelmaydi** (T-17).
