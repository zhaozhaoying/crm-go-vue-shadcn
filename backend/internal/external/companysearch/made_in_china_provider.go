package companysearch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"backend/internal/model"

	"github.com/PuerkitoBio/goquery"
)

type MadeInChinaProvider struct {
	client  *HTTPClient
	baseURL string
}

func NewMadeInChinaProvider(client *HTTPClient, baseURL string) *MadeInChinaProvider {
	if client == nil {
		client = NewDefaultHTTPClient()
	}
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = "https://www.made-in-china.com"
	}
	return &MadeInChinaProvider{client: client, baseURL: strings.TrimRight(baseURL, "/")}
}

func (p *MadeInChinaProvider) Platform() int {
	return model.ExternalCompanyPlatformMadeInChina
}

func (p *MadeInChinaProvider) Search(ctx context.Context, req SearchRequest, consume func(SearchPage) error) error {
	if p == nil || p.client == nil {
		return errors.New("made-in-china provider is not initialized")
	}
	query := strings.TrimSpace(req.RegionKeyword)
	if query == "" {
		query = strings.TrimSpace(req.Keyword)
	}
	if query == "" {
		return errors.New("search keyword is required")
	}

	currentURL := p.buildSearchURL(query)
	pageLimit := req.PageLimit
	for pageNo := 1; currentURL != ""; pageNo++ {
		if pageLimit > 0 && pageNo > pageLimit {
			break
		}
		htmlBody, err := p.client.Get(ctx, currentURL, RequestOptions{Headers: buildMadeInChinaHeaders()})
		if err != nil {
			return err
		}
		page, nextURL, err := p.parsePage(htmlBody, currentURL, query, pageNo)
		if err != nil {
			return err
		}
		page.HasNext = nextURL != ""
		page.ResumeCursor = nextURL
		if err := consume(page); err != nil {
			return err
		}
		currentURL = nextURL
	}

	return nil
}

func (p *MadeInChinaProvider) buildSearchURL(query string) string {
	values := url.Values{}
	values.Set("subaction", "hunt")
	values.Set("style", "b")
	values.Set("mode", "and")
	values.Set("code", "0")
	values.Set("comProvince", "nolimit")
	values.Set("order", "0")
	values.Set("isOpenCorrection", "1")
	values.Set("org", "top")
	values.Set("keyword", "")
	values.Set("file", "")
	values.Set("searchType", "1")
	values.Set("word", query)
	return p.baseURL + "/companysearch.do?" + values.Encode()
}

func (p *MadeInChinaProvider) parsePage(htmlBody, currentURL, query string, pageNo int) (SearchPage, string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlBody))
	if err != nil {
		return SearchPage{}, "", fmt.Errorf("parse made-in-china html: %w", err)
	}

	items := make([]FetchedCompany, 0, 20)
	doc.Find(`div[faw-module="suppliers_list"]`).Each(func(index int, sel *goquery.Selection) {
		item, parseErr := parseMadeInChinaCompany(sel, p.baseURL, query, pageNo, index+1)
		if parseErr != nil {
			return
		}
		items = append(items, item)
	})

	nextURL := extractMadeInChinaNextPage(doc, currentURL, p.baseURL)
	return SearchPage{PageNo: pageNo, Items: items}, nextURL, nil
}

func parseMadeInChinaCompany(sel *goquery.Selection, baseURL, query string, pageNo, rankNo int) (FetchedCompany, error) {
	companyName := strings.TrimSpace(sel.Find(".company-name").First().Text())
	companyURL, _ := sel.Find(".company-name a").First().Attr("href")
	companyURL = NormalizeAbsoluteURL(baseURL, companyURL)
	if companyName == "" && companyURL == "" {
		return FetchedCompany{}, errors.New("empty made-in-china company payload")
	}

	images := make([]string, 0, 4)
	logo := ""
	sel.Find("ul.rec-product img").Each(func(_ int, img *goquery.Selection) {
		candidate, _ := img.Attr("data-original")
		if strings.TrimSpace(candidate) == "" {
			candidate, _ = img.Attr("data-src")
		}
		if strings.TrimSpace(candidate) == "" {
			candidate, _ = img.Attr("src")
		}
		candidate = NormalizeAbsoluteURL(baseURL, candidate)
		if strings.TrimSpace(candidate) == "" {
			return
		}
		if logo == "" {
			logo = candidate
		}
		images = append(images, candidate)
	})

	mainProducts := ""
	city := ""
	sel.Find("tr").Each(func(_ int, tr *goquery.Selection) {
		subject := strings.TrimSpace(tr.Find("td.subject").First().Text())
		value := strings.TrimSpace(tr.Find("td").Eq(1).Text())
		switch subject {
		case "Main Products:":
			mainProducts = buildMadeInChinaProducts(value)
		case "City/Province:":
			city = value
		}
	})

	outerHTML, _ := goquery.OuterHtml(sel)
	platformCompanyID := ExtractMadeInChinaCompanyID(companyURL)
	payload := map[string]any{
		"query":       query,
		"companyName": companyName,
		"companyUrl":  companyURL,
		"city":        city,
	}
	resultPayload, _ := json.Marshal(payload)

	return FetchedCompany{
		PlatformCompanyID: platformCompanyID,
		DedupeKey:         BuildDedupeKey(platformCompanyID, companyURL, companyName, city),
		CompanyName:       companyName,
		CompanyURL:        companyURL,
		CompanyLogo:       logo,
		CompanyImages:     strings.Join(images, ","),
		City:              city,
		MainProducts:      mainProducts,
		RawPayload:        outerHTML,
		ResultPayload:     string(resultPayload),
		PageNo:            pageNo,
		RankNo:            rankNo,
	}, nil
}

func extractMadeInChinaNextPage(doc *goquery.Document, currentURL, baseURL string) string {
	selectors := []string{"a.next", "div.pager a"}
	for _, selector := range selectors {
		found := ""
		doc.Find(selector).EachWithBreak(func(_ int, sel *goquery.Selection) bool {
			label := strings.TrimSpace(sel.Text())
			if selector == "div.pager a" && !strings.EqualFold(label, "Next") {
				return true
			}
			href, ok := sel.Attr("href")
			if !ok || strings.TrimSpace(href) == "" {
				return true
			}
			found = NormalizeAbsoluteURL(baseURL, href)
			return false
		})
		if found != "" && found != currentURL {
			return found
		}
	}
	return ""
}

func buildMadeInChinaHeaders() map[string]string {
	return map[string]string{
		"User-Agent":                defaultUserAgent,
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
		"Accept-Language":           "en-US,en;q=0.9",
		"Cache-Control":             "max-age=0",
		"Connection":                "keep-alive",
		"Upgrade-Insecure-Requests": "1",
		"Sec-Fetch-Dest":            "document",
		"Sec-Fetch-Mode":            "navigate",
		"Sec-Fetch-Site":            "none",
		"Sec-Fetch-User":            "?1",
	}
}

func buildMadeInChinaProducts(raw string) string {
	parts := strings.Split(raw, ",")
	items := make([]map[string]any, 0, len(parts))
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		items = append(items, map[string]any{
			"name":  name,
			"count": nil,
		})
	}
	data, err := json.Marshal(items)
	if err != nil {
		return ""
	}
	return string(data)
}
