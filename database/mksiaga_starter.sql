-- MK Siaga - schema dan data awal siap import
-- Dibuat untuk bootstrap project baru.
-- Akun development awal: admin / qweqwe
-- Ganti password segera setelah import pada lingkungan selain development.
-- Tabel user_sessions sengaja dibuat tanpa data agar sesi login dimulai bersih.

SET NAMES utf8mb4;
SET time_zone = '+07:00';

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;
DROP TABLE IF EXISTS `permissions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `permissions` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(125) NOT NULL,
  `guard_name` varchar(50) NOT NULL DEFAULT 'web',
  `description` varchar(255) DEFAULT NULL,
  `created_at` datetime(6) NOT NULL DEFAULT current_timestamp(6),
  `updated_at` datetime(6) NOT NULL DEFAULT current_timestamp(6) ON UPDATE current_timestamp(6),
  PRIMARY KEY (`id`),
  UNIQUE KEY `permissions_name_guard_unique` (`name`,`guard_name`)
) ENGINE=InnoDB AUTO_INCREMENT=27 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `role_permissions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `role_permissions` (
  `role_id` bigint(20) unsigned NOT NULL,
  `permission_id` bigint(20) unsigned NOT NULL,
  `created_at` datetime(6) NOT NULL DEFAULT current_timestamp(6),
  PRIMARY KEY (`role_id`,`permission_id`),
  KEY `role_permissions_permission_id_index` (`permission_id`),
  CONSTRAINT `role_permissions_permission_fk` FOREIGN KEY (`permission_id`) REFERENCES `permissions` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `role_permissions_role_fk` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `roles`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `roles` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(125) NOT NULL,
  `guard_name` varchar(50) NOT NULL DEFAULT 'web',
  `description` varchar(255) DEFAULT NULL,
  `created_at` datetime(6) NOT NULL DEFAULT current_timestamp(6),
  `updated_at` datetime(6) NOT NULL DEFAULT current_timestamp(6) ON UPDATE current_timestamp(6),
  PRIMARY KEY (`id`),
  UNIQUE KEY `roles_name_guard_unique` (`name`,`guard_name`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `stores`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `stores` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `code` varchar(50) NOT NULL,
  `name` varchar(150) NOT NULL,
  `address` text DEFAULT NULL,
  `phone` varchar(30) DEFAULT NULL,
  `timezone` varchar(64) NOT NULL DEFAULT 'Asia/Jakarta',
  `status` varchar(20) NOT NULL DEFAULT 'active',
  `created_at` datetime(6) NOT NULL DEFAULT current_timestamp(6),
  `updated_at` datetime(6) NOT NULL DEFAULT current_timestamp(6) ON UPDATE current_timestamp(6),
  `deleted_at` datetime(6) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `stores_code_unique` (`code`),
  KEY `stores_status_index` (`status`),
  KEY `stores_deleted_at_index` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=104 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `user_sessions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `user_sessions` (
  `id` char(64) NOT NULL,
  `user_id` bigint(20) unsigned NOT NULL,
  `user_store_id` bigint(20) unsigned NOT NULL,
  `ip_address` varchar(45) NOT NULL,
  `user_agent` varchar(512) NOT NULL,
  `created_at` datetime(6) NOT NULL DEFAULT current_timestamp(6),
  `last_seen_at` datetime(6) NOT NULL DEFAULT current_timestamp(6),
  `expires_at` datetime(6) NOT NULL,
  `revoked_at` datetime(6) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `user_sessions_user_active_index` (`user_id`,`revoked_at`,`expires_at`),
  KEY `user_sessions_user_store_fk` (`user_store_id`),
  CONSTRAINT `user_sessions_user_fk` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `user_sessions_user_store_fk` FOREIGN KEY (`user_store_id`) REFERENCES `user_stores` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `user_store_permissions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `user_store_permissions` (
  `user_store_id` bigint(20) unsigned NOT NULL,
  `permission_id` bigint(20) unsigned NOT NULL,
  `created_at` datetime(6) NOT NULL DEFAULT current_timestamp(6),
  PRIMARY KEY (`user_store_id`,`permission_id`),
  KEY `user_store_permissions_permission_id_index` (`permission_id`),
  CONSTRAINT `user_store_permissions_permission_fk` FOREIGN KEY (`permission_id`) REFERENCES `permissions` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `user_store_permissions_user_store_fk` FOREIGN KEY (`user_store_id`) REFERENCES `user_stores` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `user_store_roles`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `user_store_roles` (
  `user_store_id` bigint(20) unsigned NOT NULL,
  `role_id` bigint(20) unsigned NOT NULL,
  `created_at` datetime(6) NOT NULL DEFAULT current_timestamp(6),
  PRIMARY KEY (`user_store_id`,`role_id`),
  KEY `user_store_roles_role_id_index` (`role_id`),
  CONSTRAINT `user_store_roles_role_fk` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `user_store_roles_user_store_fk` FOREIGN KEY (`user_store_id`) REFERENCES `user_stores` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `user_stores`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `user_stores` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) unsigned NOT NULL,
  `store_id` bigint(20) unsigned NOT NULL,
  `is_default` tinyint(1) NOT NULL DEFAULT 0,
  `status` varchar(20) NOT NULL DEFAULT 'active',
  `joined_at` datetime(6) DEFAULT NULL,
  `created_at` datetime(6) NOT NULL DEFAULT current_timestamp(6),
  `updated_at` datetime(6) NOT NULL DEFAULT current_timestamp(6) ON UPDATE current_timestamp(6),
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_stores_user_store_unique` (`user_id`,`store_id`),
  KEY `user_stores_store_status_index` (`store_id`,`status`),
  CONSTRAINT `user_stores_store_fk` FOREIGN KEY (`store_id`) REFERENCES `stores` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `user_stores_user_fk` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=13 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `users` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `employee_number` varchar(50) DEFAULT NULL,
  `name` varchar(150) NOT NULL,
  `username` varchar(100) NOT NULL,
  `email` varchar(191) DEFAULT NULL,
  `password_hash` varchar(255) NOT NULL,
  `phone` varchar(30) DEFAULT NULL,
  `status` varchar(20) NOT NULL DEFAULT 'active',
  `last_login_at` datetime(6) DEFAULT NULL,
  `created_at` datetime(6) NOT NULL DEFAULT current_timestamp(6),
  `updated_at` datetime(6) NOT NULL DEFAULT current_timestamp(6) ON UPDATE current_timestamp(6),
  `deleted_at` datetime(6) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `users_username_unique` (`username`),
  UNIQUE KEY `users_employee_number_unique` (`employee_number`),
  UNIQUE KEY `users_email_unique` (`email`),
  KEY `users_status_index` (`status`),
  KEY `users_deleted_at_index` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;



/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

LOCK TABLES `permissions` WRITE;
/*!40000 ALTER TABLE `permissions` DISABLE KEYS */;
INSERT INTO `permissions` VALUES (2,'roles.view','web','Melihat daftar role','2026-10-08 13:43:34.934316','2026-10-08 13:43:34.934316'),(3,'roles.create','web','Membuat role','2026-10-08 13:43:34.934316','2026-10-08 13:43:34.934316'),(4,'roles.update','web','Memperbarui role dan assignment permission','2026-10-08 13:43:34.934316','2026-10-08 13:43:34.934316'),(5,'roles.delete','web','Menghapus role','2026-10-08 13:43:34.934316','2026-10-08 13:43:34.934316'),(6,'permissions.view','web','Melihat daftar permission','2026-10-08 13:43:34.934316','2026-10-08 13:43:34.934316'),(7,'permissions.create','web','Membuat permission','2026-10-08 13:43:34.934316','2026-10-08 13:43:34.934316'),(8,'permissions.update','web','Memperbarui permission','2026-10-08 13:43:34.934316','2026-10-08 13:43:34.934316'),(9,'permissions.delete','web','Menghapus permission','2026-10-08 13:43:34.934316','2026-10-08 13:43:34.934316'),(19,'users.view','web','Melihat daftar user','2026-10-10 11:12:47.489454','2026-10-10 11:12:47.489454'),(20,'users.create','web','Membuat user','2026-10-10 11:12:47.489454','2026-10-10 11:12:47.489454'),(21,'users.update','web','Memperbarui user, store, dan role','2026-10-10 11:12:47.489454','2026-10-10 11:12:47.489454'),(22,'users.delete','web','Menghapus user','2026-10-10 11:12:47.489454','2026-10-10 11:12:47.489454'),(23,'stores.view','web','Melihat master data store','2026-10-10 11:12:47.510954','2026-10-10 11:12:47.510954'),(24,'stores.create','web','Membuat store','2026-10-10 11:12:47.510954','2026-10-10 11:12:47.510954'),(25,'stores.update','web','Memperbarui store','2026-10-10 11:12:47.510954','2026-10-10 11:12:47.510954'),(26,'stores.delete','web','Menghapus store','2026-10-10 11:12:47.510954','2026-10-10 11:12:47.510954');
/*!40000 ALTER TABLE `permissions` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `roles` WRITE;
/*!40000 ALTER TABLE `roles` DISABLE KEYS */;
INSERT INTO `roles` VALUES (2,'superadmin','web','Akses penuh untuk administrasi aplikasi','2026-10-08 12:43:35.508434','2026-10-08 12:43:35.508434');
/*!40000 ALTER TABLE `roles` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `stores` WRITE;
/*!40000 ALTER TABLE `stores` DISABLE KEYS */;
INSERT INTO `stores` VALUES (1,'MK1','MK1 Babarsari','Babarsari',NULL,'Asia/Jakarta','active','2026-10-09 14:42:19.608203','2026-10-09 14:42:19.608203',NULL),(2,'MK2','MK2 Simanjuntak','Simanjuntak',NULL,'Asia/Jakarta','active','2026-10-09 14:42:19.608203','2026-10-09 14:42:19.608203',NULL),(3,'MK3','MK3 Supeno','Supeno',NULL,'Asia/Jakarta','active','2026-10-08 12:43:35.506078','2026-10-09 14:42:19.608203',NULL),(4,'MK4','MK4 Palagan','Palagan',NULL,'Asia/Jakarta','active','2026-10-09 14:42:19.608203','2026-10-09 14:42:19.608203',NULL),(5,'MK5','MK5 Godean','Godean',NULL,'Asia/Jakarta','active','2026-10-09 14:42:19.608203','2026-10-09 14:42:19.608203',NULL),(6,'MK6','MK6 Imogiri','Imogiri',NULL,'Asia/Jakarta','active','2026-10-09 14:42:19.608203','2026-10-09 14:42:19.608203',NULL),(7,'MK7','MK7 Keloran','Keloran',NULL,'Asia/Jakarta','active','2026-10-09 14:42:19.608203','2026-10-09 14:42:19.608203',NULL),(101,'MKM1','MK Mini 1 Plemesewu','Plemesewu',NULL,'Asia/Jakarta','active','2026-10-09 14:42:19.608203','2026-10-09 14:42:19.608203',NULL),(102,'MKM2','MK Mini 2 Diro','Diro',NULL,'Asia/Jakarta','active','2026-10-09 14:42:19.608203','2026-10-09 14:42:19.608203',NULL),(103,'MKM3','MK Mini 3 Minomartani','Minomartani',NULL,'Asia/Jakarta','active','2026-10-09 14:42:19.608203','2026-10-09 14:42:19.608203',NULL);
/*!40000 ALTER TABLE `stores` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `users` WRITE;
/*!40000 ALTER TABLE `users` DISABLE KEYS */;
INSERT INTO `users` VALUES (2,NULL,'Admin Security','admin','admin@mksiaga.local','$2y$10$lwkQLoTL4oDoSPj6oGHuL.bhCFZzZE3hfwul5dw6f738FrThFhov6',NULL,'active','2026-10-10 11:10:54.889916','2026-10-08 12:43:35.508585','2026-10-10 11:10:54.889916',NULL);
/*!40000 ALTER TABLE `users` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `role_permissions` WRITE;
/*!40000 ALTER TABLE `role_permissions` DISABLE KEYS */;
INSERT INTO `role_permissions` VALUES (2,2,'2026-10-08 13:43:34.935167'),(2,3,'2026-10-08 13:43:34.935167'),(2,4,'2026-10-08 13:43:34.935167'),(2,5,'2026-10-08 13:43:34.935167'),(2,6,'2026-10-08 13:43:34.935167'),(2,7,'2026-10-08 13:43:34.935167'),(2,8,'2026-10-08 13:43:34.935167'),(2,9,'2026-10-08 13:43:34.935167'),(2,19,'2026-10-10 11:12:47.489710'),(2,20,'2026-10-10 11:12:47.489710'),(2,21,'2026-10-10 11:12:47.489710'),(2,22,'2026-10-10 11:12:47.489710'),(2,23,'2026-10-10 11:12:47.511209'),(2,24,'2026-10-10 11:12:47.511209'),(2,25,'2026-10-10 11:12:47.511209'),(2,26,'2026-10-10 11:12:47.511209');
/*!40000 ALTER TABLE `role_permissions` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `user_stores` WRITE;
/*!40000 ALTER TABLE `user_stores` DISABLE KEYS */;
INSERT INTO `user_stores` VALUES (3,2,3,0,'active','2026-10-08 12:43:35.508773','2026-10-08 12:43:35.508773','2026-10-09 14:54:26.405814'),(6,2,5,1,'active','2026-10-09 14:52:24.169030','2026-10-09 14:52:24.169030','2026-10-09 14:54:26.408630'),(11,2,4,0,'active','2026-10-09 14:54:26.408077','2026-10-09 14:54:26.408077','2026-10-09 14:54:26.408077');
/*!40000 ALTER TABLE `user_stores` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `user_store_permissions` WRITE;
/*!40000 ALTER TABLE `user_store_permissions` DISABLE KEYS */;
/*!40000 ALTER TABLE `user_store_permissions` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `user_store_roles` WRITE;
/*!40000 ALTER TABLE `user_store_roles` DISABLE KEYS */;
INSERT INTO `user_store_roles` VALUES (3,2,'2026-10-09 14:54:26.407844'),(6,2,'2026-10-09 14:54:26.409012'),(11,2,'2026-10-09 14:54:26.408449');
/*!40000 ALTER TABLE `user_store_roles` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

