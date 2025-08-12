-- Create system_settings table
CREATE TABLE IF NOT EXISTS system_settings (
    id VARCHAR(36) PRIMARY KEY,
    `key` VARCHAR(100) NOT NULL UNIQUE,
    value TEXT,
    description VARCHAR(255),
    category VARCHAR(50) NOT NULL,
    type VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_category (category),
    INDEX idx_key (key),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default system settings
INSERT INTO system_settings (id, `key`, value, description, category, type) VALUES
-- Site settings
('site-001', 'site_name', 'EIAM Platform', 'Site name displayed in header and browser title', 'site', 'string'),
('site-002', 'site_url', 'https://eiam.example.com', 'Main site URL', 'site', 'string'),
('site-003', 'contact_email', 'admin@example.com', 'Contact email address', 'site', 'string'),
('site-004', 'support_email', 'support@example.com', 'Support email address', 'site', 'string'),
('site-005', 'description', 'Enterprise Identity and Access Management Platform', 'Site description', 'site', 'string'),
('site-006', 'logo', '', 'Site logo path', 'site', 'string'),
('site-007', 'favicon', '', 'Site favicon path', 'site', 'string'),
('site-008', 'footer_text', '© 2024 EIAM Platform. All rights reserved.', 'Footer text', 'site', 'string'),
('site-009', 'maintenance_mode', 'false', 'Maintenance mode flag', 'site', 'boolean'),

-- Security settings - Password Policy
('security-001', 'min_password_length', '8', 'Minimum password length', 'security', 'number'),
('security-002', 'max_password_length', '128', 'Maximum password length', 'security', 'number'),
('security-003', 'password_expiry_days', '90', 'Password expiry in days', 'security', 'number'),
('security-004', 'require_uppercase', 'true', 'Require uppercase letters in password', 'security', 'boolean'),
('security-005', 'require_lowercase', 'true', 'Require lowercase letters in password', 'security', 'boolean'),
('security-006', 'require_numbers', 'true', 'Require numbers in password', 'security', 'boolean'),
('security-007', 'require_special_chars', 'true', 'Require special characters in password', 'security', 'boolean'),
('security-008', 'password_history_count', '5', 'Number of previous passwords to remember', 'security', 'number'),

-- Security settings - Session Management
('security-009', 'session_timeout', '30', 'Session timeout in minutes', 'security', 'number'),
('security-010', 'max_concurrent_sessions', '3', 'Maximum concurrent sessions per user', 'security', 'number'),
('security-011', 'remember_me_days', '7', 'Remember me duration in days', 'security', 'number'),

-- Security settings - 2FA Configuration
('security-012', 'enable_2fa', 'true', 'Enable 2FA for all users', 'security', 'boolean'),
('security-013', 'require_2fa_for_admins', 'true', 'Require 2FA for administrators', 'security', 'boolean'),
('security-014', 'allow_backup_codes', 'true', 'Allow backup codes for 2FA', 'security', 'boolean'),
('security-015', 'enable_totp', 'true', 'Enable TOTP (Time-based One-Time Password)', 'security', 'boolean'),
('security-016', 'enable_sms', 'false', 'Enable SMS verification', 'security', 'boolean'),
('security-017', 'enable_email', 'true', 'Enable email verification', 'security', 'boolean'),

-- Security settings - Login Security
('security-018', 'max_login_attempts', '5', 'Maximum failed login attempts', 'security', 'number'),
('security-019', 'lockout_duration', '15', 'Account lockout duration in minutes', 'security', 'number'),
('security-020', 'enable_ip_whitelist', 'false', 'Enable IP whitelist', 'security', 'boolean'),
('security-021', 'enable_geolocation', 'true', 'Enable geolocation tracking', 'security', 'boolean'),
('security-022', 'enable_device_fingerprinting', 'true', 'Enable device fingerprinting', 'security', 'boolean'),
('security-023', 'notify_failed_logins', 'true', 'Notify on failed login attempts', 'security', 'boolean'),
('security-024', 'notify_new_devices', 'true', 'Notify on new device login', 'security', 'boolean'),
('security-025', 'notify_password_changes', 'true', 'Notify on password changes', 'security', 'boolean');
