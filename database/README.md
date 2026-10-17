# Backup SQL

## Salinan database saat ini

`mksiaga_full.sql` adalah hasil export lengkap database `mksiaga_dev` pada 17 Oktober 2026. File ini berisi schema terbaru dan data aplikasi saat ini. Struktur tabel `user_sessions` disertakan, tetapi data sesinya sengaja dikosongkan agar sesi login dari PC lama tidak ikut dipindahkan.

Untuk memakainya pada PC lain, nyalakan MySQL/MariaDB lalu jalankan dari root project:

```powershell
& 'C:\xampp8.2.12\mysql\bin\mysql.exe' -u root -e "CREATE DATABASE mksiaga_dev CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

Get-Content -Raw database/mksiaga_full.sql |
  & 'C:\xampp8.2.12\mysql\bin\mysql.exe' -u root mksiaga_dev
```

Jika user MySQL memiliki password, tambahkan opsi `-p` pada kedua perintah. Setelah itu salin `.env.example` menjadi `.env`, sesuaikan konfigurasi database, dan buat `SESSION_SECRET` baru. Folder `storage/uploads` bukan bagian dari SQL; salin folder tersebut secara terpisah jika berisi file yang juga dibutuhkan.

Jangan import file ini ke database yang sudah berisi data karena tabel yang namanya sama akan diganti.

## SQL starter

`mksiaga_starter.sql` berisi schema lengkap dan data awal yang dapat langsung dipakai untuk project baru. Isinya mencakup store, akun admin development, role, permission, serta relasinya. Tabel sesi ikut dibuat, tetapi datanya dikosongkan.

## Import

Buat database kosong, lalu import file starter:

```powershell
& 'C:\xampp8.2.12\mysql\bin\mysql.exe' -u root -e "CREATE DATABASE nama_project CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
Get-Content -Raw database/mksiaga_starter.sql |
  & 'C:\xampp8.2.12\mysql\bin\mysql.exe' -u root nama_project
```

Sesuaikan `DB_NAME=nama_project` pada `.env`. Akun awal untuk development adalah `admin` dengan password `qweqwe`. Segera ganti password tersebut jika database dipakai di luar komputer development.

