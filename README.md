# MK Siaga Base Project

MK Siaga adalah base project admin berbasis Go, Gin, server-rendered HTML, MySQL, dan Tailwind CSS. Project ini sudah menyediakan autentikasi, manajemen user multi-store, role dan permission, master store, profil, manajemen sesi perangkat, halaman error, mode maintenance, SweetAlert, Font Awesome, filter tabel, dan pagination.

Semua waktu aplikasi menggunakan zona `Asia/Jakarta` (WIB).

## Persyaratan sistem

| Komponen | Versi/keterangan |
| --- | --- |
| Go | Sesuai [`go.mod`](go.mod), saat ini Go 1.26 |
| MySQL/MariaDB | MySQL 8 atau MariaDB 10.4+ |
| Node.js | Versi LTS yang mendukung npm dan Tailwind CSS 4 |
| npm | Mengikuti Node.js yang digunakan |
| Git | Disarankan untuk clone dan pengelolaan versi |

Pada Windows, MySQL dari XAMPP dapat digunakan. Contoh perintah dalam dokumentasi memakai lokasi `C:\xampp8.2.12\mysql\bin`.

## Instalasi

Clone atau salin project, lalu jalankan perintah berikut dari root project:

```powershell
Copy-Item .env.example .env
go mod download
npm ci
npm run build:css
```

Buat `SESSION_SECRET` baru dengan panjang minimal 32 karakter. Contoh membuat secret melalui PowerShell:

```powershell
$bytes = New-Object byte[] 32
$rng = [Security.Cryptography.RandomNumberGenerator]::Create()
$rng.GetBytes($bytes)
$rng.Dispose()
([BitConverter]::ToString($bytes) -replace '-', '').ToLower()
```

Salin hasilnya ke `SESSION_SECRET` dalam `.env`.

## Konfigurasi `.env`

Aplikasi membaca `.env` ketika startup. Environment variable sistem memiliki prioritas lebih tinggi daripada isi file tersebut.

```env
APP_NAME=MK Siaga
HTTP_ADDR=127.0.0.1:8080
GIN_MODE=debug
SESSION_SECRET=ganti-dengan-secret-acak-minimal-32-karakter
SESSION_COOKIE_SECURE=false
MAINTENANCE_MODE=false
MAINTENANCE_MESSAGE=

DB_ENABLED=true
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=mksiaga_dev
DB_USER=root
DB_PASSWORD=
```

| Variable | Keterangan |
| --- | --- |
| `APP_NAME` | Nama aplikasi yang ditampilkan pada halaman |
| `HTTP_ADDR` | Alamat dan port HTTP server |
| `GIN_MODE` | `debug`, `release`, atau `test` |
| `SESSION_SECRET` | Kunci penandatangan cookie, minimal 32 karakter |
| `SESSION_COOKIE_SECURE` | Gunakan `true` jika aplikasi berjalan melalui HTTPS |
| `MAINTENANCE_MODE` | Menampilkan halaman maintenance jika bernilai `true` |
| `MAINTENANCE_MESSAGE` | Pesan tambahan pada halaman maintenance |
| `DB_ENABLED` | Mengaktifkan koneksi database |
| `DB_HOST` | Host MySQL/MariaDB |
| `DB_PORT` | Port MySQL/MariaDB |
| `DB_NAME` | Nama database aplikasi |
| `DB_USER` | Username database |
| `DB_PASSWORD` | Password database |

Jangan commit `.env`. Gunakan secret dan kredensial berbeda untuk setiap environment.

## Database, migrasi, dan seed

Project tidak menjalankan migrasi secara otomatis. Ada dua cara menyiapkan database.

### Opsi 1: SQL starter

Opsi ini paling cepat untuk project baru karena schema dan seluruh data awal sudah digabungkan dalam [`database/mksiaga_starter.sql`](database/mksiaga_starter.sql).

```powershell
& 'C:\xampp8.2.12\mysql\bin\mysql.exe' -u root -e "CREATE DATABASE mksiaga_dev CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

Get-Content -Raw database/mksiaga_starter.sql |
  & 'C:\xampp8.2.12\mysql\bin\mysql.exe' -u root mksiaga_dev
```

Data awal menyediakan akun development berikut:

```text
Username: admin
Password: qweqwe
```

Ganti password segera jika database digunakan di luar komputer development. Data sesi perangkat tidak disertakan dalam SQL starter.

### Opsi 2: migrasi dan seed terpisah

Jalankan migrasi sesuai nomor urut:

```powershell
Get-Content -Raw migrations/000001_create_identity_and_access.up.sql |
  & 'C:\xampp8.2.12\mysql\bin\mysql.exe' -u root mksiaga_dev

Get-Content -Raw migrations/000002_create_user_sessions.up.sql |
  & 'C:\xampp8.2.12\mysql\bin\mysql.exe' -u root mksiaga_dev
```

Kemudian jalankan seed sesuai nomor urut:

```powershell
Get-ChildItem seeds/*.sql | Sort-Object Name | ForEach-Object {
  Get-Content -Raw $_.FullName |
    & 'C:\xampp8.2.12\mysql\bin\mysql.exe' -u root mksiaga_dev
}
```

Seed development bersifat idempotent. Menjalankan ulang seed superadmin akan mengembalikan password akun admin ke password development. Detail tiap migrasi tersedia di [`migrations/README.md`](migrations/README.md).

Untuk membuat perubahan schema baru:

1. Buat pasangan file bernomor berikutnya, misalnya `000003_create_incidents.up.sql` dan `000003_create_incidents.down.sql`.
2. Tulis perubahan maju pada file `.up.sql` dan rollback pada `.down.sql`.
3. Uji keduanya pada database development kosong.
4. Tambahkan seed terpisah jika modul memerlukan data awal.

## Menjalankan aplikasi

Pastikan `.env` dan database sudah siap, lalu jalankan:

```powershell
go run ./cmd/web
```

Buka [http://127.0.0.1:8080](http://127.0.0.1:8080). Endpoint pemeriksaan aplikasi:

| Endpoint | Fungsi |
| --- | --- |
| `GET /healthz` | Memastikan proses HTTP hidup |
| `GET /readyz` | Memastikan aplikasi dan database siap |

Untuk mengembangkan CSS:

```powershell
npm run dev:css
```

Untuk rebuild CSS production:

```powershell
npm run build:css
```

Project juga menyediakan `.air.toml` untuk restart otomatis menggunakan Air:

```powershell
go install github.com/air-verse/air@v1.67.4
air -c .air.toml
```

Jalankan `npm run dev:css` di terminal lain saat mengubah class atau sumber Tailwind.

## Membuat modul baru

Gunakan modul yang ada seperti `internal/store` sebagai pola. Misalnya untuk modul `incident`:

```text
internal/incident/
├── handler.go
├── model.go
├── repository.go
└── service.go

web/templates/incident/
└── index.html
```

Tanggung jawab setiap bagian:

| Bagian | Tanggung jawab |
| --- | --- |
| `model.go` | Struct input, entity, dan data tampilan |
| `repository.go` | Query database dengan parameter binding dan `context.Context` |
| `service.go` | Validasi dan aturan bisnis |
| `handler.go` | Request HTTP, response, redirect, dan render template |
| template | Tampilan server-rendered dan komponen form/tabel |

Langkah menambahkan modul:

1. Buat migrasi schema dan seed permission, misalnya `incidents.view`, `incidents.create`, `incidents.update`, dan `incidents.delete`.
2. Buat package modul dengan constructor repository, service, dan handler.
3. Daftarkan dependency dan route di [`internal/router/router.go`](internal/router/router.go).
4. Lindungi route memakai `auth.RequirePermission(db, "incidents.view")` atau permission sesuai operasinya.
5. Tambahkan menu dan permission-nya ke [`internal/navigation/navigation.go`](internal/navigation/navigation.go).
6. Buat template di `web/templates/<modul>` dengan blok `{{define "content"}}` dan gunakan partial shell admin.
7. Tambahkan JavaScript bersama di `web/static/js/app.js` hanya jika interaksi tidak dapat ditangani HTML biasa.
8. Jalankan format, test, vet, pemeriksaan JavaScript, dan build CSS.

```powershell
gofmt -w internal/incident/*.go
go test ./...
go vet ./...
node --check web/static/js/app.js
npm run build:css
```

## Struktur folder

```text
cmd/web/                    Entry point dan HTTP server
database/                   SQL starter siap import
internal/access/            Modul role dan permission
internal/auth/              Login, cookie session, CSRF, middleware akses
internal/config/            Pembacaan dan validasi environment
internal/dashboard/         Dashboard admin
internal/database/          Koneksi dan pool MySQL
internal/httperror/         Halaman 403, 404, 419, 500, maintenance
internal/navigation/        Definisi menu dan filter permission
internal/profile/           Profil dan perubahan password
internal/router/            Registrasi route dan dependency wiring
internal/sessionmanagement/ Daftar dan pencabutan sesi perangkat
internal/store/             Master data store
internal/user/              Manajemen user dan assignment store
internal/view/              Renderer layout, partial, dan template
migrations/                 Migrasi schema naik/turun
seeds/                      Data awal development
storage/uploads/            Penyimpanan unggahan privat
web/assets/css/             Sumber Tailwind CSS
web/static/                 CSS hasil build dan JavaScript publik
web/templates/              Layout, partial, dan halaman HTML
```

Binary Go belum meng-embed template dan aset. Folder `web/templates`, `web/static`, serta dependency frontend yang disajikan dari `node_modules` harus tersedia di working directory aplikasi.

## Sistem role dan permission

Hak akses ditempelkan pada hubungan user dan store, bukan langsung pada user. Karena itu satu user dapat terdaftar pada beberapa store dan memiliki role berbeda di setiap store.

```text
users --< user_stores >-- stores
              |
              +--< user_store_roles >-- roles --< role_permissions >-- permissions
              |
              +--< user_store_permissions >-------------------------- permissions
```

Alur pemeriksaan akses:

1. Login memilih assignment store default yang aktif.
2. Identitas `user_store_id`, store aktif, dan nama role disimpan dalam session.
3. `RequirePermission` memeriksa permission langsung pada `user_store_permissions` dan permission dari role.
4. Role `superadmin` memperoleh akses penuh sebagai bypass.
5. Navigasi hanya menampilkan menu yang diizinkan untuk store aktif.

Gunakan pola nama permission `<modul>.<aksi>`, misalnya `stores.view` atau `users.update`. Route tetap harus memakai middleware permission meskipun menu sudah disembunyikan, karena menu bukan pengaman request.

Session perangkat disimpan pada tabel `user_sessions`. Sesi biasa berlaku 12 jam dan sesi dengan pilihan **Ingat saya** berlaku 30 hari. Pengguna dapat mencabut sesi perangkat lain melalui menu **Profil → Perangkat & sesi**.

## Proses deployment

### 1. Siapkan environment

- Sediakan server dengan Go atau build binary di CI/build machine.
- Sediakan MySQL/MariaDB dan database khusus aplikasi.
- Gunakan user database dengan hak akses hanya pada database aplikasi.
- Siapkan HTTPS melalui reverse proxy seperti Nginx, Caddy, atau IIS.

### 2. Backup dan update database

Backup database sebelum menjalankan migrasi:

```powershell
& 'C:\xampp8.2.12\mysql\bin\mysqldump.exe' -u root -p --result-file=backup.sql nama_database
```

Jalankan hanya migrasi `.up.sql` yang belum pernah diterapkan. Untuk instalasi baru, gunakan SQL starter atau seluruh migrasi dan seed secara berurutan.

### 3. Build aplikasi

```powershell
npm ci
npm run build:css
go test ./...
go vet ./...
go build -trimpath -ldflags="-s -w" -o bin/mksiaga.exe ./cmd/web
```

Untuk Linux, ubah output menjadi `bin/mksiaga`.

### 4. Siapkan konfigurasi production

Gunakan nilai berikut sebagai dasar:

```env
GIN_MODE=release
SESSION_COOKIE_SECURE=true
DB_ENABLED=true
MAINTENANCE_MODE=false
```

Gunakan `SESSION_SECRET` acak khusus production, password database yang kuat, dan jangan menyalin `.env` development. Pastikan timezone sistem atau container mendukung `Asia/Jakarta`.

### 5. Salin artefak

Salin komponen berikut ke server:

- Binary aplikasi.
- Folder `web/templates`.
- Folder `web/static`.
- Folder `node_modules/@fortawesome/fontawesome-free`.
- Folder `node_modules/sweetalert2`.
- Folder `storage` jika aplikasi menyimpan unggahan.
- File `.env` production.

Jalankan binary dari root deployment agar path relatif template dan aset tetap ditemukan. Gunakan service manager seperti systemd, Supervisor, NSSM, atau Windows Service untuk restart otomatis.

### 6. Verifikasi deployment

1. Periksa `/healthz` dan `/readyz`.
2. Login memakai akun administrator.
3. Ganti password akun awal.
4. Periksa halaman User, Role, Permission, Store, Profil, dan Manajemen sesi.
5. Pastikan cookie memiliki atribut `Secure` melalui HTTPS.
6. Periksa log aplikasi dan reverse proxy.

Mode maintenance dapat diaktifkan saat deployment dengan:

```env
MAINTENANCE_MODE=true
MAINTENANCE_MESSAGE=Pembaruan sistem sedang berlangsung. Silakan coba kembali beberapa saat lagi.
```

Endpoint health check dan aset statis tetap tersedia selama mode maintenance.
