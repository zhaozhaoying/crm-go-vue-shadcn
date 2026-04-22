package main

import (
	"backend/internal/config"
	"backend/internal/database"
	"fmt"
	"log"
	"os"
)

func main() {
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

	action := "up"
	if len(os.Args) > 1 {
		action = os.Args[1]
	}

	switch action {
	case "up":
		if err := database.RunMigrations(db); err != nil {
			log.Fatalf("migrate up failed: %v", err)
		}
		fmt.Println("migrations applied successfully")
	case "status":
		status, err := database.ListMigrationStatus(db)
		if err != nil {
			log.Fatalf("migrate status failed: %v", err)
		}
		for _, item := range status {
			if item.Applied && item.AppliedAt != nil {
				fmt.Printf("[APPLIED] %d_%s at %s\n", item.Version, item.Name, item.AppliedAt.Format("2006-01-02 15:04:05"))
				continue
			}
			fmt.Printf("[PENDING] %d_%s\n", item.Version, item.Name)
		}
	default:
		log.Fatalf("unknown action %q, supported: up | status", action)
	}
}
