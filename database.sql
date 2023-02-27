-- MySQL dump 10.13  Distrib 8.0.25, for Linux (x86_64)
--
-- Host: localhost    Database: notthetalk
-- ------------------------------------------------------
-- Server version	8.0.25

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
--
-- Table structure for table `banned_word`
--

DROP TABLE IF EXISTS `banned_word`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `banned_word` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL,
  `word` varchar(255) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=13 DEFAULT CHARSET=utf8mb3;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `discussion`
--

DROP TABLE IF EXISTS `discussion`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `discussion` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL,
  `created_date` datetime NOT NULL,
  `folder_id` bigint NOT NULL,
  `header` varchar(1024) CHARACTER SET utf8 COLLATE utf8_general_ci DEFAULT NULL,
  `last_post` datetime NOT NULL,
  `title` varchar(128) CHARACTER SET utf8 COLLATE utf8_general_ci DEFAULT NULL,
  `user_id` bigint NOT NULL,
  `deleted` bit(1) DEFAULT NULL,
  `locked` bit(1) DEFAULT NULL,
  `post_count` int DEFAULT NULL,
  `zorder` int DEFAULT NULL,
  `status` int NOT NULL DEFAULT '0',
  `premoderate` bit(1) NOT NULL DEFAULT b'0',
  `last_updated` datetime DEFAULT NULL,
  `last_post_id` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `FK2A233828C59117F1` (`folder_id`),
  KEY `FK2A2338282AD7D091` (`user_id`),
  KEY `idx_discussion_folder_id` (`folder_id`),
  KEY `idx_discussion_last_post` (`last_post`),
  KEY `idx_discussion_last_zorder` (`zorder`),
  KEY `idx_discussion_status` (`status`),
  CONSTRAINT `FK2A2338282AD7D091` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`),
  CONSTRAINT `FK2A233828C59117F1` FOREIGN KEY (`folder_id`) REFERENCES `folder` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=52797 DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `discussion_activity`
--

DROP TABLE IF EXISTS `discussion_activity`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `discussion_activity` (
  `discussion_id` bigint NOT NULL,
  `post_count` int DEFAULT NULL,
  `last_created_date` datetime DEFAULT NULL,
  UNIQUE KEY `discussion_id` (`discussion_id`),
  KEY `idx_discussion_activity_post_count` (`post_count`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `discussion_tag`
--

DROP TABLE IF EXISTS `discussion_tag`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `discussion_tag` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL,
  `discussion_id` bigint NOT NULL,
  `tag` varchar(255) NOT NULL,
  `weight` double NOT NULL,
  `created_date` datetime DEFAULT NULL,
  `last_updated` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `FK7827D30323D27231` (`discussion_id`),
  KEY `idx_discussion_tag_tag` (`tag`),
  CONSTRAINT `FK7827D30323D27231` FOREIGN KEY (`discussion_id`) REFERENCES `discussion` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=5488 DEFAULT CHARSET=utf8mb3;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `discussion_user`
--

DROP TABLE IF EXISTS `discussion_user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `discussion_user` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL,
  `discussion_id` bigint NOT NULL,
  `user_id` bigint NOT NULL,
  `user_status` int NOT NULL,
  PRIMARY KEY (`id`),
  KEY `FK8CD345822AD7D091` (`user_id`),
  KEY `FK8CD3458223D27231` (`discussion_id`),
  CONSTRAINT `FK8CD3458223D27231` FOREIGN KEY (`discussion_id`) REFERENCES `discussion` (`id`),
  CONSTRAINT `FK8CD345822AD7D091` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=164 DEFAULT CHARSET=utf8mb3;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `external_user_connection`
--

DROP TABLE IF EXISTS `external_user_connection`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `external_user_connection` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL,
  `access_token` varchar(255) NOT NULL,
  `connection_type` varchar(255) NOT NULL,
  `last_updated` datetime NOT NULL,
  `remote_user_id` varchar(255) NOT NULL,
  `request_token` varchar(255) NOT NULL,
  `user_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `FKFF80EF5E2AD7D091` (`user_id`),
  CONSTRAINT `FKFF80EF5E2AD7D091` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=50 DEFAULT CHARSET=utf8mb3;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `folder`
--

DROP TABLE IF EXISTS `folder`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `folder` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL,
  `description` varchar(255) NOT NULL,
  `activity` int NOT NULL DEFAULT '0',
  `folder_key` varchar(255) DEFAULT NULL,
  `type` int NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=35 DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `folder_post_count`
--

DROP TABLE IF EXISTS `folder_post_count`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `folder_post_count` (
  `folder_id` int NOT NULL,
  `numposts` int NOT NULL DEFAULT '0',
  PRIMARY KEY (`folder_id`)
) ENGINE=MyISAM DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `folder_subscription`
--

DROP TABLE IF EXISTS `folder_subscription`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `folder_subscription` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL,
  `created_date` datetime NOT NULL,
  `folder_id` bigint NOT NULL,
  `last_read` datetime NOT NULL,
  `user_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `FKF112C94EC59117F1` (`folder_id`),
  KEY `FKF112C94E2AD7D091` (`user_id`),
  CONSTRAINT `FKF112C94E2AD7D091` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`),
  CONSTRAINT `FKF112C94EC59117F1` FOREIGN KEY (`folder_id`) REFERENCES `folder` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=8109 DEFAULT CHARSET=utf8mb3;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `folder_subscription_exception`
--

DROP TABLE IF EXISTS `folder_subscription_exception`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `folder_subscription_exception` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL,
  `discussion_id` bigint NOT NULL,
  `subscription_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `FK302ACA9E23D27231` (`discussion_id`),
  KEY `FK302ACA9E9F07A69F` (`subscription_id`),
  CONSTRAINT `FK302ACA9E23D27231` FOREIGN KEY (`discussion_id`) REFERENCES `discussion` (`id`),
  CONSTRAINT `FK302ACA9E9F07A69F` FOREIGN KEY (`subscription_id`) REFERENCES `folder_subscription` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=78498 DEFAULT CHARSET=utf8mb3;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `front_page_entry`
--

DROP TABLE IF EXISTS `front_page_entry`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `front_page_entry` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL,
  `discussion_id` bigint NOT NULL,
  `discussion_name` varchar(255) NOT NULL,
  `folder_id` bigint NOT NULL,
  `folder_key` varchar(255) NOT NULL,
  `folder_name` varchar(255) NOT NULL,
  `last_post` datetime(6) DEFAULT NULL,
  `last_post_id` bigint DEFAULT NULL,
  `post_count` int NOT NULL,
  `admin_only` bit(1) NOT NULL DEFAULT b'0',
  `zorder` int NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_front_page_entry_folder_id` (`folder_id`),
  KEY `idx_front_page_entry_discussion_id` (`discussion_id`),
  KEY `idx_front_page_entry_admin_only` (`admin_only`),
  KEY `idx_front_page_entry_last_post` (`last_post`)
) ENGINE=MyISAM AUTO_INCREMENT=9408184 DEFAULT CHARSET=utf8mb3;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `garrick_posts`
--

DROP TABLE IF EXISTS `garrick_posts`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `garrick_posts` (
  `post_id` bigint NOT NULL,
  PRIMARY KEY (`post_id`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `garrick_users`
--

DROP TABLE IF EXISTS `garrick_users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `garrick_users` (
  `user_id` bigint NOT NULL,
  PRIMARY KEY (`user_id`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `guardian_headline`
--

DROP TABLE IF EXISTS `guardian_headline`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `guardian_headline` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL DEFAULT '0',
  `guardian_id` varchar(256) NOT NULL,
  `deleted` bit(1) NOT NULL DEFAULT b'0',
  `section` varchar(256) NOT NULL,
  `publication_date` datetime NOT NULL,
  `url` varchar(512) NOT NULL,
  `headline` varchar(1024) NOT NULL,
  `trail_text` varchar(1024) NOT NULL,
  `discussion_id` int DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_guardian_headline_discussion_id` (`discussion_id`),
  KEY `idx_guardian_headline_publication_date` (`publication_date`)
) ENGINE=MyISAM AUTO_INCREMENT=185690 DEFAULT CHARSET=utf8mb3;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `ignore_user`
--

DROP TABLE IF EXISTS `ignore_user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ignore_user` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL,
  `ignored_user_id` bigint NOT NULL,
  `user_id` bigint NOT NULL,
  `created_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `FK5BB57A182AD7D091` (`user_id`),
  KEY `FK5BB57A18DE824224` (`ignored_user_id`),
  KEY `idx_ignore_user_user_id_ignored_user_id` (`user_id`,`ignored_user_id`),
  CONSTRAINT `FK5BB57A182AD7D091` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`),
  CONSTRAINT `FK5BB57A18DE824224` FOREIGN KEY (`ignored_user_id`) REFERENCES `user` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=27960 DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `login_history`
--

DROP TABLE IF EXISTS `login_history`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `login_history` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL,
  `ip_address` varchar(255) NOT NULL,
  `logged_in_date` datetime NOT NULL,
  `session_id` varchar(255) DEFAULT NULL,
  `user_id` int NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_login_history_ip_address` (`ip_address`)
) ENGINE=InnoDB AUTO_INCREMENT=2617264 DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `moderation_queue`
--

DROP TABLE IF EXISTS `moderation_queue`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `moderation_queue` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL,
  `created_date` datetime NOT NULL,
  `post_id` bigint NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=45601 DEFAULT CHARSET=utf8mb3;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `moderator_comment`
--

DROP TABLE IF EXISTS `moderator_comment`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `moderator_comment` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL,
  `comment` varchar(255) NOT NULL,
  `created_date` datetime NOT NULL,
  `post_id` bigint NOT NULL,
  `result` int NOT NULL,
  `user_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `FKAC9965252AD7D091` (`user_id`),
  KEY `FKAC9965251C492471` (`post_id`),
  CONSTRAINT `FKAC9965251C492471` FOREIGN KEY (`post_id`) REFERENCES `post` (`id`),
  CONSTRAINT `FKAC9965252AD7D091` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=75681 DEFAULT CHARSET=utf8mb3;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `password_reset`
--

DROP TABLE IF EXISTS `password_reset`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `password_reset` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL,
  `reset_key` varchar(255) NOT NULL,
  `user_id` bigint NOT NULL,
  `ip_address` varchar(15) DEFAULT NULL,
  `created_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_password_reset_reset_key` (`reset_key`),
  KEY `FKF4CE9AAB2AD7D091` (`user_id`),
  CONSTRAINT `FKF4CE9AAB2AD7D091` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1354 DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `post`
--

DROP TABLE IF EXISTS `post`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `post` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL,
  `created_date` datetime NOT NULL,
  `discussion_id` bigint NOT NULL,
  `text` varchar(8192) CHARACTER SET utf8 COLLATE utf8_general_ci DEFAULT NULL,
  `user_id` bigint NOT NULL,
  `deleted` bit(1) DEFAULT NULL,
  `moderation_result` int DEFAULT NULL,
  `moderation_score` double DEFAULT NULL,
  `status` int NOT NULL DEFAULT '0',
  `last_edit_date` datetime DEFAULT NULL,
  `markdown` bit(1) NOT NULL DEFAULT b'0',
  `post_count` int DEFAULT NULL,
  `post_num` int NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `FK3498A02AD7D091` (`user_id`),
  KEY `FK3498A023D27231` (`discussion_id`),
  KEY `idx_post_deleted` (`deleted`),
  KEY `idx_post_status` (`status`),
  KEY `idx_post_created_date` (`created_date`),
  KEY `idx_post_status_created_date` (`status`,`created_date`),
  KEY `idx_post_post_num` (`post_num`),
  CONSTRAINT `FK3498A023D27231` FOREIGN KEY (`discussion_id`) REFERENCES `discussion` (`id`),
  CONSTRAINT `FK3498A02AD7D091` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=11169770 DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `post_counts_by_year`
--

DROP TABLE IF EXISTS `post_counts_by_year`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `post_counts_by_year` (
  `year` int DEFAULT NULL,
  `month` int DEFAULT NULL,
  `count` int DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `post_edit`
--

DROP TABLE IF EXISTS `post_edit`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `post_edit` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL,
  `created_date` datetime NOT NULL,
  `post_id` bigint NOT NULL,
  `text` varchar(8192) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `FK775BF2491C492471` (`post_id`),
  CONSTRAINT `FK775BF2491C492471` FOREIGN KEY (`post_id`) REFERENCES `post` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1602898 DEFAULT CHARSET=utf8mb3;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `post_report`
--

DROP TABLE IF EXISTS `post_report`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `post_report` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL,
  `comment` varchar(512) NOT NULL,
  `created_date` datetime NOT NULL,
  `email` varchar(64) NOT NULL,
  `ipaddress` varchar(15) NOT NULL,
  `name` varchar(64) NOT NULL,
  `post_id` bigint NOT NULL,
  `score` double NOT NULL,
  `user_id` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `FK2668C5B32AD7D091` (`user_id`),
  KEY `FK2668C5B31C492471` (`post_id`),
  CONSTRAINT `FK2668C5B31C492471` FOREIGN KEY (`post_id`) REFERENCES `post` (`id`),
  CONSTRAINT `FK2668C5B32AD7D091` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=7816 DEFAULT CHARSET=utf8mb3;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `postcounts`
--

DROP TABLE IF EXISTS `postcounts`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `postcounts` (
  `discussion_id` int DEFAULT NULL,
  `count` int DEFAULT NULL
) ENGINE=MyISAM DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `role`
--

DROP TABLE IF EXISTS `role`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `role` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL,
  `authority` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `authority` (`authority`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `search_history`
--

DROP TABLE IF EXISTS `search_history`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `search_history` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL,
  `ip_address` varchar(255) NOT NULL,
  `query` varchar(255) NOT NULL,
  `search_date` datetime NOT NULL,
  `user_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `FK407DC8FD2AD7D091` (`user_id`),
  CONSTRAINT `FK407DC8FD2AD7D091` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=478669 DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `signup_confirmation`
--

DROP TABLE IF EXISTS `signup_confirmation`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `signup_confirmation` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL,
  `confirmation_key` varchar(255) NOT NULL,
  `created_date` datetime NOT NULL,
  `ip_address` varchar(255) DEFAULT NULL,
  `last_updated` datetime NOT NULL,
  `user_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `FK46730BFC2AD7D091` (`user_id`),
  CONSTRAINT `FK46730BFC2AD7D091` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4810 DEFAULT CHARSET=utf8mb3;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `site_var`
--

DROP TABLE IF EXISTS `site_var`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `site_var` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL,
  `banner_text` varchar(255) DEFAULT NULL,
  `last_post_date` datetime NOT NULL,
  `donation_status` int DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb3;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `subscription`
--

DROP TABLE IF EXISTS `subscription`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `subscription` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL,
  `discussion_id` bigint NOT NULL,
  `user_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `FK1456591D2AD7D091` (`user_id`),
  KEY `FK1456591D23D27231` (`discussion_id`),
  CONSTRAINT `FK1456591D23D27231` FOREIGN KEY (`discussion_id`) REFERENCES `discussion` (`id`),
  CONSTRAINT `FK1456591D2AD7D091` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=771111 DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tmpPostCounts`
--

DROP TABLE IF EXISTS `tmpPostCounts`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `tmpPostCounts` (
  `id` bigint NOT NULL DEFAULT '0',
  `real_post_count` int NOT NULL DEFAULT '0',
  `post_count` int NOT NULL DEFAULT '0'
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tmpdates`
--

DROP TABLE IF EXISTS `tmpdates`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `tmpdates` (
  `user_id` int DEFAULT NULL,
  `created_date` datetime DEFAULT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tomcat_sessions`
--

DROP TABLE IF EXISTS `tomcat_sessions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `tomcat_sessions` (
  `session_id` varchar(100) NOT NULL,
  `valid_session` char(1) NOT NULL,
  `max_inactive` int NOT NULL,
  `last_access` bigint NOT NULL,
  `app_name` varchar(255) DEFAULT NULL,
  `session_data` mediumblob,
  PRIMARY KEY (`session_id`),
  KEY `kapp_name` (`app_name`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user`
--

DROP TABLE IF EXISTS `user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL,
  `account_expired` bit(1) NOT NULL,
  `account_locked` bit(1) NOT NULL,
  `bio` varchar(1024) CHARACTER SET utf8 COLLATE utf8_general_ci DEFAULT NULL,
  `email` varchar(64) NOT NULL,
  `enabled` bit(1) NOT NULL,
  `password` varchar(255) NOT NULL,
  `password_expired` bit(1) NOT NULL,
  `username` varchar(24) CHARACTER SET utf8 COLLATE utf8_general_ci DEFAULT NULL,
  `display_email` bit(1) NOT NULL DEFAULT b'0',
  `created_date` datetime DEFAULT NULL,
  `last_updated` datetime DEFAULT NULL,
  `last_login_date` datetime DEFAULT NULL,
  `email_verified` int NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `username` (`username`)
) ENGINE=InnoDB AUTO_INCREMENT=6587 DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user_address`
--

DROP TABLE IF EXISTS `user_address`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_address` (
  `user_id` int NOT NULL,
  `user_name` varchar(32) DEFAULT NULL,
  `ip_address` varchar(32) DEFAULT NULL,
  KEY `idx_user_address_ip_address` (`ip_address`),
  KEY `idx_user_address_user_name` (`user_name`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user_discussion`
--

DROP TABLE IF EXISTS `user_discussion`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_discussion` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL,
  `discussion_id` bigint NOT NULL,
  `last_post_old` int DEFAULT NULL,
  `last_updated` datetime NOT NULL,
  `user_id` bigint NOT NULL,
  `post_last_updated_old` datetime DEFAULT NULL,
  `last_post` datetime NOT NULL,
  `last_post_count` int NOT NULL DEFAULT '0',
  `last_post_id` bigint NOT NULL DEFAULT '0',
  `created_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_userdiscussion_userid_discussionid` (`user_id`,`discussion_id`),
  KEY `FK79B7E51C2AD7D091` (`user_id`),
  KEY `FK79B7E51C23D27231` (`discussion_id`),
  CONSTRAINT `FK79B7E51C23D27231` FOREIGN KEY (`discussion_id`) REFERENCES `discussion` (`id`),
  CONSTRAINT `FK79B7E51C2AD7D091` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=6013649 DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user_history`
--

DROP TABLE IF EXISTS `user_history`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_history` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL,
  `created_date` datetime NOT NULL,
  `event_data` varchar(255) NOT NULL,
  `event_type` varchar(255) NOT NULL,
  `user_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `FK922F7C202AD7D091` (`user_id`),
  CONSTRAINT `FK922F7C202AD7D091` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=94472 DEFAULT CHARSET=utf8mb3;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user_login_location`
--

DROP TABLE IF EXISTS `user_login_location`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_login_location` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL,
  `geo_location` varchar(255) DEFAULT NULL,
  `ip_address` varchar(255) NOT NULL,
  `user_id` bigint NOT NULL,
  `last_login` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `FK37B83ABF2AD7D091` (`user_id`),
  KEY `idx_user_login_location_ip_address` (`ip_address`),
  CONSTRAINT `FK37B83ABF2AD7D091` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2074758 DEFAULT CHARSET=utf8mb3;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user_mod_note`
--

DROP TABLE IF EXISTS `user_mod_note`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_mod_note` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL,
  `created_date` datetime NOT NULL,
  `last_updated` datetime NOT NULL,
  `mod_id` bigint NOT NULL,
  `note` varchar(255) NOT NULL,
  `user_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `FKDDCA2D032AD7D091` (`user_id`),
  KEY `FKDDCA2D03F3E835FA` (`mod_id`),
  CONSTRAINT `FKDDCA2D032AD7D091` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`),
  CONSTRAINT `FKDDCA2D03F3E835FA` FOREIGN KEY (`mod_id`) REFERENCES `user` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1626 DEFAULT CHARSET=utf8mb3;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user_options`
--

DROP TABLE IF EXISTS `user_options`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_options` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `version` bigint NOT NULL,
  `auto_subs` bit(1) NOT NULL,
  `user_id` bigint NOT NULL,
  `sort_folders_by_activity` bit(1) NOT NULL DEFAULT b'1',
  `markdown` bit(1) NOT NULL DEFAULT b'0',
  `premoderate` bit(1) NOT NULL DEFAULT b'0',
  `subscription_sort_order` int NOT NULL DEFAULT '0',
  `watch` bit(1) NOT NULL DEFAULT b'0',
  `view_type` varchar(16) NOT NULL DEFAULT 'latest',
  `subs_fetch_order` int NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_id` (`user_id`),
  KEY `FK10761E2A2AD7D091` (`user_id`),
  CONSTRAINT `FK10761E2A2AD7D091` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=6573 DEFAULT CHARSET=utf8mb3;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user_role`
--

DROP TABLE IF EXISTS `user_role`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_role` (
  `role_id` bigint NOT NULL,
  `user_id` bigint NOT NULL,
  PRIMARY KEY (`role_id`,`user_id`),
  KEY `FK143BF46A85AD0CB1` (`role_id`),
  KEY `FK143BF46A2AD7D091` (`user_id`),
  CONSTRAINT `FK143BF46A2AD7D091` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`),
  CONSTRAINT `FK143BF46A85AD0CB1` FOREIGN KEY (`role_id`) REFERENCES `role` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2023-02-02  6:08:44
