package model

import "time"

const (
	ExternalCompanyPlatformAlibaba     = 1
	ExternalCompanyPlatformMadeInChina = 2
	ExternalCompanyPlatformGoogle      = 3
)

const (
	ExternalCompanySearchTaskStatusPending   = 1
	ExternalCompanySearchTaskStatusRunning   = 2
	ExternalCompanySearchTaskStatusCompleted = 3
	ExternalCompanySearchTaskStatusFailed    = 4
	ExternalCompanySearchTaskStatusCanceled  = 5
)

const (
	ExternalCompanySearchEventTaskCreated   = "task.created"
	ExternalCompanySearchEventTaskStarted   = "task.started"
	ExternalCompanySearchEventTaskProgress  = "task.progress"
	ExternalCompanySearchEventTaskCompleted = "task.completed"
	ExternalCompanySearchEventTaskFailed    = "task.failed"
	ExternalCompanySearchEventTaskCanceled  = "task.canceled"
	ExternalCompanySearchEventResultSaved   = "result.saved"
)

type ExternalCompanySearchTask struct {
	ID                int64      `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	TaskNo            string     `json:"taskNo" gorm:"column:task_no"`
	Platform          int        `json:"platform" gorm:"column:platform"`
	Keyword           string     `json:"keyword" gorm:"column:keyword"`
	KeywordNormalized string     `json:"keywordNormalized" gorm:"column:keyword_normalized"`
	RegionKeyword     string     `json:"regionKeyword" gorm:"column:region_keyword"`
	Status            int        `json:"status" gorm:"column:status"`
	Priority          int        `json:"priority" gorm:"column:priority"`
	TargetCount       int        `json:"targetCount" gorm:"column:target_count"`
	PageLimit         int        `json:"pageLimit" gorm:"column:page_limit"`
	PageNo            int        `json:"pageNo" gorm:"column:page_no"`
	ProgressPercent   int        `json:"progressPercent" gorm:"column:progress_percent"`
	FetchedCount      int        `json:"fetchedCount" gorm:"column:fetched_count"`
	SavedCount        int        `json:"savedCount" gorm:"column:saved_count"`
	DuplicateCount    int        `json:"duplicateCount" gorm:"column:duplicate_count"`
	FailedCount       int        `json:"failedCount" gorm:"column:failed_count"`
	RetryCount        int        `json:"retryCount" gorm:"column:retry_count"`
	MaxRetryCount     int        `json:"maxRetryCount" gorm:"column:max_retry_count"`
	NextRunAt         *time.Time `json:"nextRunAt,omitempty" gorm:"column:next_run_at"`
	LockedAt          *time.Time `json:"lockedAt,omitempty" gorm:"column:locked_at"`
	LastHeartbeatAt   *time.Time `json:"lastHeartbeatAt,omitempty" gorm:"column:last_heartbeat_at"`
	StartedAt         *time.Time `json:"startedAt,omitempty" gorm:"column:started_at"`
	FinishedAt        *time.Time `json:"finishedAt,omitempty" gorm:"column:finished_at"`
	WorkerToken       string     `json:"workerToken,omitempty" gorm:"column:worker_token"`
	SearchOptions     string     `json:"searchOptions,omitempty" gorm:"column:search_options"`
	ResumeCursor      string     `json:"resumeCursor,omitempty" gorm:"column:resume_cursor"`
	ErrorMessage      string     `json:"errorMessage,omitempty" gorm:"column:error_message"`
	CreatedBy         int64      `json:"createdBy" gorm:"column:created_by"`
	CreatedAt         time.Time  `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt         time.Time  `json:"updatedAt" gorm:"column:updated_at"`
}

type ExternalCompany struct {
	ID                int64      `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	CompanyNo         string     `json:"companyNo" gorm:"column:company_no"`
	Platform          int        `json:"platform" gorm:"column:platform"`
	PlatformCompanyID string     `json:"platformCompanyId" gorm:"column:platform_company_id"`
	DedupeKey         string     `json:"dedupeKey" gorm:"column:dedupe_key"`
	CompanyName       string     `json:"companyName" gorm:"column:company_name"`
	CompanyNameEn     string     `json:"companyNameEn,omitempty" gorm:"column:company_name_en"`
	CompanyURL        string     `json:"companyUrl,omitempty" gorm:"column:company_url"`
	CompanyLogo       string     `json:"companyLogo,omitempty" gorm:"column:company_logo"`
	CompanyImages     string     `json:"companyImages,omitempty" gorm:"column:company_images"`
	CompanyDesc       string     `json:"companyDesc,omitempty" gorm:"column:company_desc"`
	Country           string     `json:"country,omitempty" gorm:"column:country"`
	Province          string     `json:"province,omitempty" gorm:"column:province"`
	City              string     `json:"city,omitempty" gorm:"column:city"`
	Address           string     `json:"address,omitempty" gorm:"column:address"`
	MainProducts      string     `json:"mainProducts,omitempty" gorm:"column:main_products"`
	BusinessType      string     `json:"businessType,omitempty" gorm:"column:business_type"`
	EmployeeCount     string     `json:"employeeCount,omitempty" gorm:"column:employee_count"`
	EstablishedYear   string     `json:"establishedYear,omitempty" gorm:"column:established_year"`
	AnnualRevenue     string     `json:"annualRevenue,omitempty" gorm:"column:annual_revenue"`
	Certification     string     `json:"certification,omitempty" gorm:"column:certification"`
	Contact           string     `json:"contact,omitempty" gorm:"column:contact"`
	Phone             string     `json:"phone,omitempty" gorm:"column:phone"`
	Email             string     `json:"email,omitempty" gorm:"column:email"`
	DataVersion       int        `json:"dataVersion" gorm:"column:data_version"`
	InterestStatus    int        `json:"interestStatus" gorm:"column:interest_status"`
	IsDeleted         bool       `json:"isDeleted" gorm:"column:is_deleted"`
	RawPayload        string     `json:"rawPayload,omitempty" gorm:"column:raw_payload"`
	FirstSeenAt       *time.Time `json:"firstSeenAt,omitempty" gorm:"column:first_seen_at"`
	LastSeenAt        *time.Time `json:"lastSeenAt,omitempty" gorm:"column:last_seen_at"`
	CreateTime        time.Time  `json:"createTime" gorm:"column:create_time"`
	UpdateTime        time.Time  `json:"updateTime" gorm:"column:update_time"`
}

type ExternalCompanySearchResult struct {
	ID            int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	TaskID        int64     `json:"taskId" gorm:"column:task_id"`
	CompanyID     int64     `json:"companyId" gorm:"column:company_id"`
	Platform      int       `json:"platform" gorm:"column:platform"`
	Keyword       string    `json:"keyword" gorm:"column:keyword"`
	RegionKeyword string    `json:"regionKeyword,omitempty" gorm:"column:region_keyword"`
	PageNo        int       `json:"pageNo" gorm:"column:page_no"`
	RankNo        int       `json:"rankNo" gorm:"column:rank_no"`
	IsNewCompany  bool      `json:"isNewCompany" gorm:"column:is_new_company"`
	ResultPayload string    `json:"resultPayload,omitempty" gorm:"column:result_payload"`
	CreatedAt     time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt     time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

type ExternalCompanySearchEvent struct {
	ID        int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	TaskID    int64     `json:"taskId" gorm:"column:task_id"`
	SeqNo     int64     `json:"seqNo" gorm:"column:seq_no"`
	EventType string    `json:"eventType" gorm:"column:event_type"`
	Message   string    `json:"message,omitempty" gorm:"column:message"`
	Payload   string    `json:"payload,omitempty" gorm:"column:payload"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at"`
}

type ExternalCompanySearchTaskCreateInput struct {
	Platforms     []int  `json:"platforms"`
	Keyword       string `json:"keyword"`
	RegionKeyword string `json:"regionKeyword"`
	PageLimit     int    `json:"pageLimit"`
	TargetCount   int    `json:"targetCount"`
	Priority      int    `json:"priority"`
	SearchOptions string `json:"searchOptions"`
	CreatedBy     int64  `json:"createdBy"`
}

type ExternalCompanySearchTaskListFilter struct {
	Platform          int    `json:"platform"`
	Status            int    `json:"status"`
	Keyword           string `json:"keyword"`
	CreatedBy         int64  `json:"createdBy"`
	RestrictToCreator bool   `json:"restrictToCreator"`
	Page              int    `json:"page"`
	PageSize          int    `json:"pageSize"`
}

type ExternalCompanySearchTaskListResult struct {
	Items    []ExternalCompanySearchTask `json:"items"`
	Total    int64                       `json:"total"`
	Page     int                         `json:"page"`
	PageSize int                         `json:"pageSize"`
}

type ExternalCompanySearchResultListFilter struct {
	TaskID            int64  `json:"taskId"`
	Search            string `json:"search"`
	Platform          int    `json:"platform"`
	NewOnly           bool   `json:"newOnly"`
	CreatedBy         int64  `json:"createdBy"`
	RestrictToCreator bool   `json:"restrictToCreator"`
	Page              int    `json:"page"`
	PageSize          int    `json:"pageSize"`
}

type ExternalCompanySearchResultItem struct {
	ID                int64      `json:"id"`
	TaskID            int64      `json:"taskId"`
	CompanyID         int64      `json:"companyId"`
	Platform          int        `json:"platform"`
	Keyword           string     `json:"keyword"`
	RegionKeyword     string     `json:"regionKeyword"`
	PageNo            int        `json:"pageNo"`
	RankNo            int        `json:"rankNo"`
	IsNewCompany      bool       `json:"isNewCompany"`
	ResultPayload     string     `json:"resultPayload,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
	CompanyNo         string     `json:"companyNo"`
	PlatformCompanyID string     `json:"platformCompanyId"`
	DedupeKey         string     `json:"dedupeKey"`
	CompanyName       string     `json:"companyName"`
	CompanyNameEn     string     `json:"companyNameEn,omitempty"`
	CompanyURL        string     `json:"companyUrl,omitempty"`
	CompanyLogo       string     `json:"companyLogo,omitempty"`
	CompanyImages     string     `json:"companyImages,omitempty"`
	CompanyDesc       string     `json:"companyDesc,omitempty"`
	Country           string     `json:"country,omitempty"`
	Province          string     `json:"province,omitempty"`
	City              string     `json:"city,omitempty"`
	Address           string     `json:"address,omitempty"`
	MainProducts      string     `json:"mainProducts,omitempty"`
	BusinessType      string     `json:"businessType,omitempty"`
	EmployeeCount     string     `json:"employeeCount,omitempty"`
	EstablishedYear   string     `json:"establishedYear,omitempty"`
	AnnualRevenue     string     `json:"annualRevenue,omitempty"`
	Certification     string     `json:"certification,omitempty"`
	Contact           string     `json:"contact,omitempty"`
	Phone             string     `json:"phone,omitempty"`
	Email             string     `json:"email,omitempty"`
	DataVersion       int        `json:"dataVersion"`
	InterestStatus    int        `json:"interestStatus"`
	IsDeleted         bool       `json:"isDeleted"`
	RawPayload        string     `json:"rawPayload,omitempty"`
	FirstSeenAt       *time.Time `json:"firstSeenAt,omitempty"`
	LastSeenAt        *time.Time `json:"lastSeenAt,omitempty"`
}

type ExternalCompanySearchResultListResult struct {
	Items    []ExternalCompanySearchResultItem `json:"items"`
	Total    int64                             `json:"total"`
	Page     int                               `json:"page"`
	PageSize int                               `json:"pageSize"`
}

type ExternalCompanySearchEventListResult struct {
	Items   []ExternalCompanySearchEvent `json:"items"`
	NextSeq int64                        `json:"nextSeq"`
}

func (ExternalCompanySearchTask) TableName() string   { return "external_company_search_task" }
func (ExternalCompany) TableName() string             { return "external_company" }
func (ExternalCompanySearchResult) TableName() string { return "external_company_search_result" }
func (ExternalCompanySearchEvent) TableName() string  { return "external_company_search_event" }
