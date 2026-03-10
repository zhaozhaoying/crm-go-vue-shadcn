package database

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewSQLite(dsn string, gormCfgs ...*gorm.Config) *gorm.DB {
	gormCfg := defaultGormConfig()
	if len(gormCfgs) > 0 && gormCfgs[0] != nil {
		gormCfg = gormCfgs[0]
	}

	db, err := gorm.Open(sqlite.Open(dsn), gormCfg)
	if err != nil {
		log.Fatalf("failed to open sqlite via gorm: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get generic sql DB from gorm: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("failed to ping sqlite via gorm: %v", err)
	}
	// sqlite 在并发写场景下多连接容易出现 database is locked；单连接更稳定
	sqlDB.SetMaxOpenConns(1)

	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		log.Fatalf("failed to enable sqlite foreign keys: %v", err)
	}
	return db
}
