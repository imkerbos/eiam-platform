package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"eiam-platform/config"
	"eiam-platform/internal/router"
	"eiam-platform/pkg/database"
	"eiam-platform/pkg/logger"
	"eiam-platform/pkg/redis"
	"eiam-platform/pkg/utils"

	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		panic("Failed to load configuration file: " + err.Error())
	}

	// Initialize logger
	if err := logger.InitLogger(&cfg.Log); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	logger.ServiceInfo("EIAM IdP platform starting...")

	// Initialize database connection
	if err := database.InitDatabase(&cfg.Database); err != nil {
		logger.ErrorFatal("Database initialization failed", zap.Error(err))
	}

	// Initialize Redis
	if err := redis.InitRedis(&cfg.Redis); err != nil {
		logger.ErrorFatal("Redis initialization failed", zap.Error(err))
	}

	// Initialize JWT manager
	jwtManager := utils.NewJWTManager(&cfg.JWT)

	// Setup router
	r := router.SetupRouter(cfg, jwtManager)

	// Create HTTP server
	srv := &http.Server{
		Addr:    cfg.Server.Host + ":" + fmt.Sprintf("%d", cfg.Server.Port),
		Handler: r,
	}

	// Start server
	go func() {
		logger.ServiceInfo("Server starting",
			zap.String("host", cfg.Server.Host),
			zap.Int("port", cfg.Server.Port),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.ErrorFatal("Server startup failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.ServiceInfo("Server shutting down...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.ErrorFatal("Server forced shutdown", zap.Error(err))
	}

	// Close database connection
	if err := database.Close(); err != nil {
		logger.ErrorError("Failed to close database connection", zap.Error(err))
	}

	// Close Redis connection
	if err := redis.Close(); err != nil {
		logger.ErrorError("Failed to close Redis connection", zap.Error(err))
	}

	logger.ServiceInfo("Server shutdown complete")
}
