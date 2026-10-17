-- MariaDB dump 10.19  Distrib 10.4.32-MariaDB, for Win64 (AMD64)
--
-- Host: 127.0.0.1    Database: mksiaga_dev
-- ------------------------------------------------------
-- Server version	10.4.32-MariaDB

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

--
-- Table structure for table `daily_cctv_incident_actions`
--

DROP TABLE IF EXISTS `daily_cctv_incident_actions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `daily_cctv_incident_actions` (
  `incident_id` bigint(20) unsigned NOT NULL,
  `action_id` bigint(20) unsigned NOT NULL,
  `notes` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`incident_id`,`action_id`),
  KEY `daily_incident_action_master_fk` (`action_id`),
  CONSTRAINT `daily_incident_action_incident_fk` FOREIGN KEY (`incident_id`) REFERENCES `daily_cctv_incidents` (`id`) ON DELETE CASCADE,
  CONSTRAINT `daily_incident_action_master_fk` FOREIGN KEY (`action_id`) REFERENCES `master_tindakan` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `daily_cctv_incident_actions`
--

LOCK TABLES `daily_cctv_incident_actions` WRITE;
/*!40000 ALTER TABLE `daily_cctv_incident_actions` DISABLE KEYS */;
/*!40000 ALTER TABLE `daily_cctv_incident_actions` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `daily_cctv_incidents`
--

DROP TABLE IF EXISTS `daily_cctv_incidents`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `daily_cctv_incidents` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `report_id` bigint(20) unsigned NOT NULL,
  `incident_time` time DEFAULT NULL,
  `area_id` bigint(20) unsigned DEFAULT NULL,
  `camera_id` bigint(20) unsigned DEFAULT NULL,
  `incident_type_id` bigint(20) unsigned DEFAULT NULL,
  `other_type` varchar(150) DEFAULT NULL,
  `description` text DEFAULT NULL,
  `perpetrator_gender` varchar(30) DEFAULT NULL,
  `perpetrator_age` varchar(30) DEFAULT NULL,
  `perpetrator_clothing` varchar(255) DEFAULT NULL,
  `perpetrator_features` varchar(255) DEFAULT NULL,
  `response_minutes` smallint(5) unsigned DEFAULT NULL,
  `incident_status` enum('completed','in_progress','forwarded') DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `daily_incident_report_index` (`report_id`),
  KEY `daily_incident_area_fk` (`area_id`),
  KEY `daily_incident_camera_fk` (`camera_id`),
  KEY `daily_incident_type_fk` (`incident_type_id`),
  CONSTRAINT `daily_incident_area_fk` FOREIGN KEY (`area_id`) REFERENCES `master_area` (`id`),
  CONSTRAINT `daily_incident_camera_fk` FOREIGN KEY (`camera_id`) REFERENCES `master_kamera` (`id`),
  CONSTRAINT `daily_incident_report_fk` FOREIGN KEY (`report_id`) REFERENCES `daily_cctv_reports` (`id`) ON DELETE CASCADE,
  CONSTRAINT `daily_incident_type_fk` FOREIGN KEY (`incident_type_id`) REFERENCES `master_jenis_insiden` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `daily_cctv_incidents`
--

LOCK TABLES `daily_cctv_incidents` WRITE;
/*!40000 ALTER TABLE `daily_cctv_incidents` DISABLE KEYS */;
/*!40000 ALTER TABLE `daily_cctv_incidents` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `daily_cctv_monitoring_logs`
--

DROP TABLE IF EXISTS `daily_cctv_monitoring_logs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `daily_cctv_monitoring_logs` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `report_id` bigint(20) unsigned NOT NULL,
  `log_time` time NOT NULL,
  `area_id` bigint(20) unsigned DEFAULT NULL,
  `condition_status` enum('normal','finding') NOT NULL,
  `finding` text DEFAULT NULL,
  `action_taken` text DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `daily_monitoring_report_time_index` (`report_id`,`log_time`),
  KEY `daily_monitoring_area_fk` (`area_id`),
  CONSTRAINT `daily_monitoring_area_fk` FOREIGN KEY (`area_id`) REFERENCES `master_area` (`id`),
  CONSTRAINT `daily_monitoring_report_fk` FOREIGN KEY (`report_id`) REFERENCES `daily_cctv_reports` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `daily_cctv_monitoring_logs`
--

LOCK TABLES `daily_cctv_monitoring_logs` WRITE;
/*!40000 ALTER TABLE `daily_cctv_monitoring_logs` DISABLE KEYS */;
/*!40000 ALTER TABLE `daily_cctv_monitoring_logs` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `daily_cctv_report_checklists`
--

DROP TABLE IF EXISTS `daily_cctv_report_checklists`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `daily_cctv_report_checklists` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `report_id` bigint(20) unsigned NOT NULL,
  `checklist_id` bigint(20) unsigned NOT NULL,
  `answer` enum('yes','no') NOT NULL,
  `camera_active` smallint(5) unsigned DEFAULT NULL,
  `camera_total` smallint(5) unsigned DEFAULT NULL,
  `notes` varchar(500) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `daily_report_checklist_unique` (`report_id`,`checklist_id`),
  KEY `daily_report_checklist_master_fk` (`checklist_id`),
  CONSTRAINT `daily_report_checklist_master_fk` FOREIGN KEY (`checklist_id`) REFERENCES `master_checklist` (`id`),
  CONSTRAINT `daily_report_checklist_report_fk` FOREIGN KEY (`report_id`) REFERENCES `daily_cctv_reports` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `daily_cctv_report_checklists`
--

LOCK TABLES `daily_cctv_report_checklists` WRITE;
/*!40000 ALTER TABLE `daily_cctv_report_checklists` DISABLE KEYS */;
/*!40000 ALTER TABLE `daily_cctv_report_checklists` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `daily_cctv_reports`
--

DROP TABLE IF EXISTS `daily_cctv_reports`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `daily_cctv_reports` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `store_id` bigint(20) unsigned NOT NULL,
  `officer_id` bigint(20) unsigned NOT NULL,
  `supervisor_id` bigint(20) unsigned DEFAULT NULL,
  `shift_id` bigint(20) unsigned NOT NULL,
  `report_date` date NOT NULL,
  `work_start` time NOT NULL,
  `work_end` time NOT NULL,
  `monitoring_location` varchar(255) NOT NULL,
  `start_notes` text DEFAULT NULL,
  `end_notes` text DEFAULT NULL,
  `handover_to_user_id` bigint(20) unsigned DEFAULT NULL,
  `handover_at` time DEFAULT NULL,
  `handover_notes` text DEFAULT NULL,
  `handover_from_signature` mediumtext DEFAULT NULL,
  `handover_to_signature` mediumtext DEFAULT NULL,
  `status` enum('draft','submitted') NOT NULL DEFAULT 'draft',
  `submitted_at` datetime(6) DEFAULT NULL,
  `created_at` datetime(6) NOT NULL DEFAULT current_timestamp(6),
  `updated_at` datetime(6) NOT NULL DEFAULT current_timestamp(6) ON UPDATE current_timestamp(6),
  PRIMARY KEY (`id`),
  UNIQUE KEY `daily_report_identity_unique` (`store_id`,`officer_id`,`report_date`,`shift_id`),
  KEY `daily_report_store_date_index` (`store_id`,`report_date`),
  KEY `daily_report_officer_fk` (`officer_id`),
  KEY `daily_report_supervisor_fk` (`supervisor_id`),
  KEY `daily_report_shift_fk` (`shift_id`),
  KEY `daily_report_handover_user_fk` (`handover_to_user_id`),
  CONSTRAINT `daily_report_handover_user_fk` FOREIGN KEY (`handover_to_user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `daily_report_officer_fk` FOREIGN KEY (`officer_id`) REFERENCES `users` (`id`),
  CONSTRAINT `daily_report_shift_fk` FOREIGN KEY (`shift_id`) REFERENCES `master_shift` (`id`),
  CONSTRAINT `daily_report_store_fk` FOREIGN KEY (`store_id`) REFERENCES `stores` (`id`),
  CONSTRAINT `daily_report_supervisor_fk` FOREIGN KEY (`supervisor_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `daily_cctv_reports`
--

LOCK TABLES `daily_cctv_reports` WRITE;
/*!40000 ALTER TABLE `daily_cctv_reports` DISABLE KEYS */;
/*!40000 ALTER TABLE `daily_cctv_reports` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `master_area`
--

DROP TABLE IF EXISTS `master_area`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `master_area` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `store_id` bigint(20) unsigned NOT NULL,
  `code` varchar(50) NOT NULL,
  `name` varchar(150) NOT NULL,
  `is_priority` tinyint(1) NOT NULL DEFAULT 0,
  `sort_order` smallint(5) unsigned NOT NULL DEFAULT 0,
  `status` enum('active','inactive') NOT NULL DEFAULT 'active',
  `created_at` datetime(6) NOT NULL DEFAULT current_timestamp(6),
  `updated_at` datetime(6) NOT NULL DEFAULT current_timestamp(6) ON UPDATE current_timestamp(6),
  PRIMARY KEY (`id`),
  UNIQUE KEY `master_area_store_code_unique` (`store_id`,`code`),
  UNIQUE KEY `master_area_store_id_unique` (`store_id`,`id`),
  CONSTRAINT `master_area_store_fk` FOREIGN KEY (`store_id`) REFERENCES `stores` (`id`),
  CONSTRAINT `master_area_priority_check` CHECK (`is_priority` in (0,1))
) ENGINE=InnoDB AUTO_INCREMENT=71 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `master_area`
--

LOCK TABLES `master_area` WRITE;
/*!40000 ALTER TABLE `master_area` DISABLE KEYS */;
INSERT INTO `master_area` VALUES (1,1,'POS','Area Kasir / POS',1,1,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(2,1,'ENTRANCE_EXIT','Pintu Masuk & Keluar',1,2,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(3,1,'STOCKROOM','Area Gudang / Stockroom',1,3,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(4,1,'HIGH_VALUE','Rak Produk Bernilai Tinggi (elektronik, kosmetik, alkohol)',1,4,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(5,1,'PARKING','Area Parkir',1,5,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(6,1,'FITTING_TOILET_ACCESS','Area Akses Luar Ruang Ganti / Toilet',1,6,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(7,1,'LOADING_DOCK','Area Belakang Toko (loading dock)',1,7,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(8,2,'POS','Area Kasir / POS',1,1,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(9,2,'ENTRANCE_EXIT','Pintu Masuk & Keluar',1,2,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(10,2,'STOCKROOM','Area Gudang / Stockroom',1,3,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(11,2,'HIGH_VALUE','Rak Produk Bernilai Tinggi (elektronik, kosmetik, alkohol)',1,4,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(12,2,'PARKING','Area Parkir',1,5,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(13,2,'FITTING_TOILET_ACCESS','Area Akses Luar Ruang Ganti / Toilet',1,6,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(14,2,'LOADING_DOCK','Area Belakang Toko (loading dock)',1,7,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(15,3,'POS','Area Kasir / POS',1,1,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(16,3,'ENTRANCE_EXIT','Pintu Masuk & Keluar',1,2,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(17,3,'STOCKROOM','Area Gudang / Stockroom',1,3,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(18,3,'HIGH_VALUE','Rak Produk Bernilai Tinggi (elektronik, kosmetik, alkohol)',1,4,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(19,3,'PARKING','Area Parkir',1,5,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(20,3,'FITTING_TOILET_ACCESS','Area Akses Luar Ruang Ganti / Toilet',1,6,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(21,3,'LOADING_DOCK','Area Belakang Toko (loading dock)',1,7,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(22,4,'POS','Area Kasir / POS',1,1,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(23,4,'ENTRANCE_EXIT','Pintu Masuk & Keluar',1,2,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(24,4,'STOCKROOM','Area Gudang / Stockroom',1,3,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(25,4,'HIGH_VALUE','Rak Produk Bernilai Tinggi (elektronik, kosmetik, alkohol)',1,4,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(26,4,'PARKING','Area Parkir',1,5,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(27,4,'FITTING_TOILET_ACCESS','Area Akses Luar Ruang Ganti / Toilet',1,6,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(28,4,'LOADING_DOCK','Area Belakang Toko (loading dock)',1,7,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(29,5,'POS','Area Kasir / POS',1,1,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(30,5,'ENTRANCE_EXIT','Pintu Masuk & Keluar',1,2,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(31,5,'STOCKROOM','Area Gudang / Stockroom',1,3,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(32,5,'HIGH_VALUE','Rak Produk Bernilai Tinggi (elektronik, kosmetik, alkohol)',1,4,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(33,5,'PARKING','Area Parkir',1,5,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(34,5,'FITTING_TOILET_ACCESS','Area Akses Luar Ruang Ganti / Toilet',1,6,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(35,5,'LOADING_DOCK','Area Belakang Toko (loading dock)',1,7,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(36,6,'POS','Area Kasir / POS',1,1,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(37,6,'ENTRANCE_EXIT','Pintu Masuk & Keluar',1,2,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(38,6,'STOCKROOM','Area Gudang / Stockroom',1,3,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(39,6,'HIGH_VALUE','Rak Produk Bernilai Tinggi (elektronik, kosmetik, alkohol)',1,4,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(40,6,'PARKING','Area Parkir',1,5,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(41,6,'FITTING_TOILET_ACCESS','Area Akses Luar Ruang Ganti / Toilet',1,6,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(42,6,'LOADING_DOCK','Area Belakang Toko (loading dock)',1,7,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(43,7,'POS','Area Kasir / POS',1,1,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(44,7,'ENTRANCE_EXIT','Pintu Masuk & Keluar',1,2,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(45,7,'STOCKROOM','Area Gudang / Stockroom',1,3,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(46,7,'HIGH_VALUE','Rak Produk Bernilai Tinggi (elektronik, kosmetik, alkohol)',1,4,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(47,7,'PARKING','Area Parkir',1,5,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(48,7,'FITTING_TOILET_ACCESS','Area Akses Luar Ruang Ganti / Toilet',1,6,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(49,7,'LOADING_DOCK','Area Belakang Toko (loading dock)',1,7,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(50,101,'POS','Area Kasir / POS',1,1,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(51,101,'ENTRANCE_EXIT','Pintu Masuk & Keluar',1,2,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(52,101,'STOCKROOM','Area Gudang / Stockroom',1,3,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(53,101,'HIGH_VALUE','Rak Produk Bernilai Tinggi (elektronik, kosmetik, alkohol)',1,4,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(54,101,'PARKING','Area Parkir',1,5,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(55,101,'FITTING_TOILET_ACCESS','Area Akses Luar Ruang Ganti / Toilet',1,6,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(56,101,'LOADING_DOCK','Area Belakang Toko (loading dock)',1,7,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(57,102,'POS','Area Kasir / POS',1,1,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(58,102,'ENTRANCE_EXIT','Pintu Masuk & Keluar',1,2,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(59,102,'STOCKROOM','Area Gudang / Stockroom',1,3,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(60,102,'HIGH_VALUE','Rak Produk Bernilai Tinggi (elektronik, kosmetik, alkohol)',1,4,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(61,102,'PARKING','Area Parkir',1,5,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(62,102,'FITTING_TOILET_ACCESS','Area Akses Luar Ruang Ganti / Toilet',1,6,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(63,102,'LOADING_DOCK','Area Belakang Toko (loading dock)',1,7,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(64,103,'POS','Area Kasir / POS',1,1,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(65,103,'ENTRANCE_EXIT','Pintu Masuk & Keluar',1,2,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(66,103,'STOCKROOM','Area Gudang / Stockroom',1,3,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(67,103,'HIGH_VALUE','Rak Produk Bernilai Tinggi (elektronik, kosmetik, alkohol)',1,4,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(68,103,'PARKING','Area Parkir',1,5,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(69,103,'FITTING_TOILET_ACCESS','Area Akses Luar Ruang Ganti / Toilet',1,6,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812'),(70,103,'LOADING_DOCK','Area Belakang Toko (loading dock)',1,7,'active','2026-10-12 09:36:02.601812','2026-10-12 09:36:02.601812');
/*!40000 ALTER TABLE `master_area` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `master_checklist`
--

DROP TABLE IF EXISTS `master_checklist`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `master_checklist` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `code` varchar(50) NOT NULL,
  `phase` enum('start','end') NOT NULL,
  `name` varchar(255) NOT NULL,
  `requires_camera_count` tinyint(1) NOT NULL DEFAULT 0,
  `sort_order` smallint(5) unsigned NOT NULL DEFAULT 0,
  `status` enum('active','inactive') NOT NULL DEFAULT 'active',
  `created_at` datetime(6) NOT NULL DEFAULT current_timestamp(6),
  `updated_at` datetime(6) NOT NULL DEFAULT current_timestamp(6) ON UPDATE current_timestamp(6),
  PRIMARY KEY (`id`),
  UNIQUE KEY `master_checklist_code_unique` (`code`),
  KEY `master_checklist_phase_order_index` (`phase`,`sort_order`),
  CONSTRAINT `master_checklist_count_check` CHECK (`requires_camera_count` in (0,1))
) ENGINE=InnoDB AUTO_INCREMENT=213 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `master_checklist`
--

LOCK TABLES `master_checklist` WRITE;
/*!40000 ALTER TABLE `master_checklist` DISABLE KEYS */;
INSERT INTO `master_checklist` VALUES (1,'START_01','start','Semua monitor CCTV berfungsi normal',0,1,'active','2026-10-12 09:36:02.607163','2026-10-12 11:09:20.080908'),(2,'START_02','start','Jumlah kamera aktif sesuai standar',1,2,'active','2026-10-12 09:36:02.607163','2026-10-12 11:09:20.080908'),(3,'START_03','start','Kualitas gambar jelas (tidak blur/terhalang)',0,3,'active','2026-10-12 09:36:02.607163','2026-10-12 11:09:20.080908'),(4,'START_04','start','Rekaman (recording) berjalan normal',0,4,'active','2026-10-12 09:36:02.607163','2026-10-12 11:09:20.080908'),(5,'START_05','start','Sistem alarm berfungsi',0,5,'active','2026-10-12 09:36:02.607163','2026-10-12 11:09:20.080908'),(6,'START_06','start','Komunikasi (HT/telepon) berfungsi',0,6,'active','2026-10-12 09:36:02.607163','2026-10-12 11:09:20.080908'),(7,'START_07','start','Buku log shift sebelumnya sudah dibaca',0,7,'active','2026-10-12 09:36:02.607163','2026-10-12 11:09:20.080908'),(8,'START_08','start','Area prioritas (kasir, gudang, pintu masuk/keluar) terpantau',0,8,'active','2026-10-12 09:36:02.607163','2026-10-12 11:09:20.080908'),(9,'END_01','end','Semua rekaman hari ini tersimpan dengan baik',0,1,'active','2026-10-12 09:36:02.607163','2026-10-12 11:09:20.080908'),(10,'END_02','end','Rekaman insiden (jika ada) sudah di-clipping',0,2,'active','2026-10-12 09:36:02.607163','2026-10-12 11:09:20.080908'),(11,'END_03','end','Laporan insiden sudah diserahkan ke supervisor',0,3,'active','2026-10-12 09:36:02.607163','2026-10-12 11:09:20.080908'),(12,'END_04','end','Peralatan CCTV dalam kondisi baik',0,4,'active','2026-10-12 09:36:02.607163','2026-10-12 11:09:20.080908'),(13,'END_05','end','Buku log sudah diisi lengkap',0,5,'active','2026-10-12 09:36:02.607163','2026-10-12 11:09:20.080908'),(14,'END_06','end','Serah terima tugas ke shift berikutnya sudah dilakukan',0,6,'active','2026-10-12 09:36:02.607163','2026-10-12 11:09:20.080908');
/*!40000 ALTER TABLE `master_checklist` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `master_jenis_insiden`
--

DROP TABLE IF EXISTS `master_jenis_insiden`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `master_jenis_insiden` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `code` varchar(50) NOT NULL,
  `name` varchar(150) NOT NULL,
  `requires_description` tinyint(1) NOT NULL DEFAULT 0,
  `sort_order` smallint(5) unsigned NOT NULL DEFAULT 0,
  `status` enum('active','inactive') NOT NULL DEFAULT 'active',
  `created_at` datetime(6) NOT NULL DEFAULT current_timestamp(6),
  `updated_at` datetime(6) NOT NULL DEFAULT current_timestamp(6) ON UPDATE current_timestamp(6),
  PRIMARY KEY (`id`),
  UNIQUE KEY `master_jenis_insiden_code_unique` (`code`),
  CONSTRAINT `master_jenis_insiden_description_check` CHECK (`requires_description` in (0,1))
) ENGINE=InnoDB AUTO_INCREMENT=44 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `master_jenis_insiden`
--

LOCK TABLES `master_jenis_insiden` WRITE;
/*!40000 ALTER TABLE `master_jenis_insiden` DISABLE KEYS */;
INSERT INTO `master_jenis_insiden` VALUES (1,'THEFT','Pencurian (Shoplifting)',0,1,'active','2026-10-12 09:36:02.607564','2026-10-12 09:36:02.607564'),(2,'SUSPICIOUS','Kecurigaan (Suspicious Behavior)',0,2,'active','2026-10-12 09:36:02.607564','2026-10-12 09:36:02.607564'),(3,'ACCIDENT','Kecelakaan / Jatuh',0,3,'active','2026-10-12 09:36:02.607564','2026-10-12 09:36:02.607564'),(4,'PROPERTY_DAMAGE','Kerusakan Properti',0,4,'active','2026-10-12 09:36:02.607564','2026-10-12 09:36:02.607564'),(5,'CONFLICT','Konflik / Keributan',0,5,'active','2026-10-12 09:36:02.607564','2026-10-12 09:36:02.607564'),(6,'INTERNAL_FRAUD','Fraud Internal',0,6,'active','2026-10-12 09:36:02.607564','2026-10-12 09:36:02.607564'),(7,'OTHER','Lainnya',1,7,'active','2026-10-12 09:36:02.607564','2026-10-12 09:36:02.607564');
/*!40000 ALTER TABLE `master_jenis_insiden` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `master_kamera`
--

DROP TABLE IF EXISTS `master_kamera`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `master_kamera` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `store_id` bigint(20) unsigned NOT NULL,
  `area_id` bigint(20) unsigned NOT NULL,
  `code` varchar(50) NOT NULL,
  `name` varchar(150) NOT NULL,
  `description` text DEFAULT NULL,
  `status` enum('active','inactive') NOT NULL DEFAULT 'active',
  `created_at` datetime(6) NOT NULL DEFAULT current_timestamp(6),
  `updated_at` datetime(6) NOT NULL DEFAULT current_timestamp(6) ON UPDATE current_timestamp(6),
  PRIMARY KEY (`id`),
  UNIQUE KEY `master_kamera_store_code_unique` (`store_id`,`code`),
  KEY `master_kamera_store_area_index` (`store_id`,`area_id`),
  CONSTRAINT `master_kamera_area_fk` FOREIGN KEY (`store_id`, `area_id`) REFERENCES `master_area` (`store_id`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `master_kamera`
--

LOCK TABLES `master_kamera` WRITE;
/*!40000 ALTER TABLE `master_kamera` DISABLE KEYS */;
/*!40000 ALTER TABLE `master_kamera` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `master_shift`
--

DROP TABLE IF EXISTS `master_shift`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `master_shift` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `code` varchar(50) NOT NULL,
  `name` varchar(100) NOT NULL,
  `start_time` time DEFAULT NULL,
  `end_time` time DEFAULT NULL,
  `sort_order` smallint(5) unsigned NOT NULL DEFAULT 0,
  `status` enum('active','inactive') NOT NULL DEFAULT 'active',
  `created_at` datetime(6) NOT NULL DEFAULT current_timestamp(6),
  `updated_at` datetime(6) NOT NULL DEFAULT current_timestamp(6) ON UPDATE current_timestamp(6),
  PRIMARY KEY (`id`),
  UNIQUE KEY `master_shift_code_unique` (`code`),
  CONSTRAINT `master_shift_time_check` CHECK (`start_time` is null and `end_time` is null or `start_time` is not null and `end_time` is not null and `start_time` >= '00:00:00' and `start_time` < '24:00:00' and `end_time` >= '00:00:00' and `end_time` < '24:00:00' and `start_time` <> `end_time`)
) ENGINE=InnoDB AUTO_INCREMENT=42 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `master_shift`
--

LOCK TABLES `master_shift` WRITE;
/*!40000 ALTER TABLE `master_shift` DISABLE KEYS */;
INSERT INTO `master_shift` VALUES (1,'PAGI','Pagi',NULL,NULL,1,'active','2026-10-12 09:36:02.606693','2026-10-12 11:09:20.080279'),(2,'SIANG','Siang',NULL,NULL,2,'active','2026-10-12 09:36:02.606693','2026-10-12 11:09:20.080279'),(3,'MALAM','Malam',NULL,NULL,3,'active','2026-10-12 09:36:02.606693','2026-10-12 11:09:20.080279');
/*!40000 ALTER TABLE `master_shift` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `master_tindakan`
--

DROP TABLE IF EXISTS `master_tindakan`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `master_tindakan` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `code` varchar(50) NOT NULL,
  `name` varchar(150) NOT NULL,
  `requires_description` tinyint(1) NOT NULL DEFAULT 0,
  `sort_order` smallint(5) unsigned NOT NULL DEFAULT 0,
  `status` enum('active','inactive') NOT NULL DEFAULT 'active',
  `created_at` datetime(6) NOT NULL DEFAULT current_timestamp(6),
  `updated_at` datetime(6) NOT NULL DEFAULT current_timestamp(6) ON UPDATE current_timestamp(6),
  PRIMARY KEY (`id`),
  UNIQUE KEY `master_tindakan_code_unique` (`code`),
  CONSTRAINT `master_tindakan_description_check` CHECK (`requires_description` in (0,1))
) ENGINE=InnoDB AUTO_INCREMENT=32 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `master_tindakan`
--

LOCK TABLES `master_tindakan` WRITE;
/*!40000 ALTER TABLE `master_tindakan` DISABLE KEYS */;
INSERT INTO `master_tindakan` VALUES (1,'REPORT_SUPERVISOR','Melaporkan ke Supervisor',0,1,'active','2026-10-12 09:36:02.609428','2026-10-12 09:36:02.609428'),(2,'CONTACT_SECURITY','Menghubungi Tim LP/Security Lantai',0,2,'active','2026-10-12 09:36:02.609428','2026-10-12 09:36:02.609428'),(3,'SAVE_CLIP','Menyimpan Rekaman (Clipping)',0,3,'active','2026-10-12 09:36:02.609428','2026-10-12 09:36:02.609428'),(4,'CONTACT_POLICE','Menghubungi Kepolisian',0,4,'active','2026-10-12 09:36:02.609428','2026-10-12 09:36:02.609428'),(5,'OTHER','Lainnya',1,5,'active','2026-10-12 09:36:02.609428','2026-10-12 09:36:02.609428');
/*!40000 ALTER TABLE `master_tindakan` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `permissions`
--

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
) ENGINE=InnoDB AUTO_INCREMENT=54 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `permissions`
--

LOCK TABLES `permissions` WRITE;
/*!40000 ALTER TABLE `permissions` DISABLE KEYS */;
INSERT INTO `permissions` VALUES (2,'roles.view','web','Melihat daftar role','2026-10-08 13:43:34.934316','2026-10-08 13:43:34.934316'),(3,'roles.create','web','Membuat role','2026-10-08 13:43:34.934316','2026-10-08 13:43:34.934316'),(4,'roles.update','web','Memperbarui role dan assignment permission','2026-10-08 13:43:34.934316','2026-10-08 13:43:34.934316'),(5,'roles.delete','web','Menghapus role','2026-10-08 13:43:34.934316','2026-10-08 13:43:34.934316'),(6,'permissions.view','web','Melihat daftar permission','2026-10-08 13:43:34.934316','2026-10-08 13:43:34.934316'),(7,'permissions.create','web','Membuat permission','2026-10-08 13:43:34.934316','2026-10-08 13:43:34.934316'),(8,'permissions.update','web','Memperbarui permission','2026-10-08 13:43:34.934316','2026-10-08 13:43:34.934316'),(9,'permissions.delete','web','Menghapus permission','2026-10-08 13:43:34.934316','2026-10-08 13:43:34.934316'),(19,'users.view','web','Melihat daftar user','2026-10-10 11:12:47.489454','2026-10-10 11:12:47.489454'),(20,'users.create','web','Membuat user','2026-10-10 11:12:47.489454','2026-10-10 11:12:47.489454'),(21,'users.update','web','Memperbarui user, store, dan role','2026-10-10 11:12:47.489454','2026-10-10 11:12:47.489454'),(22,'users.delete','web','Menghapus user','2026-10-10 11:12:47.489454','2026-10-10 11:12:47.489454'),(23,'stores.view','web','Melihat master data store','2026-10-10 11:12:47.510954','2026-10-10 11:12:47.510954'),(24,'stores.create','web','Membuat store','2026-10-10 11:12:47.510954','2026-10-10 11:12:47.510954'),(25,'stores.update','web','Memperbarui store','2026-10-10 11:12:47.510954','2026-10-10 11:12:47.510954'),(26,'stores.delete','web','Menghapus store','2026-10-10 11:12:47.510954','2026-10-10 11:12:47.510954'),(27,'areas.view','web','Melihat master data area','2026-10-12 09:46:41.215520','2026-10-12 09:46:41.215520'),(28,'areas.create','web','Membuat area','2026-10-12 09:46:41.215520','2026-10-12 09:46:41.215520'),(29,'areas.update','web','Memperbarui area','2026-10-12 09:46:41.215520','2026-10-12 09:46:41.215520'),(30,'areas.delete','web','Menghapus area','2026-10-12 09:46:41.215520','2026-10-12 09:46:41.215520'),(31,'checklists.view','web','Melihat master data checklist','2026-10-12 10:23:51.347394','2026-10-12 10:23:51.347394'),(32,'checklists.create','web','Membuat checklist','2026-10-12 10:23:51.347394','2026-10-12 10:23:51.347394'),(33,'checklists.update','web','Memperbarui checklist','2026-10-12 10:23:51.347394','2026-10-12 10:23:51.347394'),(34,'checklists.delete','web','Menghapus checklist','2026-10-12 10:23:51.347394','2026-10-12 10:23:51.347394'),(35,'incident-types.view','web','Melihat master jenis insiden','2026-10-12 10:32:30.951946','2026-10-12 10:32:30.951946'),(36,'incident-types.create','web','Membuat jenis insiden','2026-10-12 10:32:30.951946','2026-10-12 10:32:30.951946'),(37,'incident-types.update','web','Memperbarui jenis insiden','2026-10-12 10:32:30.951946','2026-10-12 10:32:30.951946'),(38,'incident-types.delete','web','Menghapus jenis insiden','2026-10-12 10:32:30.951946','2026-10-12 10:32:30.951946'),(39,'cameras.view','web','Melihat master data kamera','2026-10-12 10:35:24.385069','2026-10-12 10:35:24.385069'),(40,'cameras.create','web','Membuat kamera','2026-10-12 10:35:24.385069','2026-10-12 10:35:24.385069'),(41,'cameras.update','web','Memperbarui kamera','2026-10-12 10:35:24.385069','2026-10-12 10:35:24.385069'),(42,'cameras.delete','web','Menghapus kamera','2026-10-12 10:35:24.385069','2026-10-12 10:35:24.385069'),(43,'shifts.view','web','Melihat master data shift','2026-10-12 10:37:29.798523','2026-10-12 10:37:29.798523'),(44,'shifts.create','web','Membuat shift','2026-10-12 10:37:29.798523','2026-10-12 10:37:29.798523'),(45,'shifts.update','web','Memperbarui shift','2026-10-12 10:37:29.798523','2026-10-12 10:37:29.798523'),(46,'shifts.delete','web','Menghapus shift','2026-10-12 10:37:29.798523','2026-10-12 10:37:29.798523'),(47,'actions.view','web','Melihat master data tindakan','2026-10-12 10:42:48.381689','2026-10-12 10:42:48.381689'),(48,'actions.create','web','Membuat tindakan','2026-10-12 10:42:48.381689','2026-10-12 10:42:48.381689'),(49,'actions.update','web','Memperbarui tindakan','2026-10-12 10:42:48.381689','2026-10-12 10:42:48.381689'),(50,'actions.delete','web','Menghapus tindakan','2026-10-12 10:42:48.381689','2026-10-12 10:42:48.381689'),(51,'daily-reports.view','web','Melihat laporan harian CCTV','2026-10-12 14:42:06.329424','2026-10-12 14:42:06.329424'),(52,'daily-reports.create','web','Membuat laporan harian CCTV','2026-10-12 14:42:06.329424','2026-10-12 14:42:06.329424'),(53,'daily-reports.delete','web','Menghapus laporan harian CCTV','2026-10-12 14:42:06.329424','2026-10-12 14:42:06.329424');
/*!40000 ALTER TABLE `permissions` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `role_permissions`
--

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

--
-- Dumping data for table `role_permissions`
--

LOCK TABLES `role_permissions` WRITE;
/*!40000 ALTER TABLE `role_permissions` DISABLE KEYS */;
INSERT INTO `role_permissions` VALUES (2,2,'2026-10-08 13:43:34.935167'),(2,3,'2026-10-08 13:43:34.935167'),(2,4,'2026-10-08 13:43:34.935167'),(2,5,'2026-10-08 13:43:34.935167'),(2,6,'2026-10-08 13:43:34.935167'),(2,7,'2026-10-08 13:43:34.935167'),(2,8,'2026-10-08 13:43:34.935167'),(2,9,'2026-10-08 13:43:34.935167'),(2,19,'2026-10-10 11:12:47.489710'),(2,20,'2026-10-10 11:12:47.489710'),(2,21,'2026-10-10 11:12:47.489710'),(2,22,'2026-10-10 11:12:47.489710'),(2,23,'2026-10-10 11:12:47.511209'),(2,24,'2026-10-10 11:12:47.511209'),(2,25,'2026-10-10 11:12:47.511209'),(2,26,'2026-10-10 11:12:47.511209'),(2,27,'2026-10-12 09:46:41.216988'),(2,28,'2026-10-12 09:46:41.216988'),(2,29,'2026-10-12 09:46:41.216988'),(2,30,'2026-10-12 09:46:41.216988'),(2,31,'2026-10-12 10:23:51.347989'),(2,32,'2026-10-12 10:23:51.347989'),(2,33,'2026-10-12 10:23:51.347989'),(2,34,'2026-10-12 10:23:51.347989'),(2,35,'2026-10-12 10:32:30.952486'),(2,36,'2026-10-12 10:32:30.952486'),(2,37,'2026-10-12 10:32:30.952486'),(2,38,'2026-10-12 10:32:30.952486'),(2,39,'2026-10-12 10:35:24.385537'),(2,40,'2026-10-12 10:35:24.385537'),(2,41,'2026-10-12 10:35:24.385537'),(2,42,'2026-10-12 10:35:24.385537'),(2,43,'2026-10-12 10:37:29.799139'),(2,44,'2026-10-12 10:37:29.799139'),(2,45,'2026-10-12 10:37:29.799139'),(2,46,'2026-10-12 10:37:29.799139'),(2,47,'2026-10-12 10:42:48.382221'),(2,48,'2026-10-12 10:42:48.382221'),(2,49,'2026-10-12 10:42:48.382221'),(2,50,'2026-10-12 10:42:48.382221'),(2,51,'2026-10-12 14:42:06.330491'),(2,52,'2026-10-12 14:42:06.330491'),(2,53,'2026-10-12 14:42:06.330491');
/*!40000 ALTER TABLE `role_permissions` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `roles`
--

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

--
-- Dumping data for table `roles`
--

LOCK TABLES `roles` WRITE;
/*!40000 ALTER TABLE `roles` DISABLE KEYS */;
INSERT INTO `roles` VALUES (2,'superadmin','web','Akses penuh untuk administrasi aplikasi','2026-10-08 12:43:35.508434','2026-10-08 12:43:35.508434');
/*!40000 ALTER TABLE `roles` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `stores`
--

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

--
-- Dumping data for table `stores`
--

LOCK TABLES `stores` WRITE;
/*!40000 ALTER TABLE `stores` DISABLE KEYS */;
INSERT INTO `stores` VALUES (1,'MK1','MK1 Babarsari','Babarsari',NULL,'Asia/Jakarta','active','2026-10-09 14:42:19.608203','2026-10-09 14:42:19.608203',NULL),(2,'MK2','MK2 Simanjuntak','Simanjuntak',NULL,'Asia/Jakarta','active','2026-10-09 14:42:19.608203','2026-10-09 14:42:19.608203',NULL),(3,'MK3','MK3 Supeno','Supeno',NULL,'Asia/Jakarta','active','2026-10-08 12:43:35.506078','2026-10-09 14:42:19.608203',NULL),(4,'MK4','MK4 Palagan','Palagan',NULL,'Asia/Jakarta','active','2026-10-09 14:42:19.608203','2026-10-09 14:42:19.608203',NULL),(5,'MK5','MK5 Godean','Godean',NULL,'Asia/Jakarta','active','2026-10-09 14:42:19.608203','2026-10-09 14:42:19.608203',NULL),(6,'MK6','MK6 Imogiri','Imogiri',NULL,'Asia/Jakarta','active','2026-10-09 14:42:19.608203','2026-10-09 14:42:19.608203',NULL),(7,'MK7','MK7 Keloran','Keloran',NULL,'Asia/Jakarta','active','2026-10-09 14:42:19.608203','2026-10-09 14:42:19.608203',NULL),(101,'MKM1','MK Mini 1 Plemesewu','Plemesewu',NULL,'Asia/Jakarta','active','2026-10-09 14:42:19.608203','2026-10-09 14:42:19.608203',NULL),(102,'MKM2','MK Mini 2 Diro','Diro',NULL,'Asia/Jakarta','active','2026-10-09 14:42:19.608203','2026-10-09 14:42:19.608203',NULL),(103,'MKM3','MK Mini 3 Minomartani','Minomartani',NULL,'Asia/Jakarta','active','2026-10-09 14:42:19.608203','2026-10-09 14:42:19.608203',NULL);
/*!40000 ALTER TABLE `stores` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `user_sessions`
--

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

--
-- Table structure for table `user_store_permissions`
--

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

--
-- Dumping data for table `user_store_permissions`
--

LOCK TABLES `user_store_permissions` WRITE;
/*!40000 ALTER TABLE `user_store_permissions` DISABLE KEYS */;
/*!40000 ALTER TABLE `user_store_permissions` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `user_store_roles`
--

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

--
-- Dumping data for table `user_store_roles`
--

LOCK TABLES `user_store_roles` WRITE;
/*!40000 ALTER TABLE `user_store_roles` DISABLE KEYS */;
INSERT INTO `user_store_roles` VALUES (3,2,'2026-10-09 14:54:26.407844'),(6,2,'2026-10-09 14:54:26.409012'),(11,2,'2026-10-09 14:54:26.408449');
/*!40000 ALTER TABLE `user_store_roles` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `user_stores`
--

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

--
-- Dumping data for table `user_stores`
--

LOCK TABLES `user_stores` WRITE;
/*!40000 ALTER TABLE `user_stores` DISABLE KEYS */;
INSERT INTO `user_stores` VALUES (3,2,3,0,'active','2026-10-08 12:43:35.508773','2026-10-08 12:43:35.508773','2026-10-09 14:54:26.405814'),(6,2,5,1,'active','2026-10-09 14:52:24.169030','2026-10-09 14:52:24.169030','2026-10-09 14:54:26.408630'),(11,2,4,0,'active','2026-10-09 14:54:26.408077','2026-10-09 14:54:26.408077','2026-10-09 14:54:26.408077');
/*!40000 ALTER TABLE `user_stores` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `users`
--

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

--
-- Dumping data for table `users`
--

LOCK TABLES `users` WRITE;
/*!40000 ALTER TABLE `users` DISABLE KEYS */;
INSERT INTO `users` VALUES (2,NULL,'Admin Security','admin','admin@mksiaga.local','$2y$10$lwkQLoTL4oDoSPj6oGHuL.bhCFZzZE3hfwul5dw6f738FrThFhov6',NULL,'active','2026-10-17 12:30:29.979900','2026-10-08 12:43:35.508585','2026-10-17 12:30:29.979900',NULL);
/*!40000 ALTER TABLE `users` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Dumping events for database 'mksiaga_dev'
--

--
-- Dumping routines for database 'mksiaga_dev'
--
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2026-10-17 12:46:48
