INSERT INTO permissions (name,guard_name,description) VALUES
('blmkc.view','web','Melihat dashboard dan daftar kasus BLMKC'),
('blmkc.view-sensitive','web','Melihat data identitas sensitif BLMKC'),
('blmkc.create','web','Membuat kasus BLMKC'),
('blmkc.update','web','Memperbarui kasus BLMKC'),
('blmkc.submit','web','Mengajukan kasus BLMKC untuk diverifikasi'),
('blmkc.review','web','Memverifikasi kasus BLMKC'),
('blmkc.change-status','web','Mengubah status penanganan kasus BLMKC'),
('blmkc.upload-evidence','web','Mengunggah bukti kasus BLMKC'),
('blmkc.download-evidence','web','Mengunduh bukti kasus BLMKC'),
('blmkc.resolve','web','Mencatat penyelesaian kasus BLMKC'),
('blmkc.export','web','Mengekspor data BLMKC'),
('blmkc.archive','web','Mengarsipkan data BLMKC'),
('blmkc.audit','web','Melihat audit akses BLMKC')
ON DUPLICATE KEY UPDATE description=VALUES(description);

INSERT IGNORE INTO role_permissions (role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE r.name='superadmin' AND r.guard_name='web' AND p.name LIKE 'blmkc.%';
