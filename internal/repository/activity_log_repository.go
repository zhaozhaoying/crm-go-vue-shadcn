package repository

import (
	"backend/internal/model"
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ActivityLogRepository struct {
	db *gorm.DB
}

func NewActivityLogRepository(db *gorm.DB) *ActivityLogRepository {
	return &ActivityLogRepository{db: db}
}

func (r *ActivityLogRepository) Create(ctx context.Context, log model.ActivityLog) error {
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	return r.db.WithContext(ctx).Create(&log).Error
}

func (r *ActivityLogRepository) ListRecent(ctx context.Context, limit int, userID int64, showAll bool) ([]model.ActivityLog, error) {
	var logs []model.ActivityLog
	query := r.db.WithContext(ctx).
		Table("activity_logs AS a").
		Select(`
			a.id,
			a.user_id,
			a.action,
			a.target_type,
			a.target_id,
			COALESCE(NULLIF(a.target_name, ''), NULLIF(c.name, ''), NULLIF(ct.contract_name, ''), '') AS target_name,
			a.content,
			a.created_at,
			COALESCE(NULLIF(u.nickname, ''), NULLIF(u.username, ''), '') AS user_name
		`).
		Joins("LEFT JOIN users AS u ON u.id = a.user_id").
		Joins("LEFT JOIN customers AS c ON a.target_type = ? AND c.id = a.target_id", model.TargetTypeCustomer).
		Joins("LEFT JOIN contracts AS ct ON a.target_type = ? AND ct.id = a.target_id", model.TargetTypeContract).
		Order("a.created_at DESC, a.id DESC").
		Limit(limit)
	if !showAll {
		query = query.Where("a.user_id = ?", userID)
	}
	if err := query.Scan(&logs).Error; err != nil {
		return nil, err
	}
	if err := r.fillMissingUserNames(ctx, logs); err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *ActivityLogRepository) fillMissingUserNames(ctx context.Context, logs []model.ActivityLog) error {
	if len(logs) == 0 {
		return nil
	}

	missingUserIDs := make([]int64, 0)
	seen := make(map[int64]struct{})
	for _, log := range logs {
		if log.UserID <= 0 || strings.TrimSpace(log.UserName) != "" {
			continue
		}
		if _, ok := seen[log.UserID]; ok {
			continue
		}
		seen[log.UserID] = struct{}{}
		missingUserIDs = append(missingUserIDs, log.UserID)
	}

	if len(missingUserIDs) == 0 {
		return nil
	}

	type userRow struct {
		ID       int64  `gorm:"column:id"`
		UserName string `gorm:"column:user_name"`
	}

	var rows []userRow
	if err := r.db.WithContext(ctx).
		Table("users").
		Select("id, COALESCE(NULLIF(nickname, ''), NULLIF(username, ''), '') AS user_name").
		Where("id IN ?", missingUserIDs).
		Scan(&rows).Error; err != nil {
		return err
	}

	nameMap := make(map[int64]string, len(rows))
	for _, row := range rows {
		nameMap[row.ID] = strings.TrimSpace(row.UserName)
	}

	for i := range logs {
		if strings.TrimSpace(logs[i].UserName) != "" || logs[i].UserID <= 0 {
			continue
		}
		if name, ok := nameMap[logs[i].UserID]; ok && name != "" {
			logs[i].UserName = name
		}
	}

	return nil
}
