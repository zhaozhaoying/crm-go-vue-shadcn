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
		return NewMySQL(cfg.EffectiveMySQLDSN(), gormCfg)
	default:
		return NewSQLite(cfg.DBPath, gormCfg)
	}
}
