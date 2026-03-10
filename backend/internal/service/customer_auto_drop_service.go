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
	maxFailItems int
}

type autoDropCandidateRow struct {
	ID          int64     `gorm:"column:id"`
	OwnerUserID int64     `gorm:"column:owner_user_id"`
	FollowTime  *int64    `gorm:"column:follow_time"`
	CollectTime *int64    `gorm:"column:collect_time"`
	DealTime    *int64    `gorm:"column:deal_time"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

type ownerLogRow struct {
	CustomerID      int64     `gorm:"column:customer_id"`
	FromOwnerUserID *int64    `gorm:"column:from_owner_user_id"`
	ToOwnerUserID   *int64    `gorm:"column:to_owner_user_id"`
	Action          string    `gorm:"column:action"`
	OperatorUserID  int64     `gorm:"column:operator_user_id"`
	CreatedAt       time.Time `gorm:"column:created_at"`
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

func NewCustomerAutoDropService(
	db *gorm.DB,
	settingRepo *repository.SystemSettingRepository,
) CustomerAutoDropService {
	return &customerAutoDropService{
		db:           db,
		settingRepo:  settingRepo,
		maxFailItems: 20,
	}
}

func (s *customerAutoDropService) Run(ctx context.Context) (CustomerAutoDropTaskResult, error) {
	now := time.Now().UTC()
	result := CustomerAutoDropTaskResult{
		ExecutedAt:         now,
		FollowUpDropDays:   s.getIntSetting("follow_up_drop_days", defaultFollowUpDropDays),
		DealDropDays:       s.getIntSetting("deal_drop_days", defaultDealDropDays),
		HolidayModeEnabled: s.getBoolSetting("holiday_mode_enabled", false),
		Failures:           make([]CustomerAutoDropTaskFailure, 0),
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

	followCutoffUnix := now.AddDate(0, 0, -result.FollowUpDropDays).Unix()
	dealCutoffUnix := now.AddDate(0, 0, -result.DealDropDays).Unix()

	for _, row := range candidates {
		followReference := firstNonZero(row.FollowTime, row.CollectTime, row.CreatedAt.Unix())
		dealReference := firstNonZero(row.DealTime, row.CollectTime, row.CreatedAt.Unix())

		followTimeout := result.FollowUpDropDays > 0 && followReference <= followCutoffUnix
		dealTimeout := result.DealDropDays > 0 && dealReference <= dealCutoffUnix
		if !followTimeout && !dealTimeout {
			continue
		}

		triggerType, reason := autoDropReason(result.FollowUpDropDays, result.DealDropDays, followTimeout, dealTimeout)
		dropped, dropErr := s.dropOne(ctx, row, now, triggerType, reason)
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
		Select("id, owner_user_id, NULLIF(follow_time, 0) AS follow_time, NULLIF(collect_time, 0) AS collect_time, NULLIF(deal_time, 0) AS deal_time, created_at").
		Where("owner_user_id IS NOT NULL").
		Where("status <> ?", model.CustomerStatusPool).
		Where("deal_status = ?", model.CustomerDealStatusUndone).
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
) (bool, error) {
	dropped := false
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		update := tx.Table("customers").
			Where("id = ?", row.ID).
			Where("owner_user_id = ?", row.OwnerUserID).
			Where("status <> ?", model.CustomerStatusPool).
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
		if err := tx.Table("customer_owner_logs").Create(&ownerLogRow{
			CustomerID:      row.ID,
			FromOwnerUserID: &fromOwner,
			ToOwnerUserID:   nil,
			Action:          "release",
			OperatorUserID:  row.OwnerUserID,
			CreatedAt:       now,
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

