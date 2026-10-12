# Master operasional per store

Area dan Kamera menggunakan `store_id`. Kode unik per
store, sehingga kode yang sama boleh digunakan di store berbeda.
Shift, Checklist, Jenis Insiden, dan Tindakan menjadi referensi global.

`Middleware` memuat scope ke request context. Superadmin boleh mengelola
semua store. Pengguna lain hanya boleh mengelola store aktif pada sesi,
dengan membership aktif dan permission CRUD yang sudah ada. Assignment ke
store lain tidak otomatis memberikan permission dari store aktif ke store itu.
Repository menolak akses jika tidak menerima scope, membatasi query daftar,
pilihan form, lookup ID, update, serta delete. Form `store_id` tetap divalidasi
di server; filter di browser hanya membantu pencarian.

Perubahan Shift dan Checklist berlaku untuk seluruh store.
Store baru tidak otomatis mendapatkan master: tambahkan melalui CRUD, atau
jalankan seed 000006 secara eksplisit untuk mengisi template yang belum ada.
Seed tidak menimpa perubahan master, tetapi akan mengisi kembali kode template
yang sebelumnya dihapus. Gunakan status nonaktif untuk menonaktifkan template.

Pengujian integrasi (memakai koneksi server MySQL dari `.env`):

```powershell
$env:MKSIAGA_INTEGRATION='1'
go test ./internal/storescope -v -count=1
Remove-Item Env:MKSIAGA_INTEGRATION
```

Pengujian membuat database sementara dengan nama acak
`mksiaga_scope_test_*` dan menghapusnya setelah selesai. Database aplikasi
tidak diubah. Mencakup migrasi data, seed berulang, rendering halaman,
CRUD, kode unik per store, isolasi store pada data dan pilihan form,
permission, CSRF, serta membership nonaktif.
