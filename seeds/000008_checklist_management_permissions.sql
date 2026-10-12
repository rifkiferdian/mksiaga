-- Permission master checklist dan assignment kepada superadmin.
START TRANSACTION;
INSERT INTO permissions(name,guard_name,description) VALUES
('checklists.view','web','Melihat master data checklist'),
('checklists.create','web','Membuat checklist'),
('checklists.update','web','Memperbarui checklist'),
('checklists.delete','web','Menghapus checklist')
ON DUPLICATE KEY UPDATE description=VALUES(description);
INSERT IGNORE INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE r.name='superadmin' AND r.guard_name='web' AND p.name LIKE 'checklists.%';
COMMIT;
