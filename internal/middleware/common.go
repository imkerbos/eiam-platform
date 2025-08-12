package middleware

import (
	"strings"
	"time"

	"eiam-platform/config"
	"eiam-platform/pkg/i18n"
	"eiam-platform/pkg/logger"
	"eiam-platform/pkg/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// LoggerMiddleware 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 使用访问日志记录HTTP请求
		logger.AccessInfo(i18n.LogRequestReceived,
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.Int("status", param.StatusCode),
			zap.String("ip", param.ClientIP),
			zap.String("user_agent", param.Request.UserAgent()),
			zap.Duration("latency", param.Latency),
			zap.String("trade_id", param.Request.Header.Get("X-Trade-ID")),
		)
		return ""
	})
}

// RequestIDMiddleware 请求ID中间件
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// RecoveryMiddleware 恢复中间件
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		requestID := c.GetString("request_id")
		tradeID := c.GetString("trade_id")

		// 使用错误日志记录系统异常
		logger.ErrorError(i18n.LogSystemException,
			zap.String("request_id", requestID),
			zap.String("trade_id", tradeID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
			zap.Any("error", recovered),
		)

		c.JSON(500, gin.H{
			"code":       500,
			"message":    i18n.InternalServerError,
			"request_id": requestID,
			"trade_id":   tradeID,
		})
	})
}

// CORSMiddleware CORS中间件
func CORSMiddleware(cfg *config.CORSConfig) gin.HandlerFunc {
	corsConfig := cors.Config{
		AllowOrigins:     cfg.AllowOrigins,
		AllowMethods:     cfg.AllowMethods,
		AllowHeaders:     cfg.AllowHeaders,
		ExposeHeaders:    cfg.ExposeHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           time.Duration(cfg.MaxAge) * time.Hour,
	}

	return cors.New(corsConfig)
}

// SecurityHeadersMiddleware 安全头中间件
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}

// RateLimitMiddleware 限流中间件 (简单示例)
func RateLimitMiddleware() gin.HandlerFunc {
	// TODO: 实现基于Redis的分布式限流
	return func(c *gin.Context) {
		c.Next()
	}
}

// IPWhitelistMiddleware IP白名单中间件
func IPWhitelistMiddleware(allowedIPs []string) gin.HandlerFunc {
	allowedIPSet := make(map[string]bool)
	for _, ip := range allowedIPs {
		allowedIPSet[ip] = true
	}

	return func(c *gin.Context) {
		if len(allowedIPs) == 0 {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		if !allowedIPSet[clientIP] {
			logger.AccessWarn(i18n.LogIPNotWhitelisted,
				zap.String("client_ip", clientIP),
				zap.String("path", c.Request.URL.Path),
				zap.String("trade_id", c.GetString("trade_id")),
			)
			c.JSON(403, gin.H{
				"code":    403,
				"message": i18n.LogAccessDenied,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// TradeIDMiddleware TradeID中间件
func TradeIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tradeID := c.GetHeader("X-Trade-ID")
		if tradeID == "" {
			// 根据请求路径生成不同来源的TradeID
			source := "api"
			if strings.Contains(c.Request.URL.Path, "/console/") {
				source = "console"
			} else if strings.Contains(c.Request.URL.Path, "/portal/") {
				source = "portal"
			} else if strings.Contains(c.Request.URL.Path, "/saml/") {
				source = "saml"
			} else if strings.Contains(c.Request.URL.Path, "/oauth2/") {
				source = "oauth2"
			}

			tradeID = utils.GenerateTradeIDString(source)
		}
		c.Set("trade_id", tradeID)
		c.Header("X-Trade-ID", tradeID)
		c.Set("timestamp", time.Now().Unix())
		c.Next()
	}
}
