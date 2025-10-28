DROP TABLE IF EXISTS `tbl_user`;
CREATE TABLE `tbl_user` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'id',
    `code` VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'code',
    `name` VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'name',
    `email` VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'email',
    `password_hash` VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'password_hash',
    `status` TINYINT NOT NULL DEFAULT 0 COMMENT 'status',
    `deleted` TINYINT NOT NULL DEFAULT 0 COMMENT 'deleted',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created_at',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated_at',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_code` (`code`),
    UNIQUE KEY `uk_user_email` (`email`),
    KEY `idx_user_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='tbl_user';

DROP TABLE IF EXISTS `tbl_menu`;
CREATE TABLE `tbl_menu` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'id',
    `code` VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'code',
    `name` VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'name',
    `parent_code` VARCHAR(100) DEFAULT '' COMMENT 'parent_code',
    `sort` INT NOT NULL DEFAULT 0 COMMENT 'sort',
    `url` VARCHAR(100) DEFAULT '' COMMENT 'url',
    `status` TINYINT NOT NULL DEFAULT 0 COMMENT 'status',
    `visible` TINYINT NOT NULL DEFAULT 0 COMMENT 'visible',
    `deleted` TINYINT NOT NULL DEFAULT 0 COMMENT 'deleted',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created_at',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated_at',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_menu_code` (`code`),
    KEY `idx_menu_parent_sort_deleted` (`parent_code`, `sort`, `deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='tbl_menu';

DROP TABLE IF EXISTS `tbl_role`;
CREATE TABLE `tbl_role` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'id',
    `code` VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'code',
    `name` VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'name',
    `parent_code` VARCHAR(100) DEFAULT '' COMMENT 'parent_code',
    `sort` INT NOT NULL DEFAULT 0 COMMENT 'sort',
    `status` TINYINT NOT NULL DEFAULT 0 COMMENT 'status',
    `visible` TINYINT NOT NULL DEFAULT 0 COMMENT 'visible',
    `deleted` TINYINT NOT NULL DEFAULT 0 COMMENT 'deleted',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created_at',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated_at',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_role_code` (`code`),
    KEY `idx_role_parent_sort_deleted` (`parent_code`, `sort`, `deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='tbl_role';

DROP TABLE IF EXISTS `tbl_casbin_rule`;
CREATE TABLE `tbl_casbin_rule` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'id',
    `ptype` VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'ptype',
    `v0` VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'v0',
    `v1` VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'v1',
    `v2` VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'v2',
    `v3` VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'v3',
    `v4` VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'v4',
    `v5` VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'v5',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_casbin_policy` (`ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`),
    KEY `idx_casbin_core` (`ptype`, `v0`, `v1`, `v2`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='tbl_casbin_rule';
