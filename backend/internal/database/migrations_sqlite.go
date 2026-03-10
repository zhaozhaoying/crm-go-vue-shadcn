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
	}
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
			FOREIGN KEY (owner_user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS customer_owner_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			customer_id INTEGER NOT NULL,
			from_owner_user_id INTEGER,
			to_owner_user_id INTEGER,
			action TEXT NOT NULL,
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
