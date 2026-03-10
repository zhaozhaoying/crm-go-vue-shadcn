package companysearch

import "context"

type SearchRequest struct {
	TaskID        int64
	Keyword       string
	RegionKeyword string
	PageLimit     int
	TargetCount   int
	SearchOptions string
}

type FetchedCompany struct {
	PlatformCompanyID string
	DedupeKey         string
	CompanyName       string
	CompanyNameEn     string
	CompanyURL        string
	CompanyLogo       string
	CompanyImages     string
	CompanyDesc       string
	Country           string
	Province          string
	City              string
	Address           string
	MainProducts      string
	BusinessType      string
	EmployeeCount     string
	EstablishedYear   string
	AnnualRevenue     string
	Certification     string
	Contact           string
	Phone             string
	Email             string
	RawPayload        string
	ResultPayload     string
	PageNo            int
	RankNo            int
}

type SearchPage struct {
	PageNo              int
	EstimatedTotalPages int
	ResumeCursor        string
	HasNext             bool
	Items               []FetchedCompany
}

type Provider interface {
	Platform() int
	Search(ctx context.Context, req SearchRequest, consume func(SearchPage) error) error
}
