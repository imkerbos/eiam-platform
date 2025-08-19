package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"eiam-platform/config"
	"eiam-platform/internal/models"
	"eiam-platform/pkg/database"
	"eiam-platform/pkg/i18n"
	"eiam-platform/pkg/logger"
	"eiam-platform/pkg/redis"
	"eiam-platform/pkg/session"
	"eiam-platform/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required" validate:"required,min=3,max=50"`
	Password string `json:"password" binding:"required" validate:"required,min=6"`
	OTPCode  string `json:"otp_code"` // 可选，OTP验证码
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int64    `json:"expires_in"`
	User         UserInfo `json:"user"`
	RequireOTP   bool     `json:"require_otp"`
	SessionID    string   `json:"session_id"` // 添加会话ID
}

// 全局会话管理器实例
var sessionManager *session.SessionManager

// InitSessionManager 初始化会话管理器
func InitSessionManager() {
	redisClient := redis.RDB
	sessionManager = session.NewSessionManager(redisClient, logger.GetLogger())
}

// GetSessionManager 获取会话管理器实例
func GetSessionManager() *session.SessionManager {
	return sessionManager
}

// UserInfo 用户信息
type UserInfo struct {
	ID             string                  `json:"id"`
	Username       string                  `json:"username"`
	Email          string                  `json:"email"`
	DisplayName    string                  `json:"display_name"`
	Avatar         string                  `json:"avatar"`
	Status         string                  `json:"status"`
	EmailVerified  bool                    `json:"email_verified"`
	PhoneVerified  bool                    `json:"phone_verified"`
	EnableOTP      bool                    `json:"enable_otp"`
	LastLoginAt    time.Time               `json:"last_login_at"`
	LastLoginIP    string                  `json:"last_login_ip"`
	OrganizationID string                  `json:"organization_id"`
	Organization   *OrganizationSimpleInfo `json:"organization,omitempty"`
	Roles          []string                `json:"roles,omitempty"`
	Permissions    []string                `json:"permissions,omitempty"`
}

// OrganizationSimpleInfo 组织简要信息
type OrganizationSimpleInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse 刷新令牌响应
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// LoginHandler 登录处理器
func LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.AccessInfo("Login request validation failed",
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
			logger.AccessInfo("Login failed: user not found",
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
		logger.ErrorError("Database error during login",
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
		logger.AccessInfo("Login failed: user inactive",
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

	// 验证密码（前端已经AES加密，需要先解密）
	decryptedPassword := req.Password

	// 如果密码长度表明是加密的，尝试解密
	if len(req.Password) > 32 {
		if decrypted, err := utils.DecryptPassword(req.Password); err == nil {
			decryptedPassword = decrypted
		}
		// 如果解密失败，保持原密码（可能是明文或MD5）
	}

	if !utils.CheckPassword(decryptedPassword, user.Password) {
		// 更新失败登录次数
		user.FailedCount++
		if user.FailedCount >= 5 {
			// 锁定账户30分钟
			lockTime := time.Now().Add(30 * time.Minute)
			user.LockedUntil = &lockTime
		}
		database.DB.Save(&user)

		logger.AccessInfo("Login failed: invalid password",
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

	// 检查是否需要升级密码哈希
	if utils.ShouldUpgradePassword(user.Password) {
		// 在后台升级密码哈希
		if newHash, err := utils.UpgradePasswordHash(decryptedPassword); err == nil {
			user.Password = newHash
			// 注意：这里会在更新登录信息时一起保存
		}
	}

	// 检查账户是否被锁定
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		logger.AccessInfo("Login failed: account locked",
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

	// 检查是否需要OTP验证
	if user.EnableOTP {
		if req.OTPCode == "" {
			// 需要OTP验证码
			logger.AccessInfo("Login requires OTP",
				zap.String("ip", c.ClientIP()),
				zap.String("username", user.Username),
			)
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": i18n.OTPRequired,
				"data": LoginResponse{
					RequireOTP: true,
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
					},
				},
			})
			return
		}

		// TODO: 验证OTP码
		// 这里暂时跳过OTP验证，实际应该验证TOTP
		if req.OTPCode != "123456" { // 临时使用固定码
			logger.AccessInfo("Login failed: invalid OTP",
				zap.String("ip", c.ClientIP()),
				zap.String("username", user.Username),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": i18n.InvalidOTP,
				"data":    nil,
			})
			return
		}
	}

	// 生成JWT令牌（包含完整用户信息）
	cfg := config.GetConfig()

	// 获取用户角色和权限 (这里暂时为空，后续可以扩展)
	roles := []string{}
	permissions := []string{}

	accessToken, err := utils.GenerateAccessTokenWithUserInfo(
		user.ID,
		user.Username,
		user.Email,
		user.DisplayName,
		roles,
		permissions,
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenExpire,
	)
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
		LoginType: "password",
		LoginIP:   c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		Success:   true,
	}
	database.DB.Create(&loginLog)

	// 创建Redis会话
	var sessionID string
	logger.AccessInfo("Session manager status",
		zap.Bool("session_manager_initialized", sessionManager != nil),
		zap.String("username", user.Username),
	)
	
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
	} else {
		logger.ErrorError("Session manager is nil",
			zap.String("username", user.Username),
		)
	}

	logger.AccessInfo("Login successful",
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
			},
		},
	})
}

// LogoutHandler 登出处理器
func LogoutHandler(c *gin.Context) {
	// 从请求头获取token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Authorization header required",
			"data":    nil,
		})
		return
	}

	tokenString := utils.ExtractTokenFromHeader(authHeader)
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Invalid token format",
			"data":    nil,
		})
		return
	}

	// 解析token获取用户信息
	cfg := config.GetConfig()
	jwtManager := utils.NewJWTManager(&cfg.JWT)
	claims, err := jwtManager.ValidateAccessToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Invalid token",
			"data":    nil,
		})
		return
	}

	// 将token加入黑名单
	if sessionManager != nil && claims.TradeID != "" {
		ctx := context.Background()
		// 计算token剩余有效期
		expireDuration := time.Until(claims.ExpiresAt.Time)
		if expireDuration > 0 {
			err := sessionManager.BlacklistToken(ctx, claims.TradeID, expireDuration)
			if err != nil {
				logger.ErrorError("Failed to blacklist token",
					zap.String("user_id", claims.UserID),
					zap.String("trade_id", claims.TradeID),
					zap.Error(err),
				)
			}
		}
	}

	// 删除特定会话（如果有session_id）
	if sessionManager != nil && claims.SessionID != "" {
		ctx := context.Background()
		err := sessionManager.DeleteSession(ctx, claims.SessionID)
		if err != nil {
			logger.ErrorError("Failed to delete session",
				zap.String("user_id", claims.UserID),
				zap.String("session_id", claims.SessionID),
				zap.Error(err),
			)
		}
	}

	logger.AccessInfo("User logged out",
		zap.String("user_id", claims.UserID),
		zap.String("username", claims.Username),
		zap.String("ip", c.ClientIP()),
	)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Logout successful",
		"data":    nil,
	})
}

// ForceLogoutUserHandler 强制用户下线处理器
func ForceLogoutUserHandler(c *gin.Context) {
	userID := c.Param("userID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "User ID is required",
			"data":    nil,
		})
		return
	}

	if sessionManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Session manager not initialized",
			"data":    nil,
		})
		return
	}

	ctx := context.Background()
	err := sessionManager.ForceLogoutUser(ctx, userID)
	if err != nil {
		logger.ErrorError("Failed to force logout user",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to force logout user",
			"data":    nil,
		})
		return
	}

	logger.AccessInfo("User forced logout by admin",
		zap.String("user_id", userID),
		zap.String("admin_ip", c.ClientIP()),
	)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "User forced logout successful",
		"data":    nil,
	})
}

// GetUserSessionsHandler 获取用户会话列表处理器
func GetUserSessionsHandler(c *gin.Context) {
	userID := c.Param("userID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "User ID is required",
			"data":    nil,
		})
		return
	}

	if sessionManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Session manager not initialized",
			"data":    nil,
		})
		return
	}

	ctx := context.Background()
	sessions, err := sessionManager.GetUserSessions(ctx, userID)
	if err != nil {
		logger.ErrorError("Failed to get user sessions",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to get user sessions",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data":    sessions,
	})
}

// GetAllSessionsHandler 获取所有在线会话处理器
func GetAllSessionsHandler(c *gin.Context) {
	if sessionManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Session manager not initialized",
			"data":    nil,
		})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	userID := c.Query("user_id")
	isActiveFilter := c.Query("is_active")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	ctx := context.Background()
	
	// 获取所有用户
	var users []models.User
	query := database.DB.Model(&models.User{})
	if userID != "" {
		query = query.Where("id = ?", userID)
	}
	if err := query.Find(&users).Error; err != nil {
		logger.ErrorError("Failed to get users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to get users",
			"data":    nil,
		})
		return
	}

	var allSessions []map[string]interface{}
	var total int64 = 0

	// 获取每个用户的会话
	for _, user := range users {
		sessions, err := sessionManager.GetUserSessions(ctx, user.ID)
		if err != nil {
			logger.ErrorError("Failed to get user sessions",
				zap.String("user_id", user.ID),
				zap.Error(err),
			)
			continue
		}

		for _, session := range sessions {
			// 检查会话是否过期
			isActive := time.Now().Before(session.ExpiresAt)
			
			// 过滤活跃状态
			if isActiveFilter != "" {
				isActiveBool, _ := strconv.ParseBool(isActiveFilter)
				if isActive != isActiveBool {
					continue
				}
			}

			sessionData := map[string]interface{}{
				"id":            session.SessionID,
				"user_id":       user.ID,
				"username":      user.Username,
				"session_id":    session.SessionID,
				"login_ip":      session.LoginIP,
				"user_agent":    session.UserAgent,
				"device_type":   "Unknown", // SessionInfo中没有这个字段
				"location":      "Unknown", // SessionInfo中没有这个字段
				"login_time":    session.LoginTime,
				"last_activity": session.LastActivity,
				"expires_at":    session.ExpiresAt,
				"is_active":     isActive,
			}
			allSessions = append(allSessions, sessionData)
			total++
		}
	}

	// 分页处理
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= len(allSessions) {
		allSessions = []map[string]interface{}{}
	} else if end > len(allSessions) {
		allSessions = allSessions[start:]
	} else {
		allSessions = allSessions[start:end]
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data": gin.H{
			"items":       allSessions,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
		},
	})
}



// RefreshTokenHandler 刷新令牌处理器
func RefreshTokenHandler(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": i18n.InvalidRequestData,
			"data":    nil,
		})
		return
	}

	// 验证刷新令牌
	cfg := config.GetConfig()
	claims, err := utils.ValidateRefreshToken(req.RefreshToken, cfg.JWT.Secret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": i18n.InvalidToken,
			"data":    nil,
		})
		return
	}

	// 获取用户信息
	var user models.User
	if err := database.DB.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": i18n.UserNotFound,
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

	// 生成新的令牌（包含完整用户信息）
	roles := []string{}
	permissions := []string{}

	accessToken, err := utils.GenerateAccessTokenWithUserInfo(
		user.ID,
		user.Username,
		user.Email,
		user.DisplayName,
		roles,
		permissions,
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenExpire,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Username, cfg.JWT.Secret, cfg.JWT.RefreshTokenExpire)
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
		"message": i18n.TokenRefreshed,
		"data": RefreshTokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    int64(cfg.JWT.AccessTokenExpire),
		},
	})
}
