package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"eiam-platform/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	Logger     *zap.Logger
	Sugar      *zap.SugaredLogger
	AccessLog  *zap.Logger
	ErrorLog   *zap.Logger
	ServiceLog *zap.Logger
)

// InitLogger initialize logger
func InitLogger(cfg *config.LogConfig) error {
	// Configure log level
	level := zapcore.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	}

	// Configure encoder
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Ensure log directory exists
	if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
		return err
	}

	// Configure output
	var writeSyncers []zapcore.WriteSyncer

	// Check if output to stdout
	for _, output := range cfg.Output {
		if output == "stdout" {
			writeSyncers = append(writeSyncers, zapcore.AddSync(os.Stdout))
		}
	}

	// Create main log core
	mainCore := createLogCore(encoder, level, writeSyncers, cfg, "eiam.log")
	Logger = zap.New(mainCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	Sugar = Logger.Sugar()

	// Create access log
	accessCore := createLogCore(encoder, level, writeSyncers, cfg, "access.log")
	AccessLog = zap.New(accessCore, zap.AddCaller())

	// Create error log
	errorCore := createLogCore(encoder, zapcore.ErrorLevel, writeSyncers, cfg, "error.log")
	ErrorLog = zap.New(errorCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	// Create service log
	serviceCore := createLogCore(encoder, level, writeSyncers, cfg, "service.log")
	ServiceLog = zap.New(serviceCore, zap.AddCaller())

	return nil
}

// createLogCore create log core
func createLogCore(encoder zapcore.Encoder, level zapcore.Level, writeSyncers []zapcore.WriteSyncer, cfg *config.LogConfig, filename string) zapcore.Core {
	// Check if output to file
	for _, output := range cfg.Output {
		if output == "file" {
			var logFilename string
			if cfg.RotateByDate {
				// Rotate log files by date
				baseName := strings.TrimSuffix(filename, filepath.Ext(filename))
				logFilename = filepath.Join(cfg.LogDir, fmt.Sprintf("%s.%s%s",
					baseName,
					time.Now().Format("2006-01-02"),
					filepath.Ext(filename)))
			} else {
				logFilename = filepath.Join(cfg.LogDir, filename)
			}

			lumberJackLogger := &lumberjack.Logger{
				Filename:   logFilename,
				MaxSize:    cfg.MaxSize,
				MaxBackups: cfg.MaxBackups,
				MaxAge:     cfg.MaxAge,
				Compress:   cfg.Compress,
			}
			writeSyncers = append(writeSyncers, zapcore.AddSync(lumberJackLogger))
		}
	}

	// If no output configured, default to stdout
	if len(writeSyncers) == 0 {
		writeSyncers = append(writeSyncers, zapcore.AddSync(os.Stdout))
	}

	writeSyncer := zapcore.NewMultiWriteSyncer(writeSyncers...)
	return zapcore.NewCore(encoder, writeSyncer, level)
}

// GetLogger get logger instance
func GetLogger() *zap.Logger {
	return Logger
}

// GetSugar get sugar logger instance
func GetSugar() *zap.SugaredLogger {
	return Sugar
}

// Sync sync logs
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}

// Convenience methods
func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}

// Sugar convenience methods
func Debugf(template string, args ...interface{}) {
	Sugar.Debugf(template, args...)
}

func Infof(template string, args ...interface{}) {
	Sugar.Infof(template, args...)
}

func Warnf(template string, args ...interface{}) {
	Sugar.Warnf(template, args...)
}

func Errorf(template string, args ...interface{}) {
	Sugar.Errorf(template, args...)
}

func Fatalf(template string, args ...interface{}) {
	Sugar.Fatalf(template, args...)
}

// WithFields add fields
func WithFields(fields ...zap.Field) *zap.Logger {
	return Logger.With(fields...)
}

// WithField add single field
func WithField(key string, value interface{}) *zap.SugaredLogger {
	return Sugar.With(key, value)
}

// NewRequestLogger create request logger
func NewRequestLogger(requestID string) *zap.SugaredLogger {
	return Sugar.With("request_id", requestID)
}

// Access log methods
func AccessInfo(msg string, fields ...zap.Field) {
	AccessLog.Info(msg, fields...)
}

func AccessWarn(msg string, fields ...zap.Field) {
	AccessLog.Warn(msg, fields...)
}

func AccessError(msg string, fields ...zap.Field) {
	AccessLog.Error(msg, fields...)
}

// Error log methods
func ErrorInfo(msg string, fields ...zap.Field) {
	ErrorLog.Info(msg, fields...)
}

func ErrorWarn(msg string, fields ...zap.Field) {
	ErrorLog.Warn(msg, fields...)
}

func ErrorError(msg string, fields ...zap.Field) {
	ErrorLog.Error(msg, fields...)
}

func ErrorFatal(msg string, fields ...zap.Field) {
	ErrorLog.Fatal(msg, fields...)
}

// Service log methods
func ServiceInfo(msg string, fields ...zap.Field) {
	ServiceLog.Info(msg, fields...)
}

func ServiceWarn(msg string, fields ...zap.Field) {
	ServiceLog.Warn(msg, fields...)
}

func ServiceError(msg string, fields ...zap.Field) {
	ServiceLog.Error(msg, fields...)
}

func ServiceDebug(msg string, fields ...zap.Field) {
	ServiceLog.Debug(msg, fields...)
}
