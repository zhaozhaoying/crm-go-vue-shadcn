package model

import "time"

type CallRecording struct {
	ID               string    `json:"id" gorm:"column:id;primaryKey"`
	AgentCode        int64     `json:"agentCode" gorm:"column:agent_code"`
	CallStatus       int       `json:"callStatus" gorm:"column:call_status"`
	CallStatusName   string    `json:"callStatusName" gorm:"column:call_status_name"`
	CallType         int       `json:"callType" gorm:"column:call_type"`
	CalleeAttr       string    `json:"calleeAttr" gorm:"column:callee_attr"`
	CallerAttr       string    `json:"callerAttr" gorm:"column:caller_attr"`
	CreateTime       int64     `json:"createTime" gorm:"column:create_time"`
	DeptName         string    `json:"deptName" gorm:"column:dept_name"`
	Duration         int       `json:"duration" gorm:"column:duration"`
	EndTime          int64     `json:"endTime" gorm:"column:end_time"`
	EnterpriseName   string    `json:"enterpriseName" gorm:"column:enterprise_name"`
	FinishStatus     int       `json:"finishStatus" gorm:"column:finish_status"`
	FinishStatusName string    `json:"finishStatusName" gorm:"column:finish_status_name"`
	Handle           int       `json:"handle" gorm:"column:handle"`
	InterfaceID      string    `json:"interfaceId" gorm:"column:interface_id"`
	InterfaceName    string    `json:"interfaceName" gorm:"column:interface_name"`
	LineName         string    `json:"lineName" gorm:"column:line_name"`
	Mobile           string    `json:"mobile" gorm:"column:mobile"`
	Mode             int       `json:"mode" gorm:"column:mode"`
	MoveBatchCode    *string   `json:"moveBatchCode,omitempty" gorm:"column:move_batch_code"`
	OctCustomerID    *string   `json:"octCustomerId,omitempty" gorm:"column:oct_customer_id"`
	Phone            string    `json:"phone" gorm:"column:phone"`
	Postage          float64   `json:"postage" gorm:"column:postage"`
	PreRecordURL     string    `json:"preRecordUrl" gorm:"column:pre_record_url"`
	RealName         string    `json:"realName" gorm:"column:real_name"`
	StartTime        int64     `json:"startTime" gorm:"column:start_time"`
	Status           int       `json:"status" gorm:"column:status"`
	TelA             string    `json:"telA" gorm:"column:tel_a"`
	TelB             string    `json:"telB" gorm:"column:tel_b"`
	TelX             string    `json:"telX" gorm:"column:tel_x"`
	TenantCode       string    `json:"tenantCode" gorm:"column:tenant_code"`
	UpdateTime       int64     `json:"updateTime" gorm:"column:update_time"`
	UserID           string    `json:"userId" gorm:"column:user_id"`
	WorkNum          *string   `json:"workNum,omitempty" gorm:"column:work_num"`
	DedupeKey        string    `json:"-" gorm:"column:dedupe_key"`
	CreatedAt        time.Time `json:"-" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time `json:"-" gorm:"column:updated_at;autoUpdateTime"`
}

func (CallRecording) TableName() string { return "call_recordings" }

type CallRecordingUpsertInput struct {
	ID               string
	AgentCode        int64
	CallStatus       int
	CallStatusName   string
	CallType         int
	CalleeAttr       string
	CallerAttr       string
	CreateTime       int64
	DeptName         string
	Duration         int
	EndTime          int64
	EnterpriseName   string
	FinishStatus     int
	FinishStatusName string
	Handle           int
	InterfaceID      string
	InterfaceName    string
	LineName         string
	Mobile           string
	Mode             int
	MoveBatchCode    *string
	OctCustomerID    *string
	Phone            string
	Postage          float64
	PreRecordURL     string
	RealName         string
	StartTime        int64
	Status           int
	TelA             string
	TelB             string
	TelX             string
	TenantCode       string
	UpdateTime       int64
	UserID           string
	WorkNum          *string
}

type CallRecordingListFilter struct {
	ShowAll                  bool   `gorm:"-"`
	ViewerHanghangCRMMobile  string `gorm:"-"`
	Keyword                  string `gorm:"-"`
	MinDuration              int    `gorm:"-"`
	MaxDuration              int    `gorm:"-"`
	Page                     int    `gorm:"-"`
	PageSize                 int    `gorm:"-"`
}

type CallRecordingListResult struct {
	Items    []CallRecording `json:"items" gorm:"-"`
	Total    int64           `json:"total" gorm:"-"`
	Page     int             `json:"page" gorm:"-"`
	PageSize int             `json:"pageSize" gorm:"-"`
}
