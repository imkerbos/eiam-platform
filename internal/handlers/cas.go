package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"eiam-platform/internal/models"
	"eiam-platform/pkg/database"
	"eiam-platform/pkg/logger"
	"eiam-platform/pkg/utils"
)

// CASServiceTicket CAS服务票据
type CASServiceTicket struct {
	models.BaseModel
	Ticket     string    `json:"ticket" gorm:"type:varchar(255);uniqueIndex;not null"`
	Service    string    `json:"service" gorm:"type:varchar(500);not null"`
	UserID     string    `json:"user_id" gorm:"type:varchar(36);not null;index"`
	Username   string    `json:"username" gorm:"type:varchar(100);not null"`
	ExpiresAt  time.Time `json:"expires_at" gorm:"not null"`
	Used       bool      `json:"used" gorm:"default:false"`
	Attributes string    `json:"attributes" gorm:"type:text"` // JSON格式的用户属性
}

// TableName 指定表名
func (CASServiceTicket) TableName() string {
	return "cas_service_tickets"
}

// CASProxyTicket CAS代理票据
type CASProxyTicket struct {
	models.BaseModel
	Ticket        string    `json:"ticket" gorm:"type:varchar(255);uniqueIndex;not null"`
	Service       string    `json:"service" gorm:"type:varchar(500);not null"`
	UserID        string    `json:"user_id" gorm:"type:varchar(36);not null;index"`
	Username      string    `json:"username" gorm:"type:varchar(100);not null"`
	ProxyGrantingTicket string `json:"proxy_granting_ticket" gorm:"type:varchar(255)"`
	ExpiresAt     time.Time `json:"expires_at" gorm:"not null"`
	Used          bool      `json:"used" gorm:"default:false"`
}

// TableName 指定表名
func (CASProxyTicket) TableName() string {
	return "cas_proxy_tickets"
}

// CASProxyGrantingTicket CAS代理授权票据
type CASProxyGrantingTicket struct {
	models.BaseModel
	Ticket    string    `json:"ticket" gorm:"type:varchar(255);uniqueIndex;not null"`
	UserID    string    `json:"user_id" gorm:"type:varchar(36);not null;index"`
	Username  string    `json:"username" gorm:"type:varchar(100);not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	Used      bool      `json:"used" gorm:"default:false"`
}

// TableName 指定表名
func (CASProxyGrantingTicket) TableName() string {
	return "cas_proxy_granting_tickets"
}

// CASLoginHandler CAS登录处理器
func CASLoginHandler(c *gin.Context) {
	service := c.Query("service")
	gateway := c.Query("gateway") == "true"
	renew := c.Query("renew") == "true"

	logger.Info("CAS login request",
		zap.String("service", service),
		zap.Bool("gateway", gateway),
		zap.Bool("renew", renew),
		zap.String("ip", c.ClientIP()),
	)

	// 验证service参数
	if service == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "service parameter is required",
		})
		return
	}

	// 验证service URL是否在允许的应用列表中
	var application models.Application
	if err := database.DB.Where("service_url = ? AND protocol = ? AND status = ?", service, "cas", 1).First(&application).Error; err != nil {
		logger.ErrorError("CAS service not found or not enabled", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid service URL",
		})
		return
	}

	// 检查用户是否已登录
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		// 从Cookie中获取session
		cookie, err := c.Cookie("session_id")
		if err == nil {
			sessionID = cookie
		}
	}

	var user *models.User
	var isLoggedIn bool

	if sessionID != "" {
		// 验证session
		sessionManager := GetSessionManager()
		if sessionManager != nil {
			ctx := c.Request.Context()
			sessionInfo, err := sessionManager.GetSession(ctx, sessionID)
			if err == nil && sessionInfo != nil {
				// 获取用户信息
				if err := database.DB.Where("id = ?", sessionInfo.UserID).First(&user).Error; err == nil {
					isLoggedIn = true
					logger.Info("User already logged in via session",
						zap.String("username", user.Username),
						zap.String("session_id", sessionID),
					)
				}
			}
		}
	}

	// 如果用户已登录且不是renew模式，直接生成服务票据
	if isLoggedIn && !renew {
		ticket, err := generateServiceTicket(user, service, &application)
		if err != nil {
			logger.ErrorError("Failed to generate service ticket", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "Failed to generate service ticket",
			})
			return
		}

		// 重定向到service URL with ticket
		redirectURL := buildRedirectURL(service, ticket, nil)
		logger.Info("Redirecting to service with ticket",
			zap.String("username", user.Username),
			zap.String("service", service),
			zap.String("ticket", ticket),
		)
		c.Redirect(http.StatusFound, redirectURL)
		return
	}

	// 如果gateway模式且用户未登录，直接重定向到service
	if gateway && !isLoggedIn {
		logger.Info("Gateway mode: redirecting to service without authentication",
			zap.String("service", service),
		)
		c.Redirect(http.StatusFound, service)
		return
	}

	// 需要用户登录
	// 这里可以重定向到登录页面，或者返回登录表单
	// 为了简化，我们返回一个简单的登录表单
	c.HTML(http.StatusOK, "cas_login.html", gin.H{
		"service": service,
		"gateway": gateway,
		"renew":   renew,
		"title":   "CAS Login",
	})
}

// CASLoginSubmitHandler CAS登录提交处理器
func CASLoginSubmitHandler(c *gin.Context) {
	var req struct {
		Username string `form:"username" binding:"required"`
		Password string `form:"password" binding:"required"`
		Service  string `form:"service" binding:"required"`
		Gateway  bool   `form:"gateway"`
		Renew    bool   `form:"renew"`
	}

	if err := c.ShouldBind(&req); err != nil {
		logger.ErrorError("CAS login parameter binding failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request parameters",
			"error":   err.Error(),
		})
		return
	}

	logger.Info("CAS login submit",
		zap.String("username", req.Username),
		zap.String("service", req.Service),
		zap.String("ip", c.ClientIP()),
	)

	// 验证service
	var application models.Application
	if err := database.DB.Where("service_url = ? AND protocol = ? AND status = ?", req.Service, "cas", 1).First(&application).Error; err != nil {
		logger.ErrorError("CAS service not found", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid service URL",
		})
		return
	}

	// 验证用户凭据
	var user models.User
	if err := database.DB.Where("(username = ? OR email = ?) AND status = ?", req.Username, req.Username, 1).First(&user).Error; err != nil {
		logger.ErrorError("User not found", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Invalid username or password",
		})
		return
	}

	// 验证密码
	if !utils.CheckPassword(req.Password, user.Password) {
		logger.ErrorError("Invalid password", zap.String("username", req.Username))
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Invalid username or password",
		})
		return
	}

	// 创建session
	sessionManager := GetSessionManager()
	if sessionManager != nil {
		ctx := c.Request.Context()
		sessionID, err := sessionManager.CreateSession(
			ctx,
			user.ID,
			user.Username,
			user.Email,
			user.DisplayName,
			c.ClientIP(),
			c.GetHeader("User-Agent"),
			"", // tokenID
			3600*time.Second, // 1小时过期
		)
		if err != nil {
			logger.ErrorError("Failed to create session", zap.Error(err))
		} else {
			// 设置session cookie
			c.SetCookie("session_id", sessionID, 3600*24*7, "/", "", false, true) // 7天
		}
	}

	// 生成服务票据
	ticket, err := generateServiceTicket(&user, req.Service, &application)
	if err != nil {
		logger.ErrorError("Failed to generate service ticket", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to generate service ticket",
		})
		return
	}

	// 重定向到service URL with ticket
	redirectURL := buildRedirectURL(req.Service, ticket, nil)
	logger.Info("CAS login successful, redirecting to service",
		zap.String("username", user.Username),
		zap.String("service", req.Service),
		zap.String("ticket", ticket),
	)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Login successful",
		"data": gin.H{
			"redirect_url": redirectURL,
			"ticket":       ticket,
		},
	})
}

// CASValidateHandler CAS票据验证处理器
func CASValidateHandler(c *gin.Context) {
	service := c.Query("service")
	ticket := c.Query("ticket")
	format := c.Query("format") // xml, json

	logger.Info("CAS validate request",
		zap.String("service", service),
		zap.String("ticket", ticket),
		zap.String("format", format),
		zap.String("ip", c.ClientIP()),
	)

	// 验证参数
	if service == "" || ticket == "" {
		if format == "xml" {
			c.Header("Content-Type", "application/xml")
			c.String(http.StatusOK, `<?xml version="1.0"?>
<cas:serviceResponse xmlns:cas='http://www.yale.edu/tp/cas'>
    <cas:authenticationFailure code='INVALID_REQUEST'>service and ticket parameters are required</cas:authenticationFailure>
</cas:serviceResponse>`)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "service and ticket parameters are required",
			})
		}
		return
	}

	// 查找并验证票据
	var serviceTicket CASServiceTicket
	if err := database.DB.Where("ticket = ? AND service = ? AND used = ? AND expires_at > ?", ticket, service, false, time.Now()).First(&serviceTicket).Error; err != nil {
		logger.ErrorError("Invalid or expired ticket", zap.Error(err))
		if format == "xml" {
			c.Header("Content-Type", "application/xml")
			c.String(http.StatusOK, `<?xml version="1.0"?>
<cas:serviceResponse xmlns:cas='http://www.yale.edu/tp/cas'>
    <cas:authenticationFailure code='INVALID_TICKET'>Ticket not recognized</cas:authenticationFailure>
</cas:serviceResponse>`)
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "Ticket not recognized",
			})
		}
		return
	}

	// 标记票据为已使用
	serviceTicket.Used = true
	database.DB.Save(&serviceTicket)

	// 获取用户信息
	var user models.User
	if err := database.DB.Where("id = ?", serviceTicket.UserID).First(&user).Error; err != nil {
		logger.ErrorError("User not found", zap.Error(err))
		if format == "xml" {
			c.Header("Content-Type", "application/xml")
			c.String(http.StatusOK, `<?xml version="1.0"?>
<cas:serviceResponse xmlns:cas='http://www.yale.edu/tp/cas'>
    <cas:authenticationFailure code='INTERNAL_ERROR'>User not found</cas:authenticationFailure>
</cas:serviceResponse>`)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "User not found",
			})
		}
		return
	}

	// 返回成功响应
	if format == "xml" {
		c.Header("Content-Type", "application/xml")
		c.String(http.StatusOK, `<?xml version="1.0"?>
<cas:serviceResponse xmlns:cas='http://www.yale.edu/tp/cas'>
    <cas:authenticationSuccess>
        <cas:user>%s</cas:user>
        <cas:attributes>
            <cas:email>%s</cas:email>
            <cas:displayName>%s</cas:displayName>
        </cas:attributes>
    </cas:authenticationSuccess>
</cas:serviceResponse>`, user.Username, user.Email, user.DisplayName)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "Ticket validated successfully",
			"data": gin.H{
				"user": gin.H{
					"username":    user.Username,
					"email":       user.Email,
					"displayName": user.DisplayName,
				},
			},
		})
	}

	logger.Info("CAS ticket validated successfully",
		zap.String("username", user.Username),
		zap.String("service", service),
		zap.String("ticket", ticket),
	)
}

// CASLogoutHandler CAS登出处理器
func CASLogoutHandler(c *gin.Context) {
	service := c.Query("service")

	logger.Info("CAS logout request",
		zap.String("service", service),
		zap.String("ip", c.ClientIP()),
	)

	// 清除session
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		cookie, err := c.Cookie("session_id")
		if err == nil {
			sessionID = cookie
		}
	}

	if sessionID != "" {
		sessionManager := GetSessionManager()
		if sessionManager != nil {
			ctx := c.Request.Context()
			sessionManager.DeleteSession(ctx, sessionID)
		}
		// 清除cookie
		c.SetCookie("session_id", "", -1, "/", "", false, true)
	}

	// 如果指定了service，重定向到service
	if service != "" {
		c.Redirect(http.StatusFound, service)
		return
	}

	// 否则返回登出成功页面
	c.HTML(http.StatusOK, "cas_logout.html", gin.H{
		"title": "CAS Logout",
	})
}

// generateServiceTicket 生成服务票据
func generateServiceTicket(user *models.User, service string, application *models.Application) (string, error) {
	// 生成随机票据
	ticketBytes := make([]byte, 32)
	if _, err := rand.Read(ticketBytes); err != nil {
		return "", err
	}
	ticket := "ST-" + base64.URLEncoding.EncodeToString(ticketBytes)

	// 创建服务票据记录
	serviceTicket := CASServiceTicket{
		Ticket:    ticket,
		Service:   service,
		UserID:    user.ID,
		Username:  user.Username,
		ExpiresAt: time.Now().Add(10 * time.Minute), // 10分钟过期
		Used:      false,
		Attributes: fmt.Sprintf(`{"email":"%s","displayName":"%s","organization":"%s"}`, 
			user.Email, user.DisplayName, user.OrganizationID),
	}

	if err := database.DB.Create(&serviceTicket).Error; err != nil {
		return "", err
	}

	logger.Info("Service ticket generated",
		zap.String("ticket", ticket),
		zap.String("username", user.Username),
		zap.String("service", service),
	)

	return ticket, nil
}

// buildRedirectURL 构建重定向URL
func buildRedirectURL(service, ticket string, params map[string]string) string {
	u, err := url.Parse(service)
	if err != nil {
		return service
	}

	query := u.Query()
	query.Set("ticket", ticket)
	
	if params != nil {
		for key, value := range params {
			query.Set(key, value)
		}
	}

	u.RawQuery = query.Encode()
	return u.String()
}

// CASServiceValidateHandler CAS 3.0服务验证处理器
func CASServiceValidateHandler(c *gin.Context) {
	service := c.Query("service")
	ticket := c.Query("ticket")
	pgtUrl := c.Query("pgtUrl")
	format := c.Query("format") // xml, json

	logger.Info("CAS service validate request",
		zap.String("service", service),
		zap.String("ticket", ticket),
		zap.String("pgtUrl", pgtUrl),
		zap.String("format", format),
		zap.String("ip", c.ClientIP()),
	)

	// 验证参数
	if service == "" || ticket == "" {
		if format == "xml" {
			c.Header("Content-Type", "application/xml")
			c.String(http.StatusOK, `<?xml version="1.0"?>
<cas:serviceResponse xmlns:cas='http://www.yale.edu/tp/cas'>
    <cas:authenticationFailure code='INVALID_REQUEST'>service and ticket parameters are required</cas:authenticationFailure>
</cas:serviceResponse>`)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "service and ticket parameters are required",
			})
		}
		return
	}

	// 查找并验证票据
	var serviceTicket CASServiceTicket
	if err := database.DB.Where("ticket = ? AND service = ? AND used = ? AND expires_at > ?", ticket, service, false, time.Now()).First(&serviceTicket).Error; err != nil {
		logger.ErrorError("Invalid or expired ticket", zap.Error(err))
		if format == "xml" {
			c.Header("Content-Type", "application/xml")
			c.String(http.StatusOK, `<?xml version="1.0"?>
<cas:serviceResponse xmlns:cas='http://www.yale.edu/tp/cas'>
    <cas:authenticationFailure code='INVALID_TICKET'>Ticket not recognized</cas:authenticationFailure>
</cas:serviceResponse>`)
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "Ticket not recognized",
			})
		}
		return
	}

	// 标记票据为已使用
	serviceTicket.Used = true
	database.DB.Save(&serviceTicket)

	// 获取用户信息
	var user models.User
	if err := database.DB.Where("id = ?", serviceTicket.UserID).First(&user).Error; err != nil {
		logger.ErrorError("User not found", zap.Error(err))
		if format == "xml" {
			c.Header("Content-Type", "application/xml")
			c.String(http.StatusOK, `<?xml version="1.0"?>
<cas:serviceResponse xmlns:cas='http://www.yale.edu/tp/cas'>
    <cas:authenticationFailure code='INTERNAL_ERROR'>User not found</cas:authenticationFailure>
</cas:serviceResponse>`)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "User not found",
			})
		}
		return
	}

	// 处理代理授权票据（如果提供）
	var pgtIou string
	if pgtUrl != "" {
		pgtIou = generateProxyGrantingTicket(&user, pgtUrl)
	}

	// 返回成功响应
	if format == "xml" {
		response := fmt.Sprintf(`<?xml version="1.0"?>
<cas:serviceResponse xmlns:cas='http://www.yale.edu/tp/cas'>
    <cas:authenticationSuccess>
        <cas:user>%s</cas:user>`, user.Username)

		if pgtIou != "" {
			response += fmt.Sprintf(`
        <cas:proxyGrantingTicket>%s</cas:proxyGrantingTicket>`, pgtIou)
		}

		response += fmt.Sprintf(`
        <cas:attributes>
            <cas:email>%s</cas:email>
            <cas:displayName>%s</cas:displayName>
        </cas:attributes>
    </cas:authenticationSuccess>
</cas:serviceResponse>`, user.Email, user.DisplayName)

		c.Header("Content-Type", "application/xml")
		c.String(http.StatusOK, response)
	} else {
		response := gin.H{
			"code":    200,
			"message": "Ticket validated successfully",
			"data": gin.H{
				"user": gin.H{
					"username":    user.Username,
					"email":       user.Email,
					"displayName": user.DisplayName,
				},
			},
		}

		if pgtIou != "" {
			response["data"].(gin.H)["proxyGrantingTicket"] = pgtIou
		}

		c.JSON(http.StatusOK, response)
	}

	logger.Info("CAS service ticket validated successfully",
		zap.String("username", user.Username),
		zap.String("service", service),
		zap.String("ticket", ticket),
	)
}

// generateProxyGrantingTicket 生成代理授权票据
func generateProxyGrantingTicket(user *models.User, pgtUrl string) string {
	// 生成随机票据
	ticketBytes := make([]byte, 32)
	if _, err := rand.Read(ticketBytes); err != nil {
		return ""
	}
	ticket := "PGT-" + base64.URLEncoding.EncodeToString(ticketBytes)

	// 创建代理授权票据记录
	pgt := CASProxyGrantingTicket{
		Ticket:    ticket,
		UserID:    user.ID,
		Username:  user.Username,
		ExpiresAt: time.Now().Add(2 * time.Hour), // 2小时过期
		Used:      false,
	}

	if err := database.DB.Create(&pgt).Error; err != nil {
		return ""
	}

	// 生成PGTIOU
	pgtIouBytes := make([]byte, 16)
	if _, err := rand.Read(pgtIouBytes); err != nil {
		return ""
	}
	pgtIou := "PGTIOU-" + base64.URLEncoding.EncodeToString(pgtIouBytes)

	// 这里应该向pgtUrl发送回调，包含pgt和pgtIou
	// 为了简化，我们直接返回pgtIou
	logger.Info("Proxy granting ticket generated",
		zap.String("ticket", ticket),
		zap.String("pgtIou", pgtIou),
		zap.String("username", user.Username),
	)

	return pgtIou
}
