package database

import (
	"path/filepath"
	"testing"
)

func TestUpDedupeCallRecordingsRemovesHistoricalDuplicates(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "call-recordings-migration.db")
	db := NewSQLite(dbPath)

	createLegacyTable := `
		CREATE TABLE call_recordings (
			id TEXT PRIMARY KEY,
			agent_code INTEGER NOT NULL DEFAULT 0,
			call_status INTEGER NOT NULL DEFAULT 0,
			call_status_name TEXT NOT NULL DEFAULT '',
			call_type INTEGER NOT NULL DEFAULT 0,
			callee_attr TEXT NOT NULL DEFAULT '',
			caller_attr TEXT NOT NULL DEFAULT '',
			create_time INTEGER NOT NULL DEFAULT 0,
			dept_name TEXT NOT NULL DEFAULT '',
			duration INTEGER NOT NULL DEFAULT 0,
			end_time INTEGER NOT NULL DEFAULT 0,
			enterprise_name TEXT NOT NULL DEFAULT '',
			finish_status INTEGER NOT NULL DEFAULT 0,
			finish_status_name TEXT NOT NULL DEFAULT '',
			handle INTEGER NOT NULL DEFAULT 0,
			interface_id TEXT NOT NULL DEFAULT '',
			interface_name TEXT NOT NULL DEFAULT '',
			line_name TEXT NOT NULL DEFAULT '',
			mobile TEXT NOT NULL DEFAULT '',
			mode INTEGER NOT NULL DEFAULT 0,
			move_batch_code TEXT,
			oct_customer_id TEXT,
			phone TEXT NOT NULL DEFAULT '',
			postage REAL NOT NULL DEFAULT 0,
			pre_record_url TEXT,
			real_name TEXT NOT NULL DEFAULT '',
			start_time INTEGER NOT NULL DEFAULT 0,
			status INTEGER NOT NULL DEFAULT 0,
			tel_a TEXT NOT NULL DEFAULT '',
			tel_b TEXT NOT NULL DEFAULT '',
			tel_x TEXT NOT NULL DEFAULT '',
			tenant_code TEXT NOT NULL DEFAULT '',
			update_time INTEGER NOT NULL DEFAULT 0,
			user_id TEXT NOT NULL DEFAULT '',
			work_num TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`
	if err := db.Exec(createLegacyTable).Error; err != nil {
		t.Fatalf("create legacy table failed: %v", err)
	}

	insertDuplicate1 := `
		INSERT INTO call_recordings (
			id, mobile, phone, tel_a, tel_b, start_time, call_type, duration
		) VALUES ('record-1', '13800000000', '13900000000', '13800000000', '13900000000', 1710000000000, 1, 120)
	`
	insertDuplicate2 := `
		INSERT INTO call_recordings (
			id, mobile, phone, tel_a, tel_b, start_time, call_type, duration
		) VALUES ('record-2', '13800000000', '13900000000', '13800000000', '13900000000', 1710000000000, 1, 120)
	`
	if err := db.Exec(insertDuplicate1).Error; err != nil {
		t.Fatalf("insert first duplicate failed: %v", err)
	}
	if err := db.Exec(insertDuplicate2).Error; err != nil {
		t.Fatalf("insert second duplicate failed: %v", err)
	}

	if err := upDedupeCallRecordings(db); err != nil {
		t.Fatalf("upDedupeCallRecordings failed: %v", err)
	}

	var count int64
	if err := db.Table("call_recordings").Count(&count).Error; err != nil {
		t.Fatalf("count call_recordings failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row after dedupe migration, got %d", count)
	}

	var dedupeKey string
	if err := db.Table("call_recordings").Select("dedupe_key").Limit(1).Scan(&dedupeKey).Error; err != nil {
		t.Fatalf("query dedupe_key failed: %v", err)
	}
	if dedupeKey != "1710000000000|13800000000|13900000000|13800000000|13900000000|1|120" {
		t.Fatalf("unexpected dedupe_key: %q", dedupeKey)
	}

	insertAfterMigration := `
		INSERT INTO call_recordings (
			id, mobile, phone, tel_a, tel_b, start_time, call_type, duration, dedupe_key
		) VALUES ('record-3', '13800000000', '13900000000', '13800000000', '13900000000', 1710000000000, 1, 120, '1710000000000|13800000000|13900000000|13800000000|13900000000|1|120')
	`
	if err := db.Exec(insertAfterMigration).Error; err == nil {
		t.Fatalf("expected unique index to reject duplicate dedupe_key insert")
	}
}
