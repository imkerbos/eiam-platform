package models

import (
	"time"

	"gorm.io/gorm"
)

// PasswordHistory 密码历史记录
type PasswordHistory struct {
	ID        string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID    string         `json:"user_id" gorm:"type:varchar(36);not null;index"`
	Password  string         `json:"-" gorm:"type:text;not null"` // 加密存储
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationships
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (PasswordHistory) TableName() string {
	return "password_history"
}

// PasswordPolicy 密码策略配置
type PasswordPolicy struct {
	ID                  string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	MinLength           int            `json:"min_length" gorm:"default:8"`
	MaxLength           int            `json:"max_length" gorm:"default:128"`
	RequireUppercase    bool           `json:"require_uppercase" gorm:"default:true"`
	RequireLowercase    bool           `json:"require_lowercase" gorm:"default:true"`
	RequireNumbers      bool           `json:"require_numbers" gorm:"default:true"`
	RequireSpecialChars bool           `json:"require_special_chars" gorm:"default:true"`
	HistoryCount        int            `json:"history_count" gorm:"default:5"`
	ExpiryDays          int            `json:"expiry_days" gorm:"default:90"`
	PreventCommon       bool           `json:"prevent_common" gorm:"default:true"`
	PreventUsername     bool           `json:"prevent_username" gorm:"default:true"`
	IsActive            bool           `json:"is_active" gorm:"default:true"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName 指定表名
func (PasswordPolicy) TableName() string {
	return "password_policies"
}
