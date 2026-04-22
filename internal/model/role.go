package model

import "time"

type Role struct {
	ID        int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"column:name;size:64;not null;uniqueIndex"`
	Label     string    `json:"label" gorm:"column:label;size:64;not null;default:''"`
	Sort      int       `json:"sort" gorm:"column:sort;not null;default:0"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
}

func (Role) TableName() string { return "roles" }
