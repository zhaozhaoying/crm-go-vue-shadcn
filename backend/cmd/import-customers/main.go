package main

import (
	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/service"
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"
)

func main() {
	var (
		filePath      = flag.String("file", "", "csv file path to import")
		operatorUser  = flag.Int64("operator-user-id", 1, "operator user id for import ownership/logs")
		batchSize     = flag.Int("batch-size", 1000, "batch size (default 1000, max 5000)")
		dryRun        = flag.Bool("dry-run", false, "validate and deduplicate only, no database writes")
		defaultStatus = flag.String("default-status", "owned", "default customer status when csv status is empty (owned|pool)")
		maxErrors     = flag.Int("max-errors", 200, "max error details returned in output")
		skipMigrate   = flag.Bool("skip-migrate", false, "skip startup migration check")
	)
	flag.Parse()

	if *filePath == "" {
		log.Fatal("file path is required, usage: go run ./cmd/import-customers --file ./customers.csv")
	}
	if *operatorUser <= 0 {
		log.Fatal("operator-user-id must be greater than 0")
	}

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

	if !*skipMigrate {
		if err := database.RunMigrations(db); err != nil {
			log.Fatalf("failed to run migrations before import: %v", err)
		}
	}

	file, err := os.Open(*filePath)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	importService := service.NewCustomerImportService(db)
	start := time.Now()
	result, err := importService.ImportCSV(context.Background(), file, service.CustomerCSVImportInput{
		OperatorUserID: *operatorUser,
		BatchSize:      *batchSize,
		DryRun:         *dryRun,
		DefaultStatus:  *defaultStatus,
		MaxErrors:      *maxErrors,
	})
	if err != nil {
		log.Fatalf("import failed: %v", err)
	}

	output := struct {
		DurationMS int64 `json:"durationMs"`
		Result     any   `json:"result"`
	}{
		DurationMS: time.Since(start).Milliseconds(),
		Result:     result,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		log.Fatalf("failed to print import result: %v", err)
	}
}
