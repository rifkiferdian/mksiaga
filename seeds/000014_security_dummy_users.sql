-- HANYA UNTUK DEVELOPMENT.
-- Akun security_dummy_1 dan security_dummy_2 menggunakan password: qweqwe
START TRANSACTION;

INSERT INTO roles(name,guard_name,description)
VALUES ('security','web','Petugas Security/CCTV per shift')
ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id),description=VALUES(description);
SET @security_role_id=LAST_INSERT_ID();

INSERT IGNORE INTO role_permissions(role_id,permission_id)
SELECT @security_role_id,id FROM permissions
WHERE name IN ('daily-reports.view','daily-reports.create');

-- Gunakan Store Pusat bila tersedia; pada data operasional gunakan store aktif pertama.
SELECT id INTO @store_id FROM stores
WHERE status='active' AND deleted_at IS NULL
ORDER BY (code='PUSAT') DESC,id
LIMIT 1;
SELECT id INTO @shift_pagi_id FROM master_shift WHERE code='PAGI' LIMIT 1;
SELECT id INTO @shift_siang_id FROM master_shift WHERE code='SIANG' LIMIT 1;

INSERT INTO users(employee_number,name,username,email,password_hash,status)
VALUES ('SEC-DUMMY-001','Security Dummy 1 - Shift Pagi','security_dummy_1','security1@mksiaga.local','$2y$10$lwkQLoTL4oDoSPj6oGHuL.bhCFZzZE3hfwul5dw6f738FrThFhov6','active')
ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id),name=VALUES(name),email=VALUES(email),password_hash=VALUES(password_hash),status='active',deleted_at=NULL;
SET @user_pagi_id=LAST_INSERT_ID();
INSERT INTO user_stores(user_id,store_id,default_shift_id,is_default,status,joined_at)
VALUES (@user_pagi_id,@store_id,@shift_pagi_id,1,'active',CURRENT_TIMESTAMP(6))
ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id),default_shift_id=VALUES(default_shift_id),is_default=1,status='active';
SET @user_store_pagi_id=LAST_INSERT_ID();
INSERT IGNORE INTO user_store_roles(user_store_id,role_id) VALUES (@user_store_pagi_id,@security_role_id);

INSERT INTO users(employee_number,name,username,email,password_hash,status)
VALUES ('SEC-DUMMY-002','Security Dummy 2 - Shift Siang','security_dummy_2','security2@mksiaga.local','$2y$10$lwkQLoTL4oDoSPj6oGHuL.bhCFZzZE3hfwul5dw6f738FrThFhov6','active')
ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id),name=VALUES(name),email=VALUES(email),password_hash=VALUES(password_hash),status='active',deleted_at=NULL;
SET @user_siang_id=LAST_INSERT_ID();
INSERT INTO user_stores(user_id,store_id,default_shift_id,is_default,status,joined_at)
VALUES (@user_siang_id,@store_id,@shift_siang_id,1,'active',CURRENT_TIMESTAMP(6))
ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id),default_shift_id=VALUES(default_shift_id),is_default=1,status='active';
SET @user_store_siang_id=LAST_INSERT_ID();
INSERT IGNORE INTO user_store_roles(user_store_id,role_id) VALUES (@user_store_siang_id,@security_role_id);

COMMIT;
