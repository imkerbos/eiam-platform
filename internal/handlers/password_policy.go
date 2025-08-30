package handlers

import (
	"net/http"
	"time"

	"eiam-platform/internal/models"
	"eiam-platform/pkg/database"
	"eiam-platform/pkg/logger"
	"eiam-platform/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PasswordPolicyRequest 密码策略请求
type PasswordPolicyRequest struct {
	MinLength           int  `json:"min_length" binding:"required,min=6,max=128"`
	MaxLength           int  `json:"max_length" binding:"required,min=8,max=256"`
	RequireUppercase    bool `json:"require_uppercase"`
	RequireLowercase    bool `json:"require_lowercase"`
	RequireNumbers      bool `json:"require_numbers"`
	RequireSpecialChars bool `json:"require_special_chars"`
	HistoryCount        int  `json:"history_count" binding:"min=0,max=20"`
	ExpiryDays          int  `json:"expiry_days" binding:"min=0,max=3650"`
	PreventCommon       bool `json:"prevent_common"`
	PreventUsername     bool `json:"prevent_username"`
	IsActive            bool `json:"is_active"`
}

// PasswordValidationRequest 密码验证请求
type PasswordValidationRequest struct {
	Password string `json:"password" binding:"required"`
	Username string `json:"username"`
}

// PasswordValidationResponse 密码验证响应
type PasswordValidationResponse struct {
	Valid         bool     `json:"valid"`
	Errors        []string `json:"errors"`
	Warnings      []string `json:"warnings"`
	Strength      int      `json:"strength"`
	StrengthText  string   `json:"strength_text"`
	StrengthColor string   `json:"strength_color"`
}

// GetPasswordPolicyHandler 获取密码策略
func GetPasswordPolicyHandler(c *gin.Context) {
	var policy models.PasswordPolicy

	// 获取当前激活的密码策略
	if err := database.DB.Where("is_active = ?", true).First(&policy).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果没有找到策略，返回默认策略
			defaultPolicy := utils.DefaultPasswordPolicy()
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "Success",
				"data": gin.H{
					"min_length":            defaultPolicy.MinLength,
					"max_length":            defaultPolicy.MaxLength,
					"require_uppercase":     defaultPolicy.RequireUppercase,
					"require_lowercase":     defaultPolicy.RequireLowercase,
					"require_numbers":       defaultPolicy.RequireNumbers,
					"require_special_chars": defaultPolicy.RequireSpecial,
					"history_count":         defaultPolicy.HistoryCount,
					"expiry_days":           defaultPolicy.ExpiryDays,
					"prevent_common":        defaultPolicy.PreventCommon,
					"prevent_username":      defaultPolicy.PreventUsername,
				},
			})
			return
		}

		logger.Error("Failed to get password policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to get password policy",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data":    policy,
	})
}

// UpdatePasswordPolicyHandler 更新密码策略
func UpdatePasswordPolicyHandler(c *gin.Context) {
	var req PasswordPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	// 验证策略参数
	if req.MinLength > req.MaxLength {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Minimum length cannot be greater than maximum length",
		})
		return
	}

	// 至少需要启用一种字符类型要求
	if !req.RequireUppercase && !req.RequireLowercase && !req.RequireNumbers && !req.RequireSpecialChars {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "At least one character type requirement must be enabled",
		})
		return
	}

	var policy models.PasswordPolicy

	// 查找现有策略或创建新策略
	if err := database.DB.Where("is_active = ?", true).First(&policy).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新策略
			policy = models.PasswordPolicy{
				ID: utils.GenerateTradeIDString("policy"),
			}
		} else {
			logger.Error("Failed to get password policy", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "Failed to get password policy",
			})
			return
		}
	}

	// 更新策略
	policy.MinLength = req.MinLength
	policy.MaxLength = req.MaxLength
	policy.RequireUppercase = req.RequireUppercase
	policy.RequireLowercase = req.RequireLowercase
	policy.RequireNumbers = req.RequireNumbers
	policy.RequireSpecialChars = req.RequireSpecialChars
	policy.HistoryCount = req.HistoryCount
	policy.ExpiryDays = req.ExpiryDays
	policy.PreventCommon = req.PreventCommon
	policy.PreventUsername = req.PreventUsername
	policy.IsActive = req.IsActive

	if err := database.DB.Save(&policy).Error; err != nil {
		logger.Error("Failed to update password policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to update password policy",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Password policy updated successfully",
		"data":    policy,
	})
}

// ValidatePasswordHandler 验证密码
func ValidatePasswordHandler(c *gin.Context) {
	var req PasswordValidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	// 获取当前密码策略
	var policy models.PasswordPolicy
	if err := database.DB.Where("is_active = ?", true).First(&policy).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 使用默认策略
			defaultPolicy := utils.DefaultPasswordPolicy()
			validationResult := utils.ValidatePassword(req.Password, defaultPolicy, req.Username, nil)

			strength := utils.CalculatePasswordStrength(req.Password)
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "Success",
				"data": PasswordValidationResponse{
					Valid:         validationResult.Valid,
					Errors:        validationResult.Errors,
					Warnings:      validationResult.Warnings,
					Strength:      strength,
					StrengthText:  utils.GetPasswordStrengthText(strength),
					StrengthColor: utils.GetPasswordStrengthColor(strength),
				},
			})
			return
		}

		logger.Error("Failed to get password policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to get password policy",
		})
		return
	}

	// 转换为utils.PasswordPolicy
	utilsPolicy := &utils.PasswordPolicy{
		MinLength:        policy.MinLength,
		MaxLength:        policy.MaxLength,
		RequireUppercase: policy.RequireUppercase,
		RequireLowercase: policy.RequireLowercase,
		RequireNumbers:   policy.RequireNumbers,
		RequireSpecial:   policy.RequireSpecialChars,
		HistoryCount:     policy.HistoryCount,
		ExpiryDays:       policy.ExpiryDays,
		PreventCommon:    policy.PreventCommon,
		PreventUsername:  policy.PreventUsername,
	}

	// 获取用户密码历史（如果提供了用户名）
	var passwordHistory []string
	if req.Username != "" {
		var user models.User
		if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err == nil {
			var histories []models.PasswordHistory
			if err := database.DB.Where("user_id = ?", user.ID).Order("created_at DESC").Limit(policy.HistoryCount).Find(&histories).Error; err == nil {
				for _, history := range histories {
					passwordHistory = append(passwordHistory, history.Password)
				}
			}
		}
	}

	// 验证密码
	validationResult := utils.ValidatePassword(req.Password, utilsPolicy, req.Username, passwordHistory)
	strength := utils.CalculatePasswordStrength(req.Password)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data": PasswordValidationResponse{
			Valid:         validationResult.Valid,
			Errors:        validationResult.Errors,
			Warnings:      validationResult.Warnings,
			Strength:      strength,
			StrengthText:  utils.GetPasswordStrengthText(strength),
			StrengthColor: utils.GetPasswordStrengthColor(strength),
		},
	})
}

// GeneratePasswordHandler 生成强密码
func GeneratePasswordHandler(c *gin.Context) {
	// 获取当前密码策略
	var policy models.PasswordPolicy
	if err := database.DB.Where("is_active = ?", true).First(&policy).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 使用默认策略
			defaultPolicy := utils.DefaultPasswordPolicy()
			password, err := utils.GenerateStrongPassword(defaultPolicy)
			if err != nil {
				logger.Error("Failed to generate password", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "Failed to generate password",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "Success",
				"data": gin.H{
					"password": password,
				},
			})
			return
		}

		logger.Error("Failed to get password policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to get password policy",
		})
		return
	}

	// 转换为utils.PasswordPolicy
	utilsPolicy := &utils.PasswordPolicy{
		MinLength:        policy.MinLength,
		MaxLength:        policy.MaxLength,
		RequireUppercase: policy.RequireUppercase,
		RequireLowercase: policy.RequireLowercase,
		RequireNumbers:   policy.RequireNumbers,
		RequireSpecial:   policy.RequireSpecialChars,
		HistoryCount:     policy.HistoryCount,
		ExpiryDays:       policy.ExpiryDays,
		PreventCommon:    policy.PreventCommon,
		PreventUsername:  policy.PreventUsername,
	}

	password, err := utils.GenerateStrongPassword(utilsPolicy)
	if err != nil {
		logger.Error("Failed to generate password", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to generate password",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data": gin.H{
			"password": password,
		},
	})
}

// SavePasswordHistory 保存密码历史
func SavePasswordHistory(userID, password string) error {
	history := models.PasswordHistory{
		ID:       utils.GenerateTradeIDString("pwd"),
		UserID:   userID,
		Password: password,
	}

	return database.DB.Create(&history).Error
}

// CleanOldPasswordHistory 清理旧的密码历史
func CleanOldPasswordHistory(userID string, keepCount int) error {
	var count int64
	database.DB.Model(&models.PasswordHistory{}).Where("user_id = ?", userID).Count(&count)

	if int(count) > keepCount {
		// 删除多余的记录，保留最新的
		var histories []models.PasswordHistory
		if err := database.DB.Where("user_id = ?", userID).Order("created_at DESC").Limit(keepCount).Find(&histories).Error; err != nil {
			return err
		}

		if len(histories) > 0 {
			// 删除不在保留列表中的记录
			var ids []string
			for _, history := range histories {
				ids = append(ids, history.ID)
			}

			return database.DB.Where("user_id = ? AND id NOT IN ?", userID, ids).Delete(&models.PasswordHistory{}).Error
		}
	}

	return nil
}

// CheckPasswordExpiry 检查密码是否过期
func CheckPasswordExpiry(user *models.User) (bool, error) {
	if user.PasswordExpiredAt == nil {
		return false, nil
	}

	return time.Now().After(*user.PasswordExpiredAt), nil
}

// SetPasswordExpiry 设置密码过期时间
func SetPasswordExpiry(userID string, expiryDays int) error {
	expiryTime := time.Now().AddDate(0, 0, expiryDays)
	return database.DB.Model(&models.User{}).Where("id = ?", userID).Update("password_expired_at", expiryTime).Error
}
