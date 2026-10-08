-- HANYA UNTUK DEVELOPMENT.
-- Akun: admin / qweqwe
-- Password disimpan sebagai bcrypt hash, bukan teks biasa.

START TRANSACTION;

INSERT INTO stores (code, name, timezone, status)
VALUES ('PUSAT', 'Store Pusat', 'Asia/Jakarta', 'active')
ON DUPLICATE KEY UPDATE
    id = LAST_INSERT_ID(id),
    name = VALUES(name),
    timezone = VALUES(timezone),
    status = VALUES(status),
    deleted_at = NULL;
SET @store_id = LAST_INSERT_ID();

INSERT INTO roles (name, guard_name, description)
VALUES ('superadmin', 'web', 'Akses penuh untuk administrasi aplikasi')
ON DUPLICATE KEY UPDATE
    id = LAST_INSERT_ID(id),
    description = VALUES(description);
SET @role_id = LAST_INSERT_ID();

INSERT INTO users (name, username, email, password_hash, status)
VALUES (
    'Super Administrator',
    'admin',
    'admin@mksiaga.local',
    '$2y$10$lwkQLoTL4oDoSPj6oGHuL.bhCFZzZE3hfwul5dw6f738FrThFhov6',
    'active'
)
ON DUPLICATE KEY UPDATE
    id = LAST_INSERT_ID(id),
    name = VALUES(name),
    email = VALUES(email),
    password_hash = VALUES(password_hash),
    status = VALUES(status),
    deleted_at = NULL;
SET @user_id = LAST_INSERT_ID();

INSERT INTO user_stores (user_id, store_id, is_default, status, joined_at)
VALUES (@user_id, @store_id, 1, 'active', CURRENT_TIMESTAMP(6))
ON DUPLICATE KEY UPDATE
    id = LAST_INSERT_ID(id),
    is_default = VALUES(is_default),
    status = VALUES(status);
SET @user_store_id = LAST_INSERT_ID();

INSERT IGNORE INTO user_store_roles (user_store_id, role_id)
VALUES (@user_store_id, @role_id);

COMMIT;
