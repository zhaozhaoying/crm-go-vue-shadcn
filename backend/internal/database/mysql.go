package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewMySQL(dsn string, gormCfgs ...*gorm.Config) *gorm.DB {
	return NewMySQLWithLocation(dsn, nil, gormCfgs...)
}

func NewMySQLWithLocation(dsn string, location *time.Location, gormCfgs ...*gorm.Config) *gorm.DB {
	gormCfg := defaultGormConfig()
	if len(gormCfgs) > 0 && gormCfgs[0] != nil {
		gormCfg = gormCfgs[0]
	}

	db, err := gorm.Open(mysql.Open(dsn), gormCfg)
	if err != nil {
		log.Fatalf("failed to open mysql via gorm: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get generic sql DB from gorm: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("failed to ping mysql via gorm: %v", err)
	}
	applyMySQLSessionTimeZone(db, location)

	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db
}

func applyMySQLSessionTimeZone(db *gorm.DB, location *time.Location) {
	if db == nil || location == nil {
		return
	}

	timeZoneValue := mysqlSessionTimeZoneValue(location)
	if timeZoneValue == "" {
		return
	}

	if err := db.Exec(fmt.Sprintf("SET time_zone = '%s'", timeZoneValue)).Error; err != nil {
		log.Printf("warn: failed to set mysql session time_zone to %s: %v", timeZoneValue, err)
	}
}

func mysqlSessionTimeZoneValue(location *time.Location) string {
	if location == nil {
		return ""
	}

	_, offsetSeconds := time.Now().In(location).Zone()
	sign := "+"
	if offsetSeconds < 0 {
		sign = "-"
		offsetSeconds = -offsetSeconds
	}

	hours := offsetSeconds / 3600
	minutes := (offsetSeconds % 3600) / 60
	return fmt.Sprintf("%s%02d:%02d", sign, hours, minutes)
}
