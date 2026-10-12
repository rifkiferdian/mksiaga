-- Permission master area dan assignment kepada superadmin.
START TRANSACTION;
INSERT INTO permissions(name,guard_name,description) VALUES
('areas.view','web','Melihat master data area'),
('areas.create','web','Membuat area'),
('areas.update','web','Memperbarui area'),
('areas.delete','web','Menghapus area')
ON DUPLICATE KEY UPDATE description=VALUES(description);
INSERT IGNORE INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE r.name='superadmin' AND r.guard_name='web' AND p.name LIKE 'areas.%';
COMMIT;
