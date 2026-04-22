package companysearch

import "context"

// EnrichRequest carries the inputs needed to enrich a company record.
type EnrichRequest struct {
	CompanyID         int64
	Platform          int
	CompanyURL        string
	PlatformCompanyID string
}

// EnrichResult holds the data extracted during enrichment.
type EnrichResult struct {
	ChineseCompanyName string // 中文公司名（注册名称）
	Contact            string // 联系人姓名
	Phone              string // 电话号码
	Email              string // 电子邮箱
	Address            string // 详细地址
}

// IsEmpty returns true when no useful data was extracted.
func (r *EnrichResult) IsEmpty() bool {
	if r == nil {
		return true
	}
	return r.ChineseCompanyName == "" && r.Contact == "" &&
		r.Phone == "" && r.Email == "" && r.Address == ""
}

// MergeFrom fills empty fields in r from src.
func (r *EnrichResult) MergeFrom(src *EnrichResult) {
	if src == nil {
		return
	}
	if r.ChineseCompanyName == "" {
		r.ChineseCompanyName = src.ChineseCompanyName
	}
	if r.Contact == "" {
		r.Contact = src.Contact
	}
	if r.Phone == "" {
		r.Phone = src.Phone
	}
	if r.Email == "" {
		r.Email = src.Email
	}
	if r.Address == "" {
		r.Address = src.Address
	}
}

// Enricher can enrich a company record from a specific data source.
type Enricher interface {
	Enrich(ctx context.Context, req EnrichRequest) (*EnrichResult, error)
}

// ContainsChinese reports whether s contains at least one CJK character.
func ContainsChinese(s string) bool {
	for _, r := range s {
		if r >= '\u4e00' && r <= '\u9fff' {
			return true
		}
	}
	return false
}
