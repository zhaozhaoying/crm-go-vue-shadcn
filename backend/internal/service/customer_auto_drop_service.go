package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	defaultFollowUpDropDays = 30
	defaultDealDropDays     = 90
)

type CustomerAutoDropTaskFailure struct {
	CustomerID int64  `json:"customerId"`
	Error      string `json:"error"`
}

type CustomerAutoDropTaskResult struct {
	ExecutedAt              time.Time                     `json:"executedAt"`
	AutoDropEnabled         bool                          `json:"autoDropEnabled"`
	FollowUpDropDays        int                           `json:"followUpDropDays"`
	DealDropDays            int                           `json:"dealDropDays"`
	HolidayModeEnabled      bool                          `json:"holidayModeEnabled"`
	Skipped                 bool                          `json:"skipped"`
	SkipReason              string                        `json:"skipReason,omitempty"`
	Evaluated               int                           `json:"evaluated"`
	Dropped                 int                           `json:"dropped"`
	FollowUpTimeoutDropped  int                           `json:"followUpTimeoutDropped"`
	DealTimeoutDropped      int                           `json:"dealTimeoutDropped"`
	BothRulesMatchedDropped int                           `json:"bothRulesMatchedDropped"`
	Failures                []CustomerAutoDropTaskFailure `json:"failures"`
}

type CustomerAutoDropService interface {
	Run(ctx context.Context) (CustomerAutoDropTaskResult, error)
}

type customerAutoDropService struct {
	db           *gorm.DB
	settingRepo  *repository.SystemSettingRepository
	activityRepo *repository.ActivityLogRepository
	maxFailItems int
}

type autoDropCandidateRow struct {
	ID            int64     `gorm:"column:id"`
	Name          string    `gorm:"column:name"`
	OwnerUserID   int64     `gorm:"column:owner_user_id"`
	OwnerUserName string    `gorm:"column:owner_user_name"`
	NextTime      *int64    `gorm:"column:next_time"`
	CollectTime   *int64    `gorm:"column:collect_time"`
	CreatedAt     time.Time `gorm:"column:created_at"`
}

type ownerLogRow struct {
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

type statusLogRow struct {
	CustomerID     int64  `gorm:"column:customer_id"`
	FromStatus     int    `gorm:"column:from_status"`
	ToStatus       int    `gorm:"column:to_status"`
	TriggerType    int    `gorm:"column:trigger_type"`
	Reason         string `gorm:"column:reason"`
	OperatorUserID int64  `gorm:"column:operator_user_id"`
	OperateTime    int64  `gorm:"column:operate_time"`
}

type activityLogRow struct {
	UserID     int64     `gorm:"column:user_id"`
	Action     string    `gorm:"column:action"`
	TargetType string    `gorm:"column:target_type"`
	TargetID   int64     `gorm:"column:target_id"`
	TargetName string    `gorm:"column:target_name"`
	Content    string    `gorm:"column:content"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func NewCustomerAutoDropService(
	db *gorm.DB,
	settingRepo *repository.SystemSettingRepository,
	activityRepo ...*repository.ActivityLogRepository,
) CustomerAutoDropService {
	svc := &customerAutoDropService{
		db:           db,
		settingRepo:  settingRepo,
		maxFailItems: 20,
	}
	if len(activityRepo) > 0 {
		svc.activityRepo = activityRepo[0]
	}
	return svc
}

func (s *customerAutoDropService) Run(ctx context.Context) (CustomerAutoDropTaskResult, error) {
	now := time.Now().UTC()
	result := CustomerAutoDropTaskResult{
		ExecutedAt:         now,
		AutoDropEnabled:    s.getBoolSetting("customer_auto_drop_enabled", true),
		FollowUpDropDays:   s.getIntSetting("follow_up_drop_days", defaultFollowUpDropDays),
		DealDropDays:       s.getIntSetting("deal_drop_days", defaultDealDropDays),
		HolidayModeEnabled: s.getBoolSetting("holiday_mode_enabled", false),
		Failures:           make([]CustomerAutoDropTaskFailure, 0),
	}
	claimFreezeDays := s.getIntSetting("claim_freeze_days", defaultClaimFreezeDays)

	if !result.AutoDropEnabled {
		result.Skipped = true
		result.SkipReason = "auto drop disabled"
		return result, nil
	}
	if result.HolidayModeEnabled {
		result.Skipped = true
		result.SkipReason = "holiday mode enabled"
		return result, nil
	}
	if result.FollowUpDropDays <= 0 && result.DealDropDays <= 0 {
		result.Skipped = true
		result.SkipReason = "drop rules disabled"
		return result, nil
	}

	candidates, err := s.listCandidates(ctx)
	if err != nil {
		return result, err
	}
	result.Evaluated = len(candidates)

	for _, row := range candidates {
		followTimeout := result.FollowUpDropDays > 0 &&
			row.NextTime != nil &&
			*row.NextTime+int64(result.FollowUpDropDays)*24*60*60 <= now.Unix()
		dealTimeout := result.DealDropDays > 0 &&
			row.CollectTime != nil &&
			*row.CollectTime+int64(result.DealDropDays)*24*60*60 <= now.Unix()
		if !followTimeout && !dealTimeout {
			continue
		}

		triggerType, reason := autoDropReason(result.FollowUpDropDays, result.DealDropDays, followTimeout, dealTimeout)
		dropped, dropErr := s.dropOne(
			ctx,
			row,
			now,
			triggerType,
			reason,
			claimFreezeDays,
			result.FollowUpDropDays,
			result.DealDropDays,
			followTimeout,
			dealTimeout,
		)
		if dropErr != nil {
			if len(result.Failures) < s.maxFailItems {
				result.Failures = append(result.Failures, CustomerAutoDropTaskFailure{
					CustomerID: row.ID,
					Error:      dropErr.Error(),
				})
			}
			continue
		}
		if !dropped {
			continue
		}

		result.Dropped++
		switch {
		case followTimeout && dealTimeout:
			result.BothRulesMatchedDropped++
			result.FollowUpTimeoutDropped++
			result.DealTimeoutDropped++
		case followTimeout:
			result.FollowUpTimeoutDropped++
		case dealTimeout:
			result.DealTimeoutDropped++
		}
	}

	return result, nil
}

func (s *customerAutoDropService) listCandidates(ctx context.Context) ([]autoDropCandidateRow, error) {
	rows := make([]autoDropCandidateRow, 0)
	err := s.db.WithContext(ctx).
		Table("customers").
		Select(`
			customers.id,
			customers.name,
			customers.owner_user_id,
			COALESCE(NULLIF(u.nickname, ''), NULLIF(u.username, ''), '') AS owner_user_name,
			NULLIF(customers.next_time, 0) AS next_time,
			NULLIF(customers.collect_time, 0) AS collect_time,
			customers.created_at
		`).
		Joins("LEFT JOIN users AS u ON u.id = customers.owner_user_id").
		Where("customers.owner_user_id IS NOT NULL").
		Where("customers.status <> ?", model.CustomerStatusPool).
		Where("NOT EXISTS (SELECT 1 FROM contracts ct WHERE ct.customer_id = customers.id)").
		Where("customers.deal_status = ?", model.CustomerDealStatusUndone).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *customerAutoDropService) dropOne(
	ctx context.Context,
	row autoDropCandidateRow,
	now time.Time,
	triggerType int,
	reason string,
	claimFreezeDays int,
	followUpDropDays int,
	dealDropDays int,
	followTimeout bool,
	dealTimeout bool,
) (bool, error) {
	dropped := false
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		update := tx.Table("customers").
			Where("id = ?", row.ID).
			Where("owner_user_id = ?", row.OwnerUserID).
			Where("status <> ?", model.CustomerStatusPool).
			Where("NOT EXISTS (SELECT 1 FROM contracts ct WHERE ct.customer_id = customers.id)").
			Where("deal_status = ?", model.CustomerDealStatusUndone).
			Updates(map[string]interface{}{
				"owner_user_id": nil,
				"status":        model.CustomerStatusPool,
				"drop_time":     now.Unix(),
				"drop_user_id":  row.OwnerUserID,
				"updated_at":    now,
			})
		if update.Error != nil {
			return update.Error
		}
		if update.RowsAffected == 0 {
			return nil
		}
		dropped = true

		fromOwner := row.OwnerUserID
		blockedDepartmentAnchorUserID, blockedUntil, err := buildClaimBlockInfo(ctx, repository.NewGormCustomerRepository(tx), row.OwnerUserID, claimFreezeDays, now)
		if err != nil {
			return err
		}
		if err := tx.Table("customer_owner_logs").Create(&ownerLogRow{
			CustomerID:                    row.ID,
			FromOwnerUserID:               &fromOwner,
			ToOwnerUserID:                 nil,
			Action:                        "release",
			Reason:                        model.CustomerOwnerLogReasonAutoDrop,
			Content:                       reason,
			BlockedDepartmentAnchorUserID: blockedDepartmentAnchorUserID,
			BlockedUntil:                  blockedUntil,
			OperatorUserID:                row.OwnerUserID,
			CreatedAt:                     now,
		}).Error; err != nil {
			return err
		}

		if err := tx.Table("customer_status_logs").Create(&statusLogRow{
			CustomerID:     row.ID,
			FromStatus:     0,
			ToStatus:       0,
			TriggerType:    triggerType,
			Reason:         reason,
			OperatorUserID: row.OwnerUserID,
			OperateTime:    now.Unix(),
		}).Error; err != nil {
			return err
		}

		logs := s.buildAutoDropActivityLogs(row, now, followUpDropDays, dealDropDays, followTimeout, dealTimeout)
		if len(logs) > 0 {
			if err := tx.Table("activity_logs").Create(&logs).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return dropped, err
}

func (s *customerAutoDropService) getIntSetting(key string, defaultVal int) int {
	if s.settingRepo == nil {
		return defaultVal
	}
	setting, err := s.settingRepo.GetSetting(key)
	if err != nil || setting == nil {
		return defaultVal
	}
	value, err := strconv.Atoi(strings.TrimSpace(setting.Value))
	if err != nil || value <= 0 {
		return defaultVal
	}
	return value
}

func (s *customerAutoDropService) getBoolSetting(key string, defaultVal bool) bool {
	if s.settingRepo == nil {
		return defaultVal
	}
	setting, err := s.settingRepo.GetSetting(key)
	if err != nil || setting == nil {
		return defaultVal
	}
	switch strings.ToLower(strings.TrimSpace(setting.Value)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return defaultVal
	}
}

func firstNonZero(primary, secondary *int64, fallback int64) int64 {
	if primary != nil && *primary > 0 {
		return *primary
	}
	if secondary != nil && *secondary > 0 {
		return *secondary
	}
	return fallback
}

func buildClaimBlockInfo(
	ctx context.Context,
	repo customerOwnerAssignmentRepo,
	ownerUserID int64,
	claimFreezeDays int,
	now time.Time,
) (*int64, *time.Time, error) {
	if ownerUserID <= 0 || claimFreezeDays <= 0 {
		return nil, nil, nil
	}

	anchorUserID, err := resolveSalesDirectorUserID(ctx, repo, ownerUserID)
	if err != nil {
		return nil, nil, err
	}
	if anchorUserID <= 0 {
		return nil, nil, nil
	}

	blockedUntil := now.Add(time.Duration(claimFreezeDays) * 24 * time.Hour)
	return &anchorUserID, &blockedUntil, nil
}

func autoDropReason(followDays, dealDays int, followTimeout, dealTimeout bool) (int, string) {
	switch {
	case followTimeout && dealTimeout:
		return 4, fmt.Sprintf("系统自动掉库：超过%d天未跟进且超过%d天未签单", followDays, dealDays)
	case followTimeout:
		return 3, fmt.Sprintf("系统自动掉库：超过%d天未跟进", followDays)
	case dealTimeout:
		return 4, fmt.Sprintf("系统自动掉库：超过%d天未签单", dealDays)
	default:
		return 3, "系统自动掉库"
	}
}

func (s *customerAutoDropService) buildAutoDropActivityLogs(
	row autoDropCandidateRow,
	now time.Time,
	followUpDropDays int,
	dealDropDays int,
	followTimeout bool,
	dealTimeout bool,
) []activityLogRow {
	if s.activityRepo == nil {
		return nil
	}

	logs := make([]activityLogRow, 0, 2)
	customerName := strings.TrimSpace(row.Name)
	if customerName == "" {
		customerName = fmt.Sprintf("客户%d", row.ID)
	}
	ownerName := strings.TrimSpace(row.OwnerUserName)
	if ownerName == "" {
		ownerName = fmt.Sprintf("用户%d", row.OwnerUserID)
	}
	timeText := now.In(time.Local).Format("2006-01-02 15:04:05")

	if followTimeout {
		logs = append(logs, activityLogRow{
			UserID:     row.OwnerUserID,
			Action:     model.ActionAutoDropFollowUp,
			TargetType: model.TargetTypeCustomer,
			TargetID:   row.ID,
			TargetName: customerName,
			Content: fmt.Sprintf(
				"客户【%s】因销售【%s】%d天未跟进，系统于%s自动触发掉库。",
				customerName,
				ownerName,
				followUpDropDays,
				timeText,
			),
			CreatedAt: now,
		})
	}
	if dealTimeout {
		logs = append(logs, activityLogRow{
			UserID:     row.OwnerUserID,
			Action:     model.ActionAutoDropDeal,
			TargetType: model.TargetTypeCustomer,
			TargetID:   row.ID,
			TargetName: customerName,
			Content: fmt.Sprintf(
				"客户【%s】因销售【%s】%d天未签单，系统于%s自动触发掉库。",
				customerName,
				ownerName,
				dealDropDays,
				timeText,
			),
			CreatedAt: now,
		})
	}

	return logs
}
