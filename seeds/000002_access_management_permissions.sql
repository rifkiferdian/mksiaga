-- Permission awal untuk modul role dan permission.
-- Seed aman dijalankan ulang dan otomatis diberikan kepada superadmin.

START TRANSACTION;

INSERT INTO permissions (name, guard_name, description) VALUES
('roles.view', 'web', 'Melihat daftar role'),
('roles.create', 'web', 'Membuat role'),
('roles.update', 'web', 'Memperbarui role dan assignment permission'),
('roles.delete', 'web', 'Menghapus role'),
('permissions.view', 'web', 'Melihat daftar permission'),
('permissions.create', 'web', 'Membuat permission'),
('permissions.update', 'web', 'Memperbarui permission'),
('permissions.delete', 'web', 'Menghapus permission')
ON DUPLICATE KEY UPDATE description = VALUES(description);

INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'superadmin' AND r.guard_name = 'web';

COMMIT;
