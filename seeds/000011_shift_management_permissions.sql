-- Permission master shift dan assignment kepada superadmin.
START TRANSACTION;
INSERT INTO permissions(name,guard_name,description) VALUES
('shifts.view','web','Melihat master data shift'),
('shifts.create','web','Membuat shift'),
('shifts.update','web','Memperbarui shift'),
('shifts.delete','web','Menghapus shift')
ON DUPLICATE KEY UPDATE description=VALUES(description);
INSERT IGNORE INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE r.name='superadmin' AND r.guard_name='web' AND p.name LIKE 'shifts.%';
COMMIT;
