package repository

import (
	"backend/internal/model"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	salesDailyScoreRoleNames = []string{
		"sales_director",
		"sales_manager",
		"sales_staff",
		"sales_outside",
		"sale_outside",
	}
	salesDailyScoreRoleLabels = []string{
		"销售总监",
		"销售经理",
		"销售员工",
		"Outside销售",
		"outside销售",
	}
)

type SalesDailyScoreRepository interface {
	ListEnabledSalesUsers(ctx context.Context) ([]model.SalesDailyScoreUser, error)
	ListDailyCallMetrics(ctx context.Context, scoreDate string) ([]model.DailySalesCallMetric, error)
	CountVisitByUserOnDate(ctx context.Context, scoreDate string) (map[int64]int, error)
	CountNewCustomersByUserBetween(ctx context.Context, startUTC, endUTC time.Time) (map[int64]int, error)
	UpsertBatch(ctx context.Context, items []model.SalesDailyScoreUpsertInput) ([]model.SalesDailyScore, error)
	ListByDate(ctx context.Context, scoreDate string, actorUserID int64, actorRole string) ([]model.SalesDailyScore, error)
}

type gormSalesDailyScoreRepository struct {
	db *gorm.DB
}

type salesDailyScoreRow struct {
	ID                  int64     `gorm:"column:id;primaryKey;autoIncrement"`
	ScoreDate           string    `gorm:"column:score_date"`
	UserID              int64     `gorm:"column:user_id"`
	UserName            string    `gorm:"column:user_name"`
	RoleName            string    `gorm:"column:role_name"`
	CallNum             int       `gorm:"column:call_num"`
	CallDurationSecond  int       `gorm:"column:call_duration_second"`
	CallScoreByCount    int       `gorm:"column:call_score_by_count"`
	CallScoreByDuration int       `gorm:"column:call_score_by_duration"`
	CallScoreType       string    `gorm:"column:call_score_type"`
	CallScore           int       `gorm:"column:call_score"`
	VisitCount          int       `gorm:"column:visit_count"`
	VisitScore          int       `gorm:"column:visit_score"`
	NewCustomerCount    int       `gorm:"column:new_customer_count"`
	NewCustomerScore    int       `gorm:"column:new_customer_score"`
	TotalScore          int       `gorm:"column:total_score"`
	CreatedAt           time.Time `gorm:"column:created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at"`
}

type countByUserRow struct {
	UserID int64 `gorm:"column:user_id"`
	Count  int   `gorm:"column:item_count"`
}

func NewGormSalesDailyScoreRepository(db *gorm.DB) SalesDailyScoreRepository {
	return &gormSalesDailyScoreRepository{db: db}
}

func (r *gormSalesDailyScoreRepository) ListEnabledSalesUsers(ctx context.Context) ([]model.SalesDailyScoreUser, error) {
	var users []model.SalesDailyScoreUser
	err := r.db.WithContext(ctx).
		Table("users AS u").
		Select(
			"u.id AS user_id",
			"COALESCE(u.nickname, '') AS user_name",
			"CASE WHEN COALESCE(r.label, '') <> '' THEN r.label ELSE COALESCE(r.name, '') END AS role_name",
		).
		Joins("JOIN roles AS r ON r.id = u.role_id").
		Where("u.status = ?", model.UserStatusEnabled).
		Where("(r.name IN ? OR r.label IN ?)", salesDailyScoreRoleNames, salesDailyScoreRoleLabels).
		Order("u.id ASC").
		Scan(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *gormSalesDailyScoreRepository) ListDailyCallMetrics(
	ctx context.Context,
	scoreDate string,
) ([]model.DailySalesCallMetric, error) {
	var rows []model.DailySalesCallMetric
	err := r.db.WithContext(ctx).
		Table("daily_user_call_stats").
		Select(
			"user_id",
			"COALESCE(SUM(call_num), 0) AS call_num",
			"COALESCE(SUM(total_second), 0) AS call_duration_second",
		).
		Where("stat_date = ? AND user_id IS NOT NULL", strings.TrimSpace(scoreDate)).
		Group("user_id").
		Order("user_id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *gormSalesDailyScoreRepository) CountVisitByUserOnDate(
	ctx context.Context,
	scoreDate string,
) (map[int64]int, error) {
	var rows []countByUserRow
	err := r.db.WithContext(ctx).
		Table("customer_visits").
		Select("operator_user_id AS user_id", "COUNT(1) AS item_count").
		Where("visit_date = ?", strings.TrimSpace(scoreDate)).
		Group("operator_user_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64]int, len(rows))
	for _, row := range rows {
		result[row.UserID] = row.Count
	}
	return result, nil
}

func (r *gormSalesDailyScoreRepository) CountNewCustomersByUserBetween(
	ctx context.Context,
	startUTC, endUTC time.Time,
) (map[int64]int, error) {
	var rows []countByUserRow
	err := r.db.WithContext(ctx).
		Table("customers").
		Select("create_user_id AS user_id", "COUNT(1) AS item_count").
		Where("create_user_id > 0").
		Where("created_at >= ? AND created_at < ?", startUTC, endUTC).
		Group("create_user_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64]int, len(rows))
	for _, row := range rows {
		result[row.UserID] = row.Count
	}
	return result, nil
}

func (r *gormSalesDailyScoreRepository) UpsertBatch(
	ctx context.Context,
	items []model.SalesDailyScoreUpsertInput,
) ([]model.SalesDailyScore, error) {
	items = dedupeSalesDailyScoreUpsertInputs(items)
	if len(items) == 0 {
		return []model.SalesDailyScore{}, nil
	}

	now := time.Now().UTC()
	result := make([]model.SalesDailyScore, 0, len(items))
	for _, item := range items {
		row := salesDailyScoreRow{
			ScoreDate:           item.ScoreDate,
			UserID:              item.UserID,
			UserName:            item.UserName,
			RoleName:            item.RoleName,
			CallNum:             item.CallNum,
			CallDurationSecond:  item.CallDurationSecond,
			CallScoreByCount:    item.CallScoreByCount,
			CallScoreByDuration: item.CallScoreByDuration,
			CallScoreType:       item.CallScoreType,
			CallScore:           item.CallScore,
			VisitCount:          item.VisitCount,
			VisitScore:          item.VisitScore,
			NewCustomerCount:    item.NewCustomerCount,
			NewCustomerScore:    item.NewCustomerScore,
			TotalScore:          item.TotalScore,
			CreatedAt:           now,
			UpdatedAt:           now,
		}

		err := r.db.WithContext(ctx).Table("sales_daily_scores").Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "score_date"},
				{Name: "user_id"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"user_name":              row.UserName,
				"role_name":              row.RoleName,
				"call_num":               row.CallNum,
				"call_duration_second":   row.CallDurationSecond,
				"call_score_by_count":    row.CallScoreByCount,
				"call_score_by_duration": row.CallScoreByDuration,
				"call_score_type":        row.CallScoreType,
				"call_score":             row.CallScore,
				"visit_count":            row.VisitCount,
				"visit_score":            row.VisitScore,
				"new_customer_count":     row.NewCustomerCount,
				"new_customer_score":     row.NewCustomerScore,
				"total_score":            row.TotalScore,
				"updated_at":             row.UpdatedAt,
			}),
		}).Create(&row).Error
		if err != nil {
			return nil, err
		}

		var saved salesDailyScoreRow
		if err := r.db.WithContext(ctx).
			Table("sales_daily_scores").
			Where("score_date = ? AND user_id = ?", row.ScoreDate, row.UserID).
			Take(&saved).Error; err != nil {
			return nil, err
		}
		result = append(result, mapSalesDailyScoreRowToModel(saved))
	}

	return result, nil
}

func (r *gormSalesDailyScoreRepository) ListByDate(
	ctx context.Context,
	scoreDate string,
	actorUserID int64,
	actorRole string,
) ([]model.SalesDailyScore, error) {
	showAll := isSalesDailyScoreGlobalRole(actorRole)
	scopedUserIDs, err := r.resolveSalesDailyScoreScopeUserIDs(ctx, actorUserID, actorRole, showAll)
	if err != nil {
		return nil, err
	}

	var rows []salesDailyScoreRow
	query := r.db.WithContext(ctx).
		Table("sales_daily_scores AS s").
		Select("s.*").
		Joins("JOIN users AS u ON u.id = s.user_id").
		Joins("JOIN roles AS r ON r.id = u.role_id").
		Where("s.score_date = ?", strings.TrimSpace(scoreDate)).
		Where("u.status = ?", model.UserStatusEnabled).
		Where("(r.name IN ? OR r.label IN ?)", salesDailyScoreRoleNames, salesDailyScoreRoleLabels)
	if !showAll {
		if len(scopedUserIDs) == 0 {
			return []model.SalesDailyScore{}, nil
		}
		query = query.Where("u.id IN ?", scopedUserIDs)
	}
	err = query.Order("s.total_score DESC, s.user_id ASC").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []model.SalesDailyScore{}, nil
	}

	items := make([]model.SalesDailyScore, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapSalesDailyScoreRowToModel(row))
	}
	return items, nil
}

func isSalesDailyScoreGlobalRole(role string) bool {
	switch strings.TrimSpace(strings.ToLower(role)) {
	case "admin", "管理员", "finance", "finance_manager", "财务", "财务经理":
		return true
	default:
		return false
	}
}

func isSalesDailyScoreTeamRole(role string) bool {
	switch strings.TrimSpace(strings.ToLower(role)) {
	case "sales_director", "销售总监",
		"sales_manager", "销售经理",
		"sales_staff", "销售员工",
		"sales_outside", "sale_outside", "outside销售":
		return true
	default:
		return false
	}
}

func (r *gormSalesDailyScoreRepository) resolveSalesDailyScoreScopeUserIDs(ctx context.Context, actorUserID int64, actorRole string, showAll bool) ([]int64, error) {
	if showAll {
		return []int64{}, nil
	}
	if actorUserID <= 0 {
		return []int64{}, nil
	}

	roleName, err := r.getUserRoleName(ctx, actorUserID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(roleName) == "" {
		roleName = strings.TrimSpace(actorRole)
	}
	if !isSalesDailyScoreTeamRole(roleName) {
		return []int64{}, nil
	}

	anchorUserID, err := r.resolveSalesDailyScoreAnchorUserID(ctx, actorUserID)
	if err != nil {
		return nil, err
	}
	if anchorUserID <= 0 {
		return []int64{actorUserID}, nil
	}

	descendantIDs, err := r.listAllDescendantUserIDs(ctx, anchorUserID)
	if err != nil {
		return nil, err
	}

	teamUserIDs := uniquePositiveInt64(append([]int64{anchorUserID}, descendantIDs...))
	if len(teamUserIDs) == 0 {
		return []int64{}, nil
	}

	var scopedUserIDs []int64
	err = r.db.WithContext(ctx).
		Table("users AS u").
		Select("u.id").
		Joins("JOIN roles AS r ON r.id = u.role_id").
		Where("u.id IN ?", teamUserIDs).
		Where("u.status = ?", model.UserStatusEnabled).
		Where("(r.name IN ? OR r.label IN ?)", salesDailyScoreRoleNames, salesDailyScoreRoleLabels).
		Order("u.id ASC").
		Pluck("u.id", &scopedUserIDs).Error
	if err != nil {
		return nil, err
	}
	return uniquePositiveInt64(scopedUserIDs), nil
}

func (r *gormSalesDailyScoreRepository) resolveSalesDailyScoreAnchorUserID(ctx context.Context, userID int64) (int64, error) {
	if userID <= 0 {
		return 0, nil
	}

	visited := map[int64]struct{}{}
	currentUserID := userID
	for currentUserID > 0 {
		if _, seen := visited[currentUserID]; seen {
			return 0, nil
		}
		visited[currentUserID] = struct{}{}

		roleName, err := r.getUserRoleName(ctx, currentUserID)
		if err != nil {
			return 0, err
		}
		if strings.EqualFold(strings.TrimSpace(roleName), "sales_director") || strings.TrimSpace(roleName) == "销售总监" {
			return currentUserID, nil
		}

		parentUserID, err := r.getParentUserID(ctx, currentUserID)
		if err != nil {
			return 0, err
		}
		currentUserID = parentUserID
	}

	return 0, nil
}

func (r *gormSalesDailyScoreRepository) getUserRoleName(ctx context.Context, userID int64) (string, error) {
	if userID <= 0 {
		return "", nil
	}

	var roleName string
	err := r.db.WithContext(ctx).
		Table("users AS u").
		Select("COALESCE(r.name, '')").
		Joins("LEFT JOIN roles r ON r.id = u.role_id").
		Where("u.id = ?", userID).
		Limit(1).
		Scan(&roleName).Error
	if err != nil {
		return "", err
	}
	return roleName, nil
}

func (r *gormSalesDailyScoreRepository) getParentUserID(ctx context.Context, userID int64) (int64, error) {
	if userID <= 0 {
		return 0, nil
	}

	type parentRow struct {
		ParentID *int64 `gorm:"column:parent_id"`
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
	if row.ParentID == nil || *row.ParentID <= 0 {
		return 0, nil
	}
	return *row.ParentID, nil
}

func (r *gormSalesDailyScoreRepository) listAllDescendantUserIDs(ctx context.Context, rootUserID int64) ([]int64, error) {
	if rootUserID <= 0 {
		return []int64{}, nil
	}

	queue := []int64{rootUserID}
	seen := map[int64]struct{}{rootUserID: {}}
	result := make([]int64, 0)

	for len(queue) > 0 {
		var nextLevel []int64
		if err := r.db.WithContext(ctx).
			Table("users").
			Where("parent_id IN ?", queue).
			Order("id ASC").
			Pluck("id", &nextLevel).Error; err != nil {
			return nil, err
		}

		queue = queue[:0]
		for _, id := range nextLevel {
			if id <= 0 {
				continue
			}
			if _, exists := seen[id]; exists {
				continue
			}
			seen[id] = struct{}{}
			result = append(result, id)
			queue = append(queue, id)
		}
	}

	return result, nil
}

func dedupeSalesDailyScoreUpsertInputs(
	items []model.SalesDailyScoreUpsertInput,
) []model.SalesDailyScoreUpsertInput {
	if len(items) == 0 {
		return []model.SalesDailyScoreUpsertInput{}
	}

	order := make([]string, 0, len(items))
	merged := make(map[string]model.SalesDailyScoreUpsertInput, len(items))
	for _, item := range items {
		key := fmt.Sprintf("%s\x00%d", strings.TrimSpace(item.ScoreDate), item.UserID)
		if _, exists := merged[key]; !exists {
			order = append(order, key)
		}
		merged[key] = model.SalesDailyScoreUpsertInput{
			ScoreDate:           strings.TrimSpace(item.ScoreDate),
			UserID:              item.UserID,
			UserName:            strings.TrimSpace(item.UserName),
			RoleName:            strings.TrimSpace(item.RoleName),
			CallNum:             item.CallNum,
			CallDurationSecond:  item.CallDurationSecond,
			CallScoreByCount:    item.CallScoreByCount,
			CallScoreByDuration: item.CallScoreByDuration,
			CallScoreType:       strings.TrimSpace(item.CallScoreType),
			CallScore:           item.CallScore,
			VisitCount:          item.VisitCount,
			VisitScore:          item.VisitScore,
			NewCustomerCount:    item.NewCustomerCount,
			NewCustomerScore:    item.NewCustomerScore,
			TotalScore:          item.TotalScore,
		}
	}

	result := make([]model.SalesDailyScoreUpsertInput, 0, len(order))
	for _, key := range order {
		result = append(result, merged[key])
	}
	return result
}

func mapSalesDailyScoreRowToModel(row salesDailyScoreRow) model.SalesDailyScore {
	return model.SalesDailyScore{
		ID:                  row.ID,
		ScoreDate:           row.ScoreDate,
		UserID:              row.UserID,
		UserName:            row.UserName,
		RoleName:            row.RoleName,
		CallNum:             row.CallNum,
		CallDurationSecond:  row.CallDurationSecond,
		CallScoreByCount:    row.CallScoreByCount,
		CallScoreByDuration: row.CallScoreByDuration,
		CallScoreType:       row.CallScoreType,
		CallScore:           row.CallScore,
		VisitCount:          row.VisitCount,
		VisitScore:          row.VisitScore,
		NewCustomerCount:    row.NewCustomerCount,
		NewCustomerScore:    row.NewCustomerScore,
		TotalScore:          row.TotalScore,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
	}
}
