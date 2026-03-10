package companysearch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"backend/internal/model"
)

type AlibabaProvider struct {
	client  *HTTPClient
	baseURL string
}

func NewAlibabaProvider(client *HTTPClient, baseURL string) *AlibabaProvider {
	if client == nil {
		client = NewDefaultHTTPClient()
	}
	return &AlibabaProvider{client: client, baseURL: strings.TrimSpace(baseURL)}
}

func (p *AlibabaProvider) Platform() int {
	return model.ExternalCompanyPlatformAlibaba
}

func (p *AlibabaProvider) Search(ctx context.Context, req SearchRequest, consume func(SearchPage) error) error {
	if p == nil || p.client == nil {
		return errors.New("alibaba provider is not initialized")
	}
	if strings.TrimSpace(p.baseURL) == "" {
		return errors.New("alibaba search base url is not configured")
	}
	query := strings.TrimSpace(req.RegionKeyword)
	if query == "" {
		query = strings.TrimSpace(req.Keyword)
	}
	if query == "" {
		return errors.New("search keyword is required")
	}

	firstPage, totalPages, err := p.fetchPage(ctx, query, 1)
	if err != nil {
		return err
	}
	pageLimit := totalPages
	if pageLimit <= 0 {
		pageLimit = 1
	}
	if req.PageLimit > 0 && req.PageLimit < pageLimit {
		pageLimit = req.PageLimit
	}

	firstPage.EstimatedTotalPages = totalPages
	firstPage.HasNext = pageLimit > 1
	if err := consume(firstPage); err != nil {
		return err
	}

	for pageNo := 2; pageNo <= pageLimit; pageNo++ {
		page, _, err := p.fetchPage(ctx, query, pageNo)
		if err != nil {
			return err
		}
		page.EstimatedTotalPages = totalPages
		page.HasNext = pageNo < pageLimit
		if err := consume(page); err != nil {
			return err
		}
	}

	return nil
}

func (p *AlibabaProvider) fetchPage(ctx context.Context, query string, pageNo int) (SearchPage, int, error) {
	body, err := p.client.Get(ctx, p.baseURL, RequestOptions{
		Query:   BuildAlibabaParams(query, pageNo),
		Headers: BuildAlibabaHeaders(),
	})
	if err != nil {
		return SearchPage{}, 0, err
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return SearchPage{}, 0, fmt.Errorf("decode alibaba response: %w", err)
	}
	modelObj := mapValue(payload, "model")
	if modelObj == nil {
		return SearchPage{PageNo: pageNo, Items: []FetchedCompany{}}, 0, nil
	}

	totalPages := intValue(mapValue(modelObj, "paginationData"), "totalPage")
	offers := sliceValue(modelObj["offers"])
	items := make([]FetchedCompany, 0, len(offers))
	for idx, raw := range offers {
		offer := mapValueAny(raw)
		if offer == nil {
			continue
		}
		item, err := parseAlibabaOffer(offer, query, pageNo, idx+1)
		if err != nil {
			continue
		}
		items = append(items, item)
	}

	return SearchPage{
		PageNo:       pageNo,
		ResumeCursor: fmt.Sprintf("page:%d", pageNo),
		Items:        items,
	}, totalPages, nil
}

func parseAlibabaOffer(offer map[string]any, query string, pageNo, rankNo int) (FetchedCompany, error) {
	companyID := strings.TrimSpace(JSONString(offer, "companyId"))
	companyName := StripHTML(JSONString(offer, "companyName"))
	companyURL := ProcessURLProtocol(JSONString(offer, "action"))
	if companyID == "" && companyURL == "" && companyName == "" {
		return FetchedCompany{}, errors.New("empty alibaba company payload")
	}

	images := ConcatStringArray(sliceValue(offer["companyImage"]))
	certification := ConcatArrayField(sliceValue(offer["certIconList"]), "image")
	rawPayload := mustCompactJSON(offer)
	isFactory := boolValue(offer, "isFactory")
	businessType := "贸易公司"
	if isFactory {
		businessType = "工厂"
	}

	return FetchedCompany{
		PlatformCompanyID: companyID,
		DedupeKey:         BuildDedupeKey(companyID, companyURL, companyName, JSONString(offer, "city")),
		CompanyName:       companyName,
		CompanyNameEn:     StripHTML(JSONString(offer, "companyTitle")),
		CompanyURL:        companyURL,
		CompanyLogo:       ProcessURLProtocol(JSONString(offer, "companyIcon")),
		CompanyImages:     images,
		City:              JSONString(offer, "city"),
		MainProducts:      JSONArrayString(offer, "mainProducts"),
		BusinessType:      businessType,
		EmployeeCount:     JSONString(offer, "staffNumber"),
		AnnualRevenue:     JSONString(offer, "onlineRevenue"),
		Certification:     certification,
		RawPayload:        rawPayload,
		ResultPayload:     rawPayload,
		PageNo:            pageNo,
		RankNo:            rankNo,
	}, nil
}

func mapValue(obj map[string]any, key string) map[string]any {
	if obj == nil {
		return nil
	}
	return mapValueAny(obj[key])
}

func mapValueAny(value any) map[string]any {
	mapped, _ := value.(map[string]any)
	return mapped
}

func sliceValue(value any) []any {
	items, _ := value.([]any)
	return items
}

func intValue(obj map[string]any, key string) int {
	if obj == nil {
		return 0
	}
	return ParseInt(JSONString(obj, key), 0)
}

func boolValue(obj map[string]any, key string) bool {
	if obj == nil {
		return false
	}
	value, ok := obj[key]
	if !ok || value == nil {
		return false
	}
	flag, _ := value.(bool)
	return flag
}

func mustCompactJSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(data)
}
