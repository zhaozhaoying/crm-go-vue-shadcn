package repository

import (
	"backend/internal/model"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

var (
	dashboardSalesRoleNames = []string{
		"sales_director",
		"sales_manager",
		"sales_staff",
		"sales_inside",
		"sale_inside",
		"sales_outside",
		"sale_outside",
	}
	dashboardSalesRoleLabels = []string{
		"销售总监",
		"销售经理",
		"销售员工",
		"销售",
		"Inside销售",
		"Outside销售",
		"电销员工",
	}
)

type DashboardRepository interface {
	GetOverview(ctx context.Context, now time.Time, actorUserID int64, actorRole string) (model.DashboardOverview, error)
}

type gormDashboardRepository struct {
	db *gorm.DB
}

func NewGormDashboardRepository(db *gorm.DB) DashboardRepository {
	return &gormDashboardRepository{db: db}
}

func (r *gormDashboardRepository) GetOverview(ctx context.Context, now time.Time, actorUserID int64, actorRole string) (model.DashboardOverview, error) {
	location := now.Location()
	todayStart := dayStart(now, location)
	yesterdayStart := todayStart.AddDate(0, 0, -1)
	tomorrowStart := todayStart.AddDate(0, 0, 1)
	currentMonthStart := monthStart(now, location)
	nextMonthStart := currentMonthStart.AddDate(0, 1, 0)
	lastMonthStart := currentMonthStart.AddDate(0, -1, 0)
	effectiveRole, err := r.resolveDashboardActorRole(ctx, actorUserID, actorRole)
	if err != nil {
		return model.DashboardOverview{}, err
	}
	showAll := isDashboardGlobalRole(effectiveRole)
	scopedUserIDs, err := r.resolveDashboardScopeUserIDs(ctx, actorUserID, effectiveRole, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
	}
	followUpDropDays := r.getPositiveIntSetting(ctx, "follow_up_drop_days", 30)
	dealDropDays := r.getPositiveIntSetting(ctx, "deal_drop_days", 90)

	currentRevenue, err := r.sumContractAmountBetween(ctx, currentMonthStart, nextMonthStart, scopedUserIDs, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
	}
	lastRevenue, err := r.sumContractAmountBetween(ctx, lastMonthStart, currentMonthStart, scopedUserIDs, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
	}

	currentNewCustomers, err := r.countCustomersBetween(ctx, currentMonthStart, nextMonthStart, "", scopedUserIDs, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
	}
	lastNewCustomers, err := r.countCustomersBetween(ctx, lastMonthStart, currentMonthStart, "", scopedUserIDs, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
	}

	currentNewOpportunities, err := r.countCustomersBetween(ctx, currentMonthStart, nextMonthStart, model.CustomerDealStatusUndone, scopedUserIDs, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
	}
	lastNewOpportunities, err := r.countCustomersBetween(ctx, lastMonthStart, currentMonthStart, model.CustomerDealStatusUndone, scopedUserIDs, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
	}

	currentDoneCustomers, err := r.countCustomersBetween(ctx, currentMonthStart, nextMonthStart, model.CustomerDealStatusDone, scopedUserIDs, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
	}
	lastDoneCustomers, err := r.countCustomersBetween(ctx, lastMonthStart, currentMonthStart, model.CustomerDealStatusDone, scopedUserIDs, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
	}

	monthlyRevenue, err := r.listMonthlyRevenue(ctx, currentMonthStart, 12, scopedUserIDs, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
	}

	monthlyContracts, err := r.listMonthlyContractCounts(ctx, currentMonthStart, 12, scopedUserIDs, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
	}

	recentDeals, err := r.listRecentDeals(ctx, 5, scopedUserIDs, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
	}

	recentActivities, err := r.listRecentActivities(ctx, 5, scopedUserIDs, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
	}

	followUpDueSoonCount, err := r.countFollowUpDueSoonCustomers(ctx, now, followUpDropDays, scopedUserIDs, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
	}
	dealDueSoonCount, err := r.countDealDueSoonCustomers(ctx, now, dealDropDays, scopedUserIDs, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
	}
	monthlyFollowUpDropped, err := r.countAutoDropActivitiesBetween(ctx, currentMonthStart, nextMonthStart, model.ActionAutoDropFollowUp, scopedUserIDs, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
	}
	monthlyDealDropped, err := r.countAutoDropActivitiesBetween(ctx, currentMonthStart, nextMonthStart, model.ActionAutoDropDeal, scopedUserIDs, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
	}

	var salesAdminOverview *model.DashboardSalesAdminOverview
	if isDashboardSalesOverviewRole(effectiveRole) {
		todaySalesCustomers, err := r.countSalesCustomersBetween(ctx, todayStart, tomorrowStart, scopedUserIDs, showAll)
		if err != nil {
			return model.DashboardOverview{}, err
		}
		yesterdaySalesCustomers, err := r.countSalesCustomersBetween(ctx, yesterdayStart, todayStart, scopedUserIDs, showAll)
		if err != nil {
			return model.DashboardOverview{}, err
		}
		todaySalesFollowRecords, err := r.countSalesFollowRecordsBetween(ctx, todayStart, tomorrowStart, scopedUserIDs, showAll)
		if err != nil {
			return model.DashboardOverview{}, err
		}
		yesterdaySalesFollowRecords, err := r.countSalesFollowRecordsBetween(ctx, yesterdayStart, todayStart, scopedUserIDs, showAll)
		if err != nil {
			return model.DashboardOverview{}, err
		}
		currentMonthSalesCustomers, err := r.countSalesCustomersBetween(ctx, currentMonthStart, nextMonthStart, scopedUserIDs, showAll)
		if err != nil {
			return model.DashboardOverview{}, err
		}
		lastMonthSalesCustomers, err := r.countSalesCustomersBetween(ctx, lastMonthStart, currentMonthStart, scopedUserIDs, showAll)
		if err != nil {
			return model.DashboardOverview{}, err
		}
		currentMonthSalesFollowRecords, err := r.countSalesFollowRecordsBetween(ctx, currentMonthStart, nextMonthStart, scopedUserIDs, showAll)
		if err != nil {
			return model.DashboardOverview{}, err
		}
		lastMonthSalesFollowRecords, err := r.countSalesFollowRecordsBetween(ctx, lastMonthStart, currentMonthStart, scopedUserIDs, showAll)
		if err != nil {
			return model.DashboardOverview{}, err
		}
		todayNewCustomerRanks := []model.DashboardRankingItem{}
		todayFollowRecordRanks := []model.DashboardRankingItem{}
		if showAll || isDashboardScopedTeamRole(effectiveRole) {
			todayNewCustomerRanks, err = r.listSalesCustomerRanksBetween(ctx, todayStart, tomorrowStart, 10, scopedUserIDs, showAll)
			if err != nil {
				return model.DashboardOverview{}, err
			}
			todayFollowRecordRanks, err = r.listSalesFollowRecordRanksBetween(ctx, todayStart, tomorrowStart, 10, scopedUserIDs, showAll)
			if err != nil {
				return model.DashboardOverview{}, err
			}
		}

		salesAdminOverview = &model.DashboardSalesAdminOverview{
			TodayNewCustomers: model.DashboardStat{
				Current:    float64(todaySalesCustomers),
				Previous:   float64(yesterdaySalesCustomers),
				ChangeRate: calcChangeRate(float64(todaySalesCustomers), float64(yesterdaySalesCustomers)),
			},
			TodayFollowRecords: model.DashboardStat{
				Current:    float64(todaySalesFollowRecords),
				Previous:   float64(yesterdaySalesFollowRecords),
				ChangeRate: calcChangeRate(float64(todaySalesFollowRecords), float64(yesterdaySalesFollowRecords)),
			},
			MonthlyNewCustomers: model.DashboardStat{
				Current:    float64(currentMonthSalesCustomers),
				Previous:   float64(lastMonthSalesCustomers),
				ChangeRate: calcChangeRate(float64(currentMonthSalesCustomers), float64(lastMonthSalesCustomers)),
			},
			MonthlyFollowRecords: model.DashboardStat{
				Current:    float64(currentMonthSalesFollowRecords),
				Previous:   float64(lastMonthSalesFollowRecords),
				ChangeRate: calcChangeRate(float64(currentMonthSalesFollowRecords), float64(lastMonthSalesFollowRecords)),
			},
			TodayNewCustomerRanks:  todayNewCustomerRanks,
			TodayFollowRecordRanks: todayFollowRecordRanks,
		}
	}

	currentConversionRate := calcConversionRate(currentDoneCustomers, currentNewCustomers)
	lastConversionRate := calcConversionRate(lastDoneCustomers, lastNewCustomers)

	return model.DashboardOverview{
		Revenue: model.DashboardStat{
			Current:    currentRevenue,
			Previous:   lastRevenue,
			ChangeRate: calcChangeRate(currentRevenue, lastRevenue),
		},
		NewCustomers: model.DashboardStat{
			Current:    float64(currentNewCustomers),
			Previous:   float64(lastNewCustomers),
			ChangeRate: calcChangeRate(float64(currentNewCustomers), float64(lastNewCustomers)),
		},
		NewOpportunities: model.DashboardStat{
			Current:    float64(currentNewOpportunities),
			Previous:   float64(lastNewOpportunities),
			ChangeRate: calcChangeRate(float64(currentNewOpportunities), float64(lastNewOpportunities)),
		},
		ConversionRate: model.DashboardStat{
			Current:    currentConversionRate,
			Previous:   lastConversionRate,
			ChangeRate: calcChangeRate(currentConversionRate, lastConversionRate),
		},
		MonthlyRevenue:   monthlyRevenue,
		MonthlyContracts: monthlyContracts,
		AutoDropOverview: model.DashboardAutoDropOverview{
			FollowUpDueSoonCount:   followUpDueSoonCount,
			DealDueSoonCount:       dealDueSoonCount,
			MonthlyFollowUpDropped: monthlyFollowUpDropped,
			MonthlyDealDropped:     monthlyDealDropped,
		},
		SalesAdminOverview: salesAdminOverview,
		RecentDeals:        recentDeals,
		RecentActivities:   recentActivities,
	}, nil
}

func (r *gormDashboardRepository) sumContractAmountBetween(ctx context.Context, start, end time.Time, scopedUserIDs []int64, showAll bool) (float64, error) {
	type resultRow struct {
		Amount float64 `gorm:"column:amount"`
	}

	var row resultRow
	query := r.db.WithContext(ctx).
		Table("contracts").
		Select("COALESCE(SUM(contract_amount), 0) AS amount").
		Where("created_at >= ? AND created_at < ?", start, end)
	query = applyDashboardUserScope(query, "user_id", scopedUserIDs, showAll)
	err := query.Scan(&row).Error
	if err != nil {
		return 0, err
	}
	return row.Amount, nil
}

func (r *gormDashboardRepository) countCustomersBetween(ctx context.Context, start, end time.Time, dealStatus string, scopedUserIDs []int64, showAll bool) (int64, error) {
	query := r.db.WithContext(ctx).
		Table("customers").
		Where("(delete_time IS NULL OR delete_time = 0)").
		Where("created_at >= ? AND created_at < ?", start, end)
	query = applyDashboardUserScope(query, "owner_user_id", scopedUserIDs, showAll)
	if dealStatus != "" {
		query = query.Where("deal_status = ?", dealStatus)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *gormDashboardRepository) countFollowUpDueSoonCustomers(ctx context.Context, now time.Time, followUpDropDays int, scopedUserIDs []int64, showAll bool) (int64, error) {
	if followUpDropDays <= 0 {
		return 0, nil
	}

	nowUnix := now.Unix()
	warningDeadline := now.Add(24 * time.Hour).Unix()
	referenceExpr := fmt.Sprintf("NULLIF(c.next_time, 0) + %d", followUpDropDays*24*60*60)

	query := r.db.WithContext(ctx).
		Table("customers AS c").
		Where("(c.delete_time IS NULL OR c.delete_time = 0)").
		Where("c.owner_user_id IS NOT NULL").
		Where("c.status <> ?", model.CustomerStatusPool).
		Where("NOT EXISTS (SELECT 1 FROM contracts ct WHERE ct.customer_id = c.id)").
		Where("c.deal_status = ?", model.CustomerDealStatusUndone).
		Where("NULLIF(c.next_time, 0) IS NOT NULL").
		Where(referenceExpr+" > ?", nowUnix).
		Where(referenceExpr+" <= ?", warningDeadline)
	if !showAll {
		if len(scopedUserIDs) == 0 {
			return 0, nil
		}
		query = query.Where("c.owner_user_id IN ?", scopedUserIDs)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *gormDashboardRepository) countDealDueSoonCustomers(ctx context.Context, now time.Time, dealDropDays int, scopedUserIDs []int64, showAll bool) (int64, error) {
	if dealDropDays <= 0 {
		return 0, nil
	}

	nowUnix := now.Unix()
	warningDeadline := now.Add(10 * 24 * time.Hour).Unix()
	referenceExpr := fmt.Sprintf("NULLIF(c.collect_time, 0) + %d", dealDropDays*24*60*60)

	query := r.db.WithContext(ctx).
		Table("customers AS c").
		Where("(c.delete_time IS NULL OR c.delete_time = 0)").
		Where("c.owner_user_id IS NOT NULL").
		Where("c.status <> ?", model.CustomerStatusPool).
		Where("NOT EXISTS (SELECT 1 FROM contracts ct WHERE ct.customer_id = c.id)").
		Where("c.deal_status = ?", model.CustomerDealStatusUndone).
		Where("NULLIF(c.collect_time, 0) IS NOT NULL").
		Where(referenceExpr+" > ?", nowUnix).
		Where(referenceExpr+" <= ?", warningDeadline)
	if !showAll {
		if len(scopedUserIDs) == 0 {
			return 0, nil
		}
		query = query.Where("c.owner_user_id IN ?", scopedUserIDs)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *gormDashboardRepository) countAutoDropActivitiesBetween(ctx context.Context, start, end time.Time, action string, scopedUserIDs []int64, showAll bool) (int64, error) {
	query := r.db.WithContext(ctx).
		Table("activity_logs").
		Where("target_type = ?", model.TargetTypeCustomer).
		Where("action = ?", action).
		Where("created_at >= ? AND created_at < ?", start, end)
	if !showAll {
		if len(scopedUserIDs) == 0 {
			return 0, nil
		}
		query = query.Where("user_id IN ?", scopedUserIDs)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *gormDashboardRepository) resolveDashboardScopeUserIDs(ctx context.Context, actorUserID int64, actorRole string, showAll bool) ([]int64, error) {
	if showAll {
		return []int64{}, nil
	}
	if actorUserID <= 0 {
		return []int64{}, nil
	}

	roleName := strings.TrimSpace(actorRole)
	if roleName == "" {
		var err error
		roleName, err = r.getUserRoleName(ctx, actorUserID)
		if err != nil {
			return nil, err
		}
	}
	if isDashboardScopedTeamRole(roleName) {
		descendantIDs, err := r.listAllDescendantUserIDs(ctx, actorUserID)
		if err != nil {
			return nil, err
		}
		return uniqueDashboardPositiveInt64(append([]int64{actorUserID}, descendantIDs...)), nil
	}

	return []int64{actorUserID}, nil
}

func (r *gormDashboardRepository) resolveDashboardActorRole(ctx context.Context, actorUserID int64, actorRole string) (string, error) {
	roleName := strings.TrimSpace(actorRole)
	if actorUserID <= 0 {
		return roleName, nil
	}

	dbRoleName, err := r.getUserRoleName(ctx, actorUserID)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(dbRoleName) != "" {
		return dbRoleName, nil
	}
	return roleName, nil
}

func (r *gormDashboardRepository) getUserRoleName(ctx context.Context, userID int64) (string, error) {
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

func (r *gormDashboardRepository) listAllDescendantUserIDs(ctx context.Context, rootUserID int64) ([]int64, error) {
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

func isDashboardScopedTeamRole(role string) bool {
	switch strings.TrimSpace(strings.ToLower(role)) {
	case "sales_director", "销售总监", "sales_manager", "销售经理":
		return true
	default:
		return false
	}
}

func uniqueDashboardPositiveInt64(ids []int64) []int64 {
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

func (r *gormDashboardRepository) countContractsBetween(ctx context.Context, start, end time.Time, scopedUserIDs []int64, showAll bool) (int64, error) {
	query := r.db.WithContext(ctx).
		Table("contracts").
		Where("created_at >= ? AND created_at < ?", start, end)
	query = applyDashboardUserScope(query, "user_id", scopedUserIDs, showAll)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *gormDashboardRepository) getPositiveIntSetting(ctx context.Context, key string, defaultVal int) int {
	type settingRow struct {
		Value string `gorm:"column:value"`
	}

	var row settingRow
	if err := r.db.WithContext(ctx).
		Table("system_settings").
		Select("value").
		Where("`key` = ?", key).
		Take(&row).Error; err != nil {
		return defaultVal
	}

	value, err := strconv.Atoi(strings.TrimSpace(row.Value))
	if err != nil || value <= 0 {
		return defaultVal
	}
	return value
}

func applyDashboardSalesRoleScope(query *gorm.DB) *gorm.DB {
	return query.Where("(r.name IN ? OR r.label IN ?)", dashboardSalesRoleNames, dashboardSalesRoleLabels)
}

func applyDashboardUserScope(query *gorm.DB, column string, scopedUserIDs []int64, showAll bool) *gorm.DB {
	if showAll {
		return query
	}
	if len(scopedUserIDs) == 0 {
		return query.Where("1 = 0")
	}
	return query.Where(fmt.Sprintf("%s IN ?", column), scopedUserIDs)
}

func isDashboardGlobalRole(role string) bool {
	switch strings.TrimSpace(strings.ToLower(role)) {
	case "admin", "管理员", "finance", "finance_manager", "财务", "财务经理":
		return true
	default:
		return false
	}
}

func isDashboardSalesOverviewRole(role string) bool {
	switch strings.TrimSpace(strings.ToLower(role)) {
	case "admin", "管理员",
		"sales_director", "销售总监",
		"sales_manager", "销售经理",
		"sales_staff", "销售员工",
		"sales_inside", "sale_inside", "销售", "inside销售", "电销员工",
		"sales_outside", "sale_outside", "outside销售":
		return true
	default:
		return false
	}
}

func (r *gormDashboardRepository) countSalesCustomersBetween(ctx context.Context, start, end time.Time, scopedUserIDs []int64, showAll bool) (int64, error) {
	query := r.db.WithContext(ctx).
		Table("customers AS c").
		Joins("JOIN users AS u ON u.id = c.owner_user_id").
		Joins("JOIN roles AS r ON r.id = u.role_id").
		Where("(c.delete_time IS NULL OR c.delete_time = 0)").
		Where("c.created_at >= ? AND c.created_at < ?", start, end)
	query = applyDashboardSalesRoleScope(query)
	query = applyDashboardUserScope(query, "u.id", scopedUserIDs, showAll)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *gormDashboardRepository) countSalesFollowRecordsBetween(ctx context.Context, start, end time.Time, scopedUserIDs []int64, showAll bool) (int64, error) {
	query := r.db.WithContext(ctx).
		Table("sales_follow_records AS sfr").
		Joins("JOIN users AS u ON u.id = sfr.operator_user_id").
		Joins("JOIN roles AS r ON r.id = u.role_id").
		Where("sfr.created_at >= ? AND sfr.created_at < ?", start, end)
	query = applyDashboardSalesRoleScope(query)
	query = applyDashboardUserScope(query, "u.id", scopedUserIDs, showAll)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *gormDashboardRepository) listSalesCustomerRanksBetween(ctx context.Context, start, end time.Time, limit int, scopedUserIDs []int64, showAll bool) ([]model.DashboardRankingItem, error) {
	if limit <= 0 {
		return []model.DashboardRankingItem{}, nil
	}

	type rankRow struct {
		UserID   int64  `gorm:"column:user_id"`
		UserName string `gorm:"column:user_name"`
		Count    int64  `gorm:"column:item_count"`
	}

	var rows []rankRow
	query := r.db.WithContext(ctx).
		Table("customers AS c").
		Select("u.id AS user_id, COALESCE(NULLIF(u.nickname, ''), u.username, '') AS user_name, COUNT(*) AS item_count").
		Joins("JOIN users AS u ON u.id = c.owner_user_id").
		Joins("JOIN roles AS r ON r.id = u.role_id").
		Where("(c.delete_time IS NULL OR c.delete_time = 0)").
		Where("c.created_at >= ? AND c.created_at < ?", start, end).
		Group("u.id, u.nickname, u.username").
		Order("item_count DESC, u.id ASC").
		Limit(limit)
	query = applyDashboardSalesRoleScope(query)
	query = applyDashboardUserScope(query, "u.id", scopedUserIDs, showAll)
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]model.DashboardRankingItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.DashboardRankingItem{
			UserID:   row.UserID,
			UserName: row.UserName,
			Count:    row.Count,
		})
	}
	return items, nil
}

func (r *gormDashboardRepository) listSalesFollowRecordRanksBetween(ctx context.Context, start, end time.Time, limit int, scopedUserIDs []int64, showAll bool) ([]model.DashboardRankingItem, error) {
	if limit <= 0 {
		return []model.DashboardRankingItem{}, nil
	}

	type rankRow struct {
		UserID   int64  `gorm:"column:user_id"`
		UserName string `gorm:"column:user_name"`
		Count    int64  `gorm:"column:item_count"`
	}

	var rows []rankRow
	query := r.db.WithContext(ctx).
		Table("sales_follow_records AS sfr").
		Select("u.id AS user_id, COALESCE(NULLIF(u.nickname, ''), u.username, '') AS user_name, COUNT(*) AS item_count").
		Joins("JOIN users AS u ON u.id = sfr.operator_user_id").
		Joins("JOIN roles AS r ON r.id = u.role_id").
		Where("sfr.created_at >= ? AND sfr.created_at < ?", start, end).
		Group("u.id, u.nickname, u.username").
		Order("item_count DESC, u.id ASC").
		Limit(limit)
	query = applyDashboardSalesRoleScope(query)
	query = applyDashboardUserScope(query, "u.id", scopedUserIDs, showAll)
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]model.DashboardRankingItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.DashboardRankingItem{
			UserID:   row.UserID,
			UserName: row.UserName,
			Count:    row.Count,
		})
	}
	return items, nil
}

func (r *gormDashboardRepository) listMonthlyRevenue(ctx context.Context, currentMonthStart time.Time, months int, scopedUserIDs []int64, showAll bool) ([]model.DashboardMonthlyRevenue, error) {
	if months <= 0 {
		return []model.DashboardMonthlyRevenue{}, nil
	}

	items := make([]model.DashboardMonthlyRevenue, 0, months)
	for i := months - 1; i >= 0; i-- {
		start := currentMonthStart.AddDate(0, -i, 0)
		end := start.AddDate(0, 1, 0)
		amount, err := r.sumContractAmountBetween(ctx, start, end, scopedUserIDs, showAll)
		if err != nil {
			return nil, err
		}
		items = append(items, model.DashboardMonthlyRevenue{
			Label:  fmt.Sprintf("%d月", start.Month()),
			Amount: amount,
		})
	}
	return items, nil
}

func (r *gormDashboardRepository) listMonthlyContractCounts(ctx context.Context, currentMonthStart time.Time, months int, scopedUserIDs []int64, showAll bool) ([]model.DashboardMonthlyContractCount, error) {
	if months <= 0 {
		return []model.DashboardMonthlyContractCount{}, nil
	}

	items := make([]model.DashboardMonthlyContractCount, 0, months)
	for i := months - 1; i >= 0; i-- {
		start := currentMonthStart.AddDate(0, -i, 0)
		end := start.AddDate(0, 1, 0)
		total, err := r.countContractsBetween(ctx, start, end, scopedUserIDs, showAll)
		if err != nil {
			return nil, err
		}
		items = append(items, model.DashboardMonthlyContractCount{
			Label: fmt.Sprintf("%d月", start.Month()),
			Count: total,
		})
	}
	return items, nil
}

func (r *gormDashboardRepository) listRecentDeals(ctx context.Context, limit int, scopedUserIDs []int64, showAll bool) ([]model.DashboardRecentDeal, error) {
	if limit <= 0 {
		return []model.DashboardRecentDeal{}, nil
	}

	type dealRow struct {
		ID            int64     `gorm:"column:id"`
		UserName      string    `gorm:"column:user_name"`
		CustomerName  string    `gorm:"column:customer_name"`
		CustomerEmail string    `gorm:"column:customer_email"`
		ContractName  string    `gorm:"column:contract_name"`
		Amount        float64   `gorm:"column:amount"`
		CreatedAt     time.Time `gorm:"column:created_at"`
	}

	var rows []dealRow
	query := r.db.WithContext(ctx).
		Table("contracts AS c").
		Select("c.id, c.contract_name, c.contract_amount AS amount, c.created_at, COALESCE(cu.name, '') AS customer_name, COALESCE(cu.email, '') AS customer_email, COALESCE(NULLIF(u.nickname, ''), u.username) AS user_name").
		Joins("JOIN users AS u ON u.id = c.user_id").
		Joins("LEFT JOIN customers AS cu ON cu.id = c.customer_id").
		Order("c.created_at DESC, c.id DESC").
		Limit(limit)
	query = applyDashboardUserScope(query, "c.user_id", scopedUserIDs, showAll)
	err := query.Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	items := make([]model.DashboardRecentDeal, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.DashboardRecentDeal{
			ID:            row.ID,
			UserName:      row.UserName,
			CustomerName:  row.CustomerName,
			CustomerEmail: row.CustomerEmail,
			ContractName:  row.ContractName,
			Amount:        row.Amount,
			CreatedAt:     row.CreatedAt,
		})
	}
	return items, nil
}

func (r *gormDashboardRepository) listRecentActivities(ctx context.Context, limit int, scopedUserIDs []int64, showAll bool) ([]model.DashboardRecentActivity, error) {
	if limit <= 0 {
		return []model.DashboardRecentActivity{}, nil
	}

	type row struct {
		ID         int64     `gorm:"column:id"`
		Action     string    `gorm:"column:action"`
		TargetType string    `gorm:"column:target_type"`
		TargetName string    `gorm:"column:target_name"`
		Content    string    `gorm:"column:content"`
		UserName   string    `gorm:"column:user_name"`
		CreatedAt  time.Time `gorm:"column:created_at"`
	}

	var rows []row
	query := r.db.WithContext(ctx).
		Table("activity_logs AS a").
		Select("a.id, a.action, a.target_type, a.target_name, a.content, a.created_at, COALESCE(NULLIF(u.nickname, ''), u.username, '') AS user_name").
		Joins("LEFT JOIN users AS u ON u.id = a.user_id").
		Order("a.created_at DESC, a.id DESC").
		Limit(limit)
	query = applyDashboardUserScope(query, "a.user_id", scopedUserIDs, showAll)
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]model.DashboardRecentActivity, 0, len(rows))
	for _, r := range rows {
		items = append(items, model.DashboardRecentActivity{
			ID:        r.ID,
			Type:      "activity",
			UserName:  r.UserName,
			Action:    actionLabel(r.Action),
			Target:    r.TargetName,
			Content:   r.Content,
			CreatedAt: r.CreatedAt,
		})
	}
	return items, nil
}

var actionLabelMap = map[string]string{
	model.ActionCreateContract:   "创建合同",
	model.ActionAuditContract:    "审核合同",
	model.ActionCreateCustomer:   "创建客户",
	model.ActionImportCustomer:   "导入客户",
	model.ActionClaimCustomer:    "领取客户",
	model.ActionReleaseCustomer:  "丢弃客户",
	model.ActionAutoDropFollowUp: "未跟进掉库通知",
	model.ActionAutoDropDeal:     "未签单掉库通知",
	model.ActionTransferCustomer: "转移客户",
	model.ActionSalesFollow:      "销售跟进",
	model.ActionOperationFollow:  "运营跟进",
}

func actionLabel(action string) string {
	if label, ok := actionLabelMap[action]; ok {
		return label
	}
	return action
}

func monthStart(now time.Time, location *time.Location) time.Time {
	if location == nil {
		location = time.Local
	}
	year, month, _ := now.In(location).Date()
	return time.Date(year, month, 1, 0, 0, 0, 0, location)
}

func dayStart(now time.Time, location *time.Location) time.Time {
	if location == nil {
		location = time.Local
	}
	year, month, day := now.In(location).Date()
	return time.Date(year, month, day, 0, 0, 0, 0, location)
}

func calcConversionRate(doneCount, totalCount int64) float64 {
	if totalCount <= 0 {
		return 0
	}
	return (float64(doneCount) / float64(totalCount)) * 100
}

func calcChangeRate(current, previous float64) float64 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}
	return ((current - previous) / previous) * 100
}
