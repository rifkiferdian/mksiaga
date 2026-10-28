INSERT INTO permissions (name,guard_name,description) VALUES
('security-shift-reports.view','web','Melihat laporan shift security'),
('security-shift-reports.create','web','Membuat laporan shift security')
ON DUPLICATE KEY UPDATE description=VALUES(description);

INSERT IGNORE INTO role_permissions (role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE r.guard_name='web'
  AND r.name IN ('superadmin','security','spv_security')
  AND p.name IN ('security-shift-reports.view','security-shift-reports.create');
