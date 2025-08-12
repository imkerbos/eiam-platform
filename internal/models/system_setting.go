package models

import (
	"time"

	"gorm.io/gorm"
)

// SystemSetting 系统设置模型
type SystemSetting struct {
	ID          string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Key         string         `json:"key" gorm:"uniqueIndex;type:varchar(100);not null"`
	Value       string         `json:"value" gorm:"type:text"`
	Description string         `json:"description" gorm:"type:varchar(255)"`
	Category    string         `json:"category" gorm:"type:varchar(50);not null"` // site, security, email, etc.
	Type        string         `json:"type" gorm:"type:varchar(20);not null"`     // string, number, boolean, json
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName 指定表名
func (SystemSetting) TableName() string {
	return "system_settings"
}

// SiteSettings 站点设置结构
type SiteSettings struct {
	SiteName        string `json:"site_name"`
	SiteURL         string `json:"site_url"`
	ContactEmail    string `json:"contact_email"`
	SupportEmail    string `json:"support_email"`
	Description     string `json:"description"`
	Logo            string `json:"logo"`
	Favicon         string `json:"favicon"`
	FooterText      string `json:"footer_text"`
	MaintenanceMode bool   `json:"maintenance_mode"`
}

// SecuritySettings 安全设置结构
type SecuritySettings struct {
	// 密码策略
	MinPasswordLength    int  `json:"min_password_length"`
	MaxPasswordLength    int  `json:"max_password_length"`
	PasswordExpiryDays   int  `json:"password_expiry_days"`
	RequireUppercase     bool `json:"require_uppercase"`
	RequireLowercase     bool `json:"require_lowercase"`
	RequireNumbers       bool `json:"require_numbers"`
	RequireSpecialChars  bool `json:"require_special_chars"`
	PasswordHistoryCount int  `json:"password_history_count"`

	// 会话管理
	SessionTimeout        int `json:"session_timeout"`
	MaxConcurrentSessions int `json:"max_concurrent_sessions"`
	RememberMeDays        int `json:"remember_me_days"`

	// 2FA配置
	Enable2FA           bool `json:"enable_2fa"`
	Require2FAForAdmins bool `json:"require_2fa_for_admins"`
	AllowBackupCodes    bool `json:"allow_backup_codes"`
	EnableTOTP          bool `json:"enable_totp"`
	EnableSMS           bool `json:"enable_sms"`
	EnableEmail         bool `json:"enable_email"`

	// 登录安全
	MaxLoginAttempts           int  `json:"max_login_attempts"`
	LockoutDuration            int  `json:"lockout_duration"`
	EnableIPWhitelist          bool `json:"enable_ip_whitelist"`
	EnableGeolocation          bool `json:"enable_geolocation"`
	EnableDeviceFingerprinting bool `json:"enable_device_fingerprinting"`
	NotifyFailedLogins         bool `json:"notify_failed_logins"`
	NotifyNewDevices           bool `json:"notify_new_devices"`
	NotifyPasswordChanges      bool `json:"notify_password_changes"`
}
