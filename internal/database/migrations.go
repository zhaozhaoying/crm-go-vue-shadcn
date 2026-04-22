package database

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

const schemaMigrationsTable = "schema_migrations"

type Migration struct {
	Version int64
	Name    string
	Up      func(tx *gorm.DB) error
}

type MigrationStatus struct {
	Version   int64      `json:"version"`
	Name      string     `json:"name"`
	Applied   bool       `json:"applied"`
	AppliedAt *time.Time `json:"appliedAt,omitempty"`
}

type schemaMigrationRow struct {
	Version   int64     `gorm:"column:version"`
	Name      string    `gorm:"column:name"`
	AppliedAt time.Time `gorm:"column:applied_at"`
}

func RunMigrations(db *gorm.DB) error {
	if err := ensureMigrationTable(db); err != nil {
		return err
	}

	migrations := getMigrationsByDialect(db)
	if err := validateMigrations(migrations); err != nil {
		return err
	}

	applied, err := loadAppliedMigrations(db)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if applied[migration.Version] {
			continue
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			if err := migration.Up(tx); err != nil {
				return err
			}
			return tx.Table(schemaMigrationsTable).Create(map[string]interface{}{
				"version": migration.Version,
				"name":    migration.Name,
			}).Error
		})
		if err != nil {
			return fmt.Errorf("apply migration %d_%s failed: %w", migration.Version, migration.Name, err)
		}
	}

	return nil
}

func ListMigrationStatus(db *gorm.DB) ([]MigrationStatus, error) {
	if err := ensureMigrationTable(db); err != nil {
		return nil, err
	}

	migrations := getMigrationsByDialect(db)
	if err := validateMigrations(migrations); err != nil {
		return nil, err
	}

	var rows []schemaMigrationRow
	if err := db.Table(schemaMigrationsTable).
		Select("version", "name", "applied_at").
		Order("version ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	appliedByVersion := make(map[int64]schemaMigrationRow, len(rows))
	for _, row := range rows {
		appliedByVersion[row.Version] = row
	}

	status := make([]MigrationStatus, 0, len(migrations))
	for _, migration := range migrations {
		item := MigrationStatus{
			Version: migration.Version,
			Name:    migration.Name,
		}
		if row, ok := appliedByVersion[migration.Version]; ok {
			item.Applied = true
			appliedAt := row.AppliedAt
			item.AppliedAt = &appliedAt
		}
		status = append(status, item)
	}

	return status, nil
}

func ensureMigrationTable(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
}

func loadAppliedMigrations(db *gorm.DB) (map[int64]bool, error) {
	var rows []schemaMigrationRow
	if err := db.Table(schemaMigrationsTable).Select("version").Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make(map[int64]bool, len(rows))
	for _, row := range rows {
		result[row.Version] = true
	}
	return result, nil
}

func validateMigrations(migrations []Migration) error {
	versions := make([]int64, 0, len(migrations))
	seen := make(map[int64]struct{}, len(migrations))
	for _, migration := range migrations {
		if migration.Version <= 0 {
			return fmt.Errorf("invalid migration version: %d", migration.Version)
		}
		if migration.Name == "" {
			return fmt.Errorf("migration %d has empty name", migration.Version)
		}
		if migration.Up == nil {
			return fmt.Errorf("migration %d has nil Up", migration.Version)
		}
		if _, ok := seen[migration.Version]; ok {
			return fmt.Errorf("duplicate migration version: %d", migration.Version)
		}
		seen[migration.Version] = struct{}{}
		versions = append(versions, migration.Version)
	}

	if !sort.SliceIsSorted(versions, func(i, j int) bool { return versions[i] < versions[j] }) {
		return fmt.Errorf("migrations must be ordered by version ASC")
	}

	return nil
}

func execStatements(tx *gorm.DB, statements []string) error {
	for _, stmt := range statements {
		if err := tx.Exec(stmt).Error; err != nil {
			return err
		}
	}
	return nil
}

type pragmaColumnInfo struct {
	Name string `gorm:"column:name"`
}

func addColumnIfNotExists(tx *gorm.DB, table, column, definition string) error {
	exists, err := hasColumn(tx, table, column)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	alter := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, definition)
	return tx.Exec(alter).Error
}

func addIndexIfNotExists(tx *gorm.DB, table, indexName, columns string, unique bool) error {
	exists, err := hasIndex(tx, table, indexName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	keyword := "INDEX"
	if unique {
		keyword = "UNIQUE INDEX"
	}
	stmt := fmt.Sprintf("CREATE %s %s ON %s(%s)", keyword, indexName, table, columns)
	return tx.Exec(stmt).Error
}

func renameTableIfExists(tx *gorm.DB, from, to string) error {
	if strings.EqualFold(strings.TrimSpace(from), strings.TrimSpace(to)) {
		return nil
	}
	if tx.Migrator().HasTable(to) {
		return nil
	}
	if !tx.Migrator().HasTable(from) {
		return nil
	}
	return tx.Migrator().RenameTable(from, to)
}

func getMigrationsByDialect(db *gorm.DB) []Migration {
	switch strings.ToLower(strings.TrimSpace(db.Dialector.Name())) {
	case "mysql":
		return getMySQLMigrations()
	default:
		return getSQLiteMigrations()
	}
}

func hasColumn(tx *gorm.DB, table, column string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(tx.Dialector.Name())) {
	case "mysql":
		type columnRow struct {
			ColumnName string `gorm:"column:COLUMN_NAME"`
		}
		var row columnRow
		err := tx.Raw(`
			SELECT COLUMN_NAME
			FROM information_schema.COLUMNS
			WHERE TABLE_SCHEMA = DATABASE()
				AND TABLE_NAME = ?
				AND COLUMN_NAME = ?
			LIMIT 1
		`, table, column).Scan(&row).Error
		if err != nil {
			return false, err
		}
		return strings.EqualFold(row.ColumnName, column), nil
	default:
		var cols []pragmaColumnInfo
		query := fmt.Sprintf("PRAGMA table_info(%s)", table)
		if err := tx.Raw(query).Scan(&cols).Error; err != nil {
			return false, err
		}
		for _, col := range cols {
			if strings.EqualFold(col.Name, column) {
				return true, nil
			}
		}
		return false, nil
	}
}

func hasIndex(tx *gorm.DB, table, indexName string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(tx.Dialector.Name())) {
	case "mysql":
		type indexRow struct {
			IndexName string `gorm:"column:INDEX_NAME"`
		}
		var row indexRow
		err := tx.Raw(`
			SELECT INDEX_NAME
			FROM information_schema.STATISTICS
			WHERE TABLE_SCHEMA = DATABASE()
				AND TABLE_NAME = ?
				AND INDEX_NAME = ?
			LIMIT 1
		`, table, indexName).Scan(&row).Error
		if err != nil {
			return false, err
		}
		return strings.EqualFold(row.IndexName, indexName), nil
	default:
		type pragmaIndexInfo struct {
			Name string `gorm:"column:name"`
		}
		var rows []pragmaIndexInfo
		query := fmt.Sprintf("PRAGMA index_list(%s)", table)
		if err := tx.Raw(query).Scan(&rows).Error; err != nil {
			return false, err
		}
		for _, row := range rows {
			if strings.EqualFold(row.Name, indexName) {
				return true, nil
			}
		}
		return false, nil
	}
}
