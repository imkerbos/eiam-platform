-- 添加多设备登录配置到系统设置表
INSERT INTO system_settings (id, `key`, value, description, category, type, created_at, updated_at) VALUES
('setting-001', 'allow_multi_device_login', 'false', '是否允许多设备同时登录', 'security', 'boolean', NOW(), NOW())
ON DUPLICATE KEY UPDATE
value = VALUES(value),
description = VALUES(description),
updated_at = NOW();
