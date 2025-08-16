package handlers

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"eiam-platform/config"
	"eiam-platform/internal/models"
	"eiam-platform/pkg/database"
	"eiam-platform/pkg/i18n"
	"eiam-platform/pkg/logger"
	"eiam-platform/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Console authentication handlers
func ConsoleLoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.AccessInfo("Console login request validation failed",
			zap.String("ip", c.ClientIP()),
			zap.String("username", req.Username),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": i18n.InvalidRequestData,
			"data":    nil,
		})
		return
	}

	// 获取用户信息
	var user models.User
	if err := database.DB.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.AccessInfo("Console login failed: user not found",
				zap.String("ip", c.ClientIP()),
				zap.String("username", req.Username),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": i18n.InvalidCredentials,
				"data":    nil,
			})
			return
		}
		logger.ErrorError("Database error during console login",
			zap.String("ip", c.ClientIP()),
			zap.String("username", req.Username),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// 检查用户状态
	if user.Status != models.StatusActive {
		logger.AccessInfo("Console login failed: user inactive",
			zap.String("ip", c.ClientIP()),
			zap.String("username", user.Username),
			zap.String("status", user.Status.String()),
		)
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": i18n.UserInactive,
			"data":    nil,
		})
		return
	}

	// 验证密码
	if !utils.CheckPassword(req.Password, user.Password) {
		// 更新失败登录次数
		user.FailedCount++
		if user.FailedCount >= 5 {
			// 锁定账户30分钟
			lockTime := time.Now().Add(30 * time.Minute)
			user.LockedUntil = &lockTime
		}
		database.DB.Save(&user)

		logger.AccessInfo("Console login failed: invalid password",
			zap.String("ip", c.ClientIP()),
			zap.String("username", user.Username),
			zap.Int("failed_count", user.FailedCount),
		)
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": i18n.InvalidCredentials,
			"data":    nil,
		})
		return
	}

	// 检查账户是否被锁定
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		logger.AccessInfo("Console login failed: account locked",
			zap.String("ip", c.ClientIP()),
			zap.String("username", user.Username),
			zap.Time("locked_until", *user.LockedUntil),
		)
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": i18n.AccountLocked,
			"data":    nil,
		})
		return
	}

	// 生成JWT令牌
	cfg := config.GetConfig()
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Username, cfg.JWT.Secret, cfg.JWT.AccessTokenExpire)
	if err != nil {
		logger.ErrorError("Failed to generate access token",
			zap.String("ip", c.ClientIP()),
			zap.String("username", user.Username),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Username, cfg.JWT.Secret, cfg.JWT.RefreshTokenExpire)
	if err != nil {
		logger.ErrorError("Failed to generate refresh token",
			zap.String("ip", c.ClientIP()),
			zap.String("username", user.Username),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// 更新用户登录信息
	now := time.Now()
	user.LastLoginAt = &now
	user.LastLoginIP = c.ClientIP()
	user.LoginCount++
	user.FailedCount = 0   // 重置失败次数
	user.LockedUntil = nil // 清除锁定状态
	database.DB.Save(&user)

	// 记录登录日志
	loginLog := models.UserLoginLog{
		UserID:    user.ID,
		LoginType: "console_password",
		LoginIP:   c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		Success:   true,
	}
	database.DB.Create(&loginLog)

	logger.AccessInfo("Console login successful",
		zap.String("ip", c.ClientIP()),
		zap.String("username", user.Username),
		zap.String("user_id", user.ID),
	)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": i18n.LoginSuccess,
		"data": LoginResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    int64(cfg.JWT.AccessTokenExpire),
			RequireOTP:   false,
			User: UserInfo{
				ID:            user.ID,
				Username:      user.Username,
				Email:         user.Email,
				DisplayName:   user.DisplayName,
				Avatar:        user.Avatar,
				Status:        user.Status.String(),
				EmailVerified: user.EmailVerified,
				PhoneVerified: user.PhoneVerified,
				EnableOTP:     user.EnableOTP,
				LastLoginAt:   *user.LastLoginAt,
				LastLoginIP:   user.LastLoginIP,
			},
		},
	})
}

func ConsoleLogoutHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

func ConsoleRefreshTokenHandler(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.AccessInfo("Refresh token request validation failed",
			zap.String("ip", c.ClientIP()),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": i18n.InvalidRequestData,
			"data":    nil,
		})
		return
	}

	// 验证refresh token
	claims, err := utils.ValidateRefreshToken(req.RefreshToken, config.AppConfig.JWT.Secret)
	if err != nil {
		logger.AccessInfo("Refresh token validation failed",
			zap.String("ip", c.ClientIP()),
			zap.Error(err),
		)
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Invalid refresh token",
			"data":    nil,
		})
		return
	}

	// 获取用户信息
	var user models.User
	if err := database.DB.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		logger.ErrorError("Failed to get user for refresh token",
			zap.String("user_id", claims.UserID),
			zap.Error(err),
		)
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "User not found",
			"data":    nil,
		})
		return
	}

	// 检查用户状态
	if user.Status != models.StatusActive {
		logger.AccessInfo("Refresh token failed: user inactive",
			zap.String("ip", c.ClientIP()),
			zap.String("user_id", user.ID),
			zap.String("status", user.Status.String()),
		)
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "User account is inactive",
			"data":    nil,
		})
		return
	}

	// 生成新的access token
	accessToken, err := utils.GenerateAccessTokenWithUserInfo(user.ID, user.Username, user.Email, user.DisplayName, []string{}, []string{}, config.AppConfig.JWT.Secret, config.AppConfig.JWT.AccessTokenExpire)
	if err != nil {
		logger.ErrorError("Failed to generate access token for refresh",
			zap.String("user_id", user.ID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// 生成新的refresh token
	refreshToken, err := utils.GenerateRefreshToken(user.ID, claims.SessionID, config.AppConfig.JWT.Secret, config.AppConfig.JWT.RefreshTokenExpire)
	if err != nil {
		logger.ErrorError("Failed to generate new refresh token",
			zap.String("user_id", user.ID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	logger.AccessInfo("Token refreshed successfully",
		zap.String("ip", c.ClientIP()),
		zap.String("user_id", user.ID),
	)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Token refreshed successfully",
		"data": RefreshTokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    int64(config.AppConfig.JWT.AccessTokenExpire),
		},
	})
}

func ConsoleGetMeHandler(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "User not authenticated",
			"data":    nil,
		})
		return
	}

	var user models.User
	if err := database.DB.Preload("Organization").Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "User not found",
				"data":    nil,
			})
			return
		}
		logger.ErrorError("Failed to get user info", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// Return user info in UserInfo format
	userInfo := UserInfo{
		ID:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		DisplayName:    user.DisplayName,
		Avatar:         user.Avatar,
		Status:         user.Status.String(),
		EmailVerified:  user.EmailVerified,
		PhoneVerified:  user.PhoneVerified,
		EnableOTP:      user.EnableOTP,
		LastLoginIP:    user.LastLoginIP,
		OrganizationID: user.OrganizationID,
	}
	if user.LastLoginAt != nil {
		userInfo.LastLoginAt = *user.LastLoginAt
	}

	// 添加组织信息
	if user.Organization != nil {
		userInfo.Organization = &OrganizationSimpleInfo{
			ID:   user.Organization.ID,
			Name: user.Organization.Name,
			Code: user.Organization.Code,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data":    userInfo,
	})
}

// Portal authentication handlers
func PortalLoginHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

func PortalLogoutHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

func PortalRefreshTokenHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

func PortalGetMeHandler(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "User not authenticated",
			"data":    nil,
		})
		return
	}

	var user models.User
	if err := database.DB.Preload("Organization").Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "User not found",
				"data":    nil,
			})
			return
		}
		logger.ErrorError("Failed to get user info", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// Return user info in UserInfo format
	userInfo := UserInfo{
		ID:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		DisplayName:    user.DisplayName,
		Avatar:         user.Avatar,
		Status:         user.Status.String(),
		EmailVerified:  user.EmailVerified,
		PhoneVerified:  user.PhoneVerified,
		EnableOTP:      user.EnableOTP,
		LastLoginIP:    user.LastLoginIP,
		OrganizationID: user.OrganizationID,
	}
	if user.LastLoginAt != nil {
		userInfo.LastLoginAt = *user.LastLoginAt
	}

	// 添加组织信息
	if user.Organization != nil {
		userInfo.Organization = &OrganizationSimpleInfo{
			ID:   user.Organization.ID,
			Name: user.Organization.Name,
			Code: user.Organization.Code,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data":    userInfo,
	})
}

// OTP handlers
func SendOTPHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

func VerifyOTPHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

// Password management handlers
func ForgotPasswordHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

func ResetPasswordHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

func ChangePasswordHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

// User profile handlers
func GetProfileHandler(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "User not authenticated",
			"data":    nil,
		})
		return
	}

	var user models.User
	if err := database.DB.Preload("Organization").Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "User not found",
				"data":    nil,
			})
			return
		}
		logger.ErrorError("Failed to get user profile", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// Build response data
	profileData := gin.H{
		"id":                user.ID,
		"username":          user.Username,
		"email":             user.Email,
		"display_name":      user.DisplayName,
		"phone":             user.Phone,
		"avatar":            user.Avatar,
		"status":            user.Status.String(),
		"email_verified":    user.EmailVerified,
		"phone_verified":    user.PhoneVerified,
		"enable_otp":        user.EnableOTP,
		"organization_id":   user.OrganizationID,
		"organization_name": "",
		"created_at":        user.CreatedAt,
		"updated_at":        user.UpdatedAt,
	}

	// Add organization name if available
	if user.Organization != nil {
		profileData["organization_name"] = user.Organization.Name
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data":    profileData,
	})
}

func UpdateProfileHandler(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "User not authenticated",
			"data":    nil,
		})
		return
	}

	var req struct {
		DisplayName string `json:"display_name"`
		Phone       string `json:"phone"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request data",
			"data":    nil,
		})
		return
	}

	// Check if user exists
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "User not found",
				"data":    nil,
			})
			return
		}
		logger.ErrorError("Failed to get user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// 记录更新前的值用于审计
	oldValues := gin.H{
		"display_name": user.DisplayName,
		"phone":        user.Phone,
	}

	// Update user information
	updates := make(map[string]interface{})
	changedFields := make([]string, 0)

	// 检查 DisplayName 是否有变化（允许设置为空字符串）
	if req.DisplayName != user.DisplayName {
		updates["display_name"] = req.DisplayName
		changedFields = append(changedFields, "display_name")
	}
	// 检查 Phone 是否有变化（允许设置为空字符串）
	if req.Phone != user.Phone {
		updates["phone"] = req.Phone
		changedFields = append(changedFields, "phone")
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		if err := database.DB.Model(&user).Updates(updates).Error; err != nil {
			// 记录失败的审计日志
			utils.CreateAuditLogWithError(c, utils.AuditActionUpdate, utils.AuditResourceUser, userID,
				"Failed to update user profile", err.Error(), gin.H{
					"old_values":     oldValues,
					"new_values":     req,
					"changed_fields": changedFields,
				})

			logger.ErrorError("Failed to update user profile", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "Failed to update profile",
				"data":    nil,
			})
			return
		}

		// 记录成功的审计日志
		newValues := gin.H{
			"display_name": req.DisplayName,
			"phone":        req.Phone,
		}

		utils.CreateAuditLog(c, utils.AuditActionUpdate, utils.AuditResourceUser, userID,
			"User profile updated successfully", gin.H{
				"old_values":     oldValues,
				"new_values":     newValues,
				"changed_fields": changedFields,
			})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Profile updated successfully",
		"data":    gin.H{"message": "Profile updated successfully"},
	})
}

// Avatar upload handler
func UploadAvatarHandler(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "User not authenticated",
			"data":    nil,
		})
		return
	}

	// Parse multipart form
	err := c.Request.ParseMultipartForm(10 << 20) // 10MB limit
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "File size too large (max 10MB)",
			"data":    nil,
		})
		return
	}

	// Get file from form
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "No file uploaded",
			"data":    nil,
		})
		return
	}
	defer file.Close()

	// Validate file type
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
	}

	contentType := header.Header.Get("Content-Type")
	if !allowedTypes[contentType] {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid file type. Only JPEG, PNG, and GIF are allowed",
			"data":    nil,
		})
		return
	}

	// Validate file size
	if header.Size > 10<<20 { // 10MB
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "File size too large (max 10MB)",
			"data":    nil,
		})
		return
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		// Determine extension from content type
		switch contentType {
		case "image/jpeg", "image/jpg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		default:
			ext = ".jpg"
		}
	}

	filename := fmt.Sprintf("%s_%d%s", userID, time.Now().Unix(), ext)
	avatarPath := filepath.Join("uploads", "avatars", filename)

	// Ensure upload directory exists
	uploadDir := filepath.Join("uploads", "avatars")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		logger.ErrorError("Failed to create upload directory", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// Save file
	dst, err := os.Create(avatarPath)
	if err != nil {
		logger.ErrorError("Failed to create avatar file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		logger.ErrorError("Failed to save avatar file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// Update user avatar in database
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		// Clean up uploaded file if user not found
		os.Remove(avatarPath)
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "User not found",
				"data":    nil,
			})
			return
		}
		logger.ErrorError("Failed to get user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// Remove old avatar file if exists
	if user.Avatar != "" {
		oldPath := user.Avatar
		if _, err := os.Stat(oldPath); err == nil {
			os.Remove(oldPath)
		}
	}

	// Update user avatar path
	if err := database.DB.Model(&user).Update("avatar", avatarPath).Error; err != nil {
		// Clean up uploaded file if database update fails
		os.Remove(avatarPath)
		logger.ErrorError("Failed to update user avatar", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// Create audit log
	utils.CreateAuditLog(c, utils.AuditActionUpdate, utils.AuditResourceUser, userID,
		"Avatar uploaded successfully", gin.H{
			"filename":  filename,
			"file_size": header.Size,
			"file_type": contentType,
		})

	logger.Info("Avatar uploaded successfully",
		zap.String("user_id", userID),
		zap.String("filename", filename),
		zap.Int64("file_size", header.Size),
	)

	// Generate avatar URL for response
	avatarURL := fmt.Sprintf("/uploads/avatars/%s", filename)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Avatar uploaded successfully",
		"data": gin.H{
			"avatar":     avatarURL,
			"avatar_url": avatarURL, // 保留兼容性
			"filename":   filename,
		},
	})
}

// Email verification handler
func VerifyEmailHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

// OTP setup handler
func SetupOTPHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

// Backup codes handler
func GetBackupCodesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

// System settings handlers - 实现在 system_setting.go 中

// OTP settings handlers
func EnableOTPHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

func DisableOTPHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

// User application handlers
func GetUserApplicationsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

func GetUserApplicationHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

// User management handlers - 实现在 user.go 中

// Organization management handlers - 实现在 organization.go 中

// Role management handlers
func GetRolesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

func CreateRoleHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

func UpdateRoleHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

func DeleteRoleHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

// Permission management handlers
func GetPermissionsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

func CreatePermissionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

func UpdatePermissionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

func DeletePermissionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

// Application management handlers
func GetApplicationsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

func CreateApplicationHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

func UpdateApplicationHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

func DeleteApplicationHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

// Application group management handlers
func GetApplicationGroupsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

func CreateApplicationGroupHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

func UpdateApplicationGroupHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

func DeleteApplicationGroupHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

// Log management handlers
func GetLoginLogsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

func GetAuditLogsHandler(c *gin.Context) {
	// 获取分页参数
	page := 1
	pageSize := 10 // 默认10条

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// 支持的页面大小选项：10, 20, 50, 100
	allowedPageSizes := []int{10, 20, 50, 100}
	found := false
	for _, size := range allowedPageSizes {
		if pageSize == size {
			found = true
			break
		}
	}
	if !found {
		pageSize = 10 // 如果不在允许的范围内，默认使用10
	}

	// 构建查询
	query := database.DB.Model(&models.AuditLog{}).Preload("User")

	// 添加过滤条件
	if action := c.Query("action"); action != "" {
		query = query.Where("action = ?", action)
	}

	if resource := c.Query("resource"); resource != "" {
		query = query.Where("resource = ?", resource)
	}

	if userID := c.Query("user_id"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.ErrorError("Failed to count audit logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// 分页查询
	offset := (page - 1) * pageSize
	var auditLogs []models.AuditLog
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&auditLogs).Error; err != nil {
		logger.ErrorError("Failed to get audit logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// 构建响应数据
	var logs []gin.H
	for _, log := range auditLogs {
		userName := ""
		if log.User.Username != "" {
			userName = log.User.Username
		}

		logs = append(logs, gin.H{
			"id":          log.ID,
			"user_id":     log.UserID,
			"user":        userName,
			"action":      log.Action,
			"resource":    log.Resource,
			"resource_id": log.ResourceID,
			"description": log.Description,
			"details":     log.Details,
			"ip_address":  log.IPAddress,
			"user_agent":  log.UserAgent,
			"status":      log.Status,
			"error_msg":   log.ErrorMsg,
			"created_at":  log.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"items":   logs,  // 前端期望的数据格式
		"total":   total, // 前端期望的总数格式
		"data": gin.H{
			"logs": logs,
			"pagination": gin.H{
				"page":       page,
				"page_size":  pageSize,
				"total":      total,
				"total_page": int(math.Ceil(float64(total) / float64(pageSize))),
			},
		},
	})
}
