package companysearch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	"backend/internal/model"
)

type GoogleProvider struct {
	client *HTTPClient
	apiKey string
	cx     string
	num    int
}

type GoogleSearchResponse struct {
	Items   []GoogleSearchItem `json:"items"`
	Queries struct {
		NextPage []struct {
			StartIndex int `json:"startIndex"`
		} `json:"nextPage"`
	} `json:"queries"`
	SearchInformation struct {
		TotalResults string `json:"totalResults"`
	} `json:"searchInformation"`
}

type GoogleSearchItem struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Snippet     string `json:"snippet"`
	DisplayLink string `json:"displayLink"`
	Pagemap     struct {
		Metatags []map[string]string `json:"metatags"`
	} `json:"pagemap"`
}

func NewGoogleProvider(client *HTTPClient, apiKey, cx string, num int) *GoogleProvider {
	if client == nil {
		client = NewDefaultHTTPClient()
	}
	if num <= 0 || num > 10 {
		num = 10
	}
	return &GoogleProvider{
		client: client,
		apiKey: strings.TrimSpace(apiKey),
		cx:     strings.TrimSpace(cx),
		num:    num,
	}
}

func (p *GoogleProvider) Platform() int {
	return model.ExternalCompanyPlatformGoogle
}

func (p *GoogleProvider) Search(ctx context.Context, req SearchRequest, consume func(SearchPage) error) error {
	if p == nil || p.client == nil {
		return errors.New("google provider is not initialized")
	}
	if p.apiKey == "" {
		return errors.New("google api key is not configured")
	}
	if p.cx == "" {
		return errors.New("google cx is not configured")
	}

	query := strings.TrimSpace(req.RegionKeyword)
	if query == "" {
		query = strings.TrimSpace(req.Keyword)
	}
	if query == "" {
		return errors.New("search keyword is required")
	}

	start := 1
	pageNo := 1
	pageLimit := req.PageLimit
	if pageLimit <= 0 {
		pageLimit = 100 // 默认最多100页
	}

	for {
		if pageNo > pageLimit {
			break
		}

		page, hasNext, nextStart, err := p.fetchPage(ctx, query, start, pageNo)
		if err != nil {
			return err
		}

		page.HasNext = hasNext && pageNo < pageLimit
		if err := consume(page); err != nil {
			return err
		}

		if !hasNext || len(page.Items) == 0 {
			break
		}

		start = nextStart
		pageNo++
	}

	return nil
}

func (p *GoogleProvider) fetchPage(ctx context.Context, query string, start, pageNo int) (SearchPage, bool, int, error) {
	searchURL := p.buildSearchURL(query, start)

	body, err := p.client.Get(ctx, searchURL, RequestOptions{
		Headers: buildGoogleHeaders(),
	})
	if err != nil {
		return SearchPage{}, false, 0, explainGoogleRequestError(err)
	}

	var response GoogleSearchResponse
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		return SearchPage{}, false, 0, fmt.Errorf("decode google response: %w", err)
	}

	if len(response.Items) == 0 {
		return SearchPage{PageNo: pageNo, Items: []FetchedCompany{}}, false, 0, nil
	}

	items := make([]FetchedCompany, 0, len(response.Items))
	for idx, item := range response.Items {
		company := parseGoogleSearchItem(item, query, pageNo, idx+1)
		items = append(items, company)
	}

	hasNext := len(response.Queries.NextPage) > 0
	nextStart := start + p.num
	if hasNext && len(response.Queries.NextPage) > 0 {
		nextStart = response.Queries.NextPage[0].StartIndex
	}

	return SearchPage{
		PageNo:       pageNo,
		ResumeCursor: fmt.Sprintf("start:%d", nextStart),
		Items:        items,
	}, hasNext, nextStart, nil
}

func (p *GoogleProvider) buildSearchURL(query string, start int) string {
	values := url.Values{}
	values.Set("key", p.apiKey)
	values.Set("cx", p.cx)
	values.Set("q", query)
	values.Set("start", fmt.Sprintf("%d", start))
	values.Set("num", fmt.Sprintf("%d", p.num))
	return "https://www.googleapis.com/customsearch/v1?" + values.Encode()
}

func parseGoogleSearchItem(item GoogleSearchItem, query string, pageNo, rankNo int) FetchedCompany {
	companyName := StripHTML(item.Title)
	companyURL := item.Link
	companyDesc := StripHTML(item.Snippet)
	displayLink := item.DisplayLink

	// 尝试从 metatags 中提取更多信息
	var metaDesc, metaKeywords string
	if len(item.Pagemap.Metatags) > 0 {
		meta := item.Pagemap.Metatags[0]
		metaDesc = meta["og:description"]
		if metaDesc == "" {
			metaDesc = meta["description"]
		}
		metaKeywords = meta["keywords"]
	}

	if metaDesc != "" {
		companyDesc = metaDesc
	}

	// 构建原始数据
	rawPayload, _ := json.Marshal(item)
	resultPayload := map[string]any{
		"query":       query,
		"title":       companyName,
		"url":         companyURL,
		"snippet":     companyDesc,
		"displayLink": displayLink,
		"keywords":    metaKeywords,
	}
	resultPayloadJSON, _ := json.Marshal(resultPayload)

	// 使用域名作为平台公司ID
	platformCompanyID := extractDomainFromURL(companyURL)
	dedupeKey := BuildDedupeKey(platformCompanyID, companyURL, companyName, "")

	return FetchedCompany{
		PlatformCompanyID: platformCompanyID,
		DedupeKey:         dedupeKey,
		CompanyName:       companyName,
		CompanyURL:        companyURL,
		CompanyDesc:       companyDesc,
		MainProducts:      metaKeywords,
		RawPayload:        string(rawPayload),
		ResultPayload:     string(resultPayloadJSON),
		PageNo:            pageNo,
		RankNo:            rankNo,
	}
}

func buildGoogleHeaders() map[string]string {
	return map[string]string{
		"User-Agent": defaultUserAgent,
		"Accept":     "application/json",
	}
}

func extractDomainFromURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(parsed.Host, "www.")
}

func explainGoogleRequestError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) {
		return err
	}

	var statusErr *HTTPStatusError
	if errors.As(err, &statusErr) {
		return fmt.Errorf("google custom search request failed: %w", err)
	}

	if isLikelyGoogleConnectivityError(err) {
		return fmt.Errorf("google custom search is unreachable from the current server; configure GOOGLE_PROXY_URL or HTTPS_PROXY if Google access requires a proxy: %w", err)
	}

	return fmt.Errorf("google custom search request failed: %w", err)
}

func isLikelyGoogleConnectivityError(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	message := strings.ToLower(err.Error())
	patterns := []string{
		"no such host",
		"timeout",
		"connection refused",
		"connection reset",
		"network is unreachable",
		"proxyconnect",
		"tls handshake timeout",
	}
	for _, pattern := range patterns {
		if strings.Contains(message, pattern) {
			return true
		}
	}
	return false
}
