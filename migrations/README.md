# Migrasi MySQL

Simpan perubahan skema berurutan sebagai `000001_nama.up.sql` dan `000001_nama.down.sql`.
Gunakan database development terpisah. Aplikasi belum menjalankan migrasi secara otomatis.

## Migrasi 000001

`000001_create_identity_and_access.up.sql` membuat tabel user, store, role, dan permission untuk kebutuhan multi-store. Assignment role dan permission menggunakan `user_store_id`, sehingga satu user dapat memiliki akses dan jabatan berbeda di setiap store.

Relasi utamanya:

```text
users --< user_stores >-- stores
              |
              +--< user_store_roles >-- roles --< role_permissions >-- permissions
              |
              +--< user_store_permissions >-------------------------- permissions
```

`user_store_permissions` dipakai untuk permission langsung sebagai pengecualian di luar role. Nilai role seperti `security`, `admin`, dan permission aplikasi dibuat melalui seed atau modul administrasi pada tahap berikutnya.

## Migrasi 000002

`000002_create_user_sessions.up.sql` menambahkan pencatatan sesi login per perangkat. Data ini dipakai untuk menampilkan waktu login, IP, browser/perangkat, batas waktu sesi, dan mengeluarkan perangkat lain dari halaman **Profil > Perangkat & sesi**.

```powershell
Get-Content -Raw migrations/000002_create_user_sessions.up.sql |
  & 'C:\xampp8.2.12\mysql\bin\mysql.exe' -u root mksiaga_dev
```

Setelah migrasi pertama kali diterapkan, sesi lama perlu login ulang agar tercatat sebagai perangkat aktif.

## Seed development

`seeds/000001_development_superadmin.sql` membuat akun lokal `admin`, role `superadmin`, dan `Store Pusat`. Seed bersifat idempotent sehingga aman dijalankan ulang, tetapi akan mengembalikan password akun tersebut ke nilai development yang tercantum di dalam file.

```powershell
Get-Content -Raw seeds/000001_development_superadmin.sql |
  & 'C:\xampp8.2.12\mysql\bin\mysql.exe' -u root mksiaga_dev
```

Seed ini tidak boleh dijalankan pada production.

`seeds/000002_access_management_permissions.sql` menambahkan permission awal untuk CRUD role dan permission, kemudian memberikannya kepada role `superadmin`.

`seeds/000003_user_management_permissions.sql` menambahkan permission CRUD user dan memberikannya kepada role `superadmin`.

`seeds/000004_store_management_permissions.sql` menambahkan permission CRUD master data store dan memberikannya kepada role `superadmin`.

`seeds/000005_master_stores.sql` mengisi master Store MK1–MK7 dan MK Mini 1–3 dengan ID yang sama seperti data sumber.

Menjalankan migrasi secara manual dari PowerShell/XAMPP:

```powershell
Get-Content -Raw migrations/000001_create_identity_and_access.up.sql |
  & 'C:\xampp8.2.12\mysql\bin\mysql.exe' -u root mksiaga_dev
```

Rollback akan menghapus seluruh tabel dan datanya, sehingga hanya jalankan pada database development yang memang boleh dikosongkan:

```powershell
Get-Content -Raw migrations/000001_create_identity_and_access.down.sql |
  & 'C:\xampp8.2.12\mysql\bin\mysql.exe' -u root mksiaga_dev
```
