-- Requires migrations through 000006 (global shifts and checklist).
-- Template area awal untuk setiap toko aktif; sesuaikan dengan kondisi toko.
-- Aman dijalankan ulang: data yang sudah ada tidak ditimpa.
START TRANSACTION;

INSERT INTO master_area (store_id, code, name, is_priority, sort_order)
SELECT s.id, a.code, a.name, 1, a.sort_order
FROM stores s
CROSS JOIN (
    SELECT 'POS' AS code, 'Area Kasir / POS' AS name, 1 AS sort_order
    UNION ALL SELECT 'ENTRANCE_EXIT', 'Pintu Masuk & Keluar', 2
    UNION ALL SELECT 'STOCKROOM', 'Area Gudang / Stockroom', 3
    UNION ALL SELECT 'HIGH_VALUE', 'Rak Produk Bernilai Tinggi (elektronik, kosmetik, alkohol)', 4
    UNION ALL SELECT 'PARKING', 'Area Parkir', 5
    UNION ALL SELECT 'FITTING_TOILET_ACCESS', 'Area Akses Luar Ruang Ganti / Toilet', 6
    UNION ALL SELECT 'LOADING_DOCK', 'Area Belakang Toko (loading dock)', 7
) a
WHERE s.status = 'active' AND s.deleted_at IS NULL
  AND NOT EXISTS (SELECT 1 FROM master_area existing WHERE existing.store_id = s.id AND existing.code = a.code);

-- Jam shift belum diberikan. NULL berarti belum dikonfigurasi.
INSERT INTO master_shift (code,name,sort_order) VALUES
('PAGI', 'Pagi', 1), ('SIANG', 'Siang', 2), ('MALAM', 'Malam', 3)
ON DUPLICATE KEY UPDATE code = VALUES(code);

INSERT INTO master_checklist (code, phase, name, requires_camera_count, sort_order) VALUES
('START_01', 'start', 'Semua monitor CCTV berfungsi normal', 0, 1),
('START_02', 'start', 'Jumlah kamera aktif sesuai standar', 1, 2),
('START_03', 'start', 'Kualitas gambar jelas (tidak blur/terhalang)', 0, 3),
('START_04', 'start', 'Rekaman (recording) berjalan normal', 0, 4),
('START_05', 'start', 'Sistem alarm berfungsi', 0, 5),
('START_06', 'start', 'Komunikasi (HT/telepon) berfungsi', 0, 6),
('START_07', 'start', 'Buku log shift sebelumnya sudah dibaca', 0, 7),
('START_08', 'start', 'Area prioritas (kasir, gudang, pintu masuk/keluar) terpantau', 0, 8),
('END_01', 'end', 'Semua rekaman hari ini tersimpan dengan baik', 0, 1),
('END_02', 'end', 'Rekaman insiden (jika ada) sudah di-clipping', 0, 2),
('END_03', 'end', 'Laporan insiden sudah diserahkan ke supervisor', 0, 3),
('END_04', 'end', 'Peralatan CCTV dalam kondisi baik', 0, 4),
('END_05', 'end', 'Buku log sudah diisi lengkap', 0, 5),
('END_06', 'end', 'Serah terima tugas ke shift berikutnya sudah dilakukan', 0, 6)
ON DUPLICATE KEY UPDATE code = VALUES(code);

INSERT INTO master_jenis_insiden (code, name, requires_description, sort_order) VALUES
('THEFT', 'Pencurian (Shoplifting)', 0, 1),
('SUSPICIOUS', 'Kecurigaan (Suspicious Behavior)', 0, 2),
('ACCIDENT', 'Kecelakaan / Jatuh', 0, 3),
('PROPERTY_DAMAGE', 'Kerusakan Properti', 0, 4),
('CONFLICT', 'Konflik / Keributan', 0, 5),
('INTERNAL_FRAUD', 'Fraud Internal', 0, 6),
('OTHER', 'Lainnya', 1, 7)
ON DUPLICATE KEY UPDATE code = VALUES(code);

INSERT INTO master_tindakan (code, name, requires_description, sort_order) VALUES
('REPORT_SUPERVISOR', 'Melaporkan ke Supervisor', 0, 1),
('CONTACT_SECURITY', 'Menghubungi Tim LP/Security Lantai', 0, 2),
('SAVE_CLIP', 'Menyimpan Rekaman (Clipping)', 0, 3),
('CONTACT_POLICE', 'Menghubungi Kepolisian', 0, 4),
('OTHER', 'Lainnya', 1, 5)
ON DUPLICATE KEY UPDATE code = VALUES(code);

COMMIT;
