# Unila Helpdesk Backend (Go)

Backend Go + PostgreSQL untuk aplikasi helpdesk & survey kepuasan.

## Quick Start

1. Siapkan PostgreSQL dan buat database:

```
createdb unila_helpdesk
```

2. Salin env contoh:

```
copy .env.example .env
```

3. Jalankan API:

```
go run ./cmd/api
```

API berjalan di `http://localhost:8080` (default).

## Environment

Lihat `.env.example` untuk daftar lengkap. Poin penting:

- `DATABASE_URL` koneksi Postgres
- `JWT_SECRET` kunci token
- `FCM_ENABLED=true` + `FCM_CREDENTIALS=path/to/serviceAccount.json`

## Integrasi Frontend Flutter

Jalankan aplikasi Flutter dengan base URL backend:

```
flutter run --dart-define=API_BASE_URL=http://localhost:8080
```

## API Ringkas

- `POST /auth/login`
- `POST /auth/guest`
- `GET /tickets` (auth)
- `GET /tickets/search` (public)
- `GET /tickets/:id` (optional auth)
- `POST /tickets` (auth)
- `POST /tickets/:id` (auth)
- `POST /tickets/:id/delete` (auth)
- `GET /surveys` (public)
- `GET /surveys/categories/:categoryId` (public)
- `POST /surveys` (admin)
- `POST /surveys/responses` (registered)
- `GET /notifications` (auth)
- `POST /notifications/fcm` (auth)
- `GET /reports` (admin)
- `GET /reports/cohort` (admin)

Catatan: akun guest tidak diizinkan mengisi survey. Survey hanya bisa diisi pengguna terdaftar dan tiket berstatus selesai.
