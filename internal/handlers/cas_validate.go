package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"eiam-platform/internal/models"
	"eiam-platform/pkg/database"
	"eiam-platform/pkg/logger"
)

// CASValidateHandler CAS 1.0票据验证处理器
func CASValidateHandler(c *gin.Context) {
	service := c.Query("service")
	ticket := c.Query("ticket")

	logger.Info("CAS validate request",
		zap.String("service", service),
		zap.String("ticket", ticket),
		zap.String("ip", c.ClientIP()),
	)

	// 验证参数
	if service == "" || ticket == "" {
		c.String(http.StatusBadRequest, "no\n\n")
		return
	}

	// 查找并验证票据
	var serviceTicket CASServiceTicket
	if err := database.DB.Where("ticket = ? AND service = ? AND used = ? AND expires_at > ?", ticket, service, false, time.Now()).First(&serviceTicket).Error; err != nil {
		logger.ErrorError("Invalid or expired ticket", zap.Error(err))
		c.String(http.StatusOK, "no\n\n")
		return
	}

	// 标记票据为已使用
	serviceTicket.Used = true
	database.DB.Save(&serviceTicket)

	// 获取用户信息
	var user models.User
	if err := database.DB.Where("id = ?", serviceTicket.UserID).First(&user).Error; err != nil {
		logger.ErrorError("User not found", zap.Error(err))
		c.String(http.StatusOK, "no\n\n")
		return
	}

	// CAS 1.0 返回纯文本格式: yes\nusername\n
	c.String(http.StatusOK, fmt.Sprintf("yes\n%s\n", user.Username))

	logger.Info("CAS ticket validated successfully",
		zap.String("username", user.Username),
		zap.String("service", service),
		zap.String("ticket", ticket),
	)
}
