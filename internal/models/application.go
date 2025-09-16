package models

import "time"

// ApplicationGroup application group
type ApplicationGroup struct {
	BaseModel
	Name        string `json:"name" gorm:"type:varchar(100);not null" validate:"required,max=100"`
	Code        string `json:"code" gorm:"type:varchar(50);uniqueIndex;not null" validate:"required,max=50"`
	Description string `json:"description" gorm:"type:varchar(500)"`
	Sort        int    `json:"sort" gorm:"default:0"`
	Icon        string `json:"icon" gorm:"type:varchar(200)"`
	Color       string `json:"color" gorm:"type:varchar(20);default:'#1890ff'"`
	Status      Status `json:"status" gorm:"type:tinyint;default:1;index"`

	// 关联关系
	Applications []Application `json:"applications" gorm:"foreignKey:GroupID"`
	Roles        []Role        `json:"roles" gorm:"many2many:role_application_groups;"`
}

// TableName 指定表名
func (ApplicationGroup) TableName() string {
	return "application_groups"
}

// Application 应用系统
type Application struct {
	BaseModel
	Name         string  `json:"name" gorm:"type:varchar(100);not null" validate:"required,max=100"`
	Code         string  `json:"code" gorm:"type:varchar(50);uniqueIndex;not null" validate:"required,max=50"`
	Description  string  `json:"description" gorm:"type:varchar(500)"`
	GroupID      *string `json:"group_id" gorm:"type:varchar(36);index"` // 应用分组ID
	ClientID     string  `json:"client_id" gorm:"type:varchar(100);uniqueIndex;not null"`
	ClientSecret string  `json:"client_secret" gorm:"type:varchar(255);not null"`
	RedirectURIs string  `json:"redirect_uris" gorm:"type:text"` // JSON数组格式
	LogoutURI    string  `json:"logout_uri" gorm:"type:varchar(500)"`
	HomePageURL  string  `json:"home_page_url" gorm:"type:varchar(500)"`
	Logo         string  `json:"logo" gorm:"type:varchar(500)"`
	AppType      string  `json:"app_type" gorm:"type:varchar(50);not null"` // web, mobile, api
	Protocol     string  `json:"protocol" gorm:"type:varchar(50);not null"` // oauth2, saml2, cas, oidc
	Sort         int     `json:"sort" gorm:"default:0"`
	Status       Status  `json:"status" gorm:"type:tinyint;default:1;index"`

	// OAuth2 特有配置
	GrantTypes      string `json:"grant_types" gorm:"type:varchar(500)"`    // authorization_code,refresh_token,client_credentials
	ResponseTypes   string `json:"response_types" gorm:"type:varchar(200)"` // code,token
	Scopes          string `json:"scopes" gorm:"type:varchar(500)"`         // openid,profile,email
	AccessTokenTTL  int    `json:"access_token_ttl" gorm:"default:3600"`    // 访问令牌过期时间(秒)
	RefreshTokenTTL int    `json:"refresh_token_ttl" gorm:"default:604800"` // 刷新令牌过期时间(秒)

	// SAML2 特有配置
	EntityID           string `json:"entity_id" gorm:"type:varchar(255)"`
	AcsURL             string `json:"acs_url" gorm:"type:varchar(500)"`
	SloURL             string `json:"slo_url" gorm:"type:varchar(500)"`
	Certificate        string `json:"certificate" gorm:"type:text"`
	SignatureAlgorithm string `json:"signature_algorithm" gorm:"type:varchar(100)"`
	DigestAlgorithm    string `json:"digest_algorithm" gorm:"type:varchar(100)"`

	// CAS 特有配置
	ServiceURL      string `json:"service_url" gorm:"type:varchar(500)"`
	Gateway         bool   `json:"gateway" gorm:"default:false"`
	Renew           bool   `json:"renew" gorm:"default:false"`
	AttributeMapping string `json:"attribute_mapping" gorm:"type:text"` // JSON格式的属性映射配置

	// LDAP 特有配置
	LdapURL      string `json:"ldap_url" gorm:"type:varchar(500)"`
	BaseDN       string `json:"base_dn" gorm:"type:varchar(500)"`
	BindDN       string `json:"bind_dn" gorm:"type:varchar(500)"`
	BindPassword string `json:"bind_password" gorm:"type:varchar(500)"`

	// 关联关系
	Group       *ApplicationGroup `json:"group" gorm:"foreignKey:GroupID"`
	Users       []User            `json:"users" gorm:"many2many:user_applications;"`
	Roles       []Role            `json:"roles" gorm:"many2many:role_applications;"`
	Permissions []Permission      `json:"permissions" gorm:"foreignKey:ApplicationID"`

	// OAuth2和SAML相关表（二期实现时启用外键约束）
	// OAuth2AuthorizationCodes []OAuth2AuthorizationCode `json:"oauth2_authorization_codes" gorm:"foreignKey:ClientID;references:ClientID"`
	// OAuth2AccessTokens       []OAuth2AccessToken       `json:"oauth2_access_tokens" gorm:"foreignKey:ClientID;references:ClientID"`
	// SAMLAssertions           []SAMLAssertion           `json:"saml_assertions" gorm:"foreignKey:ClientID;references:ClientID"`
}

// TableName 指定表名
func (Application) TableName() string {
	return "applications"
}

// OAuth2AuthorizationCode OAuth2授权码
type OAuth2AuthorizationCode struct {
	BaseModel
	Code            string    `json:"code" gorm:"type:varchar(255);uniqueIndex;not null"`
	ClientID        string    `json:"client_id" gorm:"type:varchar(100);not null;index"`
	UserID          string    `json:"user_id" gorm:"type:varchar(36);not null;index"`
	RedirectURI     string    `json:"redirect_uri" gorm:"type:varchar(500);not null"`
	Scope           string    `json:"scope" gorm:"type:varchar(500)"`
	State           string    `json:"state" gorm:"type:varchar(255)"`
	Challenge       string    `json:"challenge" gorm:"type:varchar(255)"`       // PKCE
	ChallengeMethod string    `json:"challenge_method" gorm:"type:varchar(10)"` // S256, plain
	ExpiresAt       time.Time `json:"expires_at" gorm:"not null"`
	Used            bool      `json:"used" gorm:"default:false"`

	// 关联关系
	User        User        `json:"user" gorm:"foreignKey:UserID"`
	Application Application `json:"application" gorm:"foreignKey:ClientID;references:ClientID"`
}

// TableName 指定表名
func (OAuth2AuthorizationCode) TableName() string {
	return "oauth2_authorization_codes"
}

// OAuth2AccessToken OAuth2访问令牌
type OAuth2AccessToken struct {
	BaseModel
	AccessToken      string     `json:"access_token" gorm:"type:varchar(500);uniqueIndex;not null"`
	RefreshToken     string     `json:"refresh_token" gorm:"type:varchar(500);uniqueIndex"`
	ClientID         string     `json:"client_id" gorm:"type:varchar(100);not null;index"`
	UserID           string     `json:"user_id" gorm:"type:varchar(36);not null;index"`
	Scope            string     `json:"scope" gorm:"type:varchar(500)"`
	TokenType        string     `json:"token_type" gorm:"type:varchar(50);default:'Bearer'"`
	ExpiresAt        time.Time  `json:"expires_at" gorm:"not null"`
	RefreshExpiresAt *time.Time `json:"refresh_expires_at"`

	// 关联关系
	User        User        `json:"user" gorm:"foreignKey:UserID"`
	Application Application `json:"application" gorm:"foreignKey:ClientID;references:ClientID"`
}

// TableName 指定表名
func (OAuth2AccessToken) TableName() string {
	return "oauth2_access_tokens"
}

// SAMLAssertion SAML断言
type SAMLAssertion struct {
	BaseModel
	AssertionID  string    `json:"assertion_id" gorm:"type:varchar(255);uniqueIndex;not null"`
	UserID       string    `json:"user_id" gorm:"type:varchar(36);not null;index"`
	ClientID     string    `json:"client_id" gorm:"type:varchar(100);not null;index"`
	Recipient    string    `json:"recipient" gorm:"type:varchar(500);not null"`
	Audience     string    `json:"audience" gorm:"type:varchar(500);not null"`
	NameID       string    `json:"name_id" gorm:"type:varchar(255);not null"`
	NameIDFormat string    `json:"name_id_format" gorm:"type:varchar(255)"`
	SessionIndex string    `json:"session_index" gorm:"type:varchar(255)"`
	Attributes   string    `json:"attributes" gorm:"type:text"` // JSON格式
	ExpiresAt    time.Time `json:"expires_at" gorm:"not null"`
	Used         bool      `json:"used" gorm:"default:false"`

	// 关联关系
	User        User        `json:"user" gorm:"foreignKey:UserID"`
	Application Application `json:"application" gorm:"foreignKey:ClientID;references:ClientID"`
}

// TableName 指定表名
func (SAMLAssertion) TableName() string {
	return "saml_assertions"
}
