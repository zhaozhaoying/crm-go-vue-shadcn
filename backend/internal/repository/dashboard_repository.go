package repository

import (
	"backend/internal/model"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type DashboardRepository interface {
	GetOverview(ctx context.Context, now time.Time, actorUserID int64, showAll bool) (model.DashboardOverview, error)
}

type gormDashboardRepository struct {
	db *gorm.DB
}

func NewGormDashboardRepository(db *gorm.DB) DashboardRepository {
	return &gormDashboardRepository{db: db}
}

func (r *gormDashboardRepository) GetOverview(ctx context.Context, now time.Time, actorUserID int64, showAll bool) (model.DashboardOverview, error) {
	location := now.Location()
	currentMonthStart := monthStart(now, location)
	nextMonthStart := currentMonthStart.AddDate(0, 1, 0)
	lastMonthStart := currentMonthStart.AddDate(0, -1, 0)

	currentRevenue, err := r.sumContractAmountBetween(ctx, currentMonthStart, nextMonthStart, actorUserID, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
	}
	lastRevenue, err := r.sumContractAmountBetween(ctx, lastMonthStart, currentMonthStart, actorUserID, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
	}

	currentNewCustomers, err := r.countCustomersBetween(ctx, currentMonthStart, nextMonthStart, "", actorUserID, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
	}
	lastNewCustomers, err := r.countCustomersBetween(ctx, lastMonthStart, currentMonthStart, "", actorUserID, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
	}

	currentNewOpportunities, err := r.countCustomersBetween(ctx, currentMonthStart, nextMonthStart, model.CustomerDealStatusUndone, actorUserID, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
	}
	lastNewOpportunities, err := r.countCustomersBetween(ctx, lastMonthStart, currentMonthStart, model.CustomerDealStatusUndone, actorUserID, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
	}

	currentDoneCustomers, err := r.countCustomersBetween(ctx, currentMonthStart, nextMonthStart, model.CustomerDealStatusDone, actorUserID, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
	}
	lastDoneCustomers, err := r.countCustomersBetween(ctx, lastMonthStart, currentMonthStart, model.CustomerDealStatusDone, actorUserID, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
	}

	monthlyRevenue, err := r.listMonthlyRevenue(ctx, currentMonthStart, 12, actorUserID, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
	}

	monthlyContracts, err := r.listMonthlyContractCounts(ctx, currentMonthStart, 12, actorUserID, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
	}

	recentDeals, err := r.listRecentDeals(ctx, 5, actorUserID, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
	}

	recentActivities, err := r.listRecentActivities(ctx, 5, actorUserID, showAll)
	if err != nil {
		return model.DashboardOverview{}, err
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
		RecentDeals:      recentDeals,
		RecentActivities: recentActivities,
	}, nil
}

func (r *gormDashboardRepository) sumContractAmountBetween(ctx context.Context, start, end time.Time, actorUserID int64, showAll bool) (float64, error) {
	type resultRow struct {
		Amount float64 `gorm:"column:amount"`
	}

	var row resultRow
	query := r.db.WithContext(ctx).
		Table("contracts").
		Select("COALESCE(SUM(contract_amount), 0) AS amount").
		Where("created_at >= ? AND created_at < ?", start, end)
	if !showAll {
		query = query.Where("user_id = ?", actorUserID)
	}
	err := query.Scan(&row).Error
	if err != nil {
		return 0, err
	}
	return row.Amount, nil
}

func (r *gormDashboardRepository) countCustomersBetween(ctx context.Context, start, end time.Time, dealStatus string, actorUserID int64, showAll bool) (int64, error) {
	query := r.db.WithContext(ctx).
		Table("customers").
		Where("(delete_time IS NULL OR delete_time = 0)").
		Where("created_at >= ? AND created_at < ?", start, end)
	if !showAll {
		query = query.Where("owner_user_id = ?", actorUserID)
	}
	if dealStatus != "" {
		query = query.Where("deal_status = ?", dealStatus)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *gormDashboardRepository) countContractsBetween(ctx context.Context, start, end time.Time, actorUserID int64, showAll bool) (int64, error) {
	query := r.db.WithContext(ctx).
		Table("contracts").
		Where("created_at >= ? AND created_at < ?", start, end)
	if !showAll {
		query = query.Where("user_id = ?", actorUserID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *gormDashboardRepository) listMonthlyRevenue(ctx context.Context, currentMonthStart time.Time, months int, actorUserID int64, showAll bool) ([]model.DashboardMonthlyRevenue, error) {
	if months <= 0 {
		return []model.DashboardMonthlyRevenue{}, nil
	}

	items := make([]model.DashboardMonthlyRevenue, 0, months)
	for i := months - 1; i >= 0; i-- {
		start := currentMonthStart.AddDate(0, -i, 0)
		end := start.AddDate(0, 1, 0)
		amount, err := r.sumContractAmountBetween(ctx, start, end, actorUserID, showAll)
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

func (r *gormDashboardRepository) listMonthlyContractCounts(ctx context.Context, currentMonthStart time.Time, months int, actorUserID int64, showAll bool) ([]model.DashboardMonthlyContractCount, error) {
	if months <= 0 {
		return []model.DashboardMonthlyContractCount{}, nil
	}

	items := make([]model.DashboardMonthlyContractCount, 0, months)
	for i := months - 1; i >= 0; i-- {
		start := currentMonthStart.AddDate(0, -i, 0)
		end := start.AddDate(0, 1, 0)
		total, err := r.countContractsBetween(ctx, start, end, actorUserID, showAll)
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

func (r *gormDashboardRepository) listRecentDeals(ctx context.Context, limit int, actorUserID int64, showAll bool) ([]model.DashboardRecentDeal, error) {
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
	if !showAll {
		query = query.Where("c.user_id = ?", actorUserID)
	}
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

func (r *gormDashboardRepository) listRecentActivities(ctx context.Context, limit int, actorUserID int64, showAll bool) ([]model.DashboardRecentActivity, error) {
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
	if !showAll {
		query = query.Where("a.user_id = ?", actorUserID)
	}
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
