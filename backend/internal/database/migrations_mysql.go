package database

import "gorm.io/gorm"

func getMySQLMigrations() []Migration {
	return []Migration{
		{
			Version: 2026030301,
			Name:    "create_base_schema",
			Up:      upCreateBaseSchemaMySQL,
		},
		{
			Version: 2026030302,
			Name:    "add_missing_columns_and_indexes",
			Up:      upAddMissingColumnsAndIndexesMySQL,
		},
		{
			Version: 2026030401,
			Name:    "create_contracts_table",
			Up:      upCreateContractsTableMySQL,
		},
		{
			Version: 2026030402,
			Name:    "seed_contract_number_prefix_setting",
			Up:      upSeedContractNumberPrefixSettingMySQL,
		},
		{
			Version: 2026030403,
			Name:    "create_auth_token_tables",
			Up:      upCreateAuthTokenTablesMySQL,
		},
		{
			Version: 2026030404,
			Name:    "create_resource_pool_table",
			Up:      upCreateResourcePoolTableMySQL,
		},
		{
			Version: 2026030405,
			Name:    "add_resource_pool_conversion_columns",
			Up:      upAddResourcePoolConversionColumnsMySQL,
		},
		{
			Version: 2026030501,
			Name:    "add_contract_remark_column",
			Up:      upAddContractRemarkColumnMySQL,
		},
		{
			Version: 2026030601,
			Name:    "create_external_company_search_tables",
			Up:      upCreateExternalCompanySearchTablesMySQL,
		},
		{
			Version: 2026030602,
			Name:    "rename_spider_crawl_tables_to_external_company_search_tables",
			Up:      upRenameSpiderCrawlTablesToExternalCompanySearchTablesMySQL,
		},
		{
			Version: 2026030604,
			Name:    "create_activity_logs_table",
			Up:      upCreateActivityLogsTableMySQL,
		},
		{
			Version: 2026030701,
			Name:    "add_contract_audit_columns",
			Up:      upAddContractAuditColumnsMySQL,
		},
		{
			Version: 2026030901,
			Name:    "fix_legacy_mysql_index_lengths",
			Up:      upFixLegacyMySQLIndexLengths,
		},
		{
			Version: 2026031101,
			Name:    "add_user_sales_type",
			Up:      upAddUserSalesTypeMySQL,
		},
		{
			Version: 2026031701,
			Name:    "seed_customer_rule_settings",
			Up:      upSeedCustomerRuleSettingsMySQL,
		},
	}
}

func upCreateBaseSchemaMySQL(tx *gorm.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS roles (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(64) NOT NULL UNIQUE,
			label VARCHAR(128) NOT NULL DEFAULT '',
			sort INT NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS users (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			username VARCHAR(64) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			salt VARCHAR(128) NOT NULL DEFAULT '',
			nickname VARCHAR(128) NOT NULL DEFAULT '',
			email VARCHAR(255) NOT NULL DEFAULT '',
			mobile VARCHAR(32) NOT NULL DEFAULT '',
			avatar VARCHAR(1024) NOT NULL DEFAULT '',
			role_id BIGINT NOT NULL DEFAULT 0,
			parent_id BIGINT NULL,
			sales_type VARCHAR(32) NOT NULL DEFAULT '',
			status VARCHAR(32) NOT NULL DEFAULT 'enabled',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (role_id) REFERENCES roles(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS customers (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(255) NOT NULL,
			legal_name VARCHAR(255) NOT NULL DEFAULT '',
			contact_name VARCHAR(255) NOT NULL DEFAULT '',
			weixin VARCHAR(128) NOT NULL DEFAULT '',
			email VARCHAR(255) NOT NULL DEFAULT '',
			status VARCHAR(32) NOT NULL DEFAULT 'pool',
			deal_status VARCHAR(32) NOT NULL DEFAULT 'undone',
			owner_user_id BIGINT NULL,
			customer_level_id INT NOT NULL DEFAULT 0,
			customer_source_id INT NOT NULL DEFAULT 0,
			province INT NOT NULL DEFAULT 0,
			city INT NOT NULL DEFAULT 0,
			area INT NOT NULL DEFAULT 0,
			detail_address VARCHAR(1024) NOT NULL DEFAULT '',
			lng DOUBLE NOT NULL DEFAULT 0,
			lat DOUBLE NOT NULL DEFAULT 0,
			next_time BIGINT NOT NULL DEFAULT 0,
			follow_time BIGINT NULL,
			remark TEXT NULL,
			deal_time BIGINT NULL,
			customer_status INT NOT NULL DEFAULT 0,
			collect_time BIGINT NULL,
			drop_time BIGINT NULL,
			drop_user_id BIGINT NULL,
			create_user_id BIGINT NOT NULL DEFAULT 0,
			operate_user_id BIGINT NOT NULL DEFAULT 0,
			is_lock TINYINT(1) NOT NULL DEFAULT 0,
			delete_time BIGINT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (owner_user_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS customer_owner_logs (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			customer_id BIGINT NOT NULL,
			from_owner_user_id BIGINT NULL,
			to_owner_user_id BIGINT NULL,
			action VARCHAR(32) NOT NULL,
			operator_user_id BIGINT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE,
			FOREIGN KEY (from_owner_user_id) REFERENCES users(id),
			FOREIGN KEY (to_owner_user_id) REFERENCES users(id),
			FOREIGN KEY (operator_user_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS customer_follow_logs (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			customer_id BIGINT NOT NULL,
			operator_user_id BIGINT NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE,
			FOREIGN KEY (operator_user_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS customer_phones (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			customer_id BIGINT NOT NULL,
			phone VARCHAR(32) NOT NULL,
			phone_label VARCHAR(64) DEFAULT NULL,
			is_primary TINYINT(1) NOT NULL DEFAULT 0,
			created_at BIGINT NOT NULL,
			updated_at BIGINT NOT NULL,
			FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS customer_status_logs (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			customer_id BIGINT NOT NULL,
			from_status INT NOT NULL,
			to_status INT NOT NULL,
			trigger_type INT NOT NULL DEFAULT 0,
			reason TEXT NULL,
			operator_user_id BIGINT NULL,
			operate_time BIGINT NOT NULL,
			FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS system_settings (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			` + "`key`" + ` VARCHAR(128) NOT NULL UNIQUE,
			value TEXT NOT NULL,
			description VARCHAR(255) NOT NULL DEFAULT '',
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS customer_levels (
			id INT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(64) NOT NULL,
			sort INT NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS customer_sources (
			id INT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(64) NOT NULL,
			sort INT NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS follow_methods (
			id INT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(64) NOT NULL,
			sort INT NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS operation_follow_records (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			customer_id BIGINT NOT NULL,
			content TEXT NOT NULL,
			next_follow_time BIGINT NULL,
			appointment_time BIGINT NULL,
			shooting_time BIGINT NULL,
			customer_level_id INT NOT NULL DEFAULT 0,
			follow_method_id INT NOT NULL DEFAULT 0,
			operator_user_id BIGINT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE,
			FOREIGN KEY (operator_user_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS sales_follow_records (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			customer_id BIGINT NOT NULL,
			content TEXT NOT NULL,
			next_follow_time BIGINT NULL,
			customer_level_id INT NOT NULL DEFAULT 0,
			customer_source_id INT NOT NULL DEFAULT 0,
			follow_method_id INT NOT NULL DEFAULT 0,
			operator_user_id BIGINT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE,
			FOREIGN KEY (operator_user_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
	if err := execStatements(tx, stmts); err != nil {
		return err
	}

	indexes := []struct {
		table   string
		name    string
		columns string
		unique  bool
	}{
		{"customers", "idx_customers_owner_user_id", "owner_user_id", false},
		{"customers", "idx_customers_status", "status", false},
		{"customers", "idx_customers_deal_status", "deal_status", false},
		{"customer_owner_logs", "idx_customer_owner_logs_customer_id", "customer_id", false},
		{"customer_owner_logs", "idx_customer_owner_logs_to_owner_user_id", "to_owner_user_id", false},
		{"customer_phones", "uk_customer_phone", "customer_id, phone", true},
		{"customer_phones", "idx_customer_phones_customer_id", "customer_id", false},
		{"customer_phones", "idx_customer_phones_phone", "phone", false},
		{"customer_phones", "idx_customer_phones_primary", "customer_id, is_primary", false},
		{"customer_status_logs", "idx_customer_status_logs_customer_id", "customer_id, operate_time", false},
		{"customer_status_logs", "idx_customer_status_logs_to_status", "to_status, operate_time", false},
		{"operation_follow_records", "idx_operation_follow_records_customer_id", "customer_id", false},
		{"sales_follow_records", "idx_sales_follow_records_customer_id", "customer_id", false},
	}
	for _, idx := range indexes {
		if err := addIndexIfNotExists(tx, idx.table, idx.name, idx.columns, idx.unique); err != nil {
			return err
		}
	}
	return nil
}

func upAddMissingColumnsAndIndexesMySQL(tx *gorm.DB) error {
	columnMigrations := []struct {
		table      string
		column     string
		definition string
	}{
		{"operation_follow_records", "appointment_time", "BIGINT NULL"},
		{"operation_follow_records", "shooting_time", "BIGINT NULL"},
		{"sales_follow_records", "customer_source_id", "INT NOT NULL DEFAULT 0"},
		{"customers", "legal_name", "VARCHAR(255) NOT NULL DEFAULT ''"},
		{"customers", "contact_name", "VARCHAR(255) NOT NULL DEFAULT ''"},
		{"customers", "weixin", "VARCHAR(128) NOT NULL DEFAULT ''"},
		{"customers", "customer_level_id", "INT NOT NULL DEFAULT 0"},
		{"customers", "customer_source_id", "INT NOT NULL DEFAULT 0"},
		{"customers", "province", "INT NOT NULL DEFAULT 0"},
		{"customers", "city", "INT NOT NULL DEFAULT 0"},
		{"customers", "area", "INT NOT NULL DEFAULT 0"},
		{"customers", "detail_address", "VARCHAR(1024) NOT NULL DEFAULT ''"},
		{"customers", "lng", "DOUBLE NOT NULL DEFAULT 0"},
		{"customers", "lat", "DOUBLE NOT NULL DEFAULT 0"},
		{"customers", "next_time", "BIGINT NOT NULL DEFAULT 0"},
		{"customers", "follow_time", "BIGINT NULL"},
		{"customers", "remark", "TEXT NULL"},
		{"customers", "deal_time", "BIGINT NULL"},
		{"customers", "customer_status", "INT NOT NULL DEFAULT 0"},
		{"customers", "collect_time", "BIGINT NULL"},
		{"customers", "drop_time", "BIGINT NULL"},
		{"customers", "drop_user_id", "BIGINT NULL"},
		{"customers", "create_user_id", "BIGINT NOT NULL DEFAULT 0"},
		{"customers", "operate_user_id", "BIGINT NOT NULL DEFAULT 0"},
		{"customers", "is_lock", "TINYINT(1) NOT NULL DEFAULT 0"},
		{"customers", "delete_time", "BIGINT NULL"},
	}
	for _, item := range columnMigrations {
		if err := addColumnIfNotExists(tx, item.table, item.column, item.definition); err != nil {
			return err
		}
	}

	indexes := []struct {
		table   string
		name    string
		columns string
		unique  bool
	}{
		{"customers", "idx_customers_owner_user_id", "owner_user_id", false},
		{"customers", "idx_customers_status", "status", false},
		{"customers", "idx_customers_deal_status", "deal_status", false},
		{"customer_owner_logs", "idx_customer_owner_logs_customer_id", "customer_id", false},
		{"customer_owner_logs", "idx_customer_owner_logs_to_owner_user_id", "to_owner_user_id", false},
		{"customer_phones", "uk_customer_phone", "customer_id, phone", true},
		{"customer_phones", "idx_customer_phones_customer_id", "customer_id", false},
		{"customer_phones", "idx_customer_phones_phone", "phone", false},
		{"customer_phones", "idx_customer_phones_primary", "customer_id, is_primary", false},
		{"customer_status_logs", "idx_customer_status_logs_customer_id", "customer_id, operate_time", false},
		{"customer_status_logs", "idx_customer_status_logs_to_status", "to_status, operate_time", false},
		{"operation_follow_records", "idx_operation_follow_records_customer_id", "customer_id", false},
		{"sales_follow_records", "idx_sales_follow_records_customer_id", "customer_id", false},
	}
	for _, idx := range indexes {
		if err := addIndexIfNotExists(tx, idx.table, idx.name, idx.columns, idx.unique); err != nil {
			return err
		}
	}
	return nil
}

func upCreateContractsTableMySQL(tx *gorm.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS contracts (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			contract_image VARCHAR(1024) NOT NULL DEFAULT '',
			payment_image VARCHAR(1024) NOT NULL DEFAULT '',
			payment_status VARCHAR(32) NOT NULL DEFAULT 'pending',
			user_id BIGINT NOT NULL,
			customer_id BIGINT NOT NULL,
			cooperation_type VARCHAR(32) NOT NULL DEFAULT 'domestic',
			contract_number VARCHAR(128) NOT NULL,
			contract_name VARCHAR(255) NOT NULL,
			contract_amount DECIMAL(18,2) NOT NULL DEFAULT 0,
			payment_amount DECIMAL(18,2) NOT NULL DEFAULT 0,
			cooperation_years INT NOT NULL DEFAULT 0,
			node_count INT NOT NULL DEFAULT 0,
			service_user_id BIGINT NULL,
			website_name VARCHAR(255) NOT NULL DEFAULT '',
			website_url VARCHAR(1024) NOT NULL DEFAULT '',
			website_username VARCHAR(255) NOT NULL DEFAULT '',
			is_online TINYINT(1) NOT NULL DEFAULT 0,
			start_date BIGINT NULL,
			end_date BIGINT NULL,
			audit_status VARCHAR(32) NOT NULL DEFAULT 'pending',
			expiry_handling_status VARCHAR(32) NOT NULL DEFAULT 'pending',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (customer_id) REFERENCES customers(id),
			FOREIGN KEY (service_user_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
	if err := execStatements(tx, stmts); err != nil {
		return err
	}

	indexes := []struct {
		table   string
		name    string
		columns string
		unique  bool
	}{
		{"contracts", "idx_contracts_user_id", "user_id", false},
		{"contracts", "idx_contracts_customer_id", "customer_id", false},
		{"contracts", "idx_contracts_service_user_id", "service_user_id", false},
		{"contracts", "idx_contracts_payment_status", "payment_status", false},
		{"contracts", "idx_contracts_cooperation_type", "cooperation_type", false},
		{"contracts", "idx_contracts_audit_status", "audit_status", false},
		{"contracts", "idx_contracts_expiry_handling_status", "expiry_handling_status", false},
		{"contracts", "uk_contracts_contract_number", "contract_number", true},
	}
	for _, idx := range indexes {
		if err := addIndexIfNotExists(tx, idx.table, idx.name, idx.columns, idx.unique); err != nil {
			return err
		}
	}
	return nil
}

func upSeedContractNumberPrefixSettingMySQL(tx *gorm.DB) error {
	stmts := []string{
		`INSERT IGNORE INTO system_settings(` + "`key`" + `, value, description) VALUES ('contract_number_prefix','zzy_','合同编号前缀')`,
	}
	return execStatements(tx, stmts)
}

func upSeedCustomerRuleSettingsMySQL(tx *gorm.DB) error {
	stmts := []string{
		`INSERT IGNORE INTO system_settings(` + "`key`" + `, value, description) VALUES ('customer_auto_drop_enabled','true','客户自动掉库总开关')`,
		`INSERT IGNORE INTO system_settings(` + "`key`" + `, value, description) VALUES ('claim_freeze_days','7','本人客户进入公海后的回捡冷冻天数')`,
	}
	return execStatements(tx, stmts)
}

func upAddContractRemarkColumnMySQL(tx *gorm.DB) error {
	return addColumnIfNotExists(tx, "contracts", "remark", "TEXT NULL")
}

func upAddContractAuditColumnsMySQL(tx *gorm.DB) error {
	if err := addColumnIfNotExists(tx, "contracts", "audit_comment", "TEXT NULL"); err != nil {
		return err
	}
	if err := addColumnIfNotExists(tx, "contracts", "audited_by", "BIGINT NULL"); err != nil {
		return err
	}
	return addColumnIfNotExists(tx, "contracts", "audited_at", "DATETIME NULL")
}

func upCreateAuthTokenTablesMySQL(tx *gorm.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS refresh_tokens (
				id BIGINT PRIMARY KEY AUTO_INCREMENT,
				token_hash VARCHAR(64) NOT NULL UNIQUE,
				user_id BIGINT NOT NULL,
				expires_at BIGINT NOT NULL,
				revoked_at BIGINT NULL,
				replaced_by_hash VARCHAR(64) NOT NULL DEFAULT '',
				created_at BIGINT NOT NULL,
				updated_at BIGINT NOT NULL,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS token_blacklist (
				id BIGINT PRIMARY KEY AUTO_INCREMENT,
				jti VARCHAR(64) NOT NULL UNIQUE,
				expires_at BIGINT NOT NULL,
				revoked_at BIGINT NOT NULL,
				reason VARCHAR(255) NOT NULL DEFAULT ''
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
	if err := execStatements(tx, stmts); err != nil {
		return err
	}

	indexes := []struct {
		table   string
		name    string
		columns string
		unique  bool
	}{
		{"refresh_tokens", "idx_refresh_tokens_user_id", "user_id", false},
		{"refresh_tokens", "idx_refresh_tokens_expires_at", "expires_at", false},
		{"token_blacklist", "idx_token_blacklist_expires_at", "expires_at", false},
	}
	for _, idx := range indexes {
		if err := addIndexIfNotExists(tx, idx.table, idx.name, idx.columns, idx.unique); err != nil {
			return err
		}
	}
	return nil
}

func upCreateResourcePoolTableMySQL(tx *gorm.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS resource_pool (
				id BIGINT PRIMARY KEY AUTO_INCREMENT,
				name VARCHAR(255) NOT NULL DEFAULT '',
				phone VARCHAR(128) NOT NULL DEFAULT '',
			address VARCHAR(1024) NOT NULL DEFAULT '',
			province VARCHAR(64) NOT NULL DEFAULT '',
			city VARCHAR(64) NOT NULL DEFAULT '',
			area VARCHAR(64) NOT NULL DEFAULT '',
				latitude DOUBLE NOT NULL DEFAULT 0,
				longitude DOUBLE NOT NULL DEFAULT 0,
				source VARCHAR(32) NOT NULL DEFAULT 'baidu',
				source_uid VARCHAR(150) NOT NULL DEFAULT '',
				search_keyword VARCHAR(255) NOT NULL DEFAULT '',
				search_radius INT NOT NULL DEFAULT 0,
				search_region VARCHAR(255) NOT NULL DEFAULT '',
			query_address VARCHAR(1024) NOT NULL DEFAULT '',
			center_latitude DOUBLE NOT NULL DEFAULT 0,
			center_longitude DOUBLE NOT NULL DEFAULT 0,
			created_by BIGINT NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uk_resource_pool_source_uid (source, source_uid)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
	if err := execStatements(tx, stmts); err != nil {
		return err
	}

	indexes := []struct {
		table   string
		name    string
		columns string
		unique  bool
	}{
		{"resource_pool", "idx_resource_pool_name", "name(191)", false},
		{"resource_pool", "idx_resource_pool_phone", "phone", false},
		{"resource_pool", "idx_resource_pool_city", "city", false},
		{"resource_pool", "idx_resource_pool_updated_at", "updated_at", false},
	}
	for _, idx := range indexes {
		if err := addIndexIfNotExists(tx, idx.table, idx.name, idx.columns, idx.unique); err != nil {
			return err
		}
	}
	return nil
}

func upAddResourcePoolConversionColumnsMySQL(tx *gorm.DB) error {
	columnMigrations := []struct {
		table      string
		column     string
		definition string
	}{
		{"resource_pool", "converted", "TINYINT(1) NOT NULL DEFAULT 0"},
		{"resource_pool", "converted_customer_id", "BIGINT NULL"},
		{"resource_pool", "converted_at", "DATETIME NULL"},
		{"resource_pool", "converted_by", "BIGINT NULL"},
	}
	for _, item := range columnMigrations {
		if err := addColumnIfNotExists(tx, item.table, item.column, item.definition); err != nil {
			return err
		}
	}

	indexes := []struct {
		table   string
		name    string
		columns string
		unique  bool
	}{
		{"resource_pool", "idx_resource_pool_converted", "converted", false},
		{"resource_pool", "idx_resource_pool_converted_customer_id", "converted_customer_id", false},
	}
	for _, idx := range indexes {
		if err := addIndexIfNotExists(tx, idx.table, idx.name, idx.columns, idx.unique); err != nil {
			return err
		}
	}
	return nil
}

func upCreateExternalCompanySearchTablesMySQL(tx *gorm.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS external_company_search_task (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			task_no VARCHAR(64) NOT NULL,
			platform TINYINT NOT NULL DEFAULT 0,
			keyword VARCHAR(255) NOT NULL,
			keyword_normalized VARCHAR(255) NOT NULL DEFAULT '',
			region_keyword VARCHAR(255) NOT NULL DEFAULT '',
			status TINYINT NOT NULL DEFAULT 1,
			priority INT NOT NULL DEFAULT 100,
			target_count INT NOT NULL DEFAULT 0,
			page_limit INT NOT NULL DEFAULT 0,
			page_no INT NOT NULL DEFAULT 0,
			progress_percent INT NOT NULL DEFAULT 0,
			fetched_count INT NOT NULL DEFAULT 0,
			saved_count INT NOT NULL DEFAULT 0,
			duplicate_count INT NOT NULL DEFAULT 0,
			failed_count INT NOT NULL DEFAULT 0,
			retry_count INT NOT NULL DEFAULT 0,
			max_retry_count INT NOT NULL DEFAULT 3,
			next_run_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			locked_at DATETIME NULL,
			last_heartbeat_at DATETIME NULL,
			started_at DATETIME NULL,
			finished_at DATETIME NULL,
			worker_token VARCHAR(64) NOT NULL DEFAULT '',
			search_options TEXT NULL,
			resume_cursor TEXT NULL,
			error_message VARCHAR(1000) NOT NULL DEFAULT '',
			created_by BIGINT NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_external_company_search_task_task_no (task_no)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS external_company (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			company_no VARCHAR(64) NOT NULL,
			platform TINYINT NOT NULL,
			platform_company_id VARCHAR(100) NOT NULL DEFAULT '',
			dedupe_key VARCHAR(191) NOT NULL,
			company_name VARCHAR(500) NOT NULL,
			company_name_en VARCHAR(500) NOT NULL DEFAULT '',
			company_url VARCHAR(1000) NOT NULL DEFAULT '',
			company_logo VARCHAR(1000) NOT NULL DEFAULT '',
			company_images LONGTEXT NULL,
			company_desc LONGTEXT NULL,
			country VARCHAR(100) NOT NULL DEFAULT '',
			province VARCHAR(100) NOT NULL DEFAULT '',
			city VARCHAR(100) NOT NULL DEFAULT '',
			address VARCHAR(500) NOT NULL DEFAULT '',
			main_products LONGTEXT NULL,
			business_type VARCHAR(200) NOT NULL DEFAULT '',
			employee_count VARCHAR(50) NOT NULL DEFAULT '',
			established_year VARCHAR(20) NOT NULL DEFAULT '',
			annual_revenue VARCHAR(100) NOT NULL DEFAULT '',
			certification LONGTEXT NULL,
			contact VARCHAR(100) NOT NULL DEFAULT '',
			phone VARCHAR(50) NOT NULL DEFAULT '',
			email VARCHAR(100) NOT NULL DEFAULT '',
			data_version INT NOT NULL DEFAULT 1,
			interest_status TINYINT NOT NULL DEFAULT 1,
			is_deleted TINYINT NOT NULL DEFAULT 0,
			raw_payload LONGTEXT NULL,
			first_seen_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			last_seen_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_external_company_company_no (company_no),
			UNIQUE KEY uk_external_company_platform_dedupe (platform, dedupe_key)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS external_company_search_result (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			task_id BIGINT NOT NULL,
			company_id BIGINT NOT NULL,
			platform TINYINT NOT NULL,
			keyword VARCHAR(255) NOT NULL,
			region_keyword VARCHAR(255) NOT NULL DEFAULT '',
			page_no INT NOT NULL DEFAULT 0,
			rank_no INT NOT NULL DEFAULT 0,
			is_new_company TINYINT(1) NOT NULL DEFAULT 0,
			result_payload LONGTEXT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (task_id) REFERENCES external_company_search_task(id) ON DELETE CASCADE,
			FOREIGN KEY (company_id) REFERENCES external_company(id) ON DELETE CASCADE,
			UNIQUE KEY uk_external_company_search_result_task_company (task_id, company_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS external_company_search_event (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			task_id BIGINT NOT NULL,
			seq_no BIGINT NOT NULL,
			event_type VARCHAR(64) NOT NULL,
			message VARCHAR(500) NOT NULL DEFAULT '',
			payload LONGTEXT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (task_id) REFERENCES external_company_search_task(id) ON DELETE CASCADE,
			UNIQUE KEY uk_external_company_search_event_task_seq (task_id, seq_no)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
	if err := execStatements(tx, stmts); err != nil {
		return err
	}

	indexes := []struct {
		table   string
		name    string
		columns string
		unique  bool
	}{
		{"external_company_search_task", "idx_external_company_search_task_status_run_at", "status, next_run_at, priority, id", false},
		{"external_company_search_task", "idx_external_company_search_task_platform_status", "platform, status, id", false},
		{"external_company_search_task", "idx_external_company_search_task_created_by", "created_by", false},
		{"external_company_search_task", "idx_external_company_search_task_created_at", "created_at", false},
		{"external_company", "idx_external_company_platform_company_id", "platform, platform_company_id", false},
		{"external_company", "idx_external_company_name", "company_name(100)", false},
		{"external_company", "idx_external_company_city", "city", false},
		{"external_company", "idx_external_company_create_time", "create_time", false},
		{"external_company", "idx_external_company_last_seen_at", "last_seen_at", false},
		{"external_company", "idx_external_company_is_deleted", "is_deleted", false},
		{"external_company", "idx_external_company_interest_status", "interest_status", false},
		{"external_company_search_result", "idx_external_company_search_result_task_id", "task_id", false},
		{"external_company_search_result", "idx_external_company_search_result_company_id", "company_id", false},
		{"external_company_search_result", "idx_external_company_search_result_created_at", "created_at", false},
		{"external_company_search_event", "idx_external_company_search_event_task_created_at", "task_id, created_at", false},
	}
	for _, idx := range indexes {
		if err := addIndexIfNotExists(tx, idx.table, idx.name, idx.columns, idx.unique); err != nil {
			return err
		}
	}
	return nil
}

func upCreateActivityLogsTableMySQL(tx *gorm.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS activity_logs (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			user_id BIGINT NOT NULL,
			action VARCHAR(64) NOT NULL,
			target_type VARCHAR(64) NOT NULL DEFAULT '',
			target_id BIGINT NOT NULL DEFAULT 0,
			target_name VARCHAR(255) NOT NULL DEFAULT '',
			content TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS notification_reads (
				id BIGINT PRIMARY KEY AUTO_INCREMENT,
				user_id BIGINT NOT NULL,
				notification_key VARCHAR(150) NOT NULL,
				created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
				UNIQUE KEY uk_notification_reads_user_key (user_id, notification_key),
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
	if err := execStatements(tx, stmts); err != nil {
		return err
	}

	indexes := []struct {
		table   string
		name    string
		columns string
		unique  bool
	}{
		{"activity_logs", "idx_activity_logs_user_id", "user_id", false},
		{"activity_logs", "idx_activity_logs_action", "action", false},
		{"activity_logs", "idx_activity_logs_target", "target_type, target_id", false},
		{"activity_logs", "idx_activity_logs_created_at", "created_at", false},
	}
	for _, idx := range indexes {
		if err := addIndexIfNotExists(tx, idx.table, idx.name, idx.columns, idx.unique); err != nil {
			return err
		}
	}
	return nil
}

func upRenameSpiderCrawlTablesToExternalCompanySearchTablesMySQL(tx *gorm.DB) error {
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

func upFixLegacyMySQLIndexLengths(tx *gorm.DB) error {
	alterStatements := []struct {
		table string
		stmt  string
	}{
		{
			table: "refresh_tokens",
			stmt:  "ALTER TABLE refresh_tokens MODIFY COLUMN token_hash VARCHAR(64) NOT NULL",
		},
		{
			table: "refresh_tokens",
			stmt:  "ALTER TABLE refresh_tokens MODIFY COLUMN replaced_by_hash VARCHAR(64) NOT NULL DEFAULT ''",
		},
		{
			table: "token_blacklist",
			stmt:  "ALTER TABLE token_blacklist MODIFY COLUMN jti VARCHAR(64) NOT NULL",
		},
		{
			table: "resource_pool",
			stmt:  "ALTER TABLE resource_pool MODIFY COLUMN source_uid VARCHAR(150) NOT NULL DEFAULT ''",
		},
		{
			table: "notification_reads",
			stmt:  "ALTER TABLE notification_reads MODIFY COLUMN notification_key VARCHAR(150) NOT NULL",
		},
	}
	for _, item := range alterStatements {
		if !tx.Migrator().HasTable(item.table) {
			continue
		}
		if err := tx.Exec(item.stmt).Error; err != nil {
			return err
		}
	}

	if !tx.Migrator().HasTable("resource_pool") {
		return nil
	}

	exists, err := hasIndex(tx, "resource_pool", "idx_resource_pool_name")
	if err != nil {
		return err
	}
	if exists {
		if err := tx.Exec("DROP INDEX idx_resource_pool_name ON resource_pool").Error; err != nil {
			return err
		}
	}

	return addIndexIfNotExists(tx, "resource_pool", "idx_resource_pool_name", "name(191)", false)
}

func upAddUserSalesTypeMySQL(tx *gorm.DB) error {
	if err := addColumnIfNotExists(tx, "users", "sales_type", "VARCHAR(32) NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	return addIndexIfNotExists(tx, "users", "idx_users_sales_type", "sales_type", false)
}
