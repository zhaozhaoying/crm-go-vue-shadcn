package companysearch

import (
	"context"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	emailRegex        = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,10}`)
	chinesePhoneRegex = regexp.MustCompile(`(?:\+?86[\s\-]?)?1[3-9]\d{9}`)
	intlPhoneRegex    = regexp.MustCompile(`\+?\d[\d\s\-\(\)\.]{5,18}\d`)
)

var emailSkipPatterns = []string{
	"example.com", "sentry.io", "domain.com", "test.com", "yourcompany",
	"schema.org", "w3.org", "openxmlformats.org",
}

var emailSkipSuffixes = []string{".png", ".jpg", ".gif", ".svg", ".jpeg", ".webp", ".ico"}

// WebsiteContactExtractor visits a company's public website and extracts
// contact information (email, phone) using HTML parsing and regex fallbacks.
type WebsiteContactExtractor struct {
	client *HTTPClient
}

// NewWebsiteContactExtractor creates a new extractor. If client is nil, a
// default HTTP client is used.
func NewWebsiteContactExtractor(client *HTTPClient) *WebsiteContactExtractor {
	if client == nil {
		client = NewDefaultHTTPClient()
	}
	return &WebsiteContactExtractor{client: client}
}

// Enrich implements Enricher. It fetches the company's website and tries to
// extract an email address and phone number.
func (e *WebsiteContactExtractor) Enrich(ctx context.Context, req EnrichRequest) (*EnrichResult, error) {
	result := &EnrichResult{}
	companyURL := strings.TrimSpace(req.CompanyURL)
	if companyURL == "" {
		return result, nil
	}

	homeBody, err := e.client.Get(ctx, companyURL, RequestOptions{
		Headers: buildEnrichmentHeaders(),
	})
	if err == nil && homeBody != "" {
		e.extractFromBody(homeBody, result)
	}

	// If we still have no email or phone, try the contact page.
	if result.Email == "" && result.Phone == "" {
		if contactURL := findContactURL(homeBody, companyURL); contactURL != "" && contactURL != companyURL {
			if body, cerr := e.client.Get(ctx, contactURL, RequestOptions{Headers: buildEnrichmentHeaders()}); cerr == nil {
				e.extractFromBody(body, result)
			}
		}
	}

	return result, nil
}

// extractFromBody parses HTML and fills any empty fields in result.
func (e *WebsiteContactExtractor) extractFromBody(body string, result *EnrichResult) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err == nil {
		// mailto: links are the most reliable source.
		if result.Email == "" {
			doc.Find(`a[href^="mailto:"]`).EachWithBreak(func(_ int, s *goquery.Selection) bool {
				href, _ := s.Attr("href")
				email := strings.TrimPrefix(href, "mailto:")
				if idx := strings.Index(email, "?"); idx != -1 {
					email = email[:idx]
				}
				email = strings.ToLower(strings.TrimSpace(email))
				if isValidExtractedEmail(email) {
					result.Email = email
					return false
				}
				return true
			})
		}

		// tel: links are the most reliable source for phone numbers.
		if result.Phone == "" {
			doc.Find(`a[href^="tel:"]`).EachWithBreak(func(_ int, s *goquery.Selection) bool {
				href, _ := s.Attr("href")
				phone := strings.TrimPrefix(href, "tel:")
				phone = strings.TrimSpace(phone)
				if phone != "" {
					result.Phone = cleanPhone(phone)
					return false
				}
				return true
			})
		}
	}

	// Regex fallback for email.
	if result.Email == "" {
		for _, email := range emailRegex.FindAllString(body, 40) {
			email = strings.ToLower(strings.TrimSpace(email))
			if isValidExtractedEmail(email) {
				result.Email = email
				break
			}
		}
	}

	// Regex fallback for phone – prefer Chinese mobile numbers.
	if result.Phone == "" {
		if m := chinesePhoneRegex.FindString(body); m != "" {
			result.Phone = cleanPhone(m)
		}
	}
	if result.Phone == "" {
		for _, phone := range intlPhoneRegex.FindAllString(body, 30) {
			digits := digitsOnly(phone)
			if len(digits) >= 7 && len(digits) <= 15 {
				result.Phone = cleanPhone(phone)
				break
			}
		}
	}
}

// findContactURL scans the homepage HTML for a link whose href or text
// suggests it is a "Contact" or "About" page on the same host.
func findContactURL(body, baseURL string) string {
	if body == "" || baseURL == "" {
		return ""
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return ""
	}

	keywords := []string{"contact", "contact-us", "contactus", "reach-us"}
	var found string

	doc.Find("a[href]").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		href, _ := s.Attr("href")
		lower := strings.ToLower(strings.TrimSpace(href))
		linkText := strings.ToLower(strings.TrimSpace(s.Text()))

		for _, kw := range keywords {
			if strings.Contains(lower, kw) || strings.Contains(linkText, kw) {
				abs := NormalizeAbsoluteURL(base.String(), href)
				parsed, perr := url.Parse(abs)
				if perr != nil || parsed.Host != base.Host {
					continue
				}
				found = abs
				return false
			}
		}
		return true
	})

	return found
}

// isValidExtractedEmail returns true if email looks like a real business email.
func isValidExtractedEmail(email string) bool {
	if email == "" || !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return false
	}
	lower := strings.ToLower(email)
	for _, skip := range emailSkipPatterns {
		if strings.Contains(lower, skip) {
			return false
		}
	}
	for _, suffix := range emailSkipSuffixes {
		if strings.HasSuffix(lower, suffix) {
			return false
		}
	}
	return true
}

// cleanPhone removes URL-encoding and trims whitespace from a phone string.
func cleanPhone(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.ReplaceAll(raw, "%20", " ")
	raw = strings.ReplaceAll(raw, "%2B", "+")
	return strings.TrimSpace(raw)
}

// digitsOnly returns only the digit characters in s.
func digitsOnly(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// buildEnrichmentHeaders returns generic browser-like HTTP headers for
// fetching company websites.
func buildEnrichmentHeaders() map[string]string {
	return map[string]string{
		"User-Agent":      defaultUserAgent,
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7",
	}
}
