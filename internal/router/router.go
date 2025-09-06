package router

import (
	"net/http"
	"time"

	"eiam-platform/config"
	"eiam-platform/internal/handlers"
	"eiam-platform/internal/middleware"
	"eiam-platform/pkg/database"
	"eiam-platform/pkg/i18n"
	"eiam-platform/pkg/logger"
	"eiam-platform/pkg/redis"
	"eiam-platform/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupRouter 设置路由
func SetupRouter(cfg *config.Config, jwtManager *utils.JWTManager) *gin.Engine {
	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 创建Gin引擎
	r := gin.New()

	// 获取sessionManager实例
	sessionManager := handlers.GetSessionManager()

	// 添加中间件
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.TradeIDMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.CORSMiddleware(&cfg.CORS))
	r.Use(middleware.SecurityHeadersMiddleware())

	// Log service startup
	logger.ServiceInfo("EIAM IdP platform started",
		zap.String("version", "1.0.0"),
		zap.String("mode", cfg.Server.Mode),
		zap.String("host", cfg.Server.Host),
		zap.Int("port", cfg.Server.Port),
	)

	// 健康检查端点
	r.GET("/health", healthCheckHandler)

	// 静态文件服务 - 头像上传
	r.Static("/uploads", "./uploads")

	// 公开API端点（不需要认证）
	r.GET("/public/site-info", handlers.GetPublicSiteInfoHandler)

	// API路由组
	api := r.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			// 公开API端点（不需要认证）
			v1.GET("/public/site-info", handlers.GetPublicSiteInfoHandler)
			// 通用认证API
			auth := v1.Group("/auth")
			{
				auth.POST("/login", handlers.LoginHandler)
				auth.POST("/refresh", handlers.RefreshTokenHandler)
				auth.POST("/logout", middleware.AuthMiddleware(jwtManager, sessionManager), handlers.LogoutHandler)
			}

			// Console管理后台API
			console := v1.Group("/console")
			{
				setupConsoleRoutes(console, jwtManager)
			}

			// Portal用户端API
			portal := v1.Group("/portal")
			{
				setupPortalRoutes(portal, jwtManager)
			}
		}
	}

	return r
}

// setupConsoleRoutes 设置Console管理后台路由
func setupConsoleRoutes(console *gin.RouterGroup, jwtManager *utils.JWTManager) {
	// 获取sessionManager实例
	sessionManager := handlers.GetSessionManager()
	// 管理员认证
	auth := console.Group("/auth")
	{
		auth.POST("/login", handlers.ConsoleLoginHandler)
		auth.POST("/logout", middleware.AuthMiddleware(jwtManager, sessionManager), handlers.LogoutHandler)
		auth.POST("/refresh", handlers.ConsoleRefreshTokenHandler)
		auth.GET("/me", middleware.AuthMiddleware(jwtManager, sessionManager), handlers.ConsoleGetMeHandler)
	}

	// 会话管理（需要管理员权限）
	sessions := console.Group("/sessions")
	sessions.Use(middleware.AuthMiddleware(jwtManager, sessionManager))
	sessions.Use(middleware.AdminMiddleware())
	{
		sessions.GET("", handlers.GetAllSessionsHandler)                        // 获取所有在线会话
		sessions.GET("/users/:userID", handlers.GetUserSessionsHandler)         // 获取用户会话列表
		sessions.DELETE("/users/:userID", handlers.ForceLogoutUserHandler)      // 强制用户下线
		sessions.POST("/force-logout-all", handlers.ForceLogoutAllUsersHandler) // 强制所有用户下线
	}

	// 用户管理（需要管理员权限）
	users := console.Group("/users")
	users.Use(middleware.AuthMiddleware(jwtManager, sessionManager))
	users.Use(middleware.AdminMiddleware())
	{
		users.GET("", handlers.GetUsersHandler)
		users.POST("", handlers.CreateUserHandler)
		users.GET("/:id", handlers.GetUserHandler)
		users.PUT("/:id", handlers.UpdateUserHandler)
		users.DELETE("/:id", handlers.DeleteUserHandler)
	}

	// 组织管理（需要管理员权限）
	organizations := console.Group("/organizations")
	organizations.Use(middleware.AuthMiddleware(jwtManager, sessionManager))
	organizations.Use(middleware.AdminMiddleware())
	{
		organizations.GET("", handlers.GetOrganizationsHandler)
		organizations.GET("/tree", handlers.GetOrganizationsTreeHandler)
		organizations.POST("", handlers.CreateOrganizationHandler)
		organizations.GET("/:id", handlers.GetOrganizationHandler)
		organizations.PUT("/:id", handlers.UpdateOrganizationHandler)
		organizations.DELETE("/:id", handlers.DeleteOrganizationHandler)
	}

	// 角色管理（需要管理员权限）
	roles := console.Group("/roles")
	roles.Use(middleware.AuthMiddleware(jwtManager, sessionManager))
	roles.Use(middleware.AdminMiddleware())
	{
		roles.GET("", handlers.GetRolesHandler)
		roles.POST("", handlers.CreateRoleHandler)
		roles.PUT("/:id", handlers.UpdateRoleHandler)
		roles.DELETE("/:id", handlers.DeleteRoleHandler)
	}

	// 管理员管理（需要管理员权限）
	administrators := console.Group("/administrators")
	administrators.Use(middleware.AuthMiddleware(jwtManager, sessionManager))
	administrators.Use(middleware.AdminMiddleware())
	{
		administrators.GET("", handlers.GetAdministratorsHandler)
		administrators.POST("/assign", handlers.AssignAdministratorRoleHandler)
		administrators.DELETE("/:userID/:roleID", handlers.RemoveAdministratorRoleHandler)
	}

	// 权限管理（需要管理员权限）
	permissions := console.Group("/permissions")
	permissions.Use(middleware.AuthMiddleware(jwtManager, sessionManager))
	permissions.Use(middleware.AdminMiddleware())
	{
		permissions.GET("", handlers.GetPermissionsHandler)
		permissions.POST("", handlers.CreatePermissionHandler)
		permissions.PUT("/:id", handlers.UpdatePermissionHandler)
		permissions.DELETE("/:id", handlers.DeletePermissionHandler)
	}

	// 角色分配管理（需要管理员权限）
	roleAssignments := console.Group("/role-assignments")
	roleAssignments.Use(middleware.AuthMiddleware(jwtManager, sessionManager))
	roleAssignments.Use(middleware.AdminMiddleware())
	{
		roleAssignments.GET("", handlers.GetRoleAssignmentsHandler)
		roleAssignments.POST("", handlers.AssignRoleToUserHandler)
		roleAssignments.DELETE("/:userID/:roleID", handlers.RemoveRoleFromUserHandler)
	}

	// 应用管理（需要管理员权限）
	applications := console.Group("/applications")
	applications.Use(middleware.AuthMiddleware(jwtManager, sessionManager))
	applications.Use(middleware.AdminMiddleware())
	{
		applications.GET("", handlers.GetApplicationsHandler)
		applications.POST("", handlers.CreateApplicationHandler)
		applications.PUT("/:id", handlers.UpdateApplicationHandler)
		applications.DELETE("/:id", handlers.DeleteApplicationHandler)
	}

	// 应用分组管理（需要管理员权限）
	appGroups := console.Group("/application-groups")
	appGroups.Use(middleware.AuthMiddleware(jwtManager, sessionManager))
	appGroups.Use(middleware.AdminMiddleware())
	{
		appGroups.GET("", handlers.GetApplicationGroupsHandler)
		appGroups.POST("", handlers.CreateApplicationGroupHandler)
		appGroups.PUT("/:id", handlers.UpdateApplicationGroupHandler)
		appGroups.DELETE("/:id", handlers.DeleteApplicationGroupHandler)
	}

	// 添加别名路由以支持前端调用
	applicationsAlias := console.Group("/applications")
	applicationsAlias.Use(middleware.AuthMiddleware(jwtManager, sessionManager))
	applicationsAlias.Use(middleware.AdminMiddleware())
	{
		applicationsAlias.GET("/groups", handlers.GetApplicationGroupsHandler)
	}

	// 系统设置管理（需要管理员权限）
	system := console.Group("/system")
	system.Use(middleware.AuthMiddleware(jwtManager, sessionManager))
	system.Use(middleware.AdminMiddleware())
	{
		system.GET("/settings", handlers.GetSystemSettingsHandler)
		system.PUT("/settings", handlers.UpdateSystemSettingsHandler)
		system.GET("/site-settings", handlers.GetSiteSettingsHandler)
		system.PUT("/site-settings", handlers.UpdateSiteSettingsHandler)
		system.GET("/security-settings", handlers.GetSecuritySettingsHandler)
		system.PUT("/security-settings", handlers.UpdateSecuritySettingsHandler)
		system.POST("/upload-logo", handlers.UploadLogoHandler)
	}

	// 密码策略管理（需要管理员权限）
	passwordPolicy := console.Group("/password-policy")
	passwordPolicy.Use(middleware.AuthMiddleware(jwtManager, sessionManager))
	passwordPolicy.Use(middleware.AdminMiddleware())
	{
		passwordPolicy.GET("", handlers.GetPasswordPolicyHandler)
		passwordPolicy.PUT("", handlers.UpdatePasswordPolicyHandler)
		passwordPolicy.POST("/validate", handlers.ValidatePasswordHandler)
		passwordPolicy.POST("/generate", handlers.GeneratePasswordHandler)
	}

	// 系统API（需要管理员权限）
	systemAPI := console.Group("/dashboard")
	systemAPI.Use(middleware.AuthMiddleware(jwtManager, sessionManager))
	systemAPI.Use(middleware.AdminMiddleware())
	{
		systemAPI.GET("", handlers.GetDashboardData)
		systemAPI.GET("/stats", handlers.GetSystemStats)
		systemAPI.GET("/activities", handlers.GetRecentActivities)
		systemAPI.GET("/top-users", handlers.GetTopLoginUsers)
		systemAPI.GET("/top-applications", handlers.GetTopLoginApplications)
	}

	// 日志管理（需要管理员权限）
	logs := console.Group("/logs")
	logs.Use(middleware.AuthMiddleware(jwtManager, sessionManager))
	logs.Use(middleware.AdminMiddleware())
	{
		logs.GET("/login", handlers.GetLoginLogsHandler)
		logs.GET("/audit", handlers.GetAuditLogsHandler)
	}
}

// setupPortalRoutes 设置Portal用户端路由
func setupPortalRoutes(portal *gin.RouterGroup, jwtManager *utils.JWTManager) {
	// 获取sessionManager实例
	sessionManager := handlers.GetSessionManager()
	// 用户认证
	auth := portal.Group("/auth")
	{
		auth.POST("/login", handlers.PortalLoginHandler)
		auth.POST("/logout", middleware.AuthMiddleware(jwtManager, sessionManager), handlers.PortalLogoutHandler)
		auth.POST("/refresh", handlers.PortalRefreshTokenHandler)
		auth.GET("/me", middleware.AuthMiddleware(jwtManager, sessionManager), handlers.PortalGetMeHandler)
	}

	// OTP相关
	otp := portal.Group("/otp")
	{
		otp.POST("/send", handlers.SendOTPHandler)
		otp.POST("/verify", handlers.VerifyOTPHandler)
	}

	// 密码管理
	password := portal.Group("/password")
	{
		password.POST("/forgot", handlers.ForgotPasswordHandler)
		password.POST("/reset", handlers.ResetPasswordHandler)
		password.PUT("/change", middleware.AuthMiddleware(jwtManager, sessionManager), handlers.ChangePasswordHandler)
	}

	// 用户资料管理（需要认证）
	profile := portal.Group("/profile")
	profile.Use(middleware.AuthMiddleware(jwtManager, sessionManager))
	{
		profile.GET("", handlers.GetProfileHandler)
		profile.PUT("", handlers.UpdateProfileHandler)
		profile.POST("/avatar", handlers.UploadAvatarHandler)
		profile.PUT("/password", handlers.ChangePasswordHandler)
		profile.POST("/verify-email", handlers.VerifyEmailHandler)
		profile.POST("/setup-otp", handlers.SetupOTPHandler)
		profile.POST("/disable-otp", handlers.DisableOTPHandler)
		profile.GET("/backup-codes", handlers.GetBackupCodesHandler)
	}

	// OTP设置（需要认证）
	otpSettings := portal.Group("/otp-settings")
	otpSettings.Use(middleware.AuthMiddleware(jwtManager, sessionManager))
	{
		otpSettings.POST("/enable", handlers.EnableOTPHandler)
		otpSettings.POST("/disable", handlers.DisableOTPHandler)
	}

	// 用户应用（需要认证）
	userApps := portal.Group("/applications")
	userApps.Use(middleware.AuthMiddleware(jwtManager, sessionManager))
	{
		userApps.GET("", handlers.GetUserApplicationsHandler)
		userApps.GET("/:id", handlers.GetUserApplicationHandler)
	}
}

// healthCheckHandler health check handler
func healthCheckHandler(c *gin.Context) {
	// Check database connection
	if err := database.HealthCheck(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":   i18n.StatusUnhealthy,
			"database": i18n.StatusDisconnected,
			"error":    err.Error(),
			"trade_id": c.GetString("trade_id"),
		})
		return
	}

	// Check Redis connection
	if err := redis.HealthCheck(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":   i18n.StatusUnhealthy,
			"redis":    i18n.StatusDisconnected,
			"error":    err.Error(),
			"trade_id": c.GetString("trade_id"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    i18n.StatusHealthy,
		"timestamp": time.Now(),
		"version":   "1.0.0",
		"database":  i18n.StatusConnected,
		"redis":     i18n.StatusConnected,
		"trade_id":  c.GetString("trade_id"),
	})
}
