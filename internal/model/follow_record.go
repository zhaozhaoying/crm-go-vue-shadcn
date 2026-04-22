package model

import "time"

// FollowMethod 跟进方式
type FollowMethod struct {
	ID        int       `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"column:name;size:128;not null;uniqueIndex"`
	Sort      int       `json:"sort" gorm:"column:sort;not null;default:0"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
}

// OperationFollowRecord 运营跟进记录
type OperationFollowRecordCustomer struct {
	ID            int64      `json:"id" gorm:"-"`
	Name          string     `json:"name,omitempty" gorm:"-"`
	LegalName     string     `json:"legalName,omitempty" gorm:"-"`
	ContactName   string     `json:"contactName,omitempty" gorm:"-"`
	Weixin        string     `json:"weixin,omitempty" gorm:"-"`
	Email         string     `json:"email,omitempty" gorm:"-"`
	PrimaryPhone  string     `json:"primaryPhone,omitempty" gorm:"-"`
	Province      int        `json:"province,omitempty" gorm:"-"`
	City          int        `json:"city,omitempty" gorm:"-"`
	Area          int        `json:"area,omitempty" gorm:"-"`
	DetailAddress string     `json:"detailAddress,omitempty" gorm:"-"`
	Remark        string     `json:"remark,omitempty" gorm:"-"`
	Status        string     `json:"status,omitempty" gorm:"-"`
	DealStatus    string     `json:"dealStatus,omitempty" gorm:"-"`
	OwnerUserID   *int64     `json:"ownerUserId,omitempty" gorm:"-"`
	OwnerUserName string     `json:"ownerUserName,omitempty" gorm:"-"`
	NextTime      *time.Time `json:"nextTime,omitempty" gorm:"-"`
	FollowTime    *time.Time `json:"followTime,omitempty" gorm:"-"`
	CollectTime   *time.Time `json:"collectTime,omitempty" gorm:"-"`
	LevelID       int        `json:"levelId,omitempty" gorm:"-"`
	LevelName     string     `json:"levelName,omitempty" gorm:"-"`
	SourceID      int        `json:"sourceId,omitempty" gorm:"-"`
	SourceName    string     `json:"sourceName,omitempty" gorm:"-"`
}

type OperationFollowRecord struct {
	ID                 int64                          `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	CustomerID         int64                          `json:"customerId" gorm:"column:customer_id;not null;index"`
	Customer           *OperationFollowRecordCustomer `json:"customer" gorm:"-"`
	Content            string                         `json:"content" gorm:"column:content;not null"`
	NextFollowTime     *time.Time                     `json:"nextFollowTime,omitempty" gorm:"-"`
	AppointmentTime    *time.Time                     `json:"appointmentTime" gorm:"-"`
	ShootingTime       *time.Time                     `json:"shootingTime" gorm:"-"`
	CustomerLevelID    int                            `json:"customerLevelId" gorm:"column:customer_level_id;not null;default:0"`
	CustomerLevelName  string                         `json:"customerLevelName,omitempty" gorm:"-"`
	CustomerSourceID   int                            `json:"customerSourceId,omitempty" gorm:"-"`
	CustomerSourceName string                         `json:"customerSourceName,omitempty" gorm:"-"`
	FollowMethodID     int                            `json:"followMethodId" gorm:"column:follow_method_id;not null;default:0"`
	FollowMethodName   string                         `json:"followMethodName,omitempty" gorm:"-"`
	OperatorUserID     int64                          `json:"operatorUserId" gorm:"column:operator_user_id;not null;index"`
	OperatorUserName   string                         `json:"operatorUserName,omitempty" gorm:"-"`
	CreatedAt          time.Time                      `json:"createdAt" gorm:"-"`
	UpdatedAt          time.Time                      `json:"updatedAt" gorm:"-"`

	CustomerRelation *Customer      `json:"-" gorm:"foreignKey:CustomerID;references:ID"`
	Operator         *User          `json:"-" gorm:"foreignKey:OperatorUserID;references:ID"`
	Level            *CustomerLevel `json:"-" gorm:"foreignKey:CustomerLevelID;references:ID"`
	FollowMethod     *FollowMethod  `json:"-" gorm:"foreignKey:FollowMethodID;references:ID"`
}

// SalesFollowRecord 销售跟进记录
type SalesFollowRecord struct {
	ID                 int64                          `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	CustomerID         int64                          `json:"customerId" gorm:"column:customer_id;not null;index"`
	Customer           *OperationFollowRecordCustomer `json:"customer" gorm:"-"`
	Content            string                         `json:"content" gorm:"column:content;not null"`
	NextFollowTime     *time.Time                     `json:"nextFollowTime,omitempty" gorm:"-"`
	CustomerLevelID    int                            `json:"customerLevelId" gorm:"column:customer_level_id;not null;default:0"`
	CustomerLevelName  string                         `json:"customerLevelName,omitempty" gorm:"-"`
	CustomerSourceID   int                            `json:"customerSourceId" gorm:"column:customer_source_id;not null;default:0"`
	CustomerSourceName string                         `json:"customerSourceName,omitempty" gorm:"-"`
	FollowMethodID     int                            `json:"followMethodId" gorm:"column:follow_method_id;not null;default:0"`
	FollowMethodName   string                         `json:"followMethodName,omitempty" gorm:"-"`
	OperatorUserID     int64                          `json:"operatorUserId" gorm:"column:operator_user_id;not null;index"`
	OperatorUserName   string                         `json:"operatorUserName,omitempty" gorm:"-"`
	CreatedAt          time.Time                      `json:"createdAt" gorm:"-"`
	UpdatedAt          time.Time                      `json:"updatedAt" gorm:"-"`

	CustomerRelation *Customer       `json:"-" gorm:"foreignKey:CustomerID;references:ID"`
	Operator         *User           `json:"-" gorm:"foreignKey:OperatorUserID;references:ID"`
	Level            *CustomerLevel  `json:"-" gorm:"foreignKey:CustomerLevelID;references:ID"`
	Source           *CustomerSource `json:"-" gorm:"foreignKey:CustomerSourceID;references:ID"`
	FollowMethod     *FollowMethod   `json:"-" gorm:"foreignKey:FollowMethodID;references:ID"`
}

// FollowRecordCreateInput 创建跟进记录输入
type FollowRecordCreateInput struct {
	CustomerID       int64      `gorm:"-"`
	Content          string     `gorm:"-"`
	NextFollowTime   *time.Time `gorm:"-"`
	AppointmentTime  *time.Time `gorm:"-"`
	ShootingTime     *time.Time `gorm:"-"`
	CustomerLevelID  int        `gorm:"-"`
	CustomerSourceID int        `gorm:"-"`
	FollowMethodID   int        `gorm:"-"`
	OperatorUserID   int64      `gorm:"-"`
}

// FollowRecordListFilter 跟进记录列表过滤
type FollowRecordListFilter struct {
	CustomerID int64 `gorm:"-"`
	Page       int   `gorm:"-"`
	PageSize   int   `gorm:"-"`
}

// OperationFollowRecordListResult 运营跟进记录列表结果
type OperationFollowRecordListResult struct {
	Items    []OperationFollowRecord `json:"items" gorm:"-"`
	Total    int64                   `json:"total" gorm:"-"`
	Page     int                     `json:"page" gorm:"-"`
	PageSize int                     `json:"pageSize" gorm:"-"`
}

// SalesFollowRecordListResult 销售跟进记录列表结果
type SalesFollowRecordListResult struct {
	Items    []SalesFollowRecord `json:"items" gorm:"-"`
	Total    int64               `json:"total" gorm:"-"`
	Page     int                 `json:"page" gorm:"-"`
	PageSize int                 `json:"pageSize" gorm:"-"`
}

// FollowMethodRequest 跟进方式请求
type FollowMethodRequest struct {
	Name string `json:"name" binding:"required" gorm:"-"`
	Sort int    `json:"sort" gorm:"-"`
}

func (FollowMethod) TableName() string          { return "follow_methods" }
func (OperationFollowRecord) TableName() string { return "operation_follow_records" }
func (SalesFollowRecord) TableName() string     { return "sales_follow_records" }
