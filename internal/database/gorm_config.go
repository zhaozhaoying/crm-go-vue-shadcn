package database

import (
	"backend/internal/config"
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func NewGormConfig(cfg config.Config) *gorm.Config {
	return &gorm.Config{
		Logger: newGormLogger(
			strings.ToLower(strings.TrimSpace(cfg.GormLogLevel)),
			cfg.GormSlowThresholdMS,
		),
	}
}

func defaultGormConfig() *gorm.Config {
	return &gorm.Config{
		Logger: newGormLogger("error", 200),
	}
}

func newGormLogger(level string, slowThresholdMS int) gormlogger.Interface {
	if slowThresholdMS <= 0 {
		slowThresholdMS = 200
	}
	return gormlogger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		gormlogger.Config{
			SlowThreshold:             time.Duration(slowThresholdMS) * time.Millisecond,
			LogLevel:                  parseGormLogLevel(level),
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
}

func parseGormLogLevel(level string) gormlogger.LogLevel {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "silent":
		return gormlogger.Silent
	case "warn":
		return gormlogger.Warn
	case "info":
		return gormlogger.Info
	case "", "error":
		return gormlogger.Error
	default:
		return gormlogger.Error
	}
}
