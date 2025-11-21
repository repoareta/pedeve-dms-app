package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

// InitLogger menginisialisasi zap logger
func InitLogger() {
	var config zap.Config
	var err error

	// Tentukan environment
	env := os.Getenv("ENV")
	if env == "production" {
		// Production: JSON output, info level
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	} else {
		// Development: Console output, debug level
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		config.EncoderConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
	}

	// Build logger
	Log, err = config.Build(
		zap.AddCaller(),           // Tambahkan caller info
		zap.AddStacktrace(zapcore.ErrorLevel), // Stack trace untuk error level
	)
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
}

// GetLogger mengembalikan instance logger
func GetLogger() *zap.Logger {
	if Log == nil {
		InitLogger()
	}
	return Log
}

// Sync flushes any buffered log entries
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

