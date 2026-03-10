package main

import (
	"backend/internal/config"
	"backend/internal/database"
	"flag"
	"fmt"
	"log"
)

func main() {
	var (
		adminUsername      = flag.String("admin-username", "admin", "admin username")
		adminPassword      = flag.String("admin-password", "admin123", "admin password")
		adminNickname      = flag.String("admin-nickname", "管理员", "admin nickname")
		resetAdminPassword = flag.Bool("reset-admin-password", false, "reset admin password if user already exists")
		skipAdmin          = flag.Bool("skip-admin", false, "skip admin account bootstrap")
		skipDictionaries   = flag.Bool("skip-dictionaries", false, "skip dictionary bootstrap")
	)
	flag.Parse()

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid config: %v", err)
	}
	db := database.Open(cfg)
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql db from gorm: %v", err)
	}
	defer sqlDB.Close()

	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("failed to run migrations before bootstrap: %v", err)
	}

	report, err := database.RunBootstrap(db, database.BootstrapOptions{
		AdminUsername:      *adminUsername,
		AdminPassword:      *adminPassword,
		AdminNickname:      *adminNickname,
		ResetAdminPassword: *resetAdminPassword,
		SkipAdmin:          *skipAdmin,
		SkipDictionaries:   *skipDictionaries,
	})
	if err != nil {
		log.Fatalf("bootstrap failed: %v", err)
	}

	fmt.Printf("roles upserted: %d\n", report.RolesUpserted)
	fmt.Printf("customer levels inserted: %d\n", report.CustomerLevelsInserted)
	fmt.Printf("customer sources inserted: %d\n", report.CustomerSourcesInserted)
	fmt.Printf("follow methods inserted: %d\n", report.FollowMethodsInserted)
	fmt.Printf("admin created: %t\n", report.AdminCreated)
	fmt.Printf("admin password reset: %t\n", report.AdminPasswordReset)
}
