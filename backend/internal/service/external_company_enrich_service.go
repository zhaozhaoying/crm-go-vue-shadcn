package service

import (
	"backend/internal/external/companysearch"
	"backend/internal/model"
	"backend/internal/repository"
	"context"
	"errors"
	"strings"
	"time"
)

var ErrExternalCompanyNotFound = errors.New("external company not found")

// ExternalCompanyEnrichService enriches a stored external company record with
// additional data fetched from the company's platform profile page and/or its
// public website (email address, phone number, Chinese registered name, …).
type ExternalCompanyEnrichService interface {
	EnrichCompany(ctx context.Context, companyID int64) (*model.ExternalCompany, error)
}

type externalCompanyEnrichService struct {
	repo             repository.ExternalCompanySearchRepository
	alibabaEnricher  *companysearch.AlibabaEnricher
	micEnricher      *companysearch.MadeInChinaEnricher
	websiteExtractor *companysearch.WebsiteContactExtractor
}

// NewExternalCompanyEnrichService constructs the service.
// Any enricher may be nil – the service will simply skip that step.
func NewExternalCompanyEnrichService(
	repo repository.ExternalCompanySearchRepository,
	alibabaEnricher *companysearch.AlibabaEnricher,
	micEnricher *companysearch.MadeInChinaEnricher,
	websiteExtractor *companysearch.WebsiteContactExtractor,
) ExternalCompanyEnrichService {
	return &externalCompanyEnrichService{
		repo:             repo,
		alibabaEnricher:  alibabaEnricher,
		micEnricher:      micEnricher,
		websiteExtractor: websiteExtractor,
	}
}

func (s *externalCompanyEnrichService) EnrichCompany(ctx context.Context, companyID int64) (*model.ExternalCompany, error) {
	company, err := s.repo.GetCompanyByID(ctx, companyID)
	if err != nil {
		return nil, err
	}
	if company == nil {
		return nil, ErrExternalCompanyNotFound
	}

	req := companysearch.EnrichRequest{
		CompanyID:         company.ID,
		Platform:          company.Platform,
		CompanyURL:        company.CompanyURL,
		PlatformCompanyID: company.PlatformCompanyID,
	}

	result := &companysearch.EnrichResult{}

	// 1. Run the platform-specific enricher first – it usually yields better
	//    structured data (Chinese name, contact person, …).
	switch company.Platform {
	case model.ExternalCompanyPlatformAlibaba:
		if s.alibabaEnricher != nil {
			if r, rerr := s.alibabaEnricher.Enrich(ctx, req); rerr == nil && r != nil {
				result.MergeFrom(r)
			}
		}
	case model.ExternalCompanyPlatformMadeInChina:
		if s.micEnricher != nil {
			if r, rerr := s.micEnricher.Enrich(ctx, req); rerr == nil && r != nil {
				result.MergeFrom(r)
			}
		}
	}

	// 2. For Google results (or when the platform enricher found nothing useful),
	//    fall back to scraping the company's own public website.
	if (result.Email == "" || result.Phone == "") &&
		strings.TrimSpace(company.CompanyURL) != "" &&
		s.websiteExtractor != nil {
		if r, rerr := s.websiteExtractor.Enrich(ctx, req); rerr == nil && r != nil {
			result.MergeFrom(r)
		}
	}

	// Nothing at all was found – return the company unchanged without writing.
	if result.IsEmpty() {
		return company, nil
	}

	// 3. Build the update map, touching only fields that are currently empty.
	updates := map[string]any{
		"update_time": time.Now().UTC(),
	}
	changed := false

	// Chinese registered name: promote only when the current name lacks CJK chars.
	if result.ChineseCompanyName != "" && !companysearch.ContainsChinese(company.CompanyName) {
		// Preserve the existing English name in company_name_en if not already set.
		if strings.TrimSpace(company.CompanyNameEn) == "" {
			updates["company_name_en"] = company.CompanyName
		}
		updates["company_name"] = result.ChineseCompanyName
		changed = true
	}

	if result.Contact != "" && strings.TrimSpace(company.Contact) == "" {
		updates["contact"] = result.Contact
		changed = true
	}
	if result.Phone != "" && strings.TrimSpace(company.Phone) == "" {
		updates["phone"] = result.Phone
		changed = true
	}
	if result.Email != "" && strings.TrimSpace(company.Email) == "" {
		updates["email"] = result.Email
		changed = true
	}
	if result.Address != "" && strings.TrimSpace(company.Address) == "" {
		updates["address"] = result.Address
		changed = true
	}

	if !changed {
		return company, nil
	}

	updates["data_version"] = company.DataVersion + 1

	return s.repo.UpdateCompanyFields(ctx, companyID, updates)
}
