package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
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

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	TotalUsers         int `json:"totalUsers"`
	TotalOrganizations int `json:"totalOrganizations"`
	ActiveSessions     int `json:"activeSessions"`
	TotalApplications  int `json:"totalApplications"`
}

// RecentActivity represents a recent activity
type RecentActivity struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Time        string    `json:"time"`
	User        string    `json:"user"`
	Action      string    `json:"action"`
	Resource    string    `json:"resource"`
	Status      string    `json:"status"`
	IPAddress   string    `json:"ip_address"`
	CreatedAt   time.Time `json:"created_at"`
}

// SystemStatus represents system component status
type SystemStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Value   string `json:"value"`
	Details any    `json:"details,omitempty"`
}

// DashboardData represents complete dashboard data
type DashboardData struct {
	Stats            DashboardStats   `json:"stats"`
	RecentActivities []RecentActivity `json:"recentActivities"`
	SystemStatus     []SystemStatus   `json:"systemStatus"`
}

// GetDashboardData returns dashboard statistics and recent activities
func GetDashboardData(c *gin.Context) {
	var stats DashboardStats
	var activities []RecentActivity
	var systemStatus []SystemStatus

	// Get user count
	var userCount int64
	if err := database.DB.Model(&models.User{}).Count(&userCount).Error; err != nil {
		logger.ErrorError("Failed to count users", zap.Error(err))
	}
	stats.TotalUsers = int(userCount)

	// Get organization count
	var orgCount int64
	if err := database.DB.Model(&models.Organization{}).Count(&orgCount).Error; err != nil {
		logger.ErrorError("Failed to count organizations", zap.Error(err))
	}
	stats.TotalOrganizations = int(orgCount)

	// Get application count (placeholder for now)
	stats.TotalApplications = 0

	// Get active sessions (placeholder for now)
	stats.ActiveSessions = 1

	// Get recent login activities
	var loginLogs []models.UserLoginLog
	if err := database.DB.Preload("User").Order("created_at DESC").Limit(10).Find(&loginLogs).Error; err != nil {
		logger.ErrorError("Failed to get login logs", zap.Error(err))
	}

	for _, log := range loginLogs {
		userName := log.UserID
		if log.User.Username != "" {
			userName = log.User.Username
		}

		activity := RecentActivity{
			ID:          log.ID,
			Title:       "User Login",
			Description: userName + " logged in",
			Icon:        "user",
			Time:        timeAgo(log.CreatedAt),
			User:        userName,
			Action:      "login",
			Resource:    "system",
			Status:      boolToString(log.Success),
			IPAddress:   log.LoginIP,
			CreatedAt:   log.CreatedAt,
		}
		activities = append(activities, activity)
	}

	// Get recent audit activities (skip if table doesn't exist)
	var auditLogs []models.AuditLog
	if err := database.DB.Preload("User").Order("created_at DESC").Limit(5).Find(&auditLogs).Error; err != nil {
		// 如果audit_logs表不存在，忽略错误继续执行
		logger.AccessInfo("Audit logs table not available, skipping audit activities", zap.Error(err))
	} else {
		for _, log := range auditLogs {
			userName := log.UserID
			if log.User.Username != "" {
				userName = log.User.Username
			}

			activity := RecentActivity{
				ID:          log.ID,
				Title:       log.Action + " " + log.Resource,
				Description: userName + " " + log.Description,
				Icon:        "audit",
				Time:        timeAgo(log.CreatedAt),
				User:        userName,
				Action:      log.Action,
				Resource:    log.Resource,
				Status:      log.Status,
				IPAddress:   log.IPAddress,
				CreatedAt:   log.CreatedAt,
			}
			activities = append(activities, activity)
		}
	}

	// Sort activities by creation time (newest first)
	sort.Slice(activities, func(i, j int) bool {
		return activities[i].CreatedAt.After(activities[j].CreatedAt)
	})

	// Limit to 10 most recent activities
	if len(activities) > 10 {
		activities = activities[:10]
	}

	// System status
	systemStatus = []SystemStatus{
		{
			Name:   "Database",
			Status: "success",
			Value:  "Connected",
		},
		{
			Name:   "Redis",
			Status: "success",
			Value:  "Connected",
		},
		{
			Name:   "API Server",
			Status: "success",
			Value:  "Running",
		},
	}

	dashboardData := DashboardData{
		Stats:            stats,
		RecentActivities: activities,
		SystemStatus:     systemStatus,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Dashboard data retrieved successfully",
		"data":    dashboardData,
	})
}

// GetSystemStats returns system statistics
func GetSystemStats(c *gin.Context) {
	var stats DashboardStats

	// Get user count
	var userCount int64
	if err := database.DB.Model(&models.User{}).Count(&userCount).Error; err != nil {
		logger.ErrorError("Failed to count users", zap.Error(err))
	}
	stats.TotalUsers = int(userCount)

	// Get organization count
	var orgCount int64
	if err := database.DB.Model(&models.Organization{}).Count(&orgCount).Error; err != nil {
		logger.ErrorError("Failed to count organizations", zap.Error(err))
	}
	stats.TotalOrganizations = int(orgCount)

	// Placeholder values
	stats.TotalApplications = 0
	stats.ActiveSessions = 1

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "System stats retrieved successfully",
		"data":    stats,
	})
}

// GetRecentActivities returns recent activities
func GetRecentActivities(c *gin.Context) {
	limit := 10 // default limit

	var loginLogs []models.UserLoginLog
	if err := database.DB.Order("created_at DESC").Limit(limit).Find(&loginLogs).Error; err != nil {
		logger.ErrorError("Failed to get login logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	var activities []RecentActivity
	for _, log := range loginLogs {
		activity := RecentActivity{
			ID:          log.ID,
			Title:       "User Login",
			Description: log.UserID + " logged in",
			Icon:        "user",
			Time:        timeAgo(log.CreatedAt),
			User:        log.UserID,
			Action:      "login",
			Resource:    "system",
			Status:      boolToString(log.Success),
			IPAddress:   log.LoginIP,
			CreatedAt:   log.CreatedAt,
		}
		activities = append(activities, activity)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Recent activities retrieved successfully",
		"data":    activities,
	})
}

// boolToString converts bool to string
func boolToString(b bool) string {
	if b {
		return "success"
	}
	return "failed"
}

// timeAgo returns a human-readable time ago string
func timeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return string(rune(minutes)) + " minutes ago"
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return string(rune(hours)) + " hours ago"
	} else {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return string(rune(days)) + " days ago"
	}
}

// GetPublicSiteInfoHandler 获取公开站点信息（不需要认证）
func GetPublicSiteInfoHandler(c *gin.Context) {
	var siteNameSetting, logoSetting models.SystemSetting

	// 获取站点名称
	if err := database.DB.Where("category = ? AND `key` = ?", "site", "site_name").First(&siteNameSetting).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			logger.ErrorError("Failed to get site name setting", zap.Error(err))
		}
	}

	// 获取logo URL
	if err := database.DB.Where("category = ? AND `key` = ?", "site", "logo_url").First(&logoSetting).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			logger.ErrorError("Failed to get logo setting", zap.Error(err))
		}
	}

	// 设置默认值
	siteName := "EIAM"
	logoUrl := "/logo.svg"

	if siteNameSetting.Value != "" {
		siteName = siteNameSetting.Value
	}

	if logoSetting.Value != "" {
		logoUrl = logoSetting.Value
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data": gin.H{
			"site_name": siteName,
			"logo_url":  logoUrl,
		},
	})
}
