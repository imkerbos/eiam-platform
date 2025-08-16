package utils

import (
	"encoding/json"

	"eiam-platform/internal/models"
	"eiam-platform/pkg/database"
	"eiam-platform/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuditAction 审计操作类型
const (
	AuditActionCreate = "create"
	AuditActionUpdate = "update"
	AuditActionDelete = "delete"
	AuditActionLogin  = "login"
	AuditActionLogout = "logout"
)

// AuditResource 审计资源类型
const (
	AuditResourceUser         = "user"
	AuditResourceOrganization = "organization"
	AuditResourceRole         = "role"
	AuditResourcePermission   = "permission"
	AuditResourceApplication  = "application"
	AuditResourceSystem       = "system"
)

// CreateAuditLog 创建审计日志
func CreateAuditLog(c *gin.Context, action, resource, resourceID, description string, details interface{}) {
	// 获取当前用户ID
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "system" // 系统操作
	}

	// 获取IP地址
	ipAddress := c.ClientIP()
	if ipAddress == "" {
		ipAddress = c.GetHeader("X-Forwarded-For")
	}

	// 获取用户代理
	userAgent := c.GetHeader("User-Agent")

	// 序列化详细信息
	var detailsJSON string
	if details != nil {
		if jsonBytes, err := json.Marshal(details); err == nil {
			detailsJSON = string(jsonBytes)
		}
	}

	// 创建审计日志记录
	auditLog := models.AuditLog{
		UserID:      userID,
		Action:      action,
		Resource:    resource,
		ResourceID:  resourceID,
		Description: description,
		Details:     detailsJSON,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Status:      "success",
	}

	// 异步保存审计日志
	go func() {
		if err := database.DB.Create(&auditLog).Error; err != nil {
			logger.Error("Failed to create audit log", zap.Error(err))
		}
	}()
}

// CreateAuditLogWithError 创建带错误信息的审计日志
func CreateAuditLogWithError(c *gin.Context, action, resource, resourceID, description, errorMsg string, details interface{}) {
	// 获取当前用户ID
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "system" // 系统操作
	}

	// 获取IP地址
	ipAddress := c.ClientIP()
	if ipAddress == "" {
		ipAddress = c.GetHeader("X-Forwarded-For")
	}

	// 获取用户代理
	userAgent := c.GetHeader("User-Agent")

	// 序列化详细信息
	var detailsJSON string
	if details != nil {
		if jsonBytes, err := json.Marshal(details); err == nil {
			detailsJSON = string(jsonBytes)
		}
	}

	// 创建审计日志记录
	auditLog := models.AuditLog{
		UserID:      userID,
		Action:      action,
		Resource:    resource,
		ResourceID:  resourceID,
		Description: description,
		Details:     detailsJSON,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Status:      "failed",
		ErrorMsg:    errorMsg,
	}

	// 异步保存审计日志
	go func() {
		if err := database.DB.Create(&auditLog).Error; err != nil {
			logger.Error("Failed to create audit log", zap.Error(err))
		}
	}()
}

// GetClientIP 获取客户端IP地址
func GetClientIP(c *gin.Context) string {
	// 尝试从各种头部获取真实IP
	headers := []string{
		"X-Forwarded-For",
		"X-Real-IP",
		"X-Client-IP",
		"CF-Connecting-IP", // Cloudflare
		"X-Forwarded",
		"Forwarded-For",
		"Forwarded",
	}

	for _, header := range headers {
		if ip := c.GetHeader(header); ip != "" {
			return ip
		}
	}

	// 如果没有找到，使用默认的ClientIP
	return c.ClientIP()
}
