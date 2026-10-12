-- Permission master kamera dan assignment kepada superadmin.
START TRANSACTION;
INSERT INTO permissions(name,guard_name,description) VALUES
('cameras.view','web','Melihat master data kamera'),
('cameras.create','web','Membuat kamera'),
('cameras.update','web','Memperbarui kamera'),
('cameras.delete','web','Menghapus kamera')
ON DUPLICATE KEY UPDATE description=VALUES(description);
INSERT IGNORE INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE r.name='superadmin' AND r.guard_name='web' AND p.name LIKE 'cameras.%';
COMMIT;
