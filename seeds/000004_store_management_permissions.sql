-- Permission master data store dan assignment kepada superadmin.
START TRANSACTION;
INSERT INTO permissions(name,guard_name,description) VALUES
('stores.view','web','Melihat master data store'),
('stores.create','web','Membuat store'),
('stores.update','web','Memperbarui store'),
('stores.delete','web','Menghapus store')
ON DUPLICATE KEY UPDATE description=VALUES(description);
INSERT IGNORE INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE r.name='superadmin' AND r.guard_name='web' AND p.name LIKE 'stores.%';
COMMIT;
