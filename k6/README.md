# iMed Backend — k6 Load Tests

## O'rnatish

```bash
# macOS
brew install k6

# yoki Docker
docker run --rm -i grafana/k6 run - <k6/smoke.js
```

## Tez boshlash

```bash
# 1. Token bilan (eng oson)
export BASE_URL=http://localhost:8004
export ACCESS_TOKEN="eyJhbGciOiJIUzI1NiIs..."
export COURSE_ID="uuid-bu-yerda"
export LESSON_ID="uuid-bu-yerda"

# 2. Smoke (30s, 1 user)
k6 run k6/smoke.js

# 3. Asosiy load test (~9 min, 30→80 VU)
k6 run k6/load-read.js

# 4. Og'ir endpointlar (50 VU, 5 min)
k6 run k6/load-heavy.js

# 5. Spike (20→200→20 VU)
k6 run k6/spike.js

# 6. Read + write aralash
k6 run k6/load-mixed.js
```

## Test fayllar

| Fayl | Maqsad |
|------|--------|
| `smoke.js` | Deploy keyin tez tekshirish |
| `load-read.js` | Asosiy mobile read trafik |
| `load-heavy.js` | lesson + rating + course (DB stress) |
| `spike.js` | Push notification keyin trafik cho'qqisi |
| `load-mixed.js` | 85% read, 15% write |

## Natijalarni tushunish

- `http_req_duration p(95)` — 95% so'rov shu vaqtdan tez
- `http_req_failed` — xato foizi (5% dan oshmasligi kerak)
- `checks` — status 200 tekshiruvlari

HTML hisobot:

```bash
k6 run --out json=k6/results.json k6/load-read.js
```

## Muhim

- **Production** da faqat `prod-smoke.js` va `prod-load.js` ishlating (past VU)
- `POST /click/*`, `otp/send`, `file-upload-youtube` testlarga kiritilmagan
- `LESSON_ID` va `COURSE_ID` bo'lmasa ba'zi checklar o'tkazib yuboriladi

---

## Production test (keyinroq)

Prod da test qilishdan oldin:

1. **Vaqtni tanlang** — kechki 23:00–06:00 (kam trafik)
2. **Staging yo'q bo'lsa** — avval `prod-smoke.js`, keyin `prod-load.js`
3. **Token** — test akkaunt tokeni; login endpointini stress qilmang
4. **Monitoring** — server CPU/RAM, Postgres connections, Redis, loglar
5. **`ALLOW_PROD=1`** — tasodifiy prod testni oldini oladi

### Prod smoke (~1 min, 1 user, 3 iteratsiya)

```bash
ALLOW_PROD=1 \
BASE_URL=https://prod.imedteam.uz \
ACCESS_TOKEN="eyJ..." \
COURSE_ID="uuid" \
LESSON_ID="uuid" \
k6 run k6/prod-smoke.js
```

### Prod load (~5 min, max 15 VU — default)

```bash
ALLOW_PROD=1 \
BASE_URL=https://prod.imedteam.uz \
ACCESS_TOKEN="eyJ..." \
COURSE_ID="uuid" \
LESSON_ID="uuid" \
PROD_MAX_VUS=15 \
k6 run k6/prod-load.js
```

### `.env.prod` bilan

```bash
set -a && source k6/.env.prod && set +a
k6 run k6/prod-smoke.js
```

`k6/.env.prod` ni `.gitignore` ga qo'shing (tokenlar).

### Prod da ISHLATMANG

| Skript | Sabab |
|--------|--------|
| `spike.js` | 200 VU — prodni qulatishi mumkin |
| `load-heavy.js` | 50 VU og'ir SQL |
| `load-mixed.js` | DB ga yozadi (activity) |
| Login orqali token | auth endpoint stress |

### Natijani saqlash

```bash
mkdir -p k6/results
k6 run --out json=k6/results/prod-$(date +%Y%m%d-%H%M).json k6/prod-load.js
```

### Kutilgan prod natija (taxmin)

| Test | Maqsad |
|------|--------|
| `prod-smoke` | Barcha endpoint 200, p95 < 2s |
| `prod-load` 15 VU | p95 < 3s, failed < 3% |
| 15 VU dan oshsa | Scaling yoki optimizatsiya kerak |
