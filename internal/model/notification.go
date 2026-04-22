package model

import "time"

type NotificationRead struct {
	ID              int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	UserID          int64     `json:"userId" gorm:"column:user_id;not null"`
	NotificationKey string    `json:"notificationKey" gorm:"column:notification_key;not null"`
	CreatedAt       time.Time `json:"createdAt" gorm:"column:created_at"`
}

func (NotificationRead) TableName() string {
	return "notification_reads"
}
