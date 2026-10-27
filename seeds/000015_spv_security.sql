-- HANYA UNTUK DEVELOPMENT.
-- Akun SPV Security menggunakan password: qweqwe
START TRANSACTION;

INSERT INTO roles(name,guard_name,description)
VALUES ('spv_security','web','Supervisor Security yang memverifikasi laporan staf')
ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id),description=VALUES(description);
SET @spv_security_role_id=LAST_INSERT_ID();

INSERT IGNORE INTO role_permissions(role_id,permission_id)
SELECT @spv_security_role_id,id FROM permissions
WHERE name IN (
    'daily-reports.view','daily-reports.review',
    'blmkc.view','blmkc.view-sensitive','blmkc.review','blmkc.download-evidence'
);

SELECT id INTO @store_id FROM stores
WHERE status='active' AND deleted_at IS NULL
ORDER BY (code='MK1') DESC,(code='PUSAT') DESC,id
LIMIT 1;

INSERT INTO users(employee_number,name,username,email,password_hash,status)
VALUES ('SPV-SEC-001','SPV Security','spv_security','spv.security@mksiaga.local','$2y$10$lwkQLoTL4oDoSPj6oGHuL.bhCFZzZE3hfwul5dw6f738FrThFhov6','active')
ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id),name=VALUES(name),email=VALUES(email),password_hash=VALUES(password_hash),status='active',deleted_at=NULL;
SET @spv_security_user_id=LAST_INSERT_ID();

INSERT INTO user_stores(user_id,store_id,is_default,status,joined_at)
VALUES (@spv_security_user_id,@store_id,1,'active',CURRENT_TIMESTAMP(6))
ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id),is_default=1,status='active';
SET @spv_security_user_store_id=LAST_INSERT_ID();

INSERT IGNORE INTO user_store_roles(user_store_id,role_id)
VALUES (@spv_security_user_store_id,@spv_security_role_id);

COMMIT;
