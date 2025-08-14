package main

import (
	"flag"
	"fmt"
	"log"

	"eiam-platform/config"
	"eiam-platform/internal/models"
	"eiam-platform/pkg/database"
	"eiam-platform/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {
	// Parse command line arguments
	var configPath string
	var username string
	flag.StringVar(&configPath, "config", "config/config.yaml", "Configuration file path")
	flag.StringVar(&username, "username", "admin", "Username to reset")
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

	logger.ServiceInfo("User reset started")

	// Initialize database connection
	if err := database.InitDatabase(&cfg.Database); err != nil {
		logger.ErrorFatal("Database initialization failed", zap.Error(err))
	}

	// Reset user
	if err := resetUser(username); err != nil {
		logger.ErrorFatal("User reset failed", zap.Error(err))
	}

	logger.ServiceInfo("User reset completed")
}

func resetUser(username string) error {
	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("user %s not found", username)
		}
		return fmt.Errorf("failed to find user: %v", err)
	}

	// Reset user status
	updates := map[string]interface{}{
		"failed_count": 0,
		"locked_until": nil,
		"status":       models.StatusActive,
	}

	if err := database.DB.Model(&user).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to reset user: %v", err)
	}

	logger.Info("User reset successfully",
		zap.String("username", username),
		zap.String("user_id", user.ID),
	)

	fmt.Printf("User %s has been reset successfully\n", username)
	return nil
}
