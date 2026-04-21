package repository

import (
	"backend/internal/model"
	"context"
	"fmt"
	"sort"
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
	salesDailyScoreTelemarketingRoleNames = []string{
		"sales_inside",
		"sale_inside",
	}
	salesDailyScoreTelemarketingRoleLabels = []string{
		"Inside销售",
		"inside销售",
		"电销员工",
		"电销",
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
	ListEnabledTelemarketingUsersByUsernames(ctx context.Context, usernames []string) (map[string]model.TelemarketingLocalUser, error)
	SpxxjjListEnabledTelemarketingUsersByWorkNumbers(ctx context.Context, workNumbers []string) (map[string]model.TelemarketingLocalUser, error)
	CountNewCustomersByTelemarketingUserBetween(ctx context.Context, userIDs []int64, startUTC, endUTC time.Time) (map[int64]int, error)
	CountCustomerVisitsByInvitersOnDate(ctx context.Context, scoreDate string, inviters []string) (map[string]int, error)
	ListNewCustomerEventTimesByTelemarketingUserBetween(ctx context.Context, userIDs []int64, startUTC, endUTC time.Time) (map[int64][]time.Time, error)
	ListCustomerVisitEventTimesByInvitersOnDate(ctx context.Context, scoreDate string, inviters []string) (map[string][]time.Time, error)
	UpsertBatch(ctx context.Context, items []model.SalesDailyScoreUpsertInput) ([]model.SalesDailyScore, error)
	ListByDate(ctx context.Context, scoreDate string, actorUserID int64, actorRole string) ([]model.SalesDailyScore, error)
	SpxxjjUpsertMiHuaSeatStatistics(ctx context.Context, items []model.SpxxjjMiHuaSeatStatisticUpsertInput) error
	SpxxjjUpsertTelemarketingDailyScores(ctx context.Context, items []model.SpxxjjTelemarketingDailyScoreUpsertInput) ([]model.TelemarketingDailyScore, error)
	SpxxjjListTelemarketingDailyScoresByDate(ctx context.Context, scoreDate string) ([]model.TelemarketingDailyScore, error)
	ListRankingLeaderboard(ctx context.Context, startDate string, endDate string) ([]model.RankingLeaderboardItem, error)
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

type spxxjjMiHuaSeatStatisticRow struct {
	ID                     int64      `gorm:"column:id;primaryKey;autoIncrement"`
	ScoreDate              string     `gorm:"column:score_date"`
	SeatID                 int64      `gorm:"column:seat_id"`
	SeatName               string     `gorm:"column:seat_name"`
	WorkNumber             string     `gorm:"column:work_number"`
	ServiceNumber          string     `gorm:"column:service_number"`
	IsMobileSeat           string     `gorm:"column:is_mobile_seat"`
	SeatType               int        `gorm:"column:seat_type"`
	Ccgeid                 int64      `gorm:"column:ccgeid"`
	SuccessCallCount       int        `gorm:"column:success_call_count"`
	OutTotalSuccess        int        `gorm:"column:out_total_success"`
	OutTotalCallCount      int        `gorm:"column:out_total_call_count"`
	CallTotalTimeSecond    int        `gorm:"column:call_total_time_second"`
	CallValidTimeSecond    int        `gorm:"column:call_valid_time_second"`
	OutCallTotalTimeSecond int        `gorm:"column:out_call_total_time_second"`
	OutCallValidTimeSecond int        `gorm:"column:out_call_valid_time_second"`
	LatestStateTime        *time.Time `gorm:"column:latest_state_time"`
	LatestStateID          int        `gorm:"column:latest_state_id"`
	StatTimestamp          *time.Time `gorm:"column:stat_timestamp"`
	EnterpriseName         string     `gorm:"column:enterprise_name"`
	DepartmentName         string     `gorm:"column:department_name"`
	GroupName              string     `gorm:"column:group_name"`
	SeatRealTimeStateJSON  string     `gorm:"column:seat_real_time_state_json"`
	GroupsJSON             string     `gorm:"column:groups_json"`
	RawPayload             string     `gorm:"column:raw_payload"`
	MatchedUserID          *int64     `gorm:"column:matched_user_id"`
	MatchedUserName        string     `gorm:"column:matched_user_name"`
	RoleName               string     `gorm:"column:role_name"`
	CreatedAt              time.Time  `gorm:"column:created_at"`
	UpdatedAt              time.Time  `gorm:"column:updated_at"`
}

type spxxjjTelemarketingDailyScoreRow struct {
	ID                  int64      `gorm:"column:id;primaryKey;autoIncrement"`
	ScoreDate           string     `gorm:"column:score_date"`
	SeatWorkNumber      string     `gorm:"column:seat_work_number"`
	SeatName            string     `gorm:"column:seat_name"`
	MatchedUserID       *int64     `gorm:"column:matched_user_id"`
	MatchedUserName     string     `gorm:"column:matched_user_name"`
	ServiceNumber       string     `gorm:"column:service_number"`
	GroupName           string     `gorm:"column:group_name"`
	RoleName            string     `gorm:"column:role_name"`
	CallNum             int        `gorm:"column:call_num"`
	AnsweredCallCount   int        `gorm:"column:answered_call_count"`
	MissedCallCount     int        `gorm:"column:missed_call_count"`
	AnswerRate          float64    `gorm:"column:answer_rate"`
	CallDurationSecond  int        `gorm:"column:call_duration_second"`
	NewCustomerCount    int        `gorm:"column:new_customer_count"`
	InvitationCount     int        `gorm:"column:invitation_count"`
	CallScoreByCount    int        `gorm:"column:call_score_by_count"`
	CallScoreByDuration int        `gorm:"column:call_score_by_duration"`
	CallScoreType       string     `gorm:"column:call_score_type"`
	CallScore           int        `gorm:"column:call_score"`
	InvitationScore     int        `gorm:"column:invitation_score"`
	NewCustomerScore    int        `gorm:"column:new_customer_score"`
	TotalScore          int        `gorm:"column:total_score"`
	ScoreReachedAt      *time.Time `gorm:"column:score_reached_at"`
	DataUpdatedAt       *time.Time `gorm:"column:data_updated_at"`
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

type inviterCountRow struct {
	Inviter string `gorm:"column:inviter"`
	Count   int    `gorm:"column:item_count"`
}

type rankingLeaderboardIdentityRow struct {
	SeatWorkNumber  string `gorm:"column:seat_work_number"`
	MatchedUserID   *int64 `gorm:"column:matched_user_id"`
	SeatName        string `gorm:"column:seat_name"`
	MatchedUserName string `gorm:"column:matched_user_name"`
	GroupName       string `gorm:"column:group_name"`
	RoleName        string `gorm:"column:role_name"`
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

func (r *gormSalesDailyScoreRepository) ListEnabledTelemarketingUsersByUsernames(
	ctx context.Context,
	usernames []string,
) (map[string]model.TelemarketingLocalUser, error) {
	usernames = uniqueTrimmedStrings(usernames)
	if len(usernames) == 0 {
		return map[string]model.TelemarketingLocalUser{}, nil
	}

	type telemarketingUserRow struct {
		UserID   int64  `gorm:"column:user_id"`
		Username string `gorm:"column:username"`
		Nickname string `gorm:"column:nickname"`
		UserName string `gorm:"column:user_name"`
		RoleName string `gorm:"column:role_name"`
	}

	var rows []telemarketingUserRow
	err := r.db.WithContext(ctx).
		Table("users AS u").
		Select(
			"u.id AS user_id",
			"u.username AS username",
			"COALESCE(NULLIF(u.nickname, ''), '') AS nickname",
			"COALESCE(NULLIF(u.nickname, ''), NULLIF(u.username, ''), '') AS user_name",
			"CASE WHEN COALESCE(r.label, '') <> '' THEN r.label ELSE COALESCE(r.name, '') END AS role_name",
		).
		Joins("JOIN roles AS r ON r.id = u.role_id").
		Where("u.status = ?", model.UserStatusEnabled).
		Where("u.username IN ?", usernames).
		Where("(r.name IN ? OR r.label IN ?)", salesDailyScoreTelemarketingRoleNames, salesDailyScoreTelemarketingRoleLabels).
		Order("u.id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]model.TelemarketingLocalUser, len(rows))
	for _, row := range rows {
		key := strings.TrimSpace(row.Username)
		if key == "" {
			continue
		}
		result[key] = model.TelemarketingLocalUser{
			UserID:   row.UserID,
			Username: key,
			Nickname: strings.TrimSpace(row.Nickname),
			UserName: strings.TrimSpace(row.UserName),
			RoleName: strings.TrimSpace(row.RoleName),
		}
	}
	return result, nil
}

func (r *gormSalesDailyScoreRepository) SpxxjjListEnabledTelemarketingUsersByWorkNumbers(
	ctx context.Context,
	workNumbers []string,
) (map[string]model.TelemarketingLocalUser, error) {
	workNumbers = uniqueTrimmedStrings(workNumbers)
	if len(workNumbers) == 0 {
		return map[string]model.TelemarketingLocalUser{}, nil
	}

	type telemarketingUserRow struct {
		UserID     int64  `gorm:"column:user_id"`
		Username   string `gorm:"column:username"`
		Nickname   string `gorm:"column:nickname"`
		UserName   string `gorm:"column:user_name"`
		RoleName   string `gorm:"column:role_name"`
		WorkNumber string `gorm:"column:work_number"`
	}

	var rows []telemarketingUserRow
	err := r.db.WithContext(ctx).
		Table("users AS u").
		Select(
			"u.id AS user_id",
			"u.username AS username",
			"COALESCE(NULLIF(u.nickname, ''), '') AS nickname",
			"COALESCE(NULLIF(u.nickname, ''), NULLIF(u.username, ''), '') AS user_name",
			"CASE WHEN COALESCE(r.label, '') <> '' THEN r.label ELSE COALESCE(r.name, '') END AS role_name",
			"COALESCE(NULLIF(u.mihua_work_number, ''), '') AS work_number",
		).
		Joins("JOIN roles AS r ON r.id = u.role_id").
		Where("u.status = ?", model.UserStatusEnabled).
		Where("COALESCE(NULLIF(u.mihua_work_number, ''), '') IN ?", workNumbers).
		Where("(r.name IN ? OR r.label IN ?)", salesDailyScoreTelemarketingRoleNames, salesDailyScoreTelemarketingRoleLabels).
		Order("u.id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]model.TelemarketingLocalUser, len(rows))
	for _, row := range rows {
		key := strings.TrimSpace(row.WorkNumber)
		if key == "" {
			continue
		}
		result[key] = model.TelemarketingLocalUser{
			UserID:     row.UserID,
			Username:   strings.TrimSpace(row.Username),
			Nickname:   strings.TrimSpace(row.Nickname),
			UserName:   strings.TrimSpace(row.UserName),
			RoleName:   strings.TrimSpace(row.RoleName),
			WorkNumber: key,
		}
	}
	return result, nil
}

func (r *gormSalesDailyScoreRepository) CountNewCustomersByTelemarketingUserBetween(
	ctx context.Context,
	userIDs []int64,
	startUTC, endUTC time.Time,
) (map[int64]int, error) {
	userIDs = uniquePositiveInt64Slice(userIDs)
	if len(userIDs) == 0 {
		return map[int64]int{}, nil
	}

	const telemarketingUserExpr = "CASE WHEN inside_sales_user_id IS NOT NULL AND inside_sales_user_id > 0 THEN inside_sales_user_id ELSE create_user_id END"

	var rows []countByUserRow
	err := r.db.WithContext(ctx).
		Table("customers").
		Select(telemarketingUserExpr+" AS user_id", "COUNT(DISTINCT id) AS item_count").
		Where("created_at >= ? AND created_at < ?", startUTC, endUTC).
		Where(telemarketingUserExpr+" IN ?", userIDs).
		Group(telemarketingUserExpr).
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

func (r *gormSalesDailyScoreRepository) CountCustomerVisitsByInvitersOnDate(
	ctx context.Context,
	scoreDate string,
	inviters []string,
) (map[string]int, error) {
	inviters = uniqueTrimmedStrings(inviters)
	if len(inviters) == 0 {
		return map[string]int{}, nil
	}

	var rows []inviterCountRow
	err := r.db.WithContext(ctx).
		Table("customer_visits").
		Select("TRIM(inviter) AS inviter", "COUNT(1) AS item_count").
		Where("visit_date = ?", strings.TrimSpace(scoreDate)).
		Where("TRIM(inviter) IN ?", inviters).
		Group("TRIM(inviter)").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]int, len(rows))
	for _, row := range rows {
		key := strings.TrimSpace(row.Inviter)
		if key == "" {
			continue
		}
		result[key] = row.Count
	}
	return result, nil
}

func (r *gormSalesDailyScoreRepository) ListNewCustomerEventTimesByTelemarketingUserBetween(
	ctx context.Context,
	userIDs []int64,
	startUTC, endUTC time.Time,
) (map[int64][]time.Time, error) {
	userIDs = uniquePositiveInt64Slice(userIDs)
	if len(userIDs) == 0 {
		return map[int64][]time.Time{}, nil
	}

	const telemarketingUserExpr = "CASE WHEN inside_sales_user_id IS NOT NULL AND inside_sales_user_id > 0 THEN inside_sales_user_id ELSE create_user_id END"

	var rows []timeByUserRow
	err := r.db.WithContext(ctx).
		Table("customers").
		Select(telemarketingUserExpr+" AS user_id", "created_at AS event_time").
		Where("created_at >= ? AND created_at < ?", startUTC, endUTC).
		Where(telemarketingUserExpr+" IN ?", userIDs).
		Order(telemarketingUserExpr + " ASC, created_at ASC, id ASC").
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

func (r *gormSalesDailyScoreRepository) ListCustomerVisitEventTimesByInvitersOnDate(
	ctx context.Context,
	scoreDate string,
	inviters []string,
) (map[string][]time.Time, error) {
	inviters = uniqueTrimmedStrings(inviters)
	if len(inviters) == 0 {
		return map[string][]time.Time{}, nil
	}

	type inviterTimeRow struct {
		Inviter   string    `gorm:"column:inviter"`
		EventTime time.Time `gorm:"column:event_time"`
	}

	var rows []inviterTimeRow
	err := r.db.WithContext(ctx).
		Table("customer_visits").
		Select("TRIM(inviter) AS inviter", "created_at AS event_time").
		Where("visit_date = ?", strings.TrimSpace(scoreDate)).
		Where("TRIM(inviter) IN ?", inviters).
		Order("TRIM(inviter) ASC, created_at ASC, id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string][]time.Time, len(rows))
	for _, row := range rows {
		key := strings.TrimSpace(row.Inviter)
		if key == "" {
			continue
		}
		result[key] = append(result[key], row.EventTime)
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

func (r *gormSalesDailyScoreRepository) SpxxjjUpsertMiHuaSeatStatistics(
	ctx context.Context,
	items []model.SpxxjjMiHuaSeatStatisticUpsertInput,
) error {
	items = dedupeSpxxjjMiHuaSeatStatisticInputs(items)
	if len(items) == 0 {
		return nil
	}

	now := time.Now().UTC()
	for _, item := range items {
		row := spxxjjMiHuaSeatStatisticRow{
			ScoreDate:              strings.TrimSpace(item.ScoreDate),
			SeatID:                 item.SeatID,
			SeatName:               strings.TrimSpace(item.SeatName),
			WorkNumber:             strings.TrimSpace(item.WorkNumber),
			ServiceNumber:          strings.TrimSpace(item.ServiceNumber),
			IsMobileSeat:           strings.TrimSpace(item.IsMobileSeat),
			SeatType:               item.SeatType,
			Ccgeid:                 item.Ccgeid,
			SuccessCallCount:       item.SuccessCallCount,
			OutTotalSuccess:        item.OutTotalSuccess,
			OutTotalCallCount:      item.OutTotalCallCount,
			CallTotalTimeSecond:    item.CallTotalTimeSecond,
			CallValidTimeSecond:    item.CallValidTimeSecond,
			OutCallTotalTimeSecond: item.OutCallTotalTimeSecond,
			OutCallValidTimeSecond: item.OutCallValidTimeSecond,
			LatestStateTime:        item.LatestStateTime,
			LatestStateID:          item.LatestStateID,
			StatTimestamp:          item.StatTimestamp,
			EnterpriseName:         strings.TrimSpace(item.EnterpriseName),
			DepartmentName:         strings.TrimSpace(item.DepartmentName),
			GroupName:              strings.TrimSpace(item.GroupName),
			SeatRealTimeStateJSON:  strings.TrimSpace(item.SeatRealTimeStateJSON),
			GroupsJSON:             strings.TrimSpace(item.GroupsJSON),
			RawPayload:             strings.TrimSpace(item.RawPayload),
			MatchedUserID:          item.MatchedUserID,
			MatchedUserName:        strings.TrimSpace(item.MatchedUserName),
			RoleName:               strings.TrimSpace(item.RoleName),
			CreatedAt:              now,
			UpdatedAt:              now,
		}

		if err := r.db.WithContext(ctx).Table("spxxjj_mihua_seat_statistics").Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "score_date"},
				{Name: "work_number"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"seat_id":                    row.SeatID,
				"seat_name":                  row.SeatName,
				"service_number":             row.ServiceNumber,
				"is_mobile_seat":             row.IsMobileSeat,
				"seat_type":                  row.SeatType,
				"ccgeid":                     row.Ccgeid,
				"success_call_count":         row.SuccessCallCount,
				"out_total_success":          row.OutTotalSuccess,
				"out_total_call_count":       row.OutTotalCallCount,
				"call_total_time_second":     row.CallTotalTimeSecond,
				"call_valid_time_second":     row.CallValidTimeSecond,
				"out_call_total_time_second": row.OutCallTotalTimeSecond,
				"out_call_valid_time_second": row.OutCallValidTimeSecond,
				"latest_state_time":          row.LatestStateTime,
				"latest_state_id":            row.LatestStateID,
				"stat_timestamp":             row.StatTimestamp,
				"enterprise_name":            row.EnterpriseName,
				"department_name":            row.DepartmentName,
				"group_name":                 row.GroupName,
				"seat_real_time_state_json":  row.SeatRealTimeStateJSON,
				"groups_json":                row.GroupsJSON,
				"raw_payload":                row.RawPayload,
				"matched_user_id":            row.MatchedUserID,
				"matched_user_name":          row.MatchedUserName,
				"role_name":                  row.RoleName,
				"updated_at":                 row.UpdatedAt,
			}),
		}).Create(&row).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *gormSalesDailyScoreRepository) SpxxjjUpsertTelemarketingDailyScores(
	ctx context.Context,
	items []model.SpxxjjTelemarketingDailyScoreUpsertInput,
) ([]model.TelemarketingDailyScore, error) {
	items = dedupeSpxxjjTelemarketingDailyScoreInputs(items)
	if len(items) == 0 {
		return []model.TelemarketingDailyScore{}, nil
	}

	now := time.Now().UTC()
	result := make([]model.TelemarketingDailyScore, 0, len(items))
	for _, item := range items {
		row := spxxjjTelemarketingDailyScoreRow{
			ScoreDate:           strings.TrimSpace(item.ScoreDate),
			SeatWorkNumber:      strings.TrimSpace(item.SeatWorkNumber),
			SeatName:            strings.TrimSpace(item.SeatName),
			MatchedUserID:       item.MatchedUserID,
			MatchedUserName:     strings.TrimSpace(item.MatchedUserName),
			ServiceNumber:       strings.TrimSpace(item.ServiceNumber),
			GroupName:           strings.TrimSpace(item.GroupName),
			RoleName:            strings.TrimSpace(item.RoleName),
			CallNum:             item.CallNum,
			AnsweredCallCount:   item.AnsweredCallCount,
			MissedCallCount:     item.MissedCallCount,
			AnswerRate:          item.AnswerRate,
			CallDurationSecond:  item.CallDurationSecond,
			NewCustomerCount:    item.NewCustomerCount,
			InvitationCount:     item.InvitationCount,
			CallScoreByCount:    item.CallScoreByCount,
			CallScoreByDuration: item.CallScoreByDuration,
			CallScoreType:       strings.TrimSpace(item.CallScoreType),
			CallScore:           item.CallScore,
			InvitationScore:     item.InvitationScore,
			NewCustomerScore:    item.NewCustomerScore,
			TotalScore:          item.TotalScore,
			ScoreReachedAt:      item.ScoreReachedAt,
			DataUpdatedAt:       item.DataUpdatedAt,
			CreatedAt:           now,
			UpdatedAt:           now,
		}

		err := r.db.WithContext(ctx).Table("spxxjj_telemarketing_daily_scores").Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "score_date"},
				{Name: "seat_work_number"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"seat_name":              row.SeatName,
				"matched_user_id":        row.MatchedUserID,
				"matched_user_name":      row.MatchedUserName,
				"service_number":         row.ServiceNumber,
				"group_name":             row.GroupName,
				"role_name":              row.RoleName,
				"call_num":               row.CallNum,
				"answered_call_count":    row.AnsweredCallCount,
				"missed_call_count":      row.MissedCallCount,
				"answer_rate":            row.AnswerRate,
				"call_duration_second":   row.CallDurationSecond,
				"new_customer_count":     row.NewCustomerCount,
				"invitation_count":       row.InvitationCount,
				"call_score_by_count":    row.CallScoreByCount,
				"call_score_by_duration": row.CallScoreByDuration,
				"call_score_type":        row.CallScoreType,
				"call_score":             row.CallScore,
				"invitation_score":       row.InvitationScore,
				"new_customer_score":     row.NewCustomerScore,
				"total_score":            row.TotalScore,
				"score_reached_at":       row.ScoreReachedAt,
				"data_updated_at":        row.DataUpdatedAt,
				"updated_at":             row.UpdatedAt,
			}),
		}).Create(&row).Error
		if err != nil {
			return nil, err
		}

		var saved spxxjjTelemarketingDailyScoreRow
		if err := r.db.WithContext(ctx).
			Table("spxxjj_telemarketing_daily_scores").
			Where("score_date = ? AND seat_work_number = ?", row.ScoreDate, row.SeatWorkNumber).
			Take(&saved).Error; err != nil {
			return nil, err
		}
		result = append(result, mapSpxxjjTelemarketingDailyScoreRowToModel(saved))
	}

	return result, nil
}

func (r *gormSalesDailyScoreRepository) ListRankingLeaderboard(
	ctx context.Context,
	startDate string,
	endDate string,
) ([]model.RankingLeaderboardItem, error) {
	var rows []model.RankingLeaderboardItem
	aggregateKeyExpr := "CASE WHEN TRIM(COALESCE(s.seat_work_number, '')) <> '' THEN CONCAT('w:', TRIM(s.seat_work_number)) WHEN s.matched_user_id IS NOT NULL AND s.matched_user_id > 0 THEN CONCAT('u:', CAST(s.matched_user_id AS CHAR)) ELSE '' END"
	err := r.db.WithContext(ctx).
		Table("spxxjj_telemarketing_daily_scores AS s").
		Select(
			aggregateKeyExpr+" AS aggregate_key",
			"MAX(NULLIF(TRIM(s.seat_work_number), '')) AS seat_work_number",
			"MAX(s.matched_user_id) AS matched_user_id",
			"MAX(NULLIF(TRIM(s.seat_name), '')) AS seat_name",
			"MAX(NULLIF(TRIM(s.matched_user_name), '')) AS matched_user_name",
			"MAX(NULLIF(TRIM(s.group_name), '')) AS group_name",
			"MAX(NULLIF(TRIM(s.role_name), '')) AS role_name",
			"COALESCE(SUM(s.call_num), 0) AS call_num",
			"COALESCE(SUM(s.answered_call_count), 0) AS answered_call_count",
			"CASE WHEN COALESCE(SUM(s.call_num), 0) <= 0 THEN 0 ELSE ROUND(COALESCE(SUM(s.answered_call_count), 0) * 100.0 / COALESCE(SUM(s.call_num), 0), 1) END AS answer_rate",
			"COALESCE(SUM(s.call_duration_second), 0) AS call_duration_second",
			"COALESCE(SUM(s.new_customer_count), 0) AS new_customer_count",
			"COALESCE(SUM(s.invitation_count), 0) AS invitation_count",
			"COALESCE(SUM(s.call_score), 0) AS call_score",
			"COALESCE(SUM(s.invitation_score), 0) AS invitation_score",
			"COALESCE(SUM(s.new_customer_score), 0) AS new_customer_score",
			"COALESCE(SUM(s.total_score), 0) AS total_score",
			"COUNT(DISTINCT s.score_date) AS score_days",
		).
		Where("s.score_date >= ? AND s.score_date <= ?", strings.TrimSpace(startDate), strings.TrimSpace(endDate)).
		Where(aggregateKeyExpr + " <> ''").
		Group(aggregateKeyExpr).
		Order("total_score DESC, answered_call_count DESC, call_duration_second DESC, invitation_count DESC, new_customer_count DESC, call_num DESC, score_days DESC, seat_work_number ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []model.RankingLeaderboardItem{}, nil
	}

	if err := r.fillRankingLeaderboardIdentities(ctx, rows); err != nil {
		return nil, err
	}
	return mergeRankingLeaderboardItems(rows), nil
}

func (r *gormSalesDailyScoreRepository) fillRankingLeaderboardIdentities(
	ctx context.Context,
	items []model.RankingLeaderboardItem,
) error {
	if len(items) == 0 {
		return nil
	}

	workNumbers := make([]string, 0, len(items))
	userIDs := make([]int64, 0, len(items))
	for _, item := range items {
		if workNumber := strings.TrimSpace(item.SeatWorkNumber); workNumber != "" {
			workNumbers = append(workNumbers, workNumber)
		}
		if item.MatchedUserID != nil && *item.MatchedUserID > 0 {
			userIDs = append(userIDs, *item.MatchedUserID)
		}
	}

	rawByWorkNumber, err := r.listRankingLeaderboardRawByWorkNumbers(ctx, workNumbers)
	if err != nil {
		return err
	}
	localUserByWorkNumber, err := r.SpxxjjListEnabledTelemarketingUsersByWorkNumbers(ctx, workNumbers)
	if err != nil {
		return err
	}
	rawByUserID, err := r.listRankingLeaderboardRawByUserIDs(ctx, userIDs)
	if err != nil {
		return err
	}
	userByID, err := r.listRankingLeaderboardUsersByIDs(ctx, userIDs)
	if err != nil {
		return err
	}

	for idx := range items {
		item := &items[idx]

		var rawByWork rankingLeaderboardIdentityRow
		var localUserByWork model.TelemarketingLocalUser
		if workNumber := strings.TrimSpace(item.SeatWorkNumber); workNumber != "" {
			rawByWork = rawByWorkNumber[workNumber]
			localUserByWork = localUserByWorkNumber[workNumber]
			if item.MatchedUserID == nil && rawByWork.MatchedUserID != nil && *rawByWork.MatchedUserID > 0 {
				matchedUserID := *rawByWork.MatchedUserID
				item.MatchedUserID = &matchedUserID
			}
			if item.MatchedUserID == nil && localUserByWork.UserID > 0 {
				matchedUserID := localUserByWork.UserID
				item.MatchedUserID = &matchedUserID
			}
		}

		var rawByUser rankingLeaderboardIdentityRow
		var localUser model.TelemarketingLocalUser
		if item.MatchedUserID != nil && *item.MatchedUserID > 0 {
			rawByUser = rawByUserID[*item.MatchedUserID]
			localUser = userByID[*item.MatchedUserID]
		}

		if strings.TrimSpace(item.SeatWorkNumber) == "" {
			item.SeatWorkNumber = firstNonEmptyTelemarketing(
				rawByWork.SeatWorkNumber,
				rawByUser.SeatWorkNumber,
				localUser.WorkNumber,
			)
		}
		item.SeatName = firstNonEmptyTelemarketing(
			item.SeatName,
			rawByWork.SeatName,
			rawByUser.SeatName,
			localUserByWork.UserName,
			localUserByWork.Nickname,
			localUserByWork.Username,
			localUser.UserName,
			localUser.Nickname,
			localUser.Username,
		)
		item.MatchedUserName = firstNonEmptyTelemarketing(
			item.MatchedUserName,
			rawByWork.MatchedUserName,
			rawByUser.MatchedUserName,
			localUserByWork.UserName,
			localUserByWork.Nickname,
			localUserByWork.Username,
			localUser.UserName,
			localUser.Nickname,
			localUser.Username,
			item.SeatName,
		)
		item.GroupName = firstNonEmptyTelemarketing(
			item.GroupName,
			rawByWork.GroupName,
			rawByUser.GroupName,
		)
		item.RoleName = firstNonEmptyTelemarketing(
			item.RoleName,
			rawByWork.RoleName,
			rawByUser.RoleName,
			localUserByWork.RoleName,
			localUser.RoleName,
			"电销",
		)
	}

	return nil
}

func (r *gormSalesDailyScoreRepository) listRankingLeaderboardRawByWorkNumbers(
	ctx context.Context,
	workNumbers []string,
) (map[string]rankingLeaderboardIdentityRow, error) {
	workNumbers = uniqueTrimmedStrings(workNumbers)
	if len(workNumbers) == 0 {
		return map[string]rankingLeaderboardIdentityRow{}, nil
	}

	var rows []rankingLeaderboardIdentityRow
	err := r.db.WithContext(ctx).
		Table("spxxjj_mihua_seat_statistics").
		Select(
			"MAX(NULLIF(TRIM(work_number), '')) AS seat_work_number",
			"MAX(matched_user_id) AS matched_user_id",
			"MAX(NULLIF(TRIM(seat_name), '')) AS seat_name",
			"MAX(NULLIF(TRIM(matched_user_name), '')) AS matched_user_name",
			"MAX(NULLIF(TRIM(group_name), '')) AS group_name",
			"MAX(NULLIF(TRIM(role_name), '')) AS role_name",
		).
		Where("TRIM(COALESCE(work_number, '')) IN ?", workNumbers).
		Group("work_number").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]rankingLeaderboardIdentityRow, len(rows))
	for _, row := range rows {
		if key := strings.TrimSpace(row.SeatWorkNumber); key != "" {
			result[key] = row
		}
	}
	return result, nil
}

func (r *gormSalesDailyScoreRepository) listRankingLeaderboardRawByUserIDs(
	ctx context.Context,
	userIDs []int64,
) (map[int64]rankingLeaderboardIdentityRow, error) {
	userIDs = uniquePositiveInt64Slice(userIDs)
	if len(userIDs) == 0 {
		return map[int64]rankingLeaderboardIdentityRow{}, nil
	}

	type rawByUserRow struct {
		UserID          int64  `gorm:"column:user_id"`
		SeatWorkNumber  string `gorm:"column:seat_work_number"`
		MatchedUserID   *int64 `gorm:"column:matched_user_id"`
		SeatName        string `gorm:"column:seat_name"`
		MatchedUserName string `gorm:"column:matched_user_name"`
		GroupName       string `gorm:"column:group_name"`
		RoleName        string `gorm:"column:role_name"`
	}

	var rows []rawByUserRow
	err := r.db.WithContext(ctx).
		Table("spxxjj_mihua_seat_statistics").
		Select(
			"matched_user_id AS user_id",
			"MAX(NULLIF(TRIM(work_number), '')) AS seat_work_number",
			"MAX(matched_user_id) AS matched_user_id",
			"MAX(NULLIF(TRIM(seat_name), '')) AS seat_name",
			"MAX(NULLIF(TRIM(matched_user_name), '')) AS matched_user_name",
			"MAX(NULLIF(TRIM(group_name), '')) AS group_name",
			"MAX(NULLIF(TRIM(role_name), '')) AS role_name",
		).
		Where("matched_user_id IN ?", userIDs).
		Group("matched_user_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64]rankingLeaderboardIdentityRow, len(rows))
	for _, row := range rows {
		if row.UserID > 0 {
			result[row.UserID] = rankingLeaderboardIdentityRow{
				SeatWorkNumber:  strings.TrimSpace(row.SeatWorkNumber),
				MatchedUserID:   row.MatchedUserID,
				SeatName:        strings.TrimSpace(row.SeatName),
				MatchedUserName: strings.TrimSpace(row.MatchedUserName),
				GroupName:       strings.TrimSpace(row.GroupName),
				RoleName:        strings.TrimSpace(row.RoleName),
			}
		}
	}
	return result, nil
}

func (r *gormSalesDailyScoreRepository) listRankingLeaderboardUsersByIDs(
	ctx context.Context,
	userIDs []int64,
) (map[int64]model.TelemarketingLocalUser, error) {
	userIDs = uniquePositiveInt64Slice(userIDs)
	if len(userIDs) == 0 {
		return map[int64]model.TelemarketingLocalUser{}, nil
	}

	type telemarketingUserRow struct {
		UserID     int64  `gorm:"column:user_id"`
		Username   string `gorm:"column:username"`
		Nickname   string `gorm:"column:nickname"`
		UserName   string `gorm:"column:user_name"`
		RoleName   string `gorm:"column:role_name"`
		WorkNumber string `gorm:"column:work_number"`
	}

	var rows []telemarketingUserRow
	err := r.db.WithContext(ctx).
		Table("users AS u").
		Select(
			"u.id AS user_id",
			"u.username AS username",
			"COALESCE(NULLIF(u.nickname, ''), '') AS nickname",
			"COALESCE(NULLIF(u.nickname, ''), NULLIF(u.username, ''), '') AS user_name",
			"CASE WHEN COALESCE(r.label, '') <> '' THEN r.label ELSE COALESCE(r.name, '') END AS role_name",
			"COALESCE(NULLIF(u.mihua_work_number, ''), '') AS work_number",
		).
		Joins("JOIN roles AS r ON r.id = u.role_id").
		Where("u.status = ?", model.UserStatusEnabled).
		Where("u.id IN ?", userIDs).
		Order("u.id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64]model.TelemarketingLocalUser, len(rows))
	for _, row := range rows {
		if row.UserID <= 0 {
			continue
		}
		result[row.UserID] = model.TelemarketingLocalUser{
			UserID:     row.UserID,
			Username:   strings.TrimSpace(row.Username),
			Nickname:   strings.TrimSpace(row.Nickname),
			UserName:   strings.TrimSpace(row.UserName),
			RoleName:   strings.TrimSpace(row.RoleName),
			WorkNumber: strings.TrimSpace(row.WorkNumber),
		}
	}
	return result, nil
}

func mergeRankingLeaderboardItems(items []model.RankingLeaderboardItem) []model.RankingLeaderboardItem {
	if len(items) <= 1 {
		return items
	}

	order := make([]string, 0, len(items))
	merged := make(map[string]model.RankingLeaderboardItem, len(items))
	for _, item := range items {
		key := rankingLeaderboardMergeKey(item)
		item.IdentityKey = key
		if _, exists := merged[key]; !exists {
			order = append(order, key)
			merged[key] = item
			continue
		}

		current := merged[key]
		current.IdentityKey = key
		current.SeatWorkNumber = firstNonEmptyTelemarketing(current.SeatWorkNumber, item.SeatWorkNumber)
		if current.MatchedUserID == nil && item.MatchedUserID != nil && *item.MatchedUserID > 0 {
			matchedUserID := *item.MatchedUserID
			current.MatchedUserID = &matchedUserID
		}
		current.SeatName = firstNonEmptyTelemarketing(current.SeatName, item.SeatName)
		current.MatchedUserName = firstNonEmptyTelemarketing(current.MatchedUserName, item.MatchedUserName)
		current.GroupName = firstNonEmptyTelemarketing(current.GroupName, item.GroupName)
		current.RoleName = firstNonEmptyTelemarketing(current.RoleName, item.RoleName, "电销")
		current.CallNum += item.CallNum
		current.AnsweredCallCount += item.AnsweredCallCount
		current.CallDurationSecond += item.CallDurationSecond
		current.NewCustomerCount += item.NewCustomerCount
		current.InvitationCount += item.InvitationCount
		current.CallScore += item.CallScore
		current.InvitationScore += item.InvitationScore
		current.NewCustomerScore += item.NewCustomerScore
		current.TotalScore += item.TotalScore
		current.ScoreDays += item.ScoreDays
		if current.CallNum > 0 {
			current.AnswerRate = float64(current.AnsweredCallCount) * 100 / float64(current.CallNum)
		} else {
			current.AnswerRate = 0
		}
		merged[key] = current
	}

	result := make([]model.RankingLeaderboardItem, 0, len(order))
	for _, key := range order {
		result = append(result, merged[key])
	}
	sort.Slice(result, func(i, j int) bool {
		left := result[i]
		right := result[j]
		switch {
		case left.TotalScore != right.TotalScore:
			return left.TotalScore > right.TotalScore
		case left.AnsweredCallCount != right.AnsweredCallCount:
			return left.AnsweredCallCount > right.AnsweredCallCount
		case left.CallDurationSecond != right.CallDurationSecond:
			return left.CallDurationSecond > right.CallDurationSecond
		case left.InvitationCount != right.InvitationCount:
			return left.InvitationCount > right.InvitationCount
		case left.NewCustomerCount != right.NewCustomerCount:
			return left.NewCustomerCount > right.NewCustomerCount
		case left.CallNum != right.CallNum:
			return left.CallNum > right.CallNum
		case left.ScoreDays != right.ScoreDays:
			return left.ScoreDays > right.ScoreDays
		default:
			return strings.Compare(strings.TrimSpace(left.SeatWorkNumber), strings.TrimSpace(right.SeatWorkNumber)) < 0
		}
	})
	return result
}

func rankingLeaderboardMergeKey(item model.RankingLeaderboardItem) string {
	if workNumber := strings.TrimSpace(item.SeatWorkNumber); workNumber != "" {
		return "w:" + workNumber
	}
	if item.MatchedUserID != nil && *item.MatchedUserID > 0 {
		return fmt.Sprintf("u:%d", *item.MatchedUserID)
	}
	if identityKey := strings.TrimSpace(item.IdentityKey); identityKey != "" {
		return identityKey
	}
	return "n:" + firstNonEmptyTelemarketing(item.MatchedUserName, item.SeatName, item.GroupName, fmt.Sprintf("%d", item.TotalScore))
}

func (r *gormSalesDailyScoreRepository) SpxxjjListTelemarketingDailyScoresByDate(
	ctx context.Context,
	scoreDate string,
) ([]model.TelemarketingDailyScore, error) {
	var rows []spxxjjTelemarketingDailyScoreRow
	err := r.db.WithContext(ctx).
		Table("spxxjj_telemarketing_daily_scores").
		Where("score_date = ?", strings.TrimSpace(scoreDate)).
		Order(spxxjjTelemarketingDailyScoreOrderClause("")).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []model.TelemarketingDailyScore{}, nil
	}

	items := make([]model.TelemarketingDailyScore, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapSpxxjjTelemarketingDailyScoreRowToModel(row))
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

func dedupeSpxxjjMiHuaSeatStatisticInputs(
	items []model.SpxxjjMiHuaSeatStatisticUpsertInput,
) []model.SpxxjjMiHuaSeatStatisticUpsertInput {
	if len(items) == 0 {
		return []model.SpxxjjMiHuaSeatStatisticUpsertInput{}
	}

	order := make([]string, 0, len(items))
	merged := make(map[string]model.SpxxjjMiHuaSeatStatisticUpsertInput, len(items))
	for _, item := range items {
		scoreDate := strings.TrimSpace(item.ScoreDate)
		workNumber := strings.TrimSpace(item.WorkNumber)
		if scoreDate == "" || workNumber == "" {
			continue
		}
		key := scoreDate + "\x00" + workNumber
		if _, exists := merged[key]; !exists {
			order = append(order, key)
		}
		item.ScoreDate = scoreDate
		item.WorkNumber = workNumber
		item.SeatName = strings.TrimSpace(item.SeatName)
		item.ServiceNumber = strings.TrimSpace(item.ServiceNumber)
		item.IsMobileSeat = strings.TrimSpace(item.IsMobileSeat)
		item.EnterpriseName = strings.TrimSpace(item.EnterpriseName)
		item.DepartmentName = strings.TrimSpace(item.DepartmentName)
		item.GroupName = strings.TrimSpace(item.GroupName)
		item.SeatRealTimeStateJSON = strings.TrimSpace(item.SeatRealTimeStateJSON)
		item.GroupsJSON = strings.TrimSpace(item.GroupsJSON)
		item.RawPayload = strings.TrimSpace(item.RawPayload)
		item.MatchedUserName = strings.TrimSpace(item.MatchedUserName)
		item.RoleName = strings.TrimSpace(item.RoleName)
		merged[key] = item
	}

	result := make([]model.SpxxjjMiHuaSeatStatisticUpsertInput, 0, len(order))
	for _, key := range order {
		result = append(result, merged[key])
	}
	return result
}

func dedupeSpxxjjTelemarketingDailyScoreInputs(
	items []model.SpxxjjTelemarketingDailyScoreUpsertInput,
) []model.SpxxjjTelemarketingDailyScoreUpsertInput {
	if len(items) == 0 {
		return []model.SpxxjjTelemarketingDailyScoreUpsertInput{}
	}

	order := make([]string, 0, len(items))
	merged := make(map[string]model.SpxxjjTelemarketingDailyScoreUpsertInput, len(items))
	for _, item := range items {
		scoreDate := strings.TrimSpace(item.ScoreDate)
		seatWorkNumber := strings.TrimSpace(item.SeatWorkNumber)
		if scoreDate == "" || seatWorkNumber == "" {
			continue
		}
		key := scoreDate + "\x00" + seatWorkNumber
		if _, exists := merged[key]; !exists {
			order = append(order, key)
		}
		item.ScoreDate = scoreDate
		item.SeatWorkNumber = seatWorkNumber
		item.SeatName = strings.TrimSpace(item.SeatName)
		item.MatchedUserName = strings.TrimSpace(item.MatchedUserName)
		item.ServiceNumber = strings.TrimSpace(item.ServiceNumber)
		item.GroupName = strings.TrimSpace(item.GroupName)
		item.RoleName = strings.TrimSpace(item.RoleName)
		item.CallScoreType = strings.TrimSpace(item.CallScoreType)
		merged[key] = item
	}

	result := make([]model.SpxxjjTelemarketingDailyScoreUpsertInput, 0, len(order))
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

func mapSpxxjjTelemarketingDailyScoreRowToModel(row spxxjjTelemarketingDailyScoreRow) model.TelemarketingDailyScore {
	updatedAt := row.UpdatedAt
	if row.DataUpdatedAt != nil && !row.DataUpdatedAt.IsZero() {
		updatedAt = *row.DataUpdatedAt
	}
	return model.TelemarketingDailyScore{
		ScoreDate:           row.ScoreDate,
		SeatWorkNumber:      row.SeatWorkNumber,
		SeatName:            row.SeatName,
		MatchedUserID:       row.MatchedUserID,
		MatchedUserName:     row.MatchedUserName,
		ServiceNumber:       row.ServiceNumber,
		GroupName:           row.GroupName,
		RoleName:            row.RoleName,
		CallNum:             row.CallNum,
		AnsweredCallCount:   row.AnsweredCallCount,
		MissedCallCount:     row.MissedCallCount,
		AnswerRate:          row.AnswerRate,
		CallDurationSecond:  row.CallDurationSecond,
		NewCustomerCount:    row.NewCustomerCount,
		InvitationCount:     row.InvitationCount,
		CallScoreByCount:    row.CallScoreByCount,
		CallScoreByDuration: row.CallScoreByDuration,
		CallScoreType:       row.CallScoreType,
		CallScore:           row.CallScore,
		InvitationScore:     row.InvitationScore,
		NewCustomerScore:    row.NewCustomerScore,
		TotalScore:          row.TotalScore,
		ScoreReachedAt:      row.ScoreReachedAt,
		UpdatedAt:           updatedAt,
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

func spxxjjTelemarketingDailyScoreOrderClause(alias string) string {
	prefix := ""
	if trimmed := strings.TrimSpace(alias); trimmed != "" {
		prefix = trimmed + "."
	}
	return fmt.Sprintf(
		"%stotal_score DESC, %sanswered_call_count DESC, %scall_duration_second DESC, %sinvitation_count DESC, %snew_customer_count DESC, %scall_num DESC, CASE WHEN %sscore_reached_at IS NULL THEN 1 ELSE 0 END ASC, %sscore_reached_at ASC, %sseat_work_number ASC",
		prefix,
		prefix,
		prefix,
		prefix,
		prefix,
		prefix,
		prefix,
		prefix,
		prefix,
	)
}

func uniqueTrimmedStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func uniquePositiveInt64Slice(values []int64) []int64 {
	if len(values) == 0 {
		return []int64{}
	}

	seen := make(map[int64]struct{}, len(values))
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func firstNonEmptyTelemarketing(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
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
