DROP TABLE IF EXISTS `tbl_subject`;
CREATE TABLE `tbl_subject`
(
    `id`          INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'id',
    `type`        INT          NOT NULL DEFAULT 0 COMMENT 'type',
    `code`        VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'code',
    `name`        VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'name',
    `parent_code` VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'parent_code',
    `sort`        INT          NOT NULL DEFAULT 0 COMMENT 'sort',
    `status`      INT          NOT NULL DEFAULT 0 COMMENT 'status',
    `deleted`     INT          NOT NULL DEFAULT 0 COMMENT 'deleted',
    `created_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created_at',
    `updated_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated_at',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_subject_code` (`code`),
    KEY           `idx_subject_name` (`name`),
    KEY           `idx_subject_parent_deleted_status` (`parent_code`, `deleted`, `status`),
    KEY           `idx_subject_type_deleted_status` (`type`, `deleted`, `status`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;

DROP TABLE IF EXISTS `tbl_object`;
CREATE TABLE `tbl_object`
(
    `id`          INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'id',
    `type`        INT          NOT NULL DEFAULT 0 COMMENT 'type',
    `code`        VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'code',
    `name`        VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'name',
    `parent_code` VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'parent_code',
    `sort`        INT          NOT NULL DEFAULT 0 COMMENT 'sort',
    `status`      INT          NOT NULL DEFAULT 0 COMMENT 'status',
    `deleted`     INT          NOT NULL DEFAULT 0 COMMENT 'deleted',
    `created_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created_at',
    `updated_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated_at',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_object_code` (`code`),
    KEY           `idx_object_name` (`name`),
    KEY           `idx_object_parent_deleted_status` (`parent_code`, `deleted`, `status`),
    KEY           `idx_object_type_deleted_status` (`type`, `deleted`, `status`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT 'tbl_object';

DROP TABLE IF EXISTS `tbl_casbin_rule`;
CREATE TABLE `tbl_casbin_rule`
(
    `id`    INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'id',
    `ptype` VARCHAR(10)  NOT NULL DEFAULT '' COMMENT 'ptype',
    `v0`    VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'v0',
    `v1`    VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'v1',
    `v2`    VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'v2',
    `v3`    VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'v3',
    `v4`    VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'v4',
    `v5`    VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'v5',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_casbin_policy` (`ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`),
    KEY     `idx_casbin_ptype_v0` (`ptype`, `v0`),
    KEY     `idx_casbin_ptype_v1` (`ptype`, `v1`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT 'tbl_casbin_rule';