package models

import (
	"time"
)

// User user model
type User struct {
	BaseModel
	Username    string     `json:"username" gorm:"type:varchar(50);uniqueIndex;not null" validate:"required,min=3,max=50"`
	Email       string     `json:"email" gorm:"type:varchar(100);uniqueIndex;not null" validate:"required,email"`
	Phone       string     `json:"phone" gorm:"type:varchar(20);index"`
	Password    string     `json:"-" gorm:"type:varchar(255);not null"`
	Salt        string     `json:"-" gorm:"type:varchar(255);not null"`
	DisplayName string     `json:"display_name" gorm:"type:varchar(100)"`
	Avatar      string     `json:"avatar" gorm:"type:varchar(500)"`
	Gender      Gender     `json:"gender" gorm:"type:tinyint;default:0"`
	Birthday    *time.Time `json:"birthday" gorm:"type:date"`
	Status      Status     `json:"status" gorm:"type:tinyint;default:1;index"`
	LastLoginAt *time.Time `json:"last_login_at"`
	LastLoginIP string     `json:"last_login_ip" gorm:"type:varchar(45)"`
	LoginCount  int64      `json:"login_count" gorm:"default:0"`
	FailedCount int        `json:"failed_count" gorm:"default:0"`
	LockedUntil *time.Time `json:"locked_until"`

	// Authentication related
	EmailVerified   bool       `json:"email_verified" gorm:"default:false"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	PhoneVerified   bool       `json:"phone_verified" gorm:"default:false"`
	PhoneVerifiedAt *time.Time `json:"phone_verified_at"`

	// Security settings
	EnableOTP   bool   `json:"enable_otp" gorm:"default:false"`
	OTPSecret   string `json:"-" gorm:"type:varchar(255)"` // TOTP secret key
	BackupCodes string `json:"-" gorm:"type:text"`         // Backup codes in JSON format

	// Password policy
	PasswordExpiredAt  *time.Time `json:"password_expired_at"`
	MustChangePassword bool       `json:"must_change_password" gorm:"default:false"`

	// Organization
	OrganizationID string `json:"organization_id" gorm:"type:varchar(36);index"`

	// Relationships
	Organization *Organization   `json:"organization" gorm:"foreignKey:OrganizationID"`
	Roles        []Role          `json:"roles" gorm:"many2many:user_roles;"`
	Groups       []Group         `json:"groups" gorm:"many2many:user_groups;"`
	Applications []Application   `json:"applications" gorm:"many2many:user_applications;"`
	Sessions     []UserSession   `json:"sessions" gorm:"foreignKey:UserID"`
	Profiles     []UserProfile   `json:"profiles" gorm:"foreignKey:UserID"`
	OTPRecords   []UserOTPRecord `json:"otp_records" gorm:"foreignKey:UserID"`
}

// TableName specify table name
func (User) TableName() string {
	return "users"
}

// UserProfile user extended configuration
type UserProfile struct {
	BaseModel
	UserID      string `json:"user_id" gorm:"type:varchar(36);not null;index"`
	Key         string `json:"key" gorm:"type:varchar(100);not null"`
	Value       string `json:"value" gorm:"type:text"`
	Description string `json:"description" gorm:"type:varchar(500)"`

	// Relationships
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// TableName specify table name
func (UserProfile) TableName() string {
	return "user_profiles"
}

// UserSession user session
type UserSession struct {
	BaseModel
	UserID       string    `json:"user_id" gorm:"type:varchar(36);not null;index"`
	SessionID    string    `json:"session_id" gorm:"type:varchar(255);uniqueIndex;not null"`
	AccessToken  string    `json:"access_token" gorm:"type:text"`
	RefreshToken string    `json:"refresh_token" gorm:"type:text"`
	ExpiresAt    time.Time `json:"expires_at" gorm:"not null"`
	UserAgent    string    `json:"user_agent" gorm:"type:varchar(500)"`
	ClientIP     string    `json:"client_ip" gorm:"type:varchar(45)"`
	DeviceType   string    `json:"device_type" gorm:"type:varchar(50)"`
	Location     string    `json:"location" gorm:"type:varchar(200)"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`

	// Relationships
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// TableName specify table name
func (UserSession) TableName() string {
	return "user_sessions"
}

// UserLoginLog user login log
type UserLoginLog struct {
	BaseModel
	UserID     string `json:"user_id" gorm:"type:varchar(36);not null;index"`
	LoginType  string `json:"login_type" gorm:"type:varchar(50);not null"` // password, sso, oauth, saml
	LoginIP    string `json:"login_ip" gorm:"type:varchar(45)"`
	UserAgent  string `json:"user_agent" gorm:"type:varchar(500)"`
	DeviceType string `json:"device_type" gorm:"type:varchar(50)"`
	Location   string `json:"location" gorm:"type:varchar(200)"`
	Success    bool   `json:"success" gorm:"default:false"`
	FailReason string `json:"fail_reason" gorm:"type:varchar(500)"`
	Duration   int64  `json:"duration"` // Login duration (milliseconds)

	// Relationships
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// TableName specify table name
func (UserLoginLog) TableName() string {
	return "user_login_logs"
}

// UserOTPRecord user OTP record
type UserOTPRecord struct {
	BaseModel
	UserID      string    `json:"user_id" gorm:"type:varchar(36);not null;index"`
	Code        string    `json:"code" gorm:"type:varchar(32);not null"`
	Purpose     string    `json:"purpose" gorm:"type:varchar(50);not null"` // login, reset_password, verify_email
	ExpiresAt   time.Time `json:"expires_at" gorm:"not null"`
	Attempts    int       `json:"attempts" gorm:"default:0"`
	MaxAttempts int       `json:"max_attempts" gorm:"default:3"`
	Used        bool      `json:"used" gorm:"default:false"`
	SentTo      string    `json:"sent_to" gorm:"type:varchar(255)"` // Send target (email or phone)

	// Relationships
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// TableName specify table name
func (UserOTPRecord) TableName() string {
	return "user_otp_records"
}
