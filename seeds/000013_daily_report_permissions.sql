INSERT INTO permissions (name,guard_name,description) VALUES
('daily-reports.view','web','Melihat laporan harian CCTV'),
('daily-reports.create','web','Membuat laporan harian CCTV'),
('daily-reports.review','web','Memverifikasi dan menyetujui laporan harian CCTV'),
('daily-reports.delete','web','Menghapus laporan harian CCTV')
ON DUPLICATE KEY UPDATE description=VALUES(description);

INSERT IGNORE INTO role_permissions (role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE r.name='superadmin' AND r.guard_name='web' AND p.name LIKE 'daily-reports.%';
