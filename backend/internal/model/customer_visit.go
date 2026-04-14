package model

import "time"

// CustomerVisit 上门拜访签到记录
type CustomerVisit struct {
	ID               int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	OperatorUserID   int64     `json:"operatorUserId" gorm:"column:operator_user_id;not null;index"`
	OperatorUserName string    `json:"operatorUserName,omitempty" gorm:"-"`
	CustomerName     string    `json:"customerName" gorm:"column:customer_name;not null"`
	Inviter          string    `json:"inviter" gorm:"column:inviter;not null;default:''"`
	CheckInIP        string    `json:"-" gorm:"column:check_in_ip;size:64;not null;default:''"`
	CheckInLat       float64   `json:"checkInLat" gorm:"column:check_in_lat;not null;default:0"`
	CheckInLng       float64   `json:"checkInLng" gorm:"column:check_in_lng;not null;default:0"`
	Province         string    `json:"province" gorm:"column:province;size:64;not null;default:''"`
	City             string    `json:"city" gorm:"column:city;size:64;not null;default:''"`
	Area             string    `json:"area" gorm:"column:area;size:64;not null;default:''"`
	DetailAddress    string    `json:"detailAddress" gorm:"column:detail_address;not null;default:''"`
	Images           string    `json:"images" gorm:"column:images;not null;default:'[]'"`
	VisitPurpose     string    `json:"visitPurpose" gorm:"column:visit_purpose;not null;default:''"`
	Remark           string    `json:"remark" gorm:"column:remark;not null;default:''"`
	VisitDate        string    `json:"visitDate" gorm:"column:visit_date;not null"`
	CreatedAt        time.Time `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time `json:"updatedAt" gorm:"column:updated_at;autoUpdateTime"`

	Operator *User `json:"-" gorm:"foreignKey:OperatorUserID;references:ID"`
}

func (CustomerVisit) TableName() string { return "customer_visits" }

// CustomerVisitCreateInput 创建上门拜访输入
type CustomerVisitCreateInput struct {
	OperatorUserID int64   `gorm:"-"`
	CustomerName   string  `gorm:"-"`
	Inviter        string  `gorm:"-"`
	CheckInIP      string  `gorm:"-"`
	CheckInLat     float64 `gorm:"-"`
	CheckInLng     float64 `gorm:"-"`
	Province       string  `gorm:"-"`
	City           string  `gorm:"-"`
	Area           string  `gorm:"-"`
	DetailAddress  string  `gorm:"-"`
	Images         string  `gorm:"-"`
	VisitPurpose   string  `gorm:"-"`
	Remark         string  `gorm:"-"`
	VisitDate      string  `gorm:"-"`
}

// CustomerVisitListFilter 上门拜访列表过滤
type CustomerVisitListFilter struct {
	OperatorUserID int64      `gorm:"-"`
	CanViewAll     bool       `gorm:"-"`
	Keyword        string     `gorm:"-"`
	StartTime      *time.Time `gorm:"-"`
	EndTime        *time.Time `gorm:"-"`
	Page           int        `gorm:"-"`
	PageSize       int        `gorm:"-"`
}

// CustomerVisitListResult 上门拜访列表结果
type CustomerVisitListResult struct {
	Items    []CustomerVisit `json:"items" gorm:"-"`
	Total    int64           `json:"total" gorm:"-"`
	Page     int             `json:"page" gorm:"-"`
	PageSize int             `json:"pageSize" gorm:"-"`
}
