package model

import "time"

type DailyUserCallStat struct {
	ID                  int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	StatDate            string    `json:"statDate" gorm:"column:stat_date"`
	UserID              *int64    `json:"userId,omitempty" gorm:"column:user_id"`
	RealName            string    `json:"realName" gorm:"column:real_name"`
	Mobile              string    `json:"mobile" gorm:"column:mobile"`
	BindNum             int       `json:"bindNum" gorm:"column:bind_num"`
	CallNum             int       `json:"callNum" gorm:"column:call_num"`
	NotConnected        int       `json:"notConnected" gorm:"column:not_connected"`
	ConnectionRate      float64   `json:"connectionRate" gorm:"column:connection_rate"`
	TimeTotal           int       `json:"timeTotal" gorm:"column:time_total"`
	TotalMinute         string    `json:"totalMinute" gorm:"column:total_minute"`
	TotalSecond         int       `json:"totalSecond" gorm:"column:total_second"`
	AverageCallDuration float64   `json:"averageCallDuration" gorm:"column:average_call_duration"`
	AverageCallSecond   float64   `json:"averageCallSecond" gorm:"column:average_call_second"`
	CreatedAt           time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt           time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

type DailyUserCallStatUpsertInput struct {
	StatDate            string
	UserID              *int64
	RealName            string
	Mobile              string
	BindNum             int
	CallNum             int
	NotConnected        int
	ConnectionRate      float64
	TimeTotal           int
	TotalMinute         string
	TotalSecond         int
	AverageCallDuration float64
	AverageCallSecond   float64
}

func (DailyUserCallStat) TableName() string { return "daily_user_call_stats" }
