package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Log        LogConfig        `mapstructure:"log"`
	Encryption EncryptionConfig `mapstructure:"encryption"`
	CORS       CORSConfig       `mapstructure:"cors"`
	Login      LoginConfig      `mapstructure:"login"`
	IdP        IdPConfig        `mapstructure:"idp"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string `mapstructure:"driver"`
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	Charset         string `mapstructure:"charset"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret             string `mapstructure:"secret"`
	AccessTokenExpire  int    `mapstructure:"access_token_expire"`
	RefreshTokenExpire int    `mapstructure:"refresh_token_expire"`
	Issuer             string `mapstructure:"issuer"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level        string   `mapstructure:"level"`
	Format       string   `mapstructure:"format"`
	Output       []string `mapstructure:"output"`  // stdout, file, 可以同时配置
	LogDir       string   `mapstructure:"log_dir"` // 日志目录，默认为logs
	MaxSize      int      `mapstructure:"max_size"`
	MaxBackups   int      `mapstructure:"max_backups"`
	MaxAge       int      `mapstructure:"max_age"`
	Compress     bool     `mapstructure:"compress"`
	RotateByDate bool     `mapstructure:"rotate_by_date"`
}

// EncryptionConfig 加密配置
type EncryptionConfig struct {
	BcryptCost int `mapstructure:"bcrypt_cost"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

// LoginConfig 登录配置
type LoginConfig struct {
	DefaultMethod    string `mapstructure:"default_method"`
	EnableOTP        bool   `mapstructure:"enable_otp"`
	EnableThirdParty bool   `mapstructure:"enable_third_party"`
}

// IdPConfig IdP配置（用于对接外部SP）
type IdPConfig struct {
	BaseURL               string `mapstructure:"base_url"`
	DefaultSessionTimeout int    `mapstructure:"default_session_timeout"`
	MaxConcurrentSessions int    `mapstructure:"max_concurrent_sessions"`
	EnableSingleLogout    bool   `mapstructure:"enable_single_logout"`
	EnableRememberMe      bool   `mapstructure:"enable_remember_me"`
	RememberMeDuration    int    `mapstructure:"remember_me_duration"`
}

var AppConfig *Config

// LoadConfig 加载配置
func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigType("yaml")

	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath("./config")
		viper.AddConfigPath(".")
	}

	// 环境变量支持
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Failed to read config file: %v", err)
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Printf("Failed to parse config file: %v", err)
		return nil, err
	}

	AppConfig = &config
	log.Printf("Config file loaded successfully: %s", viper.ConfigFileUsed())
	return &config, nil
}

// GetConfig 获取配置实例
func GetConfig() *Config {
	return AppConfig
}
