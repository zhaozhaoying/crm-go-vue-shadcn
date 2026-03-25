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

func (SalesDailyScore) TableName() string { return "sales_daily_scores" }
