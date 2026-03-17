# Unila Helpdesk Backend (Go)

Backend Go + PostgreSQL untuk aplikasi helpdesk dan survey kepuasan.

## Prasyarat

- Go 1.22+ (disarankan versi terbaru yang kompatibel dengan `go.mod`)
- PostgreSQL 13+
- File `.env` (disalin dari `.env.example`)

## Quick Start

1. Buat database PostgreSQL.

Windows (PowerShell):

```powershell
createdb unila_helpdesk
```

Linux/macOS:

```bash
createdb unila_helpdesk
```

2. Salin env contoh menjadi `.env`.

Windows (PowerShell):

```powershell
Copy-Item .env.example .env
```

Linux/macOS:

```bash
cp .env.example .env
```

3. Atur nilai penting di `.env`, minimal:

- `DATABASE_URL` untuk koneksi Postgres
- `JWT_SECRET` untuk signing token
- `FCM_ENABLED=true` dan `FCM_CREDENTIALS=path/to/serviceAccount.json` jika memakai push notification

4. Jalankan API.

```bash
go run ./cmd/api
```

Default API berjalan di `http://localhost:8080`.

## Seed Data

Project menyediakan command seed berikut:

```bash
go run ./cmd/seed
go run ./cmd/seed_random_tickets
```

Gunakan sesuai kebutuhan untuk data awal/testing.

## Integrasi Frontend Flutter

Saat menjalankan Flutter app, set base URL backend:

```bash
flutter run --dart-define=API_BASE_URL=http://localhost:8080
```

## API Ringkas

### Authentication

- `POST /auth/login` - login dengan username/password
- `POST /auth/refresh` - refresh access token

### Tickets

- `GET /tickets` (auth)
- `GET /tickets/search` (public)
- `GET /tickets/:id` (optional auth)
- `POST /tickets` (auth)
- `POST /tickets/:id` (auth)
- `POST /tickets/:id/delete` (auth)

### Surveys and Reports

- `GET /surveys` (public)
- `GET /surveys/categories/:categoryId` (public)
- `POST /surveys` (admin)
- `POST /surveys/responses` (registered)
- `GET /notifications` (auth)
- `POST /notifications/fcm` (auth)
- `GET /reports` (admin)
- `GET /reports/cohort` (admin)

Dokumen detail endpoint tersedia di folder `docs/`.

## JWT Token Management

Aplikasi menggunakan dual-token system:

1. Access Token: JWT untuk otentikasi API (expired: 12 jam untuk user, 8 jam untuk admin)
2. Refresh Token: token untuk mendapatkan access token baru (expired: 30 hari untuk user, 7 hari untuk admin)

### Contoh Response Login

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

### Contoh Refresh Token

```bash
curl -X POST http://localhost:8080/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "your-refresh-token"}'
```

Catatan: akun guest tidak diizinkan mengisi survey. Survey hanya bisa diisi pengguna terdaftar dan tiket berstatus selesai.
