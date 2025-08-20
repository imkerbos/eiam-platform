package handlers

import (
	"context"
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

	// 获取用户角色和权限
	var roles []string
	var permissions []string

	// 预加载用户角色
	if err := database.DB.Preload("Roles").Where("id = ?", user.ID).First(&user).Error; err == nil {
		for _, role := range user.Roles {
			roles = append(roles, role.Code)
		}
	}

	// 获取配置
	cfg := config.GetConfig()

	// 创建Redis会话（在生成token之前）
	var sessionID string
	var err error
	if sessionManager != nil {
		ctx := context.Background()
		sessionID, err = sessionManager.CreateSession(
			ctx,
			user.ID,
			user.Username,
			user.Email,
			user.DisplayName,
			c.ClientIP(),
			c.GetHeader("User-Agent"),
			"", // tokenID将在后续设置
			time.Duration(cfg.JWT.AccessTokenExpire)*time.Second,
		)
		if err != nil {
			logger.ErrorError("Failed to create session",
				zap.String("ip", c.ClientIP()),
				zap.String("username", user.Username),
				zap.Error(err),
			)
			// 会话创建失败不应该影响登录流程，只记录错误
		} else {
			logger.AccessInfo("Session created successfully",
				zap.String("session_id", sessionID),
				zap.String("username", user.Username),
			)
		}
	}

	// 生成JWT令牌（包含session_id）
	jwtManager := utils.NewJWTManager(&cfg.JWT)

	// 生成trade_id
	tradeID := utils.GenerateTradeIDString("login")

	// 使用新的token生成方式，包含session_id
	tokenInfo := &utils.TokenInfo{
		UserID:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Roles:       roles,
		Permissions: permissions,
		SessionID:   sessionID, // 包含session_id
		TradeID:     tradeID,
	}

	accessToken, err := jwtManager.GenerateAccessToken(tokenInfo)
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

	refreshToken, err := jwtManager.GenerateRefreshToken(user.ID, sessionID, tradeID)
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

	// 会话已在前面创建，这里只需要记录日志
	logger.AccessInfo("Session manager status",
		zap.Bool("session_manager_initialized", sessionManager != nil),
		zap.String("username", user.Username),
	)

	logger.AccessInfo("Console login successful",
		zap.String("ip", c.ClientIP()),
		zap.String("username", user.Username),
		zap.String("user_id", user.ID),
		zap.String("session_id", sessionID),
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
			SessionID:    sessionID,
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
				Roles:         roles,
				Permissions:   permissions,
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
	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "User not authenticated",
			"data":    nil,
		})
		return
	}

	// 获取用户信息
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "User not found",
			"data":    nil,
		})
		return
	}

	// 从数据库获取应用组和应用数据
	var groups []models.ApplicationGroup
	if err := database.DB.Preload("Applications").Where("status = ?", models.StatusActive).Order("sort ASC").Find(&groups).Error; err != nil {
		logger.ErrorError("Failed to get application groups", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to get application groups",
			"data":    nil,
		})
		return
	}

	// 转换为前端需要的格式
	var result []map[string]interface{}
	for _, group := range groups {
		// 只包含有应用的应用组
		if len(group.Applications) > 0 {
			groupData := map[string]interface{}{
				"id":           group.ID,
				"name":         group.Name,
				"description":  group.Description,
				"color":        "#1890ff", // 默认颜色，可以从group.Icon或新增的Color字段获取
				"applications": []map[string]interface{}{},
			}

			// 转换应用数据
			for _, app := range group.Applications {
				if app.Status == models.StatusActive {
					appData := map[string]interface{}{
						"id":          app.ID,
						"name":        app.Name,
						"description": app.Description,
						"type":        app.AppType,
						"status":      "active",
						"url":         app.HomePageURL,
						"logo":        app.Logo,
						"group":       group.Name,
						"group_color": "#1890ff",
					}
					groupData["applications"] = append(
						groupData["applications"].([]map[string]interface{}),
						appData,
					)
				}
			}

			// 只添加有应用的应用组
			if len(groupData["applications"].([]map[string]interface{})) > 0 {
				result = append(result, groupData)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data":    result,
	})
}

func GetUserApplicationHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": i18n.APINotImplemented, "trade_id": c.GetString("trade_id")})
}

// User management handlers - 实现在 user.go 中

// Organization management handlers - 实现在 organization.go 中

// Role management handlers
// RoleInfo 角色信息
type RoleInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Type        string `json:"type"`
	IsSystem    bool   `json:"is_system"`
	Scope       string `json:"scope"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// RoleListResponse 角色列表响应
type RoleListResponse struct {
	Items      []RoleInfo `json:"items"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	TotalPages int        `json:"total_pages"`
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Scope       string `json:"scope"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Scope       string `json:"scope"`
	Status      string `json:"status"`
}

func GetRolesHandler(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")
	status := c.Query("status")
	typeFilter := c.Query("type")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := database.DB.Model(&models.Role{})

	// 搜索过滤
	if search != "" {
		query = query.Where("name LIKE ? OR code LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// 状态过滤
	if status != "" {
		statusInt, err := strconv.Atoi(status)
		if err == nil {
			query = query.Where("status = ?", statusInt)
		}
	}

	// 类型过滤
	if typeFilter != "" {
		query = query.Where("type = ?", typeFilter)
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 计算总页数
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	// 分页查询
	offset := (page - 1) * pageSize
	var roles []models.Role
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&roles).Error
	if err != nil {
		logger.ErrorError("Failed to get roles", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// 转换为响应格式
	items := make([]RoleInfo, len(roles))
	for i, role := range roles {
		items[i] = RoleInfo{
			ID:          role.ID,
			Name:        role.Name,
			Code:        role.Code,
			Description: role.Description,
			Type:        role.Type,
			IsSystem:    role.IsSystem,
			Scope:       role.Scope,
			Status:      role.Status.String(),
			CreatedAt:   role.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   role.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": i18n.Success,
		"data": RoleListResponse{
			Items:      items,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	})
}

func CreateRoleHandler(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": i18n.InvalidRequestData,
			"data":    nil,
		})
		return
	}

	// 检查角色代码是否已存在
	var existingRole models.Role
	if err := database.DB.Where("code = ?", req.Code).First(&existingRole).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Role code already exists",
			"data":    nil,
		})
		return
	}

	// 创建角色
	role := models.Role{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Type:        req.Type,
		Scope:       req.Scope,
		Status:      models.StatusActive,
	}

	if err := database.DB.Create(&role).Error; err != nil {
		logger.ErrorError("Failed to create role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Role created successfully",
		"data":    role,
	})
}

func UpdateRoleHandler(c *gin.Context) {
	roleID := c.Param("id")
	if roleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Role ID is required",
			"data":    nil,
		})
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": i18n.InvalidRequestData,
			"data":    nil,
		})
		return
	}

	// 查找角色
	var role models.Role
	if err := database.DB.Where("id = ?", roleID).First(&role).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "Role not found",
			"data":    nil,
		})
		return
	}

	// 更新角色
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Type != "" {
		updates["type"] = req.Type
	}
	if req.Scope != "" {
		updates["scope"] = req.Scope
	}
	if req.Status != "" {
		statusInt, err := strconv.Atoi(req.Status)
		if err == nil {
			updates["status"] = statusInt
		}
	}
	updates["updated_at"] = time.Now()

	if err := database.DB.Model(&role).Updates(updates).Error; err != nil {
		logger.ErrorError("Failed to update role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Role updated successfully",
		"data":    role,
	})
}

func DeleteRoleHandler(c *gin.Context) {
	roleID := c.Param("id")
	if roleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Role ID is required",
			"data":    nil,
		})
		return
	}

	// 查找角色
	var role models.Role
	if err := database.DB.Where("id = ?", roleID).First(&role).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "Role not found",
			"data":    nil,
		})
		return
	}

	// 检查是否为系统角色
	if role.IsSystem {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Cannot delete system role",
			"data":    nil,
		})
		return
	}

	// 检查是否有用户使用此角色
	var userCount int64
	database.DB.Table("user_roles").Where("role_id = ?", roleID).Count(&userCount)
	if userCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Cannot delete role that is assigned to users",
			"data":    nil,
		})
		return
	}

	// 软删除角色
	if err := database.DB.Delete(&role).Error; err != nil {
		logger.ErrorError("Failed to delete role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Role deleted successfully",
		"data":    nil,
	})
}

// AdministratorInfo 管理员信息
type AdministratorInfo struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	RoleCode    string `json:"role_code"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

// AdministratorListResponse 管理员列表响应
type AdministratorListResponse struct {
	Items      []AdministratorInfo `json:"items"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
}

// AssignRoleRequest 分配角色请求
type AssignRoleRequest struct {
	UserID string `json:"user_id" binding:"required"`
	RoleID string `json:"role_id" binding:"required"`
}

// GetAdministratorsHandler 获取管理员列表
func GetAdministratorsHandler(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 查询有管理员角色的用户
	query := `
		SELECT DISTINCT 
			u.id,
			u.username,
			u.display_name,
			u.email,
			u.status,
			u.created_at,
			r.name as role_name,
			r.code as role_code
		FROM users u
		JOIN user_roles ur ON u.id = ur.user_id
		JOIN roles r ON ur.role_id = r.id
		WHERE r.code LIKE '%ADMIN%' AND u.deleted_at IS NULL
		ORDER BY u.created_at DESC
	`

	rows, err := database.DB.Raw(query).Rows()
	if err != nil {
		logger.ErrorError("Failed to get administrators", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}
	defer rows.Close()

	var administrators []AdministratorInfo
	for rows.Next() {
		var admin AdministratorInfo
		var createdAt time.Time
		var statusInt int

		err := rows.Scan(
			&admin.ID,
			&admin.Username,
			&admin.DisplayName,
			&admin.Email,
			&statusInt,
			&createdAt,
			&admin.Role,
			&admin.RoleCode,
		)
		if err != nil {
			logger.ErrorError("Failed to scan administrator row", zap.Error(err))
			continue
		}

		admin.Status = models.Status(statusInt).String()
		admin.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
		administrators = append(administrators, admin)
	}

	// 手动分页
	total := int64(len(administrators))
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(administrators) {
		administrators = []AdministratorInfo{}
	} else if end > len(administrators) {
		administrators = administrators[start:]
	} else {
		administrators = administrators[start:end]
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": i18n.Success,
		"data": AdministratorListResponse{
			Items:      administrators,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	})
}

// AssignAdministratorRoleHandler 分配管理员角色
func AssignAdministratorRoleHandler(c *gin.Context) {
	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": i18n.InvalidRequestData,
			"data":    nil,
		})
		return
	}

	// 检查用户是否存在
	var user models.User
	if err := database.DB.Where("id = ?", req.UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "User not found",
			"data":    nil,
		})
		return
	}

	// 检查角色是否存在
	var role models.Role
	if err := database.DB.Where("id = ?", req.RoleID).First(&role).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "Role not found",
			"data":    nil,
		})
		return
	}

	// 检查是否已经分配了该角色
	var existingUserRole int64
	database.DB.Table("user_roles").Where("user_id = ? AND role_id = ?", req.UserID, req.RoleID).Count(&existingUserRole)
	if existingUserRole > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "User already has this role",
			"data":    nil,
		})
		return
	}

	// 分配角色
	userRole := struct {
		UserID string `gorm:"column:user_id"`
		RoleID string `gorm:"column:role_id"`
	}{
		UserID: req.UserID,
		RoleID: req.RoleID,
	}

	if err := database.DB.Table("user_roles").Create(&userRole).Error; err != nil {
		logger.ErrorError("Failed to assign role to user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Administrator role assigned successfully",
		"data":    nil,
	})
}

// RemoveAdministratorRoleHandler 移除管理员角色
func RemoveAdministratorRoleHandler(c *gin.Context) {
	userID := c.Param("userID")
	roleID := c.Param("roleID")

	if userID == "" || roleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "User ID and Role ID are required",
			"data":    nil,
		})
		return
	}

	// 检查是否为系统管理员角色
	var role models.Role
	if err := database.DB.Where("id = ?", roleID).First(&role).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "Role not found",
			"data":    nil,
		})
		return
	}

	// 不允许移除系统管理员角色
	if role.IsSystem && role.Code == "SYSTEM_ADMIN" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Cannot remove system administrator role",
			"data":    nil,
		})
		return
	}

	// 移除角色
	if err := database.DB.Table("user_roles").Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&struct{}{}).Error; err != nil {
		logger.ErrorError("Failed to remove role from user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Administrator role removed successfully",
		"data":    nil,
	})
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
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")
	groupID := c.Query("group_id")
	appType := c.Query("type")
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 构建查询
	query := database.DB.Model(&models.Application{}).Preload("Group")

	// 搜索条件
	if search != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if groupID != "" {
		query = query.Where("group_id = ?", groupID)
	}
	if appType != "" {
		query = query.Where("type = ?", appType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 计算总页数
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	// 分页查询
	offset := (page - 1) * pageSize
	var applications []models.Application
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&applications).Error
	if err != nil {
		logger.ErrorError("Failed to get applications", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to get applications",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data": gin.H{
			"items":       applications,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
		},
	})
}

func CreateApplicationHandler(c *gin.Context) {
	var req struct {
		Name        string                 `json:"name" binding:"required"`
		Type        string                 `json:"type" binding:"required"`
		Description string                 `json:"description"`
		GroupID     string                 `json:"groupId"`
		Status      int                    `json:"status"`
		HomepageURL string                 `json:"homepageUrl"`
		LogoURL     string                 `json:"logoUrl"`
		Config      map[string]interface{} `json:"config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request data",
			"data":    nil,
		})
		return
	}

	// 验证应用组是否存在
	if req.GroupID != "" {
		var group models.ApplicationGroup
		if err := database.DB.Where("id = ?", req.GroupID).First(&group).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "Application group not found",
				"data":    nil,
			})
			return
		}
	}

	// 生成唯一的ClientID和ClientSecret
	clientID := utils.GenerateTradeIDString("client")
	clientSecret := utils.GenerateTradeIDString("secret")
	appCode := utils.GenerateTradeIDString("app")
	
	application := models.Application{
		BaseModel:   models.BaseModel{ID: utils.GenerateTradeIDString("app")},
		Name:        req.Name,
		Code:        appCode, // 使用生成的唯一code
		Description: req.Description,
		GroupID:     &req.GroupID,
		ClientID:    clientID,
		ClientSecret: clientSecret,
		Status:      models.Status(req.Status),
		HomePageURL: req.HomepageURL,
		Logo:        req.LogoURL,
		Protocol:    req.Type,
		AppType:     "web", // 默认web类型
	}

	if err := database.DB.Create(&application).Error; err != nil {
		logger.ErrorError("Failed to create application", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to create application",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Application created successfully",
		"data":    application,
	})
}

func UpdateApplicationHandler(c *gin.Context) {
	appID := c.Param("id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Application ID is required",
			"data":    nil,
		})
		return
	}

	var req struct {
		Name        string                 `json:"name" binding:"required"`
		Type        string                 `json:"type" binding:"required"`
		Description string                 `json:"description"`
		GroupID     string                 `json:"groupId"`
		Status      int                    `json:"status"`
		HomepageURL string                 `json:"homepageUrl"`
		LogoURL     string                 `json:"logoUrl"`
		Config      map[string]interface{} `json:"config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request data",
			"data":    nil,
		})
		return
	}

	var application models.Application
	if err := database.DB.Where("id = ?", appID).First(&application).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "Application not found",
			"data":    nil,
		})
		return
	}

	// 验证应用组是否存在
	if req.GroupID != "" {
		var group models.ApplicationGroup
		if err := database.DB.Where("id = ?", req.GroupID).First(&group).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "Application group not found",
				"data":    nil,
			})
			return
		}
	}

	// 更新字段
	application.Name = req.Name
	application.Code = req.Name // 使用name作为code
	application.Description = req.Description
	application.GroupID = &req.GroupID
	application.Status = models.Status(req.Status)
	application.HomePageURL = req.HomepageURL
	application.Logo = req.LogoURL
	application.Protocol = req.Type

	if err := database.DB.Save(&application).Error; err != nil {
		logger.ErrorError("Failed to update application", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to update application",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Application updated successfully",
		"data":    application,
	})
}

func DeleteApplicationHandler(c *gin.Context) {
	appID := c.Param("id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Application ID is required",
			"data":    nil,
		})
		return
	}

	var application models.Application
	if err := database.DB.Where("id = ?", appID).First(&application).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "Application not found",
			"data":    nil,
		})
		return
	}

	if err := database.DB.Delete(&application).Error; err != nil {
		logger.ErrorError("Failed to delete application", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to delete application",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Application deleted successfully",
		"data":    nil,
	})
}

// Application group management handlers
func GetApplicationGroupsHandler(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 构建查询
	query := database.DB.Model(&models.ApplicationGroup{})

	// 搜索条件
	if search != "" {
		query = query.Where("name LIKE ? OR code LIKE ? OR description LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 计算总页数
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	// 分页查询
	offset := (page - 1) * pageSize
	var groups []models.ApplicationGroup
	err := query.Offset(offset).Limit(pageSize).Order("sort ASC, created_at DESC").Find(&groups).Error
	if err != nil {
		logger.ErrorError("Failed to get application groups", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to get application groups",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data": gin.H{
			"items":       groups,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
		},
	})
}

func CreateApplicationGroupHandler(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Code        string `json:"code" binding:"required"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
		Color       string `json:"color"`
		Sort        int    `json:"sort"`
		Status      int    `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request data",
			"data":    nil,
		})
		return
	}

	group := models.ApplicationGroup{
		BaseModel:   models.BaseModel{ID: utils.GenerateTradeIDString("appgroup")},
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Icon:        req.Icon,
		Color:       req.Color,
		Sort:        req.Sort,
		Status:      models.Status(req.Status),
	}

	if err := database.DB.Create(&group).Error; err != nil {
		logger.ErrorError("Failed to create application group", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to create application group",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Application group created successfully",
		"data":    group,
	})
}

func UpdateApplicationGroupHandler(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Group ID is required",
			"data":    nil,
		})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Code        string `json:"code" binding:"required"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
		Color       string `json:"color"`
		Sort        int    `json:"sort"`
		Status      int    `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request data",
			"data":    nil,
		})
		return
	}

	var group models.ApplicationGroup
	if err := database.DB.Where("id = ?", groupID).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "Application group not found",
			"data":    nil,
		})
		return
	}

	// 更新字段
	group.Name = req.Name
	group.Code = req.Code
	group.Description = req.Description
	group.Icon = req.Icon
	group.Color = req.Color
	group.Sort = req.Sort
	group.Status = models.Status(req.Status)

	if err := database.DB.Save(&group).Error; err != nil {
		logger.ErrorError("Failed to update application group", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to update application group",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Application group updated successfully",
		"data":    group,
	})
}

func DeleteApplicationGroupHandler(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Group ID is required",
			"data":    nil,
		})
		return
	}

	var group models.ApplicationGroup
	if err := database.DB.Where("id = ?", groupID).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "Application group not found",
			"data":    nil,
		})
		return
	}

	if err := database.DB.Delete(&group).Error; err != nil {
		logger.ErrorError("Failed to delete application group", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to delete application group",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Application group deleted successfully",
		"data":    nil,
	})
}

// Log management handlers
func GetLoginLogsHandler(c *gin.Context) {
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
	query := database.DB.Model(&models.UserLoginLog{}).Preload("User")

	// 添加过滤条件
	if loginType := c.Query("login_type"); loginType != "" {
		query = query.Where("login_type = ?", loginType)
	}

	if userID := c.Query("user_id"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if success := c.Query("success"); success != "" {
		if success == "true" {
			query = query.Where("success = ?", true)
		} else if success == "false" {
			query = query.Where("success = ?", false)
		}
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.ErrorError("Failed to count login logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// 分页查询
	offset := (page - 1) * pageSize
	var loginLogs []models.UserLoginLog
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&loginLogs).Error; err != nil {
		logger.ErrorError("Failed to get login logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// 构建响应数据
	var logs []gin.H
	for _, log := range loginLogs {
		userName := ""
		if log.User.Username != "" {
			userName = log.User.Username
		}

		logs = append(logs, gin.H{
			"id":          log.ID,
			"user_id":     log.UserID,
			"user":        userName,
			"login_type":  log.LoginType,
			"login_ip":    log.LoginIP,
			"user_agent":  log.UserAgent,
			"device_type": log.DeviceType,
			"location":    log.Location,
			"success":     log.Success,
			"fail_reason": log.FailReason,
			"duration":    log.Duration,
			"created_at":  log.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data": gin.H{
			"items":       logs,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": int(math.Ceil(float64(total) / float64(pageSize))),
		},
	})
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
