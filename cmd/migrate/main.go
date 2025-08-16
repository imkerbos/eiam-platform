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

	logger.ServiceInfo("Database migration started")

	// Initialize database connection
	if err := database.InitDatabase(&cfg.Database); err != nil {
		logger.ErrorFatal("Database initialization failed", zap.Error(err))
	}

	// Execute migration
	if err := migrate(); err != nil {
		logger.ErrorFatal("Database migration failed", zap.Error(err))
	}

	logger.ServiceInfo("Database migration completed")
}

func migrate() error {
	// Disable foreign key checks temporarily
	if err := database.DB.Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
		return fmt.Errorf("failed to disable foreign key checks: %v", err)
	}

	// Core table migration in correct order
	coreTables := []interface{}{
		&models.User{},
		&models.UserProfile{},
		&models.UserSession{},
		&models.UserLoginLog{},
		&models.UserOTPRecord{},
		&models.AuditLog{},
		&models.Organization{},
		&models.Role{},
		&models.Permission{},
		&models.ApplicationGroup{},
		&models.Application{},
		&models.SystemSetting{},
	}

	// Phase 2 tables (commented for now)
	// oauth2Tables := []interface{}{
	// 	&models.OAuth2AuthorizationCode{},
	// 	&models.OAuth2AccessToken{},
	// 	&models.SAMLAssertion{},
	// }

	logger.ServiceInfo("Starting core table migration")
	for i, table := range coreTables {
		logger.ServiceInfo(fmt.Sprintf("Migrating table %d/%d", i+1, len(coreTables)))
		if err := database.DB.AutoMigrate(table); err != nil {
			return fmt.Errorf("table migration failed: %v", err)
		}
	}

	// Re-enable foreign key checks
	if err := database.DB.Exec("SET FOREIGN_KEY_CHECKS = 1").Error; err != nil {
		return fmt.Errorf("failed to enable foreign key checks: %v", err)
	}

	logger.ServiceInfo("Core table migration completed")
	return nil
}
