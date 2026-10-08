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

## Seed development

`seeds/000001_development_superadmin.sql` membuat akun lokal `admin`, role `superadmin`, dan `Store Pusat`. Seed bersifat idempotent sehingga aman dijalankan ulang, tetapi akan mengembalikan password akun tersebut ke nilai development yang tercantum di dalam file.

```powershell
Get-Content -Raw seeds/000001_development_superadmin.sql |
  & 'C:\xampp8.2.12\mysql\bin\mysql.exe' -u root mksiaga_dev
```

Seed ini tidak boleh dijalankan pada production.

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
