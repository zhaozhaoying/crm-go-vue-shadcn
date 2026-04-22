package repository

import (
	"backend/internal/model"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

var (
	ErrCustomerNotFound      = errors.New("customer not found")
	ErrCustomerNotInPool     = errors.New("customer not in pool")
	ErrCustomerAlreadyInPool = errors.New("customer already in pool")
	ErrCustomerNotOwned      = errors.New("customer not owned")
	ErrPhoneNotFound         = errors.New("phone not found")
	ErrPhoneAlreadyExists    = errors.New("phone already exists for this customer")
	ErrInvalidPhoneFormat    = errors.New("invalid phone format")
)

type CustomerRepository interface {
	List(ctx context.Context, filter model.CustomerListFilter) (model.CustomerListResult, error)
	ListAssignments(ctx context.Context, filter model.CustomerAssignmentListFilter) (model.CustomerAssignmentListResult, error)
	FindByID(ctx context.Context, customerID int64) (*model.Customer, error)
	ListUserIDsByRoleNames(ctx context.Context, roleNames []string) ([]int64, error)
	ListEnabledUserIDsByRoleNames(ctx context.Context, roleNames []string) ([]int64, error)
	ListDirectSubordinateUserIDsByRoleNames(ctx context.Context, parentIDs []int64, roleNames []string) ([]int64, error)
	GetUserRoleName(ctx context.Context, userID int64) (string, error)
	GetUserDisplayName(ctx context.Context, userID int64) (string, error)
	GetParentUserID(ctx context.Context, userID int64) (int64, error)
	ListAutoAssignRankedOwnerScores(ctx context.Context, referenceDate string, userIDs []int64) ([]model.SalesDailyScore, error)
	ListRecentContractExemptOwnerUserIDs(ctx context.Context, since time.Time, userIDs []int64) ([]int64, error)
	FindEnabledUserIDByNickname(ctx context.Context, nickname string) (int64, error)
	FindLatestAutoAssignOwnerUserID(ctx context.Context, ownerUserIDs []int64, since time.Time) (*int64, error)
	ResolveDepartmentAnchorUserID(ctx context.Context, userID int64) (int64, error)
	GetActiveBlockedUntilByDepartmentAnchor(ctx context.Context, customerID, departmentAnchorUserID int64, now time.Time) (*time.Time, error)
	CountOwnedActiveByOwner(ctx context.Context, ownerUserID int64) (int64, error)
	Create(ctx context.Context, input model.CustomerCreateInput) (*model.Customer, error)
	Update(ctx context.Context, customerID int64, input model.CustomerUpdateInput) (*model.Customer, error)
	CheckUnique(ctx context.Context, input model.CustomerUniqueCheckInput) (model.CustomerUniqueCheckResult, error)
	Claim(ctx context.Context, customerID, ownerUserID, operatorUserID int64, insideSalesUserID *int64) (*model.Customer, error)
	Release(ctx context.Context, customerID, operatorUserID int64) (*model.Customer, error)
	Transfer(ctx context.Context, input model.CustomerTransferInput) (*model.Customer, error)
	Convert(ctx context.Context, customerID, ownerUserID, operatorUserID int64) (*model.Customer, error)

	// Phone management
	AddPhone(ctx context.Context, phone *model.CustomerPhone) error
	ListPhones(ctx context.Context, customerID int64) ([]model.CustomerPhone, error)
	UpdatePhone(ctx context.Context, phone *model.CustomerPhone) error
	DeletePhone(ctx context.Context, customerID, phoneID int64) error
	GetPhone(ctx context.Context, phoneID int64) (*model.CustomerPhone, error)
	FindCustomerIDByPhone(ctx context.Context, phone string) (int64, error)

	// Status log management
	CreateStatusLog(ctx context.Context, log *model.CustomerStatusLog) error
	ListStatusLogs(ctx context.Context, customerID int64, page, pageSize int) ([]model.CustomerStatusLog, error)
}

type gormCustomerRepository struct {
	db *gorm.DB
}

const defaultClaimFreezeDays = 7

type customerListRow struct {
	ID                         int64         `gorm:"column:id"`
	Name                       string        `gorm:"column:name"`
	LegalName                  string        `gorm:"column:legal_name"`
	ContactName                string        `gorm:"column:contact_name"`
	Weixin                     string        `gorm:"column:weixin"`
	Email                      string        `gorm:"column:email"`
	CustomerLevelID            int           `gorm:"column:customer_level_id"`
	CustomerSourceID           int           `gorm:"column:customer_source_id"`
	CustomerLevelName          string        `gorm:"column:customer_level_name"`
	CustomerSourceName         string        `gorm:"column:customer_source_name"`
	Province                   int           `gorm:"column:province"`
	City                       int           `gorm:"column:city"`
	Area                       int           `gorm:"column:area"`
	DetailAddress              string        `gorm:"column:detail_address"`
	Remark                     string        `gorm:"column:remark"`
	Status                     string        `gorm:"column:status"`
	DealStatus                 string        `gorm:"column:deal_status"`
	CreateUserID               int64         `gorm:"column:create_user_id"`
	InsideSalesUserID          sql.NullInt64 `gorm:"column:inside_sales_user_id"`
	InsideSalesUserName        string        `gorm:"column:inside_sales_user_name"`
	ConvertedAt                sql.NullTime  `gorm:"column:converted_at"`
	OwnerUserID                sql.NullInt64 `gorm:"column:owner_user_id"`
	OwnerUserName              string        `gorm:"column:owner_user_name"`
	AssignmentReason           string        `gorm:"column:assignment_reason"`
	AssignmentOperatorUserID   sql.NullInt64 `gorm:"column:assignment_operator_user_id"`
	AssignmentOperatorUserName string        `gorm:"column:assignment_operator_user_name"`
	CreatedAt                  time.Time     `gorm:"column:created_at"`
	UpdatedAt                  time.Time     `gorm:"column:updated_at"`
	NextTimeUnix               sql.NullInt64 `gorm:"column:next_time_unix"`
	FollowTimeUnix             sql.NullInt64 `gorm:"column:follow_time_unix"`
	CollectTimeUnix            sql.NullInt64 `gorm:"column:collect_time_unix"`
	AssignTimeUnix             sql.NullInt64 `gorm:"column:assign_time_unix"`
	DropTimeUnix               sql.NullInt64 `gorm:"column:drop_time_unix"`
	DropUserID                 sql.NullInt64 `gorm:"column:drop_user_id"`
	DropUserName               string        `gorm:"column:drop_user_name"`
	IsInPool                   bool          `gorm:"column:is_in_pool"`
}

type customerPhoneRow struct {
	ID         int64  `gorm:"column:id;primaryKey;autoIncrement"`
	CustomerID int64  `gorm:"column:customer_id"`
	Phone      string `gorm:"column:phone"`
	PhoneLabel string `gorm:"column:phone_label"`
	IsPrimary  int    `gorm:"column:is_primary"`
	CreatedAt  int64  `gorm:"column:created_at"`
	UpdatedAt  int64  `gorm:"column:updated_at"`
}

type customerStatusLogRow struct {
	ID             int64         `gorm:"column:id;primaryKey;autoIncrement"`
	CustomerID     int64         `gorm:"column:customer_id"`
	FromStatus     int           `gorm:"column:from_status"`
	ToStatus       int           `gorm:"column:to_status"`
	TriggerType    int           `gorm:"column:trigger_type"`
	Reason         string        `gorm:"column:reason"`
	OperatorUserID sql.NullInt64 `gorm:"column:operator_user_id"`
	OperateTime    int64         `gorm:"column:operate_time"`
}

type customerStatusLogListRow struct {
	customerStatusLogRow
	OperatorName string `gorm:"column:operator_name"`
}

type customerAssignmentListRow struct {
	ID              int64     `gorm:"column:id"`
	Date            time.Time `gorm:"column:date"`
	InsideSalesName string    `gorm:"column:inside_sales_name"`
	SalesName       string    `gorm:"column:sales_name"`
	CustomerName    string    `gorm:"column:customer_name"`
	LegalName       string    `gorm:"column:legal_name"`
	ContactName     string    `gorm:"column:contact_name"`
	Mobile          string    `gorm:"column:mobile"`
	Address         string    `gorm:"column:address"`
	Remark          string    `gorm:"column:remark"`
}

type customerOwnerLogRow struct {
	CustomerID                    int64      `gorm:"column:customer_id"`
	FromOwnerUserID               *int64     `gorm:"column:from_owner_user_id"`
	ToOwnerUserID                 *int64     `gorm:"column:to_owner_user_id"`
	Action                        string     `gorm:"column:action"`
	Reason                        string     `gorm:"column:reason"`
	Content                       string     `gorm:"column:content"`
	BlockedDepartmentAnchorUserID *int64     `gorm:"column:blocked_department_anchor_user_id"`
	BlockedUntil                  *time.Time `gorm:"column:blocked_until"`
	OperatorUserID                int64      `gorm:"column:operator_user_id"`
	CreatedAt                     time.Time  `gorm:"column:created_at"`
}

func newCustomerOwnerLogRow(
	customerID int64,
	fromOwnerUserID *int64,
	toOwnerUserID *int64,
	action string,
	reason string,
	content string,
	operatorUserID int64,
	createdAt time.Time,
) customerOwnerLogRow {
	return customerOwnerLogRow{
		CustomerID:      customerID,
		FromOwnerUserID: fromOwnerUserID,
		ToOwnerUserID:   toOwnerUserID,
		Action:          action,
		Reason:          reason,
		Content:         strings.TrimSpace(content),
		OperatorUserID:  operatorUserID,
		CreatedAt:       createdAt,
	}
}

func customerOwnerLogContent(note string, fallback string) string {
	trimmed := strings.TrimSpace(note)
	if trimmed != "" {
		return trimmed
	}
	return fallback
}

func assignedSalesUnix(insideSalesUserID *int64, ownerUserID int64, now time.Time) *int64 {
	if insideSalesUserID == nil || *insideSalesUserID <= 0 || ownerUserID <= 0 {
		return nil
	}
	if ownerUserID == *insideSalesUserID {
		return nil
	}
	assignedAt := now.Unix()
	return &assignedAt
}

func collectTimeUnixOrNow(current *time.Time, now time.Time) int64 {
	if current != nil && !current.IsZero() {
		return current.Unix()
	}
	return now.Unix()
}

func NewGormCustomerRepository(db *gorm.DB) CustomerRepository {
	return &gormCustomerRepository{db: db}
}

func NewSQLiteCustomerRepository(db *gorm.DB) CustomerRepository {
	return NewGormCustomerRepository(db)
}

func buildCustomerAssignmentMeta(hasInsideSalesUser bool, reason string, isInPool bool) (string, string) {
	if hasInsideSalesUser {
		return "auto_assign", "电销分配"
	}

	switch strings.TrimSpace(reason) {
	case model.CustomerOwnerLogReasonCreateInitialAssign:
		return "self_add", "自己添加"
	case model.CustomerOwnerLogReasonImportInitialAssign:
		return "import_assign", "导入分配"
	case model.CustomerOwnerLogReasonClaimFromPool:
		return "pool_claim", "公海领取"
	case model.CustomerOwnerLogReasonManualTransfer:
		return "manual_transfer", "手动转移"
	case model.CustomerOwnerLogReasonManualRelease:
		if isInPool {
			return "manual_release", "手动丢弃"
		}
		return "manual_release", "手动处理"
	case model.CustomerOwnerLogReasonAutoDrop:
		if isInPool {
			return "auto_drop", "自动掉库"
		}
		return "auto_drop", "自动处理"
	default:
		if isInPool {
			return "", "-"
		}
		return "", "-"
	}
}

func (r *gormCustomerRepository) FindByID(ctx context.Context, customerID int64) (*model.Customer, error) {
	return r.getCustomer(ctx, customerID)
}

func (r *gormCustomerRepository) ListAssignments(ctx context.Context, filter model.CustomerAssignmentListFilter) (model.CustomerAssignmentListResult, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}

	base := r.db.WithContext(ctx).
		Table("customer_owner_logs l").
		Joins("INNER JOIN customers c ON c.id = l.customer_id").
		Joins("LEFT JOIN users iu ON iu.id = c.inside_sales_user_id").
		Joins("LEFT JOIN users su ON su.id = l.to_owner_user_id").
		Where(`
			l.reason IN ?
			OR (
				l.reason = ?
				AND TRIM(COALESCE(l.content, '')) = ?
			)
		`, []string{
			model.CustomerOwnerLogReasonInsideSalesCreate,
			model.CustomerOwnerLogReasonInsideSalesClaim,
			model.CustomerOwnerLogReasonInsideSalesConvert,
		}, model.CustomerOwnerLogReasonManualTransfer, "按昨日排名重新分配客户")

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return model.CustomerAssignmentListResult{}, err
	}

	phoneExpr := `
		COALESCE(
			(
				SELECT cp.phone
				FROM customer_phones cp
				WHERE cp.customer_id = c.id
				ORDER BY cp.is_primary DESC, cp.id ASC
				LIMIT 1
			),
			''
		)
	`

	rows := make([]customerAssignmentListRow, 0, pageSize)
	if err := base.
		Select(`
			l.id AS id,
			l.created_at AS date,
			COALESCE(NULLIF(iu.nickname, ''), NULLIF(iu.username, ''), '-') AS inside_sales_name,
			COALESCE(NULLIF(su.nickname, ''), NULLIF(su.username, ''), '-') AS sales_name,
			c.name AS customer_name,
			c.legal_name AS legal_name,
			c.contact_name AS contact_name,
			` + phoneExpr + ` AS mobile,
			c.detail_address AS address,
			c.remark AS remark
		`).
		Order("l.created_at DESC, l.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error; err != nil {
		return model.CustomerAssignmentListResult{}, err
	}

	items := make([]model.CustomerAssignmentItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.CustomerAssignmentItem{
			ID:              row.ID,
			Date:            row.Date,
			InsideSalesName: row.InsideSalesName,
			SalesName:       row.SalesName,
			CustomerName:    row.CustomerName,
			LegalName:       row.LegalName,
			ContactName:     row.ContactName,
			Mobile:          row.Mobile,
			Address:         row.Address,
			Remark:          row.Remark,
		})
	}

	return model.CustomerAssignmentListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (r *gormCustomerRepository) List(ctx context.Context, filter model.CustomerListFilter) (model.CustomerListResult, error) {
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

	where, args := buildCustomerListWhere(filter)
	dropUserExpr := `
		COALESCE(
			c.drop_user_id,
			(
				SELECT COALESCE(l.from_owner_user_id, l.operator_user_id)
				FROM customer_owner_logs l
				WHERE l.customer_id = c.id
					AND l.action = 'release'
				ORDER BY l.created_at DESC, l.id DESC
				LIMIT 1
			)
		)
	`
	assignmentLogOwnerMatchExpr := `
		(
			(c.owner_user_id IS NOT NULL AND l.to_owner_user_id = c.owner_user_id)
			OR (c.owner_user_id IS NULL AND l.to_owner_user_id IS NULL)
		)
	`
	assignmentReasonExpr := `
		COALESCE(
			(
				SELECT COALESCE(l.reason, '')
				FROM customer_owner_logs l
				WHERE l.customer_id = c.id
					AND ` + assignmentLogOwnerMatchExpr + `
				ORDER BY l.created_at DESC, l.id DESC
				LIMIT 1
			),
			''
		)
	`
	assignmentOperatorUserIDExpr := `
		(
			SELECT l.operator_user_id
			FROM customer_owner_logs l
			WHERE l.customer_id = c.id
				AND ` + assignmentLogOwnerMatchExpr + `
			ORDER BY l.created_at DESC, l.id DESC
			LIMIT 1
		)
	`
	assignmentOperatorUserNameExpr := `
		COALESCE(
			(
				SELECT COALESCE(NULLIF(op.nickname, ''), NULLIF(op.username, ''), '')
				FROM customer_owner_logs l
				LEFT JOIN users op ON op.id = l.operator_user_id
				WHERE l.customer_id = c.id
					AND ` + assignmentLogOwnerMatchExpr + `
				ORDER BY l.created_at DESC, l.id DESC
				LIMIT 1
			),
			''
		)
	`
	base := r.db.WithContext(ctx).
		Table("customers AS c").
		Joins("LEFT JOIN users u ON c.owner_user_id = u.id").
		Joins("LEFT JOIN users iu ON iu.id = c.inside_sales_user_id").
		Joins("LEFT JOIN users du ON du.id = " + dropUserExpr).
		Joins("LEFT JOIN customer_levels cl ON c.customer_level_id = cl.id").
		Joins("LEFT JOIN customer_sources cs ON c.customer_source_id = cs.id")
	if len(where) > 0 {
		base = base.Where(strings.Join(where, " AND "), args...)
	}

	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return model.CustomerListResult{}, err
	}

	var rows []customerListRow
	orderBy := buildCustomerListOrderBy(filter)
	err := base.Session(&gorm.Session{}).
		Select(`
			c.id AS id,
			c.name AS name,
			COALESCE(c.legal_name, '') AS legal_name,
			COALESCE(c.contact_name, '') AS contact_name,
			COALESCE(c.weixin, '') AS weixin,
			COALESCE(c.email, '') AS email,
			COALESCE(c.customer_level_id, 0) AS customer_level_id,
			COALESCE(c.customer_source_id, 0) AS customer_source_id,
			COALESCE(cl.name, '') AS customer_level_name,
			COALESCE(cs.name, '') AS customer_source_name,
			COALESCE(c.province, 0) AS province,
			COALESCE(c.city, 0) AS city,
			COALESCE(c.area, 0) AS area,
			COALESCE(c.detail_address, '') AS detail_address,
			COALESCE(c.remark, '') AS remark,
			c.status AS status,
			CASE
				WHEN EXISTS (SELECT 1 FROM contracts ct WHERE ct.customer_id = c.id) THEN 'done'
				ELSE 'undone'
			END AS deal_status,
			c.create_user_id AS create_user_id,
			c.inside_sales_user_id AS inside_sales_user_id,
			COALESCE(NULLIF(iu.nickname, ''), NULLIF(iu.username, ''), '') AS inside_sales_user_name,
			c.converted_at AS converted_at,
			c.owner_user_id AS owner_user_id,
			COALESCE(u.nickname, '') AS owner_user_name,
			` + assignmentReasonExpr + ` AS assignment_reason,
			` + assignmentOperatorUserIDExpr + ` AS assignment_operator_user_id,
			` + assignmentOperatorUserNameExpr + ` AS assignment_operator_user_name,
			c.created_at AS created_at,
			c.updated_at AS updated_at,
			NULLIF(c.next_time, 0) AS next_time_unix,
			NULLIF(c.follow_time, 0) AS follow_time_unix,
			NULLIF(c.collect_time, 0) AS collect_time_unix,
			NULLIF(c.assign_time, 0) AS assign_time_unix,
			NULLIF(c.drop_time, 0) AS drop_time_unix,
			` + dropUserExpr + ` AS drop_user_id,
			COALESCE(NULLIF(du.nickname, ''), NULLIF(du.username, ''), '') AS drop_user_name,
			CASE WHEN c.owner_user_id IS NULL OR c.status = 'pool' THEN 1 ELSE 0 END AS is_in_pool
		`).
		Order(orderBy).
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Scan(&rows).Error
	if err != nil {
		return model.CustomerListResult{}, err
	}

	list := make([]model.Customer, 0)
	for _, row := range rows {
		item := model.Customer{
			ID:                         row.ID,
			Name:                       row.Name,
			LegalName:                  row.LegalName,
			ContactName:                row.ContactName,
			Weixin:                     row.Weixin,
			Email:                      row.Email,
			CustomerLevelID:            row.CustomerLevelID,
			CustomerSourceID:           row.CustomerSourceID,
			CustomerLevelName:          row.CustomerLevelName,
			CustomerSourceName:         row.CustomerSourceName,
			Province:                   row.Province,
			City:                       row.City,
			Area:                       row.Area,
			DetailAddress:              row.DetailAddress,
			Remark:                     row.Remark,
			Status:                     row.Status,
			DealStatus:                 row.DealStatus,
			CreateUserID:               row.CreateUserID,
			OwnerUserName:              row.OwnerUserName,
			DropUserName:               row.DropUserName,
			CreatedAt:                  row.CreatedAt,
			UpdatedAt:                  row.UpdatedAt,
			IsInPool:                   row.IsInPool,
			AssignmentReason:           row.AssignmentReason,
			AssignmentOperatorUserName: row.AssignmentOperatorUserName,
		}
		if row.InsideSalesUserID.Valid {
			item.InsideSalesUserID = &row.InsideSalesUserID.Int64
			item.AssignmentOperatorUserID = &row.InsideSalesUserID.Int64
			item.AssignmentOperatorUserName = row.InsideSalesUserName
		}
		if row.ConvertedAt.Valid {
			convertedAt := row.ConvertedAt.Time
			item.ConvertedAt = &convertedAt
		}
		if row.OwnerUserID.Valid {
			item.OwnerUserID = &row.OwnerUserID.Int64
		}
		if item.AssignmentOperatorUserID == nil && row.AssignmentOperatorUserID.Valid {
			item.AssignmentOperatorUserID = &row.AssignmentOperatorUserID.Int64
		}
		if item.AssignmentOperatorUserName == "" {
			item.AssignmentOperatorUserName = row.AssignmentOperatorUserName
		}
		item.AssignmentType, item.AssignmentLabel = buildCustomerAssignmentMeta(
			row.InsideSalesUserID.Valid,
			row.AssignmentReason,
			row.IsInPool,
		)
		item.NextTime = nullableUnixToTime(row.NextTimeUnix)
		item.FollowTime = nullableUnixToTime(row.FollowTimeUnix)
		item.CollectTime = nullableUnixToTime(row.CollectTimeUnix)
		item.AssignTime = nullableUnixToTime(row.AssignTimeUnix)
		item.DropTime = nullableUnixToTime(row.DropTimeUnix)
		if row.DropUserID.Valid {
			item.DropUserID = &row.DropUserID.Int64
		}
		list = append(list, item)
	}

	for i := range list {
		history, err := r.listHistoricalOwnerIDs(ctx, list[i].ID)
		if err != nil {
			return model.CustomerListResult{}, err
		}
		list[i].HistoricalOwnerIDs = history

		// Load phones for each customer
		phones, err := r.ListPhones(ctx, list[i].ID)
		if err != nil {
			return model.CustomerListResult{}, err
		}
		list[i].Phones = phones
	}

	return model.CustomerListResult{
		Items:    list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (r *gormCustomerRepository) ListDirectSubordinateUserIDsByRoleNames(ctx context.Context, parentIDs []int64, roleNames []string) ([]int64, error) {
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

func (r *gormCustomerRepository) ListUserIDsByRoleNames(ctx context.Context, roleNames []string) ([]int64, error) {
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

func (r *gormCustomerRepository) ListEnabledUserIDsByRoleNames(ctx context.Context, roleNames []string) ([]int64, error) {
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
		Where("u.status = ?", model.UserStatusEnabled).
		Order("u.id ASC").
		Pluck("u.id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *gormCustomerRepository) GetUserRoleName(ctx context.Context, userID int64) (string, error) {
	if userID <= 0 {
		return "", nil
	}

	var roleName string
	err := r.db.WithContext(ctx).
		Table("users AS u").
		Select("COALESCE(r.name, '')").
		Joins("LEFT JOIN roles r ON u.role_id = r.id").
		Where("u.id = ?", userID).
		Limit(1).
		Scan(&roleName).Error
	if err != nil {
		return "", err
	}
	return roleName, nil
}

func (r *gormCustomerRepository) GetUserDisplayName(ctx context.Context, userID int64) (string, error) {
	if userID <= 0 {
		return "", nil
	}

	var displayName string
	err := r.db.WithContext(ctx).
		Table("users").
		Select("COALESCE(NULLIF(nickname, ''), NULLIF(username, ''), '')").
		Where("id = ?", userID).
		Limit(1).
		Scan(&displayName).Error
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(displayName), nil
}

func (r *gormCustomerRepository) FindEnabledUserIDByNickname(ctx context.Context, nickname string) (int64, error) {
	trimmed := strings.TrimSpace(nickname)
	if trimmed == "" {
		return 0, nil
	}

	type userRow struct {
		ID int64 `gorm:"column:id"`
	}

	var row userRow
	err := r.db.WithContext(ctx).
		Table("users").
		Select("id").
		Where("nickname = ?", trimmed).
		Where("status = ?", model.UserStatusEnabled).
		Order("id ASC").
		Limit(1).
		Scan(&row).Error
	if err != nil {
		return 0, err
	}
	return row.ID, nil
}

func (r *gormCustomerRepository) ResolveDepartmentAnchorUserID(ctx context.Context, userID int64) (int64, error) {
	if userID <= 0 {
		return 0, nil
	}

	visited := map[int64]struct{}{}
	currentID := userID

	for currentID > 0 {
		if _, seen := visited[currentID]; seen {
			return currentID, nil
		}
		visited[currentID] = struct{}{}

		subordinateIDs, err := r.ListDirectSubordinateUserIDsByRoleNames(ctx, []int64{currentID}, nil)
		if err != nil {
			return 0, err
		}
		if len(uniquePositiveInt64(subordinateIDs)) > 0 {
			return currentID, nil
		}

		type parentRow struct {
			ParentID sql.NullInt64 `gorm:"column:parent_id"`
		}

		var row parentRow
		if err := r.db.WithContext(ctx).
			Table("users").
			Select("parent_id").
			Where("id = ?", currentID).
			Take(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, nil
			}
			return 0, err
		}

		if !row.ParentID.Valid || row.ParentID.Int64 <= 0 {
			return currentID, nil
		}
		currentID = row.ParentID.Int64
	}

	return 0, nil
}

func (r *gormCustomerRepository) GetParentUserID(ctx context.Context, userID int64) (int64, error) {
	if userID <= 0 {
		return 0, nil
	}

	type parentRow struct {
		ParentID sql.NullInt64 `gorm:"column:parent_id"`
	}

	var row parentRow
	if err := r.db.WithContext(ctx).
		Table("users").
		Select("parent_id").
		Where("id = ?", userID).
		Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}

	if !row.ParentID.Valid || row.ParentID.Int64 <= 0 {
		return 0, nil
	}
	return row.ParentID.Int64, nil
}

func (r *gormCustomerRepository) ListAutoAssignRankedOwnerScores(ctx context.Context, referenceDate string, userIDs []int64) ([]model.SalesDailyScore, error) {
	cleanUserIDs := uniquePositiveInt64(userIDs)
	if len(cleanUserIDs) == 0 {
		return []model.SalesDailyScore{}, nil
	}

	type latestDateRow struct {
		ScoreDate string `gorm:"column:score_date"`
	}

	var latest latestDateRow
	err := r.db.WithContext(ctx).
		Table("sales_daily_scores").
		Select("MAX(score_date) AS score_date").
		Where("user_id IN ?", cleanUserIDs).
		Where("score_date <= ?", strings.TrimSpace(referenceDate)).
		Where("total_score > 0").
		Scan(&latest).Error
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(latest.ScoreDate) == "" {
		return []model.SalesDailyScore{}, nil
	}

	rankedScores := make([]model.SalesDailyScore, 0, len(cleanUserIDs))
	err = r.db.WithContext(ctx).
		Table("sales_daily_scores").
		Select("user_id", "total_score", "score_reached_at").
		Where("score_date = ?", latest.ScoreDate).
		Where("user_id IN ?", cleanUserIDs).
		Order(salesDailyScoreOrderClause("")).
		Scan(&rankedScores).Error
	if err != nil {
		return nil, err
	}
	return rankedScores, nil
}

func (r *gormCustomerRepository) FindLatestAutoAssignOwnerUserID(ctx context.Context, ownerUserIDs []int64, since time.Time) (*int64, error) {
	cleanOwnerUserIDs := uniquePositiveInt64(ownerUserIDs)
	if len(cleanOwnerUserIDs) == 0 {
		return nil, nil
	}

	type latestOwnerRow struct {
		ToOwnerUserID sql.NullInt64 `gorm:"column:to_owner_user_id"`
	}

	query := r.db.WithContext(ctx).
		Table("customer_owner_logs").
		Select("to_owner_user_id").
		Where("to_owner_user_id IN ?", cleanOwnerUserIDs).
		Where("reason IN ?", []string{
			model.CustomerOwnerLogReasonInsideSalesCreate,
			model.CustomerOwnerLogReasonInsideSalesClaim,
			model.CustomerOwnerLogReasonInsideSalesConvert,
		})
	if !since.IsZero() {
		query = query.Where("created_at >= ?", since)
	}

	var row latestOwnerRow
	if err := query.Order("created_at DESC, id DESC").Limit(1).Scan(&row).Error; err != nil {
		return nil, err
	}
	if !row.ToOwnerUserID.Valid || row.ToOwnerUserID.Int64 <= 0 {
		return nil, nil
	}
	ownerUserID := row.ToOwnerUserID.Int64
	return &ownerUserID, nil
}

func (r *gormCustomerRepository) ListRecentContractExemptOwnerUserIDs(ctx context.Context, since time.Time, userIDs []int64) ([]int64, error) {
	cleanUserIDs := uniquePositiveInt64(userIDs)
	if len(cleanUserIDs) == 0 {
		return []int64{}, nil
	}

	type exemptOwnerRow struct {
		UserID           int64     `gorm:"column:user_id"`
		LatestContractAt time.Time `gorm:"column:latest_contract_at"`
	}

	rows := make([]exemptOwnerRow, 0, len(cleanUserIDs))
	query := r.db.WithContext(ctx).
		Table("contracts").
		Select("user_id, MAX(created_at) AS latest_contract_at").
		Where("user_id IN ?", cleanUserIDs)
	if !since.IsZero() {
		query = query.Where("created_at >= ?", since)
	}
	if err := query.
		Group("user_id").
		Order("latest_contract_at DESC, user_id ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row.UserID <= 0 {
			continue
		}
		result = append(result, row.UserID)
	}
	return result, nil
}

func (r *gormCustomerRepository) GetActiveBlockedUntilByDepartmentAnchor(ctx context.Context, customerID, departmentAnchorUserID int64, now time.Time) (*time.Time, error) {
	if customerID <= 0 || departmentAnchorUserID <= 0 {
		return nil, nil
	}

	type blockedUntilRow struct {
		BlockedDepartmentAnchorUserID sql.NullInt64 `gorm:"column:blocked_department_anchor_user_id"`
		BlockedUntil                  sql.NullTime  `gorm:"column:blocked_until"`
		OperatorUserID                sql.NullInt64 `gorm:"column:operator_user_id"`
	}

	rows := make([]blockedUntilRow, 0, 4)
	err := r.db.WithContext(ctx).
		Table("customer_owner_logs").
		Select("blocked_department_anchor_user_id", "blocked_until", "operator_user_id").
		Where("customer_id = ?", customerID).
		Where("action = ?", "release").
		Where("reason = ?", model.CustomerOwnerLogReasonManualRelease).
		Where("blocked_until IS NOT NULL").
		Where("blocked_until > ?", now).
		Order("blocked_until ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	resolvedOperatorAnchors := make(map[int64]int64)
	for _, row := range rows {
		if !row.BlockedUntil.Valid {
			continue
		}
		if row.BlockedDepartmentAnchorUserID.Valid && row.BlockedDepartmentAnchorUserID.Int64 == departmentAnchorUserID {
			blockedUntil := row.BlockedUntil.Time
			return &blockedUntil, nil
		}
		if !row.OperatorUserID.Valid || row.OperatorUserID.Int64 <= 0 {
			continue
		}

		operatorAnchorUserID, ok := resolvedOperatorAnchors[row.OperatorUserID.Int64]
		if !ok {
			operatorAnchorUserID, err = r.resolveSalesDirectorAnchorUserID(ctx, row.OperatorUserID.Int64)
			if err != nil {
				return nil, err
			}
			resolvedOperatorAnchors[row.OperatorUserID.Int64] = operatorAnchorUserID
		}
		if operatorAnchorUserID == departmentAnchorUserID {
			blockedUntil := row.BlockedUntil.Time
			return &blockedUntil, nil
		}
	}
	return nil, nil
}

func (r *gormCustomerRepository) resolveClaimBlockInfo(ctx context.Context, ownerUserID int64, now time.Time) (*int64, *time.Time, error) {
	if ownerUserID <= 0 {
		return nil, nil, nil
	}

	anchorUserID, err := r.resolveSalesDirectorAnchorUserID(ctx, ownerUserID)
	if err != nil {
		return nil, nil, err
	}
	if anchorUserID <= 0 {
		return nil, nil, nil
	}

	var settingValue string
	if err := r.db.WithContext(ctx).
		Table("system_settings").
		Select("value").
		Where("`key` = ?", "claim_freeze_days").
		Limit(1).
		Scan(&settingValue).Error; err != nil {
		return nil, nil, err
	}

	freezeDays := defaultClaimFreezeDays
	if trimmed := strings.TrimSpace(settingValue); trimmed != "" {
		if value, convErr := strconv.Atoi(trimmed); convErr == nil {
			freezeDays = value
		}
	}
	if freezeDays <= 0 {
		return nil, nil, nil
	}

	blockedUntil := now.Add(time.Duration(freezeDays) * 24 * time.Hour)
	return &anchorUserID, &blockedUntil, nil
}

func (r *gormCustomerRepository) resolveSalesDirectorAnchorUserID(ctx context.Context, userID int64) (int64, error) {
	if userID <= 0 {
		return 0, nil
	}

	visited := map[int64]struct{}{}
	currentID := userID

	for currentID > 0 {
		if _, seen := visited[currentID]; seen {
			return 0, nil
		}
		visited[currentID] = struct{}{}

		roleName, err := r.GetUserRoleName(ctx, currentID)
		if err != nil {
			return 0, err
		}
		if isSalesDirectorRoleName(roleName) {
			return currentID, nil
		}

		parentID, err := r.GetParentUserID(ctx, currentID)
		if err != nil {
			return 0, err
		}
		currentID = parentID
	}

	return 0, nil
}

func isSalesDirectorRoleName(roleName string) bool {
	normalized := strings.TrimSpace(roleName)
	return strings.EqualFold(normalized, "sales_director") || normalized == "销售总监"
}

func (r *gormCustomerRepository) CountOwnedActiveByOwner(ctx context.Context, ownerUserID int64) (int64, error) {
	if ownerUserID <= 0 {
		return 0, nil
	}

	var total int64
	err := r.db.WithContext(ctx).
		Table("customers").
		Where("owner_user_id = ?", ownerUserID).
		Where("status = ?", model.CustomerStatusOwned).
		Where("NOT EXISTS (SELECT 1 FROM contracts ct WHERE ct.customer_id = customers.id)").
		Where("(delete_time IS NULL OR delete_time = 0)").
		Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

func buildCustomerListOrderBy(filter model.CustomerListFilter) string {
	if filter.Category != "pool" {
		return "c.updated_at DESC, c.id DESC"
	}

	if strings.TrimSpace(filter.SortBy) == "" {
		return "COALESCE(c.drop_time, 0) DESC, COALESCE(c.follow_time, 0) DESC, c.updated_at DESC, c.id DESC"
	}

	switch normalizePoolSortBy(filter.SortBy) {
	case "follow_time":
		return "COALESCE(c.follow_time, 0) DESC, COALESCE(c.drop_time, 0) DESC, c.updated_at DESC, c.id DESC"
	case "updated_at":
		return "c.updated_at DESC, c.id DESC"
	default:
		return "COALESCE(c.drop_time, 0) DESC, COALESCE(c.follow_time, 0) DESC, c.updated_at DESC, c.id DESC"
	}
}

func normalizePoolSortBy(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "followtime", "follow_time", "follow", "跟进", "跟进时间":
		return "follow_time"
	case "updatedat", "updated_at", "updated", "更新时间":
		return "updated_at"
	default:
		return "drop_time"
	}
}

func buildMyCustomerOwnershipCondition(filter model.CustomerListFilter) (string, []interface{}) {
	ownerIDs := uniquePositiveInt64(filter.AllowedOwnerUserIDs)
	insideSalesIDs := uniquePositiveInt64(filter.AllowedInsideSalesUserIDs)
	hasExplicitScope := len(filter.AllowedOwnerUserIDs) > 0 || len(filter.AllowedInsideSalesUserIDs) > 0

	if hasExplicitScope {
		switch {
		case len(ownerIDs) > 0 && len(insideSalesIDs) > 0:
			return "(c.owner_user_id IN ? OR c.inside_sales_user_id IN ?)", []interface{}{ownerIDs, insideSalesIDs}
		case len(ownerIDs) > 0:
			return "c.owner_user_id IN ?", []interface{}{ownerIDs}
		case len(insideSalesIDs) > 0:
			return "c.inside_sales_user_id IN ?", []interface{}{insideSalesIDs}
		default:
			return "1 = 0", nil
		}
	}

	return "c.owner_user_id = ?", []interface{}{filter.ViewerID}
}

func buildMyCustomerPendingConvertCondition(filter model.CustomerListFilter) (string, []interface{}) {
	if !filter.IncludePendingConvertScope {
		return "", nil
	}
	insideSalesIDs := uniquePositiveInt64(filter.AllowedInsideSalesUserIDs)
	if len(insideSalesIDs) == 0 {
		return "", nil
	}
	return `(c.inside_sales_user_id IN ? AND c.create_user_id IN ? AND c.converted_at IS NULL AND (c.owner_user_id IS NULL OR c.status = 'pool'))`,
		[]interface{}{insideSalesIDs, insideSalesIDs}
}

func buildCustomerListWhere(filter model.CustomerListFilter) ([]string, []interface{}) {
	var where []string
	var args []interface{}

	if filter.Keyword != "" {
		pattern := "%" + filter.Keyword + "%"
		where = append(where, "(c.name LIKE ? OR c.email LIKE ? OR COALESCE(u.nickname, '') LIKE ? OR EXISTS (SELECT 1 FROM customer_phones WHERE customer_id = c.id AND phone LIKE ?))")
		args = append(args, pattern, pattern, pattern, pattern)
	}
	if filter.Name != "" {
		where = append(where, "c.name LIKE ?")
		args = append(args, "%"+filter.Name+"%")
	}
	if filter.ContactName != "" {
		where = append(where, "COALESCE(c.contact_name, '') LIKE ?")
		args = append(args, "%"+filter.ContactName+"%")
	}
	if filter.Phone != "" {
		where = append(where, "EXISTS (SELECT 1 FROM customer_phones p WHERE p.customer_id = c.id AND p.phone LIKE ?)")
		args = append(args, "%"+filter.Phone+"%")
	}
	if filter.Weixin != "" {
		where = append(where, "COALESCE(c.weixin, '') LIKE ?")
		args = append(args, "%"+filter.Weixin+"%")
	}
	if filter.OwnerUserID > 0 {
		if filter.Category == "pool" {
			where = append(where, "du.id = ?")
		} else {
			where = append(where, "c.owner_user_id = ?")
		}
		args = append(args, filter.OwnerUserID)
	}
	if filter.OwnerUserName != "" {
		if filter.Category == "pool" {
			where = append(where, "COALESCE(NULLIF(du.nickname, ''), NULLIF(du.username, ''), '') LIKE ?")
		} else {
			where = append(where, "COALESCE(u.nickname, '') LIKE ?")
		}
		args = append(args, "%"+filter.OwnerUserName+"%")
	}
	if filter.Province > 0 {
		where = append(where, "c.province = ?")
		args = append(args, filter.Province)
	}
	if filter.City > 0 {
		where = append(where, "c.city = ?")
		args = append(args, filter.City)
	}
	if filter.Area > 0 {
		where = append(where, "c.area = ?")
		args = append(args, filter.Area)
	}
	if filter.ExcludePool {
		where = append(where, "c.status != 'pool'")
	}

	switch filter.Category {
	case "my":
		where = append(where, "NOT EXISTS (SELECT 1 FROM contracts ct WHERE ct.customer_id = c.id)")
		if filter.RequireInsideSalesAssociation {
			where = append(where, "c.inside_sales_user_id IS NOT NULL")
		}
		if filter.HasViewer {
			if filter.SkipViewerOwnerLimit {
				if !filter.IncludePoolInMyScope {
					where = append(where, "c.status != 'pool'")
				}
				break
			}

			ownershipCondition, ownershipArgs := buildMyCustomerOwnershipCondition(filter)
			pendingConvertCondition, pendingConvertArgs := buildMyCustomerPendingConvertCondition(filter)

			switch {
			case pendingConvertCondition != "" && ownershipCondition == "1 = 0":
				where = append(where, pendingConvertCondition)
				args = append(args, pendingConvertArgs...)
			case pendingConvertCondition != "":
				where = append(where, "((c.status != 'pool' AND ("+ownershipCondition+")) OR "+pendingConvertCondition+")")
				args = append(args, ownershipArgs...)
				args = append(args, pendingConvertArgs...)
			default:
				if !filter.IncludePoolInMyScope {
					where = append(where, "c.status != 'pool'")
				}
				where = append(where, ownershipCondition)
				args = append(args, ownershipArgs...)
			}
		} else {
			where = append(where, "1 = 0")
		}
	case "pool":
		where = append(where, "(c.owner_user_id IS NULL OR c.status = 'pool')")
	case "potential":
		if filter.HasViewer {
			where = append(where, "c.deal_status = 'undone'")
			where = append(where, "(c.owner_user_id IS NULL OR c.status = 'pool')")
			where = append(where, `EXISTS (
				SELECT 1 FROM customer_owner_logs l
				WHERE l.customer_id = c.id AND l.to_owner_user_id = ?
			)`)
			args = append(args, filter.ViewerID)
		} else {
			where = append(where, "1 = 0")
		}
	case "partner":
		if filter.HasViewer {
			where = append(where, "EXISTS (SELECT 1 FROM contracts ct WHERE ct.customer_id = c.id)")
			where = append(where, "c.status != 'pool'")
			switch {
			case filter.ForceServiceUserID != nil:
				where = append(where, "EXISTS (SELECT 1 FROM contracts ct WHERE ct.customer_id = c.id AND ct.service_user_id = ?)")
				args = append(args, *filter.ForceServiceUserID)
			case len(filter.AllowedServiceUserIDs) > 0:
				serviceUserIDs := uniquePositiveInt64(filter.AllowedServiceUserIDs)
				if len(serviceUserIDs) == 0 {
					where = append(where, "1 = 0")
				} else {
					where = append(where, "EXISTS (SELECT 1 FROM contracts ct WHERE ct.customer_id = c.id AND ct.service_user_id IN ?)")
					args = append(args, serviceUserIDs)
				}
			case len(filter.AllowedOwnerUserIDs) > 0:
				ownerUserIDs := uniquePositiveInt64(filter.AllowedOwnerUserIDs)
				if len(ownerUserIDs) == 0 {
					where = append(where, "1 = 0")
				} else {
					where = append(where, "c.owner_user_id IN ?")
					args = append(args, ownerUserIDs)
				}
			default:
				where = append(where, "c.owner_user_id = ?")
				args = append(args, filter.ViewerID)
			}
		} else {
			where = append(where, "1 = 0")
		}
	case "search", "":
		// no extra condition
	default:
		where = append(where, "1 = 0")
	}

	return where, args
}

func (r *gormCustomerRepository) Create(ctx context.Context, input model.CustomerCreateInput) (*model.Customer, error) {
	type customerCreateRow struct {
		ID                int64      `gorm:"column:id;primaryKey;autoIncrement"`
		Name              string     `gorm:"column:name"`
		LegalName         string     `gorm:"column:legal_name"`
		ContactName       string     `gorm:"column:contact_name"`
		Weixin            string     `gorm:"column:weixin"`
		Email             string     `gorm:"column:email"`
		CustomerLevelID   int        `gorm:"column:customer_level_id"`
		CustomerSourceID  int        `gorm:"column:customer_source_id"`
		Province          int        `gorm:"column:province"`
		City              int        `gorm:"column:city"`
		Area              int        `gorm:"column:area"`
		DetailAddress     string     `gorm:"column:detail_address"`
		Remark            string     `gorm:"column:remark"`
		Status            string     `gorm:"column:status"`
		DealStatus        string     `gorm:"column:deal_status"`
		OwnerUserID       *int64     `gorm:"column:owner_user_id"`
		InsideSalesUserID *int64     `gorm:"column:inside_sales_user_id"`
		ConvertedAt       *time.Time `gorm:"column:converted_at"`
		CreateUserID      int64      `gorm:"column:create_user_id"`
		OperateUserID     int64      `gorm:"column:operate_user_id"`
		CollectTime       *int64     `gorm:"column:collect_time"`
		AssignTime        *int64     `gorm:"column:assign_time"`
		NextTime          int64      `gorm:"column:next_time"`
		CreatedAt         time.Time  `gorm:"column:created_at"`
		UpdatedAt         time.Time  `gorm:"column:updated_at"`
	}

	tx := r.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return nil, err
	}
	defer tx.Rollback()

	status := input.Status
	if status != model.CustomerStatusOwned && status != model.CustomerStatusPool {
		status = model.CustomerStatusPool
	}

	var ownerUserID *int64
	if status == model.CustomerStatusOwned {
		if input.OwnerUserID != nil {
			ownerUserID = input.OwnerUserID
		} else if input.OperatorUserID > 0 {
			defaultOwnerID := input.OperatorUserID
			ownerUserID = &defaultOwnerID
		}
	}

	now := time.Now().UTC()
	var collectTime *int64
	var assignTime *int64
	nextTime := int64(0)
	if ownerUserID != nil {
		collectAt := now.Unix()
		collectTime = &collectAt
		nextTime = collectAt
		assignTime = assignedSalesUnix(input.InsideSalesUserID, *ownerUserID, now)
	}

	row := customerCreateRow{
		Name:              input.Name,
		LegalName:         input.LegalName,
		ContactName:       input.ContactName,
		Weixin:            input.Weixin,
		Email:             input.Email,
		CustomerLevelID:   0,
		CustomerSourceID:  0,
		Province:          input.Province,
		City:              input.City,
		Area:              input.Area,
		DetailAddress:     input.DetailAddress,
		Remark:            input.Remark,
		Status:            status,
		DealStatus:        model.CustomerDealStatusUndone,
		OwnerUserID:       ownerUserID,
		InsideSalesUserID: input.InsideSalesUserID,
		ConvertedAt:       input.ConvertedAt,
		CreateUserID:      input.OperatorUserID,
		OperateUserID:     input.OperatorUserID,
		CollectTime:       collectTime,
		AssignTime:        assignTime,
		NextTime:          nextTime,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := tx.Table("customers").Create(&row).Error; err != nil {
		return nil, err
	}
	customerID := row.ID

	if len(input.Phones) > 0 {
		if err := r.replacePhonesTx(tx, customerID, input.Phones); err != nil {
			return nil, err
		}
	}

	if ownerUserID != nil {
		reason := model.CustomerOwnerLogReasonCreateInitialAssign
		content := "创建客户后直接分配负责人"
		if input.InsideSalesUserID != nil && *input.InsideSalesUserID > 0 {
			reason = model.CustomerOwnerLogReasonInsideSalesCreate
			if *ownerUserID == *input.InsideSalesUserID && input.ConvertedAt == nil {
				content = "电销创建客户后暂未分配负责人，客户先归属电销本人"
			} else {
				content = "电销创建客户后按分配规则自动分配负责人"
			}
		}
		logRow := newCustomerOwnerLogRow(
			customerID,
			nil,
			ownerUserID,
			"claim",
			reason,
			content,
			input.OperatorUserID,
			now,
		)
		if err := tx.Table("customer_owner_logs").Create(&logRow).Error; err != nil {
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return r.getCustomer(ctx, customerID)
}

func (r *gormCustomerRepository) Update(ctx context.Context, customerID int64, input model.CustomerUpdateInput) (*model.Customer, error) {
	tx := r.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err := r.getCustomerForUpdate(tx, customerID); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if err := tx.Table("customers").
		Where("id = ?", customerID).
		Updates(map[string]interface{}{
			"name":            input.Name,
			"legal_name":      input.LegalName,
			"contact_name":    input.ContactName,
			"weixin":          input.Weixin,
			"email":           input.Email,
			"province":        input.Province,
			"city":            input.City,
			"area":            input.Area,
			"detail_address":  input.DetailAddress,
			"remark":          input.Remark,
			"operate_user_id": input.OperatorUserID,
			"updated_at":      now,
		}).Error; err != nil {
		return nil, err
	}

	if len(input.Phones) > 0 {
		if err := r.replacePhonesTx(tx, customerID, input.Phones); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return r.getCustomer(ctx, customerID)
}

func (r *gormCustomerRepository) CheckUnique(ctx context.Context, input model.CustomerUniqueCheckInput) (model.CustomerUniqueCheckResult, error) {
	result := model.CustomerUniqueCheckResult{
		DuplicatePhones: []string{},
	}

	name, err := r.existsCustomerField(ctx, "name", input.Name, input.ExcludeCustomerID)
	if err != nil {
		return result, err
	}
	result.NameExists = name

	weixin, err := r.existsCustomerField(ctx, "weixin", input.Weixin, input.ExcludeCustomerID)
	if err != nil {
		return result, err
	}
	result.WeixinExists = weixin

	duplicatePhones, err := r.findDuplicatePhones(ctx, input.Phones, input.ExcludeCustomerID)
	if err != nil {
		return result, err
	}
	result.DuplicatePhones = duplicatePhones

	return result, nil
}

func (r *gormCustomerRepository) Claim(ctx context.Context, customerID, ownerUserID, operatorUserID int64, insideSalesUserID *int64) (*model.Customer, error) {
	tx := r.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return nil, err
	}
	defer tx.Rollback()

	customer, err := r.getCustomerForUpdate(tx, customerID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	updates := map[string]interface{}{
		"owner_user_id": ownerUserID,
		"status":        "owned",
		"collect_time":  now.Unix(),
		"assign_time":   nil,
		"follow_time":   nil,
		"next_time":     now.Unix(),
		"drop_time":     nil,
		"updated_at":    now,
	}
	if insideSalesUserID != nil && *insideSalesUserID > 0 {
		updates["inside_sales_user_id"] = *insideSalesUserID
		updates["assign_time"] = assignedSalesUnix(insideSalesUserID, ownerUserID, now)
		if ownerUserID != *insideSalesUserID {
			updates["converted_at"] = now
		} else {
			updates["converted_at"] = nil
		}
	}
	if err := tx.Table("customers").
		Where("id = ?", customerID).
		Updates(updates).Error; err != nil {
		return nil, err
	}

	toOwnerID := ownerUserID
	logContent := "从公海领取客户"
	logReason := model.CustomerOwnerLogReasonClaimFromPool
	if insideSalesUserID != nil && *insideSalesUserID > 0 {
		if ownerUserID == *insideSalesUserID {
			logContent = "电销从公海领取客户后暂未分配负责人，客户先归属电销本人"
		} else {
			logContent = "电销从公海领取客户后自动分配负责人"
		}
		logReason = model.CustomerOwnerLogReasonInsideSalesClaim
	}
	logRow := newCustomerOwnerLogRow(
		customerID,
		customer.OwnerUserID,
		&toOwnerID,
		"claim",
		logReason,
		logContent,
		operatorUserID,
		now,
	)
	if err := tx.Table("customer_owner_logs").Create(&logRow).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return r.getCustomer(ctx, customerID)
}

func (r *gormCustomerRepository) Convert(ctx context.Context, customerID, ownerUserID, operatorUserID int64) (*model.Customer, error) {
	tx := r.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return nil, err
	}
	defer tx.Rollback()

	customer, err := r.getCustomerForUpdate(tx, customerID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	insideSalesUserID := customer.InsideSalesUserID
	if insideSalesUserID == nil && customer.CreateUserID > 0 {
		fallbackInsideSalesUserID := customer.CreateUserID
		insideSalesUserID = &fallbackInsideSalesUserID
	}
	var insideSalesValue interface{}
	if insideSalesUserID != nil && *insideSalesUserID > 0 {
		insideSalesValue = *insideSalesUserID
	}
	convertedAtValue := interface{}(nil)
	if insideSalesUserID == nil || *insideSalesUserID <= 0 || ownerUserID != *insideSalesUserID {
		convertedAtValue = now
	}
	collectTimeValue := collectTimeUnixOrNow(customer.CollectTime, now)
	assignTimeValue := interface{}(nil)
	if assignedAt := assignedSalesUnix(insideSalesUserID, ownerUserID, now); assignedAt != nil {
		assignTimeValue = *assignedAt
	}
	if err := tx.Table("customers").
		Where("id = ?", customerID).
		Updates(map[string]interface{}{
			"owner_user_id":        ownerUserID,
			"inside_sales_user_id": insideSalesValue,
			"converted_at":         convertedAtValue,
			"status":               "owned",
			"collect_time":         collectTimeValue,
			"assign_time":          assignTimeValue,
			"follow_time":          nil,
			"next_time":            now.Unix(),
			"drop_time":            nil,
			"drop_user_id":         nil,
			"updated_at":           now,
		}).Error; err != nil {
		return nil, err
	}

	toOwnerID := ownerUserID
	action := "claim"
	if customer.OwnerUserID != nil && *customer.OwnerUserID > 0 {
		action = "transfer"
	}
	logRow := newCustomerOwnerLogRow(
		customerID,
		customer.OwnerUserID,
		&toOwnerID,
		action,
		model.CustomerOwnerLogReasonInsideSalesConvert,
		func() string {
			if insideSalesUserID != nil && *insideSalesUserID > 0 && ownerUserID == *insideSalesUserID {
				return "电销转化客户时暂无评分可用，客户先归属电销本人"
			}
			return "电销转化客户后按分配规则完成分配"
		}(),
		operatorUserID,
		now,
	)
	if err := tx.Table("customer_owner_logs").Create(&logRow).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return r.getCustomer(ctx, customerID)
}

func (r *gormCustomerRepository) Release(ctx context.Context, customerID, operatorUserID int64) (*model.Customer, error) {
	tx := r.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return nil, err
	}
	defer tx.Rollback()

	customer, err := r.getCustomerForUpdate(tx, customerID)
	if err != nil {
		return nil, err
	}
	if customer.IsInPool {
		return nil, ErrCustomerAlreadyInPool
	}
	if customer.OwnerUserID == nil || *customer.OwnerUserID != operatorUserID {
		return nil, ErrCustomerNotOwned
	}

	now := time.Now().UTC()
	blockedDepartmentAnchorUserID, blockedUntil, err := r.resolveClaimBlockInfo(ctx, operatorUserID, now)
	if err != nil {
		return nil, err
	}
	if err := tx.Table("customers").
		Where("id = ?", customerID).
		Updates(map[string]interface{}{
			"owner_user_id": nil,
			"status":        "pool",
			"assign_time":   nil,
			"drop_time":     now.Unix(),
			"drop_user_id":  operatorUserID,
			"updated_at":    now,
		}).Error; err != nil {
		return nil, err
	}

	logRow := newCustomerOwnerLogRow(
		customerID,
		customer.OwnerUserID,
		nil,
		"release",
		model.CustomerOwnerLogReasonManualRelease,
		"手动丢弃客户回公海",
		operatorUserID,
		now,
	)
	logRow.BlockedDepartmentAnchorUserID = blockedDepartmentAnchorUserID
	logRow.BlockedUntil = blockedUntil
	if err := tx.Table("customer_owner_logs").Create(&logRow).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return r.getCustomer(ctx, customerID)
}

func (r *gormCustomerRepository) Transfer(ctx context.Context, input model.CustomerTransferInput) (*model.Customer, error) {
	tx := r.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return nil, err
	}
	defer tx.Rollback()

	customer, err := r.getCustomerForUpdate(tx, input.CustomerID)
	if err != nil {
		return nil, err
	}
	if customer.IsInPool || customer.OwnerUserID == nil {
		return nil, ErrCustomerNotOwned
	}
	if !input.AllowAnyOwner && *customer.OwnerUserID != input.OperatorUserID {
		return nil, ErrCustomerNotOwned
	}

	now := time.Now().UTC()
	updates := map[string]interface{}{
		"owner_user_id": input.ToOwnerUserID,
		"status":        "owned",
		"collect_time":  now.Unix(),
		"next_time":     now.Unix(),
		"follow_time":   nil,
		"drop_time":     nil,
		"updated_at":    now,
	}
	if customer.InsideSalesUserID != nil && *customer.InsideSalesUserID > 0 {
		switch {
		case input.ToOwnerUserID == *customer.InsideSalesUserID:
			updates["assign_time"] = nil
		case customer.OwnerUserID != nil && *customer.OwnerUserID == *customer.InsideSalesUserID:
			updates["collect_time"] = collectTimeUnixOrNow(customer.CollectTime, now)
			updates["assign_time"] = now.Unix()
		}
	}
	if err := tx.Table("customers").
		Where("id = ?", input.CustomerID).
		Updates(updates).Error; err != nil {
		return nil, err
	}

	toOwnerID := input.ToOwnerUserID
	logRow := newCustomerOwnerLogRow(
		input.CustomerID,
		customer.OwnerUserID,
		&toOwnerID,
		"transfer",
		model.CustomerOwnerLogReasonManualTransfer,
		customerOwnerLogContent(input.Note, "手动转移客户"),
		input.OperatorUserID,
		now,
	)
	if err := tx.Table("customer_owner_logs").Create(&logRow).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return r.getCustomer(ctx, input.CustomerID)
}

func (r *gormCustomerRepository) getCustomerForUpdate(tx *gorm.DB, customerID int64) (*model.Customer, error) {
	type customerLockRow struct {
		ID                int64         `gorm:"column:id"`
		Name              string        `gorm:"column:name"`
		Email             string        `gorm:"column:email"`
		Status            string        `gorm:"column:status"`
		DealStatus        string        `gorm:"column:deal_status"`
		CreateUserID      int64         `gorm:"column:create_user_id"`
		CollectTime       sql.NullInt64 `gorm:"column:collect_time"`
		InsideSalesUserID sql.NullInt64 `gorm:"column:inside_sales_user_id"`
		ConvertedAt       sql.NullTime  `gorm:"column:converted_at"`
		OwnerUser         sql.NullInt64 `gorm:"column:owner_user_id"`
		DropUserID        sql.NullInt64 `gorm:"column:drop_user_id"`
		CreatedAt         time.Time     `gorm:"column:created_at"`
		UpdatedAt         time.Time     `gorm:"column:updated_at"`
	}

	var row customerLockRow
	if err := tx.Table("customers").
		Select(`
			id,
			name,
			email,
			status,
			deal_status,
			create_user_id,
			NULLIF(collect_time, 0) AS collect_time,
			inside_sales_user_id,
			converted_at,
			owner_user_id,
			COALESCE(
				drop_user_id,
				(
					SELECT COALESCE(l.from_owner_user_id, l.operator_user_id)
					FROM customer_owner_logs l
					WHERE l.customer_id = customers.id
						AND l.action = 'release'
					ORDER BY l.created_at DESC, l.id DESC
					LIMIT 1
				)
			) AS drop_user_id,
			created_at,
			updated_at
		`).
		Where("id = ?", customerID).
		Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}

	customer := &model.Customer{
		ID:           row.ID,
		Name:         row.Name,
		Email:        row.Email,
		Status:       row.Status,
		DealStatus:   row.DealStatus,
		CreateUserID: row.CreateUserID,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
		IsInPool:     !row.OwnerUser.Valid || row.Status == model.CustomerStatusPool,
	}
	if row.InsideSalesUserID.Valid {
		customer.InsideSalesUserID = &row.InsideSalesUserID.Int64
	}
	customer.CollectTime = nullableUnixToTime(row.CollectTime)
	if row.ConvertedAt.Valid {
		convertedAt := row.ConvertedAt.Time
		customer.ConvertedAt = &convertedAt
	}
	if row.OwnerUser.Valid {
		customer.OwnerUserID = &row.OwnerUser.Int64
	}
	if row.DropUserID.Valid {
		customer.DropUserID = &row.DropUserID.Int64
	}
	return customer, nil
}

func (r *gormCustomerRepository) getCustomer(ctx context.Context, customerID int64) (*model.Customer, error) {
	var row customerListRow
	err := r.db.WithContext(ctx).
		Table("customers AS c").
		Select(`
			c.id AS id,
			c.name AS name,
			COALESCE(c.legal_name, '') AS legal_name,
			COALESCE(c.contact_name, '') AS contact_name,
			COALESCE(c.weixin, '') AS weixin,
			COALESCE(c.email, '') AS email,
			COALESCE(c.customer_level_id, 0) AS customer_level_id,
			COALESCE(c.customer_source_id, 0) AS customer_source_id,
			COALESCE(cl.name, '') AS customer_level_name,
			COALESCE(cs.name, '') AS customer_source_name,
			COALESCE(c.province, 0) AS province,
			COALESCE(c.city, 0) AS city,
			COALESCE(c.area, 0) AS area,
			COALESCE(c.detail_address, '') AS detail_address,
			COALESCE(c.remark, '') AS remark,
			c.status AS status,
			CASE
				WHEN EXISTS (SELECT 1 FROM contracts ct WHERE ct.customer_id = c.id) THEN 'done'
				ELSE 'undone'
			END AS deal_status,
			c.create_user_id AS create_user_id,
			c.inside_sales_user_id AS inside_sales_user_id,
			c.converted_at AS converted_at,
			c.owner_user_id AS owner_user_id,
			COALESCE(u.nickname, '') AS owner_user_name,
			COALESCE(NULLIF(du.nickname, ''), NULLIF(du.username, ''), '') AS drop_user_name,
			c.created_at AS created_at,
			c.updated_at AS updated_at,
			NULLIF(c.next_time, 0) AS next_time_unix,
			NULLIF(c.follow_time, 0) AS follow_time_unix,
			NULLIF(c.collect_time, 0) AS collect_time_unix,
			NULLIF(c.assign_time, 0) AS assign_time_unix,
			NULLIF(c.drop_time, 0) AS drop_time_unix,
			COALESCE(
				c.drop_user_id,
					(
						SELECT COALESCE(l.from_owner_user_id, l.operator_user_id)
						FROM customer_owner_logs l
						WHERE l.customer_id = c.id
							AND l.action = 'release'
						ORDER BY l.created_at DESC, l.id DESC
						LIMIT 1
					)
				) AS drop_user_id,
				CASE WHEN c.owner_user_id IS NULL OR c.status = 'pool' THEN 1 ELSE 0 END AS is_in_pool
		`).
		Joins("LEFT JOIN users u ON c.owner_user_id = u.id").
		Joins(`LEFT JOIN users du ON du.id = COALESCE(
			c.drop_user_id,
			(
				SELECT COALESCE(l.from_owner_user_id, l.operator_user_id)
				FROM customer_owner_logs l
				WHERE l.customer_id = c.id
					AND l.action = 'release'
				ORDER BY l.created_at DESC, l.id DESC
				LIMIT 1
			)
		)`).
		Joins("LEFT JOIN customer_levels cl ON c.customer_level_id = cl.id").
		Joins("LEFT JOIN customer_sources cs ON c.customer_source_id = cs.id").
		Where("c.id = ?", customerID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}

	customer := model.Customer{
		ID:                 row.ID,
		Name:               row.Name,
		LegalName:          row.LegalName,
		ContactName:        row.ContactName,
		Weixin:             row.Weixin,
		Email:              row.Email,
		CustomerLevelID:    row.CustomerLevelID,
		CustomerSourceID:   row.CustomerSourceID,
		CustomerLevelName:  row.CustomerLevelName,
		CustomerSourceName: row.CustomerSourceName,
		Province:           row.Province,
		City:               row.City,
		Area:               row.Area,
		DetailAddress:      row.DetailAddress,
		Remark:             row.Remark,
		Status:             row.Status,
		DealStatus:         row.DealStatus,
		CreateUserID:       row.CreateUserID,
		OwnerUserName:      row.OwnerUserName,
		DropUserName:       row.DropUserName,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
		IsInPool:           row.IsInPool,
	}
	if row.InsideSalesUserID.Valid {
		customer.InsideSalesUserID = &row.InsideSalesUserID.Int64
	}
	if row.ConvertedAt.Valid {
		convertedAt := row.ConvertedAt.Time
		customer.ConvertedAt = &convertedAt
	}
	if row.OwnerUserID.Valid {
		customer.OwnerUserID = &row.OwnerUserID.Int64
	}
	customer.NextTime = nullableUnixToTime(row.NextTimeUnix)
	customer.FollowTime = nullableUnixToTime(row.FollowTimeUnix)
	customer.CollectTime = nullableUnixToTime(row.CollectTimeUnix)
	customer.AssignTime = nullableUnixToTime(row.AssignTimeUnix)
	customer.DropTime = nullableUnixToTime(row.DropTimeUnix)
	if row.DropUserID.Valid {
		customer.DropUserID = &row.DropUserID.Int64
	}

	history, err := r.listHistoricalOwnerIDs(ctx, customer.ID)
	if err != nil {
		return nil, err
	}
	customer.HistoricalOwnerIDs = history
	phones, err := r.ListPhones(ctx, customer.ID)
	if err != nil {
		return nil, err
	}
	customer.Phones = phones
	return &customer, nil
}

func nullableUnixToTime(value sql.NullInt64) *time.Time {
	if !value.Valid || value.Int64 <= 0 {
		return nil
	}
	t := time.Unix(value.Int64, 0).UTC()
	return &t
}

func (r *gormCustomerRepository) listHistoricalOwnerIDs(ctx context.Context, customerID int64) ([]int64, error) {
	var ids []int64
	if err := r.db.WithContext(ctx).
		Table("customer_owner_logs").
		Distinct("to_owner_user_id").
		Where("customer_id = ? AND to_owner_user_id IS NOT NULL", customerID).
		Order("to_owner_user_id").
		Pluck("to_owner_user_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// AddPhone adds a new phone number for a customer
func (r *gormCustomerRepository) AddPhone(ctx context.Context, phone *model.CustomerPhone) error {
	tx := r.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return err
	}
	defer tx.Rollback()

	// If this is a primary phone, unset other primary phones for this customer
	if phone.IsPrimary {
		if err := tx.Table("customer_phones").
			Where("customer_id = ?", phone.CustomerID).
			Update("is_primary", 0).Error; err != nil {
			return err
		}
	}

	now := time.Now().Unix()
	row := customerPhoneRow{
		CustomerID: phone.CustomerID,
		Phone:      phone.Phone,
		PhoneLabel: phone.PhoneLabel,
		IsPrimary:  boolToInt(phone.IsPrimary),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := tx.Table("customer_phones").Create(&row).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrPhoneAlreadyExists
		}
		return err
	}

	phone.ID = row.ID
	phone.CreatedAt = time.Unix(now, 0)
	phone.UpdatedAt = time.Unix(now, 0)

	return tx.Commit().Error
}

// ListPhones retrieves all phone numbers for a customer
func (r *gormCustomerRepository) ListPhones(ctx context.Context, customerID int64) ([]model.CustomerPhone, error) {
	var rows []customerPhoneRow
	if err := r.db.WithContext(ctx).
		Table("customer_phones").
		Select("id", "customer_id", "phone", "COALESCE(phone_label, '') AS phone_label", "is_primary", "created_at", "updated_at").
		Where("customer_id = ?", customerID).
		Order("is_primary DESC, id ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	var phones []model.CustomerPhone
	for _, row := range rows {
		p := model.CustomerPhone{
			ID:         row.ID,
			CustomerID: row.CustomerID,
			Phone:      row.Phone,
			PhoneLabel: row.PhoneLabel,
			IsPrimary:  row.IsPrimary == 1,
			CreatedAt:  time.Unix(row.CreatedAt, 0),
			UpdatedAt:  time.Unix(row.UpdatedAt, 0),
		}
		phones = append(phones, p)
	}
	if phones == nil {
		phones = []model.CustomerPhone{}
	}
	return phones, nil
}

// UpdatePhone updates an existing phone number
func (r *gormCustomerRepository) UpdatePhone(ctx context.Context, phone *model.CustomerPhone) error {
	tx := r.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if phone exists
	existing, err := r.GetPhone(ctx, phone.ID)
	if err != nil {
		return err
	}

	// If setting as primary, unset other primary phones for this customer
	if phone.IsPrimary && !existing.IsPrimary {
		if err := tx.Table("customer_phones").
			Where("customer_id = ?", phone.CustomerID).
			Update("is_primary", 0).Error; err != nil {
			return err
		}
	}

	now := time.Now().Unix()
	if err := tx.Table("customer_phones").
		Where("id = ? AND customer_id = ?", phone.ID, phone.CustomerID).
		Updates(map[string]interface{}{
			"phone":       phone.Phone,
			"phone_label": phone.PhoneLabel,
			"is_primary":  boolToInt(phone.IsPrimary),
			"updated_at":  now,
		}).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrPhoneAlreadyExists
		}
		return err
	}

	phone.UpdatedAt = time.Unix(now, 0)
	return tx.Commit().Error
}

// DeletePhone deletes a phone number
func (r *gormCustomerRepository) DeletePhone(ctx context.Context, customerID, phoneID int64) error {
	result := r.db.WithContext(ctx).
		Table("customer_phones").
		Where("id = ? AND customer_id = ?", phoneID, customerID).
		Delete(nil)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrPhoneNotFound
	}

	return nil
}

// GetPhone retrieves a single phone by ID
func (r *gormCustomerRepository) GetPhone(ctx context.Context, phoneID int64) (*model.CustomerPhone, error) {
	var row customerPhoneRow
	err := r.db.WithContext(ctx).
		Table("customer_phones").
		Select("id", "customer_id", "phone", "COALESCE(phone_label, '') AS phone_label", "is_primary", "created_at", "updated_at").
		Where("id = ?", phoneID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPhoneNotFound
		}
		return nil, err
	}
	p := model.CustomerPhone{
		ID:         row.ID,
		CustomerID: row.CustomerID,
		Phone:      row.Phone,
		PhoneLabel: row.PhoneLabel,
		IsPrimary:  row.IsPrimary == 1,
		CreatedAt:  time.Unix(row.CreatedAt, 0),
		UpdatedAt:  time.Unix(row.UpdatedAt, 0),
	}
	return &p, nil
}

func (r *gormCustomerRepository) FindCustomerIDByPhone(ctx context.Context, phone string) (int64, error) {
	trimmed := strings.TrimSpace(phone)
	if trimmed == "" {
		return 0, nil
	}

	type customerIDRow struct {
		CustomerID int64 `gorm:"column:customer_id"`
	}
	var row customerIDRow
	err := r.db.WithContext(ctx).
		Table("customer_phones").
		Select("customer_id").
		Where("phone = ?", trimmed).
		Order("is_primary DESC, id ASC").
		Limit(1).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return row.CustomerID, nil
}

// CreateStatusLog creates a new status log entry
func (r *gormCustomerRepository) CreateStatusLog(ctx context.Context, log *model.CustomerStatusLog) error {
	now := time.Now().Unix()
	operator := sql.NullInt64{}
	if log.OperatorUserID != nil {
		operator.Valid = true
		operator.Int64 = *log.OperatorUserID
	}
	row := customerStatusLogRow{
		CustomerID:     log.CustomerID,
		FromStatus:     log.FromStatus,
		ToStatus:       log.ToStatus,
		TriggerType:    log.TriggerType,
		Reason:         log.Reason,
		OperatorUserID: operator,
		OperateTime:    now,
	}
	if err := r.db.WithContext(ctx).Table("customer_status_logs").Create(&row).Error; err != nil {
		return err
	}
	log.ID = row.ID
	log.OperateTime = time.Unix(now, 0)
	return nil
}

// ListStatusLogs retrieves status logs for a customer with pagination
func (r *gormCustomerRepository) ListStatusLogs(ctx context.Context, customerID int64, page, pageSize int) ([]model.CustomerStatusLog, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	var rows []customerStatusLogListRow
	if err := r.db.WithContext(ctx).
		Table("customer_status_logs AS l").
		Select(
			"l.id", "l.customer_id", "l.from_status", "l.to_status", "l.trigger_type",
			"COALESCE(l.reason, '') AS reason", "l.operator_user_id", "COALESCE(u.nickname, '') AS operator_name", "l.operate_time",
		).
		Joins("LEFT JOIN users u ON l.operator_user_id = u.id").
		Where("l.customer_id = ?", customerID).
		Order("l.operate_time DESC, l.id DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	var logs []model.CustomerStatusLog
	for _, row := range rows {
		logItem := model.CustomerStatusLog{
			ID:           row.ID,
			CustomerID:   row.CustomerID,
			FromStatus:   row.FromStatus,
			ToStatus:     row.ToStatus,
			TriggerType:  row.TriggerType,
			Reason:       row.Reason,
			OperatorName: row.OperatorName,
			OperateTime:  time.Unix(row.OperateTime, 0),
		}
		if row.OperatorUserID.Valid {
			logItem.OperatorUserID = &row.OperatorUserID.Int64
		}
		logs = append(logs, logItem)
	}
	if logs == nil {
		logs = []model.CustomerStatusLog{}
	}
	return logs, nil
}

func (r *gormCustomerRepository) replacePhonesTx(tx *gorm.DB, customerID int64, phones []model.CustomerPhoneInput) error {
	if err := tx.Table("customer_phones").Where("customer_id = ?", customerID).Delete(nil).Error; err != nil {
		return err
	}
	if len(phones) == 0 {
		return nil
	}

	primaryIndex := -1
	for idx, phone := range phones {
		if phone.IsPrimary {
			primaryIndex = idx
			break
		}
	}
	if primaryIndex < 0 {
		primaryIndex = 0
	}

	now := time.Now().Unix()
	for idx, phone := range phones {
		trimmedPhone := strings.TrimSpace(phone.Phone)
		if trimmedPhone == "" {
			continue
		}
		row := customerPhoneRow{
			CustomerID: customerID,
			Phone:      trimmedPhone,
			PhoneLabel: strings.TrimSpace(phone.PhoneLabel),
			IsPrimary:  boolToInt(idx == primaryIndex),
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if err := tx.Table("customer_phones").Create(&row).Error; err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				return ErrPhoneAlreadyExists
			}
			return err
		}
	}

	return nil
}

func (r *gormCustomerRepository) existsCustomerField(
	ctx context.Context,
	field string,
	value string,
	excludeCustomerID *int64,
) (bool, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false, nil
	}

	switch field {
	case "name", "legal_name", "contact_name", "weixin":
	default:
		return false, fmt.Errorf("unsupported field: %s", field)
	}

	query := r.db.WithContext(ctx).
		Table("customers").
		Select("id").
		Where(fmt.Sprintf("TRIM(%s) = TRIM(?)", field), trimmed)
	if excludeCustomerID != nil {
		query = query.Where("id <> ?", *excludeCustomerID)
	}

	var id int64
	if err := query.Take(&id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *gormCustomerRepository) findDuplicatePhones(
	ctx context.Context,
	phones []string,
	excludeCustomerID *int64,
) ([]string, error) {
	uniquePhones := make([]string, 0, len(phones))
	seen := make(map[string]struct{})
	for _, phone := range phones {
		trimmed := strings.TrimSpace(phone)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		uniquePhones = append(uniquePhones, trimmed)
	}
	if len(uniquePhones) == 0 {
		return []string{}, nil
	}

	query := r.db.WithContext(ctx).
		Table("customer_phones").
		Distinct("phone").
		Where("phone IN ?", uniquePhones)
	if excludeCustomerID != nil {
		query = query.Where("customer_id <> ?", *excludeCustomerID)
	}

	found := make(map[string]struct{})
	var foundPhones []string
	if err := query.Pluck("phone", &foundPhones).Error; err != nil {
		return nil, err
	}
	for _, phone := range foundPhones {
		found[phone] = struct{}{}
	}

	duplicates := make([]string, 0, len(found))
	for _, phone := range uniquePhones {
		if _, exists := found[phone]; exists {
			duplicates = append(duplicates, phone)
		}
	}
	return duplicates, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
