package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

// InitLogger initializes the logger with the specified environment
func InitLogger(env string) {
	var config zap.Config

	if env == "production" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.OutputPaths = []string{"stdout", "/var/log/facedrop/app.log"}
		config.ErrorOutputPaths = []string{"stderr", "/var/log/facedrop/error.log"}
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.OutputPaths = []string{"stdout"}
		config.ErrorOutputPaths = []string{"stderr"}
	}

	// Add custom fields to all logs
	config.InitialFields = map[string]interface{}{
		"service": "facedrop",
		"env":     env,
	}

	var err error
	log, err = config.Build()
	if err != nil {
		panic(err)
	}
}

// GetLogger returns the logger instance
func GetLogger() *zap.Logger {
	if log == nil {
		InitLogger("development")
	}
	return log
}

// Info logs an info message with fields
func Info(msg string, fields ...zapcore.Field) {
	GetLogger().Info(msg, fields...)
}

// Error logs an error message with fields
func Error(err error) zapcore.Field {
	return zap.Error(err)
}

// Debug logs a debug message with fields
func Debug(msg string, fields ...zapcore.Field) {
	GetLogger().Debug(msg, fields...)
}

// Warn logs a warning message with fields
func Warn(msg string, fields ...zapcore.Field) {
	GetLogger().Warn(msg, fields...)
}

// Fatal logs a fatal message with fields and exits
func Fatal(msg string, fields ...zapcore.Field) {
	GetLogger().Fatal(msg, fields...)
}

// String creates a string field
func String(key, value string) zapcore.Field {
	return zap.String(key, value)
}

// Int creates an integer field
func Int(key string, value int) zapcore.Field {
	return zap.Int(key, value)
}

// WithFields creates a new logger with additional fields
func WithFields(fields ...zapcore.Field) *zap.Logger {
	return GetLogger().With(fields...)
}

// HTTPLogger is a middleware that logs HTTP requests
func HTTPLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Stop timer
		duration := time.Since(start)

		// Log request
		GetLogger().Info("HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)
	}
} 