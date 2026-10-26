# TibMentor: admin panel uchun API o'zgarishlari

Bu hujjat TibMentor yangilanishida (TZ T-01…T-17) admin panelga qo'shilishi kerak bo'lgan ekranlar va ular chaqiradigan API'larni tushuntiradi.

- Barcha so'rovlar admin tokeni bilan yuboriladi: `Authorization: Bearer <access_token>`.
- To'liq sxema Swagger'da: `/api/swagger/index.html`.

## Qisqacha: nima qilish kerak

| TZ | Ekran | Ish |
|----|-------|-----|
| T-04 | Foydalanuvchilar | Birinchi kirish va oxirgi faollik ustunlari |
| T-07, T-08 | Darslar → Testlar | Test yaratish, savol/variant qo'shish, Excel'dan import |
| T-06 | Manbalar (sources) | `audio` turi qo'shildi |
| T-11, T-12 | Premium sozlamalari | Oylik narxni kiritish |
| T-16 | Foydalanuvchilar | "Ustoz" unvonini berish/olish |
| T-17 | Xabarnomalar | Avtomatik push o'chirildi, xabarni admin o'zi yuboradi |

---

## 1. Foydalanuvchilar ro'yxati (T-04, T-15, T-16)

`GET /web/user` va `GET /web/user/:id` endi qo'shimcha maydonlarni qaytaradi:

```json
{
  "id": "…",
  "name": "…",
  "phone_number": "…",
  "created_at": "…",
  "first_login_at": "2026-03-12T09:14:00+05:00",
  "last_seen_at": "2026-10-26T12:58:00+05:00",
  "is_premium": true,
  "premium_until": "2026-11-26T13:00:00+05:00",
  "is_ustoz": false
}
```

Jadvalga qo'shiladigan ustunlar:

| Ustun | Maydon | Izoh |
|-------|--------|------|
| Birinchi kirish | `first_login_at` | Eski foydalanuvchilarda ro'yxatdan o'tgan vaqt yozilgan |
| Oxirgi faollik | `last_seen_at` | Taxminan 5 daqiqa aniqlikda yangilanadi. Yangilanishdan keyin hali kirmaganlarda `null` |
| Premium | `is_premium`, `premium_until` | |
| Ustoz | `is_ustoz` | Belgi yoki tugma |

### "Ustoz" unvoni (T-16)

`PUT /web/user/:id/ustoz`

```json
{ "is_ustoz": true }
```

Javob: `{"message": "ustoz updated successfully"}`

- Unvon berilgan zahoti foydalanuvchiga **barcha kurslar doimiy ochiladi**, profilida "Ustoz" belgisi chiqadi.
- `false` yuborilsa, unvon olib tashlanadi.
- Bu `teachers` (o'qituvchilar) bo'limidan alohida: ustoz oddiy ilova foydalanuvchisi.

## 2. Testlar (T-07, T-08)

Test darsga (`lesson_id`) bog'lanadi. Bitta darsda bir nechta test bo'lishi mumkin.

**Qoidalar:**
- Savollar va variantlar soni cheklanmagan.
- Har savolda kamida **2 ta variant** va aniq **1 ta to'g'ri javob** bo'lishi shart.
- Vaqt chegarasi (`time_limit_sec`) ixtiyoriy. `null` bo'lsa, vaqt cheklanmagan.

### 2.1 Ro'yxat

`GET /web/test?lesson_id=…&page=1&limit=10`

```json
{
  "total": 2, "page": 1, "limit": 10,
  "data": [
    { "id": "…", "lesson_id": "…", "title": {"uz": "1-test", "ru": "Тест 1", "en": "Test 1"},
      "time_limit_sec": 600, "order_num": 1, "question_count": 20, "created_at": "…", "updated_at": "…" }
  ]
}
```

### 2.2 Yaratish

`POST /web/test`. Savollarni darhol shu so'rovda yuborish mumkin, keyin alohida qo'shish ham mumkin.

```json
{
  "lesson_id": "lesson-uuid",
  "title": { "uz": "1-test", "ru": "Тест 1", "en": "Test 1" },
  "time_limit_sec": 600,
  "order_num": 1,
  "questions": [
    {
      "text": "Yurak nechta kameradan iborat?",
      "image_url": "",
      "options": [
        { "text": "2", "is_correct": false },
        { "text": "4", "is_correct": true }
      ]
    }
  ]
}
```

Javob: `{"message": "test created successfully", "Id": "test-uuid"}`

Biror savol qoidaga to'g'ri kelmasa, hech narsa saqlanmaydi va xato qaytadi, masalan: `question 3: question "…" must have exactly 1 correct option`.

### 2.3 Ko'rish

`GET /web/test/:id` testni barcha savol va variantlari bilan qaytaradi (`is_correct` bilan).

### 2.4 Tahrirlash va o'chirish

| Amal | So'rov | Body |
|------|--------|------|
| Test nomi, vaqti, tartibi | `PUT /web/test/:id/update` | `{ "title": {...}, "time_limit_sec": 600, "order_num": 1 }` |
| Testni o'chirish | `DELETE /web/test/:id/delete` | — |
| Savol qo'shish | `POST /web/test/:id/question` | `{ "text": "…", "image_url": "", "options": [...] }` |
| Savolni tahrirlash | `PUT /web/test/question/:question_id/update` | `{ "text": "…", "order_num": 2, "options": [...] }` |
| Savolni o'chirish | `DELETE /web/test/question/:question_id/delete` | — |

- `time_limit_sec` ni olib tashlash uchun `null` yuboring.
- Savolni tahrirlashda variantlar **to'liq almashtiriladi**, shuning uchun hamma variantni qayta yuboring.

### 2.5 Excel'dan import

1. Namuna faylni yuklab oling: `GET /web/test/import-template` (`test_template.xlsx`).
2. Faylni to'ldirib yuboring: `POST /web/test/:id/import`, `multipart/form-data`, maydon nomi `file`.

Javob: `{"imported": 25}`

Fayl formati: **birinchi qator sarlavha**, keyingi har bir qator bitta savol.

| Savol | A | B | C | D | To'g'ri javob |
|-------|---|---|---|---|---------------|
| Yurak nechta kameradan iborat? | 2 | 3 | 4 | 5 | C |
| Eng katta organ qaysi? | Jigar | Teri | O'pka | | B |

- Sarlavhadagi **oxirgi ustun** to'g'ri javob, unga variant harfi yoziladi (`A`, `B`, …).
- Ko'proq variant kerak bo'lsa, sarlavhaga ustun qo'shing (`E`, `F`, …).
- Variantlar o'rtasida bo'sh katak bo'lmasligi kerak. Oxiridagi bo'sh kataklar mumkin.
- Savollar testning mavjud savollari oxiriga qo'shiladi.
- Bitta qatorda xato bo'lsa, **butun fayl saqlanmaydi**. Xatolar qator raqami bilan qaytadi, masalan: `row 4: invalid correct answer "E"; row 9: options must not have gaps`.

### Eski PDF testlar

`sources` dagi `type: "test"` (PDF) yozuvlar bazada qoladi, lekin mobil ilovada **ko'rinmaydi**. Yangi testlarni faqat yuqoridagi API orqali qo'shing.

## 3. Manbalar: audio turi (T-06)

`POST /web/source` endi `type: "audio"` ni ham qabul qiladi: `video`, `audio`, `document`, `test`.

1. Faylni yuklang: `POST /web/file-upload-s3`, masalan mp3.
2. Qaytgan url bilan manba yarating:

```json
{ "lesson_id": "…", "name": {"uz": "Audio ma'ruza"}, "url": {"uz": "https://…/lecture.mp3"}, "type": "audio", "order_num": 1 }
```

Mobil ilova faqat kontenti bor bo'limlarni ko'rsatadi: darsda kontent yo'q bo'lsa, u ilovada umuman chiqmaydi.

## 4. Premium sozlamalari (T-11, T-12)

| Amal | So'rov |
|------|--------|
| Ko'rish | `GET /web/premium/settings` |
| O'zgartirish | `PUT /web/premium/settings` |

Ikkalasida ham format bir xil:

```json
{ "price": 50000, "duration_months": 1 }
```

- `price` so'mda. **0 bo'lsa, foydalanuvchilar premium sotib ola olmaydi.** Ishga tushirishdan oldin narx qo'yilishi shart.
- `duration_months`: bitta to'lov necha oy beradi (standart 1).
- Narx o'zgarsa, yangi buyurtmalarga ta'sir qiladi. Allaqachon to'langanlar o'zgarmaydi.

### Buyurtmalar

`GET /web/order` va `GET /web/order/:id` javobida yangi maydon bor: `order_type` (`course` yoki `premium`).

- Premium buyurtmada `course_id`, `course_name`, `tariff_id`, `tariff_name` **`null`** bo'ladi. Jadvalda "Premium" deb ko'rsating.
- Promokod premium uchun ham ishlaydi, lekin faqat "barcha kurslar" (`type: all`) turidagisi. Muayyan kurslarga bog'langan promokod premiumga qo'llanmaydi.

### Tariflar

Eski 1/3/6/12 oylik tariflar va kurs narxlari **o'chirilmadi**. Ilgari sotib olganlar muddati tugaguncha kursdan foydalanadi. Yangi sotuv mobil ilovada faqat Premium orqali bo'ladi.

## 5. Xabarnomalar (T-17)

Kurs, modul, dars, manba, fan yoki banner qo'shilganda yoki o'zgartirilganda **avtomatik push ketmaydi**. Yangilash tugagach, xabarni o'zingiz yuboring: avvalgidek `POST /web/notification`.

24 soat kirmagan foydalanuvchilarga eslatma (T-05) avtomatik ketadi, admin hech narsa qilmaydi.
