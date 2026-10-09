-- Permission modul user dan assignment kepada superadmin.
START TRANSACTION;
INSERT INTO permissions(name,guard_name,description) VALUES
('users.view','web','Melihat daftar user'),
('users.create','web','Membuat user'),
('users.update','web','Memperbarui user, store, dan role'),
('users.delete','web','Menghapus user')
ON DUPLICATE KEY UPDATE description=VALUES(description);
INSERT IGNORE INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE r.name='superadmin' AND r.guard_name='web' AND p.name LIKE 'users.%';
COMMIT;
