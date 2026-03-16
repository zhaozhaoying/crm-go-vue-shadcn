package model

import "time"

const (
	CustomerDealStatusUndone = "undone"
	CustomerDealStatusDone   = "done"

	CustomerStatusPool  = "pool"
	CustomerStatusOwned = "owned"
)

type CustomerPhone struct {
	ID         int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	CustomerID int64     `json:"customerId" gorm:"column:customer_id;not null;index"`
	Phone      string    `json:"phone" gorm:"column:phone;not null"`
	PhoneLabel string    `json:"phoneLabel,omitempty" gorm:"column:phone_label"` // 手机/座机/其他
	IsPrimary  bool      `json:"isPrimary" gorm:"column:is_primary;not null;default:0"`
	CreatedAt  time.Time `json:"createdAt" gorm:"-"`
	UpdatedAt  time.Time `json:"updatedAt" gorm:"-"`

	CreatedAtUnix int64 `json:"-" gorm:"column:created_at"`
	UpdatedAtUnix int64 `json:"-" gorm:"column:updated_at"`

	Customer *Customer `json:"-" gorm:"foreignKey:CustomerID;references:ID"`
}

type CustomerStatusLog struct {
	ID             int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	CustomerID     int64     `json:"customerId" gorm:"column:customer_id;not null;index"`
	FromStatus     int       `json:"fromStatus" gorm:"column:from_status;not null"`
	ToStatus       int       `json:"toStatus" gorm:"column:to_status;not null"`
	TriggerType    int       `json:"triggerType" gorm:"column:trigger_type;not null;default:0"` // 0手动 1领取 2丢弃 3跟进超时 4签单超时 5成交
	Reason         string    `json:"reason,omitempty" gorm:"column:reason"`
	OperatorUserID *int64    `json:"operatorUserId,omitempty" gorm:"column:operator_user_id"`
	OperatorName   string    `json:"operatorName,omitempty" gorm:"-"`
	OperateTime    time.Time `json:"operateTime" gorm:"-"`

	OperateTimeUnix int64 `json:"-" gorm:"column:operate_time;not null"`

	Customer *Customer `json:"-" gorm:"foreignKey:CustomerID;references:ID"`
	Operator *User     `json:"-" gorm:"foreignKey:OperatorUserID;references:ID"`
}

type Customer struct {
	ID                 int64           `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Name               string          `json:"name" gorm:"column:name;not null"`
	LegalName          string          `json:"legalName,omitempty" gorm:"column:legal_name;not null;default:''"`
	ContactName        string          `json:"contactName,omitempty" gorm:"column:contact_name;not null;default:''"`
	Weixin             string          `json:"weixin,omitempty" gorm:"column:weixin;not null;default:''"`
	Email              string          `json:"email" gorm:"column:email;not null;default:''"`
	CustomerLevelID    int             `json:"customerLevelId,omitempty" gorm:"column:customer_level_id;not null;default:0"`
	CustomerSourceID   int             `json:"customerSourceId,omitempty" gorm:"column:customer_source_id;not null;default:0"`
	CustomerLevelName  string          `json:"customerLevelName,omitempty" gorm:"-"`
	CustomerSourceName string          `json:"customerSourceName,omitempty" gorm:"-"`
	Province           int             `json:"province,omitempty" gorm:"column:province;not null;default:0"`
	City               int             `json:"city,omitempty" gorm:"column:city;not null;default:0"`
	Area               int             `json:"area,omitempty" gorm:"column:area;not null;default:0"`
	DetailAddress      string          `json:"detailAddress,omitempty" gorm:"column:detail_address;not null;default:''"`
	Lng                float64         `json:"lng,omitempty" gorm:"column:lng;not null;default:0"`
	Lat                float64         `json:"lat,omitempty" gorm:"column:lat;not null;default:0"`
	NextTime           *time.Time      `json:"nextTime,omitempty" gorm:"-"`
	FollowTime         *time.Time      `json:"followTime,omitempty" gorm:"-"`
	Remark             string          `json:"remark,omitempty" gorm:"column:remark"`
	Status             string          `json:"status" gorm:"column:status;not null;default:'pool';index"`
	DealStatus         string          `json:"dealStatus" gorm:"column:deal_status;not null;default:'undone';index"`
	DealTime           *time.Time      `json:"dealTime,omitempty" gorm:"-"`
	CustomerStatus     int             `json:"customerStatus,omitempty" gorm:"column:customer_status;not null;default:0"`
	CollectTime        *time.Time      `json:"collectTime,omitempty" gorm:"-"`
	DropTime           *time.Time      `json:"dropTime,omitempty" gorm:"-"`
	DropUserID         *int64          `json:"dropUserId,omitempty" gorm:"column:drop_user_id"`
	CreateUserID       int64           `json:"createUserId,omitempty" gorm:"column:create_user_id;not null;default:0"`
	OwnerUserID        *int64          `json:"ownerUserId" gorm:"column:owner_user_id;index"`
	OperateUserID      int64           `json:"operateUserId,omitempty" gorm:"column:operate_user_id;not null;default:0"`
	OwnerUserName      string          `json:"ownerUserName,omitempty" gorm:"-"`
	IsLock             bool            `json:"isLock,omitempty" gorm:"column:is_lock;not null;default:0"`
	Phones             []CustomerPhone `json:"phones,omitempty" gorm:"foreignKey:CustomerID;references:ID"`
	IsInPool           bool            `json:"isInPool" gorm:"-"`
	HistoricalOwnerIDs []int64         `json:"historicalOwnerIds,omitempty" gorm:"-"`
	CreatedAt          time.Time       `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt          time.Time       `json:"updatedAt" gorm:"column:updated_at;autoUpdateTime"`
	DeleteTime         *time.Time      `json:"deleteTime,omitempty" gorm:"-"`

	NextTimeUnix    *int64 `json:"-" gorm:"column:next_time"`
	FollowTimeUnix  *int64 `json:"-" gorm:"column:follow_time"`
	DealTimeUnix    *int64 `json:"-" gorm:"column:deal_time"`
	CollectTimeUnix *int64 `json:"-" gorm:"column:collect_time"`
	DropTimeUnix    *int64 `json:"-" gorm:"column:drop_time"`
	DeleteTimeUnix  *int64 `json:"-" gorm:"column:delete_time"`

	OwnerUser *User           `json:"-" gorm:"foreignKey:OwnerUserID;references:ID"`
	Level     *CustomerLevel  `json:"-" gorm:"foreignKey:CustomerLevelID;references:ID"`
	Source    *CustomerSource `json:"-" gorm:"foreignKey:CustomerSourceID;references:ID"`
}

type CustomerListFilter struct {
	Category       string `gorm:"-"`
	OwnershipScope string `gorm:"-"`
	Keyword        string `gorm:"-"`
	Name           string `gorm:"-"`
	ContactName    string `gorm:"-"`
	Phone          string `gorm:"-"`
	Weixin         string `gorm:"-"`
	OwnerUserName  string `gorm:"-"`
	Province       int    `gorm:"-"`
	City           int    `gorm:"-"`
	Area           int    `gorm:"-"`
	ExcludePool    bool   `gorm:"-"`
	SortBy         string `gorm:"-"`
	Page           int    `gorm:"-"`
	PageSize       int    `gorm:"-"`
	ViewerID       int64  `gorm:"-"`
	HasViewer      bool   `gorm:"-"`
	ActorRole      string `gorm:"-"`

	// AllowedOwnerUserIDs is used by role-based scope for "my" customer list.
	AllowedOwnerUserIDs []int64 `gorm:"-"`
	// SkipViewerOwnerLimit is used by admin "all" scope on "my" category.
	SkipViewerOwnerLimit bool `gorm:"-"`
	// IncludePoolInMyScope allows the "my" category to include pool customers.
	IncludePoolInMyScope bool `gorm:"-"`
	// IncludeCreatorScope allows the "my" category to also include customers
	// where create_user_id = ViewerID, used for inside-sales roles so they can
	// see customers they created and assigned to outside-sales staff.
	IncludeCreatorScope bool `gorm:"-"`
}

type CustomerListResult struct {
	Items    []Customer `json:"items" gorm:"-"`
	Total    int64      `json:"total" gorm:"-"`
	Page     int        `json:"page" gorm:"-"`
	PageSize int        `json:"pageSize" gorm:"-"`
}

type CustomerTransferInput struct {
	CustomerID     int64 `gorm:"-"`
	ToOwnerUserID  int64 `gorm:"-"`
	OperatorUserID int64 `gorm:"-"`
}

type CustomerPhoneInput struct {
	Phone      string `gorm:"-"`
	PhoneLabel string `gorm:"-"`
	IsPrimary  bool   `gorm:"-"`
}

type CustomerCreateInput struct {
	Name           string               `gorm:"-"`
	LegalName      string               `gorm:"-"`
	ContactName    string               `gorm:"-"`
	Weixin         string               `gorm:"-"`
	Email          string               `gorm:"-"`
	Province       int                  `gorm:"-"`
	City           int                  `gorm:"-"`
	Area           int                  `gorm:"-"`
	DetailAddress  string               `gorm:"-"`
	Remark         string               `gorm:"-"`
	Status         string               `gorm:"-"`
	OwnerUserID    *int64               `gorm:"-"`
	OperatorUserID int64                `gorm:"-"`
	Phones         []CustomerPhoneInput `gorm:"-"`
}

type CustomerUpdateInput struct {
	Name           string               `gorm:"-"`
	LegalName      string               `gorm:"-"`
	ContactName    string               `gorm:"-"`
	Weixin         string               `gorm:"-"`
	Email          string               `gorm:"-"`
	Province       int                  `gorm:"-"`
	City           int                  `gorm:"-"`
	Area           int                  `gorm:"-"`
	DetailAddress  string               `gorm:"-"`
	Remark         string               `gorm:"-"`
	OperatorUserID int64                `gorm:"-"`
	Phones         []CustomerPhoneInput `gorm:"-"`
}

type CustomerUniqueCheckInput struct {
	ExcludeCustomerID *int64   `gorm:"-"`
	Name              string   `gorm:"-"`
	LegalName         string   `gorm:"-"`
	Weixin            string   `gorm:"-"`
	Phones            []string `gorm:"-"`
}

type CustomerUniqueCheckResult struct {
	NameExists      bool     `json:"nameExists" gorm:"-"`
	LegalNameExists bool     `json:"legalNameExists" gorm:"-"`
	WeixinExists    bool     `json:"weixinExists" gorm:"-"`
	DuplicatePhones []string `json:"duplicatePhones" gorm:"-"`
}

func (Customer) TableName() string          { return "customers" }
func (CustomerPhone) TableName() string     { return "customer_phones" }
func (CustomerStatusLog) TableName() string { return "customer_status_logs" }
