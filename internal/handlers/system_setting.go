package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"eiam-platform/internal/models"
	"eiam-platform/pkg/database"
	"eiam-platform/pkg/i18n"
	"eiam-platform/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// GetSystemSettingsHandler 获取系统设置
func GetSystemSettingsHandler(c *gin.Context) {
	category := c.Query("category")

	var settings []models.SystemSetting
	query := database.DB.Model(&models.SystemSetting{})

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if err := query.Find(&settings).Error; err != nil {
		logger.ErrorError("Failed to get system settings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// 转换为键值对格式
	settingsMap := make(map[string]interface{})
	for _, setting := range settings {
		value := convertValue(setting.Value, setting.Type)
		settingsMap[setting.Key] = value
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": i18n.Success,
		"data":    settingsMap,
	})
}

// UpdateSystemSettingsHandler 更新系统设置
func UpdateSystemSettingsHandler(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": i18n.InvalidRequestData,
			"data":    nil,
		})
		return
	}

	// 批量更新设置
	for key, value := range req {
		valueStr := convertToString(value)

		var setting models.SystemSetting
		result := database.DB.Where("key = ?", key).First(&setting)

		if result.Error == gorm.ErrRecordNotFound {
			// 创建新设置
			setting = models.SystemSetting{
				ID:        uuid.New().String(),
				Key:       key,
				Value:     valueStr,
				Type:      getTypeFromValue(value),
				Category:  getCategoryFromKey(key),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := database.DB.Create(&setting).Error; err != nil {
				logger.ErrorError("Failed to create setting", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": i18n.InternalServerError,
					"data":    nil,
				})
				return
			}
		} else if result.Error != nil {
			logger.ErrorError("Failed to get setting", zap.Error(result.Error))
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": i18n.InternalServerError,
				"data":    nil,
			})
			return
		} else {
			// 更新现有设置
			if err := database.DB.Model(&setting).Updates(map[string]interface{}{
				"value":      valueStr,
				"type":       getTypeFromValue(value),
				"updated_at": time.Now(),
			}).Error; err != nil {
				logger.ErrorError("Failed to update setting", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": i18n.InternalServerError,
					"data":    nil,
				})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": i18n.Success,
		"data":    nil,
	})
}

// GetSiteSettingsHandler 获取站点设置
func GetSiteSettingsHandler(c *gin.Context) {
	var settings []models.SystemSetting
	if err := database.DB.Where("category = ?", "site").Find(&settings).Error; err != nil {
		logger.ErrorError("Failed to get site settings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// 转换为SiteSettings结构
	siteSettings := models.SiteSettings{}
	for _, setting := range settings {
		value := convertValue(setting.Value, setting.Type)
		switch setting.Key {
		case "site_name":
			if str, ok := value.(string); ok {
				siteSettings.SiteName = str
			}
		case "site_url":
			if str, ok := value.(string); ok {
				siteSettings.SiteURL = str
			}
		case "contact_email":
			if str, ok := value.(string); ok {
				siteSettings.ContactEmail = str
			}
		case "support_email":
			if str, ok := value.(string); ok {
				siteSettings.SupportEmail = str
			}
		case "description":
			if str, ok := value.(string); ok {
				siteSettings.Description = str
			}
		case "logo":
			if str, ok := value.(string); ok {
				siteSettings.Logo = str
			}
		case "favicon":
			if str, ok := value.(string); ok {
				siteSettings.Favicon = str
			}
		case "footer_text":
			if str, ok := value.(string); ok {
				siteSettings.FooterText = str
			}
		case "maintenance_mode":
			if b, ok := value.(bool); ok {
				siteSettings.MaintenanceMode = b
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": i18n.Success,
		"data":    siteSettings,
	})
}

// GetSecuritySettingsHandler 获取安全设置
func GetSecuritySettingsHandler(c *gin.Context) {
	var settings []models.SystemSetting
	if err := database.DB.Where("category = ?", "security").Find(&settings).Error; err != nil {
		logger.ErrorError("Failed to get security settings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// 转换为SecuritySettings结构
	securitySettings := models.SecuritySettings{}
	for _, setting := range settings {
		value := convertValue(setting.Value, setting.Type)
		switch setting.Key {
		case "min_password_length":
			if num, ok := value.(int); ok {
				securitySettings.MinPasswordLength = num
			}
		case "max_password_length":
			if num, ok := value.(int); ok {
				securitySettings.MaxPasswordLength = num
			}
		case "password_expiry_days":
			if num, ok := value.(int); ok {
				securitySettings.PasswordExpiryDays = num
			}
		case "require_uppercase":
			if b, ok := value.(bool); ok {
				securitySettings.RequireUppercase = b
			}
		case "require_lowercase":
			if b, ok := value.(bool); ok {
				securitySettings.RequireLowercase = b
			}
		case "require_numbers":
			if b, ok := value.(bool); ok {
				securitySettings.RequireNumbers = b
			}
		case "require_special_chars":
			if b, ok := value.(bool); ok {
				securitySettings.RequireSpecialChars = b
			}
		case "password_history_count":
			if num, ok := value.(int); ok {
				securitySettings.PasswordHistoryCount = num
			}
		case "session_timeout":
			if num, ok := value.(int); ok {
				securitySettings.SessionTimeout = num
			}
		case "max_concurrent_sessions":
			if num, ok := value.(int); ok {
				securitySettings.MaxConcurrentSessions = num
			}
		case "remember_me_days":
			if num, ok := value.(int); ok {
				securitySettings.RememberMeDays = num
			}
		case "enable_2fa":
			if b, ok := value.(bool); ok {
				securitySettings.Enable2FA = b
			}
		case "require_2fa_for_admins":
			if b, ok := value.(bool); ok {
				securitySettings.Require2FAForAdmins = b
			}
		case "allow_backup_codes":
			if b, ok := value.(bool); ok {
				securitySettings.AllowBackupCodes = b
			}
		case "enable_totp":
			if b, ok := value.(bool); ok {
				securitySettings.EnableTOTP = b
			}
		case "enable_sms":
			if b, ok := value.(bool); ok {
				securitySettings.EnableSMS = b
			}
		case "enable_email":
			if b, ok := value.(bool); ok {
				securitySettings.EnableEmail = b
			}
		case "max_login_attempts":
			if num, ok := value.(int); ok {
				securitySettings.MaxLoginAttempts = num
			}
		case "lockout_duration":
			if num, ok := value.(int); ok {
				securitySettings.LockoutDuration = num
			}
		case "enable_ip_whitelist":
			if b, ok := value.(bool); ok {
				securitySettings.EnableIPWhitelist = b
			}
		case "enable_geolocation":
			if b, ok := value.(bool); ok {
				securitySettings.EnableGeolocation = b
			}
		case "enable_device_fingerprinting":
			if b, ok := value.(bool); ok {
				securitySettings.EnableDeviceFingerprinting = b
			}
		case "notify_failed_logins":
			if b, ok := value.(bool); ok {
				securitySettings.NotifyFailedLogins = b
			}
		case "notify_new_devices":
			if b, ok := value.(bool); ok {
				securitySettings.NotifyNewDevices = b
			}
		case "notify_password_changes":
			if b, ok := value.(bool); ok {
				securitySettings.NotifyPasswordChanges = b
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": i18n.Success,
		"data":    securitySettings,
	})
}

// Helper functions
func convertValue(value string, valueType string) interface{} {
	switch valueType {
	case "number":
		if num, err := strconv.Atoi(value); err == nil {
			return num
		}
		return 0
	case "boolean":
		if value == "true" {
			return true
		}
		return false
	case "json":
		var result interface{}
		if err := json.Unmarshal([]byte(value), &result); err == nil {
			return result
		}
		return value
	default:
		return value
	}
}

func convertToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int32, int64:
		return strconv.Itoa(v.(int))
	case bool:
		return strconv.FormatBool(v)
	default:
		if jsonBytes, err := json.Marshal(v); err == nil {
			return string(jsonBytes)
		}
		return ""
	}
}

func getTypeFromValue(value interface{}) string {
	switch value.(type) {
	case string:
		return "string"
	case int, int32, int64:
		return "number"
	case bool:
		return "boolean"
	default:
		return "json"
	}
}

func getCategoryFromKey(key string) string {
	if len(key) >= 4 && key[:4] == "site" {
		return "site"
	}
	if len(key) >= 8 && key[:8] == "security" {
		return "security"
	}
	if len(key) >= 5 && key[:5] == "email" {
		return "email"
	}
	return "other"
}
