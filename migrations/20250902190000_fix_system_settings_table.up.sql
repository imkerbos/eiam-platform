-- 确保system_settings表存在并有正确的结构
CREATE TABLE IF NOT EXISTS `system_settings` (
  `id` varchar(36) NOT NULL,
  `key` varchar(100) NOT NULL,
  `value` text,
  `description` varchar(255) DEFAULT NULL,
  `category` varchar(50) NOT NULL,
  `type` varchar(20) NOT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_system_settings_key` (`key`),
  KEY `idx_system_settings_category` (`category`),
  KEY `idx_system_settings_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 插入默认的安全设置
INSERT IGNORE INTO `system_settings` (`id`, `key`, `value`, `description`, `category`, `type`, `created_at`, `updated_at`) VALUES
('setting-001', 'allow_multi_device_login', 'false', '是否允许多设备同时登录', 'security', 'boolean', NOW(), NOW()),
('setting-002', 'session_timeout', '30', '会话超时时间（分钟）', 'security', 'number', NOW(), NOW()),
('setting-003', 'max_concurrent_sessions', '3', '最大并发会话数', 'security', 'number', NOW(), NOW()),
('setting-004', 'remember_me_days', '30', '记住我功能持续时间（天）', 'security', 'number', NOW(), NOW()),
('setting-005', 'min_password_length', '8', '最小密码长度', 'security', 'number', NOW(), NOW()),
('setting-006', 'max_password_length', '64', '最大密码长度', 'security', 'number', NOW(), NOW()),
('setting-007', 'password_expiry_days', '90', '密码过期时间（天）', 'security', 'number', NOW(), NOW()),
('setting-008', 'require_uppercase', 'true', '密码必须包含大写字母', 'security', 'boolean', NOW(), NOW()),
('setting-009', 'require_lowercase', 'true', '密码必须包含小写字母', 'security', 'boolean', NOW(), NOW()),
('setting-010', 'require_numbers', 'true', '密码必须包含数字', 'security', 'boolean', NOW(), NOW()),
('setting-011', 'require_special_chars', 'true', '密码必须包含特殊字符', 'security', 'boolean', NOW(), NOW()),
('setting-012', 'password_history_count', '5', '密码历史记录数量', 'security', 'number', NOW(), NOW()),
('setting-013', 'enable_2fa', 'false', '启用双因素认证', 'security', 'boolean', NOW(), NOW()),
('setting-014', 'require_2fa_for_admins', 'true', '管理员必须启用双因素认证', 'security', 'boolean', NOW(), NOW()),
('setting-015', 'max_login_attempts', '5', '最大登录尝试次数', 'security', 'number', NOW(), NOW()),
('setting-016', 'lockout_duration', '15', '账户锁定时间（分钟）', 'security', 'number', NOW(), NOW());
