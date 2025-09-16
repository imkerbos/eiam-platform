-- 添加CAS属性映射配置字段

-- 为applications表添加attribute_mapping字段
ALTER TABLE `applications` ADD COLUMN `attribute_mapping` text NULL COMMENT 'CAS属性映射配置(JSON格式)' AFTER `renew`;
