package main

import (
	"backend/internal/config"
	"backend/internal/database"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
)

var migrationTableOrder = []string{
	"roles",
	"users",
	"system_settings",
	"customer_levels",
	"customer_sources",
	"follow_methods",
	"customers",
	"customer_phones",
	"customer_owner_logs",
	"customer_follow_logs",
	"customer_status_logs",
	"operation_follow_records",
	"sales_follow_records",
	"contracts",
	"refresh_tokens",
	"token_blacklist",
	"resource_pool",
	"external_company_search_task",
	"external_company",
	"external_company_search_result",
	"external_company_search_event",
}

type tableReport struct {
	Table string
	Rows  int64
}

func main() {
	var (
		sqlitePath = flag.String("sqlite-path", "", "sqlite database path (default from DB_PATH)")
		mysqlDSN   = flag.String("mysql-dsn", "", "mysql dsn (default from MYSQL_DSN or MYSQL_* env)")
		dryRun     = flag.Bool("dry-run", false, "read source rows without writing target")
	)
	flag.Parse()

	cfg := config.Load()

	srcPath := strings.TrimSpace(*sqlitePath)
	if srcPath == "" {
		srcPath = strings.TrimSpace(cfg.DBPath)
	}
	if srcPath == "" {
		log.Fatal("sqlite path is empty, please set --sqlite-path or DB_PATH")
	}

	targetDSN := strings.TrimSpace(*mysqlDSN)
	if targetDSN == "" {
		targetDSN = strings.TrimSpace(cfg.EffectiveMySQLDSN())
	}
	if targetDSN == "" {
		log.Fatal("mysql dsn is empty, please set --mysql-dsn or MYSQL_DSN / MYSQL_*")
	}

	sourceDB := database.NewSQLite(srcPath)
	sourceSQLDB, err := sourceDB.DB()
	if err != nil {
		log.Fatalf("failed to get sqlite sql db: %v", err)
	}
	defer sourceSQLDB.Close()

	targetDB := database.NewMySQL(targetDSN)
	targetSQLDB, err := targetDB.DB()
	if err != nil {
		log.Fatalf("failed to get mysql sql db: %v", err)
	}
	defer targetSQLDB.Close()

	totalRows := int64(0)
	reports := make([]tableReport, 0, len(migrationTableOrder))

	for _, table := range migrationTableOrder {
		exists := sourceDB.Migrator().HasTable(table)
		if !exists {
			log.Printf("skip table %s: not found in sqlite", table)
			continue
		}

		count, err := migrateTable(sourceDB, targetDB, table, *dryRun)
		if err != nil {
			log.Fatalf("migrate table %s failed: %v", table, err)
		}
		reports = append(reports, tableReport{
			Table: table,
			Rows:  count,
		})
		totalRows += count
	}

	log.Printf("sqlite -> mysql migration finished, dryRun=%t, tables=%d, rows=%d", *dryRun, len(reports), totalRows)
	for _, item := range reports {
		log.Printf("table=%s rows=%d", item.Table, item.Rows)
	}
}

func migrateTable(sourceDB *gorm.DB, targetDB *gorm.DB, table string, dryRun bool) (int64, error) {
	rows, err := sourceDB.Table(table).Rows()
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return 0, err
	}
	if len(columns) == 0 {
		return 0, nil
	}

	insertSQL := buildUpsertSQL(table, columns)

	tx := targetDB.Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	defer tx.Rollback()

	var migrated int64
	for rows.Next() {
		values := make([]interface{}, len(columns))
		scanArgs := make([]interface{}, len(columns))
		for i := range values {
			scanArgs[i] = &values[i]
		}
		if err := rows.Scan(scanArgs...); err != nil {
			return migrated, err
		}

		for idx := range values {
			values[idx] = normalizeSQLiteValue(values[idx])
		}

		if !dryRun {
			if err := tx.Exec(insertSQL, values...).Error; err != nil {
				return migrated, fmt.Errorf("exec upsert failed: table=%s err=%w", table, err)
			}
		}
		migrated++
	}
	if err := rows.Err(); err != nil {
		return migrated, err
	}

	if dryRun {
		return migrated, tx.Rollback().Error
	}

	if hasColumn(columns, "id") {
		if err := resetAutoIncrement(tx, table); err != nil {
			return migrated, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return migrated, err
	}
	return migrated, nil
}

func buildUpsertSQL(table string, columns []string) string {
	quotedCols := make([]string, 0, len(columns))
	placeholders := make([]string, 0, len(columns))
	updates := make([]string, 0, len(columns))

	for _, col := range columns {
		quoted := quoteIdent(col)
		quotedCols = append(quotedCols, quoted)
		placeholders = append(placeholders, "?")
		if strings.EqualFold(col, "id") {
			continue
		}
		updates = append(updates, fmt.Sprintf("%s=VALUES(%s)", quoted, quoted))
	}

	insertPrefix := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		quoteIdent(table),
		strings.Join(quotedCols, ","),
		strings.Join(placeholders, ","),
	)
	if len(updates) == 0 {
		return insertPrefix
	}
	return insertPrefix + " ON DUPLICATE KEY UPDATE " + strings.Join(updates, ",")
}

func normalizeSQLiteValue(v interface{}) interface{} {
	switch value := v.(type) {
	case []byte:
		return string(value)
	case time.Time:
		return value.UTC()
	default:
		return value
	}
}

func hasColumn(columns []string, target string) bool {
	for _, col := range columns {
		if strings.EqualFold(col, target) {
			return true
		}
	}
	return false
}

func resetAutoIncrement(tx *gorm.DB, table string) error {
	var nextID int64
	if err := tx.Raw(fmt.Sprintf("SELECT COALESCE(MAX(id), 0) + 1 AS next_id FROM %s", quoteIdent(table))).
		Scan(&nextID).Error; err != nil {
		return err
	}
	if nextID < 1 {
		nextID = 1
	}
	stmt := fmt.Sprintf("ALTER TABLE %s AUTO_INCREMENT = %d", quoteIdent(table), nextID)
	return tx.Exec(stmt).Error
}

func quoteIdent(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}
