-- 删除多设备登录配置
DELETE FROM system_settings WHERE `key` = 'allow_multi_device_login' AND category = 'security';
