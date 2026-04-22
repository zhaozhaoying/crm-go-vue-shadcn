package model

import "time"

const (
	ContractCooperationTypeDomestic = "domestic"
	ContractCooperationTypeForeign  = "foreign"
)

const (
	ContractPaymentStatusPending = "pending"
	ContractPaymentStatusPaid    = "paid"
	ContractPaymentStatusPartial = "partial"
)

const (
	ContractAuditStatusPending = "pending"
	ContractAuditStatusSuccess = "success"
	ContractAuditStatusFailed  = "failed"
)

const (
	ContractExpiryHandlingStatusPending = "pending"
	ContractExpiryHandlingStatusRenewed = "renewed"
	ContractExpiryHandlingStatusEnded   = "ended"
)

type Contract struct {
	ID                   int64      `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ContractImage        string     `json:"contractImage" gorm:"column:contract_image;not null;default:''"`
	PaymentImage         string     `json:"paymentImage" gorm:"column:payment_image;not null;default:''"`
	PaymentStatus        string     `json:"paymentStatus" gorm:"column:payment_status;not null;default:'pending';index"`
	Remark               string     `json:"remark" gorm:"column:remark;type:text"`
	UserID               int64      `json:"userId" gorm:"column:user_id;not null;index"`
	CustomerID           int64      `json:"customerId" gorm:"column:customer_id;not null;index"`
	CooperationType      string     `json:"cooperationType" gorm:"column:cooperation_type;not null;default:'domestic';index"`
	ContractNumber       string     `json:"contractNumber" gorm:"column:contract_number;not null;uniqueIndex"`
	ContractName         string     `json:"contractName" gorm:"column:contract_name;not null"`
	ContractAmount       float64    `json:"contractAmount" gorm:"column:contract_amount;not null;default:0"`
	PaymentAmount        float64    `json:"paymentAmount" gorm:"column:payment_amount;not null;default:0"`
	CooperationYears     int        `json:"cooperationYears" gorm:"column:cooperation_years;not null;default:0"`
	NodeCount            int        `json:"nodeCount" gorm:"column:node_count;not null;default:0"`
	ServiceUserID        *int64     `json:"serviceUserId" gorm:"column:service_user_id;index"`
	WebsiteName          string     `json:"websiteName" gorm:"column:website_name;not null;default:''"`
	WebsiteURL           string     `json:"websiteUrl" gorm:"column:website_url;not null;default:''"`
	WebsiteUsername      string     `json:"websiteUsername" gorm:"column:website_username;not null;default:''"`
	IsOnline             bool       `json:"isOnline" gorm:"column:is_online;not null;default:0"`
	StartDateUnix        *int64     `json:"-" gorm:"column:start_date"`
	EndDateUnix          *int64     `json:"-" gorm:"column:end_date"`
	StartDate            *time.Time `json:"startDate,omitempty" gorm:"-"`
	EndDate              *time.Time `json:"endDate,omitempty" gorm:"-"`
	AuditStatus          string     `json:"auditStatus" gorm:"column:audit_status;not null;default:'pending';index"`
	AuditComment         string     `json:"auditComment,omitempty" gorm:"column:audit_comment;type:text"`
	AuditedBy            *int64     `json:"auditedBy,omitempty" gorm:"column:audited_by"`
	AuditedAt            *time.Time `json:"auditedAt,omitempty" gorm:"column:audited_at"`
	ExpiryHandlingStatus string     `json:"expiryHandlingStatus" gorm:"column:expiry_handling_status;not null;default:'pending';index"`
	UserName             string     `json:"userName,omitempty" gorm:"-"`
	CustomerName         string     `json:"customerName,omitempty" gorm:"-"`
	ServiceUserName      string     `json:"serviceUserName,omitempty" gorm:"-"`
	AuditedByName        string     `json:"auditedByName,omitempty" gorm:"-"`
	CreatedAt            time.Time  `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt            time.Time  `json:"updatedAt" gorm:"column:updated_at;autoUpdateTime"`
}

type ContractListFilter struct {
	Keyword               string  `gorm:"-"`
	PaymentStatus         string  `gorm:"-"`
	CooperationType       string  `gorm:"-"`
	AuditStatus           string  `gorm:"-"`
	ExpiryHandlingStatus  string  `gorm:"-"`
	UserID                int64   `gorm:"-"`
	CustomerID            int64   `gorm:"-"`
	ActorUserID           int64   `gorm:"-"`
	ActorRole             string  `gorm:"-"`
	AllowedUserIDs        []int64 `gorm:"-"`
	AllowedServiceUserIDs []int64 `gorm:"-"`
	ForceServiceUserID    *int64  `gorm:"-"`
	Page                  int     `gorm:"-"`
	PageSize              int     `gorm:"-"`
}

type ContractListResult struct {
	Items    []Contract `json:"items" gorm:"-"`
	Total    int64      `json:"total" gorm:"-"`
	Page     int        `json:"page" gorm:"-"`
	PageSize int        `json:"pageSize" gorm:"-"`
}

type ContractCreateInput struct {
	ContractImage        string     `gorm:"-"`
	PaymentImage         string     `gorm:"-"`
	PaymentStatus        string     `gorm:"-"`
	Remark               string     `gorm:"-"`
	UserID               int64      `gorm:"-"`
	CustomerID           int64      `gorm:"-"`
	CooperationType      string     `gorm:"-"`
	ContractNumber       string     `gorm:"-"`
	ContractName         string     `gorm:"-"`
	ContractAmount       float64    `gorm:"-"`
	PaymentAmount        float64    `gorm:"-"`
	CooperationYears     int        `gorm:"-"`
	NodeCount            int        `gorm:"-"`
	ServiceUserID        *int64     `gorm:"-"`
	WebsiteName          string     `gorm:"-"`
	WebsiteURL           string     `gorm:"-"`
	WebsiteUsername      string     `gorm:"-"`
	IsOnline             bool       `gorm:"-"`
	StartDate            *int64     `gorm:"-"`
	EndDate              *int64     `gorm:"-"`
	AuditStatus          string     `gorm:"-"`
	AuditComment         string     `gorm:"-"`
	AuditedBy            *int64     `gorm:"-"`
	AuditedAt            *time.Time `gorm:"-"`
	ExpiryHandlingStatus string     `gorm:"-"`
}

type ContractUpdateInput struct {
	ContractImage        string     `gorm:"-"`
	PaymentImage         string     `gorm:"-"`
	PaymentStatus        string     `gorm:"-"`
	Remark               string     `gorm:"-"`
	UserID               int64      `gorm:"-"`
	CustomerID           int64      `gorm:"-"`
	CooperationType      string     `gorm:"-"`
	ContractNumber       string     `gorm:"-"`
	ContractName         string     `gorm:"-"`
	ContractAmount       float64    `gorm:"-"`
	PaymentAmount        float64    `gorm:"-"`
	CooperationYears     int        `gorm:"-"`
	NodeCount            int        `gorm:"-"`
	ServiceUserID        *int64     `gorm:"-"`
	WebsiteName          string     `gorm:"-"`
	WebsiteURL           string     `gorm:"-"`
	WebsiteUsername      string     `gorm:"-"`
	IsOnline             bool       `gorm:"-"`
	StartDate            *int64     `gorm:"-"`
	EndDate              *int64     `gorm:"-"`
	AuditStatus          string     `gorm:"-"`
	AuditComment         string     `gorm:"-"`
	AuditedBy            *int64     `gorm:"-"`
	AuditedAt            *time.Time `gorm:"-"`
	ExpiryHandlingStatus string     `gorm:"-"`
}

func (Contract) TableName() string { return "contracts" }
