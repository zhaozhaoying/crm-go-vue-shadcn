package repository

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"backend/internal/model"
	"gorm.io/gorm"
)

type FollowRecordRepository struct {
	db *gorm.DB
}

type operationFollowRecordRow struct {
	ID              int64     `gorm:"column:id;primaryKey;autoIncrement"`
	CustomerID      int64     `gorm:"column:customer_id"`
	Content         string    `gorm:"column:content"`
	NextFollowTime  int64     `gorm:"column:next_follow_time"`
	AppointmentTime *int64    `gorm:"column:appointment_time"`
	ShootingTime    *int64    `gorm:"column:shooting_time"`
	CustomerLevelID int       `gorm:"column:customer_level_id"`
	FollowMethodID  int       `gorm:"column:follow_method_id"`
	OperatorUserID  int64     `gorm:"column:operator_user_id"`
	CreatedAt       time.Time `gorm:"column:created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at"`
}

type salesFollowRecordRow struct {
	ID               int64     `gorm:"column:id;primaryKey;autoIncrement"`
	CustomerID       int64     `gorm:"column:customer_id"`
	Content          string    `gorm:"column:content"`
	NextFollowTime   int64     `gorm:"column:next_follow_time"`
	CustomerLevelID  int       `gorm:"column:customer_level_id"`
	CustomerSourceID int       `gorm:"column:customer_source_id"`
	FollowMethodID   int       `gorm:"column:follow_method_id"`
	OperatorUserID   int64     `gorm:"column:operator_user_id"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
}

type followMethodRow struct {
	ID        int       `gorm:"column:id;primaryKey;autoIncrement"`
	Name      string    `gorm:"column:name"`
	Sort      int       `gorm:"column:sort"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func NewGormFollowRecordRepository(db *gorm.DB) *FollowRecordRepository {
	return &FollowRecordRepository{db: db}
}

func NewFollowRecordRepository(db *gorm.DB) *FollowRecordRepository {
	return NewGormFollowRecordRepository(db)
}

// CreateOperationFollowRecord 创建运营跟进记录
func (r *FollowRecordRepository) CreateOperationFollowRecord(input model.FollowRecordCreateInput) (int64, error) {
	nowUnix := time.Now().Unix()
	now := time.Now().UTC()

	nextFollowTime := nowUnix
	if input.NextFollowTime != nil {
		nextFollowTime = input.NextFollowTime.Unix()
	}

	var appointmentTime *int64
	if input.AppointmentTime != nil {
		t := input.AppointmentTime.Unix()
		appointmentTime = &t
	}

	var shootingTime *int64
	if input.ShootingTime != nil {
		t := input.ShootingTime.Unix()
		shootingTime = &t
	}

	row := operationFollowRecordRow{}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		customerLevelID := input.CustomerLevelID
		var err error
		if customerLevelID == 0 {
			customerLevelID, err = r.getCustomerLevelIDByCustomerID(tx, input.CustomerID)
			if err != nil {
				return err
			}
		}

		row = operationFollowRecordRow{
			CustomerID:      input.CustomerID,
			Content:         input.Content,
			NextFollowTime:  nextFollowTime,
			AppointmentTime: appointmentTime,
			ShootingTime:    shootingTime,
			CustomerLevelID: customerLevelID,
			FollowMethodID:  input.FollowMethodID,
			OperatorUserID:  input.OperatorUserID,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		return tx.Table("operation_follow_records").Create(&row).Error
	})
	if err != nil {
		return 0, err
	}

	return row.ID, nil
}

func (r *FollowRecordRepository) getCustomerLevelIDByCustomerID(tx *gorm.DB, customerID int64) (int, error) {
	var levelID int
	result := tx.Table("customers").
		Select("COALESCE(customer_level_id, 0)").
		Where("id = ?", customerID).
		Scan(&levelID)
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected == 0 {
		return 0, nil
	}
	return levelID, nil
}

// CreateSalesFollowRecord 创建销售跟进记录
func (r *FollowRecordRepository) CreateSalesFollowRecord(input model.FollowRecordCreateInput) (int64, error) {
	nowUnix := time.Now().Unix()
	now := time.Now().UTC()

	nextFollowTime := nowUnix
	if input.NextFollowTime != nil {
		nextFollowTime = input.NextFollowTime.Unix()
	}

	row := salesFollowRecordRow{}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		row = salesFollowRecordRow{
			CustomerID:       input.CustomerID,
			Content:          input.Content,
			NextFollowTime:   nextFollowTime,
			CustomerLevelID:  input.CustomerLevelID,
			CustomerSourceID: input.CustomerSourceID,
			FollowMethodID:   input.FollowMethodID,
			OperatorUserID:   input.OperatorUserID,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if err := tx.Table("sales_follow_records").Create(&row).Error; err != nil {
			return err
		}
		return r.syncCustomerFollowSnapshot(tx, input)
	})
	if err != nil {
		return 0, err
	}

	return row.ID, nil
}

func (r *FollowRecordRepository) syncCustomerFollowSnapshot(tx *gorm.DB, input model.FollowRecordCreateInput) error {
	nowUnix := time.Now().Unix()
	now := time.Now().UTC()

	updates := map[string]interface{}{
		"follow_time":     nowUnix,
		"next_time":       nowUnix,
		"operate_user_id": input.OperatorUserID,
		"updated_at":      now,
	}
	if input.NextFollowTime != nil {
		updates["next_time"] = input.NextFollowTime.Unix()
	}
	if input.CustomerLevelID != 0 {
		updates["customer_level_id"] = input.CustomerLevelID
	}
	if input.CustomerSourceID != 0 {
		updates["customer_source_id"] = input.CustomerSourceID
	}

	return tx.Table("customers").Where("id = ?", input.CustomerID).Updates(updates).Error
}

// ListOperationFollowRecords 获取运营跟进记录列表（按客户ID）
func (r *FollowRecordRepository) ListOperationFollowRecords(filter model.FollowRecordListFilter) (model.OperationFollowRecordListResult, error) {
	var result model.OperationFollowRecordListResult
	result.Page = filter.Page
	result.PageSize = filter.PageSize

	if err := r.db.Table("operation_follow_records").Where("customer_id = ?", filter.CustomerID).Count(&result.Total).Error; err != nil {
		return result, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	rows, err := r.db.Table("operation_follow_records AS ofr").
		Select(
			"ofr.id", "ofr.customer_id",
			"COALESCE(c.name, '')", "COALESCE(c.legal_name, '')", "COALESCE(c.contact_name, '')",
			"COALESCE(c.weixin, '')", "COALESCE(c.email, '')",
			"COALESCE((SELECT cp.phone FROM customer_phones cp WHERE cp.customer_id = c.id ORDER BY cp.is_primary DESC, cp.id ASC LIMIT 1), '')",
			"COALESCE(c.province, 0)", "COALESCE(c.city, 0)", "COALESCE(c.area, 0)",
			"COALESCE(c.detail_address, '')", "COALESCE(c.remark, '')",
			"COALESCE(c.status, '')", "COALESCE(c.deal_status, '')",
			"COALESCE(c.owner_user_id, 0)", "COALESCE(owner_u.nickname, '')",
			"NULLIF(c.next_time, 0)", "NULLIF(c.follow_time, 0)", "NULLIF(c.collect_time, 0)",
			"COALESCE(c.customer_source_id, 0)", "COALESCE(cs.name, '')",
			"COALESCE(c.customer_level_id, 0)", "COALESCE(cl_customer.name, '')",
			"ofr.content", "ofr.next_follow_time", "ofr.appointment_time", "ofr.shooting_time",
			"COALESCE(NULLIF(ofr.customer_level_id, 0), c.customer_level_id, 0)", "COALESCE(cl_record.name, cl_customer.name, '')",
			"ofr.follow_method_id", "COALESCE(fm.name, '')",
			"ofr.operator_user_id", "COALESCE(u.nickname, '')", "ofr.created_at", "ofr.updated_at",
		).
		Joins("LEFT JOIN users u ON ofr.operator_user_id = u.id").
		Joins("LEFT JOIN customers c ON ofr.customer_id = c.id").
		Joins("LEFT JOIN users owner_u ON c.owner_user_id = owner_u.id").
		Joins("LEFT JOIN customer_levels cl_record ON ofr.customer_level_id = cl_record.id").
		Joins("LEFT JOIN customer_levels cl_customer ON c.customer_level_id = cl_customer.id").
		Joins("LEFT JOIN customer_sources cs ON c.customer_source_id = cs.id").
		Joins("LEFT JOIN follow_methods fm ON ofr.follow_method_id = fm.id").
		Where("ofr.customer_id = ?", filter.CustomerID).
		Order("ofr.created_at DESC").
		Limit(filter.PageSize).
		Offset(offset).
		Rows()
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var record model.OperationFollowRecord
		var customerName, customerLegalName, customerContactName string
		var customerWeixin, customerEmail, customerPrimaryPhone string
		var customerProvince, customerCity, customerArea int
		var customerDetailAddress, customerRemark string
		var customerStatus, customerDealStatus string
		var customerOwnerUserIDRaw int64
		var customerOwnerUserName string
		var customerNextTimeRaw, customerFollowTimeRaw, customerCollectTimeRaw interface{}
		var customerSourceIDRaw, customerLevelIDRaw int
		var customerSourceNameRaw, customerLevelNameRaw string
		var nextFollowTimeRaw, appointmentTimeRaw, shootingTimeRaw interface{}
		var recordCustomerLevelID int
		var recordCustomerLevelName string
		var createdAtRaw, updatedAtRaw interface{}

		err := rows.Scan(
			&record.ID, &record.CustomerID,
			&customerName, &customerLegalName, &customerContactName,
			&customerWeixin, &customerEmail, &customerPrimaryPhone,
			&customerProvince, &customerCity, &customerArea,
			&customerDetailAddress, &customerRemark,
			&customerStatus, &customerDealStatus,
			&customerOwnerUserIDRaw, &customerOwnerUserName,
			&customerNextTimeRaw, &customerFollowTimeRaw, &customerCollectTimeRaw,
			&customerSourceIDRaw, &customerSourceNameRaw, &customerLevelIDRaw, &customerLevelNameRaw,
			&record.Content, &nextFollowTimeRaw, &appointmentTimeRaw, &shootingTimeRaw,
			&recordCustomerLevelID, &recordCustomerLevelName,
			&record.FollowMethodID, &record.FollowMethodName,
			&record.OperatorUserID, &record.OperatorUserName, &createdAtRaw, &updatedAtRaw,
		)
		if err != nil {
			return result, err
		}

		nextFollowTime, err := parseNullableDBTime(nextFollowTimeRaw)
		if err != nil {
			return result, err
		}
		record.NextFollowTime = nextFollowTime

		customerNextTime, err := parseNullableDBTime(customerNextTimeRaw)
		if err != nil {
			return result, err
		}

		customerFollowTime, err := parseNullableDBTime(customerFollowTimeRaw)
		if err != nil {
			return result, err
		}

		customerCollectTime, err := parseNullableDBTime(customerCollectTimeRaw)
		if err != nil {
			return result, err
		}

		appointmentTime, err := parseNullableDBTime(appointmentTimeRaw)
		if err != nil {
			return result, err
		}
		record.AppointmentTime = appointmentTime

		shootingTime, err := parseNullableDBTime(shootingTimeRaw)
		if err != nil {
			return result, err
		}
		record.ShootingTime = shootingTime

		record.CreatedAt, err = parseDBTime(createdAtRaw)
		if err != nil {
			return result, err
		}
		record.UpdatedAt, err = parseDBTime(updatedAtRaw)
		if err != nil {
			return result, err
		}

		record.CustomerLevelID = recordCustomerLevelID
		record.CustomerLevelName = recordCustomerLevelName
		record.CustomerSourceID = customerSourceIDRaw
		record.CustomerSourceName = customerSourceNameRaw

		customer := &model.OperationFollowRecordCustomer{
			ID:            record.CustomerID,
			Name:          customerName,
			LegalName:     customerLegalName,
			ContactName:   customerContactName,
			Weixin:        customerWeixin,
			Email:         customerEmail,
			PrimaryPhone:  customerPrimaryPhone,
			Province:      customerProvince,
			City:          customerCity,
			Area:          customerArea,
			DetailAddress: customerDetailAddress,
			Remark:        customerRemark,
			Status:        customerStatus,
			DealStatus:    customerDealStatus,
			OwnerUserName: customerOwnerUserName,
			NextTime:      customerNextTime,
			FollowTime:    customerFollowTime,
			CollectTime:   customerCollectTime,
			LevelID:       customerLevelIDRaw,
			LevelName:     customerLevelNameRaw,
			SourceID:      customerSourceIDRaw,
			SourceName:    customerSourceNameRaw,
		}
		if customerOwnerUserIDRaw > 0 {
			ownerID := customerOwnerUserIDRaw
			customer.OwnerUserID = &ownerID
		}
		record.Customer = customer

		result.Items = append(result.Items, record)
	}

	if result.Items == nil {
		result.Items = []model.OperationFollowRecord{}
	}

	return result, nil
}

// ListAllOperationFollowRecords 获取所有运营跟进记录列表
func (r *FollowRecordRepository) ListAllOperationFollowRecords(page, pageSize int) (model.OperationFollowRecordListResult, error) {
	var result model.OperationFollowRecordListResult
	result.Page = page
	result.PageSize = pageSize

	if err := r.db.Table("operation_follow_records").Count(&result.Total).Error; err != nil {
		return result, err
	}

	offset := (page - 1) * pageSize
	rows, err := r.db.Table("operation_follow_records AS ofr").
		Select(
			"ofr.id", "ofr.customer_id",
			"COALESCE(c.name, '')", "COALESCE(c.legal_name, '')", "COALESCE(c.contact_name, '')",
			"COALESCE(c.weixin, '')", "COALESCE(c.email, '')",
			"COALESCE((SELECT cp.phone FROM customer_phones cp WHERE cp.customer_id = c.id ORDER BY cp.is_primary DESC, cp.id ASC LIMIT 1), '')",
			"COALESCE(c.province, 0)", "COALESCE(c.city, 0)", "COALESCE(c.area, 0)",
			"COALESCE(c.detail_address, '')", "COALESCE(c.remark, '')",
			"COALESCE(c.status, '')", "COALESCE(c.deal_status, '')",
			"COALESCE(c.owner_user_id, 0)", "COALESCE(owner_u.nickname, '')",
			"NULLIF(c.next_time, 0)", "NULLIF(c.follow_time, 0)", "NULLIF(c.collect_time, 0)",
			"COALESCE(c.customer_source_id, 0)", "COALESCE(cs.name, '')",
			"COALESCE(c.customer_level_id, 0)", "COALESCE(cl_customer.name, '')",
			"ofr.content", "ofr.next_follow_time", "ofr.appointment_time", "ofr.shooting_time",
			"COALESCE(NULLIF(ofr.customer_level_id, 0), c.customer_level_id, 0)", "COALESCE(cl_record.name, cl_customer.name, '')",
			"ofr.follow_method_id", "COALESCE(fm.name, '')",
			"ofr.operator_user_id", "COALESCE(u.nickname, '')", "ofr.created_at", "ofr.updated_at",
		).
		Joins("LEFT JOIN users u ON ofr.operator_user_id = u.id").
		Joins("LEFT JOIN customers c ON ofr.customer_id = c.id").
		Joins("LEFT JOIN users owner_u ON c.owner_user_id = owner_u.id").
		Joins("LEFT JOIN customer_levels cl_record ON ofr.customer_level_id = cl_record.id").
		Joins("LEFT JOIN customer_levels cl_customer ON c.customer_level_id = cl_customer.id").
		Joins("LEFT JOIN customer_sources cs ON c.customer_source_id = cs.id").
		Joins("LEFT JOIN follow_methods fm ON ofr.follow_method_id = fm.id").
		Order("ofr.created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Rows()
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var record model.OperationFollowRecord
		var customerName, customerLegalName, customerContactName string
		var customerWeixin, customerEmail, customerPrimaryPhone string
		var customerProvince, customerCity, customerArea int
		var customerDetailAddress, customerRemark string
		var customerStatus, customerDealStatus string
		var customerOwnerUserIDRaw int64
		var customerOwnerUserName string
		var customerNextTimeRaw, customerFollowTimeRaw, customerCollectTimeRaw interface{}
		var customerSourceIDRaw, customerLevelIDRaw int
		var customerSourceNameRaw, customerLevelNameRaw string
		var nextFollowTimeRaw, appointmentTimeRaw, shootingTimeRaw interface{}
		var recordCustomerLevelID int
		var recordCustomerLevelName string
		var createdAtRaw, updatedAtRaw interface{}

		err := rows.Scan(
			&record.ID, &record.CustomerID,
			&customerName, &customerLegalName, &customerContactName,
			&customerWeixin, &customerEmail, &customerPrimaryPhone,
			&customerProvince, &customerCity, &customerArea,
			&customerDetailAddress, &customerRemark,
			&customerStatus, &customerDealStatus,
			&customerOwnerUserIDRaw, &customerOwnerUserName,
			&customerNextTimeRaw, &customerFollowTimeRaw, &customerCollectTimeRaw,
			&customerSourceIDRaw, &customerSourceNameRaw, &customerLevelIDRaw, &customerLevelNameRaw,
			&record.Content, &nextFollowTimeRaw, &appointmentTimeRaw, &shootingTimeRaw,
			&recordCustomerLevelID, &recordCustomerLevelName,
			&record.FollowMethodID, &record.FollowMethodName,
			&record.OperatorUserID, &record.OperatorUserName, &createdAtRaw, &updatedAtRaw,
		)
		if err != nil {
			return result, err
		}

		nextFollowTime, err := parseNullableDBTime(nextFollowTimeRaw)
		if err != nil {
			return result, err
		}
		record.NextFollowTime = nextFollowTime

		customerNextTime, err := parseNullableDBTime(customerNextTimeRaw)
		if err != nil {
			return result, err
		}

		customerFollowTime, err := parseNullableDBTime(customerFollowTimeRaw)
		if err != nil {
			return result, err
		}

		customerCollectTime, err := parseNullableDBTime(customerCollectTimeRaw)
		if err != nil {
			return result, err
		}

		appointmentTime, err := parseNullableDBTime(appointmentTimeRaw)
		if err != nil {
			return result, err
		}
		record.AppointmentTime = appointmentTime

		shootingTime, err := parseNullableDBTime(shootingTimeRaw)
		if err != nil {
			return result, err
		}
		record.ShootingTime = shootingTime

		record.CreatedAt, err = parseDBTime(createdAtRaw)
		if err != nil {
			return result, err
		}
		record.UpdatedAt, err = parseDBTime(updatedAtRaw)
		if err != nil {
			return result, err
		}

		record.CustomerLevelID = recordCustomerLevelID
		record.CustomerLevelName = recordCustomerLevelName
		record.CustomerSourceID = customerSourceIDRaw
		record.CustomerSourceName = customerSourceNameRaw

		customer := &model.OperationFollowRecordCustomer{
			ID:            record.CustomerID,
			Name:          customerName,
			LegalName:     customerLegalName,
			ContactName:   customerContactName,
			Weixin:        customerWeixin,
			Email:         customerEmail,
			PrimaryPhone:  customerPrimaryPhone,
			Province:      customerProvince,
			City:          customerCity,
			Area:          customerArea,
			DetailAddress: customerDetailAddress,
			Remark:        customerRemark,
			Status:        customerStatus,
			DealStatus:    customerDealStatus,
			OwnerUserName: customerOwnerUserName,
			NextTime:      customerNextTime,
			FollowTime:    customerFollowTime,
			CollectTime:   customerCollectTime,
			LevelID:       customerLevelIDRaw,
			LevelName:     customerLevelNameRaw,
			SourceID:      customerSourceIDRaw,
			SourceName:    customerSourceNameRaw,
		}
		if customerOwnerUserIDRaw > 0 {
			ownerID := customerOwnerUserIDRaw
			customer.OwnerUserID = &ownerID
		}
		record.Customer = customer

		result.Items = append(result.Items, record)
	}

	if result.Items == nil {
		result.Items = []model.OperationFollowRecord{}
	}

	return result, nil
}

// ListSalesFollowRecords 获取销售跟进记录列表（按客户ID）
func (r *FollowRecordRepository) ListSalesFollowRecords(filter model.FollowRecordListFilter) (model.SalesFollowRecordListResult, error) {
	var result model.SalesFollowRecordListResult
	result.Page = filter.Page
	result.PageSize = filter.PageSize

	if err := r.db.Table("sales_follow_records").Where("customer_id = ?", filter.CustomerID).Count(&result.Total).Error; err != nil {
		return result, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	rows, err := r.db.Table("sales_follow_records AS sfr").
		Select(
			"sfr.id", "sfr.customer_id",
			"COALESCE(c.name, '')", "COALESCE(c.legal_name, '')", "COALESCE(c.contact_name, '')",
			"COALESCE(c.weixin, '')", "COALESCE(c.email, '')",
			"COALESCE((SELECT cp.phone FROM customer_phones cp WHERE cp.customer_id = c.id ORDER BY cp.is_primary DESC, cp.id ASC LIMIT 1), '')",
			"COALESCE(c.province, 0)", "COALESCE(c.city, 0)", "COALESCE(c.area, 0)",
			"COALESCE(c.detail_address, '')", "COALESCE(c.remark, '')",
			"COALESCE(c.status, '')", "COALESCE(c.deal_status, '')",
			"COALESCE(c.owner_user_id, 0)", "COALESCE(owner_u.nickname, '')",
			"NULLIF(c.next_time, 0)", "NULLIF(c.follow_time, 0)", "NULLIF(c.collect_time, 0)",
			"COALESCE(c.customer_source_id, 0)", "COALESCE(cs_customer.name, '')",
			"COALESCE(c.customer_level_id, 0)", "COALESCE(cl_customer.name, '')",
			"sfr.content", "sfr.next_follow_time",
			"COALESCE(NULLIF(sfr.customer_level_id, 0), c.customer_level_id, 0)", "COALESCE(cl_record.name, cl_customer.name, '')",
			"COALESCE(NULLIF(sfr.customer_source_id, 0), c.customer_source_id, 0)", "COALESCE(cs_record.name, cs_customer.name, '')",
			"sfr.follow_method_id", "COALESCE(fm.name, '')",
			"sfr.operator_user_id", "COALESCE(u.nickname, '')", "sfr.created_at", "sfr.updated_at",
		).
		Joins("LEFT JOIN users u ON sfr.operator_user_id = u.id").
		Joins("LEFT JOIN customers c ON sfr.customer_id = c.id").
		Joins("LEFT JOIN users owner_u ON c.owner_user_id = owner_u.id").
		Joins("LEFT JOIN customer_levels cl_record ON sfr.customer_level_id = cl_record.id").
		Joins("LEFT JOIN customer_levels cl_customer ON c.customer_level_id = cl_customer.id").
		Joins("LEFT JOIN customer_sources cs_record ON sfr.customer_source_id = cs_record.id").
		Joins("LEFT JOIN customer_sources cs_customer ON c.customer_source_id = cs_customer.id").
		Joins("LEFT JOIN follow_methods fm ON sfr.follow_method_id = fm.id").
		Where("sfr.customer_id = ?", filter.CustomerID).
		Order("sfr.created_at DESC").
		Limit(filter.PageSize).
		Offset(offset).
		Rows()
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var record model.SalesFollowRecord
		var customerName, customerLegalName, customerContactName string
		var customerWeixin, customerEmail, customerPrimaryPhone string
		var customerProvince, customerCity, customerArea int
		var customerDetailAddress, customerRemark string
		var customerStatus, customerDealStatus string
		var customerOwnerUserIDRaw int64
		var customerOwnerUserName string
		var customerNextTimeRaw, customerFollowTimeRaw, customerCollectTimeRaw interface{}
		var customerSourceIDRaw, customerLevelIDRaw int
		var customerSourceNameRaw, customerLevelNameRaw string
		var nextFollowTimeRaw interface{}
		var recordCustomerLevelID, recordCustomerSourceID int
		var recordCustomerLevelName, recordCustomerSourceName string
		var createdAtRaw, updatedAtRaw interface{}

		err := rows.Scan(
			&record.ID, &record.CustomerID,
			&customerName, &customerLegalName, &customerContactName,
			&customerWeixin, &customerEmail, &customerPrimaryPhone,
			&customerProvince, &customerCity, &customerArea,
			&customerDetailAddress, &customerRemark,
			&customerStatus, &customerDealStatus,
			&customerOwnerUserIDRaw, &customerOwnerUserName,
			&customerNextTimeRaw, &customerFollowTimeRaw, &customerCollectTimeRaw,
			&customerSourceIDRaw, &customerSourceNameRaw, &customerLevelIDRaw, &customerLevelNameRaw,
			&record.Content, &nextFollowTimeRaw,
			&recordCustomerLevelID, &recordCustomerLevelName,
			&recordCustomerSourceID, &recordCustomerSourceName,
			&record.FollowMethodID, &record.FollowMethodName, &record.OperatorUserID, &record.OperatorUserName,
			&createdAtRaw, &updatedAtRaw,
		)
		if err != nil {
			return result, err
		}

		nextFollowTime, err := parseNullableDBTime(nextFollowTimeRaw)
		if err != nil {
			return result, err
		}
		record.NextFollowTime = nextFollowTime

		customerNextTime, err := parseNullableDBTime(customerNextTimeRaw)
		if err != nil {
			return result, err
		}

		customerFollowTime, err := parseNullableDBTime(customerFollowTimeRaw)
		if err != nil {
			return result, err
		}

		customerCollectTime, err := parseNullableDBTime(customerCollectTimeRaw)
		if err != nil {
			return result, err
		}

		record.CreatedAt, err = parseDBTime(createdAtRaw)
		if err != nil {
			return result, err
		}
		record.UpdatedAt, err = parseDBTime(updatedAtRaw)
		if err != nil {
			return result, err
		}

		record.CustomerLevelID = recordCustomerLevelID
		record.CustomerLevelName = recordCustomerLevelName
		record.CustomerSourceID = recordCustomerSourceID
		record.CustomerSourceName = recordCustomerSourceName

		customer := &model.OperationFollowRecordCustomer{
			ID:            record.CustomerID,
			Name:          customerName,
			LegalName:     customerLegalName,
			ContactName:   customerContactName,
			Weixin:        customerWeixin,
			Email:         customerEmail,
			PrimaryPhone:  customerPrimaryPhone,
			Province:      customerProvince,
			City:          customerCity,
			Area:          customerArea,
			DetailAddress: customerDetailAddress,
			Remark:        customerRemark,
			Status:        customerStatus,
			DealStatus:    customerDealStatus,
			OwnerUserName: customerOwnerUserName,
			NextTime:      customerNextTime,
			FollowTime:    customerFollowTime,
			CollectTime:   customerCollectTime,
			LevelID:       customerLevelIDRaw,
			LevelName:     customerLevelNameRaw,
			SourceID:      customerSourceIDRaw,
			SourceName:    customerSourceNameRaw,
		}
		if customerOwnerUserIDRaw > 0 {
			ownerID := customerOwnerUserIDRaw
			customer.OwnerUserID = &ownerID
		}
		record.Customer = customer

		result.Items = append(result.Items, record)
	}

	if result.Items == nil {
		result.Items = []model.SalesFollowRecord{}
	}

	return result, nil
}

// ListAllSalesFollowRecords 获取所有销售跟进记录列表
func (r *FollowRecordRepository) ListAllSalesFollowRecords(page, pageSize int) (model.SalesFollowRecordListResult, error) {
	var result model.SalesFollowRecordListResult
	result.Page = page
	result.PageSize = pageSize

	if err := r.db.Table("sales_follow_records").Count(&result.Total).Error; err != nil {
		return result, err
	}

	offset := (page - 1) * pageSize
	rows, err := r.db.Table("sales_follow_records AS sfr").
		Select(
			"sfr.id", "sfr.customer_id",
			"COALESCE(c.name, '')", "COALESCE(c.legal_name, '')", "COALESCE(c.contact_name, '')",
			"COALESCE(c.weixin, '')", "COALESCE(c.email, '')",
			"COALESCE((SELECT cp.phone FROM customer_phones cp WHERE cp.customer_id = c.id ORDER BY cp.is_primary DESC, cp.id ASC LIMIT 1), '')",
			"COALESCE(c.province, 0)", "COALESCE(c.city, 0)", "COALESCE(c.area, 0)",
			"COALESCE(c.detail_address, '')", "COALESCE(c.remark, '')",
			"COALESCE(c.status, '')", "COALESCE(c.deal_status, '')",
			"COALESCE(c.owner_user_id, 0)", "COALESCE(owner_u.nickname, '')",
			"NULLIF(c.next_time, 0)", "NULLIF(c.follow_time, 0)", "NULLIF(c.collect_time, 0)",
			"COALESCE(c.customer_source_id, 0)", "COALESCE(cs_customer.name, '')",
			"COALESCE(c.customer_level_id, 0)", "COALESCE(cl_customer.name, '')",
			"sfr.content", "sfr.next_follow_time",
			"COALESCE(NULLIF(sfr.customer_level_id, 0), c.customer_level_id, 0)", "COALESCE(cl_record.name, cl_customer.name, '')",
			"COALESCE(NULLIF(sfr.customer_source_id, 0), c.customer_source_id, 0)", "COALESCE(cs_record.name, cs_customer.name, '')",
			"sfr.follow_method_id", "COALESCE(fm.name, '')",
			"sfr.operator_user_id", "COALESCE(u.nickname, '')", "sfr.created_at", "sfr.updated_at",
		).
		Joins("LEFT JOIN users u ON sfr.operator_user_id = u.id").
		Joins("LEFT JOIN customers c ON sfr.customer_id = c.id").
		Joins("LEFT JOIN users owner_u ON c.owner_user_id = owner_u.id").
		Joins("LEFT JOIN customer_levels cl_record ON sfr.customer_level_id = cl_record.id").
		Joins("LEFT JOIN customer_levels cl_customer ON c.customer_level_id = cl_customer.id").
		Joins("LEFT JOIN customer_sources cs_record ON sfr.customer_source_id = cs_record.id").
		Joins("LEFT JOIN customer_sources cs_customer ON c.customer_source_id = cs_customer.id").
		Joins("LEFT JOIN follow_methods fm ON sfr.follow_method_id = fm.id").
		Order("sfr.created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Rows()
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var record model.SalesFollowRecord
		var customerName, customerLegalName, customerContactName string
		var customerWeixin, customerEmail, customerPrimaryPhone string
		var customerProvince, customerCity, customerArea int
		var customerDetailAddress, customerRemark string
		var customerStatus, customerDealStatus string
		var customerOwnerUserIDRaw int64
		var customerOwnerUserName string
		var customerNextTimeRaw, customerFollowTimeRaw, customerCollectTimeRaw interface{}
		var customerSourceIDRaw, customerLevelIDRaw int
		var customerSourceNameRaw, customerLevelNameRaw string
		var nextFollowTimeRaw interface{}
		var recordCustomerLevelID, recordCustomerSourceID int
		var recordCustomerLevelName, recordCustomerSourceName string
		var createdAtRaw, updatedAtRaw interface{}

		err := rows.Scan(
			&record.ID, &record.CustomerID,
			&customerName, &customerLegalName, &customerContactName,
			&customerWeixin, &customerEmail, &customerPrimaryPhone,
			&customerProvince, &customerCity, &customerArea,
			&customerDetailAddress, &customerRemark,
			&customerStatus, &customerDealStatus,
			&customerOwnerUserIDRaw, &customerOwnerUserName,
			&customerNextTimeRaw, &customerFollowTimeRaw, &customerCollectTimeRaw,
			&customerSourceIDRaw, &customerSourceNameRaw, &customerLevelIDRaw, &customerLevelNameRaw,
			&record.Content, &nextFollowTimeRaw,
			&recordCustomerLevelID, &recordCustomerLevelName,
			&recordCustomerSourceID, &recordCustomerSourceName,
			&record.FollowMethodID, &record.FollowMethodName, &record.OperatorUserID, &record.OperatorUserName,
			&createdAtRaw, &updatedAtRaw,
		)
		if err != nil {
			return result, err
		}

		nextFollowTime, err := parseNullableDBTime(nextFollowTimeRaw)
		if err != nil {
			return result, err
		}
		record.NextFollowTime = nextFollowTime

		customerNextTime, err := parseNullableDBTime(customerNextTimeRaw)
		if err != nil {
			return result, err
		}

		customerFollowTime, err := parseNullableDBTime(customerFollowTimeRaw)
		if err != nil {
			return result, err
		}

		customerCollectTime, err := parseNullableDBTime(customerCollectTimeRaw)
		if err != nil {
			return result, err
		}

		record.CreatedAt, err = parseDBTime(createdAtRaw)
		if err != nil {
			return result, err
		}
		record.UpdatedAt, err = parseDBTime(updatedAtRaw)
		if err != nil {
			return result, err
		}

		record.CustomerLevelID = recordCustomerLevelID
		record.CustomerLevelName = recordCustomerLevelName
		record.CustomerSourceID = recordCustomerSourceID
		record.CustomerSourceName = recordCustomerSourceName

		customer := &model.OperationFollowRecordCustomer{
			ID:            record.CustomerID,
			Name:          customerName,
			LegalName:     customerLegalName,
			ContactName:   customerContactName,
			Weixin:        customerWeixin,
			Email:         customerEmail,
			PrimaryPhone:  customerPrimaryPhone,
			Province:      customerProvince,
			City:          customerCity,
			Area:          customerArea,
			DetailAddress: customerDetailAddress,
			Remark:        customerRemark,
			Status:        customerStatus,
			DealStatus:    customerDealStatus,
			OwnerUserName: customerOwnerUserName,
			NextTime:      customerNextTime,
			FollowTime:    customerFollowTime,
			CollectTime:   customerCollectTime,
			LevelID:       customerLevelIDRaw,
			LevelName:     customerLevelNameRaw,
			SourceID:      customerSourceIDRaw,
			SourceName:    customerSourceNameRaw,
		}
		if customerOwnerUserIDRaw > 0 {
			ownerID := customerOwnerUserIDRaw
			customer.OwnerUserID = &ownerID
		}
		record.Customer = customer

		result.Items = append(result.Items, record)
	}

	if result.Items == nil {
		result.Items = []model.SalesFollowRecord{}
	}

	return result, nil
}

// GetFollowMethodByID 根据ID获取跟进方式
func (r *FollowRecordRepository) GetFollowMethodByID(id int) (*model.FollowMethod, error) {
	var method model.FollowMethod
	err := r.db.Table("follow_methods").
		Select("id", "name", "sort", "created_at").
		Where("id = ?", id).
		Take(&method).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &method, nil
}

// ListFollowMethods 获取所有跟进方式
func (r *FollowRecordRepository) ListFollowMethods() ([]model.FollowMethod, error) {
	var methods []model.FollowMethod
	err := r.db.Table("follow_methods").
		Select("id", "name", "sort", "created_at").
		Order("sort ASC, id ASC").
		Find(&methods).Error
	if err != nil {
		return nil, err
	}

	if methods == nil {
		methods = []model.FollowMethod{}
	}

	return methods, nil
}

// CreateFollowMethod 创建跟进方式
func (r *FollowRecordRepository) CreateFollowMethod(req model.FollowMethodRequest) (int64, error) {
	row := followMethodRow{Name: req.Name, Sort: req.Sort}
	if err := r.db.Table("follow_methods").Create(&row).Error; err != nil {
		return 0, err
	}
	return int64(row.ID), nil
}

// UpdateFollowMethod 更新跟进方式
func (r *FollowRecordRepository) UpdateFollowMethod(id int, req model.FollowMethodRequest) error {
	return r.db.Table("follow_methods").
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"name": req.Name,
			"sort": req.Sort,
		}).Error
}

// DeleteFollowMethod 删除跟进方式
func (r *FollowRecordRepository) DeleteFollowMethod(id int) error {
	return r.db.Table("follow_methods").Where("id = ?", id).Delete(nil).Error
}

func parseNullableDBTime(value interface{}) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := parseDBTime(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseDBTime(value interface{}) (time.Time, error) {
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case int64:
		return time.Unix(v, 0), nil
	case int:
		return time.Unix(int64(v), 0), nil
	case float64:
		return time.Unix(int64(v), 0), nil
	case []byte:
		return parseDBTimeString(string(v))
	case string:
		return parseDBTimeString(v)
	default:
		return time.Time{}, fmt.Errorf("unsupported datetime type: %T", value)
	}
}

func parseDBTimeString(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, fmt.Errorf("empty datetime string")
	}

	if unixSec, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		return time.Unix(unixSec, 0), nil
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05 -0700 MST",
		"2006-01-02 15:04:05 -0700",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, trimmed); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid datetime value: %s", trimmed)
}
