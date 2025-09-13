-- 删除CAS相关表

-- 删除CAS表
DROP TABLE IF EXISTS `cas_proxy_granting_tickets`;
DROP TABLE IF EXISTS `cas_proxy_tickets`;
DROP TABLE IF EXISTS `cas_service_tickets`;

-- 删除CAS字段（如果存在）
SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
     WHERE table_name = 'applications' 
     AND table_schema = DATABASE() 
     AND column_name = 'renew') > 0,
    'ALTER TABLE applications DROP COLUMN renew',
    'SELECT "Column renew does not exist"'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
     WHERE table_name = 'applications' 
     AND table_schema = DATABASE() 
     AND column_name = 'gateway') > 0,
    'ALTER TABLE applications DROP COLUMN gateway',
    'SELECT "Column gateway does not exist"'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
     WHERE table_name = 'applications' 
     AND table_schema = DATABASE() 
     AND column_name = 'service_url') > 0,
    'ALTER TABLE applications DROP COLUMN service_url',
    'SELECT "Column service_url does not exist"'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
