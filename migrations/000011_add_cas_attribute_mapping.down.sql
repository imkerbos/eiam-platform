-- 回滚CAS属性映射配置字段

-- 删除applications表的attribute_mapping字段
ALTER TABLE `applications` DROP COLUMN `attribute_mapping`;
