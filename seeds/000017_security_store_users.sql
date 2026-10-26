-- HANYA UNTUK DEVELOPMENT / INITIAL PROVISIONING.
-- Username memakai NIP dan password awal seluruh akun: qweqwe
START TRANSACTION;

INSERT INTO roles(name,guard_name,description)
VALUES ('security','web','Petugas Security/CCTV per shift')
ON DUPLICATE KEY UPDATE description=VALUES(description);

INSERT IGNORE INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE r.name='security' AND r.guard_name='web' AND p.name IN (
    'daily-reports.view','daily-reports.create',
    'blmkc.view','blmkc.view-sensitive','blmkc.create','blmkc.update',
    'blmkc.change-status','blmkc.upload-evidence','blmkc.download-evidence','blmkc.resolve'
);

INSERT INTO users(employee_number,name,username,email,password_hash,status) VALUES
('190136','Ginanjar Nur Rohman','190136','190136@mksiaga.local','$2y$10$lwkQLoTL4oDoSPj6oGHuL.bhCFZzZE3hfwul5dw6f738FrThFhov6','active'),
('260093','Dian Safitri','260093','260093@mksiaga.local','$2y$10$lwkQLoTL4oDoSPj6oGHuL.bhCFZzZE3hfwul5dw6f738FrThFhov6','active'),
('250014','Dera Kurnia Sandy','250014','250014@mksiaga.local','$2y$10$lwkQLoTL4oDoSPj6oGHuL.bhCFZzZE3hfwul5dw6f738FrThFhov6','active'),
('060006','Dwi Indarsanto','060006','060006@mksiaga.local','$2y$10$lwkQLoTL4oDoSPj6oGHuL.bhCFZzZE3hfwul5dw6f738FrThFhov6','active'),
('260169','Puji Santoso','260169','260169@mksiaga.local','$2y$10$lwkQLoTL4oDoSPj6oGHuL.bhCFZzZE3hfwul5dw6f738FrThFhov6','active'),
('100037','Irfan Setyo Putro','100037','100037@mksiaga.local','$2y$10$lwkQLoTL4oDoSPj6oGHuL.bhCFZzZE3hfwul5dw6f738FrThFhov6','active'),
('250111','Faqih Alfian Darmawan','250111','250111@mksiaga.local','$2y$10$lwkQLoTL4oDoSPj6oGHuL.bhCFZzZE3hfwul5dw6f738FrThFhov6','active'),
('230188','Ragil Saputro','230188','230188@mksiaga.local','$2y$10$lwkQLoTL4oDoSPj6oGHuL.bhCFZzZE3hfwul5dw6f738FrThFhov6','active'),
('260205','Taufan Bayu Kriswanto','260205','260205@mksiaga.local','$2y$10$lwkQLoTL4oDoSPj6oGHuL.bhCFZzZE3hfwul5dw6f738FrThFhov6','active')
ON DUPLICATE KEY UPDATE
    employee_number=VALUES(employee_number),name=VALUES(name),username=VALUES(username),
    email=VALUES(email),password_hash=VALUES(password_hash),status='active',deleted_at=NULL;

INSERT INTO user_stores(user_id,store_id,is_default,status,joined_at)
SELECT u.id,s.id,1,'active',CURRENT_TIMESTAMP(6)
FROM users u
JOIN stores s ON s.code=CASE u.username
    WHEN '190136' THEN 'MK3' WHEN '260093' THEN 'MK3'
    WHEN '250014' THEN 'MK5' WHEN '060006' THEN 'MK5' WHEN '260169' THEN 'MK5'
    WHEN '100037' THEN 'MK6' WHEN '250111' THEN 'MK6'
    WHEN '230188' THEN 'MK7' WHEN '260205' THEN 'MK7'
END
WHERE u.username IN ('190136','260093','250014','060006','260169','100037','250111','230188','260205')
  AND s.deleted_at IS NULL
ON DUPLICATE KEY UPDATE is_default=1,status='active';

INSERT IGNORE INTO user_store_roles(user_store_id,role_id)
SELECT us.id,r.id
FROM user_stores us
JOIN users u ON u.id=us.user_id
JOIN roles r ON r.name='security' AND r.guard_name='web'
WHERE u.username IN ('190136','260093','250014','060006','260169','100037','250111','230188','260205');

COMMIT;
