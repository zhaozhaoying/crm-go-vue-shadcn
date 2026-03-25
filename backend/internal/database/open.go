package database

import (
	"backend/internal/config"
	"strings"

	"gorm.io/gorm"
)

func Open(cfg config.Config) *gorm.DB {
	gormCfg := NewGormConfig(cfg)
	switch strings.ToLower(strings.TrimSpace(cfg.DBDriver)) {
	case "mysql":
		return NewMySQLWithLocation(cfg.EffectiveMySQLDSN(), cfg.ScheduleLocation(), gormCfg)
	default:
		return NewSQLite(cfg.DBPath, gormCfg)
	}
}
