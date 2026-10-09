# MK Siaga

Starter monolith modular menggunakan Go, Gin, server-rendered HTML, MySQL (`database/sql`), dan Tailwind CLI v4. Folder PHP `mk_sc` tidak diubah.

## Menjalankan aplikasi

Jalankan semua perintah dari root `D:\CODING\mksiaga`. Prasyarat: Go sesuai `go.mod`, Node.js/npm, dan MySQL jika koneksi database diaktifkan.

```powershell
Copy-Item .env.example .env
go mod download
npm ci
npm run build:css
go run ./cmd/web
```

Buka http://127.0.0.1:8080. Salin `.env.example` hanya saat `.env` belum ada. Environment sistem mengalahkan nilai `.env`.

Untuk mengembangkan CSS, jalankan `npm run dev:css` di terminal kedua. Ubah `web/assets/css/input.css` atau class template; jangan edit hasil `web/static/css/app.css`.

## Development dengan Air

[Air](https://github.com/air-verse/air) otomatis melakukan build dan restart aplikasi saat kode Go, template HTML, `.env`, `go.mod`, atau `go.sum` berubah. Konfigurasi proyek ada di `.air.toml`, dengan executable `.exe` untuk Windows.

Instal sekali (versi yang dipakai proyek):

```powershell
go install github.com/air-verse/air@v1.67.4
```

Air versi ini membutuhkan Go 1.26; Go dapat mengunduh toolchain tersebut otomatis saat instalasi. Versi Go aplikasi pada `go.mod` tidak perlu diubah.

Terminal pertama, dari root proyek:

```powershell
air -c .air.toml
```

Terminal kedua:

```powershell
npm run dev:css
```

Buka http://127.0.0.1:8080. Hentikan proses `go run` sebelumnya jika masih menggunakan port tersebut. Simpan perubahan untuk memicu restart otomatis; refresh browser untuk melihat hasilnya. Air tidak mengaktifkan refresh browser otomatis pada konfigurasi ini. Tekan Ctrl+C untuk menghentikan Air.

Jika perintah `air` belum dikenali, jalankan executable langsung (untuk instalasi dengan GOBIN default):

```powershell
& "$(go env GOPATH)\bin\air.exe" -c .air.toml
```

Jika memakai GOBIN khusus, gunakan `air.exe` dari direktori `go env GOBIN`. Output build Air berada di `tmp/` yang sudah diabaikan Git. File static, node_modules, dan uploads tidak memicu build Go.

## MySQL

Default `DB_ENABLED=false` memungkinkan halaman awal dijalankan tanpa database. Untuk mengaktifkannya:

1. Jalankan MySQL, misalnya dari XAMPP.
2. Buat database development terpisah bernama `mksiaga_dev` menggunakan charset `utf8mb4`.
3. Atur `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, dan `DB_PASSWORD` di `.env`, lalu ubah `DB_ENABLED=true`.
4. Restart aplikasi. Startup akan gagal jika koneksi yang diaktifkan tidak berhasil.

Tidak ada skema atau migrasi yang dijalankan otomatis. Skema PHP lama belum ditinjau; folder `migrations` disediakan untuk perubahan SQL setelah pemetaan data. Jangan arahkan pengembangan ke database produksi.

## Struktur dan tanggung jawab

```text
cmd/web/                 Entry point, dependency wiring, graceful shutdown
internal/config/         Konfigurasi environment
internal/database/       Koneksi dan connection pool MySQL
internal/router/         Route Gin dan pemasangan middleware
internal/dashboard/      Handler halaman awal
internal/view/           Renderer template per halaman
internal/middleware/     Tempat middleware aplikasi berikutnya
internal/auth/           Tempat modul autentikasi berikutnya
internal/user/           Tempat modul pengguna berikutnya
internal/pengaduan/      Calon modul; sesuaikan kebutuhan aplikasi lama
web/templates/layouts/   Kerangka HTML
web/templates/partials/  Komponen HTML bersama
web/templates/<fitur>/   Halaman tiap fitur
web/assets/css/          Sumber Tailwind
web/static/              CSS hasil build, JS, dan gambar publik
migrations/              Tempat migrasi SQL
storage/uploads/         File unggahan privat, tidak dilayani sebagai static
```

Modul nyata pertama adalah dashboard. Folder auth, user, pengaduan, dan middleware berisi petunjuk, bukan implementasi semu. Tambahkan `model.go`, `handler.go`, `repository.go`, serta `service.go` sesuai kebutuhan fitur.

Handler menangani HTTP dan validasi format. Service menangani aturan bisnis tanpa `gin.Context`. Repository menangani SQL dengan parameter binding dan `context.Context`. Dependency diberikan melalui constructor; hindari koneksi global dan interface yang belum diperlukan.

Setiap halaman mendefinisikan `{{define "content"}}...{{end}}`. Renderer menggabungkan layout dan partials secara terpisah untuk setiap halaman, sehingga blok halaman berbeda tidak saling menimpa. Nama render contohnya `dashboard/index.html`.

## Endpoint awal

| Endpoint | Perilaku |
| --- | --- |
| `GET /login` | Form login |
| `POST /login` | Verifikasi akun dan membuat session |
| `GET /` | Dashboard; memerlukan login |
| `POST /logout` | Menghapus session; memerlukan login dan token CSRF |
| `GET /settings/roles` | Pengelolaan role dan assignment permission |
| `GET /settings/permissions` | Pengelolaan permission |
| `GET /healthz` | HTTP 200 ketika proses HTTP hidup |
| `GET /readyz` | HTTP 200 jika MySQL dapat dijangkau; HTTP 503 jika nonaktif/tidak tersedia |
| `GET /static/*` | Aset publik |

Login memakai bcrypt, session cookie yang ditandatangani, serta token CSRF pada login dan logout. `SESSION_SECRET` wajib berisi minimal 32 karakter. Gunakan nilai acak yang berbeda pada setiap environment dan jangan memasukkannya ke Git. Set `SESSION_COOKIE_SECURE=true` ketika aplikasi dijalankan melalui HTTPS.

Seed akun development ada di `seeds/000001_development_superadmin.sql`, sedangkan permission awal pengelolaan akses ada di `seeds/000002_access_management_permissions.sql`. Route modul akses memakai permission terkait dan role `superadmin` selalu memperoleh akses penuh. Otorisasi untuk fitur bisnis lain, rate limiting login, pergantian store aktif, dan migrasi otomatis belum diimplementasikan. Proxy belum dipercaya; konfigurasikan alamat proxy spesifik jika nanti memakai reverse proxy.

Menu aplikasi didefinisikan terpusat di `internal/navigation/navigation.go`, difilter menggunakan permission pada store aktif, lalu dirender oleh partial `web/templates/partials/app_shell.html`. Dashboard dan seluruh halaman pengelolaan memakai partial yang sama. Tambahkan definisi menu dan permission di sana ketika modul baru tersedia.

## Pemeriksaan dan build

```powershell
go test ./...
go vet ./...
npm run build:css
go build -o bin/mksiaga.exe ./cmd/web
```

Jalankan `./bin/mksiaga.exe` dari root proyek. Deployment perlu membawa folder `web/templates` dan `web/static` bersama binary karena aset belum di-embed. Node.js diperlukan hanya pada tahap build CSS.

Referensi: [Gin](https://gin-gonic.com/en/docs/quickstart/) dan [Tailwind CLI](https://tailwindcss.com/docs/installation/tailwind-cli).
