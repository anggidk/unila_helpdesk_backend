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

### Authentication
- `POST /auth/login` - Login dengan username/password
- `POST /auth/refresh` - Refresh access token

### Tickets
- `GET /tickets` (auth)
- `GET /tickets/search` (public)
- `GET /tickets/:id` (optional auth)
- `POST /tickets` (auth)
- `POST /tickets/:id` (auth)
- `POST /tickets/:id/delete` (auth)

### Surveys & Reports
- `GET /surveys` (public)
- `GET /surveys/categories/:categoryId` (public)
- `POST /surveys` (admin)
- `POST /surveys/responses` (registered)
- `GET /notifications` (auth)
- `POST /notifications/fcm` (auth)
- `GET /reports` (admin)
- `GET /reports/cohort` (admin)

## JWT Token Management

Aplikasi menggunakan dual-token system:
1. **Access Token**: JWT token untuk otentikasi API (expired: 12h untuk user, 8h untuk admin)
2. **Refresh Token**: Long-lived token untuk mendapatkan access token baru (expired: 30 hari untuk user, 7 hari untuk admin)

### Login Response
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expiresAt": "2026-02-04T12:03:45Z",
  "refreshToken": "random-base64-encoded-string",
  "refreshExpiresAt": "2026-03-05T12:03:45Z",
  "user": {
    "id": "user-id",
    "username": "user123",
    "name": "User Name",
    "email": "user@example.com",
    "role": "registered"
  }
}
```

### Refresh Token Usage
```bash
curl -X POST http://localhost:8080/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "your-refresh-token"}'
```

Catatan: akun guest tidak diizinkan mengisi survey. Survey hanya bisa diisi pengguna terdaftar dan tiket berstatus selesai.
