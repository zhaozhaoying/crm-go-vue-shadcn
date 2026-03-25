package repository

import (
	"backend/internal/model"
	"context"
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
	ListDailyCallEventsByUser(ctx context.Context, startUTC, endUTC time.Time) (map[int64][]model.DailySalesCallEvent, error)
	CountVisitByUserOnDate(ctx context.Context, scoreDate string) (map[int64]int, error)
	ListVisitEventTimesByUserOnDate(ctx context.Context, scoreDate string) (map[int64][]time.Time, error)
	CountNewCustomersByUserBetween(ctx context.Context, startUTC, endUTC time.Time) (map[int64]int, error)
	ListNewCustomerEventTimesByUserBetween(ctx context.Context, startUTC, endUTC time.Time) (map[int64][]time.Time, error)
	UpsertBatch(ctx context.Context, items []model.SalesDailyScoreUpsertInput) ([]model.SalesDailyScore, error)
	ListByDate(ctx context.Context, scoreDate string, actorUserID int64, actorRole string) ([]model.SalesDailyScore, error)
}

type gormSalesDailyScoreRepository struct {
	db *gorm.DB
}

type salesDailyScoreRow struct {
	ID                  int64      `gorm:"column:id;primaryKey;autoIncrement"`
	ScoreDate           string     `gorm:"column:score_date"`
	UserID              int64      `gorm:"column:user_id"`
	UserName            string     `gorm:"column:user_name"`
	RoleName            string     `gorm:"column:role_name"`
	CallNum             int        `gorm:"column:call_num"`
	CallDurationSecond  int        `gorm:"column:call_duration_second"`
	CallScoreByCount    int        `gorm:"column:call_score_by_count"`
	CallScoreByDuration int        `gorm:"column:call_score_by_duration"`
	CallScoreType       string     `gorm:"column:call_score_type"`
	CallScore           int        `gorm:"column:call_score"`
	VisitCount          int        `gorm:"column:visit_count"`
	VisitScore          int        `gorm:"column:visit_score"`
	NewCustomerCount    int        `gorm:"column:new_customer_count"`
	NewCustomerScore    int        `gorm:"column:new_customer_score"`
	TotalScore          int        `gorm:"column:total_score"`
	ScoreReachedAt      *time.Time `gorm:"column:score_reached_at"`
	CreatedAt           time.Time  `gorm:"column:created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at"`
}

type countByUserRow struct {
	UserID int64 `gorm:"column:user_id"`
	Count  int   `gorm:"column:item_count"`
}

type timeByUserRow struct {
	UserID    int64     `gorm:"column:user_id"`
	EventTime time.Time `gorm:"column:event_time"`
}

type callEventByUserRow struct {
	UserID         int64     `gorm:"column:user_id"`
	CallID         string    `gorm:"column:call_id"`
	StartTime      int64     `gorm:"column:start_time"`
	EndTime        int64     `gorm:"column:end_time"`
	CreateTime     int64     `gorm:"column:create_time"`
	DurationSecond int       `gorm:"column:duration_second"`
	CreatedAt      time.Time `gorm:"column:created_at"`
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

func (r *gormSalesDailyScoreRepository) ListDailyCallEventsByUser(
	ctx context.Context,
	startUTC, endUTC time.Time,
) (map[int64][]model.DailySalesCallEvent, error) {
	startMillis := startUTC.UnixMilli()
	endMillis := endUTC.UnixMilli()

	var rows []callEventByUserRow
	err := r.db.WithContext(ctx).
		Table("call_recordings AS c").
		Select(
			"DISTINCT um.user_id AS user_id",
			"c.id AS call_id",
			"c.start_time AS start_time",
			"c.end_time AS end_time",
			"c.create_time AS create_time",
			"c.duration AS duration_second",
			"c.created_at AS created_at",
		).
		Joins("JOIN user_hanghang_crm_mobiles AS um ON (um.mobile = c.mobile OR um.mobile = c.tel_a)").
		Where("c.start_time >= ? AND c.start_time < ?", startMillis, endMillis).
		Order("um.user_id ASC, c.start_time ASC, c.id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64][]model.DailySalesCallEvent, len(rows))
	for _, row := range rows {
		result[row.UserID] = append(result[row.UserID], model.DailySalesCallEvent{
			UserID:         row.UserID,
			EventTime:      resolveCallEventTime(row.StartTime, row.EndTime, row.CreateTime, row.DurationSecond, row.CreatedAt),
			DurationSecond: row.DurationSecond,
		})
	}
	return result, nil
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

func (r *gormSalesDailyScoreRepository) ListVisitEventTimesByUserOnDate(
	ctx context.Context,
	scoreDate string,
) (map[int64][]time.Time, error) {
	var rows []timeByUserRow
	err := r.db.WithContext(ctx).
		Table("customer_visits").
		Select("operator_user_id AS user_id", "created_at AS event_time").
		Where("visit_date = ?", strings.TrimSpace(scoreDate)).
		Order("operator_user_id ASC, created_at ASC, id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64][]time.Time, len(rows))
	for _, row := range rows {
		result[row.UserID] = append(result[row.UserID], row.EventTime)
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

func (r *gormSalesDailyScoreRepository) ListNewCustomerEventTimesByUserBetween(
	ctx context.Context,
	startUTC, endUTC time.Time,
) (map[int64][]time.Time, error) {
	var rows []timeByUserRow
	err := r.db.WithContext(ctx).
		Table("customers").
		Select("create_user_id AS user_id", "created_at AS event_time").
		Where("create_user_id > 0").
		Where("created_at >= ? AND created_at < ?", startUTC, endUTC).
		Order("create_user_id ASC, created_at ASC, id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64][]time.Time, len(rows))
	for _, row := range rows {
		result[row.UserID] = append(result[row.UserID], row.EventTime)
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
			ScoreReachedAt:      item.ScoreReachedAt,
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
				"score_reached_at":       row.ScoreReachedAt,
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
	_ int64,
	_ string,
) ([]model.SalesDailyScore, error) {
	var rows []salesDailyScoreRow
	query := r.db.WithContext(ctx).
		Table("sales_daily_scores AS s").
		Select("s.*").
		Joins("JOIN users AS u ON u.id = s.user_id").
		Joins("JOIN roles AS r ON r.id = u.role_id").
		Where("s.score_date = ?", strings.TrimSpace(scoreDate)).
		Where("u.status = ?", model.UserStatusEnabled).
		Where("(r.name IN ? OR r.label IN ?)", salesDailyScoreRoleNames, salesDailyScoreRoleLabels)
	err := query.Order(salesDailyScoreOrderClause("s")).Find(&rows).Error
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
			ScoreReachedAt:      item.ScoreReachedAt,
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
		ScoreReachedAt:      row.ScoreReachedAt,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
	}
}

func salesDailyScoreOrderClause(alias string) string {
	prefix := ""
	if trimmed := strings.TrimSpace(alias); trimmed != "" {
		prefix = trimmed + "."
	}
	return fmt.Sprintf(
		"%stotal_score DESC, CASE WHEN %sscore_reached_at IS NULL THEN 1 ELSE 0 END ASC, %sscore_reached_at ASC, %suser_id ASC",
		prefix,
		prefix,
		prefix,
		prefix,
	)
}

func resolveCallEventTime(
	startTime int64,
	endTime int64,
	createTime int64,
	durationSecond int,
	createdAt time.Time,
) time.Time {
	switch {
	case endTime > 0:
		return time.UnixMilli(endTime).UTC()
	case startTime > 0 && durationSecond > 0:
		return time.UnixMilli(startTime).Add(time.Duration(durationSecond) * time.Second).UTC()
	case startTime > 0:
		return time.UnixMilli(startTime).UTC()
	case createTime > 0:
		return time.UnixMilli(createTime).UTC()
	default:
		return createdAt.UTC()
	}
}
