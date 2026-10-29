INSERT INTO permissions (name,guard_name,description) VALUES
('security-shift-reports.review','web','Melihat dashboard monitoring dan memeriksa laporan shift security'),
('security-shift-reports.export','web','Mencetak dan mengekspor laporan shift security')
ON DUPLICATE KEY UPDATE description=VALUES(description);

INSERT IGNORE INTO role_permissions (role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE r.guard_name='web' AND (
  (r.name IN ('superadmin','spv_security') AND p.name='security-shift-reports.review') OR
  (r.name IN ('superadmin','spv_security','security') AND p.name='security-shift-reports.export')
);
