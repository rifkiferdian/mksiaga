-- Data demonstrasi BLMKC. Seluruh identitas dan kejadian di file ini fiktif.
-- Aman dijalankan ulang: nomor subjek dan nomor kasus demo dipertahankan,
-- sedangkan detail anak milik lima kasus demo dibangun kembali.
START TRANSACTION;

SET @creator_id = COALESCE(
    (SELECT id FROM users WHERE username='admin' AND deleted_at IS NULL LIMIT 1),
    (SELECT id FROM users WHERE deleted_at IS NULL ORDER BY id LIMIT 1)
);

-- Hentikan seed jika database sama sekali belum mempunyai user.
SET @assert_creator = IF(@creator_id IS NULL, NULL, 1);
INSERT INTO blmkc_subjects (
    subject_number,full_name,gender,address,identity_status,identification_notes,created_by_user_id,verified_by_user_id,verified_at
) VALUES
('DEMO-SUB-001',NULL,'male',NULL,'unknown','Subjek fiktif belum teridentifikasi; gunakan hanya untuk demonstrasi Wanted.',@creator_id,NULL,NULL),
('DEMO-SUB-002','Budi Contoh','male','Alamat dummy Yogyakarta','identified','Identitas fiktif untuk demonstrasi aplikasi.',@creator_id,@creator_id,'2026-10-12 10:00:00'),
('DEMO-SUB-003','Sari Simulasi','female','Alamat dummy Sleman','verified','Identitas fiktif untuk demonstrasi aplikasi.',@creator_id,@creator_id,'2026-10-13 11:00:00'),
('DEMO-SUB-004','Andi Latihan','male','Alamat dummy Bantul','verified','Identitas fiktif untuk demonstrasi aplikasi.',@creator_id,@creator_id,'2026-10-14 13:00:00'),
('DEMO-SUB-005','Rina Uji Data','female','Alamat dummy Kulon Progo','verified','Identitas fiktif untuk demonstrasi aplikasi.',@creator_id,@creator_id,'2026-10-15 15:00:00')
ON DUPLICATE KEY UPDATE
    id=LAST_INSERT_ID(id),full_name=VALUES(full_name),gender=VALUES(gender),address=VALUES(address),
    identity_status=VALUES(identity_status),identification_notes=VALUES(identification_notes),
    verified_by_user_id=VALUES(verified_by_user_id),verified_at=VALUES(verified_at),archived_at=NULL;

SET @subject_1=(SELECT id FROM blmkc_subjects WHERE subject_number='DEMO-SUB-001');
SET @subject_2=(SELECT id FROM blmkc_subjects WHERE subject_number='DEMO-SUB-002');
SET @subject_3=(SELECT id FROM blmkc_subjects WHERE subject_number='DEMO-SUB-003');
SET @subject_4=(SELECT id FROM blmkc_subjects WHERE subject_number='DEMO-SUB-004');
SET @subject_5=(SELECT id FROM blmkc_subjects WHERE subject_number='DEMO-SUB-005');

INSERT INTO blmkc_subject_aliases(subject_id,alias_name,notes) VALUES
(@subject_2,'Budi Dummy','Alias fiktif'),
(@subject_3,'Sari Demo','Alias fiktif'),
(@subject_5,'Rina Test','Alias fiktif')
ON DUPLICATE KEY UPDATE notes=VALUES(notes);

SET @store_1=(SELECT id FROM stores WHERE code='MK1' AND deleted_at IS NULL LIMIT 1);
SET @store_2=(SELECT id FROM stores WHERE code='MK2' AND deleted_at IS NULL LIMIT 1);
SET @store_3=(SELECT id FROM stores WHERE code='MK3' AND deleted_at IS NULL LIMIT 1);
SET @store_4=(SELECT id FROM stores WHERE code='MK4' AND deleted_at IS NULL LIMIT 1);
SET @store_5=(SELECT id FROM stores WHERE code='MK5' AND deleted_at IS NULL LIMIT 1);
SET @theft_type=(SELECT id FROM master_jenis_insiden WHERE code='THEFT' LIMIT 1);

INSERT INTO blmkc_cases (
    case_number,store_id,incident_type_id,area_id,occurred_at,reported_at,title,chronology,
    case_status,verification_status,legal_status,loss_amount,reported_by_user_id,reviewed_by_user_id,reviewed_at,review_notes,closed_at,archived_at
) VALUES (
    'DEMO-BLMKC-2026-001',@store_1,@theft_type,(SELECT id FROM master_area WHERE store_id=@store_1 AND code='HIGH_VALUE' LIMIT 1),
    '2026-10-10 18:25:00','2026-10-10 19:00:00','Pencurian produk kosmetik','DATA DUMMY: Rekaman memperlihatkan subjek mengambil dua produk kosmetik dan meninggalkan area tanpa pembayaran.',
    'wanted','submitted','not_reported',350000,@creator_id,NULL,NULL,NULL,NULL,NULL
) ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id),store_id=VALUES(store_id),incident_type_id=VALUES(incident_type_id),area_id=VALUES(area_id),occurred_at=VALUES(occurred_at),reported_at=VALUES(reported_at),title=VALUES(title),chronology=VALUES(chronology),case_status=VALUES(case_status),verification_status=VALUES(verification_status),legal_status=VALUES(legal_status),loss_amount=VALUES(loss_amount),reported_by_user_id=VALUES(reported_by_user_id),reviewed_by_user_id=VALUES(reviewed_by_user_id),reviewed_at=VALUES(reviewed_at),review_notes=VALUES(review_notes),closed_at=VALUES(closed_at),archived_at=NULL;
SET @case_1=LAST_INSERT_ID();

INSERT INTO blmkc_cases (
    case_number,store_id,incident_type_id,area_id,occurred_at,reported_at,title,chronology,
    case_status,verification_status,legal_status,loss_amount,reported_by_user_id,reviewed_by_user_id,reviewed_at,review_notes,closed_at,archived_at
) VALUES (
    'DEMO-BLMKC-2026-002',@store_2,@theft_type,(SELECT id FROM master_area WHERE store_id=@store_2 AND code='ENTRANCE_EXIT' LIMIT 1),
    '2026-10-11 14:10:00','2026-10-11 14:40:00','Percobaan pencurian barang toko','DATA DUMMY: Subjek mencoba membawa barang tanpa pembayaran, telah dikenali, tetapi belum berhasil diamankan.',
    'identified','revision','not_reported',225000,@creator_id,@creator_id,'2026-10-11 16:00:00','Mohon lengkapi identitas subjek dan waktu kejadian.',NULL,NULL
) ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id),store_id=VALUES(store_id),incident_type_id=VALUES(incident_type_id),area_id=VALUES(area_id),occurred_at=VALUES(occurred_at),reported_at=VALUES(reported_at),title=VALUES(title),chronology=VALUES(chronology),case_status=VALUES(case_status),verification_status=VALUES(verification_status),legal_status=VALUES(legal_status),loss_amount=VALUES(loss_amount),reported_by_user_id=VALUES(reported_by_user_id),reviewed_by_user_id=VALUES(reviewed_by_user_id),reviewed_at=VALUES(reviewed_at),review_notes=VALUES(review_notes),closed_at=VALUES(closed_at),archived_at=NULL;
SET @case_2=LAST_INSERT_ID();

INSERT INTO blmkc_cases (
    case_number,store_id,incident_type_id,area_id,occurred_at,reported_at,title,chronology,
    case_status,verification_status,legal_status,loss_amount,reported_by_user_id,reviewed_by_user_id,reviewed_at,review_notes,closed_at,archived_at
) VALUES (
    'DEMO-BLMKC-2026-003',@store_3,@theft_type,(SELECT id FROM master_area WHERE store_id=@store_3 AND code='POS' LIMIT 1),
    '2026-10-12 20:05:00','2026-10-12 20:20:00','Pengambilan aksesori dekat kasir','DATA DUMMY: Petugas mengamankan subjek setelah pemeriksaan rekaman CCTV dan pencocokan barang.',
    'apprehended','verified','not_reported',180000,@creator_id,@creator_id,'2026-10-12 21:00:00','Penanganan contoh dinyatakan lengkap.',NULL,NULL
) ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id),store_id=VALUES(store_id),incident_type_id=VALUES(incident_type_id),area_id=VALUES(area_id),occurred_at=VALUES(occurred_at),reported_at=VALUES(reported_at),title=VALUES(title),chronology=VALUES(chronology),case_status=VALUES(case_status),verification_status=VALUES(verification_status),legal_status=VALUES(legal_status),loss_amount=VALUES(loss_amount),reported_by_user_id=VALUES(reported_by_user_id),reviewed_by_user_id=VALUES(reviewed_by_user_id),reviewed_at=VALUES(reviewed_at),review_notes=VALUES(review_notes),closed_at=VALUES(closed_at),archived_at=NULL;
SET @case_3=LAST_INSERT_ID();

INSERT INTO blmkc_cases (
    case_number,store_id,incident_type_id,area_id,occurred_at,reported_at,title,chronology,
    case_status,verification_status,legal_status,loss_amount,reported_by_user_id,reviewed_by_user_id,reviewed_at,review_notes,closed_at,archived_at
) VALUES (
    'DEMO-BLMKC-2026-004',@store_4,@theft_type,(SELECT id FROM master_area WHERE store_id=@store_4 AND code='HIGH_VALUE' LIMIT 1),
    '2026-10-13 09:30:00','2026-10-13 10:00:00','Pencurian produk bernilai tinggi','DATA DUMMY: Subjek mengakui pengambilan barang dan menyelesaikan penggantian biaya melalui kesepakatan tertulis.',
    'resolved','rejected','completed',750000,@creator_id,@creator_id,'2026-10-13 13:00:00','Pengajuan ditolak karena data kejadian tidak sesuai dengan bukti yang dilampirkan.','2026-10-14 10:00:00',NULL
) ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id),store_id=VALUES(store_id),incident_type_id=VALUES(incident_type_id),area_id=VALUES(area_id),occurred_at=VALUES(occurred_at),reported_at=VALUES(reported_at),title=VALUES(title),chronology=VALUES(chronology),case_status=VALUES(case_status),verification_status=VALUES(verification_status),legal_status=VALUES(legal_status),loss_amount=VALUES(loss_amount),reported_by_user_id=VALUES(reported_by_user_id),reviewed_by_user_id=VALUES(reviewed_by_user_id),reviewed_at=VALUES(reviewed_at),review_notes=VALUES(review_notes),closed_at=VALUES(closed_at),archived_at=NULL;
SET @case_4=LAST_INSERT_ID();

INSERT INTO blmkc_cases (
    case_number,store_id,incident_type_id,area_id,occurred_at,reported_at,title,chronology,
    case_status,verification_status,legal_status,loss_amount,reported_by_user_id,reviewed_by_user_id,reviewed_at,review_notes,closed_at,archived_at
) VALUES (
    'DEMO-BLMKC-2026-005',@store_5,@theft_type,(SELECT id FROM master_area WHERE store_id=@store_5 AND code='STOCKROOM' LIMIT 1),
    '2026-10-14 16:45:00','2026-10-14 18:00:00','Pencurian stok barang gudang','DATA DUMMY: Pemeriksaan internal menemukan pengambilan stok barang tanpa dokumen pendukung.',
    'reported_to_police','verified','reported',2400000,@creator_id,@creator_id,'2026-10-15 09:00:00','Kasus contoh diteruskan untuk proses hukum.',NULL,NULL
) ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id),store_id=VALUES(store_id),incident_type_id=VALUES(incident_type_id),area_id=VALUES(area_id),occurred_at=VALUES(occurred_at),reported_at=VALUES(reported_at),title=VALUES(title),chronology=VALUES(chronology),case_status=VALUES(case_status),verification_status=VALUES(verification_status),legal_status=VALUES(legal_status),loss_amount=VALUES(loss_amount),reported_by_user_id=VALUES(reported_by_user_id),reviewed_by_user_id=VALUES(reviewed_by_user_id),reviewed_at=VALUES(reviewed_at),review_notes=VALUES(review_notes),closed_at=VALUES(closed_at),archived_at=NULL;
SET @case_5=LAST_INSERT_ID();

DELETE FROM blmkc_case_events WHERE case_id IN (@case_1,@case_2,@case_3,@case_4,@case_5);
DELETE FROM blmkc_case_resolutions WHERE case_id IN (@case_1,@case_2,@case_3,@case_4,@case_5);
DELETE FROM blmkc_vehicles WHERE case_id IN (@case_1,@case_2,@case_3,@case_4,@case_5);
DELETE FROM blmkc_case_items WHERE case_id IN (@case_1,@case_2,@case_3,@case_4,@case_5);
DELETE FROM blmkc_evidence_files WHERE case_id IN (@case_1,@case_2,@case_3,@case_4,@case_5);
DELETE FROM blmkc_case_subjects WHERE case_id IN (@case_1,@case_2,@case_3,@case_4,@case_5);

INSERT INTO blmkc_case_subjects(case_id,subject_id,involvement_role,allegation_status,is_primary,notes) VALUES
(@case_1,@subject_1,'suspected','unverified',1,'Subjek utama dummy belum dikenal.'),
(@case_2,@subject_2,'suspected','identified',1,'Subjek dummy telah teridentifikasi.'),
(@case_3,@subject_3,'perpetrator','admitted',1,'Subjek dummy telah diamankan.'),
(@case_4,@subject_4,'perpetrator','admitted',1,'Subjek dummy menyelesaikan ganti rugi.'),
(@case_5,@subject_5,'perpetrator','legally_processed',1,'Subjek dummy sedang diproses hukum.');

INSERT INTO blmkc_case_items(case_id,item_name,brand,quantity,estimated_value,item_status,notes) VALUES
(@case_1,'Produk kosmetik demo','Merek Dummy',2,350000,'missing','Barang fiktif untuk demonstrasi.'),
(@case_2,'Produk kebutuhan rumah tangga demo','Merek Dummy',3,225000,'recovered','Barang fiktif ditemukan kembali.'),
(@case_3,'Aksesori ponsel demo','Merek Dummy',2,180000,'returned','Barang fiktif dikembalikan.'),
(@case_4,'Produk elektronik demo','Merek Dummy',1,750000,'returned','Barang fiktif telah dikembalikan.'),
(@case_5,'Stok produk gudang demo','Merek Dummy',12,2400000,'missing','Selisih stok fiktif untuk demonstrasi.');

INSERT INTO blmkc_vehicles(case_id,subject_id,vehicle_type,brand,model,color,license_plate,license_plate_normalized,notes) VALUES
(@case_1,@subject_1,'Sepeda motor','Merek Dummy','Model A','Hitam','AB 0001 XX','AB0001XX','Nomor polisi fiktif.'),
(@case_2,@subject_2,'Sepeda motor','Merek Dummy','Model B','Merah','AB 0002 XX','AB0002XX','Nomor polisi fiktif.');

INSERT INTO blmkc_evidence_files(
    case_id,subject_id,file_category,original_filename,storage_path,mime_type,file_size,file_hash,
    captured_at,description,verification_status,verified_by_user_id,verified_at,verification_notes,uploaded_by_user_id,created_at
) VALUES
(@case_5,@subject_5,'incident_photo','dummy-005-incident-stockroom.png','storage/uploads/blmkc/demo/dummy-005-incident-stockroom.png','image/png',2352273,UNHEX('49b712f98632d6ba129e0bfd9df2633a72f3128ffe3e8d92498dcc3e9b94e8ed'),'2026-10-14 17:05:00','DATA DUMMY / SIMULASI: Kondisi rak gudang saat pemeriksaan selisih stok dilakukan.','verified',@creator_id,'2026-10-15 09:00:00','Aset fiktif hasil generasi untuk demonstrasi aplikasi.',@creator_id,'2026-10-14 18:05:00'),
(@case_5,@subject_5,'stolen_item_photo','dummy-005-stolen-items.png','storage/uploads/blmkc/demo/dummy-005-stolen-items.png','image/png',1959183,UNHEX('c99608a49f166c16ce4014cc7c685b5e48cd43fe9a6ef6a2bd9a166343ea5d0b'),'2026-10-14 17:15:00','DATA DUMMY / SIMULASI: Dua belas kotak produk generik yang didokumentasikan untuk pencocokan inventaris.','verified',@creator_id,'2026-10-15 09:00:00','Aset fiktif hasil generasi untuk demonstrasi aplikasi.',@creator_id,'2026-10-14 18:10:00');

INSERT INTO blmkc_case_resolutions(case_id,resolution_type,resolution_date,summary,compensation_amount,created_by_user_id,verified_by_user_id) VALUES
(@case_3,'statement','2026-10-12 21:15:00','DATA DUMMY: Subjek membuat pernyataan dan barang dikembalikan.',0,@creator_id,@creator_id),
(@case_4,'compensation','2026-10-14 10:00:00','DATA DUMMY: Biaya perbaikan fasilitas telah diganti.',750000,@creator_id,@creator_id),
(@case_5,'handed_to_police','2026-10-15 10:00:00','DATA DUMMY: Kasus diteruskan kepada kepolisian untuk simulasi workflow.',NULL,@creator_id,@creator_id);

INSERT INTO blmkc_case_events(case_id,actor_user_id,event_type,from_status,to_status,notes,created_at) VALUES
(@case_1,@creator_id,'verified','draft','wanted','Data dummy diverifikasi sebagai Wanted.','2026-10-10 20:00:00'),
(@case_2,@creator_id,'identified','wanted','identified','Identitas dummy berhasil dicatat.','2026-10-11 16:00:00'),
(@case_3,@creator_id,'apprehended','identified','apprehended','Subjek dummy ditandai telah diamankan.','2026-10-12 21:00:00'),
(@case_4,@creator_id,'resolved','apprehended','resolved','Kasus dummy selesai melalui ganti rugi.','2026-10-14 10:00:00'),
(@case_5,@creator_id,'reported_to_police','apprehended','reported_to_police','Kasus dummy diteruskan ke kepolisian.','2026-10-15 10:00:00');

COMMIT;
