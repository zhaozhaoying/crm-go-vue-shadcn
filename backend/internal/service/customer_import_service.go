package service

import (
	"backend/internal/errmsg"
	"bufio"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"backend/internal/model"
	"backend/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrCustomerImportInvalidFile   = errors.New("invalid customer import file")
	ErrCustomerImportInvalidHeader = errors.New("invalid customer import header")
)

var nonDigitRegex = regexp.MustCompile(`\D+`)

type CustomerImportService interface {
	ImportCSV(ctx context.Context, reader io.Reader, input CustomerCSVImportInput) (model.CustomerImportResult, error)
}

type CustomerCSVImportInput struct {
	OperatorUserID int64
	BatchSize      int
	DryRun         bool
	DefaultStatus  string
	MaxErrors      int
}

type customerImportService struct {
	db              *gorm.DB
	activityLogRepo *repository.ActivityLogRepository
}

type customerImportRow struct {
	RowNum           int
	Name             string
	LegalName        string
	ContactName      string
	Phone            string
	PhoneLabel       string
	Weixin           string
	Email            string
	Province         int
	City             int
	Area             int
	DetailAddress    string
	Remark           string
	CustomerLevelID  int
	CustomerSourceID int
	Status           string
	OwnerUserID      *int64
}

type customerCreateImportRow struct {
	ID               int64      `gorm:"column:id;primaryKey;autoIncrement"`
	Name             string     `gorm:"column:name"`
	LegalName        string     `gorm:"column:legal_name"`
	ContactName      string     `gorm:"column:contact_name"`
	Weixin           string     `gorm:"column:weixin"`
	Email            string     `gorm:"column:email"`
	CustomerLevelID  int        `gorm:"column:customer_level_id"`
	CustomerSourceID int        `gorm:"column:customer_source_id"`
	Province         int        `gorm:"column:province"`
	City             int        `gorm:"column:city"`
	Area             int        `gorm:"column:area"`
	DetailAddress    string     `gorm:"column:detail_address"`
	Remark           string     `gorm:"column:remark"`
	Status           string     `gorm:"column:status"`
	DealStatus       string     `gorm:"column:deal_status"`
	OwnerUserID      *int64     `gorm:"column:owner_user_id"`
	CreateUserID     int64      `gorm:"column:create_user_id"`
	OperateUserID    int64      `gorm:"column:operate_user_id"`
	CollectTime      *int64     `gorm:"column:collect_time"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
	DeleteTime       *time.Time `gorm:"column:delete_time"`
}

type customerPhoneImportRow struct {
	CustomerID int64  `gorm:"column:customer_id"`
	Phone      string `gorm:"column:phone"`
	PhoneLabel string `gorm:"column:phone_label"`
	IsPrimary  int    `gorm:"column:is_primary"`
	CreatedAt  int64  `gorm:"column:created_at"`
	UpdatedAt  int64  `gorm:"column:updated_at"`
}

type customerOwnerLogImportRow struct {
	CustomerID      int64     `gorm:"column:customer_id"`
	FromOwnerUserID *int64    `gorm:"column:from_owner_user_id"`
	ToOwnerUserID   *int64    `gorm:"column:to_owner_user_id"`
	Action          string    `gorm:"column:action"`
	OperatorUserID  int64     `gorm:"column:operator_user_id"`
	CreatedAt       time.Time `gorm:"column:created_at"`
}

func NewCustomerImportService(db *gorm.DB, activityLogRepo ...*repository.ActivityLogRepository) CustomerImportService {
	svc := &customerImportService{db: db}
	if len(activityLogRepo) > 0 {
		svc.activityLogRepo = activityLogRepo[0]
	}
	return svc
}

func (s *customerImportService) ImportCSV(ctx context.Context, reader io.Reader, input CustomerCSVImportInput) (model.CustomerImportResult, error) {
	if reader == nil {
		return model.CustomerImportResult{}, ErrCustomerImportInvalidFile
	}
	if input.OperatorUserID <= 0 {
		return model.CustomerImportResult{}, fmt.Errorf("operator user id is required")
	}
	if input.BatchSize <= 0 {
		input.BatchSize = 1000
	}
	if input.BatchSize > 5000 {
		input.BatchSize = 5000
	}
	if input.MaxErrors <= 0 {
		input.MaxErrors = 200
	}

	result := model.CustomerImportResult{
		DryRun: input.DryRun,
		Errors: make([]model.CustomerImportError, 0),
	}

	csvReader := csv.NewReader(bufio.NewReaderSize(reader, 64*1024))
	csvReader.FieldsPerRecord = -1
	csvReader.TrimLeadingSpace = true
	csvReader.LazyQuotes = true
	csvReader.ReuseRecord = false

	header, err := csvReader.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return result, ErrCustomerImportInvalidFile
		}
		return result, err
	}

	indexes, err := resolveCustomerImportHeader(header)
	if err != nil {
		return result, err
	}

	seenNames := make(map[string]struct{})
	seenLegalNames := make(map[string]struct{})
	seenWeixins := make(map[string]struct{})
	seenPhones := make(map[string]struct{})

	rows := make([]customerImportRow, 0, input.BatchSize)
	rowNum := 1
	for {
		record, readErr := csvReader.Read()
		rowNum++
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			result.TotalRows++
			result.FailedRows++
			appendCustomerImportError(&result, input.MaxErrors, model.CustomerImportError{
				Row:    rowNum,
				Reason: "csv parse error: " + readErr.Error(),
			})
			continue
		}

		row, empty, parseErr := parseCustomerImportRecord(record, rowNum, indexes, input.DefaultStatus)
		if empty {
			continue
		}

		result.TotalRows++
		if parseErr != nil {
			result.FailedRows++
			appendCustomerImportError(&result, input.MaxErrors, model.CustomerImportError{
				Row:    rowNum,
				Reason: parseErr.Error(),
			})
			continue
		}

		if duplicateReason := duplicateInFile(row, seenNames, seenLegalNames, seenWeixins, seenPhones); duplicateReason != "" {
			result.SkippedRows++
			appendCustomerImportError(&result, input.MaxErrors, model.CustomerImportError{
				Row:    rowNum,
				Name:   row.Name,
				Phone:  row.Phone,
				Reason: duplicateReason,
			})
			continue
		}

		rememberCustomerImportRow(row, seenNames, seenLegalNames, seenWeixins, seenPhones)
		rows = append(rows, row)
		if len(rows) >= input.BatchSize {
			if err := s.applyImportBatch(ctx, rows, input, &result); err != nil {
				return result, err
			}
			rows = rows[:0]
		}
	}

	if len(rows) > 0 {
		if err := s.applyImportBatch(ctx, rows, input, &result); err != nil {
			return result, err
		}
	}

	if result.SuccessRows > 0 && s.activityLogRepo != nil {
		content := fmt.Sprintf("成功导入 %d 条客户", result.SuccessRows)
		_ = s.activityLogRepo.Create(ctx, model.ActivityLog{
			UserID:     input.OperatorUserID,
			Action:     model.ActionImportCustomer,
			TargetType: model.TargetTypeCustomer,
			Content:    content,
		})
	}

	return result, nil
}

func (s *customerImportService) applyImportBatch(
	ctx context.Context,
	rows []customerImportRow,
	input CustomerCSVImportInput,
	result *model.CustomerImportResult,
) error {
	if len(rows) == 0 {
		return nil
	}

	operatorRole, err := s.getUserRoleName(ctx, input.OperatorUserID)
	if err != nil {
		return err
	}

	nameSet, legalNameSet, weixinSet, phoneSet := collectImportBatchKeys(rows)
	existingNames, err := s.findExistingCustomerFieldValues(ctx, "name", nameSet)
	if err != nil {
		return err
	}
	existingLegalNames, err := s.findExistingCustomerFieldValues(ctx, "legal_name", legalNameSet)
	if err != nil {
		return err
	}
	existingWeixins, err := s.findExistingCustomerFieldValues(ctx, "weixin", weixinSet)
	if err != nil {
		return err
	}
	existingPhones, err := s.findExistingPhones(ctx, phoneSet)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	nowUnix := now.Unix()

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer tx.Rollback()
	assignmentRepo := repository.NewGormCustomerRepository(tx)

	for idx, row := range rows {
		if row.Name != "" {
			if _, ok := existingNames[row.Name]; ok {
				result.SkippedRows++
				appendCustomerImportError(result, input.MaxErrors, model.CustomerImportError{
					Row:    row.RowNum,
					Name:   row.Name,
					Phone:  row.Phone,
					Reason: "duplicate name in database",
				})
				continue
			}
		}
		if row.LegalName != "" {
			if _, ok := existingLegalNames[row.LegalName]; ok {
				result.SkippedRows++
				appendCustomerImportError(result, input.MaxErrors, model.CustomerImportError{
					Row:    row.RowNum,
					Name:   row.Name,
					Phone:  row.Phone,
					Reason: "duplicate legalName in database",
				})
				continue
			}
		}
		if row.Weixin != "" {
			if _, ok := existingWeixins[row.Weixin]; ok {
				result.SkippedRows++
				appendCustomerImportError(result, input.MaxErrors, model.CustomerImportError{
					Row:    row.RowNum,
					Name:   row.Name,
					Phone:  row.Phone,
					Reason: "duplicate weixin in database",
				})
				continue
			}
		}
		if _, ok := existingPhones[row.Phone]; ok {
			result.SkippedRows++
			appendCustomerImportError(result, input.MaxErrors, model.CustomerImportError{
				Row:    row.RowNum,
				Name:   row.Name,
				Phone:  row.Phone,
				Reason: "duplicate phone in database",
			})
			continue
		}

		savePoint := fmt.Sprintf("customer_import_%d", idx)
		if err := tx.SavePoint(savePoint).Error; err != nil {
			return err
		}

		ownerUserID, collectTime, err := s.resolveImportOwner(ctx, assignmentRepo, row, input.OperatorUserID, operatorRole, nowUnix)
		if err != nil {
			return err
		}
		customerRow := customerCreateImportRow{
			Name:             row.Name,
			LegalName:        row.LegalName,
			ContactName:      row.ContactName,
			Weixin:           row.Weixin,
			Email:            row.Email,
			CustomerLevelID:  row.CustomerLevelID,
			CustomerSourceID: row.CustomerSourceID,
			Province:         row.Province,
			City:             row.City,
			Area:             row.Area,
			DetailAddress:    row.DetailAddress,
			Remark:           row.Remark,
			Status:           row.Status,
			DealStatus:       model.CustomerDealStatusUndone,
			OwnerUserID:      ownerUserID,
			CreateUserID:     input.OperatorUserID,
			OperateUserID:    input.OperatorUserID,
			CollectTime:      collectTime,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		if err := tx.Table("customers").Create(&customerRow).Error; err != nil {
			_ = tx.RollbackTo(savePoint).Error
			if isCustomerImportDuplicateError(err) {
				result.SkippedRows++
				appendCustomerImportError(result, input.MaxErrors, model.CustomerImportError{
					Row:    row.RowNum,
					Name:   row.Name,
					Phone:  row.Phone,
					Reason: "duplicate record in database",
				})
				continue
			}
			result.FailedRows++
			appendCustomerImportError(result, input.MaxErrors, model.CustomerImportError{
				Row:    row.RowNum,
				Name:   row.Name,
				Phone:  row.Phone,
				Reason: "insert customer failed: " + err.Error(),
			})
			continue
		}

		phoneRow := customerPhoneImportRow{
			CustomerID: customerRow.ID,
			Phone:      row.Phone,
			PhoneLabel: row.PhoneLabel,
			IsPrimary:  1,
			CreatedAt:  nowUnix,
			UpdatedAt:  nowUnix,
		}
		if err := tx.Table("customer_phones").Create(&phoneRow).Error; err != nil {
			_ = tx.RollbackTo(savePoint).Error
			if isCustomerImportDuplicateError(err) {
				result.SkippedRows++
				appendCustomerImportError(result, input.MaxErrors, model.CustomerImportError{
					Row:    row.RowNum,
					Name:   row.Name,
					Phone:  row.Phone,
					Reason: "duplicate phone in database",
				})
				continue
			}
			result.FailedRows++
			appendCustomerImportError(result, input.MaxErrors, model.CustomerImportError{
				Row:    row.RowNum,
				Name:   row.Name,
				Phone:  row.Phone,
				Reason: "insert customer phone failed: " + err.Error(),
			})
			continue
		}

		if ownerUserID != nil {
			logRow := customerOwnerLogImportRow{
				CustomerID:      customerRow.ID,
				FromOwnerUserID: nil,
				ToOwnerUserID:   ownerUserID,
				Action:          "claim",
				OperatorUserID:  input.OperatorUserID,
				CreatedAt:       now,
			}
			if err := tx.Table("customer_owner_logs").Create(&logRow).Error; err != nil {
				_ = tx.RollbackTo(savePoint).Error
				result.FailedRows++
				appendCustomerImportError(result, input.MaxErrors, model.CustomerImportError{
					Row:    row.RowNum,
					Name:   row.Name,
					Phone:  row.Phone,
					Reason: "insert owner log failed: " + err.Error(),
				})
				continue
			}
		}

		result.SuccessRows++
		existingNames[row.Name] = struct{}{}
		if row.LegalName != "" {
			existingLegalNames[row.LegalName] = struct{}{}
		}
		if row.Weixin != "" {
			existingWeixins[row.Weixin] = struct{}{}
		}
		existingPhones[row.Phone] = struct{}{}
	}

	if input.DryRun {
		return tx.Rollback().Error
	}
	return tx.Commit().Error
}

func (s *customerImportService) resolveImportOwner(ctx context.Context, assignmentRepo customerOwnerAssignmentRepo, row customerImportRow, operatorUserID int64, operatorRole string, nowUnix int64) (*int64, *int64, error) {
	if row.Status == model.CustomerStatusPool {
		return nil, nil, nil
	}
	if isInsideSalesRole(operatorRole) {
		owner, err := pickBalancedSalesOwnerUserID(ctx, assignmentRepo, operatorUserID)
		if err != nil {
			return nil, nil, err
		}
		return &owner, &nowUnix, nil
	}
	if isOutsideSalesRole(operatorRole) {
		owner := operatorUserID
		return &owner, &nowUnix, nil
	}
	owner := row.OwnerUserID
	if owner == nil {
		defaultOwner := operatorUserID
		owner = &defaultOwner
	}
	return owner, &nowUnix, nil
}

func (s *customerImportService) getUserRoleName(ctx context.Context, userID int64) (string, error) {
	if userID <= 0 {
		return "", nil
	}

	var roleName string
	err := s.db.WithContext(ctx).
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

func collectImportBatchKeys(rows []customerImportRow) (names []string, legalNames []string, weixins []string, phones []string) {
	nameSet := make(map[string]struct{})
	legalNameSet := make(map[string]struct{})
	weixinSet := make(map[string]struct{})
	phoneSet := make(map[string]struct{})

	for _, row := range rows {
		if row.Name != "" {
			nameSet[row.Name] = struct{}{}
		}
		if row.LegalName != "" {
			legalNameSet[row.LegalName] = struct{}{}
		}
		if row.Weixin != "" {
			weixinSet[row.Weixin] = struct{}{}
		}
		if row.Phone != "" {
			phoneSet[row.Phone] = struct{}{}
		}
	}

	names = mapKeys(nameSet)
	legalNames = mapKeys(legalNameSet)
	weixins = mapKeys(weixinSet)
	phones = mapKeys(phoneSet)
	return
}

func mapKeys(set map[string]struct{}) []string {
	if len(set) == 0 {
		return nil
	}
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	return keys
}

func (s *customerImportService) findExistingCustomerFieldValues(ctx context.Context, field string, values []string) (map[string]struct{}, error) {
	result := make(map[string]struct{})
	if len(values) == 0 {
		return result, nil
	}

	switch field {
	case "name", "legal_name", "weixin":
	default:
		return nil, fmt.Errorf("unsupported field: %s", field)
	}

	selectSQL := fmt.Sprintf("DISTINCT TRIM(%s) AS value", field)
	whereSQL := fmt.Sprintf("TRIM(%s) IN ?", field)
	var rows []struct {
		Value string `gorm:"column:value"`
	}
	if err := s.db.WithContext(ctx).
		Table("customers").
		Select(selectSQL).
		Where(whereSQL, values).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		trimmed := strings.TrimSpace(row.Value)
		if trimmed != "" {
			result[trimmed] = struct{}{}
		}
	}
	return result, nil
}

func (s *customerImportService) findExistingPhones(ctx context.Context, phones []string) (map[string]struct{}, error) {
	result := make(map[string]struct{})
	if len(phones) == 0 {
		return result, nil
	}

	var rows []struct {
		Phone string `gorm:"column:phone"`
	}
	if err := s.db.WithContext(ctx).
		Table("customer_phones").
		Select("DISTINCT phone").
		Where("phone IN ?", phones).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		trimmed := strings.TrimSpace(row.Phone)
		if trimmed != "" {
			result[trimmed] = struct{}{}
		}
	}
	return result, nil
}

func duplicateInFile(
	row customerImportRow,
	seenNames map[string]struct{},
	seenLegalNames map[string]struct{},
	seenWeixins map[string]struct{},
	seenPhones map[string]struct{},
) string {
	if row.Name != "" {
		if _, exists := seenNames[row.Name]; exists {
			return "duplicate name in import file"
		}
	}
	if row.LegalName != "" {
		if _, exists := seenLegalNames[row.LegalName]; exists {
			return "duplicate legalName in import file"
		}
	}
	if row.Weixin != "" {
		if _, exists := seenWeixins[row.Weixin]; exists {
			return "duplicate weixin in import file"
		}
	}
	if row.Phone != "" {
		if _, exists := seenPhones[row.Phone]; exists {
			return "duplicate phone in import file"
		}
	}
	return ""
}

func rememberCustomerImportRow(
	row customerImportRow,
	seenNames map[string]struct{},
	seenLegalNames map[string]struct{},
	seenWeixins map[string]struct{},
	seenPhones map[string]struct{},
) {
	if row.Name != "" {
		seenNames[row.Name] = struct{}{}
	}
	if row.LegalName != "" {
		seenLegalNames[row.LegalName] = struct{}{}
	}
	if row.Weixin != "" {
		seenWeixins[row.Weixin] = struct{}{}
	}
	if row.Phone != "" {
		seenPhones[row.Phone] = struct{}{}
	}
}

func appendCustomerImportError(result *model.CustomerImportResult, maxErrors int, errItem model.CustomerImportError) {
	if maxErrors <= 0 {
		return
	}
	if len(result.Errors) >= maxErrors {
		return
	}
	errItem.Reason = errmsg.Normalize(errItem.Reason)
	result.Errors = append(result.Errors, errItem)
}

func isCustomerImportDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate entry") ||
		strings.Contains(msg, "unique constraint failed")
}

func parseCustomerImportRecord(
	record []string,
	rowNum int,
	indexes map[string]int,
	defaultStatus string,
) (customerImportRow, bool, error) {
	get := func(field string) string {
		idx, ok := indexes[field]
		if !ok || idx < 0 || idx >= len(record) {
			return ""
		}
		return strings.TrimSpace(record[idx])
	}

	name := strings.TrimSpace(get("name"))
	phoneRaw := strings.TrimSpace(get("phone"))

	if name == "" && phoneRaw == "" {
		return customerImportRow{}, true, nil
	}
	if name == "" {
		return customerImportRow{}, false, fmt.Errorf("name is required")
	}

	phone := normalizeImportPhone(phoneRaw)
	if phone == "" {
		return customerImportRow{}, false, fmt.Errorf("phone is invalid")
	}

	province, err := parseImportInt(get("province"))
	if err != nil {
		return customerImportRow{}, false, fmt.Errorf("province is invalid")
	}
	city, err := parseImportInt(get("city"))
	if err != nil {
		return customerImportRow{}, false, fmt.Errorf("city is invalid")
	}
	area, err := parseImportInt(get("area"))
	if err != nil {
		return customerImportRow{}, false, fmt.Errorf("area is invalid")
	}
	customerLevelID, err := parseImportInt(get("customer_level_id"))
	if err != nil {
		return customerImportRow{}, false, fmt.Errorf("customerLevelId is invalid")
	}
	customerSourceID, err := parseImportInt(get("customer_source_id"))
	if err != nil {
		return customerImportRow{}, false, fmt.Errorf("customerSourceId is invalid")
	}
	ownerUserID, err := parseImportInt64Pointer(get("owner_user_id"))
	if err != nil {
		return customerImportRow{}, false, fmt.Errorf("ownerUserId is invalid")
	}

	status := normalizeImportStatus(get("status"), defaultStatus)
	phoneLabel := strings.TrimSpace(get("phone_label"))
	if phoneLabel == "" {
		phoneLabel = "导入"
	}

	return customerImportRow{
		RowNum:           rowNum,
		Name:             name,
		LegalName:        strings.TrimSpace(get("legal_name")),
		ContactName:      strings.TrimSpace(get("contact_name")),
		Phone:            phone,
		PhoneLabel:       phoneLabel,
		Weixin:           strings.TrimSpace(get("weixin")),
		Email:            strings.TrimSpace(get("email")),
		Province:         province,
		City:             city,
		Area:             area,
		DetailAddress:    strings.TrimSpace(get("detail_address")),
		Remark:           strings.TrimSpace(get("remark")),
		CustomerLevelID:  customerLevelID,
		CustomerSourceID: customerSourceID,
		Status:           status,
		OwnerUserID:      ownerUserID,
	}, false, nil
}

func parseImportInt(raw string) (int, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, nil
	}
	value, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, err
	}
	if value < 0 {
		return 0, fmt.Errorf("negative value")
	}
	return value, nil
}

func parseImportInt64Pointer(raw string) (*int64, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}
	value, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		return nil, err
	}
	if value <= 0 {
		return nil, nil
	}
	return &value, nil
}

func normalizeImportPhone(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	digits := nonDigitRegex.ReplaceAllString(trimmed, "")
	if strings.HasPrefix(digits, "86") && len(digits) == 13 {
		digits = digits[2:]
	}
	if len(digits) < 7 || len(digits) > 15 {
		return ""
	}
	return digits
}

func normalizeImportStatus(raw string, defaultStatus string) string {
	parse := func(value string) string {
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "owned", "my", "private", "私有", "我的", "已分配":
			return model.CustomerStatusOwned
		case "pool", "public", "公海":
			return model.CustomerStatusPool
		default:
			return ""
		}
	}

	if parsed := parse(raw); parsed != "" {
		return parsed
	}
	if parsed := parse(defaultStatus); parsed != "" {
		return parsed
	}
	return model.CustomerStatusOwned
}

func resolveCustomerImportHeader(header []string) (map[string]int, error) {
	indexes := map[string]int{
		"name":               -1,
		"legal_name":         -1,
		"contact_name":       -1,
		"phone":              -1,
		"phone_label":        -1,
		"weixin":             -1,
		"email":              -1,
		"province":           -1,
		"city":               -1,
		"area":               -1,
		"detail_address":     -1,
		"remark":             -1,
		"customer_level_id":  -1,
		"customer_source_id": -1,
		"status":             -1,
		"owner_user_id":      -1,
	}

	aliases := map[string][]string{
		"name":               {"name", "customername", "customer", "客户名称", "客户名", "公司名称"},
		"legal_name":         {"legalname", "主体名称", "客户主体", "企业名称", "营业执照名称"},
		"contact_name":       {"contactname", "contact", "联系人", "联系人姓名"},
		"phone":              {"phone", "mobile", "primaryphone", "手机号", "手机", "电话", "联系电话"},
		"phone_label":        {"phonelabel", "phone_tag", "电话标签"},
		"weixin":             {"weixin", "wechat", "微信", "微信号"},
		"email":              {"email", "邮箱", "电子邮箱"},
		"province":           {"province", "省", "省编码"},
		"city":               {"city", "市", "市编码"},
		"area":               {"area", "区", "区编码"},
		"detail_address":     {"detailaddress", "address", "详细地址", "地址"},
		"remark":             {"remark", "备注"},
		"customer_level_id":  {"customerlevelid", "客户等级id", "客户级别id"},
		"customer_source_id": {"customersourceid", "客户来源id"},
		"status":             {"status", "客户状态", "状态"},
		"owner_user_id":      {"owneruserid", "负责人id", "所属人id"},
	}

	aliasToField := make(map[string]string)
	for field, keys := range aliases {
		for _, key := range keys {
			aliasToField[normalizeCustomerImportHeader(key)] = field
		}
	}

	for idx, col := range header {
		key := normalizeCustomerImportHeader(col)
		field, ok := aliasToField[key]
		if !ok {
			continue
		}
		if indexes[field] < 0 {
			indexes[field] = idx
		}
	}

	if indexes["name"] < 0 || indexes["phone"] < 0 {
		return nil, fmt.Errorf("%w: required columns: name, phone", ErrCustomerImportInvalidHeader)
	}
	return indexes, nil
}

func normalizeCustomerImportHeader(raw string) string {
	value := strings.TrimSpace(strings.TrimPrefix(raw, "\uFEFF"))
	value = strings.ToLower(value)
	value = strings.ReplaceAll(value, "_", "")
	value = strings.ReplaceAll(value, "-", "")
	value = strings.ReplaceAll(value, " ", "")
	return value
}
