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
| T-11 | Kurs, to'lov | Tarif tanlash yo'q; eski tarif to'lovi `426` qaytaradi |
| T-12 | Premium, kurs | 3 ta paket (Fundamental, Preklinik, Klinik), har biri alohida sotiladi; kursda `package` |
| T-14 | Reyting | `by=test` parametri qo'shildi, `me` doim keladi |
| T-15 | Profil | Paketlar holati, `is_ustoz`, `tests_count`, test bo'yicha guruhlangan tarix |
| T-17 | — | Kontent qo'shilganda avtomatik push kelmaydi, faqat admin yuboradi |
| — | Dars, test | Yopiq darsga kirilsa `403` qaytadi |

Yangi endpointlar avtomatik ravishda foydalanuvchining "oxirgi faollik" vaqtini yangilaydi. Ilova alohida hech narsa yuborishi shart emas.

---

## 1. Kurs sahifasi (T-06)

`GET /mobile/course/:id` avvalgidek ishlaydi, lekin endi:

- Kontenti yo'q darslar va darssiz modullar **qaytmaydi**. Bo'sh bo'lim ko'rsatmang.
- Har bir darsda `content_types` bor: shu darsda nima borligi.
- `has_access` endi `true` bo'ladi, agar foydalanuvchi kursni sotib olgan **yoki** kurs fani tegishli paketda premium faol **yoki** ustoz bo'lsa.
- `tests` soni endi ilova ichidagi testlar soni (eski PDF testlar hisoblanmaydi).
- `package`: kursni qaysi premium paket ochadi (`fundamental`, `preclinical`, `clinical` yoki `null`). `has_access: false` bo'lsa tugma: «♛ <paket nomi> Premiumini olish», bosilganda Premium ekrani shu paket tanlangan holda ochiladi.

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
  "last_seen_at": "…",
  "tests_count": 12,
  "packages": [
    { "package": "fundamental", "is_active": true,  "until": "2026-11-29T13:00:00+05:00" },
    { "package": "preclinical", "is_active": false, "until": null },
    { "package": "clinical",    "is_active": false, "until": null }
  ]
}
```

- `is_ustoz: true` bo'lsa, "Ustoz" belgisini ko'rsating.
- Premium kartochkasi: `is_ustoz: true` bo'lsa «Ustoz — doimiy kirish», tugmasiz. Aks holda `packages` dagi har bir faol paket uchun alohida qator («Premium: <paket> · DD.MM.YYYY gacha» + «Uzaytirish»). Hech biri faol bo'lmasa «Premium yo'q» + «Premium olish».
- `tests_count`: «Test natijalari» bandi yonidagi son (kamida bir marta yakunlangan testlar).

### Test natijalari tarixi

`GET /mobile/user/test-history?page=1&limit=10`

Urinishlar **test bo'yicha guruhlangan**: har test bitta qator, eng yangisi yuqorida.

```json
{
  "summary": { "tests_count": 12, "average_percent": 84, "best_percent": 96 },
  "total": 12, "page": 1, "limit": 10, "total_pages": 2, "has_next": true, "has_previous": false,
  "data": [
    {
      "test_id": "…", "test_title": "Osteologiya", "lesson_id": "…",
      "course_id": "…", "course_name": "Anatomiya",
      "best_percent": 96, "attempts": 2,
      "last_attempt_id": "…", "last_finished_at": "2026-10-12T10:00:00+05:00"
    }
  ]
}
```

- `summary`: yuqoridagi ko'rsatkichlar. `average_percent` har testning eng yaxshi natijasi bo'yicha o'rtacha.
- Qatorda: foiz doirasi `best_percent`, ostida `course_name · sana · attempts urinish`.
- Bosilganda natija ekrani (T-10): `POST /mobile/test/attempt/{last_attempt_id}/finish` yakunlangan urinish natijasini qayta qaytaradi.

## 5. Premium (T-11, T-12)

Alohida 1, 3, 6, 12 oylik tariflar endi ishlatilmaydi. Premium **3 ta paketga** bo'lingan, har biri alohida sotib olinadi:

| `package` | Nomi | Nimani ochadi |
|-----------|------|---------------|
| `fundamental` | Fundamental fanlar | Shu paketga biriktirilgan fanlarning barcha kurslari |
| `preclinical` | Preklinik fanlar | 〃 |
| `clinical` | Klinik fanlar | 〃 |

Qaysi fan qaysi paketda ekanini admin belgilaydi. `GET /mobile/subject` javobidagi har bir fanda `package` maydoni bor (`null` bo'lsa, fan hech qaysi paketda emas).

### 5.1 Narx va holat

`GET /mobile/premium`

```json
{
  "is_premium": true,
  "premium_until": "2026-11-26T13:00:00+05:00",
  "is_ustoz": false,
  "packages": [
    { "package": "fundamental", "price": 50000, "duration_months": 1, "is_active": true,  "until": "2026-11-26T13:00:00+05:00" },
    { "package": "preclinical", "price": 60000, "duration_months": 1, "is_active": false, "until": null },
    { "package": "clinical",    "price": 70000, "duration_months": 1, "is_active": false, "until": null }
  ]
}
```

- `price` so'mda. `price: 0` bo'lsa, admin bu paketga hali narx qo'ymagan: sotib olish tugmasini yashiring.
- `is_premium` kamida bitta paket faol ekanini bildiradi.
- Paketlarga bo'linishdan oldin umumiy premium sotib olganlarda muddati tugaguncha **uchala paket ham** `is_active: true` bo'ladi.

### 5.2 Sotib olish

`POST /mobile/premium/order`

```json
{ "package": "fundamental", "promocode_id": "…" }
```

`package` majburiy, `promocode_id` ixtiyoriy.

```json
{ "message": "order created successfully", "order_id": "…", "url": "https://my.click.uz/services/pay?..." }
```

- `url` ni Click'da oching.
- To'lov o'tgach, `GET /mobile/premium` ni qayta chaqirib holatni yangilang.
- Push ham keladi: `payment_success`.
- Paket faol paytda yana to'lansa, **shu paket** muddati uzaytiriladi (tugash sanasiga +1 oy). Boshqa paketlarga ta'sir qilmaydi.
- Promokod paket narxiga qo'llanadi. Faqat "barcha kurslar" turidagi promokod premiumga ishlaydi.

**Promokodni tekshirish:** `POST /mobile/promocode/check`, body'da `type: "premium"` va `package`. Bu holda `course_id` va `tariff_id` kerak emas.

```json
{ "type": "premium", "package": "fundamental", "promocode": "BAHOR10" }
```

Javob avvalgidek: `is_valid`, `total_amount`, `discount_amount` va boshqalar. Promokod tekshiruvdan o'tsa, uning `promocode_id` qiymatini `POST /mobile/premium/order` ga yuboring.

Xatolar (`400`):

| `message` | Ma'nosi |
|-----------|---------|
| `premium price is not configured` | Bu paketga narx qo'yilmagan |
| `package must be one of: fundamental, preclinical, clinical` | `package` noto'g'ri yoki yuborilmagan |

### 5.3 Eski tarif to'lovi (T-11)

`POST /mobile/order` (kurs + tarif bo'yicha to'lov) yopilgandan keyin `426 {"message": "tariff sales are closed, please update the app"}` qaytaradi. Eski ilova versiyasida bu holatda «Ilovani yangilang» ko'rsatiladi. Yangi ilova bu endpointni chaqirmaydi.

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
