-- Permission master tindakan dan assignment kepada superadmin.
START TRANSACTION;
INSERT INTO permissions(name,guard_name,description) VALUES
('actions.view','web','Melihat master data tindakan'),
('actions.create','web','Membuat tindakan'),
('actions.update','web','Memperbarui tindakan'),
('actions.delete','web','Menghapus tindakan')
ON DUPLICATE KEY UPDATE description=VALUES(description);
INSERT IGNORE INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE r.name='superadmin' AND r.guard_name='web' AND p.name LIKE 'actions.%';
COMMIT;
