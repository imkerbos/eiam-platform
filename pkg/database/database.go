package database

import (
	"fmt"
	"time"

	"eiam-platform/config"
	"eiam-platform/pkg/i18n"
	"eiam-platform/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase(cfg *config.DatabaseConfig) error {
	var dialector gorm.Dialector

	switch cfg.Driver {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			cfg.Username,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.DBName,
			cfg.Charset,
		)
		dialector = mysql.Open(dsn)
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
			cfg.Host,
			cfg.Username,
			cfg.Password,
			cfg.DBName,
			cfg.Port,
		)
		dialector = postgres.Open(dsn)
	default:
		return fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	// GORM配置
	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",
			SingularTable: false,
		},
		Logger: gormLogger.Default.LogMode(gormLogger.Info),
	}

	var err error
	DB, err = gorm.Open(dialector, gormConfig)
	if err != nil {
		return fmt.Errorf("database connection failed: %v", err)
	}

	// Get underlying database connection and configure connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection pool: %v", err)
	}

	// 配置连接池
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	// Test database connection
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database connection test failed: %v", err)
	}

	logger.Info(i18n.DatabaseConnected,
		zap.String("driver", cfg.Driver),
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.DBName),
	)

	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// Close 关闭数据库连接
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// Transaction 执行事务
func Transaction(fn func(tx *gorm.DB) error) error {
	return DB.Transaction(fn)
}

// Migrate 执行数据库迁移
func Migrate(models ...interface{}) error {
	if err := DB.AutoMigrate(models...); err != nil {
		logger.Error("数据库迁移失败", zap.Error(err))
		return err
	}

	logger.Info("数据库迁移完成")
	return nil
}

// HealthCheck 健康检查
func HealthCheck() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
