package model

import "time"

type SystemSetting struct {
	ID          int       `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Key         string    `json:"key" gorm:"column:key;size:128;not null;uniqueIndex"`
	Value       string    `json:"value" gorm:"column:value;not null"`
	Description string    `json:"description" gorm:"column:description;not null;default:''"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"column:updated_at;autoUpdateTime"`
}

type CustomerLevel struct {
	ID        int       `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"column:name;size:128;not null"`
	Sort      int       `json:"sort" gorm:"column:sort;not null;default:0"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
}

type CustomerSource struct {
	ID        int       `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"column:name;size:128;not null"`
	Sort      int       `json:"sort" gorm:"column:sort;not null;default:0"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
}

type SystemSettingsResponse struct {
	CustomerAutoDropEnabled bool             `json:"customerAutoDropEnabled" gorm:"-"`
	FollowUpDropDays        int              `json:"followUpDropDays" gorm:"-"`
	DealDropDays            int              `json:"dealDropDays" gorm:"-"`
	ClaimFreezeDays         int              `json:"claimFreezeDays" gorm:"-"`
	HolidayModeEnabled      bool             `json:"holidayModeEnabled" gorm:"-"`
	CustomerLimit           int              `json:"customerLimit" gorm:"-"`
	ShowFullContact         bool             `json:"showFullContact" gorm:"-"`
	ContractNumberPrefix    string           `json:"contractNumberPrefix" gorm:"-"`
	VisitPurposes           []string         `json:"visitPurposes" gorm:"-"`
	CustomerLevels          []CustomerLevel  `json:"customerLevels" gorm:"-"`
	CustomerSources         []CustomerSource `json:"customerSources" gorm:"-"`
}

type UpdateSystemSettingsRequest struct {
	CustomerAutoDropEnabled *bool    `json:"customerAutoDropEnabled" gorm:"-"`
	FollowUpDropDays        *int     `json:"followUpDropDays" gorm:"-"`
	DealDropDays            *int     `json:"dealDropDays" gorm:"-"`
	ClaimFreezeDays         *int     `json:"claimFreezeDays" gorm:"-"`
	HolidayModeEnabled      *bool    `json:"holidayModeEnabled" gorm:"-"`
	CustomerLimit           *int     `json:"customerLimit" gorm:"-"`
	ShowFullContact         *bool    `json:"showFullContact" gorm:"-"`
	ContractNumberPrefix    *string  `json:"contractNumberPrefix" gorm:"-"`
	VisitPurposes           []string `json:"visitPurposes" gorm:"-"`
}

type CustomerLevelRequest struct {
	Name string `json:"name" binding:"required" gorm:"-"`
	Sort int    `json:"sort" gorm:"-"`
}

type CustomerSourceRequest struct {
	Name string `json:"name" binding:"required" gorm:"-"`
	Sort int    `json:"sort" gorm:"-"`
}

func (SystemSetting) TableName() string  { return "system_settings" }
func (CustomerLevel) TableName() string  { return "customer_levels" }
func (CustomerSource) TableName() string { return "customer_sources" }
