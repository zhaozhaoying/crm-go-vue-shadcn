package companysearch

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// AlibabaEnricher fetches an Alibaba supplier profile page and extracts
// structured company data including Chinese registered name and contact info.
type AlibabaEnricher struct {
	client *HTTPClient
}

// NewAlibabaEnricher creates a new AlibabaEnricher. If client is nil, a
// default HTTP client is used.
func NewAlibabaEnricher(client *HTTPClient) *AlibabaEnricher {
	if client == nil {
		client = NewDefaultHTTPClient()
	}
	return &AlibabaEnricher{client: client}
}

// Enrich implements Enricher. It fetches the Alibaba supplier profile page
// (e.g. https://[id].en.alibaba.com/) and extracts company data.
func (e *AlibabaEnricher) Enrich(ctx context.Context, req EnrichRequest) (*EnrichResult, error) {
	result := &EnrichResult{}
	companyURL := strings.TrimSpace(req.CompanyURL)
	if companyURL == "" {
		return result, nil
	}

	body, err := e.client.Get(ctx, companyURL, RequestOptions{
		Headers: buildAlibabaEnrichHeaders(),
	})
	if err != nil {
		return result, err
	}

	// Try JSON-LD <script type="application/ld+json"> blocks first.
	extractJSONLD(body, result)

	// Parse HTML table rows / definition lists for label-value pairs.
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return result, nil
	}

	doc.Find("tr, .comp-basic-row, .company-info-item, dl").Each(func(_ int, s *goquery.Selection) {
		label := strings.TrimSpace(s.Find("th, dt, .label, .subject").First().Text())
		value := strings.TrimSpace(s.Find("td, dd, .value").Last().Text())
		if label == "" || value == "" {
			return
		}
		applyLabelValue(label, value, result)
	})

	// Also run the generic mailto/tel extractor on the fetched page.
	websiteExtractor := &WebsiteContactExtractor{client: e.client}
	websiteExtractor.extractFromBody(body, result)

	return result, nil
}

// extractJSONLD scans every <script type="application/ld+json"> block and
// populates result from the first Organization / LocalBusiness node found.
func extractJSONLD(body string, result *EnrichResult) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return
	}

	doc.Find(`script[type="application/ld+json"]`).Each(func(_ int, s *goquery.Selection) {
		var data map[string]any
		if err := json.Unmarshal([]byte(s.Text()), &data); err != nil {
			return
		}

		schemaType, _ := data["@type"].(string)
		if !strings.EqualFold(schemaType, "Organization") &&
			!strings.EqualFold(schemaType, "LocalBusiness") &&
			!strings.EqualFold(schemaType, "Corporation") {
			return
		}

		// alternateName is often the native-language (Chinese) registered name.
		if alt, _ := data["alternateName"].(string); alt != "" &&
			ContainsChinese(alt) && result.ChineseCompanyName == "" {
			result.ChineseCompanyName = strings.TrimSpace(alt)
		}

		// legalName may also carry the registered Chinese name.
		if legal, _ := data["legalName"].(string); legal != "" &&
			ContainsChinese(legal) && result.ChineseCompanyName == "" {
			result.ChineseCompanyName = strings.TrimSpace(legal)
		}

		// telephone
		if phone, _ := data["telephone"].(string); phone != "" && result.Phone == "" {
			result.Phone = cleanPhone(phone)
		}

		// email
		if email, _ := data["email"].(string); email != "" && result.Email == "" {
			if isValidExtractedEmail(strings.ToLower(strings.TrimSpace(email))) {
				result.Email = strings.ToLower(strings.TrimSpace(email))
			}
		}

		// address
		if result.Address == "" {
			switch addr := data["address"].(type) {
			case string:
				if strings.TrimSpace(addr) != "" {
					result.Address = strings.TrimSpace(addr)
				}
			case map[string]any:
				parts := []string{
					jsonStr(addr, "streetAddress"),
					jsonStr(addr, "addressLocality"),
					jsonStr(addr, "addressRegion"),
					jsonStr(addr, "addressCountry"),
				}
				var nonEmpty []string
				for _, p := range parts {
					if strings.TrimSpace(p) != "" {
						nonEmpty = append(nonEmpty, strings.TrimSpace(p))
					}
				}
				if len(nonEmpty) > 0 {
					result.Address = strings.Join(nonEmpty, ", ")
				}
			}
		}
	})
}

// applyLabelValue maps common label strings to result fields.
func applyLabelValue(label, value string, result *EnrichResult) {
	ll := strings.ToLower(label)
	switch {
	case strings.Contains(ll, "registered name") ||
		strings.Contains(ll, "legal name") ||
		(strings.Contains(ll, "company name") && ContainsChinese(value)):
		if result.ChineseCompanyName == "" && ContainsChinese(value) {
			result.ChineseCompanyName = value
		}
	case strings.Contains(ll, "contact person") ||
		strings.Contains(ll, "representative") ||
		strings.Contains(ll, "contact name"):
		if result.Contact == "" {
			result.Contact = value
		}
	case (strings.Contains(ll, "phone") || strings.Contains(ll, "tel") ||
		strings.Contains(ll, "mobile")) && result.Phone == "":
		result.Phone = cleanPhone(value)
	case strings.Contains(ll, "email") && result.Email == "":
		lower := strings.ToLower(value)
		if isValidExtractedEmail(lower) {
			result.Email = lower
		}
	case (strings.Contains(ll, "address") || strings.Contains(ll, "location")) &&
		result.Address == "":
		result.Address = value
	}
}

// jsonStr is a small helper to pull a string value from a map[string]any.
func jsonStr(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return v
}

func buildAlibabaEnrichHeaders() map[string]string {
	return map[string]string{
		"User-Agent":      defaultUserAgent,
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "en-US,en;q=0.9,zh-CN;q=0.8",
		"Referer":         "https://www.alibaba.com/",
	}
}
