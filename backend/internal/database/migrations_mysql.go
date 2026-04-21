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
		{
			Version: 2026031801,
			Name:    "add_customer_owner_log_reason_and_content",
			Up:      upAddCustomerOwnerLogReasonAndContentMySQL,
		},
		{
			Version: 2026031802,
			Name:    "add_customer_inside_sales_fields",
			Up:      upAddCustomerInsideSalesFieldsMySQL,
		},
		{
			Version: 2026031901,
			Name:    "add_customer_owner_log_claim_block_fields",
			Up:      upAddCustomerOwnerLogClaimBlockFieldsMySQL,
		},
		{
			Version: 2026031902,
			Name:    "create_daily_user_call_stats_and_add_users_hanghang_crm_mobile",
			Up:      upCreateDailyUserCallStatsAndAddUsersHanghangCrmMobileMySQL,
		},
		{
			Version: 2026032002,
			Name:    "create_customer_visits_table",
			Up:      upCreateCustomerVisitsTableMySQL,
		},
		{
			Version: 2026032003,
			Name:    "create_sales_daily_scores",
			Up:      upCreateSalesDailyScoresMySQL,
		},
		{
			Version: 2026032004,
			Name:    "alter_customer_visits_region_columns_to_varchar",
			Up:      upAlterCustomerVisitsRegionColumnsToVarcharMySQL,
		},
		{
			Version: 2026032005,
			Name:    "add_customer_visits_check_in_ip",
			Up:      upAddCustomerVisitsCheckInIPMySQL,
		},
		{
			Version: 2026032301,
			Name:    "create_call_recordings",
			Up:      upCreateCallRecordingsMySQL,
		},
		{
			Version: 2026032302,
			Name:    "create_user_hanghang_crm_mobiles",
			Up:      upCreateUserHanghangCRMMobilesMySQL,
		},
		{
			Version: 2026032401,
			Name:    "dedupe_call_recordings",
			Up:      upDedupeCallRecordingsMySQL,
		},
		{
			Version: 2026032501,
			Name:    "add_sales_daily_score_reached_at",
			Up:      upAddSalesDailyScoreReachedAtMySQL,
		},
		{
			Version: 2026032601,
			Name:    "add_customer_assign_time_and_sales_assign_drop_setting",
			Up:      upAddCustomerAssignTimeAndSalesAssignDropSettingMySQL,
		},
		{
			Version: 2026040301,
			Name:    "add_customer_visits_inviter",
			Up:      upAddCustomerVisitsInviterMySQL,
		},
		{
			Version: 2026042001,
			Name:    "add_user_mihua_work_number",
			Up:      upAddUserMihuaWorkNumberMySQL,
		},
		{
			Version: 2026042002,
			Name:    "create_spxxjj_telemarketing_tables",
			Up:      upCreateSpxxjjTelemarketingTablesMySQL,
		},
		{
			Version: 2026042101,
			Name:    "create_mihua_call_recordings",
			Up:      upCreateMiHuaCallRecordingsMySQL,
		},
	}
}

func upAddUserMihuaWorkNumberMySQL(tx *gorm.DB) error {
	return addColumnIfNotExists(tx, "users", "mihua_work_number", "VARCHAR(64) NOT NULL DEFAULT ''")
}

func upCreateSpxxjjTelemarketingTablesMySQL(tx *gorm.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS spxxjj_mihua_seat_statistics (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			score_date DATE NOT NULL COMMENT '统计日期',
			seat_id BIGINT NOT NULL DEFAULT 0 COMMENT '米话坐席ID',
			seat_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '坐席展示名',
			work_number VARCHAR(64) NOT NULL DEFAULT '' COMMENT '米话工号',
			service_number VARCHAR(64) NOT NULL DEFAULT '' COMMENT '坐席分机号',
			is_mobile_seat VARCHAR(8) NOT NULL DEFAULT '' COMMENT '是否手机坐席',
			seat_type INT NOT NULL DEFAULT 0 COMMENT '坐席类型',
			ccgeid BIGINT NOT NULL DEFAULT 0 COMMENT '企业组ID',
			success_call_count INT NOT NULL DEFAULT 0 COMMENT '接通数',
			out_total_success INT NOT NULL DEFAULT 0 COMMENT '外呼接通数',
			out_total_call_count INT NOT NULL DEFAULT 0 COMMENT '外呼总通次',
			call_total_time_second INT NOT NULL DEFAULT 0 COMMENT '总通话秒数',
			call_valid_time_second INT NOT NULL DEFAULT 0 COMMENT '总有效通话秒数',
			out_call_total_time_second INT NOT NULL DEFAULT 0 COMMENT '外呼总通话秒数',
			out_call_valid_time_second INT NOT NULL DEFAULT 0 COMMENT '外呼有效通话秒数',
			latest_state_time DATETIME NULL COMMENT '最新状态时间',
			latest_state_id INT NOT NULL DEFAULT 0 COMMENT '最新状态ID',
			stat_timestamp DATETIME NULL COMMENT '统计时间戳',
			enterprise_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '企业名称',
			department_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '部门名称',
			group_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '分组名称',
			seat_real_time_state_json LONGTEXT NULL COMMENT '坐席实时状态JSON',
			groups_json LONGTEXT NULL COMMENT '分组JSON',
			raw_payload LONGTEXT NULL COMMENT '米话原始记录JSON',
			matched_user_id BIGINT NULL COMMENT '匹配到的本地用户ID',
			matched_user_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '匹配到的本地用户名',
			role_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '角色名',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
			UNIQUE KEY uk_spxxjj_mihua_seat_statistics_date_work_number (score_date, work_number),
			KEY idx_spxxjj_mihua_seat_statistics_user_id (matched_user_id),
			KEY idx_spxxjj_mihua_seat_statistics_score_date (score_date)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='spxxjj 米话电销原始统计表'`,
		`CREATE TABLE IF NOT EXISTS spxxjj_telemarketing_daily_scores (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			score_date DATE NOT NULL COMMENT '统计日期',
			seat_work_number VARCHAR(64) NOT NULL DEFAULT '' COMMENT '米话工号',
			seat_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '坐席展示名',
			matched_user_id BIGINT NULL COMMENT '匹配到的本地用户ID',
			matched_user_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '匹配到的本地用户名',
			service_number VARCHAR(64) NOT NULL DEFAULT '' COMMENT '坐席分机号',
			group_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '分组名称',
			role_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '角色名',
			call_num INT NOT NULL DEFAULT 0 COMMENT '拨打数',
			answered_call_count INT NOT NULL DEFAULT 0 COMMENT '接通数',
			missed_call_count INT NOT NULL DEFAULT 0 COMMENT '未接通数',
			answer_rate DOUBLE NOT NULL DEFAULT 0 COMMENT '接通率',
			call_duration_second INT NOT NULL DEFAULT 0 COMMENT '通话时长秒数',
			new_customer_count INT NOT NULL DEFAULT 0 COMMENT '新增客户数',
			invitation_count INT NOT NULL DEFAULT 0 COMMENT '邀约数',
			call_score_by_count INT NOT NULL DEFAULT 0 COMMENT '按通话量积分',
			call_score_by_duration INT NOT NULL DEFAULT 0 COMMENT '按通话时长积分',
			call_score_type VARCHAR(32) NOT NULL DEFAULT 'none' COMMENT '电话积分口径',
			call_score INT NOT NULL DEFAULT 0 COMMENT '电话积分',
			invitation_score INT NOT NULL DEFAULT 0 COMMENT '邀约积分',
			new_customer_score INT NOT NULL DEFAULT 0 COMMENT '新增客户积分',
			total_score INT NOT NULL DEFAULT 0 COMMENT '总积分',
			score_reached_at DATETIME NULL COMMENT '达到当前积分时间',
			data_updated_at DATETIME NULL COMMENT '米话数据更新时间',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
			UNIQUE KEY uk_spxxjj_telemarketing_daily_scores_date_work_number (score_date, seat_work_number),
			KEY idx_spxxjj_telemarketing_daily_scores_user_id (matched_user_id),
			KEY idx_spxxjj_telemarketing_daily_scores_rank (score_date, total_score, answered_call_count, call_duration_second, seat_work_number)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='spxxjj 电销每日积分表'`,
	}
	return execStatements(tx, stmts)
}

func upCreateMiHuaCallRecordingsMySQL(tx *gorm.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS mihua_call_recordings (
			id VARCHAR(64) PRIMARY KEY COMMENT '米话录音ID',
			cc_number VARCHAR(128) NOT NULL DEFAULT '' COMMENT '通话唯一标识',
			sid BIGINT NOT NULL DEFAULT 0 COMMENT 'sid',
			seid BIGINT NOT NULL DEFAULT 0 COMMENT '企业ID',
			ccgeid BIGINT NOT NULL DEFAULT 0 COMMENT '企业组ID',
			call_type INT NOT NULL DEFAULT 0 COMMENT '通话类型',
			outline_number VARCHAR(64) NOT NULL DEFAULT '' COMMENT '客户号码',
			encrypted_outline_number VARCHAR(255) NOT NULL DEFAULT '' COMMENT '加密客户号码',
			switch_number VARCHAR(64) NOT NULL DEFAULT '' COMMENT '交换机号码',
			initiator VARCHAR(128) NOT NULL DEFAULT '' COMMENT '发起方',
			initiator_call_id VARCHAR(128) NOT NULL DEFAULT '' COMMENT '发起 call id',
			service_number VARCHAR(64) NOT NULL DEFAULT '' COMMENT '坐席分机号',
			service_uid BIGINT NOT NULL DEFAULT 0 COMMENT '坐席远端 UID',
			service_seat_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '坐席名称',
			service_seat_worknumber VARCHAR(64) NOT NULL DEFAULT '' COMMENT '坐席工号',
			service_group_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '坐席分组',
			initiate_time BIGINT NOT NULL DEFAULT 0 COMMENT '发起时间（秒）',
			ring_time BIGINT NOT NULL DEFAULT 0 COMMENT '振铃时间（秒）',
			confirm_time BIGINT NOT NULL DEFAULT 0 COMMENT '接通时间（秒）',
			disconnect_time BIGINT NOT NULL DEFAULT 0 COMMENT '挂断时间（秒）',
			conversation_time BIGINT NOT NULL DEFAULT 0 COMMENT '通话开始时间（秒）',
			duration_second INT NOT NULL DEFAULT 0 COMMENT '通话时长秒数',
			duration_text VARCHAR(64) NOT NULL DEFAULT '' COMMENT '通话时长文本',
			valid_duration_text VARCHAR(64) NOT NULL DEFAULT '' COMMENT '有效通话时长文本',
			customer_ring_duration INT NOT NULL DEFAULT 0 COMMENT '客户振铃秒数',
			seat_ring_duration INT NOT NULL DEFAULT 0 COMMENT '坐席振铃秒数',
			record_status INT NOT NULL DEFAULT 0 COMMENT '录音状态',
			record_filename VARCHAR(255) NOT NULL DEFAULT '' COMMENT '录音文件名',
			record_res_token VARCHAR(255) NOT NULL DEFAULT '' COMMENT '录音 token',
			evaluate_value VARCHAR(128) NOT NULL DEFAULT '' COMMENT '评价值',
			cm_result VARCHAR(128) NOT NULL DEFAULT '' COMMENT '结果',
			cm_description VARCHAR(255) NOT NULL DEFAULT '' COMMENT '结果说明',
			attribution VARCHAR(128) NOT NULL DEFAULT '' COMMENT '归属地',
			stop_reason INT NOT NULL DEFAULT 0 COMMENT '结束原因',
			customer_fail_reason VARCHAR(255) NOT NULL DEFAULT '' COMMENT '客户失败原因',
			customer_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '客户名称',
			customer_company VARCHAR(255) NOT NULL DEFAULT '' COMMENT '客户公司',
			group_names VARCHAR(255) NOT NULL DEFAULT '' COMMENT '分组名集合',
			seat_names VARCHAR(255) NOT NULL DEFAULT '' COMMENT '坐席名集合',
			seat_numbers VARCHAR(255) NOT NULL DEFAULT '' COMMENT '坐席分机集合',
			seat_work_numbers VARCHAR(255) NOT NULL DEFAULT '' COMMENT '坐席工号集合',
			enterprise_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '企业名',
			district_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '地区',
			service_device_number VARCHAR(64) NOT NULL DEFAULT '' COMMENT '设备号',
			call_answer_result INT NOT NULL DEFAULT 0 COMMENT '接通结果',
			call_hangup_party INT NOT NULL DEFAULT 0 COMMENT '挂断方',
			matched_user_id BIGINT NULL COMMENT '匹配本地用户 ID',
			matched_user_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '匹配本地用户名',
			role_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '角色名',
			remote_created_at DATETIME NULL COMMENT '远端创建时间',
			remote_updated_at DATETIME NULL COMMENT '远端更新时间',
			raw_payload LONGTEXT NULL COMMENT '原始 JSON',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
			UNIQUE KEY uk_mihua_call_recordings_cc_number (cc_number),
			KEY idx_mihua_call_recordings_initiate_time (initiate_time),
			KEY idx_mihua_call_recordings_seat_worknumber (service_seat_worknumber),
			KEY idx_mihua_call_recordings_outline_number (outline_number),
			KEY idx_mihua_call_recordings_matched_user_id (matched_user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='米话电销录音库'`,
	}
	return execStatements(tx, stmts)
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
			inside_sales_user_id BIGINT NULL,
			converted_at DATETIME NULL,
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
			FOREIGN KEY (owner_user_id) REFERENCES users(id),
			FOREIGN KEY (inside_sales_user_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS customer_owner_logs (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			customer_id BIGINT NOT NULL,
			from_owner_user_id BIGINT NULL,
			to_owner_user_id BIGINT NULL,
			action VARCHAR(32) NOT NULL,
			reason VARCHAR(64) NOT NULL DEFAULT '',
			content TEXT NULL,
			blocked_department_anchor_user_id BIGINT NULL,
			blocked_until DATETIME NULL,
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
		{"customers", "idx_customers_inside_sales_user_id", "inside_sales_user_id", false},
		{"customers", "idx_customers_converted_at", "converted_at", false},
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
		{"customers", "inside_sales_user_id", "BIGINT NULL"},
		{"customers", "converted_at", "DATETIME NULL"},
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
		{"customers", "idx_customers_inside_sales_user_id", "inside_sales_user_id", false},
		{"customers", "idx_customers_converted_at", "converted_at", false},
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

func upAddCustomerOwnerLogReasonAndContentMySQL(tx *gorm.DB) error {
	if err := addColumnIfNotExists(tx, "customer_owner_logs", "reason", "VARCHAR(64) NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	return addColumnIfNotExists(tx, "customer_owner_logs", "content", "TEXT NULL")
}

func upAddCustomerInsideSalesFieldsMySQL(tx *gorm.DB) error {
	if err := addColumnIfNotExists(tx, "customers", "inside_sales_user_id", "BIGINT NULL"); err != nil {
		return err
	}
	if err := addColumnIfNotExists(tx, "customers", "converted_at", "DATETIME NULL"); err != nil {
		return err
	}
	if err := addIndexIfNotExists(tx, "customers", "idx_customers_inside_sales_user_id", "inside_sales_user_id", false); err != nil {
		return err
	}
	return addIndexIfNotExists(tx, "customers", "idx_customers_converted_at", "converted_at", false)
}

func upAddCustomerOwnerLogClaimBlockFieldsMySQL(tx *gorm.DB) error {
	if err := addColumnIfNotExists(tx, "customer_owner_logs", "blocked_department_anchor_user_id", "BIGINT NULL"); err != nil {
		return err
	}
	if err := addColumnIfNotExists(tx, "customer_owner_logs", "blocked_until", "DATETIME NULL"); err != nil {
		return err
	}
	return addIndexIfNotExists(
		tx,
		"customer_owner_logs",
		"idx_customer_owner_logs_customer_blocked_until_anchor",
		"customer_id, blocked_until, blocked_department_anchor_user_id",
		false,
	)
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

func upCreateDailyUserCallStatsAndAddUsersHanghangCrmMobileMySQL(tx *gorm.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS daily_user_call_stats (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			stat_date DATE NOT NULL COMMENT '统计日期，每天每个用户一条',
			user_id BIGINT NULL COMMENT '匹配 users.id 后回填',
			real_name VARCHAR(64) NOT NULL DEFAULT '' COMMENT '真实姓名，用于匹配 users.nickname',
			mobile VARCHAR(32) NOT NULL DEFAULT '' COMMENT '航航CRM手机号，用于匹配 users.hanghang_crm_mobile',
			bind_num INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '绑定数量',
			call_num INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '拨打数量',
			not_connected INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '未接通数量',
			connection_rate DECIMAL(8,4) NOT NULL DEFAULT 0.0000 COMMENT '接通率',
			time_total INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '通话总时长原始值',
			total_minute VARCHAR(32) NOT NULL DEFAULT '' COMMENT '总时长文本',
			total_second INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '总时长秒数',
			average_call_duration DECIMAL(10,4) NOT NULL DEFAULT 0.0000 COMMENT '平均通话时长',
			average_call_second DECIMAL(10,4) NOT NULL DEFAULT 0.0000 COMMENT '平均通话秒数',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
			UNIQUE KEY uk_daily_user_call_stats_date_name_mobile (stat_date, real_name, mobile),
			KEY idx_daily_user_call_stats_stat_date (stat_date),
			KEY idx_daily_user_call_stats_user_id (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='每日用户电话统计表'`,
	}
	if err := execStatements(tx, stmts); err != nil {
		return err
	}

	if err := addColumnIfNotExists(tx, "users", "hanghang_crm_mobile", "VARCHAR(32) NOT NULL DEFAULT ''"); err != nil {
		return err
	}

	return addIndexIfNotExists(
		tx,
		"users",
		"idx_users_nickname_hanghang_crm_mobile",
		"nickname, hanghang_crm_mobile",
		false,
	)
}

func upCreateSalesDailyScoresMySQL(tx *gorm.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS sales_daily_scores (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			score_date DATE NOT NULL COMMENT '评分日期',
			user_id BIGINT NOT NULL COMMENT '销售ID',
			user_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '销售姓名',
			role_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '销售角色',
			call_num INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '通话量',
			call_duration_second INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '通话时长秒数',
			call_score_by_count INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '按通话量评分',
			call_score_by_duration INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '按通话时长评分',
			call_score_type VARCHAR(32) NOT NULL DEFAULT 'none' COMMENT '通话得分口径',
			call_score INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '通话得分',
			visit_count INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '上门拜访数',
			visit_score INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '上门拜访得分',
			new_customer_count INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '新增客户数',
			new_customer_score INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '新增客户得分',
			total_score INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '总分',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
			UNIQUE KEY uk_sales_daily_scores_date_user (score_date, user_id),
			KEY idx_sales_daily_scores_user_id (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='销售每日考核得分表'`,
	}
	return execStatements(tx, stmts)
}

func upAddSalesDailyScoreReachedAtMySQL(tx *gorm.DB) error {
	if !tx.Migrator().HasTable("sales_daily_scores") {
		return nil
	}

	if err := addColumnIfNotExists(
		tx,
		"sales_daily_scores",
		"score_reached_at",
		"DATETIME NULL COMMENT '达到当日最终分数的最早时间'",
	); err != nil {
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

func upCreateCustomerVisitsTableMySQL(tx *gorm.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS customer_visits (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			operator_user_id BIGINT NOT NULL,
			customer_name VARCHAR(255) NOT NULL,
			check_in_lat DOUBLE NOT NULL DEFAULT 0,
			check_in_lng DOUBLE NOT NULL DEFAULT 0,
			province VARCHAR(64) NOT NULL DEFAULT '',
			city VARCHAR(64) NOT NULL DEFAULT '',
			area VARCHAR(64) NOT NULL DEFAULT '',
			detail_address VARCHAR(1024) NOT NULL DEFAULT '',
			images TEXT NOT NULL,
			visit_purpose VARCHAR(500) NOT NULL DEFAULT '',
			remark TEXT NOT NULL,
			visit_date VARCHAR(10) NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
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
		{"customer_visits", "idx_customer_visits_operator_user_id", "operator_user_id", false},
		{"customer_visits", "idx_customer_visits_visit_date", "visit_date", false},
		{"customer_visits", "idx_customer_visits_created_at", "created_at", false},
	}
	for _, idx := range indexes {
		if err := addIndexIfNotExists(tx, idx.table, idx.name, idx.columns, idx.unique); err != nil {
			return err
		}
	}
	return nil
}

func upCreateUserHanghangCRMMobilesMySQL(tx *gorm.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS user_hanghang_crm_mobiles (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			user_id BIGINT NOT NULL,
			mobile VARCHAR(32) NOT NULL DEFAULT '',
			is_primary TINYINT(1) NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_user_hanghang_crm_mobiles_mobile (mobile),
			UNIQUE KEY uk_user_hanghang_crm_mobiles_user_mobile (user_id, mobile),
			KEY idx_user_hanghang_crm_mobiles_user_id (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户航航CRM手机号映射表'`,
		`INSERT INTO user_hanghang_crm_mobiles (user_id, mobile, is_primary, created_at, updated_at)
		 SELECT id, hanghang_crm_mobile, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		 FROM users
		 WHERE COALESCE(hanghang_crm_mobile, '') <> ''
		 ON DUPLICATE KEY UPDATE
			is_primary = VALUES(is_primary),
			updated_at = CURRENT_TIMESTAMP`,
	}
	return execStatements(tx, stmts)
}

func upAlterCustomerVisitsRegionColumnsToVarcharMySQL(tx *gorm.DB) error {
	if !tx.Migrator().HasTable("customer_visits") {
		return nil
	}

	stmts := []string{
		`ALTER TABLE customer_visits MODIFY COLUMN province VARCHAR(64) NOT NULL DEFAULT ''`,
		`ALTER TABLE customer_visits MODIFY COLUMN city VARCHAR(64) NOT NULL DEFAULT ''`,
		`ALTER TABLE customer_visits MODIFY COLUMN area VARCHAR(64) NOT NULL DEFAULT ''`,
		`UPDATE customer_visits SET province = '' WHERE province = '0'`,
		`UPDATE customer_visits SET city = '' WHERE city = '0'`,
		`UPDATE customer_visits SET area = '' WHERE area = '0'`,
	}
	return execStatements(tx, stmts)
}

func upAddCustomerVisitsCheckInIPMySQL(tx *gorm.DB) error {
	if !tx.Migrator().HasTable("customer_visits") {
		return nil
	}

	if err := addColumnIfNotExists(tx, "customer_visits", "check_in_ip", "VARCHAR(64) NOT NULL DEFAULT ''"); err != nil {
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

func upAddCustomerVisitsInviterMySQL(tx *gorm.DB) error {
	if !tx.Migrator().HasTable("customer_visits") {
		return nil
	}

	return addColumnIfNotExists(tx, "customer_visits", "inviter", "VARCHAR(255) NOT NULL DEFAULT ''")
}

func upAddCustomerAssignTimeAndSalesAssignDropSettingMySQL(tx *gorm.DB) error {
	if err := addColumnIfNotExists(tx, "customers", "assign_time", "BIGINT NULL"); err != nil {
		return err
	}
	if err := addIndexIfNotExists(tx, "customers", "idx_customers_assign_time", "assign_time", false); err != nil {
		return err
	}

	stmts := []string{
		`UPDATE customers
		SET assign_time = COALESCE(
			CASE
				WHEN converted_at IS NOT NULL THEN UNIX_TIMESTAMP(converted_at)
				ELSE NULL
			END,
			collect_time
		)
		WHERE COALESCE(assign_time, 0) = 0
			AND inside_sales_user_id IS NOT NULL
			AND owner_user_id IS NOT NULL
			AND owner_user_id <> inside_sales_user_id
			AND status <> 'pool'`,
		`INSERT IGNORE INTO system_settings(` + "`key`" + `, value, description) VALUES ('sales_assign_deal_drop_days','30','电销分配给销售后多少天未签单自动掉库')`,
	}
	return execStatements(tx, stmts)
}

func upCreateCallRecordingsMySQL(tx *gorm.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS call_recordings (
			id VARCHAR(64) PRIMARY KEY COMMENT '航航CRM录音ID',
			agent_code BIGINT NOT NULL DEFAULT 0 COMMENT '坐席编码',
			call_status INT NOT NULL DEFAULT 0 COMMENT '通话状态',
			call_status_name VARCHAR(64) NOT NULL DEFAULT '' COMMENT '通话状态名称',
			call_type INT NOT NULL DEFAULT 0 COMMENT '通话类型',
			callee_attr VARCHAR(128) NOT NULL DEFAULT '' COMMENT '被叫归属地',
			caller_attr VARCHAR(128) NOT NULL DEFAULT '' COMMENT '主叫归属地',
			create_time BIGINT NOT NULL DEFAULT 0 COMMENT '远端创建时间毫秒',
			dept_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '部门名称',
			duration INT NOT NULL DEFAULT 0 COMMENT '通话时长秒数',
			end_time BIGINT NOT NULL DEFAULT 0 COMMENT '结束时间毫秒',
			enterprise_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '企业名称',
			finish_status INT NOT NULL DEFAULT 0 COMMENT '完成状态',
			finish_status_name VARCHAR(64) NOT NULL DEFAULT '' COMMENT '完成状态名称',
			handle INT NOT NULL DEFAULT 0 COMMENT '处理标记',
			interface_id VARCHAR(64) NOT NULL DEFAULT '' COMMENT '接口ID',
			interface_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '接口名称',
			line_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '线路名称',
			mobile VARCHAR(32) NOT NULL DEFAULT '' COMMENT '坐席手机号',
			mode INT NOT NULL DEFAULT 0 COMMENT '模式',
			move_batch_code VARCHAR(128) NULL COMMENT '批次编码',
			oct_customer_id VARCHAR(128) NULL COMMENT '外部客户ID',
			phone VARCHAR(32) NOT NULL DEFAULT '' COMMENT '客户电话',
			postage DECIMAL(10,4) NOT NULL DEFAULT 0.0000 COMMENT '通话资费',
			pre_record_url TEXT NULL COMMENT '录音地址',
			real_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '坐席名称',
			start_time BIGINT NOT NULL DEFAULT 0 COMMENT '开始时间毫秒',
			status INT NOT NULL DEFAULT 0 COMMENT '状态',
			tel_a VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'A号码',
			tel_b VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'B号码',
			tel_x VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'X号码',
			tenant_code VARCHAR(64) NOT NULL DEFAULT '' COMMENT '租户编码',
			update_time BIGINT NOT NULL DEFAULT 0 COMMENT '远端更新时间毫秒',
			user_id VARCHAR(64) NOT NULL DEFAULT '' COMMENT '远端用户ID',
			work_num VARCHAR(128) NULL COMMENT '工号',
			dedupe_key VARCHAR(191) NOT NULL DEFAULT '' COMMENT '业务去重键',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
			KEY idx_call_recordings_mobile (mobile),
			KEY idx_call_recordings_phone (phone),
			KEY idx_call_recordings_start_time (start_time),
			KEY idx_call_recordings_create_time (create_time),
			KEY idx_call_recordings_user_id (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通话录音表'`,
	}
	return execStatements(tx, stmts)
}

func upDedupeCallRecordingsMySQL(tx *gorm.DB) error {
	if !tx.Migrator().HasTable("call_recordings") {
		return nil
	}

	if err := addColumnIfNotExists(tx, "call_recordings", "dedupe_key", "VARCHAR(191) NOT NULL DEFAULT '' COMMENT '业务去重键'"); err != nil {
		return err
	}

	stmts := []string{
		`UPDATE call_recordings
		SET dedupe_key = CONCAT(
			COALESCE(CAST(start_time AS CHAR), ''), '|',
			COALESCE(mobile, ''), '|',
			COALESCE(phone, ''), '|',
			COALESCE(tel_a, ''), '|',
			COALESCE(tel_b, ''), '|',
			COALESCE(CAST(call_type AS CHAR), ''), '|',
			COALESCE(CAST(duration AS CHAR), '')
		)`,
		`DELETE cr1 FROM call_recordings cr1
		INNER JOIN call_recordings cr2
			ON cr1.dedupe_key = cr2.dedupe_key
			AND cr1.id < cr2.id`,
	}
	if err := execStatements(tx, stmts); err != nil {
		return err
	}

	return addIndexIfNotExists(tx, "call_recordings", "uk_call_recordings_dedupe_key", "dedupe_key", true)
}
