# Merge Plan v2 - Integrasi dengan Sistem Existing PHP + MySQL

Dokumen ini adalah versi pembaruan berdasarkan keputusan diskusi:

1. Sistem web existing PHP tetap berjalan untuk manajemen tiket.
2. APK tetap dipakai.
3. Integrasi utama tiket dilakukan lewat endpoint PHP existing.
4. Backend Go dipakai sebagai adapter/orchestrator + modul survei/analitik.

## 1. Tujuan Implementasi

1. APK bisa create/edit/delete/list/detail tiket dari data existing.
2. Data lama dari sistem PHP bisa tampil di APK.
3. Modul survei kepuasan dan analitik indeks tetap aktif.
4. Perubahan Flutter seminimal mungkin.
5. Tidak perlu akses langsung ke DB existing pada fase awal.

## 2. Arsitektur Target (Disarankan)

Alur request:

1. Flutter -> Go API (endpoint tetap).
2. Go API -> PHP API existing (khusus operasi tiket + auth/SSO jika tersedia).
3. Go API -> MySQL modul baru (khusus survey/fcm/report).

Prinsip:

1. Tabel tiket existing tetap dimiliki PHP.
2. Tabel baru untuk modul survey/fcm dikelola Go.
3. Integrasi data antar modul memakai `ticket_id` dan `user_id`.

## 3. Keputusan Data Model

1. Karena sistem PHP saat ini tidak punya `ticket history`, maka response detail tiket di APK:
   - `history` dikirim kosong (`[]`) atau tidak dipakai di UI.
   - status diambil dari status tiket terbaru (single source dari tabel tiket existing).
2. Ticket lifecycle tetap mengikuti rule sistem existing.

## 4. Persiapan Wajib

## 4.1 Dari Tim Sistem Existing (PHP)

1. Daftar endpoint tiket (method + URL):
   - create
   - update/edit
   - delete/close
   - list/pagination
   - detail by id
   - search/filter status
2. Kontrak request/response JSON tiap endpoint.
3. Mekanisme auth:
   - token/JWT/session
   - refresh token (jika ada)
4. Endpoint auth SSO (jika login APK harus ikut akun existing).

## 4.2 Dari Sisi Go

1. Konfigurasi env integrasi PHP:
   - `PHP_API_BASE_URL`
   - `PHP_API_KEY` atau credentials integrasi
   - timeout dan retry policy
2. MySQL untuk modul baru (survey/fcm/report).
3. Staging environment terpisah.

## 5. Perubahan File (Estimasi)

## 5.1 Backend Go (utama)

1. `internal/config/config.go` - env konfigurasi integrasi PHP.
2. `cmd/api/main.go` - wiring client PHP dan dependency service.
3. `internal/service/ticket_service.go` - operasi tiket dialihkan ke PHP API.
4. `internal/service/auth_service.go` - opsional, jika auth/SSO via PHP endpoint.
5. `internal/handler/ticket_handler.go` - minor error mapping.
6. `internal/handler/auth_handler.go` - minor mapping jika auth via PHP.
7. `README.md` + `.env.example` - dokumentasi env baru.

## 5.2 Backend Go (file baru)

1. `internal/integration/php/client.go`
2. `internal/integration/php/ticket_client.go`
3. `internal/integration/php/auth_client.go` (opsional)
4. `internal/integration/php/models.go`

## 5.3 Flutter

Target minimal:

1. Tidak ubah field inti model tiket.
2. Jika tidak ada history, UI detail tiket menampilkan status terakhir saja.
3. Kontrak endpoint APK -> Go dipertahankan agar perubahan Flutter kecil.

## 6. Rencana Eksekusi Bertahap

## Fase 0 - Kontrak Integrasi

1. Kunci kontrak endpoint PHP (request/response + error code).
2. Buat mapping PHP DTO <-> Go DTO.
3. Definisikan fallback:
   - timeout PHP
   - error gateway
   - retry policy.

Output:

1. Mapping document final.
2. Daftar endpoint prioritas implementasi.

## Fase 1 - Adapter Ticket di Go

1. Implement HTTP client ke PHP endpoint.
2. Alihkan operasi tiket di `ticket_service` dari DB langsung ke PHP API.
3. Pertahankan response contract ke Flutter.

Output:

1. Endpoint tiket Go tetap sama, sumber data berubah ke PHP.

## Fase 2 - Survey/FCM/Report di MySQL

1. Siapkan tabel baru modul survey/fcm/report di MySQL.
2. Simpan referensi `ticket_id`/`user_id` dari sistem existing.
3. Jalankan perhitungan indeks kepuasan otomatis sesuai framework.

Output:

1. Modul survei dan report aktif tanpa mengganggu tabel tiket existing.

## Fase 3 - UAT dan Pilot

1. Uji end-to-end:
   - login
   - create/edit/delete/list/detail tiket
   - submit survey
   - lihat report indeks.
2. Pilot ke subset user.
3. Monitoring error dan performa.

Output:

1. Siap roll-out penuh.

## 7. Testing Checklist

1. Contract test:
   - parser JSON dari PHP stabil.
2. Functional test:
   - tiket create/edit/delete/list/detail berjalan dari APK.
3. Consistency test:
   - status tiket di APK sama dengan web existing.
4. Resilience test:
   - PHP timeout/error ditangani baik.
5. Survey/report test:
   - indeks kepuasan terhitung otomatis dan konsisten.

## 8. Rollback Plan

Jika ada isu:

1. Nonaktifkan adapter tiket Go (feature flag).
2. Arahkan sementara APK langsung ke endpoint yang paling stabil.
3. Tetap jaga web PHP sebagai jalur utama operasional tiket.
4. Perbaiki integrasi di staging, lalu deploy ulang.

## 9. Estimasi Waktu Implementasi

Dengan kontrak PHP endpoint sudah tersedia:

1. Mapping + desain integrasi: 1-2 hari.
2. Implement adapter tiket di Go: 3-5 hari.
3. Integrasi auth/SSO + testing: 2-4 hari.
4. UAT + hardening + pilot: 2-4 hari.

Total estimasi: 8-15 hari kerja.

## 10. Kesimpulan

1. Merge bisa dilakukan tanpa mematikan sistem existing.
2. APK tetap bisa membuat dan mengelola tiket dengan data existing.
3. Ketiadaan ticket history di sistem existing bukan blocker, cukup gunakan status terbaru.
4. Pendekatan paling aman: tiket via PHP endpoint, modul baru via Go + MySQL.
