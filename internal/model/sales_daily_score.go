package model

import "time"

const (
	SalesDailyScoreCallScoreTypeNone     = "none"
	SalesDailyScoreCallScoreTypeCallNum  = "call_num"
	SalesDailyScoreCallScoreTypeDuration = "call_duration"
)

type SalesDailyScore struct {
	ID                  int64      `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ScoreDate           string     `json:"scoreDate" gorm:"column:score_date"`
	UserID              int64      `json:"userId" gorm:"column:user_id"`
	UserName            string     `json:"userName" gorm:"column:user_name"`
	RoleName            string     `json:"roleName" gorm:"column:role_name"`
	CallNum             int        `json:"callNum" gorm:"column:call_num"`
	CallDurationSecond  int        `json:"callDurationSecond" gorm:"column:call_duration_second"`
	CallScoreByCount    int        `json:"callScoreByCount" gorm:"column:call_score_by_count"`
	CallScoreByDuration int        `json:"callScoreByDuration" gorm:"column:call_score_by_duration"`
	CallScoreType       string     `json:"callScoreType" gorm:"column:call_score_type"`
	CallScore           int        `json:"callScore" gorm:"column:call_score"`
	VisitCount          int        `json:"visitCount" gorm:"column:visit_count"`
	VisitScore          int        `json:"visitScore" gorm:"column:visit_score"`
	NewCustomerCount    int        `json:"newCustomerCount" gorm:"column:new_customer_count"`
	NewCustomerScore    int        `json:"newCustomerScore" gorm:"column:new_customer_score"`
	TotalScore          int        `json:"totalScore" gorm:"column:total_score"`
	ScoreReachedAt      *time.Time `json:"scoreReachedAt,omitempty" gorm:"column:score_reached_at"`
	CreatedAt           time.Time  `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt           time.Time  `json:"updatedAt" gorm:"column:updated_at"`
}

type SalesDailyScoreUpsertInput struct {
	ScoreDate           string
	UserID              int64
	UserName            string
	RoleName            string
	CallNum             int
	CallDurationSecond  int
	CallScoreByCount    int
	CallScoreByDuration int
	CallScoreType       string
	CallScore           int
	VisitCount          int
	VisitScore          int
	NewCustomerCount    int
	NewCustomerScore    int
	TotalScore          int
	ScoreReachedAt      *time.Time
}

type SalesDailyScoreUser struct {
	UserID   int64  `json:"userId" gorm:"column:user_id"`
	UserName string `json:"userName" gorm:"column:user_name"`
	RoleName string `json:"roleName" gorm:"column:role_name"`
}

type DailySalesCallMetric struct {
	UserID             int64 `json:"userId" gorm:"column:user_id"`
	CallNum            int   `json:"callNum" gorm:"column:call_num"`
	CallDurationSecond int   `json:"callDurationSecond" gorm:"column:call_duration_second"`
}

type DailySalesCallEvent struct {
	UserID         int64     `json:"userId" gorm:"column:user_id"`
	EventTime      time.Time `json:"eventTime" gorm:"column:event_time"`
	DurationSecond int       `json:"durationSecond" gorm:"column:duration_second"`
}

type SalesDailyScoreRankingItem struct {
	Rank int `json:"rank" gorm:"-"`
	SalesDailyScore
}

type SalesDailyScoreRankingListResult struct {
	ScoreDate string                       `json:"scoreDate" gorm:"-"`
	Total     int                          `json:"total" gorm:"-"`
	Items     []SalesDailyScoreRankingItem `json:"items" gorm:"-"`
}

type SalesDailyScoreDetail struct {
	ScoreDate  string          `json:"scoreDate" gorm:"-"`
	Rank       int             `json:"rank" gorm:"-"`
	TotalUsers int             `json:"totalUsers" gorm:"-"`
	HasData    bool            `json:"hasData" gorm:"-"`
	Score      SalesDailyScore `json:"score" gorm:"-"`
}

type TelemarketingDailyScore struct {
	ScoreDate           string     `json:"scoreDate" gorm:"-"`
	SeatWorkNumber      string     `json:"seatWorkNumber" gorm:"-"`
	SeatName            string     `json:"seatName" gorm:"-"`
	MatchedUserID       *int64     `json:"matchedUserId,omitempty" gorm:"-"`
	MatchedUserName     string     `json:"matchedUserName" gorm:"-"`
	ServiceNumber       string     `json:"serviceNumber" gorm:"-"`
	GroupName           string     `json:"groupName" gorm:"-"`
	RoleName            string     `json:"roleName" gorm:"-"`
	CallNum             int        `json:"callNum" gorm:"-"`
	AnsweredCallCount   int        `json:"answeredCallCount" gorm:"-"`
	MissedCallCount     int        `json:"missedCallCount" gorm:"-"`
	AnswerRate          float64    `json:"answerRate" gorm:"-"`
	CallDurationSecond  int        `json:"callDurationSecond" gorm:"-"`
	NewCustomerCount    int        `json:"newCustomerCount" gorm:"-"`
	InvitationCount     int        `json:"invitationCount" gorm:"-"`
	CallScoreByCount    int        `json:"callScoreByCount" gorm:"-"`
	CallScoreByDuration int        `json:"callScoreByDuration" gorm:"-"`
	CallScoreType       string     `json:"callScoreType" gorm:"-"`
	CallScore           int        `json:"callScore" gorm:"-"`
	InvitationScore     int        `json:"invitationScore" gorm:"-"`
	NewCustomerScore    int        `json:"newCustomerScore" gorm:"-"`
	TotalScore          int        `json:"totalScore" gorm:"-"`
	ScoreReachedAt      *time.Time `json:"scoreReachedAt,omitempty" gorm:"-"`
	UpdatedAt           time.Time  `json:"updatedAt" gorm:"-"`
}

type TelemarketingDailyScoreRankingItem struct {
	Rank int `json:"rank" gorm:"-"`
	TelemarketingDailyScore
}

type TelemarketingDailyScoreRankingListResult struct {
	ScoreDate string                               `json:"scoreDate" gorm:"-"`
	Total     int                                  `json:"total" gorm:"-"`
	Items     []TelemarketingDailyScoreRankingItem `json:"items" gorm:"-"`
}

type TelemarketingDailyScoreDetail struct {
	ScoreDate  string                  `json:"scoreDate" gorm:"-"`
	Rank       int                     `json:"rank" gorm:"-"`
	TotalUsers int                     `json:"totalUsers" gorm:"-"`
	HasData    bool                    `json:"hasData" gorm:"-"`
	Score      TelemarketingDailyScore `json:"score" gorm:"-"`
}

type TelemarketingLocalUser struct {
	UserID     int64  `json:"userId" gorm:"-"`
	Username   string `json:"username" gorm:"-"`
	Nickname   string `json:"nickname" gorm:"-"`
	UserName   string `json:"userName" gorm:"-"`
	RoleName   string `json:"roleName" gorm:"-"`
	WorkNumber string `json:"workNumber" gorm:"-"`
}

type SpxxjjMiHuaSeatStatisticUpsertInput struct {
	ScoreDate              string
	SeatID                 int64
	SeatName               string
	WorkNumber             string
	ServiceNumber          string
	IsMobileSeat           string
	SeatType               int
	Ccgeid                 int64
	SuccessCallCount       int
	OutTotalSuccess        int
	OutTotalCallCount      int
	CallTotalTimeSecond    int
	CallValidTimeSecond    int
	OutCallTotalTimeSecond int
	OutCallValidTimeSecond int
	LatestStateTime        *time.Time
	LatestStateID          int
	StatTimestamp          *time.Time
	EnterpriseName         string
	DepartmentName         string
	GroupName              string
	SeatRealTimeStateJSON  string
	GroupsJSON             string
	RawPayload             string
	MatchedUserID          *int64
	MatchedUserName        string
	RoleName               string
}

type SpxxjjTelemarketingDailyScoreUpsertInput struct {
	ScoreDate           string
	SeatWorkNumber      string
	SeatName            string
	MatchedUserID       *int64
	MatchedUserName     string
	ServiceNumber       string
	GroupName           string
	RoleName            string
	CallNum             int
	AnsweredCallCount   int
	MissedCallCount     int
	AnswerRate          float64
	CallDurationSecond  int
	NewCustomerCount    int
	InvitationCount     int
	CallScoreByCount    int
	CallScoreByDuration int
	CallScoreType       string
	CallScore           int
	InvitationScore     int
	NewCustomerScore    int
	TotalScore          int
	ScoreReachedAt      *time.Time
	DataUpdatedAt       *time.Time
}

type RankingLeaderboardItem struct {
	IdentityKey        string  `json:"identityKey" gorm:"column:aggregate_key"`
	Rank               int     `json:"rank" gorm:"-"`
	SeatWorkNumber     string  `json:"seatWorkNumber" gorm:"column:seat_work_number"`
	SeatName           string  `json:"seatName" gorm:"column:seat_name"`
	MatchedUserID      *int64  `json:"matchedUserId,omitempty" gorm:"column:matched_user_id"`
	MatchedUserName    string  `json:"matchedUserName" gorm:"column:matched_user_name"`
	GroupName          string  `json:"groupName" gorm:"column:group_name"`
	RoleName           string  `json:"roleName" gorm:"column:role_name"`
	CallNum            int     `json:"callNum" gorm:"column:call_num"`
	AnsweredCallCount  int     `json:"answeredCallCount" gorm:"column:answered_call_count"`
	AnswerRate         float64 `json:"answerRate" gorm:"column:answer_rate"`
	CallDurationSecond int     `json:"callDurationSecond" gorm:"column:call_duration_second"`
	NewCustomerCount   int     `json:"newCustomerCount" gorm:"column:new_customer_count"`
	InvitationCount    int     `json:"invitationCount" gorm:"column:invitation_count"`
	CallScore          int     `json:"callScore" gorm:"column:call_score"`
	InvitationScore    int     `json:"invitationScore" gorm:"column:invitation_score"`
	NewCustomerScore   int     `json:"newCustomerScore" gorm:"column:new_customer_score"`
	TotalScore         int     `json:"totalScore" gorm:"column:total_score"`
	ScoreDays          int     `json:"scoreDays" gorm:"column:score_days"`
}

type RankingLeaderboardResult struct {
	Period    string                   `json:"period" gorm:"-"`
	StartDate string                   `json:"startDate" gorm:"-"`
	EndDate   string                   `json:"endDate" gorm:"-"`
	Total     int                      `json:"total" gorm:"-"`
	Items     []RankingLeaderboardItem `json:"items" gorm:"-"`
}

type RankingLeaderboardDetail struct {
	Period     string                 `json:"period" gorm:"-"`
	StartDate  string                 `json:"startDate" gorm:"-"`
	EndDate    string                 `json:"endDate" gorm:"-"`
	Rank       int                    `json:"rank" gorm:"-"`
	TotalUsers int                    `json:"totalUsers" gorm:"-"`
	HasData    bool                   `json:"hasData" gorm:"-"`
	Score      RankingLeaderboardItem `json:"score" gorm:"-"`
}

func (SalesDailyScore) TableName() string { return "sales_daily_scores" }
