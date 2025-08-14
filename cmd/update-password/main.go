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
	var username string
	var newPassword string
	flag.StringVar(&configPath, "config", "config/config.yaml", "Configuration file path")
	flag.StringVar(&username, "username", "admin", "Username to update")
	flag.StringVar(&newPassword, "password", "admin123", "New password")
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

	logger.ServiceInfo("Password update started")

	// Initialize database connection
	if err := database.InitDatabase(&cfg.Database); err != nil {
		logger.ErrorFatal("Database initialization failed", zap.Error(err))
	}

	// Update password
	if err := updatePassword(username, newPassword); err != nil {
		logger.ErrorFatal("Password update failed", zap.Error(err))
	}

	logger.ServiceInfo("Password update completed")
}

func updatePassword(username, password string) error {
	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("user %s not found", username)
		}
		return fmt.Errorf("failed to find user: %v", err)
	}

	// 生成新的MD5密码
	hashedPassword, err := utils.HashPassword(password, 12)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	// 更新密码
	if err := database.DB.Model(&user).Update("password", hashedPassword).Error; err != nil {
		return fmt.Errorf("failed to update password: %v", err)
	}

	logger.Info("Password updated successfully",
		zap.String("username", username),
		zap.String("user_id", user.ID),
	)

	fmt.Printf("Password for user %s has been updated successfully\n", username)
	fmt.Printf("New password hash: %s\n", hashedPassword)
	return nil
}
