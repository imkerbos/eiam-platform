package handlers

import (
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

// LaunchApplicationHandler 应用启动处理器 - 支持多种协议的应用跳转
func LaunchApplicationHandler(c *gin.Context) {
	appID := c.Param("id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Application ID is required",
		})
		return
	}

	// 获取当前用户信息（从JWT token或session）
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "User not authenticated",
		})
		return
	}

	// 查询应用信息
	var application models.Application
	if err := database.DB.Where("id = ? AND status = ?", appID, models.StatusActive).First(&application).Error; err != nil {
		logger.Error("Application not found", zap.String("app_id", appID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "Application not found or inactive",
		})
		return
	}

	// 获取用户信息
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		logger.Error("User not found", zap.String("user_id", userID), zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "User not found",
		})
		return
	}

	logger.Info("User launching application",
		zap.String("user_id", user.ID),
		zap.String("username", user.Username),
		zap.String("app_id", application.ID),
		zap.String("app_name", application.Name),
		zap.String("protocol", application.Protocol),
		zap.String("ip", c.ClientIP()),
	)

	// 根据应用协议类型进行不同的处理
	switch application.Protocol {
	case "saml":
		handleSAMLAppLaunch(c, &user, &application)
	case "cas":
		handleCASAppLaunch(c, &user, &application)
	case "oauth2", "oidc":
		handleOAuth2AppLaunch(c, &user, &application)
	case "ldap":
		handleDirectAppLaunch(c, &user, &application)
	default:
		// 直接跳转（无SSO）
		handleDirectAppLaunch(c, &user, &application)
	}
}

// handleSAMLAppLaunch 处理SAML应用启动 (IdP-initiated SSO)
func handleSAMLAppLaunch(c *gin.Context, user *models.User, app *models.Application) {
	if app.AcsURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "SAML ACS URL not configured",
		})
		return
	}

	// 生成SAML断言并跳转到SP
	// 这里需要调用SAML IdP库生成断言
	samlResponse, err := generateSAMLResponseForUser(user, app)
	if err != nil {
		logger.Error("Failed to generate SAML response", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to generate SAML response",
		})
		return
	}

	// 记录应用访问日志
	recordApplicationAccess(user.ID, app.ID, "saml", c.ClientIP())

	// 返回SAML POST表单，让浏览器自动提交到SP
	c.HTML(http.StatusOK, "saml_post_form.html", gin.H{
		"AcsURL":       app.AcsURL,
		"SAMLResponse": samlResponse,
		"RelayState":   "", // 可选的状态信息
	})
}

// handleCASAppLaunch 处理CAS应用启动
func handleCASAppLaunch(c *gin.Context, user *models.User, app *models.Application) {
	if app.ServiceURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "CAS Service URL not configured",
		})
		return
	}

	// 生成CAS服务票据
	ticket, err := casTicketManager.GenerateServiceTicket(user.Username, app.ServiceURL)
	if err != nil {
		logger.Error("Failed to generate CAS ticket", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to generate CAS ticket",
		})
		return
	}

	// 构建重定向URL
	redirectURL := buildAppLaunchCASRedirectURL(app.ServiceURL, ticket)

	// 记录应用访问日志
	recordApplicationAccess(user.ID, app.ID, "cas", c.ClientIP())

	logger.Info("CAS application launch",
		zap.String("username", user.Username),
		zap.String("service_url", app.ServiceURL),
		zap.String("ticket", ticket),
	)

	// 重定向到CAS服务
	c.Redirect(http.StatusFound, redirectURL)
}

// handleOAuth2AppLaunch 处理OAuth2/OIDC应用启动
func handleOAuth2AppLaunch(c *gin.Context, user *models.User, app *models.Application) {
	// OAuth2/OIDC通常需要authorization code flow
	// 这里简化处理，实际应该实现完整的OAuth2流程

	if app.RedirectURIs == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "OAuth2 redirect URIs not configured",
		})
		return
	}

	// 生成授权码（简化实现）
	authCode := utils.GenerateTradeIDString("auth")
	state := c.Query("state") // 从查询参数获取state

	// 构建重定向URL
	redirectURL := fmt.Sprintf("%s?code=%s&state=%s", app.RedirectURIs, authCode, state)

	// 记录应用访问日志
	recordApplicationAccess(user.ID, app.ID, "oauth2", c.ClientIP())

	logger.Info("OAuth2 application launch",
		zap.String("username", user.Username),
		zap.String("redirect_uri", app.RedirectURIs),
		zap.String("auth_code", authCode),
	)

	// 重定向到OAuth2客户端
	c.Redirect(http.StatusFound, redirectURL)
}

// handleDirectAppLaunch 处理直接跳转（无SSO）
func handleDirectAppLaunch(c *gin.Context, user *models.User, app *models.Application) {
	if app.HomePageURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Application homepage URL not configured",
		})
		return
	}

	// 记录应用访问日志
	recordApplicationAccess(user.ID, app.ID, "direct", c.ClientIP())

	logger.Info("Direct application launch",
		zap.String("username", user.Username),
		zap.String("homepage_url", app.HomePageURL),
	)

	// 直接重定向到应用首页
	c.Redirect(http.StatusFound, app.HomePageURL)
}

// generateSAMLResponseForUser 为用户生成SAML响应（IdP-initiated）
func generateSAMLResponseForUser(user *models.User, app *models.Application) (string, error) {
	if samlIDP == nil {
		return "", fmt.Errorf("SAML IdP not initialized")
	}

	// 生成SAML响应XML
	samlResponseXML, err := createSAMLResponseXML(user, app)
	if err != nil {
		return "", fmt.Errorf("failed to create SAML response XML: %w", err)
	}

	// Base64编码
	samlResponseB64 := utils.Base64Encode([]byte(samlResponseXML))

	logger.Info("SAML response generated successfully",
		zap.String("username", user.Username),
		zap.String("app_entity_id", app.EntityID),
		zap.String("acs_url", app.AcsURL),
	)

	return samlResponseB64, nil
}

// recordApplicationAccess 记录应用访问日志
func recordApplicationAccess(userID, appID, protocol, clientIP string) {
	// 这里可以记录到数据库或日志文件
	logger.Info("Application access recorded",
		zap.String("user_id", userID),
		zap.String("app_id", appID),
		zap.String("protocol", protocol),
		zap.String("client_ip", clientIP),
		zap.Time("access_time", time.Now()),
	)

	// TODO: 可以在这里更新应用访问统计
}

// buildAppLaunchCASRedirectURL 构建CAS重定向URL（应用启动专用）
func buildAppLaunchCASRedirectURL(serviceURL, ticket string) string {
	u, err := url.Parse(serviceURL)
	if err != nil {
		return serviceURL + "?ticket=" + ticket
	}

	query := u.Query()
	query.Set("ticket", ticket)
	u.RawQuery = query.Encode()
	return u.String()
}
