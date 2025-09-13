-- 创建CAS相关表

-- CAS服务票据表
CREATE TABLE IF NOT EXISTS `cas_service_tickets` (
    `id` varchar(36) NOT NULL PRIMARY KEY,
    `created_at` datetime(3) NOT NULL,
    `updated_at` datetime(3) NOT NULL,
    `deleted_at` datetime(3) NULL,
    `ticket` varchar(255) NOT NULL UNIQUE,
    `service` varchar(500) NOT NULL,
    `user_id` varchar(36) NOT NULL,
    `username` varchar(100) NOT NULL,
    `expires_at` datetime(3) NOT NULL,
    `used` boolean DEFAULT false,
    `attributes` text,
    INDEX `idx_cas_service_tickets_user_id` (`user_id`),
    INDEX `idx_cas_service_tickets_ticket` (`ticket`),
    INDEX `idx_cas_service_tickets_expires_at` (`expires_at`),
    INDEX `idx_cas_service_tickets_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- CAS代理票据表
CREATE TABLE IF NOT EXISTS `cas_proxy_tickets` (
    `id` varchar(36) NOT NULL PRIMARY KEY,
    `created_at` datetime(3) NOT NULL,
    `updated_at` datetime(3) NOT NULL,
    `deleted_at` datetime(3) NULL,
    `ticket` varchar(255) NOT NULL UNIQUE,
    `service` varchar(500) NOT NULL,
    `user_id` varchar(36) NOT NULL,
    `username` varchar(100) NOT NULL,
    `proxy_granting_ticket` varchar(255),
    `expires_at` datetime(3) NOT NULL,
    `used` boolean DEFAULT false,
    INDEX `idx_cas_proxy_tickets_user_id` (`user_id`),
    INDEX `idx_cas_proxy_tickets_ticket` (`ticket`),
    INDEX `idx_cas_proxy_tickets_expires_at` (`expires_at`),
    INDEX `idx_cas_proxy_tickets_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- CAS代理授权票据表
CREATE TABLE IF NOT EXISTS `cas_proxy_granting_tickets` (
    `id` varchar(36) NOT NULL PRIMARY KEY,
    `created_at` datetime(3) NOT NULL,
    `updated_at` datetime(3) NOT NULL,
    `deleted_at` datetime(3) NULL,
    `ticket` varchar(255) NOT NULL UNIQUE,
    `user_id` varchar(36) NOT NULL,
    `username` varchar(100) NOT NULL,
    `expires_at` datetime(3) NOT NULL,
    `used` boolean DEFAULT false,
    INDEX `idx_cas_proxy_granting_tickets_user_id` (`user_id`),
    INDEX `idx_cas_proxy_granting_tickets_ticket` (`ticket`),
    INDEX `idx_cas_proxy_granting_tickets_expires_at` (`expires_at`),
    INDEX `idx_cas_proxy_granting_tickets_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 添加CAS字段到applications表（如果不存在）
SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
     WHERE table_name = 'applications' 
     AND table_schema = DATABASE() 
     AND column_name = 'service_url') = 0,
    'ALTER TABLE applications ADD COLUMN service_url VARCHAR(500) AFTER digest_algorithm',
    'SELECT "Column service_url already exists"'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
     WHERE table_name = 'applications' 
     AND table_schema = DATABASE() 
     AND column_name = 'gateway') = 0,
    'ALTER TABLE applications ADD COLUMN gateway BOOLEAN DEFAULT FALSE AFTER service_url',
    'SELECT "Column gateway already exists"'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
     WHERE table_name = 'applications' 
     AND table_schema = DATABASE() 
     AND column_name = 'renew') = 0,
    'ALTER TABLE applications ADD COLUMN renew BOOLEAN DEFAULT FALSE AFTER gateway',
    'SELECT "Column renew already exists"'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
