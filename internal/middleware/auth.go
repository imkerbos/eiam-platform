package middleware

import (
	"context"
	"net/http"

	"eiam-platform/pkg/i18n"
	"eiam-platform/pkg/logger"
	"eiam-platform/pkg/session"
	"eiam-platform/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthMiddleware JWT authentication middleware
func AuthMiddleware(jwtManager *utils.JWTManager, sessionManager *session.SessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Missing Authorization header",
			})
			c.Abort()
			return
		}

		token := utils.ExtractTokenFromHeader(authHeader)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Invalid Authorization format",
			})
			c.Abort()
			return
		}

		claims, err := jwtManager.ValidateAccessToken(token)
		if err != nil {
			logger.Warn(i18n.LogJWTValidationFailed,
				zap.String("error", err.Error()),
				zap.String("token", token[:min(len(token), 20)]+"..."),
				zap.String("trade_id", c.GetString("trade_id")),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":     401,
				"message":  i18n.InvalidToken,
				"trade_id": c.GetString("trade_id"),
			})
			c.Abort()
			return
		}

		// 检查token是否在黑名单中
		if sessionManager != nil && claims.TradeID != "" {
			ctx := context.Background()
			if sessionManager.IsTokenBlacklisted(ctx, claims.TradeID) {
				logger.Warn("Token is blacklisted",
					zap.String("user_id", claims.UserID),
					zap.String("username", claims.Username),
					zap.String("trade_id", claims.TradeID),
				)
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    401,
					"message": "Token has been revoked",
				})
				c.Abort()
				return
			}
		}

		// 验证会话是否有效（如果启用了会话管理且有session_id）
		if sessionManager != nil && claims.SessionID != "" {
			ctx := context.Background()
			if !sessionManager.IsSessionValid(ctx, claims.SessionID) {
				logger.Warn("Session is invalid",
					zap.String("user_id", claims.UserID),
					zap.String("username", claims.Username),
					zap.String("session_id", claims.SessionID),
				)
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    401,
					"message": "Session expired",
				})
				c.Abort()
				return
			}
		}

		// Store user information in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("display_name", claims.DisplayName)
		c.Set("roles", claims.Roles)
		c.Set("permissions", claims.Permissions)
		c.Set("claims", claims)

		c.Next()
	}
}

// OptionalAuthMiddleware optional JWT authentication middleware
func OptionalAuthMiddleware(jwtManager *utils.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			token := utils.ExtractTokenFromHeader(authHeader)
			if token != "" {
				claims, err := jwtManager.ValidateAccessToken(token)
				if err == nil {
					c.Set("user_id", claims.UserID)
					c.Set("username", claims.Username)
					c.Set("email", claims.Email)
					c.Set("display_name", claims.DisplayName)
					c.Set("roles", claims.Roles)
					c.Set("permissions", claims.Permissions)
					c.Set("claims", claims)
				}
			}
		}
		c.Next()
	}
}

// RoleMiddleware role permission middleware
func RoleMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("roles")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": i18n.Forbidden,
			})
			c.Abort()
			return
		}

		userRoles, ok := roles.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "Role information format error",
			})
			c.Abort()
			return
		}

		// Check if user has required role
		hasRequiredRole := false
		for _, requiredRole := range requiredRoles {
			for _, userRole := range userRoles {
				if userRole == requiredRole || userRole == "admin" || userRole == "SYSTEM_ADMIN" { // admin roles have all permissions
					hasRequiredRole = true
					break
				}
			}
			if hasRequiredRole {
				break
			}
		}

		if !hasRequiredRole {
			logger.Warn("Insufficient role permissions",
				zap.String("user_id", c.GetString("user_id")),
				zap.Strings("user_roles", userRoles),
				zap.Strings("required_roles", requiredRoles),
			)
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "Insufficient role permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminMiddleware admin permission middleware
func AdminMiddleware() gin.HandlerFunc {
	return RoleMiddleware("admin")
}

// GetCurrentUserID get current user ID
func GetCurrentUserID(c *gin.Context) string {
	userID, _ := c.Get("user_id")
	if id, ok := userID.(string); ok {
		return id
	}
	return ""
}

// GetCurrentUsername get current username
func GetCurrentUsername(c *gin.Context) string {
	username, _ := c.Get("username")
	if name, ok := username.(string); ok {
		return name
	}
	return ""
}

// GetCurrentUserRoles get current user roles
func GetCurrentUserRoles(c *gin.Context) []string {
	roles, _ := c.Get("roles")
	if r, ok := roles.([]string); ok {
		return r
	}
	return []string{}
}

// IsAdmin check if user is admin
func IsAdmin(c *gin.Context) bool {
	roles := GetCurrentUserRoles(c)
	for _, role := range roles {
		if role == "admin" {
			return true
		}
	}
	return false
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
