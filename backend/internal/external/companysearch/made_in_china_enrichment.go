package companysearch

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// MadeInChinaEnricher fetches a MadeInChina supplier's about/profile page and
// extracts the Chinese registered company name and any visible contact info.
type MadeInChinaEnricher struct {
	client *HTTPClient
}

// NewMadeInChinaEnricher creates a new MadeInChinaEnricher.
// If client is nil a default HTTP client is used.
func NewMadeInChinaEnricher(client *HTTPClient) *MadeInChinaEnricher {
	if client == nil {
		client = NewDefaultHTTPClient()
	}
	return &MadeInChinaEnricher{client: client}
}

// Enrich implements Enricher. It tries the company's /about.html page first,
// then falls back to the main company URL if the about page is unreachable.
func (e *MadeInChinaEnricher) Enrich(ctx context.Context, req EnrichRequest) (*EnrichResult, error) {
	result := &EnrichResult{}

	companyURL := strings.TrimSpace(req.CompanyURL)
	if companyURL == "" {
		return result, nil
	}

	// Prefer the structured about page.
	aboutURL := buildMICAboutURL(companyURL)
	body, err := e.client.Get(ctx, aboutURL, RequestOptions{
		Headers: buildMadeInChinaEnrichHeaders(),
	})
	if err != nil || strings.TrimSpace(body) == "" {
		// Fall back to the company homepage.
		body, err = e.client.Get(ctx, companyURL, RequestOptions{
			Headers: buildMadeInChinaEnrichHeaders(),
		})
		if err != nil {
			return result, fmt.Errorf("made-in-china enrichment fetch failed: %w", err)
		}
	}

	e.parseBody(body, result)

	// Also run JSON-LD extraction (some MIC pages embed ld+json).
	extractJSONLD(body, result)

	// Run the generic mailto/tel link extractor as a final pass.
	websiteExtractor := &WebsiteContactExtractor{client: e.client}
	websiteExtractor.extractFromBody(body, result)

	return result, nil
}

// parseBody walks table rows and definition-list items looking for
// label→value pairs that reveal company registration or contact details.
func (e *MadeInChinaEnricher) parseBody(body string, result *EnrichResult) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return
	}

	// MadeInChina about pages typically use <tr> with <td class="subject"> /
	// <td class="value">, plus various .company-info-* div patterns.
	doc.Find("tr").Each(func(_ int, s *goquery.Selection) {
		label := strings.TrimSpace(s.Find("td.subject, th").First().Text())
		value := strings.TrimSpace(s.Find("td.value, td:not(.subject):not(th)").Last().Text())
		if label == "" || value == "" {
			return
		}
		applyMICLabelValue(label, value, result)
	})

	// Also handle definition-list and div-based layouts.
	doc.Find("dl, .info-item, .company-detail-item").Each(func(_ int, s *goquery.Selection) {
		label := strings.TrimSpace(s.Find("dt, .label, .detail-label").First().Text())
		value := strings.TrimSpace(s.Find("dd, .value, .detail-value").Last().Text())
		if label == "" || value == "" {
			return
		}
		applyMICLabelValue(label, value, result)
	})

	// Some pages expose the registered Chinese name in a dedicated element.
	if result.ChineseCompanyName == "" {
		doc.Find(".company-name-cn, .cn-name, .reg-name").Each(func(_ int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if text != "" && ContainsChinese(text) {
				result.ChineseCompanyName = text
			}
		})
	}
}

// applyMICLabelValue maps MadeInChina-specific label strings to result fields.
func applyMICLabelValue(label, value string, result *EnrichResult) {
	ll := strings.ToLower(strings.TrimSpace(label))

	switch {
	// Chinese registered company name
	case strings.Contains(ll, "registered name") ||
		strings.Contains(ll, "registration name") ||
		strings.Contains(ll, "legal name") ||
		strings.Contains(ll, "company name in chinese") ||
		(strings.Contains(ll, "company name") && ContainsChinese(value)):
		if result.ChineseCompanyName == "" && ContainsChinese(value) {
			result.ChineseCompanyName = strings.TrimSpace(value)
		}

	// Contact / legal representative
	case strings.Contains(ll, "legal representative") ||
		strings.Contains(ll, "contact person") ||
		strings.Contains(ll, "general manager") ||
		strings.Contains(ll, "ceo") ||
		strings.Contains(ll, "owner"):
		if result.Contact == "" {
			result.Contact = strings.TrimSpace(value)
		}

	// Phone / mobile / fax
	case (strings.Contains(ll, "phone") ||
		strings.Contains(ll, "telephone") ||
		strings.Contains(ll, "mobile") ||
		strings.Contains(ll, "tel")) && result.Phone == "":
		result.Phone = cleanPhone(value)

	// Email
	case strings.Contains(ll, "email") && result.Email == "":
		lower := strings.ToLower(strings.TrimSpace(value))
		if isValidExtractedEmail(lower) {
			result.Email = lower
		}

	// Address
	case (strings.Contains(ll, "address") ||
		strings.Contains(ll, "location") ||
		strings.Contains(ll, "factory address")) && result.Address == "":
		result.Address = strings.TrimSpace(value)
	}
}

// buildMICAboutURL constructs the /about.html URL for a made-in-china.com
// supplier. Company URLs have the form https://[slug].en.made-in-china.com/.
func buildMICAboutURL(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return rawURL
	}
	parsed.Path = "/about.html"
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}

func buildMadeInChinaEnrichHeaders() map[string]string {
	return map[string]string{
		"User-Agent":                defaultUserAgent,
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language":           "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7",
		"Cache-Control":             "no-cache",
		"Upgrade-Insecure-Requests": "1",
	}
}
