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

	// 手动查询用户角色
	var userRoles []models.Role
	if err := database.DB.Table("user_roles").
		Select("roles.*").
		Joins("JOIN roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", user.ID).
		Unscoped().
		Find(&userRoles).Error; err == nil {
		for _, role := range userRoles {
			roles = append(roles, role.Code)
		}
		logger.Info("Loaded user roles",
			zap.String("username", user.Username),
			zap.Strings("roles", roles),
		)
	} else {
		logger.ErrorError("Failed to load user roles",
			zap.String("username", user.Username),
			zap.Error(err),
		)
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
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.AccessInfo("Portal login request validation failed",
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
			logger.AccessInfo("Portal login failed: user not found",
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
		logger.ErrorError("Database error during portal login",
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
		logger.AccessInfo("Portal login failed: user inactive",
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

		logger.AccessInfo("Portal login failed: invalid password",
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
		logger.AccessInfo("Portal login failed: account locked",
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

	// 手动查询用户角色
	var userRoles []models.Role
	if err := database.DB.Table("user_roles").
		Select("roles.*").
		Joins("JOIN roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", user.ID).
		Unscoped().
		Find(&userRoles).Error; err == nil {
		for _, role := range userRoles {
			roles = append(roles, role.Code)
		}
		logger.Info("Loaded user roles for portal login",
			zap.String("username", user.Username),
			zap.Strings("roles", roles),
		)
	} else {
		logger.ErrorError("Failed to load user roles for portal login",
			zap.String("username", user.Username),
			zap.Error(err),
		)
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
			logger.ErrorError("Failed to create session for portal login",
				zap.String("ip", c.ClientIP()),
				zap.String("username", user.Username),
				zap.Error(err),
			)
			// 会话创建失败不应该影响登录流程，只记录错误
		} else {
			logger.AccessInfo("Session created successfully for portal login",
				zap.String("session_id", sessionID),
				zap.String("username", user.Username),
			)
		}
	}

	// 生成JWT令牌（包含session_id）
	jwtManager := utils.NewJWTManager(&cfg.JWT)

	// 生成trade_id
	tradeID := utils.GenerateTradeIDString("portal_login")

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
		logger.ErrorError("Failed to generate access token for portal login",
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
		logger.ErrorError("Failed to generate refresh token for portal login",
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
	user.FailedCount = 0   // 重置失败次数
	user.LockedUntil = nil // 清除锁定状态
	database.DB.Save(&user)

	// 记录登录日志
	logger.AccessInfo("Portal login successful",
		zap.String("ip", c.ClientIP()),
		zap.String("username", user.Username),
		zap.String("user_id", user.ID),
		zap.String("session_id", sessionID),
		zap.String("trade_id", tradeID),
	)

	// 返回登录成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Login successful",
		"data": gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"token_type":    "Bearer",
			"expires_in":    cfg.JWT.AccessTokenExpire,
			"user": gin.H{
				"id":           user.ID,
				"username":     user.Username,
				"email":        user.Email,
				"display_name": user.DisplayName,
				"roles":        roles,
				"permissions":  permissions,
			},
		},
		"trade_id": tradeID,
	})
}

func PortalLogoutHandler(c *gin.Context) {
	userID := c.GetString("user_id")
	sessionID := c.GetString("session_id")

	if sessionManager != nil && sessionID != "" {
		ctx := context.Background()
		err := sessionManager.DeleteSession(ctx, sessionID)
		if err != nil {
			logger.ErrorError("Failed to delete session during portal logout",
				zap.String("user_id", userID),
				zap.String("session_id", sessionID),
				zap.Error(err),
			)
		} else {
			logger.AccessInfo("Session deleted successfully during portal logout",
				zap.String("user_id", userID),
				zap.String("session_id", sessionID),
			)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     200,
		"message":  "Logout successful",
		"data":     nil,
		"trade_id": c.GetString("trade_id"),
	})
}

func PortalRefreshTokenHandler(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": i18n.InvalidRequestData,
			"data":    nil,
		})
		return
	}

	// 验证refresh token
	cfg := config.GetConfig()
	jwtManager := utils.NewJWTManager(&cfg.JWT)

	claims, err := jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
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
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "User not found",
			"data":    nil,
		})
		return
	}

	// 检查用户状态
	if user.Status != models.StatusActive {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": i18n.UserInactive,
			"data":    nil,
		})
		return
	}

	// 获取用户角色
	var roles []string
	var userRoles []models.Role
	if err := database.DB.Table("user_roles").
		Select("roles.*").
		Joins("JOIN roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", user.ID).
		Unscoped().
		Find(&userRoles).Error; err == nil {
		for _, role := range userRoles {
			roles = append(roles, role.Code)
		}
	}

	// 生成新的access token
	tokenInfo := &utils.TokenInfo{
		UserID:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Roles:       roles,
		Permissions: []string{}, // 暂时为空
		SessionID:   claims.SessionID,
		TradeID:     utils.GenerateTradeIDString("portal_refresh"),
	}

	newAccessToken, err := jwtManager.GenerateAccessToken(tokenInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Token refreshed successfully",
		"data": gin.H{
			"access_token": newAccessToken,
			"token_type":   "Bearer",
			"expires_in":   cfg.JWT.AccessTokenExpire,
		},
	})
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
// PermissionInfo 权限信息
type PermissionInfo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Code          string `json:"code"`
	Resource      string `json:"resource"`
	Action        string `json:"action"`
	Description   string `json:"description"`
	Category      string `json:"category"`
	ApplicationID string `json:"application_id,omitempty"`
	IsSystem      bool   `json:"is_system"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// PermissionListResponse 权限列表响应
type PermissionListResponse struct {
	Items      []PermissionInfo `json:"items"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

// CreatePermissionRequest 创建权限请求
type CreatePermissionRequest struct {
	Name          string `json:"name" binding:"required"`
	Code          string `json:"code" binding:"required"`
	Resource      string `json:"resource" binding:"required"`
	Action        string `json:"action" binding:"required"`
	Description   string `json:"description"`
	Category      string `json:"category" binding:"required"`
	ApplicationID string `json:"application_id"`
	IsSystem      bool   `json:"is_system"`
}

// UpdatePermissionRequest 更新权限请求
type UpdatePermissionRequest struct {
	Name          string `json:"name"`
	Resource      string `json:"resource"`
	Action        string `json:"action"`
	Description   string `json:"description"`
	Category      string `json:"category"`
	ApplicationID string `json:"application_id"`
	Status        string `json:"status"`
}

func GetPermissionsHandler(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")
	status := c.Query("status")
	category := c.Query("category")
	resource := c.Query("resource")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := database.DB.Model(&models.Permission{})

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

	// 分类过滤
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// 资源过滤
	if resource != "" {
		query = query.Where("resource = ?", resource)
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 计算总页数
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	// 分页查询
	offset := (page - 1) * pageSize
	var permissions []models.Permission
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&permissions).Error
	if err != nil {
		logger.ErrorError("Failed to get permissions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// 转换为响应格式
	items := make([]PermissionInfo, len(permissions))
	for i, permission := range permissions {
		applicationID := ""
		if permission.ApplicationID != nil {
			applicationID = *permission.ApplicationID
		}
		items[i] = PermissionInfo{
			ID:            permission.ID,
			Name:          permission.Name,
			Code:          permission.Code,
			Resource:      permission.Resource,
			Action:        permission.Action,
			Description:   permission.Description,
			Category:      permission.Category,
			ApplicationID: applicationID,
			IsSystem:      permission.IsSystem,
			Status:        permission.Status.String(),
			CreatedAt:     permission.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:     permission.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": i18n.Success,
		"data": PermissionListResponse{
			Items:      items,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	})
}

func CreatePermissionHandler(c *gin.Context) {
	var req CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": i18n.InvalidRequestData,
			"data":    nil,
		})
		return
	}

	// 检查权限代码是否已存在
	var existingPermission models.Permission
	if err := database.DB.Where("code = ?", req.Code).First(&existingPermission).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Permission code already exists",
			"data":    nil,
		})
		return
	}

	// 创建权限
	permission := models.Permission{
		Name:        req.Name,
		Code:        req.Code,
		Resource:    req.Resource,
		Action:      req.Action,
		Description: req.Description,
		Category:    req.Category,
		IsSystem:    req.IsSystem,
		Status:      models.StatusActive,
	}

	if req.ApplicationID != "" {
		permission.ApplicationID = &req.ApplicationID
	}

	if err := database.DB.Create(&permission).Error; err != nil {
		logger.ErrorError("Failed to create permission", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Permission created successfully",
		"data":    permission,
	})
}

func UpdatePermissionHandler(c *gin.Context) {
	permissionID := c.Param("id")
	if permissionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Permission ID is required",
			"data":    nil,
		})
		return
	}

	var req UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": i18n.InvalidRequestData,
			"data":    nil,
		})
		return
	}

	// 查找权限
	var permission models.Permission
	if err := database.DB.Where("id = ?", permissionID).First(&permission).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "Permission not found",
			"data":    nil,
		})
		return
	}

	// 更新权限
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Resource != "" {
		updates["resource"] = req.Resource
	}
	if req.Action != "" {
		updates["action"] = req.Action
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Category != "" {
		updates["category"] = req.Category
	}
	if req.ApplicationID != "" {
		updates["application_id"] = req.ApplicationID
	}
	if req.Status != "" {
		statusInt, err := strconv.Atoi(req.Status)
		if err == nil {
			updates["status"] = statusInt
		}
	}
	updates["updated_at"] = time.Now()

	if err := database.DB.Model(&permission).Updates(updates).Error; err != nil {
		logger.ErrorError("Failed to update permission", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Permission updated successfully",
		"data":    permission,
	})
}

func DeletePermissionHandler(c *gin.Context) {
	permissionID := c.Param("id")
	if permissionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Permission ID is required",
			"data":    nil,
		})
		return
	}

	// 查找权限
	var permission models.Permission
	if err := database.DB.Where("id = ?", permissionID).First(&permission).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "Permission not found",
			"data":    nil,
		})
		return
	}

	// 检查是否为系统权限
	if permission.IsSystem {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Cannot delete system permission",
			"data":    nil,
		})
		return
	}

	// 软删除权限
	if err := database.DB.Delete(&permission).Error; err != nil {
		logger.ErrorError("Failed to delete permission", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Permission deleted successfully",
		"data":    nil,
	})
}

// RoleAssignmentInfo 角色分配信息
type RoleAssignmentInfo struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	RoleID      string `json:"role_id"`
	RoleName    string `json:"role_name"`
	RoleCode    string `json:"role_code"`
	Status      string `json:"status"`
	AssignedAt  string `json:"assigned_at"`
}

// RoleAssignmentListResponse 角色分配列表响应
type RoleAssignmentListResponse struct {
	Items      []RoleAssignmentInfo `json:"items"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
}

// AssignRoleToUserRequest 分配角色给用户请求
type AssignRoleToUserRequest struct {
	UserID string `json:"user_id" binding:"required"`
	RoleID string `json:"role_id" binding:"required"`
}

func GetRoleAssignmentsHandler(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")
	roleID := c.Query("role_id")
	userID := c.Query("user_id")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 构建查询
	query := `
		SELECT 
			ur.user_id,
			u.username,
			u.display_name,
			u.email,
			u.status,
			ur.role_id,
			r.name as role_name,
			r.code as role_code,
			NOW() as assigned_at
		FROM user_roles ur
		JOIN users u ON ur.user_id = u.id
		JOIN roles r ON ur.role_id = r.id
		WHERE u.deleted_at IS NULL AND r.deleted_at IS NULL
	`

	args := []interface{}{}

	// 添加搜索条件
	if search != "" {
		query += " AND (u.username LIKE ? OR u.display_name LIKE ? OR r.name LIKE ?)"
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	if roleID != "" {
		query += " AND ur.role_id = ?"
		args = append(args, roleID)
	}

	if userID != "" {
		query += " AND ur.user_id = ?"
		args = append(args, userID)
	}

	query += " ORDER BY u.username ASC"

	// 执行查询
	rows, err := database.DB.Raw(query, args...).Rows()
	if err != nil {
		logger.ErrorError("Failed to get role assignments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}
	defer rows.Close()

	var assignments []RoleAssignmentInfo
	for rows.Next() {
		var assignment RoleAssignmentInfo
		var statusInt int
		var assignedAt time.Time

		err := rows.Scan(
			&assignment.UserID,
			&assignment.Username,
			&assignment.DisplayName,
			&assignment.Email,
			&statusInt,
			&assignment.RoleID,
			&assignment.RoleName,
			&assignment.RoleCode,
			&assignedAt,
		)
		if err != nil {
			logger.ErrorError("Failed to scan role assignment row", zap.Error(err))
			continue
		}

		assignment.ID = assignment.UserID + "_" + assignment.RoleID // 生成唯一ID
		assignment.Status = models.Status(statusInt).String()
		assignment.AssignedAt = assignedAt.Format("2006-01-02 15:04:05")
		assignments = append(assignments, assignment)
	}

	// 手动分页
	total := int64(len(assignments))
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(assignments) {
		assignments = []RoleAssignmentInfo{}
	} else if end > len(assignments) {
		assignments = assignments[start:]
	} else {
		assignments = assignments[start:end]
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": i18n.Success,
		"data": RoleAssignmentListResponse{
			Items:      assignments,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	})
}

func AssignRoleToUserHandler(c *gin.Context) {
	var req AssignRoleToUserRequest
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
		"message": "Role assigned successfully",
		"data":    nil,
	})
}

// Permission Route management handlers
// PermissionRouteInfo 权限路由信息
type PermissionRouteInfo struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Code              string   `json:"code"`
	Description       string   `json:"description"`
	Applications      []string `json:"applications"`
	ApplicationGroups []string `json:"application_groups"`
	Status            string   `json:"status"`
	CreatedAt         string   `json:"created_at"`
	UpdatedAt         string   `json:"updated_at"`
}

// PermissionRouteListResponse 权限路由列表响应
type PermissionRouteListResponse struct {
	Items      []PermissionRouteInfo `json:"items"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalPages int                   `json:"total_pages"`
}

// CreatePermissionRouteRequest 创建权限路由请求
type CreatePermissionRouteRequest struct {
	Name              string   `json:"name" binding:"required"`
	Code              string   `json:"code" binding:"required"`
	Description       string   `json:"description"`
	Applications      []string `json:"applications"`
	ApplicationGroups []string `json:"application_groups"`
	Status            string   `json:"status"`
}

// UpdatePermissionRouteRequest 更新权限路由请求
type UpdatePermissionRouteRequest struct {
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	Applications      []string `json:"applications"`
	ApplicationGroups []string `json:"application_groups"`
	Status            string   `json:"status"`
}

// PermissionRouteAssignmentInfo 权限路由分配信息
type PermissionRouteAssignmentInfo struct {
	ID                string `json:"id"`
	PermissionRouteID string `json:"permission_route_id"`
	PermissionName    string `json:"permission_name"`
	PermissionCode    string `json:"permission_code"`
	AssigneeType      string `json:"assignee_type"` // user, organization
	AssigneeID        string `json:"assignee_id"`
	AssigneeName      string `json:"assignee_name"`
	Status            string `json:"status"`
	AssignedAt        string `json:"assigned_at"`
}

// PermissionRouteAssignmentListResponse 权限路由分配列表响应
type PermissionRouteAssignmentListResponse struct {
	Items      []PermissionRouteAssignmentInfo `json:"items"`
	Total      int64                           `json:"total"`
	Page       int                             `json:"page"`
	PageSize   int                             `json:"page_size"`
	TotalPages int                             `json:"total_pages"`
}

// AssignPermissionRouteRequest 分配权限路由请求
type AssignPermissionRouteRequest struct {
	PermissionRouteID string `json:"permission_route_id" binding:"required"`
	AssigneeType      string `json:"assignee_type" binding:"required"` // user, organization
	AssigneeID        string `json:"assignee_id" binding:"required"`
	Status            string `json:"status"`
}

func GetPermissionRoutesHandler(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := database.DB.Model(&models.PermissionRoute{})

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

	// 获取总数
	var total int64
	query.Count(&total)

	// 计算总页数
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	// 分页查询
	offset := (page - 1) * pageSize
	var permissionRoutes []models.PermissionRoute
	err := query.Preload("Applications").Preload("ApplicationGroups").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&permissionRoutes).Error
	if err != nil {
		logger.ErrorError("Failed to get permission routes", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// 转换为响应格式
	items := make([]PermissionRouteInfo, len(permissionRoutes))
	for i, route := range permissionRoutes {
		applications := make([]string, len(route.Applications))
		for j, app := range route.Applications {
			applications[j] = app.Name
		}

		groups := make([]string, len(route.ApplicationGroups))
		for j, group := range route.ApplicationGroups {
			groups[j] = group.Name
		}

		items[i] = PermissionRouteInfo{
			ID:                route.ID,
			Name:              route.Name,
			Code:              route.Code,
			Description:       route.Description,
			Applications:      applications,
			ApplicationGroups: groups,
			Status:            route.Status.String(),
			CreatedAt:         route.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:         route.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data": PermissionRouteListResponse{
			Items:      items,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	})
}

func CreatePermissionRouteHandler(c *gin.Context) {
	var req CreatePermissionRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": i18n.InvalidRequestData,
			"data":    nil,
		})
		return
	}

	// 检查权限路由代码是否已存在
	var existingRoute models.PermissionRoute
	if err := database.DB.Where("code = ?", req.Code).First(&existingRoute).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Permission route code already exists",
			"data":    nil,
		})
		return
	}

	// 创建权限路由
	route := models.PermissionRoute{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Status:      models.StatusActive,
	}

	if req.Status != "" {
		if req.Status == "inactive" {
			route.Status = models.StatusInactive
		}
	}

	// 开始事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(&route).Error; err != nil {
		tx.Rollback()
		logger.ErrorError("Failed to create permission route", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// 关联应用
	if len(req.Applications) > 0 {
		var applications []models.Application
		if err := tx.Where("id IN ?", req.Applications).Find(&applications).Error; err == nil {
			tx.Model(&route).Association("Applications").Append(applications)
		}
	}

	// 关联应用组
	if len(req.ApplicationGroups) > 0 {
		var groups []models.ApplicationGroup
		if err := tx.Where("id IN ?", req.ApplicationGroups).Find(&groups).Error; err == nil {
			tx.Model(&route).Association("ApplicationGroups").Append(groups)
		}
	}

	if err := tx.Commit().Error; err != nil {
		logger.ErrorError("Failed to commit permission route creation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Permission route created successfully",
		"data":    route,
	})
}

func UpdatePermissionRouteHandler(c *gin.Context) {
	routeID := c.Param("id")
	if routeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Permission route ID is required",
			"data":    nil,
		})
		return
	}

	var req UpdatePermissionRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": i18n.InvalidRequestData,
			"data":    nil,
		})
		return
	}

	// 查找权限路由
	var route models.PermissionRoute
	if err := database.DB.Where("id = ?", routeID).First(&route).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "Permission route not found",
			"data":    nil,
		})
		return
	}

	// 开始事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新基本信息
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Status != "" {
		statusInt := models.StatusActive
		if req.Status == "inactive" {
			statusInt = models.StatusInactive
		}
		updates["status"] = statusInt
	}
	updates["updated_at"] = time.Now()

	if err := tx.Model(&route).Updates(updates).Error; err != nil {
		tx.Rollback()
		logger.ErrorError("Failed to update permission route", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// 更新应用关联
	if req.Applications != nil {
		tx.Model(&route).Association("Applications").Clear()
		if len(req.Applications) > 0 {
			var applications []models.Application
			if err := tx.Where("id IN ?", req.Applications).Find(&applications).Error; err == nil {
				tx.Model(&route).Association("Applications").Append(applications)
			}
		}
	}

	// 更新应用组关联
	if req.ApplicationGroups != nil {
		tx.Model(&route).Association("ApplicationGroups").Clear()
		if len(req.ApplicationGroups) > 0 {
			var groups []models.ApplicationGroup
			if err := tx.Where("id IN ?", req.ApplicationGroups).Find(&groups).Error; err == nil {
				tx.Model(&route).Association("ApplicationGroups").Append(groups)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		logger.ErrorError("Failed to commit permission route update", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Permission route updated successfully",
		"data":    route,
	})
}

func DeletePermissionRouteHandler(c *gin.Context) {
	routeID := c.Param("id")
	if routeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Permission route ID is required",
			"data":    nil,
		})
		return
	}

	// 查找权限路由
	var route models.PermissionRoute
	if err := database.DB.Where("id = ?", routeID).First(&route).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "Permission route not found",
			"data":    nil,
		})
		return
	}

	// 软删除权限路由
	if err := database.DB.Delete(&route).Error; err != nil {
		logger.ErrorError("Failed to delete permission route", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Permission route deleted successfully",
		"data":    nil,
	})
}

func GetPermissionRouteAssignmentsHandler(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 查询用户分配
	var userAssignments []struct {
		PermissionRouteID string `json:"permission_route_id"`
		PermissionName    string `json:"permission_name"`
		PermissionCode    string `json:"permission_code"`
		UserID            string `json:"user_id"`
		Username          string `json:"username"`
		DisplayName       string `json:"display_name"`
		AssignedAt        string `json:"assigned_at"`
	}

	userQuery := `
		SELECT 
			pru.permission_route_id,
			pr.name as permission_name,
			pr.code as permission_code,
			pru.user_id,
			u.username,
			u.display_name,
			pru.created_at as assigned_at
		FROM permission_route_users pru
		JOIN permission_routes pr ON pru.permission_route_id = pr.id
		JOIN users u ON pru.user_id = u.id
		WHERE pr.deleted_at IS NULL AND u.deleted_at IS NULL
		ORDER BY pru.created_at DESC
	`

	if err := database.DB.Raw(userQuery).Scan(&userAssignments).Error; err != nil {
		logger.ErrorError("Failed to get user permission route assignments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// 查询组织分配
	var orgAssignments []struct {
		PermissionRouteID string `json:"permission_route_id"`
		PermissionName    string `json:"permission_name"`
		PermissionCode    string `json:"permission_code"`
		OrganizationID    string `json:"organization_id"`
		OrganizationName  string `json:"organization_name"`
		AssignedAt        string `json:"assigned_at"`
	}

	orgQuery := `
		SELECT 
			pro.permission_route_id,
			pr.name as permission_name,
			pr.code as permission_code,
			pro.organization_id,
			o.name as organization_name,
			pro.created_at as assigned_at
		FROM permission_route_organizations pro
		JOIN permission_routes pr ON pro.permission_route_id = pr.id
		JOIN organizations o ON pro.organization_id = o.id
		WHERE pr.deleted_at IS NULL AND o.deleted_at IS NULL
		ORDER BY pro.created_at DESC
	`

	if err := database.DB.Raw(orgQuery).Scan(&orgAssignments).Error; err != nil {
		logger.ErrorError("Failed to get organization permission route assignments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// 合并结果
	allAssignments := make([]PermissionRouteAssignmentInfo, 0, len(userAssignments)+len(orgAssignments))

	// 添加用户分配
	for _, assignment := range userAssignments {
		allAssignments = append(allAssignments, PermissionRouteAssignmentInfo{
			ID:                assignment.UserID + "_" + assignment.PermissionRouteID,
			PermissionRouteID: assignment.PermissionRouteID,
			PermissionName:    assignment.PermissionName,
			PermissionCode:    assignment.PermissionCode,
			AssigneeType:      "user",
			AssigneeID:        assignment.UserID,
			AssigneeName:      assignment.DisplayName,
			Status:            "active",
			AssignedAt:        assignment.AssignedAt,
		})
	}

	// 添加组织分配
	for _, assignment := range orgAssignments {
		allAssignments = append(allAssignments, PermissionRouteAssignmentInfo{
			ID:                assignment.OrganizationID + "_" + assignment.PermissionRouteID,
			PermissionRouteID: assignment.PermissionRouteID,
			PermissionName:    assignment.PermissionName,
			PermissionCode:    assignment.PermissionCode,
			AssigneeType:      "organization",
			AssigneeID:        assignment.OrganizationID,
			AssigneeName:      assignment.OrganizationName,
			Status:            "active",
			AssignedAt:        assignment.AssignedAt,
		})
	}

	// 分页处理
	total := int64(len(allAssignments))
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	offset := (page - 1) * pageSize
	end := offset + pageSize
	if end > len(allAssignments) {
		end = len(allAssignments)
	}

	var items []PermissionRouteAssignmentInfo
	if offset < len(allAssignments) {
		items = allAssignments[offset:end]
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data": PermissionRouteAssignmentListResponse{
			Items:      items,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	})
}

func AssignPermissionRouteHandler(c *gin.Context) {
	var req AssignPermissionRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": i18n.InvalidRequestData,
			"data":    nil,
		})
		return
	}

	// 检查权限路由是否存在
	var route models.PermissionRoute
	if err := database.DB.Where("id = ?", req.PermissionRouteID).First(&route).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "Permission route not found",
			"data":    nil,
		})
		return
	}

	// 开始事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if req.AssigneeType == "user" {
		// 检查用户是否存在
		var user models.User
		if err := tx.Where("id = ?", req.AssigneeID).First(&user).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "User not found",
				"data":    nil,
			})
			return
		}

		// 检查是否已经分配
		var existingAssignment int64
		tx.Table("permission_route_users").Where("permission_route_id = ? AND user_id = ?", req.PermissionRouteID, req.AssigneeID).Count(&existingAssignment)
		if existingAssignment > 0 {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "User already has this permission route",
				"data":    nil,
			})
			return
		}

		// 分配权限路由给用户
		assignment := struct {
			PermissionRouteID string `gorm:"column:permission_route_id"`
			UserID            string `gorm:"column:user_id"`
		}{
			PermissionRouteID: req.PermissionRouteID,
			UserID:            req.AssigneeID,
		}

		if err := tx.Table("permission_route_users").Create(&assignment).Error; err != nil {
			tx.Rollback()
			logger.ErrorError("Failed to assign permission route to user", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": i18n.InternalServerError,
				"data":    nil,
			})
			return
		}

	} else if req.AssigneeType == "organization" {
		// 检查组织是否存在
		var org models.Organization
		if err := tx.Where("id = ?", req.AssigneeID).First(&org).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "Organization not found",
				"data":    nil,
			})
			return
		}

		// 检查是否已经分配
		var existingAssignment int64
		tx.Table("permission_route_organizations").Where("permission_route_id = ? AND organization_id = ?", req.PermissionRouteID, req.AssigneeID).Count(&existingAssignment)
		if existingAssignment > 0 {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "Organization already has this permission route",
				"data":    nil,
			})
			return
		}

		// 分配权限路由给组织
		assignment := struct {
			PermissionRouteID string `gorm:"column:permission_route_id"`
			OrganizationID    string `gorm:"column:organization_id"`
		}{
			PermissionRouteID: req.PermissionRouteID,
			OrganizationID:    req.AssigneeID,
		}

		if err := tx.Table("permission_route_organizations").Create(&assignment).Error; err != nil {
			tx.Rollback()
			logger.ErrorError("Failed to assign permission route to organization", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": i18n.InternalServerError,
				"data":    nil,
			})
			return
		}
	} else {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid assignee type",
			"data":    nil,
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		logger.ErrorError("Failed to commit permission route assignment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Permission route assigned successfully",
		"data":    nil,
	})
}

func RemovePermissionRouteAssignmentHandler(c *gin.Context) {
	assigneeType := c.Param("assigneeType")
	assigneeID := c.Param("assigneeId")
	permissionRouteID := c.Param("permissionRouteId")

	if assigneeType == "" || assigneeID == "" || permissionRouteID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Assignee type, assignee ID and permission route ID are required",
			"data":    nil,
		})
		return
	}

	if assigneeType == "user" {
		if err := database.DB.Table("permission_route_users").Where("permission_route_id = ? AND user_id = ?", permissionRouteID, assigneeID).Delete(&struct{}{}).Error; err != nil {
			logger.ErrorError("Failed to remove permission route from user", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": i18n.InternalServerError,
				"data":    nil,
			})
			return
		}
	} else if assigneeType == "organization" {
		if err := database.DB.Table("permission_route_organizations").Where("permission_route_id = ? AND organization_id = ?", permissionRouteID, assigneeID).Delete(&struct{}{}).Error; err != nil {
			logger.ErrorError("Failed to remove permission route from organization", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": i18n.InternalServerError,
				"data":    nil,
			})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid assignee type",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Permission route assignment removed successfully",
		"data":    nil,
	})
}

func RemoveRoleFromUserHandler(c *gin.Context) {
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

	// 检查是否为系统角色
	if role.IsSystem && role.Code == "SYSTEM_ADMIN" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Cannot remove system admin role",
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
		"message": "Role removed successfully",
		"data":    nil,
	})
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
		BaseModel:    models.BaseModel{ID: utils.GenerateTradeIDString("app")},
		Name:         req.Name,
		Code:         appCode, // 使用生成的唯一code
		Description:  req.Description,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Status:       models.Status(req.Status),
		HomePageURL:  req.HomepageURL,
		Logo:         req.LogoURL,
		Protocol:     req.Type,
		AppType:      "web", // 默认web类型
	}

	// 只有当提供了有效的GroupID时才设置
	if req.GroupID != "" {
		application.GroupID = &req.GroupID
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

		// OAuth2/OIDC配置字段
		ClientID        string `json:"clientId"`
		ClientSecret    string `json:"clientSecret"`
		RedirectURIs    string `json:"redirectUris"`
		Scopes          string `json:"scopes"`
		GrantTypes      string `json:"grantTypes"`
		ResponseTypes   string `json:"responseTypes"`
		AccessTokenTTL  int    `json:"accessTokenTTL"`
		RefreshTokenTTL int    `json:"refreshTokenTTL"`

		// SAML配置字段
		EntityID           string `json:"entity_id"`
		AcsURL             string `json:"acs_url"`
		SloURL             string `json:"slo_url"`
		Certificate        string `json:"certificate"`
		SignatureAlgorithm string `json:"signature_algorithm"`
		DigestAlgorithm    string `json:"digest_algorithm"`

		// CAS配置字段
		ServiceURL string `json:"serviceUrl"`
		Gateway    bool   `json:"gateway"`
		Renew      bool   `json:"renew"`

		// LDAP配置字段
		LdapURL      string `json:"ldapUrl"`
		BaseDN       string `json:"baseDn"`
		BindDN       string `json:"bindDn"`
		BindPassword string `json:"bindPassword"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.ErrorError("Failed to bind JSON for application update", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request data",
			"data":    gin.H{"error": err.Error()},
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

	// 准备更新数据
	updateData := map[string]interface{}{
		"name":          req.Name,
		"code":          req.Name, // 使用name作为code
		"description":   req.Description,
		"status":        models.Status(req.Status),
		"home_page_url": req.HomepageURL,
		"logo":          req.LogoURL,
		"protocol":      req.Type,
	}

	// 只有当提供了新的GroupID时才更新
	if req.GroupID != "" {
		updateData["group_id"] = req.GroupID
	}

	// 更新OAuth2/OIDC配置字段
	if req.Type == "oauth2" || req.Type == "oidc" {
		if req.ClientID != "" {
			updateData["client_id"] = req.ClientID
		}
		if req.ClientSecret != "" {
			updateData["client_secret"] = req.ClientSecret
		}
		if req.RedirectURIs != "" {
			updateData["redirect_uris"] = req.RedirectURIs
		}
		if len(req.Scopes) > 0 {
			updateData["scopes"] = req.Scopes
		}
		if req.GrantTypes != "" {
			updateData["grant_types"] = req.GrantTypes
		}
		if req.ResponseTypes != "" {
			updateData["response_types"] = req.ResponseTypes
		}
		if req.AccessTokenTTL > 0 {
			updateData["access_token_ttl"] = req.AccessTokenTTL
		}
		if req.RefreshTokenTTL > 0 {
			updateData["refresh_token_ttl"] = req.RefreshTokenTTL
		}
	}

	// 更新SAML配置字段
	if req.Type == "saml" {
		if req.EntityID != "" {
			updateData["entity_id"] = req.EntityID
		}
		if req.AcsURL != "" {
			updateData["acs_url"] = req.AcsURL
		}
		if req.SloURL != "" {
			updateData["slo_url"] = req.SloURL
		}
		if req.Certificate != "" {
			updateData["certificate"] = req.Certificate
		}
		if req.SignatureAlgorithm != "" {
			updateData["signature_algorithm"] = req.SignatureAlgorithm
		}
		if req.DigestAlgorithm != "" {
			updateData["digest_algorithm"] = req.DigestAlgorithm
		}
	}

	// 更新CAS配置字段
	if req.Type == "cas" {
		if req.ServiceURL != "" {
			updateData["service_url"] = req.ServiceURL
		}
		updateData["gateway"] = req.Gateway
		updateData["renew"] = req.Renew
	}

	// 更新LDAP配置字段
	if req.Type == "ldap" {
		if req.LdapURL != "" {
			updateData["ldap_url"] = req.LdapURL
		}
		if req.BaseDN != "" {
			updateData["base_dn"] = req.BaseDN
		}
		if req.BindDN != "" {
			updateData["bind_dn"] = req.BindDN
		}
		if req.BindPassword != "" {
			updateData["bind_password"] = req.BindPassword
		}
	}

	if err := database.DB.Model(&application).Updates(updateData).Error; err != nil {
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

	// 检查是否有应用属于该分组
	var appCount int64
	database.DB.Model(&models.Application{}).Where("group_id = ?", groupID).Count(&appCount)
	if appCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Cannot delete application group that contains applications. Please move or delete the applications first.",
			"data": gin.H{
				"application_count": appCount,
			},
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
		"data": gin.H{
			"items":       logs,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": int(math.Ceil(float64(total) / float64(pageSize))),
		},
	})
}
