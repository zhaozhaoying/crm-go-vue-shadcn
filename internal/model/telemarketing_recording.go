package model

import "time"

type TelemarketingRecording struct {
	ID                    string     `json:"id" gorm:"column:id;primaryKey"`
	CCNumber              string     `json:"ccNumber" gorm:"column:cc_number"`
	SID                   int64      `json:"sid" gorm:"column:sid"`
	SeID                  int64      `json:"seid" gorm:"column:seid"`
	Ccgeid                int64      `json:"ccgeid" gorm:"column:ccgeid"`
	CallType              int        `json:"callType" gorm:"column:call_type"`
	OutlineNumber         string     `json:"outlineNumber" gorm:"column:outline_number"`
	EncryptedOutlineNum   string     `json:"encryptedOutlineNumber" gorm:"column:encrypted_outline_number"`
	SwitchNumber          string     `json:"switchNumber" gorm:"column:switch_number"`
	Initiator             string     `json:"initiator" gorm:"column:initiator"`
	InitiatorCallID       string     `json:"initiatorCallId" gorm:"column:initiator_call_id"`
	ServiceNumber         string     `json:"serviceNumber" gorm:"column:service_number"`
	ServiceUID            int64      `json:"serviceUid" gorm:"column:service_uid"`
	ServiceSeatName       string     `json:"serviceSeatName" gorm:"column:service_seat_name"`
	ServiceSeatWorkNumber string     `json:"serviceSeatWorkNumber" gorm:"column:service_seat_worknumber"`
	ServiceGroupName      string     `json:"serviceGroupName" gorm:"column:service_group_name"`
	InitiateTime          int64      `json:"initiateTime" gorm:"column:initiate_time"`
	RingTime              int64      `json:"ringTime" gorm:"column:ring_time"`
	ConfirmTime           int64      `json:"confirmTime" gorm:"column:confirm_time"`
	DisconnectTime        int64      `json:"disconnectTime" gorm:"column:disconnect_time"`
	ConversationTime      int64      `json:"conversationTime" gorm:"column:conversation_time"`
	DurationSecond        int        `json:"durationSecond" gorm:"column:duration_second"`
	DurationText          string     `json:"durationText" gorm:"column:duration_text"`
	ValidDurationText     string     `json:"validDurationText" gorm:"column:valid_duration_text"`
	CustomerRingDuration  int        `json:"customerRingDuration" gorm:"column:customer_ring_duration"`
	SeatRingDuration      int        `json:"seatRingDuration" gorm:"column:seat_ring_duration"`
	RecordStatus          int        `json:"recordStatus" gorm:"column:record_status"`
	RecordFilename        string     `json:"recordFilename" gorm:"column:record_filename"`
	RecordResToken        string     `json:"recordResToken" gorm:"column:record_res_token"`
	EvaluateValue         string     `json:"evaluateValue" gorm:"column:evaluate_value"`
	CMResult              string     `json:"cmResult" gorm:"column:cm_result"`
	CMDescription         string     `json:"cmDescription" gorm:"column:cm_description"`
	Attribution           string     `json:"attribution" gorm:"column:attribution"`
	StopReason            int        `json:"stopReason" gorm:"column:stop_reason"`
	CustomerFailReason    string     `json:"customerFailReason" gorm:"column:customer_fail_reason"`
	CustomerName          string     `json:"customerName" gorm:"column:customer_name"`
	CustomerCompany       string     `json:"customerCompany" gorm:"column:customer_company"`
	GroupNames            string     `json:"groupNames" gorm:"column:group_names"`
	SeatNames             string     `json:"seatNames" gorm:"column:seat_names"`
	SeatNumbers           string     `json:"seatNumbers" gorm:"column:seat_numbers"`
	SeatWorkNumbers       string     `json:"seatWorkNumbers" gorm:"column:seat_work_numbers"`
	EnterpriseName        string     `json:"enterpriseName" gorm:"column:enterprise_name"`
	DistrictName          string     `json:"districtName" gorm:"column:district_name"`
	ServiceDeviceNumber   string     `json:"serviceDeviceNumber" gorm:"column:service_device_number"`
	CallAnswerResult      int        `json:"callAnswerResult" gorm:"column:call_answer_result"`
	CallHangupParty       int        `json:"callHangupParty" gorm:"column:call_hangup_party"`
	MatchedUserID         *int64     `json:"matchedUserId,omitempty" gorm:"column:matched_user_id"`
	MatchedUserName       string     `json:"matchedUserName" gorm:"column:matched_user_name"`
	RoleName              string     `json:"roleName" gorm:"column:role_name"`
	RemoteCreatedAt       *time.Time `json:"remoteCreatedAt,omitempty" gorm:"column:remote_created_at"`
	RemoteUpdatedAt       *time.Time `json:"remoteUpdatedAt,omitempty" gorm:"column:remote_updated_at"`
	RawPayload            string     `json:"-" gorm:"column:raw_payload"`
	CreatedAt             time.Time  `json:"-" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt             time.Time  `json:"-" gorm:"column:updated_at;autoUpdateTime"`
}

func (TelemarketingRecording) TableName() string { return "mihua_call_recordings" }

type TelemarketingRecordingMatchedUser struct {
	UserID     int64  `json:"userId"`
	Username   string `json:"username"`
	Nickname   string `json:"nickname"`
	UserName   string `json:"userName"`
	RoleName   string `json:"roleName"`
	WorkNumber string `json:"workNumber"`
}

type TelemarketingRecordingUpsertInput struct {
	ID                    string
	CCNumber              string
	SID                   int64
	SeID                  int64
	Ccgeid                int64
	CallType              int
	OutlineNumber         string
	EncryptedOutlineNum   string
	SwitchNumber          string
	Initiator             string
	InitiatorCallID       string
	ServiceNumber         string
	ServiceUID            int64
	ServiceSeatName       string
	ServiceSeatWorkNumber string
	ServiceGroupName      string
	InitiateTime          int64
	RingTime              int64
	ConfirmTime           int64
	DisconnectTime        int64
	ConversationTime      int64
	DurationSecond        int
	DurationText          string
	ValidDurationText     string
	CustomerRingDuration  int
	SeatRingDuration      int
	RecordStatus          int
	RecordFilename        string
	RecordResToken        string
	EvaluateValue         string
	CMResult              string
	CMDescription         string
	Attribution           string
	StopReason            int
	CustomerFailReason    string
	CustomerName          string
	CustomerCompany       string
	GroupNames            string
	SeatNames             string
	SeatNumbers           string
	SeatWorkNumbers       string
	EnterpriseName        string
	DistrictName          string
	ServiceDeviceNumber   string
	CallAnswerResult      int
	CallHangupParty       int
	MatchedUserID         *int64
	MatchedUserName       string
	RoleName              string
	RemoteCreatedAt       *time.Time
	RemoteUpdatedAt       *time.Time
	RawPayload            string
}

type TelemarketingRecordingListFilter struct {
	ShowAll           bool   `gorm:"-"`
	ViewerMihuaWorkNo string `gorm:"-"`
	Keyword           string `gorm:"-"`
	StartDate         string `gorm:"-"`
	EndDate           string `gorm:"-"`
	MinDuration       int    `gorm:"-"`
	MaxDuration       int    `gorm:"-"`
	Page              int    `gorm:"-"`
	PageSize          int    `gorm:"-"`
}

type TelemarketingRecordingListResult struct {
	Items    []TelemarketingRecording `json:"items" gorm:"-"`
	Total    int64                    `json:"total" gorm:"-"`
	Page     int                      `json:"page" gorm:"-"`
	PageSize int                      `json:"pageSize" gorm:"-"`
}

type TelemarketingRecordingDetail struct {
	Recording         TelemarketingRecording `json:"recording"`
	PlaybackURL       string                 `json:"playbackUrl"`
	PlaybackFilename  string                 `json:"playbackFilename"`
	PlaybackExpiresAt int64                  `json:"playbackExpiresAt"`
}
