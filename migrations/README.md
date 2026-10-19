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

## Migrasi 000003: master CCTV

`000003_create_cctv_masters.up.sql` menambahkan enam tabel master CCTV:
`master_area`, `master_kamera`, `master_shift`, `master_checklist`,
`master_jenis_insiden`, dan `master_tindakan`. Lokasi tetap menggunakan
`stores`; petugas dan supervisor tetap menggunakan `users`.

Jalankan migrasi ini sekali pada database yang sudah memiliki migrasi 000001
dan 000002, lanjutkan 000004 sampai 000006, kemudian jalankan `seeds/000006_cctv_masters.sql`.
Jika memakai `database/mksiaga_starter.sql` untuk instalasi baru, tetap
terapkan migrasi 000003–000006 dan seed mulai 000006 sesudahnya.

Seed mengisi tujuh template area per toko aktif dan tiga shift global. Delapan checklist
awal, enam checklist akhir, tujuh jenis insiden, dan lima tindakan bersifat global. Area akses
ruang ganti/toilet berarti area luarnya. Sesuaikan area dengan kondisi toko.
Seed dapat diulang tanpa menimpa perubahan master; toko baru mendapatkan
template area serta Shift dan Checklist global saat seed dijalankan kembali. Kamera belum diisi dan jam shift
masih NULL karena data sebenarnya belum tersedia. Shift berlaku untuk seluruh toko;
jam selesai lebih kecil dari jam mulai berarti melewati tengah malam.

Foreign key gabungan memastikan kamera hanya terhubung ke area pada toko
yang sama. Master yang sudah dipakai sebaiknya dinonaktifkan melalui `status`.
`requires_camera_count` menandai checklist yang nanti memerlukan angka kamera
aktif/total, sedangkan `requires_description` menandai pilihan Lainnya.
Penyimpanan jawaban dan validasi kolom transaksi menjadi bagian modul laporan.
Migrasi ini belum membuat tabel transaksi ataupun menu aplikasi.

## Migrasi 000004: shift dan checklist per store

`000004_scope_shift_checklist_to_store.up.sql` merupakan migrasi transisi yang menambahkan `store_id` wajib,
foreign key ke `stores`, dan unique key `(store_id, code)` pada Shift dan
Checklist. Backup database sebelum menjalankan sekali pada database lama.
MySQL DDL auto-commit; kegagalan sebagian perlu diperiksa sebelum mencoba ulang.

Prasyarat database lama: minimal satu store belum dihapus jika master sudah
terisi, serta belum ada tabel transaksi yang merujuk Shift/Checklist. Migrasi
mempertahankan ID asli pada store dengan ID terkecil dan menyalin nilai master
yang ada, termasuk perubahan pengguna, ke setiap store lain yang belum dihapus.
Jam kerja yang NULL tetap NULL. Pada instalasi baru dengan master kosong,
migrasi dapat dijalankan sebelum seed store.

Rollback ke master global tidak disediakan otomatis karena perubahan tiap
store tidak dapat digabungkan tanpa keputusan pengguna. Untuk membatalkan
migrasi, pulihkan backup sebelum migrasi beserta versi kode yang sesuai.

Setelah 000004, seed 000006 membutuhkan schema per store. Jalankan seluruh
migrasi dahulu, baru seed. Jangan menjalankan ulang seed superadmin pada
database aktif hanya untuk pembaruan master karena seed tersebut mereset password.

## Migrasi 000005: checklist global

`000005_make_checklist_global.up.sql` mengembalikan Checklist menjadi master
global karena seluruh store menggunakan daftar pemeriksaan yang sama. Migrasi
memastikan isi tiap kode masih identik, menyisakan satu salinan, lalu menghapus
`store_id`. Jika ditemukan perbedaan antar-store, migrasi berhenti sebelum
menghapus data agar perubahan dapat diselaraskan secara manual.

## Migrasi 000006: shift global

`000006_make_shift_global.up.sql` mengembalikan Shift menjadi master global.
Migrasi memastikan isi setiap kode Shift identik antar-store, menyisakan satu
salinan, lalu menghapus `store_id`. Jika ada perbedaan, migrasi berhenti sebelum
data digabung agar jadwal dapat diselaraskan secara manual.

## Migrasi 000007: laporan harian CCTV

`000007_create_daily_cctv_reports.up.sql` membuat laporan induk dan tabel
detail untuk checklist, monitoring rutin, insiden, serta tindakan insiden.
Data laporan selalu terikat pada store dan petugas aktif, sedangkan Shift,
Checklist, Jenis Insiden, dan Tindakan menggunakan master global.

`000008_add_daily_report_signatures.up.sql` menambahkan penyimpanan tanda
tangan digital petugas serah dan petugas terima pada laporan harian.

`000009_add_daily_report_workflow.up.sql` menambahkan shift default petugas,
status serah-terima dan verifikasi supervisor, waktu penerimaan/persetujuan,
catatan koreksi, serta tabel riwayat perubahan status laporan.

Aturan akses dan pengujian dijelaskan di
[`internal/storescope/README.md`](../internal/storescope/README.md).

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
