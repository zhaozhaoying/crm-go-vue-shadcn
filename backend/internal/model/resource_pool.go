package model

import "time"

type ResourcePoolItem struct {
	ID                  int64      `json:"id"`
	Name                string     `json:"name"`
	Phone               string     `json:"phone"`
	Address             string     `json:"address"`
	Province            string     `json:"province"`
	City                string     `json:"city"`
	Area                string     `json:"area"`
	Latitude            float64    `json:"latitude"`
	Longitude           float64    `json:"longitude"`
	Source              string     `json:"source"`
	SourceUID           string     `json:"sourceUid"`
	SearchKeyword       string     `json:"searchKeyword"`
	SearchRadius        int        `json:"searchRadius"`
	SearchRegion        string     `json:"searchRegion"`
	QueryAddress        string     `json:"queryAddress"`
	CenterLatitude      float64    `json:"centerLatitude"`
	CenterLongitude     float64    `json:"centerLongitude"`
	CreatedBy           int64      `json:"createdBy"`
	Converted           bool       `json:"converted"`
	ConvertedCustomerID *int64     `json:"convertedCustomerId,omitempty"`
	ConvertedAt         *time.Time `json:"convertedAt,omitempty"`
	ConvertedBy         *int64     `json:"convertedBy,omitempty"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}

type ResourcePoolListFilter struct {
	Keyword  string `gorm:"-"`
	HasPhone *bool  `gorm:"-"`
	Page     int    `gorm:"-"`
	PageSize int    `gorm:"-"`
}

type ResourcePoolListResult struct {
	Items    []ResourcePoolItem `json:"items"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"pageSize"`
}

type ResourcePoolSearchInput struct {
	Region          string   `gorm:"-"`
	Address         string   `gorm:"-"`
	Radius          int      `gorm:"-"`
	Keyword         string   `gorm:"-"`
	CenterLatitude  *float64 `gorm:"-"`
	CenterLongitude *float64 `gorm:"-"`
}

type ResourcePoolSearchResult struct {
	CenterLatitude  float64            `json:"centerLatitude"`
	CenterLongitude float64            `json:"centerLongitude"`
	TotalFetched    int                `json:"totalFetched"`
	TotalSaved      int                `json:"totalSaved"`
	Items           []ResourcePoolItem `json:"items"`
}

type ResourcePoolConvertResult struct {
	ResourceID    int64 `json:"resourceId"`
	CustomerID    int64 `json:"customerId"`
	AlreadyLinked bool  `json:"alreadyLinked"`
}

type ResourcePoolBatchConvertItemResult struct {
	ResourceID    int64  `json:"resourceId"`
	CustomerID    int64  `json:"customerId"`
	AlreadyLinked bool   `json:"alreadyLinked"`
	Success       bool   `json:"success"`
	Error         string `json:"error,omitempty"`
}

type ResourcePoolBatchConvertResult struct {
	Total   int                                  `json:"total"`
	Success int                                  `json:"success"`
	Failed  int                                  `json:"failed"`
	Items   []ResourcePoolBatchConvertItemResult `json:"items"`
}

type ResourcePoolItemUpsertInput struct {
	Name            string  `gorm:"-"`
	Phone           string  `gorm:"-"`
	Address         string  `gorm:"-"`
	Province        string  `gorm:"-"`
	City            string  `gorm:"-"`
	Area            string  `gorm:"-"`
	Latitude        float64 `gorm:"-"`
	Longitude       float64 `gorm:"-"`
	Source          string  `gorm:"-"`
	SourceUID       string  `gorm:"-"`
	SearchKeyword   string  `gorm:"-"`
	SearchRadius    int     `gorm:"-"`
	SearchRegion    string  `gorm:"-"`
	QueryAddress    string  `gorm:"-"`
	CenterLatitude  float64 `gorm:"-"`
	CenterLongitude float64 `gorm:"-"`
	CreatedBy       int64   `gorm:"-"`
}
