package repository

import (
	"backend/internal/model"
	"context"
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
		Order("created_at DESC, id DESC").
		Limit(limit)
	if !showAll {
		query = query.Where("user_id = ?", userID)
	}
	if err := query.Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}
