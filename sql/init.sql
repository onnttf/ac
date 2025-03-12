/*
 Navicat Premium Dump SQL

 Source Server         : localhost
 Source Server Type    : MySQL
 Source Server Version : 80404 (8.4.4)
 Source Host           : localhost:3306
 Source Schema         : ac

 Target Server Type    : MySQL
 Target Server Version : 80404 (8.4.4)
 File Encoding         : 65001

 Date: 11/03/2025 13:33:22
*/

SET NAMES utf8mb4;
SET
FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for casbin_rule
-- ----------------------------
DROP TABLE IF EXISTS `casbin_rule`;
CREATE TABLE `casbin_rule`
(
    `id`    int unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `ptype` varchar(10)  NOT NULL DEFAULT '' COMMENT 'ptype',
    `v0`    varchar(255) NOT NULL DEFAULT '' COMMENT 'v0',
    `v1`    varchar(255) NOT NULL DEFAULT '' COMMENT 'v1',
    `v2`    varchar(255) NOT NULL DEFAULT '' COMMENT 'v2',
    `v3`    varchar(255) NOT NULL DEFAULT '' COMMENT 'v3',
    `v4`    varchar(255) NOT NULL DEFAULT '' COMMENT 'v4',
    `v5`    varchar(255) NOT NULL DEFAULT '' COMMENT 'v5',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_ptype_v0_v1` (`ptype`,`v0`,`v1`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Table structure for casbin_rule_log
-- ----------------------------
DROP TABLE IF EXISTS `casbin_rule_log`;
CREATE TABLE `casbin_rule_log`
(
    `id`          int unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `log_code`    varchar(50)  NOT NULL DEFAULT '' COMMENT 'log_code',
    `operate`     tinyint      NOT NULL DEFAULT '0' COMMENT 'operate, 1. add, 2. delete',
    `ptype`       varchar(10)  NOT NULL DEFAULT '' COMMENT 'ptype',
    `v0`          varchar(255) NOT NULL DEFAULT '' COMMENT 'v0',
    `v1`          varchar(255) NOT NULL DEFAULT '' COMMENT 'v1',
    `v2`          varchar(255) NOT NULL DEFAULT '' COMMENT 'v2',
    `v3`          varchar(255) NOT NULL DEFAULT '' COMMENT 'v3',
    `v4`          varchar(255) NOT NULL DEFAULT '' COMMENT 'v4',
    `v5`          varchar(255) NOT NULL DEFAULT '' COMMENT 'v5',
    `modified_by` varchar(50)  NOT NULL DEFAULT '' COMMENT 'modified_by',
    `created_at`  datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created_at',
    PRIMARY KEY (`id`),
    KEY           `idx_ptype_v0_v1` (`ptype`,`v0`,`v1`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Table structure for resource
-- ----------------------------
DROP TABLE IF EXISTS `resource`;
CREATE TABLE `resource`
(
    `id`          int         NOT NULL AUTO_INCREMENT COMMENT 'id',
    `system_code` varchar(50) NOT NULL DEFAULT '' COMMENT 'system_code',
    `code`        varchar(50) NOT NULL DEFAULT '' COMMENT 'code',
    `name`        varchar(50) NOT NULL DEFAULT '' COMMENT 'name',
    `parent_code` varchar(50) NOT NULL DEFAULT '' COMMENT 'parent_code',
    `description` varchar(50) NOT NULL DEFAULT '' COMMENT 'description',
    `modified_by` varchar(50) NOT NULL DEFAULT '' COMMENT 'modified_by',
    `created_at`  datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created_at',
    `updated_at`  datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated_at',
    `deleted_at`  datetime             DEFAULT NULL COMMENT 'deleted_at',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_system_code_code` (`system_code`,`code`),
    KEY           `idx_system_code_name` (`system_code`,`name`),
    KEY           `idx_system_code_parent_code` (`system_code`,`parent_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Table structure for role
-- ----------------------------
DROP TABLE IF EXISTS `role`;
CREATE TABLE `role`
(
    `id`          int unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `system_code` varchar(50) NOT NULL DEFAULT '' COMMENT 'system_code',
    `name`        varchar(50) NOT NULL DEFAULT '' COMMENT 'name',
    `code`        varchar(50) NOT NULL DEFAULT '' COMMENT 'code',
    `description` varchar(50) NOT NULL DEFAULT '' COMMENT 'description',
    `modified_by` varchar(50) NOT NULL DEFAULT '' COMMENT 'modified_by',
    `created_at`  datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created_at',
    `updated_at`  datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated_at',
    `deleted_at`  datetime             DEFAULT NULL COMMENT 'deleted_at',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_system_code_code` (`system_code`,`code`),
    KEY           `idx_system_code_name` (`system_code`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Table structure for system
-- ----------------------------
DROP TABLE IF EXISTS `system`;
CREATE TABLE `system`
(
    `id`          int unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `code`        varchar(50) NOT NULL DEFAULT '' COMMENT 'code',
    `name`        varchar(50) NOT NULL DEFAULT '' COMMENT 'name',
    `description` varchar(50) NOT NULL DEFAULT '' COMMENT 'description',
    `modified_by` varchar(50) NOT NULL DEFAULT '' COMMENT 'modified_by',
    `created_at`  datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created_at',
    `updated_at`  datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated_at',
    `deleted_at`  datetime             DEFAULT NULL COMMENT 'deleted_at',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_code` (`code`),
    KEY           `idx_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Table structure for user
-- ----------------------------
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user`
(
    `id`          int unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `system_code` varchar(50) NOT NULL DEFAULT '' COMMENT 'system_code',
    `name`        varchar(50) NOT NULL DEFAULT '' COMMENT 'name',
    `code`        varchar(50) NOT NULL DEFAULT '' COMMENT 'code',
    `description` varchar(50) NOT NULL DEFAULT '' COMMENT 'description',
    `modified_by` varchar(50) NOT NULL DEFAULT '' COMMENT 'modified_by',
    `created_at`  datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created_at',
    `updated_at`  datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated_at',
    `deleted_at`  datetime             DEFAULT NULL COMMENT 'deleted_at',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_system_code_code` (`system_code`,`code`),
    KEY           `idx_system_code_name` (`system_code`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

SET
FOREIGN_KEY_CHECKS = 1;
