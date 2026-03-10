package repository

import (
	"backend/internal/model"
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type NotificationRepository interface {
	ListReadKeys(ctx context.Context, userID int64) ([]string, error)
	MarkAsRead(ctx context.Context, userID int64, keys []string) error
}

type gormNotificationRepository struct {
	db *gorm.DB
}

func NewGormNotificationRepository(db *gorm.DB) NotificationRepository {
	return &gormNotificationRepository{db: db}
}

func (r *gormNotificationRepository) ListReadKeys(ctx context.Context, userID int64) ([]string, error) {
	var keys []string
	err := r.db.WithContext(ctx).
		Model(&model.NotificationRead{}).
		Where("user_id = ?", userID).
		Pluck("notification_key", &keys).Error
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (r *gormNotificationRepository) MarkAsRead(ctx context.Context, userID int64, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	records := make([]model.NotificationRead, 0, len(keys))
	for _, key := range keys {
		records = append(records, model.NotificationRead{
			UserID:          userID,
			NotificationKey: key,
		})
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&records).Error
}
