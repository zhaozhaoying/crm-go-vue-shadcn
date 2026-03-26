package database

import "gorm.io/gorm"

func getSQLiteMigrations() []Migration {
	return []Migration{
		{
			Version: 2026030301,
			Name:    "create_base_schema",
			Up:      upCreateBaseSchema,
		},
		{
			Version: 2026030302,
			Name:    "add_missing_columns_and_indexes",
			Up:      upAddMissingColumnsAndIndexes,
		},
		{
			Version: 2026030401,
			Name:    "create_contracts_table",
			Up:      upCreateContractsTable,
		},
		{
			Version: 2026030402,
			Name:    "seed_contract_number_prefix_setting",
			Up:      upSeedContractNumberPrefixSetting,
		},
		{
			Version: 2026030403,
			Name:    "create_auth_token_tables",
			Up:      upCreateAuthTokenTables,
		},
		{
			Version: 2026030404,
			Name:    "create_resource_pool_table",
			Up:      upCreateResourcePoolTable,
		},
		{
			Version: 2026030405,
			Name:    "add_resource_pool_conversion_columns",
			Up:      upAddResourcePoolConversionColumns,
		},
		{
			Version: 2026030501,
			Name:    "add_contract_remark_column",
			Up:      upAddContractRemarkColumn,
		},
		{
			Version: 2026030601,
			Name:    "create_external_company_search_tables",
			Up:      upCreateExternalCompanySearchTables,
		},
		{
			Version: 2026030602,
			Name:    "rename_spider_crawl_tables_to_external_company_search_tables",
			Up:      upRenameSpiderCrawlTablesToExternalCompanySearchTables,
		},
		{
			Version: 2026030604,
			Name:    "create_activity_logs_table",
			Up:      upCreateActivityLogsTable,
		},
		{
			Version: 2026030701,
			Name:    "add_contract_audit_columns",
			Up:      upAddContractAuditColumns,
		},
		{
			Version: 2026031101,
			Name:    "add_user_sales_type",
			Up:      upAddUserSalesType,
		},
		{
			Version: 2026031701,
			Name:    "seed_customer_rule_settings",
			Up:      upSeedCustomerRuleSettings,
		},
		{
			Version: 2026031801,
			Name:    "add_customer_owner_log_reason_and_content",
			Up:      upAddCustomerOwnerLogReasonAndContent,
		},
		{
			Version: 2026031802,
			Name:    "add_customer_inside_sales_fields",
			Up:      upAddCustomerInsideSalesFields,
		},
		{
			Version: 2026031901,
			Name:    "add_customer_owner_log_claim_block_fields",
			Up:      upAddCustomerOwnerLogClaimBlockFields,
		},
		{
			Version: 2026031902,
			Name:    "create_daily_user_call_stats_and_add_users_hanghang_crm_mobile",
			Up:      upCreateDailyUserCallStatsAndAddUsersHanghangCrmMobile,
		},
		{
			Version: 2026032002,
			Name:    "create_customer_visits_table",
			Up:      upCreateCustomerVisitsTable,
		},
		{
			Version: 2026032003,
			Name:    "create_sales_daily_scores",
			Up:      upCreateSalesDailyScores,
		},
		{
			Version: 2026032004,
			Name:    "alter_customer_visits_region_columns_to_text",
			Up:      upAlterCustomerVisitsRegionColumnsToText,
		},
		{
			Version: 2026032005,
			Name:    "add_customer_visits_check_in_ip",
			Up:      upAddCustomerVisitsCheckInIP,
		},
		{
			Version: 2026032301,
			Name:    "create_call_recordings",
			Up:      upCreateCallRecordings,
		},
		{
			Version: 2026032302,
			Name:    "create_user_hanghang_crm_mobiles",
			Up:      upCreateUserHanghangCRMMobiles,
		},
		{
			Version: 2026032401,
			Name:    "dedupe_call_recordings",
			Up:      upDedupeCallRecordings,
		},
		{
			Version: 2026032501,
			Name:    "add_sales_daily_score_reached_at",
			Up:      upAddSalesDailyScoreReachedAt,
		},
		{
			Version: 2026032601,
			Name:    "add_customer_assign_time_and_sales_assign_drop_setting",
			Up:      upAddCustomerAssignTimeAndSalesAssignDropSetting,
		},
	}
}

func upCreateCustomerVisitsTable(tx *gorm.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS customer_visits (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			operator_user_id INTEGER NOT NULL,
			customer_name TEXT NOT NULL,
			check_in_lat REAL NOT NULL DEFAULT 0,
			check_in_lng REAL NOT NULL DEFAULT 0,
			province TEXT NOT NULL DEFAULT '',
			city TEXT NOT NULL DEFAULT '',
			area TEXT NOT NULL DEFAULT '',
			detail_address TEXT NOT NULL DEFAULT '',
			images TEXT NOT NULL DEFAULT '[]',
			visit_purpose TEXT NOT NULL DEFAULT '',
			remark TEXT NOT NULL DEFAULT '',
			visit_date TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (operator_user_id) REFERENCES users(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_customer_visits_operator_user_id ON customer_visits(operator_user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_customer_visits_visit_date ON customer_visits(visit_date)`,
		`CREATE INDEX IF NOT EXISTS idx_customer_visits_created_at ON customer_visits(created_at)`,
	}
	return execStatements(tx, stmts)
}

func upAlterCustomerVisitsRegionColumnsToText(tx *gorm.DB) error {
	if !tx.Migrator().HasTable("customer_visits") {
		return nil
	}

	stmts := []string{
		`ALTER TABLE customer_visits RENAME TO customer_visits_old`,
		`CREATE TABLE customer_visits (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			operator_user_id INTEGER NOT NULL,
			customer_name TEXT NOT NULL,
			check_in_lat REAL NOT NULL DEFAULT 0,
			check_in_lng REAL NOT NULL DEFAULT 0,
			province TEXT NOT NULL DEFAULT '',
			city TEXT NOT NULL DEFAULT '',
			area TEXT NOT NULL DEFAULT '',
			detail_address TEXT NOT NULL DEFAULT '',
			images TEXT NOT NULL DEFAULT '[]',
			visit_purpose TEXT NOT NULL DEFAULT '',
			remark TEXT NOT NULL DEFAULT '',
			visit_date TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (operator_user_id) REFERENCES users(id)
		)`,
		`INSERT INTO customer_visits (
			id, operator_user_id, customer_name, check_in_lat, check_in_lng,
			province, city, area, detail_address, images, visit_purpose, remark,
			visit_date, created_at, updated_at
		)
		SELECT
			id, operator_user_id, customer_name, check_in_lat, check_in_lng,
			CASE WHEN COALESCE(CAST(province AS TEXT), '') IN ('', '0') THEN '' ELSE CAST(province AS TEXT) END,
			CASE WHEN COALESCE(CAST(city AS TEXT), '') IN ('', '0') THEN '' ELSE CAST(city AS TEXT) END,
			CASE WHEN COALESCE(CAST(area AS TEXT), '') IN ('', '0') THEN '' ELSE CAST(area AS TEXT) END,
			COALESCE(detail_address, ''),
			COALESCE(images, '[]'),
			COALESCE(visit_purpose, ''),
			COALESCE(remark, ''),
			visit_date,
			created_at,
			updated_at
		FROM customer_visits_old`,
		`DROP TABLE customer_visits_old`,
		`CREATE INDEX IF NOT EXISTS idx_customer_visits_operator_user_id ON customer_visits(operator_user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_customer_visits_visit_date ON customer_visits(visit_date)`,
		`CREATE INDEX IF NOT EXISTS idx_customer_visits_created_at ON customer_visits(created_at)`,
	}
	return execStatements(tx, stmts)
}

func upCreateUserHanghangCRMMobiles(tx *gorm.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS user_hanghang_crm_mobiles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			mobile TEXT NOT NULL DEFAULT '',
			is_primary INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uk_user_hanghang_crm_mobiles_mobile ON user_hanghang_crm_mobiles(mobile)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uk_user_hanghang_crm_mobiles_user_mobile ON user_hanghang_crm_mobiles(user_id, mobile)`,
		`CREATE INDEX IF NOT EXISTS idx_user_hanghang_crm_mobiles_user_id ON user_hanghang_crm_mobiles(user_id)`,
		`INSERT OR IGNORE INTO user_hanghang_crm_mobiles (user_id, mobile, is_primary, created_at, updated_at)
		 SELECT id, hanghang_crm_mobile, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		 FROM users
		 WHERE COALESCE(hanghang_crm_mobile, '') <> ''`,
	}
	return execStatements(tx, stmts)
}

func upAddCustomerVisitsCheckInIP(tx *gorm.DB) error {
	if !tx.Migrator().HasTable("customer_visits") {
		return nil
	}

	if err := addColumnIfNotExists(tx, "customer_visits", "check_in_ip", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	return addIndexIfNotExists(
		tx,
		"customer_visits",
		"idx_customer_visits_user_date_ip_customer",
		"operator_user_id, visit_date, check_in_ip, customer_name",
		false,
	)
}

func upAddCustomerAssignTimeAndSalesAssignDropSetting(tx *gorm.DB) error {
	if err := addColumnIfNotExists(tx, "customers", "assign_time", "INTEGER"); err != nil {
		return err
	}
	if err := addIndexIfNotExists(tx, "customers", "idx_customers_assign_time", "assign_time", false); err != nil {
		return err
	}

	stmts := []string{
		`UPDATE customers
		SET assign_time = COALESCE(
			CASE
				WHEN converted_at IS NOT NULL THEN CAST(strftime('%s', converted_at) AS INTEGER)
				ELSE NULL
			END,
			collect_time
		)
		WHERE COALESCE(assign_time, 0) = 0
			AND inside_sales_user_id IS NOT NULL
			AND owner_user_id IS NOT NULL
			AND owner_user_id <> inside_sales_user_id
			AND status <> 'pool'`,
		`INSERT OR IGNORE INTO system_settings(key,value,description) VALUES ('sales_assign_deal_drop_days','30','电销分配给销售后多少天未签单自动掉库')`,
	}
	return execStatements(tx, stmts)
}

func upCreateCallRecordings(tx *gorm.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS call_recordings (
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
			dedupe_key TEXT NOT NULL DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_call_recordings_mobile ON call_recordings(mobile)`,
		`CREATE INDEX IF NOT EXISTS idx_call_recordings_phone ON call_recordings(phone)`,
		`CREATE INDEX IF NOT EXISTS idx_call_recordings_start_time ON call_recordings(start_time)`,
		`CREATE INDEX IF NOT EXISTS idx_call_recordings_create_time ON call_recordings(create_time)`,
		`CREATE INDEX IF NOT EXISTS idx_call_recordings_user_id ON call_recordings(user_id)`,
	}
	return execStatements(tx, stmts)
}

func upDedupeCallRecordings(tx *gorm.DB) error {
	if !tx.Migrator().HasTable("call_recordings") {
		return nil
	}

	if err := addColumnIfNotExists(tx, "call_recordings", "dedupe_key", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}

	stmts := []string{
		`UPDATE call_recordings
		SET dedupe_key =
			COALESCE(CAST(start_time AS TEXT), '') || '|' ||
			COALESCE(mobile, '') || '|' ||
			COALESCE(phone, '') || '|' ||
			COALESCE(tel_a, '') || '|' ||
			COALESCE(tel_b, '') || '|' ||
			COALESCE(CAST(call_type AS TEXT), '') || '|' ||
			COALESCE(CAST(duration AS TEXT), '')`,
		`DELETE FROM call_recordings
		WHERE id IN (
			SELECT older.id
			FROM call_recordings AS older
			INNER JOIN call_recordings AS newer
				ON older.dedupe_key = newer.dedupe_key
				AND older.id < newer.id
		)`,
	}
	if err := execStatements(tx, stmts); err != nil {
		return err
	}

	return addIndexIfNotExists(tx, "call_recordings", "uk_call_recordings_dedupe_key", "dedupe_key", true)
}

func upCreateBaseSchema(tx *gorm.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS roles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			label TEXT NOT NULL DEFAULT '',
			sort INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			salt TEXT NOT NULL DEFAULT '',
			nickname TEXT NOT NULL DEFAULT '',
			email TEXT NOT NULL DEFAULT '',
			mobile TEXT NOT NULL DEFAULT '',
			avatar TEXT NOT NULL DEFAULT '',
			role_id INTEGER NOT NULL DEFAULT 0,
			parent_id INTEGER,
			sales_type TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'enabled',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (role_id) REFERENCES roles(id)
		)`,
		`CREATE TABLE IF NOT EXISTS customers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			legal_name TEXT NOT NULL DEFAULT '',
			contact_name TEXT NOT NULL DEFAULT '',
			weixin TEXT NOT NULL DEFAULT '',
			email TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'pool',
			deal_status TEXT NOT NULL DEFAULT 'undone',
			owner_user_id INTEGER,
			inside_sales_user_id INTEGER,
			converted_at DATETIME,
			customer_level_id INTEGER NOT NULL DEFAULT 0,
			customer_source_id INTEGER NOT NULL DEFAULT 0,
			province INTEGER NOT NULL DEFAULT 0,
			city INTEGER NOT NULL DEFAULT 0,
			area INTEGER NOT NULL DEFAULT 0,
			detail_address TEXT NOT NULL DEFAULT '',
			lng REAL NOT NULL DEFAULT 0,
			lat REAL NOT NULL DEFAULT 0,
			next_time INTEGER NOT NULL DEFAULT 0,
			follow_time INTEGER,
			remark TEXT,
			deal_time INTEGER,
			customer_status INTEGER NOT NULL DEFAULT 0,
			collect_time INTEGER,
			drop_time INTEGER,
			drop_user_id INTEGER,
			create_user_id INTEGER NOT NULL DEFAULT 0,
			operate_user_id INTEGER NOT NULL DEFAULT 0,
			is_lock INTEGER NOT NULL DEFAULT 0,
			delete_time INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (owner_user_id) REFERENCES users(id),
			FOREIGN KEY (inside_sales_user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS customer_owner_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			customer_id INTEGER NOT NULL,
			from_owner_user_id INTEGER,
			to_owner_user_id INTEGER,
			action TEXT NOT NULL,
			reason TEXT NOT NULL DEFAULT '',
			content TEXT,
			blocked_department_anchor_user_id INTEGER,
			blocked_until DATETIME,
			operator_user_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE,
			FOREIGN KEY (from_owner_user_id) REFERENCES users(id),
			FOREIGN KEY (to_owner_user_id) REFERENCES users(id),
			FOREIGN KEY (operator_user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS customer_follow_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			customer_id INTEGER NOT NULL,
			operator_user_id INTEGER NOT NULL,
			content TEXT NOT NULL DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE,
			FOREIGN KEY (operator_user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS customer_phones (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			customer_id INTEGER NOT NULL,
			phone TEXT NOT NULL,
			phone_label TEXT DEFAULT NULL,
			is_primary INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS customer_status_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			customer_id INTEGER NOT NULL,
			from_status INTEGER NOT NULL,
			to_status INTEGER NOT NULL,
			trigger_type INTEGER NOT NULL DEFAULT 0,
			reason TEXT,
			operator_user_id INTEGER,
			operate_time INTEGER NOT NULL,
			FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS system_settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT NOT NULL UNIQUE,
			value TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS customer_levels (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			sort INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS customer_sources (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			sort INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS follow_methods (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			sort INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS operation_follow_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			customer_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			next_follow_time INTEGER,
			appointment_time INTEGER,
			shooting_time INTEGER,
			customer_level_id INTEGER NOT NULL DEFAULT 0,
			follow_method_id INTEGER NOT NULL DEFAULT 0,
			operator_user_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE,
			FOREIGN KEY (operator_user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS sales_follow_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			customer_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			next_follow_time INTEGER,
			customer_level_id INTEGER NOT NULL DEFAULT 0,
			customer_source_id INTEGER NOT NULL DEFAULT 0,
			follow_method_id INTEGER NOT NULL DEFAULT 0,
			operator_user_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE,
			FOREIGN KEY (operator_user_id) REFERENCES users(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_customers_owner_user_id ON customers(owner_user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_customers_inside_sales_user_id ON customers(inside_sales_user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_customers_converted_at ON customers(converted_at)`,
		`CREATE INDEX IF NOT EXISTS idx_customers_status ON customers(status)`,
		`CREATE INDEX IF NOT EXISTS idx_customers_deal_status ON customers(deal_status)`,
		`CREATE INDEX IF NOT EXISTS idx_customer_owner_logs_customer_id ON customer_owner_logs(customer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_customer_owner_logs_to_owner_user_id ON customer_owner_logs(to_owner_user_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uk_customer_phone ON customer_phones(customer_id, phone)`,
		`CREATE INDEX IF NOT EXISTS idx_customer_phones_customer_id ON customer_phones(customer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_customer_phones_phone ON customer_phones(phone)`,
		`CREATE INDEX IF NOT EXISTS idx_customer_phones_primary ON customer_phones(customer_id, is_primary)`,
		`CREATE INDEX IF NOT EXISTS idx_customer_status_logs_customer_id ON customer_status_logs(customer_id, operate_time)`,
		`CREATE INDEX IF NOT EXISTS idx_customer_status_logs_to_status ON customer_status_logs(to_status, operate_time)`,
		`CREATE INDEX IF NOT EXISTS idx_operation_follow_records_customer_id ON operation_follow_records(customer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sales_follow_records_customer_id ON sales_follow_records(customer_id)`,
	}
	return execStatements(tx, stmts)
}

func upAddMissingColumnsAndIndexes(tx *gorm.DB) error {
	columnMigrations := []struct {
		table      string
		column     string
		definition string
	}{
		{"operation_follow_records", "appointment_time", "INTEGER"},
		{"operation_follow_records", "shooting_time", "INTEGER"},
		{"sales_follow_records", "customer_source_id", "INTEGER NOT NULL DEFAULT 0"},
		{"customers", "legal_name", "TEXT NOT NULL DEFAULT ''"},
		{"customers", "contact_name", "TEXT NOT NULL DEFAULT ''"},
		{"customers", "weixin", "TEXT NOT NULL DEFAULT ''"},
		{"customers", "customer_level_id", "INTEGER NOT NULL DEFAULT 0"},
		{"customers", "customer_source_id", "INTEGER NOT NULL DEFAULT 0"},
		{"customers", "province", "INTEGER NOT NULL DEFAULT 0"},
		{"customers", "city", "INTEGER NOT NULL DEFAULT 0"},
		{"customers", "area", "INTEGER NOT NULL DEFAULT 0"},
		{"customers", "detail_address", "TEXT NOT NULL DEFAULT ''"},
		{"customers", "lng", "REAL NOT NULL DEFAULT 0"},
		{"customers", "lat", "REAL NOT NULL DEFAULT 0"},
		{"customers", "next_time", "INTEGER NOT NULL DEFAULT 0"},
		{"customers", "inside_sales_user_id", "INTEGER"},
		{"customers", "converted_at", "DATETIME"},
		{"customers", "follow_time", "INTEGER"},
		{"customers", "remark", "TEXT"},
		{"customers", "deal_time", "INTEGER"},
		{"customers", "customer_status", "INTEGER NOT NULL DEFAULT 0"},
		{"customers", "collect_time", "INTEGER"},
		{"customers", "drop_time", "INTEGER"},
		{"customers", "drop_user_id", "INTEGER"},
		{"customers", "create_user_id", "INTEGER NOT NULL DEFAULT 0"},
		{"customers", "operate_user_id", "INTEGER NOT NULL DEFAULT 0"},
		{"customers", "is_lock", "INTEGER NOT NULL DEFAULT 0"},
		{"customers", "delete_time", "INTEGER"},
	}
	for _, item := range columnMigrations {
		if err := addColumnIfNotExists(tx, item.table, item.column, item.definition); err != nil {
			return err
		}
	}

	indexStmts := []string{
		`CREATE INDEX IF NOT EXISTS idx_customers_owner_user_id ON customers(owner_user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_customers_inside_sales_user_id ON customers(inside_sales_user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_customers_converted_at ON customers(converted_at)`,
		`CREATE INDEX IF NOT EXISTS idx_customers_status ON customers(status)`,
		`CREATE INDEX IF NOT EXISTS idx_customers_deal_status ON customers(deal_status)`,
		`CREATE INDEX IF NOT EXISTS idx_customer_owner_logs_customer_id ON customer_owner_logs(customer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_customer_owner_logs_to_owner_user_id ON customer_owner_logs(to_owner_user_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uk_customer_phone ON customer_phones(customer_id, phone)`,
		`CREATE INDEX IF NOT EXISTS idx_customer_phones_customer_id ON customer_phones(customer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_customer_phones_phone ON customer_phones(phone)`,
		`CREATE INDEX IF NOT EXISTS idx_customer_phones_primary ON customer_phones(customer_id, is_primary)`,
		`CREATE INDEX IF NOT EXISTS idx_customer_status_logs_customer_id ON customer_status_logs(customer_id, operate_time)`,
		`CREATE INDEX IF NOT EXISTS idx_customer_status_logs_to_status ON customer_status_logs(to_status, operate_time)`,
		`CREATE INDEX IF NOT EXISTS idx_operation_follow_records_customer_id ON operation_follow_records(customer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sales_follow_records_customer_id ON sales_follow_records(customer_id)`,
	}
	return execStatements(tx, indexStmts)
}

func upCreateContractsTable(tx *gorm.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS contracts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			contract_image TEXT NOT NULL DEFAULT '',
			payment_image TEXT NOT NULL DEFAULT '',
			payment_status TEXT NOT NULL DEFAULT 'pending',
			user_id INTEGER NOT NULL,
			customer_id INTEGER NOT NULL,
			cooperation_type TEXT NOT NULL DEFAULT 'domestic',
			contract_number TEXT NOT NULL,
			contract_name TEXT NOT NULL,
			contract_amount REAL NOT NULL DEFAULT 0,
			payment_amount REAL NOT NULL DEFAULT 0,
			cooperation_years INTEGER NOT NULL DEFAULT 0,
			node_count INTEGER NOT NULL DEFAULT 0,
			service_user_id INTEGER,
			website_name TEXT NOT NULL DEFAULT '',
			website_url TEXT NOT NULL DEFAULT '',
			website_username TEXT NOT NULL DEFAULT '',
			is_online INTEGER NOT NULL DEFAULT 0,
			start_date INTEGER,
			end_date INTEGER,
			audit_status TEXT NOT NULL DEFAULT 'pending',
			expiry_handling_status TEXT NOT NULL DEFAULT 'pending',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (customer_id) REFERENCES customers(id),
			FOREIGN KEY (service_user_id) REFERENCES users(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_contracts_user_id ON contracts(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_contracts_customer_id ON contracts(customer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_contracts_service_user_id ON contracts(service_user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_contracts_payment_status ON contracts(payment_status)`,
		`CREATE INDEX IF NOT EXISTS idx_contracts_cooperation_type ON contracts(cooperation_type)`,
		`CREATE INDEX IF NOT EXISTS idx_contracts_audit_status ON contracts(audit_status)`,
		`CREATE INDEX IF NOT EXISTS idx_contracts_expiry_handling_status ON contracts(expiry_handling_status)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uk_contracts_contract_number ON contracts(contract_number)`,
	}
	return execStatements(tx, stmts)
}

func upSeedContractNumberPrefixSetting(tx *gorm.DB) error {
	stmts := []string{
		`INSERT OR IGNORE INTO system_settings(key,value,description) VALUES ('contract_number_prefix','zzy_','合同编号前缀')`,
	}
	return execStatements(tx, stmts)
}

func upSeedCustomerRuleSettings(tx *gorm.DB) error {
	stmts := []string{
		`INSERT OR IGNORE INTO system_settings(key,value,description) VALUES ('customer_auto_drop_enabled','true','客户自动掉库总开关')`,
		`INSERT OR IGNORE INTO system_settings(key,value,description) VALUES ('claim_freeze_days','7','本人客户进入公海后的回捡冷冻天数')`,
	}
	return execStatements(tx, stmts)
}

func upAddCustomerOwnerLogReasonAndContent(tx *gorm.DB) error {
	if err := addColumnIfNotExists(tx, "customer_owner_logs", "reason", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	return addColumnIfNotExists(tx, "customer_owner_logs", "content", "TEXT")
}

func upAddCustomerInsideSalesFields(tx *gorm.DB) error {
	if err := addColumnIfNotExists(tx, "customers", "inside_sales_user_id", "INTEGER"); err != nil {
		return err
	}
	if err := addColumnIfNotExists(tx, "customers", "converted_at", "DATETIME"); err != nil {
		return err
	}
	if err := addIndexIfNotExists(tx, "customers", "idx_customers_inside_sales_user_id", "inside_sales_user_id", false); err != nil {
		return err
	}
	return addIndexIfNotExists(tx, "customers", "idx_customers_converted_at", "converted_at", false)
}

func upAddCustomerOwnerLogClaimBlockFields(tx *gorm.DB) error {
	if err := addColumnIfNotExists(tx, "customer_owner_logs", "blocked_department_anchor_user_id", "INTEGER"); err != nil {
		return err
	}
	if err := addColumnIfNotExists(tx, "customer_owner_logs", "blocked_until", "DATETIME"); err != nil {
		return err
	}
	return execStatements(tx, []string{
		`CREATE INDEX IF NOT EXISTS idx_customer_owner_logs_customer_blocked_until_anchor ON customer_owner_logs(customer_id, blocked_until, blocked_department_anchor_user_id)`,
	})
}

func upAddContractRemarkColumn(tx *gorm.DB) error {
	return addColumnIfNotExists(tx, "contracts", "remark", "TEXT")
}

func upAddContractAuditColumns(tx *gorm.DB) error {
	if err := addColumnIfNotExists(tx, "contracts", "audit_comment", "TEXT"); err != nil {
		return err
	}
	if err := addColumnIfNotExists(tx, "contracts", "audited_by", "INTEGER"); err != nil {
		return err
	}
	return addColumnIfNotExists(tx, "contracts", "audited_at", "DATETIME")
}

func upAddUserSalesType(tx *gorm.DB) error {
	if err := addColumnIfNotExists(tx, "users", "sales_type", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	return execStatements(tx, []string{
		`CREATE INDEX IF NOT EXISTS idx_users_sales_type ON users(sales_type)`,
	})
}

func upCreateDailyUserCallStatsAndAddUsersHanghangCrmMobile(tx *gorm.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS daily_user_call_stats (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			stat_date DATE NOT NULL,
			user_id INTEGER,
			real_name TEXT NOT NULL DEFAULT '',
			mobile TEXT NOT NULL DEFAULT '',
			bind_num INTEGER NOT NULL DEFAULT 0,
			call_num INTEGER NOT NULL DEFAULT 0,
			not_connected INTEGER NOT NULL DEFAULT 0,
			connection_rate REAL NOT NULL DEFAULT 0,
			time_total INTEGER NOT NULL DEFAULT 0,
			total_minute TEXT NOT NULL DEFAULT '',
			total_second INTEGER NOT NULL DEFAULT 0,
			average_call_duration REAL NOT NULL DEFAULT 0,
			average_call_second REAL NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uk_daily_user_call_stats_date_name_mobile ON daily_user_call_stats(stat_date, real_name, mobile)`,
		`CREATE INDEX IF NOT EXISTS idx_daily_user_call_stats_stat_date ON daily_user_call_stats(stat_date)`,
		`CREATE INDEX IF NOT EXISTS idx_daily_user_call_stats_user_id ON daily_user_call_stats(user_id)`,
	}
	if err := execStatements(tx, stmts); err != nil {
		return err
	}

	if err := addColumnIfNotExists(tx, "users", "hanghang_crm_mobile", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}

	return execStatements(tx, []string{
		`CREATE INDEX IF NOT EXISTS idx_users_nickname_hanghang_crm_mobile ON users(nickname, hanghang_crm_mobile)`,
	})
}

func upCreateSalesDailyScores(tx *gorm.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS sales_daily_scores (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			score_date DATE NOT NULL,
			user_id INTEGER NOT NULL,
			user_name TEXT NOT NULL DEFAULT '',
			role_name TEXT NOT NULL DEFAULT '',
			call_num INTEGER NOT NULL DEFAULT 0,
			call_duration_second INTEGER NOT NULL DEFAULT 0,
			call_score_by_count INTEGER NOT NULL DEFAULT 0,
			call_score_by_duration INTEGER NOT NULL DEFAULT 0,
			call_score_type TEXT NOT NULL DEFAULT 'none',
			call_score INTEGER NOT NULL DEFAULT 0,
			visit_count INTEGER NOT NULL DEFAULT 0,
			visit_score INTEGER NOT NULL DEFAULT 0,
			new_customer_count INTEGER NOT NULL DEFAULT 0,
			new_customer_score INTEGER NOT NULL DEFAULT 0,
			total_score INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uk_sales_daily_scores_date_user ON sales_daily_scores(score_date, user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sales_daily_scores_user_id ON sales_daily_scores(user_id)`,
	}
	return execStatements(tx, stmts)
}

func upAddSalesDailyScoreReachedAt(tx *gorm.DB) error {
	if !tx.Migrator().HasTable("sales_daily_scores") {
		return nil
	}

	if err := addColumnIfNotExists(tx, "sales_daily_scores", "score_reached_at", "DATETIME"); err != nil {
		return err
	}
	return addIndexIfNotExists(
		tx,
		"sales_daily_scores",
		"idx_sales_daily_scores_date_score_time_user",
		"score_date, total_score, score_reached_at, user_id",
		false,
	)
}

func upCreateAuthTokenTables(tx *gorm.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS refresh_tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			token_hash TEXT NOT NULL UNIQUE,
			user_id INTEGER NOT NULL,
			expires_at INTEGER NOT NULL,
			revoked_at INTEGER,
			replaced_by_hash TEXT NOT NULL DEFAULT '',
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at)`,
		`CREATE TABLE IF NOT EXISTS token_blacklist (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			jti TEXT NOT NULL UNIQUE,
			expires_at INTEGER NOT NULL,
			revoked_at INTEGER NOT NULL,
			reason TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE INDEX IF NOT EXISTS idx_token_blacklist_expires_at ON token_blacklist(expires_at)`,
	}
	return execStatements(tx, stmts)
}

func upCreateResourcePoolTable(tx *gorm.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS resource_pool (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL DEFAULT '',
			phone TEXT NOT NULL DEFAULT '',
			address TEXT NOT NULL DEFAULT '',
			province TEXT NOT NULL DEFAULT '',
			city TEXT NOT NULL DEFAULT '',
			area TEXT NOT NULL DEFAULT '',
			latitude REAL NOT NULL DEFAULT 0,
			longitude REAL NOT NULL DEFAULT 0,
			source TEXT NOT NULL DEFAULT 'baidu',
			source_uid TEXT NOT NULL DEFAULT '',
			search_keyword TEXT NOT NULL DEFAULT '',
			search_radius INTEGER NOT NULL DEFAULT 0,
			search_region TEXT NOT NULL DEFAULT '',
			query_address TEXT NOT NULL DEFAULT '',
			center_latitude REAL NOT NULL DEFAULT 0,
			center_longitude REAL NOT NULL DEFAULT 0,
			created_by INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (source, source_uid)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_resource_pool_name ON resource_pool(name)`,
		`CREATE INDEX IF NOT EXISTS idx_resource_pool_phone ON resource_pool(phone)`,
		`CREATE INDEX IF NOT EXISTS idx_resource_pool_city ON resource_pool(city)`,
		`CREATE INDEX IF NOT EXISTS idx_resource_pool_updated_at ON resource_pool(updated_at)`,
	}
	return execStatements(tx, stmts)
}

func upAddResourcePoolConversionColumns(tx *gorm.DB) error {
	columnMigrations := []struct {
		table      string
		column     string
		definition string
	}{
		{"resource_pool", "converted", "INTEGER NOT NULL DEFAULT 0"},
		{"resource_pool", "converted_customer_id", "INTEGER"},
		{"resource_pool", "converted_at", "DATETIME"},
		{"resource_pool", "converted_by", "INTEGER"},
	}
	for _, item := range columnMigrations {
		if err := addColumnIfNotExists(tx, item.table, item.column, item.definition); err != nil {
			return err
		}
	}

	indexStmts := []string{
		`CREATE INDEX IF NOT EXISTS idx_resource_pool_converted ON resource_pool(converted)`,
		`CREATE INDEX IF NOT EXISTS idx_resource_pool_converted_customer_id ON resource_pool(converted_customer_id)`,
	}
	return execStatements(tx, indexStmts)
}

func upCreateExternalCompanySearchTables(tx *gorm.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS external_company_search_task (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_no TEXT NOT NULL,
			platform INTEGER NOT NULL DEFAULT 0,
			keyword TEXT NOT NULL,
			keyword_normalized TEXT NOT NULL DEFAULT '',
			region_keyword TEXT NOT NULL DEFAULT '',
			status INTEGER NOT NULL DEFAULT 1,
			priority INTEGER NOT NULL DEFAULT 100,
			target_count INTEGER NOT NULL DEFAULT 0,
			page_limit INTEGER NOT NULL DEFAULT 0,
			page_no INTEGER NOT NULL DEFAULT 0,
			progress_percent INTEGER NOT NULL DEFAULT 0,
			fetched_count INTEGER NOT NULL DEFAULT 0,
			saved_count INTEGER NOT NULL DEFAULT 0,
			duplicate_count INTEGER NOT NULL DEFAULT 0,
			failed_count INTEGER NOT NULL DEFAULT 0,
			retry_count INTEGER NOT NULL DEFAULT 0,
			max_retry_count INTEGER NOT NULL DEFAULT 3,
			next_run_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			locked_at DATETIME,
			last_heartbeat_at DATETIME,
			started_at DATETIME,
			finished_at DATETIME,
			worker_token TEXT NOT NULL DEFAULT '',
			search_options TEXT,
			resume_cursor TEXT,
			error_message TEXT NOT NULL DEFAULT '',
			created_by INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(task_no)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_external_company_search_task_status_run_at ON external_company_search_task(status, next_run_at, priority, id)`,
		`CREATE INDEX IF NOT EXISTS idx_external_company_search_task_platform_status ON external_company_search_task(platform, status, id)`,
		`CREATE INDEX IF NOT EXISTS idx_external_company_search_task_created_by ON external_company_search_task(created_by)`,
		`CREATE INDEX IF NOT EXISTS idx_external_company_search_task_created_at ON external_company_search_task(created_at)`,
		`CREATE TABLE IF NOT EXISTS external_company (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			company_no TEXT NOT NULL,
			platform INTEGER NOT NULL,
			platform_company_id TEXT NOT NULL DEFAULT '',
			dedupe_key TEXT NOT NULL,
			company_name TEXT NOT NULL,
			company_name_en TEXT NOT NULL DEFAULT '',
			company_url TEXT NOT NULL DEFAULT '',
			company_logo TEXT NOT NULL DEFAULT '',
			company_images TEXT,
			company_desc TEXT,
			country TEXT NOT NULL DEFAULT '',
			province TEXT NOT NULL DEFAULT '',
			city TEXT NOT NULL DEFAULT '',
			address TEXT NOT NULL DEFAULT '',
			main_products TEXT,
			business_type TEXT NOT NULL DEFAULT '',
			employee_count TEXT NOT NULL DEFAULT '',
			established_year TEXT NOT NULL DEFAULT '',
			annual_revenue TEXT NOT NULL DEFAULT '',
			certification TEXT,
			contact TEXT NOT NULL DEFAULT '',
			phone TEXT NOT NULL DEFAULT '',
			email TEXT NOT NULL DEFAULT '',
			data_version INTEGER NOT NULL DEFAULT 1,
			interest_status INTEGER NOT NULL DEFAULT 1,
			is_deleted INTEGER NOT NULL DEFAULT 0,
			raw_payload TEXT,
			first_seen_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_seen_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(company_no),
			UNIQUE(platform, dedupe_key)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_external_company_platform_company_id ON external_company(platform, platform_company_id)`,
		`CREATE INDEX IF NOT EXISTS idx_external_company_name ON external_company(company_name)`,
		`CREATE INDEX IF NOT EXISTS idx_external_company_city ON external_company(city)`,
		`CREATE INDEX IF NOT EXISTS idx_external_company_create_time ON external_company(create_time)`,
		`CREATE INDEX IF NOT EXISTS idx_external_company_last_seen_at ON external_company(last_seen_at)`,
		`CREATE INDEX IF NOT EXISTS idx_external_company_is_deleted ON external_company(is_deleted)`,
		`CREATE INDEX IF NOT EXISTS idx_external_company_interest_status ON external_company(interest_status)`,
		`CREATE TABLE IF NOT EXISTS external_company_search_result (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL,
			company_id INTEGER NOT NULL,
			platform INTEGER NOT NULL,
			keyword TEXT NOT NULL,
			region_keyword TEXT NOT NULL DEFAULT '',
			page_no INTEGER NOT NULL DEFAULT 0,
			rank_no INTEGER NOT NULL DEFAULT 0,
			is_new_company INTEGER NOT NULL DEFAULT 0,
			result_payload TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (task_id) REFERENCES external_company_search_task(id) ON DELETE CASCADE,
			FOREIGN KEY (company_id) REFERENCES external_company(id) ON DELETE CASCADE,
			UNIQUE(task_id, company_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_external_company_search_result_task_id ON external_company_search_result(task_id)`,
		`CREATE INDEX IF NOT EXISTS idx_external_company_search_result_company_id ON external_company_search_result(company_id)`,
		`CREATE INDEX IF NOT EXISTS idx_external_company_search_result_created_at ON external_company_search_result(created_at)`,
		`CREATE TABLE IF NOT EXISTS external_company_search_event (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL,
			seq_no INTEGER NOT NULL,
			event_type TEXT NOT NULL,
			message TEXT NOT NULL DEFAULT '',
			payload TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (task_id) REFERENCES external_company_search_task(id) ON DELETE CASCADE,
			UNIQUE(task_id, seq_no)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_external_company_search_event_task_created_at ON external_company_search_event(task_id, created_at)`,
	}
	return execStatements(tx, stmts)
}

func upCreateActivityLogsTable(tx *gorm.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS activity_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			action TEXT NOT NULL,
			target_type TEXT NOT NULL DEFAULT '',
			target_id INTEGER NOT NULL DEFAULT 0,
			target_name TEXT NOT NULL DEFAULT '',
			content TEXT NOT NULL DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS notification_reads (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			notification_key TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uk_notification_reads_user_key ON notification_reads(user_id, notification_key)`,
		`CREATE INDEX IF NOT EXISTS idx_activity_logs_user_id ON activity_logs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_activity_logs_action ON activity_logs(action)`,
		`CREATE INDEX IF NOT EXISTS idx_activity_logs_target ON activity_logs(target_type, target_id)`,
		`CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at ON activity_logs(created_at)`,
	}
	return execStatements(tx, stmts)
}

func upRenameSpiderCrawlTablesToExternalCompanySearchTables(tx *gorm.DB) error {
	tableRenames := []struct {
		from string
		to   string
	}{
		{from: "spider_crawl_task", to: "external_company_search_task"},
		{from: "spider_crawl_company", to: "external_company"},
		{from: "spider_crawl_task_result", to: "external_company_search_result"},
		{from: "spider_crawl_task_event", to: "external_company_search_event"},
	}
	for _, item := range tableRenames {
		if err := renameTableIfExists(tx, item.from, item.to); err != nil {
			return err
		}
	}
	return nil
}
