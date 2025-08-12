package main

import (
	"flag"
	"fmt"
	"log"

	"eiam-platform/config"
	"eiam-platform/internal/models"
	"eiam-platform/pkg/database"
	"eiam-platform/pkg/logger"
	"eiam-platform/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {
	// Parse command line arguments
	var configPath string
	flag.StringVar(&configPath, "config", "config/config.yaml", "Configuration file path")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration file: %v", err)
	}

	// Initialize logger
	if err := logger.InitLogger(&cfg.Log); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	logger.ServiceInfo("Database seeding started")

	// Initialize database connection
	if err := database.InitDatabase(&cfg.Database); err != nil {
		logger.ErrorFatal("Database initialization failed", zap.Error(err))
	}

	// Execute seeding
	if err := seed(); err != nil {
		logger.ErrorFatal("Database seeding failed", zap.Error(err))
	}

	logger.ServiceInfo("Database seeding completed")
}

func seed() error {
	// 检查用户是否已存在
	var user models.User
	if err := database.DB.Where("username = ?", "admin").First(&user).Error; err == nil {
		logger.Info("Admin user already exists, using existing user", zap.String("user_id", user.ID))
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check existing user: %v", err)
	} else {
		// 创建新用户
		hashedPassword, err := utils.HashPassword("admin123", 12)
		if err != nil {
			return fmt.Errorf("failed to hash password: %v", err)
		}

		salt, err := utils.GenerateSalt(16)
		if err != nil {
			return fmt.Errorf("failed to generate salt: %v", err)
		}

		user = models.User{
			BaseModel: models.BaseModel{
				ID: "user-001",
			},
			Username:      "admin",
			Email:         "admin@example.com",
			DisplayName:   "系统管理员",
			Password:      hashedPassword,
			Salt:          salt,
			Status:        models.StatusActive,
			EmailVerified: true,
			PhoneVerified: false,
			EnableOTP:     false,
		}

		if err := database.DB.Create(&user).Error; err != nil {
			return fmt.Errorf("failed to create user: %v", err)
		}

		logger.Info("Created test user", zap.String("user_id", user.ID), zap.String("username", user.Username))
	}

	// 检查组织是否已存在
	var org models.Organization
	if err := database.DB.Where("code = ?", "TEST_ORG").First(&org).Error; err == nil {
		logger.Info("Test organization already exists, using existing organization", zap.String("org_id", org.ID))
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check existing organization: %v", err)
	} else {
		// 创建测试组织
		org = models.Organization{
			BaseModel: models.BaseModel{
				ID: "org-001",
			},
			Name:        "测试组织",
			Code:        "TEST_ORG",
			Type:        models.OrgTypeHeadquarters,
			Level:       1,
			Path:        "/1",
			Sort:        1,
			Description: "测试用的根组织",
			Manager:     user.ID,
			Status:      models.StatusActive,
		}

		if err := database.DB.Create(&org).Error; err != nil {
			return fmt.Errorf("failed to create organization: %v", err)
		}

		logger.Info("Created test organization", zap.String("org_id", org.ID))

		// 更新用户的组织ID
		if err := database.DB.Model(&user).Update("organization_id", org.ID).Error; err != nil {
			return fmt.Errorf("failed to update user organization: %v", err)
		}
	}

	// 检查角色是否已存在
	var adminRole models.Role
	if err := database.DB.Where("code = ?", "SYSTEM_ADMIN").First(&adminRole).Error; err == nil {
		logger.Info("Admin role already exists, using existing role", zap.String("role_id", adminRole.ID))
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check existing role: %v", err)
	} else {
		// 创建系统角色
		adminRole = models.Role{
			BaseModel: models.BaseModel{
				ID: "role-001",
			},
			Name:        "系统管理员",
			Code:        "SYSTEM_ADMIN",
			Description: "系统管理员角色，拥有所有权限",
			Type:        "system",
			IsSystem:    true,
			Scope:       "global",
			Status:      models.StatusActive,
		}

		if err := database.DB.Create(&adminRole).Error; err != nil {
			return fmt.Errorf("failed to create admin role: %v", err)
		}

		logger.Info("Created admin role", zap.String("role_id", adminRole.ID))
	}

	// 创建基础权限
	permissions := []models.Permission{
		{
			BaseModel: models.BaseModel{
				ID: "perm-001",
			},
			Name:        "用户管理",
			Code:        "user:manage",
			Resource:    "user",
			Action:      "manage",
			Description: "用户管理权限",
			Category:    "system",
			IsSystem:    true,
			Status:      models.StatusActive,
		},
		{
			BaseModel: models.BaseModel{
				ID: "perm-002",
			},
			Name:        "组织管理",
			Code:        "organization:manage",
			Resource:    "organization",
			Action:      "manage",
			Description: "组织管理权限",
			Category:    "system",
			IsSystem:    true,
			Status:      models.StatusActive,
		},
		{
			BaseModel: models.BaseModel{
				ID: "perm-003",
			},
			Name:        "系统设置",
			Code:        "system:manage",
			Resource:    "system",
			Action:      "manage",
			Description: "系统设置权限",
			Category:    "system",
			IsSystem:    true,
			Status:      models.StatusActive,
		},
	}

	for _, perm := range permissions {
		var existingPerm models.Permission
		if err := database.DB.Where("code = ?", perm.Code).First(&existingPerm).Error; err == nil {
			logger.Info("Permission already exists", zap.String("perm_id", existingPerm.ID), zap.String("code", perm.Code))
		} else if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to check existing permission %s: %v", perm.Code, err)
		} else {
			if err := database.DB.Create(&perm).Error; err != nil {
				return fmt.Errorf("failed to create permission %s: %v", perm.Code, err)
			}
			logger.Info("Created permission", zap.String("perm_id", perm.ID), zap.String("code", perm.Code))
		}
	}

	// 检查用户角色关系是否已存在
	var userRoleCount int64
	database.DB.Table("user_roles").Where("user_id = ? AND role_id = ?", user.ID, adminRole.ID).Count(&userRoleCount)
	if userRoleCount == 0 {
		// 为用户分配管理员角色
		userRole := struct {
			UserID string `gorm:"column:user_id"`
			RoleID string `gorm:"column:role_id"`
		}{
			UserID: user.ID,
			RoleID: adminRole.ID,
		}

		if err := database.DB.Table("user_roles").Create(&userRole).Error; err != nil {
			return fmt.Errorf("failed to assign role to user: %v", err)
		}

		logger.Info("Assigned admin role to user", zap.String("user_id", user.ID), zap.String("role_id", adminRole.ID))
	} else {
		logger.Info("User role relationship already exists")
	}

	// 为角色分配权限
	for _, perm := range permissions {
		var rolePermCount int64
		database.DB.Table("role_permissions").Where("role_id = ? AND permission_id = ?", adminRole.ID, perm.ID).Count(&rolePermCount)
		if rolePermCount == 0 {
			rolePerm := struct {
				RoleID       string `gorm:"column:role_id"`
				PermissionID string `gorm:"column:permission_id"`
			}{
				RoleID:       adminRole.ID,
				PermissionID: perm.ID,
			}

			if err := database.DB.Table("role_permissions").Create(&rolePerm).Error; err != nil {
				return fmt.Errorf("failed to assign permission %s to role: %v", perm.Code, err)
			}
			logger.Info("Assigned permission to role", zap.String("role_id", adminRole.ID), zap.String("perm_id", perm.ID))
		} else {
			logger.Info("Role permission relationship already exists", zap.String("role_id", adminRole.ID), zap.String("perm_id", perm.ID))
		}
	}

	return nil
}
