-- Permission master jenis insiden dan assignment kepada superadmin.
START TRANSACTION;
INSERT INTO permissions(name,guard_name,description) VALUES
('incident-types.view','web','Melihat master jenis insiden'),
('incident-types.create','web','Membuat jenis insiden'),
('incident-types.update','web','Memperbarui jenis insiden'),
('incident-types.delete','web','Menghapus jenis insiden')
ON DUPLICATE KEY UPDATE description=VALUES(description);
INSERT IGNORE INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE r.name='superadmin' AND r.guard_name='web' AND p.name LIKE 'incident-types.%';
COMMIT;
