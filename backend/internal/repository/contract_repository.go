package repository

import (
	"backend/internal/model"
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	gomysql "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrContractNotFound           = errors.New("contract not found")
	ErrContractNumberExists       = errors.New("contract number already exists")
	ErrContractInvalidUser        = errors.New("invalid user")
	ErrContractInvalidCustomer    = errors.New("invalid customer")
	ErrContractInvalidServiceUser = errors.New("invalid service user")
)

type ContractRepository interface {
	List(ctx context.Context, filter model.ContractListFilter) (model.ContractListResult, error)
	GetByID(ctx context.Context, id int64) (*model.Contract, error)
	Create(ctx context.Context, input model.ContractCreateInput) (*model.Contract, error)
	Update(ctx context.Context, id int64, input model.ContractUpdateInput) (*model.Contract, error)
	Delete(ctx context.Context, id int64) error
	ExistsContractNumber(ctx context.Context, contractNumber string, excludeID int64) (bool, error)
	ExistsUser(ctx context.Context, id int64) (bool, error)
	ExistsCustomer(ctx context.Context, id int64) (bool, error)
	ListUserIDsByRoleNames(ctx context.Context, roleNames []string) ([]int64, error)
	ListDirectSubordinateUserIDsByRoleNames(ctx context.Context, parentIDs []int64, roleNames []string) ([]int64, error)
}

type gormContractRepository struct {
	db *gorm.DB
}

type contractListRow struct {
	ID                   int64         `gorm:"column:id"`
	ContractImage        string        `gorm:"column:contract_image"`
	PaymentImage         string        `gorm:"column:payment_image"`
	PaymentStatus        string        `gorm:"column:payment_status"`
	Remark               string        `gorm:"column:remark"`
	UserID               int64         `gorm:"column:user_id"`
	CustomerID           int64         `gorm:"column:customer_id"`
	CooperationType      string        `gorm:"column:cooperation_type"`
	ContractNumber       string        `gorm:"column:contract_number"`
	ContractName         string        `gorm:"column:contract_name"`
	ContractAmount       float64       `gorm:"column:contract_amount"`
	PaymentAmount        float64       `gorm:"column:payment_amount"`
	CooperationYears     int           `gorm:"column:cooperation_years"`
	NodeCount            int           `gorm:"column:node_count"`
	ServiceUserID        sql.NullInt64 `gorm:"column:service_user_id"`
	WebsiteName          string        `gorm:"column:website_name"`
	WebsiteURL           string        `gorm:"column:website_url"`
	WebsiteUsername      string        `gorm:"column:website_username"`
	IsOnline             bool          `gorm:"column:is_online"`
	StartDateUnix        sql.NullInt64 `gorm:"column:start_date"`
	EndDateUnix          sql.NullInt64 `gorm:"column:end_date"`
	AuditStatus          string        `gorm:"column:audit_status"`
	AuditComment         string        `gorm:"column:audit_comment"`
	AuditedBy            sql.NullInt64 `gorm:"column:audited_by"`
	AuditedAt            sql.NullTime  `gorm:"column:audited_at"`
	ExpiryHandlingStatus string        `gorm:"column:expiry_handling_status"`
	CreatedAt            time.Time     `gorm:"column:created_at"`
	UpdatedAt            time.Time     `gorm:"column:updated_at"`
	UserName             string        `gorm:"column:user_name"`
	CustomerName         string        `gorm:"column:customer_name"`
	ServiceUserName      string        `gorm:"column:service_user_name"`
	AuditedByName        string        `gorm:"column:audited_by_name"`
}

func NewGormContractRepository(db *gorm.DB) ContractRepository {
	return &gormContractRepository{db: db}
}

func NewSQLiteContractRepository(db *gorm.DB) ContractRepository {
	return NewGormContractRepository(db)
}

func (r *gormContractRepository) List(ctx context.Context, filter model.ContractListFilter) (model.ContractListResult, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 200 {
		pageSize = 200
	}

	where, args := buildContractListWhere(filter)
	base := r.baseQuery(ctx)
	if len(where) > 0 {
		base = base.Where(strings.Join(where, " AND "), args...)
	}

	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return model.ContractListResult{}, err
	}

	var rows []contractListRow
	err := base.Session(&gorm.Session{}).
		Select(contractSelectColumns).
		Order("c.updated_at DESC, c.id DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Scan(&rows).Error
	if err != nil {
		return model.ContractListResult{}, err
	}

	items := make([]model.Contract, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapContractRow(row))
	}

	return model.ContractListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (r *gormContractRepository) GetByID(ctx context.Context, id int64) (*model.Contract, error) {
	var row contractListRow
	err := r.baseQuery(ctx).
		Select(contractSelectColumns).
		Where("c.id = ?", id).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContractNotFound
		}
		return nil, err
	}
	contract := mapContractRow(row)
	return &contract, nil
}

func (r *gormContractRepository) Create(ctx context.Context, input model.ContractCreateInput) (*model.Contract, error) {
	type contractCreateRow struct {
		ID                   int64      `gorm:"column:id;primaryKey;autoIncrement"`
		ContractImage        string     `gorm:"column:contract_image"`
		PaymentImage         string     `gorm:"column:payment_image"`
		PaymentStatus        string     `gorm:"column:payment_status"`
		Remark               string     `gorm:"column:remark"`
		UserID               int64      `gorm:"column:user_id"`
		CustomerID           int64      `gorm:"column:customer_id"`
		CooperationType      string     `gorm:"column:cooperation_type"`
		ContractNumber       string     `gorm:"column:contract_number"`
		ContractName         string     `gorm:"column:contract_name"`
		ContractAmount       float64    `gorm:"column:contract_amount"`
		PaymentAmount        float64    `gorm:"column:payment_amount"`
		CooperationYears     int        `gorm:"column:cooperation_years"`
		NodeCount            int        `gorm:"column:node_count"`
		ServiceUserID        *int64     `gorm:"column:service_user_id"`
		WebsiteName          string     `gorm:"column:website_name"`
		WebsiteURL           string     `gorm:"column:website_url"`
		WebsiteUsername      string     `gorm:"column:website_username"`
		IsOnline             bool       `gorm:"column:is_online"`
		StartDate            *int64     `gorm:"column:start_date"`
		EndDate              *int64     `gorm:"column:end_date"`
		AuditStatus          string     `gorm:"column:audit_status"`
		AuditComment         string     `gorm:"column:audit_comment"`
		AuditedBy            *int64     `gorm:"column:audited_by"`
		AuditedAt            *time.Time `gorm:"column:audited_at"`
		ExpiryHandlingStatus string     `gorm:"column:expiry_handling_status"`
		CreatedAt            time.Time  `gorm:"column:created_at"`
		UpdatedAt            time.Time  `gorm:"column:updated_at"`
	}

	now := time.Now().UTC()
	row := contractCreateRow{
		ContractImage:        input.ContractImage,
		PaymentImage:         input.PaymentImage,
		PaymentStatus:        input.PaymentStatus,
		Remark:               input.Remark,
		UserID:               input.UserID,
		CustomerID:           input.CustomerID,
		CooperationType:      input.CooperationType,
		ContractNumber:       input.ContractNumber,
		ContractName:         input.ContractName,
		ContractAmount:       input.ContractAmount,
		PaymentAmount:        input.PaymentAmount,
		CooperationYears:     input.CooperationYears,
		NodeCount:            input.NodeCount,
		ServiceUserID:        input.ServiceUserID,
		WebsiteName:          input.WebsiteName,
		WebsiteURL:           input.WebsiteURL,
		WebsiteUsername:      input.WebsiteUsername,
		IsOnline:             input.IsOnline,
		StartDate:            input.StartDate,
		EndDate:              input.EndDate,
		AuditStatus:          input.AuditStatus,
		AuditComment:         input.AuditComment,
		AuditedBy:            input.AuditedBy,
		AuditedAt:            input.AuditedAt,
		ExpiryHandlingStatus: input.ExpiryHandlingStatus,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("contracts").Create(&row).Error; err != nil {
			if isContractNumberUniqueErr(err) {
				return ErrContractNumberExists
			}
			return err
		}
		return r.syncCustomerDealStatusTx(tx, row.CustomerID, now)
	}); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, row.ID)
}

func (r *gormContractRepository) Update(ctx context.Context, id int64, input model.ContractUpdateInput) (*model.Contract, error) {
	now := time.Now().UTC()
	type contractCustomerRow struct {
		CustomerID int64 `gorm:"column:customer_id"`
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing contractCustomerRow
		if err := tx.Table("contracts").Select("customer_id").Where("id = ?", id).Take(&existing).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrContractNotFound
			}
			return err
		}

		result := tx.Table("contracts").Where("id = ?", id).Updates(map[string]interface{}{
			"contract_image":         input.ContractImage,
			"payment_image":          input.PaymentImage,
			"payment_status":         input.PaymentStatus,
			"remark":                 input.Remark,
			"user_id":                input.UserID,
			"customer_id":            input.CustomerID,
			"cooperation_type":       input.CooperationType,
			"contract_number":        input.ContractNumber,
			"contract_name":          input.ContractName,
			"contract_amount":        input.ContractAmount,
			"payment_amount":         input.PaymentAmount,
			"cooperation_years":      input.CooperationYears,
			"node_count":             input.NodeCount,
			"service_user_id":        input.ServiceUserID,
			"website_name":           input.WebsiteName,
			"website_url":            input.WebsiteURL,
			"website_username":       input.WebsiteUsername,
			"is_online":              input.IsOnline,
			"start_date":             input.StartDate,
			"end_date":               input.EndDate,
			"audit_status":           input.AuditStatus,
			"audit_comment":          input.AuditComment,
			"audited_by":             input.AuditedBy,
			"audited_at":             input.AuditedAt,
			"expiry_handling_status": input.ExpiryHandlingStatus,
			"updated_at":             now,
		})
		if result.Error != nil {
			if isContractNumberUniqueErr(result.Error) {
				return ErrContractNumberExists
			}
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrContractNotFound
		}

		if err := r.syncCustomerDealStatusTx(tx, existing.CustomerID, now); err != nil {
			return err
		}
		if input.CustomerID != existing.CustomerID {
			if err := r.syncCustomerDealStatusTx(tx, input.CustomerID, now); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *gormContractRepository) Delete(ctx context.Context, id int64) error {
	type contractCustomerRow struct {
		CustomerID int64 `gorm:"column:customer_id"`
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing contractCustomerRow
		if err := tx.Table("contracts").Select("customer_id").Where("id = ?", id).Take(&existing).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrContractNotFound
			}
			return err
		}

		result := tx.Table("contracts").Where("id = ?", id).Delete(nil)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrContractNotFound
		}

		return r.syncCustomerDealStatusTx(tx, existing.CustomerID, time.Now().UTC())
	})
}

func (r *gormContractRepository) syncCustomerDealStatusTx(tx *gorm.DB, customerID int64, now time.Time) error {
	if customerID <= 0 {
		return nil
	}

	type contractSummary struct {
		ContractCount          int64        `gorm:"column:contract_count"`
		FirstContractCreatedAt sql.NullTime `gorm:"column:first_contract_created_at"`
	}

	var summary contractSummary
	if err := tx.Table("contracts").
		Select("COUNT(*) AS contract_count, MIN(created_at) AS first_contract_created_at").
		Where("customer_id = ?", customerID).
		Scan(&summary).Error; err != nil {
		return err
	}

	updates := map[string]interface{}{
		"updated_at": now,
	}
	if summary.ContractCount > 0 {
		updates["deal_status"] = model.CustomerDealStatusDone
		if summary.FirstContractCreatedAt.Valid {
			updates["deal_time"] = summary.FirstContractCreatedAt.Time.Unix()
		} else {
			updates["deal_time"] = now.Unix()
		}
	} else {
		updates["deal_status"] = model.CustomerDealStatusUndone
		updates["deal_time"] = nil
	}

	return tx.Table("customers").Where("id = ?", customerID).Updates(updates).Error
}

func (r *gormContractRepository) ExistsContractNumber(ctx context.Context, contractNumber string, excludeID int64) (bool, error) {
	number := strings.TrimSpace(contractNumber)
	if number == "" {
		return false, nil
	}
	query := r.db.WithContext(ctx).Table("contracts").Where("contract_number = ?", number)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *gormContractRepository) ExistsUser(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, nil
	}
	var count int64
	if err := r.db.WithContext(ctx).Table("users").Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *gormContractRepository) ExistsCustomer(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, nil
	}
	var count int64
	if err := r.db.WithContext(ctx).Table("customers").Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *gormContractRepository) ListUserIDsByRoleNames(ctx context.Context, roleNames []string) ([]int64, error) {
	if len(roleNames) == 0 {
		return []int64{}, nil
	}

	cleanRoleNames := make([]string, 0, len(roleNames))
	for _, roleName := range roleNames {
		name := strings.TrimSpace(roleName)
		if name == "" {
			continue
		}
		cleanRoleNames = append(cleanRoleNames, name)
	}
	if len(cleanRoleNames) == 0 {
		return []int64{}, nil
	}

	var ids []int64
	err := r.db.WithContext(ctx).
		Table("users AS u").
		Joins("LEFT JOIN roles r ON u.role_id = r.id").
		Where("r.name IN ?", cleanRoleNames).
		Order("u.id ASC").
		Pluck("u.id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *gormContractRepository) ListDirectSubordinateUserIDsByRoleNames(ctx context.Context, parentIDs []int64, roleNames []string) ([]int64, error) {
	if len(parentIDs) == 0 {
		return []int64{}, nil
	}
	cleanParentIDs := uniquePositiveInt64(parentIDs)
	if len(cleanParentIDs) == 0 {
		return []int64{}, nil
	}

	query := r.db.WithContext(ctx).
		Table("users AS u").
		Joins("LEFT JOIN roles r ON u.role_id = r.id").
		Where("u.parent_id IN ?", cleanParentIDs)

	cleanRoleNames := make([]string, 0, len(roleNames))
	for _, roleName := range roleNames {
		name := strings.TrimSpace(roleName)
		if name == "" {
			continue
		}
		cleanRoleNames = append(cleanRoleNames, name)
	}
	if len(cleanRoleNames) > 0 {
		query = query.Where("r.name IN ?", cleanRoleNames)
	}

	var ids []int64
	err := query.Order("u.id ASC").Pluck("u.id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *gormContractRepository) baseQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Table("contracts AS c").
		Joins("LEFT JOIN users u ON c.user_id = u.id").
		Joins("LEFT JOIN customers cu ON c.customer_id = cu.id").
		Joins("LEFT JOIN users su ON c.service_user_id = su.id").
		Joins("LEFT JOIN users au ON c.audited_by = au.id")
}

func buildContractListWhere(filter model.ContractListFilter) ([]string, []interface{}) {
	var where []string
	var args []interface{}

	if filter.Keyword != "" {
		pattern := "%" + filter.Keyword + "%"
		where = append(where, "(c.contract_number LIKE ? OR c.contract_name LIKE ? OR COALESCE(cu.name, '') LIKE ? OR COALESCE(u.nickname, '') LIKE ?)")
		args = append(args, pattern, pattern, pattern, pattern)
	}
	if filter.PaymentStatus != "" {
		where = append(where, "c.payment_status = ?")
		args = append(args, filter.PaymentStatus)
	}
	if filter.CooperationType != "" {
		where = append(where, "c.cooperation_type = ?")
		args = append(args, filter.CooperationType)
	}
	if filter.AuditStatus != "" {
		where = append(where, "c.audit_status = ?")
		args = append(args, filter.AuditStatus)
	}
	if filter.ExpiryHandlingStatus != "" {
		where = append(where, "c.expiry_handling_status = ?")
		args = append(args, filter.ExpiryHandlingStatus)
	}
	if filter.UserID > 0 {
		where = append(where, "c.user_id = ?")
		args = append(args, filter.UserID)
	}
	if filter.CustomerID > 0 {
		where = append(where, "c.customer_id = ?")
		args = append(args, filter.CustomerID)
	}
	if len(filter.AllowedUserIDs) > 0 {
		where = append(where, "c.user_id IN ?")
		args = append(args, filter.AllowedUserIDs)
	}
	if len(filter.AllowedServiceUserIDs) > 0 {
		where = append(where, "c.service_user_id IN ?")
		args = append(args, filter.AllowedServiceUserIDs)
	}
	if filter.ForceServiceUserID != nil {
		where = append(where, "c.service_user_id = ?")
		args = append(args, *filter.ForceServiceUserID)
	}
	return where, args
}

func uniquePositiveInt64(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

const contractSelectColumns = `
	c.id AS id,
	COALESCE(c.contract_image, '') AS contract_image,
	COALESCE(c.payment_image, '') AS payment_image,
	COALESCE(c.payment_status, '') AS payment_status,
	COALESCE(c.remark, '') AS remark,
	c.user_id AS user_id,
	c.customer_id AS customer_id,
	COALESCE(c.cooperation_type, '') AS cooperation_type,
	COALESCE(c.contract_number, '') AS contract_number,
	COALESCE(c.contract_name, '') AS contract_name,
	COALESCE(c.contract_amount, 0) AS contract_amount,
	COALESCE(c.payment_amount, 0) AS payment_amount,
	COALESCE(c.cooperation_years, 0) AS cooperation_years,
	COALESCE(c.node_count, 0) AS node_count,
	c.service_user_id AS service_user_id,
	COALESCE(c.website_name, '') AS website_name,
	COALESCE(c.website_url, '') AS website_url,
	COALESCE(c.website_username, '') AS website_username,
	COALESCE(c.is_online, 0) AS is_online,
	c.start_date AS start_date,
	c.end_date AS end_date,
	COALESCE(c.audit_status, '') AS audit_status,
	COALESCE(c.audit_comment, '') AS audit_comment,
	c.audited_by AS audited_by,
	c.audited_at AS audited_at,
	COALESCE(c.expiry_handling_status, '') AS expiry_handling_status,
	c.created_at AS created_at,
	c.updated_at AS updated_at,
	COALESCE(u.nickname, '') AS user_name,
	COALESCE(cu.name, '') AS customer_name,
	COALESCE(su.nickname, '') AS service_user_name,
	COALESCE(au.nickname, '') AS audited_by_name
`

func mapContractRow(row contractListRow) model.Contract {
	item := model.Contract{
		ID:                   row.ID,
		ContractImage:        row.ContractImage,
		PaymentImage:         row.PaymentImage,
		PaymentStatus:        row.PaymentStatus,
		Remark:               row.Remark,
		UserID:               row.UserID,
		CustomerID:           row.CustomerID,
		CooperationType:      row.CooperationType,
		ContractNumber:       row.ContractNumber,
		ContractName:         row.ContractName,
		ContractAmount:       row.ContractAmount,
		PaymentAmount:        row.PaymentAmount,
		CooperationYears:     row.CooperationYears,
		NodeCount:            row.NodeCount,
		WebsiteName:          row.WebsiteName,
		WebsiteURL:           row.WebsiteURL,
		WebsiteUsername:      row.WebsiteUsername,
		IsOnline:             row.IsOnline,
		AuditStatus:          row.AuditStatus,
		AuditComment:         row.AuditComment,
		ExpiryHandlingStatus: row.ExpiryHandlingStatus,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
		UserName:             row.UserName,
		CustomerName:         row.CustomerName,
		ServiceUserName:      row.ServiceUserName,
		AuditedByName:        row.AuditedByName,
	}
	if row.ServiceUserID.Valid {
		item.ServiceUserID = &row.ServiceUserID.Int64
	}
	if row.AuditedBy.Valid {
		item.AuditedBy = &row.AuditedBy.Int64
	}
	if row.StartDateUnix.Valid {
		v := row.StartDateUnix.Int64
		item.StartDateUnix = &v
		t := time.Unix(v, 0).UTC()
		item.StartDate = &t
	}
	if row.EndDateUnix.Valid {
		v := row.EndDateUnix.Int64
		item.EndDateUnix = &v
		t := time.Unix(v, 0).UTC()
		item.EndDate = &t
	}
	if row.AuditedAt.Valid {
		auditedAt := row.AuditedAt.Time.UTC()
		item.AuditedAt = &auditedAt
	}
	return item
}

func isContractNumberUniqueErr(err error) bool {
	if err == nil {
		return false
	}
	if isSQLiteUniqueErr(err, "contracts.contract_number") {
		return true
	}
	var mysqlErr *gomysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return strings.Contains(strings.ToLower(mysqlErr.Message), "contract_number") ||
			strings.Contains(strings.ToLower(mysqlErr.Message), "uk_contracts_contract_number")
	}
	return false
}

func isSQLiteUniqueErr(err error, key string) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique constraint failed") && strings.Contains(msg, strings.ToLower(key))
}
