# SQL starter

`mksiaga_starter.sql` berisi schema lengkap dan data awal yang dapat langsung dipakai untuk project baru. Isinya mencakup store, akun admin development, role, permission, serta relasinya. Tabel sesi ikut dibuat, tetapi datanya dikosongkan.

## Import

Buat database kosong, lalu import file starter:

```powershell
& 'C:\xampp8.2.12\mysql\bin\mysql.exe' -u root -e "CREATE DATABASE nama_project CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
Get-Content -Raw database/mksiaga_starter.sql |
  & 'C:\xampp8.2.12\mysql\bin\mysql.exe' -u root nama_project
```

Sesuaikan `DB_NAME=nama_project` pada `.env`. Akun awal untuk development adalah `admin` dengan password `qweqwe`. Segera ganti password tersebut jika database dipakai di luar komputer development.

