package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	ErrResourcePoolInvalidInput          = errors.New("resource pool invalid input")
	ErrResourcePoolProviderNotConfigured = errors.New("resource pool provider not configured")
	ErrResourcePoolLocationNotFound      = errors.New("resource pool location not found")
	ErrResourcePoolSearchFailed          = errors.New("resource pool search failed")
	ErrResourcePoolItemNotFound          = errors.New("resource pool item not found")
	ErrResourcePoolNoConvertiblePhone    = errors.New("resource pool no convertible phone")
	ErrResourcePoolConvertFailed         = errors.New("resource pool convert failed")
)

var resourcePoolMobileRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)

type ResourcePoolService interface {
	List(ctx context.Context, filter model.ResourcePoolListFilter) (model.ResourcePoolListResult, error)
	SearchAndStore(ctx context.Context, operatorUserID int64, input model.ResourcePoolSearchInput) (model.ResourcePoolSearchResult, error)
	ConvertToCustomer(ctx context.Context, operatorUserID int64, resourceID int64) (model.ResourcePoolConvertResult, error)
	ConvertBatchToCustomer(ctx context.Context, operatorUserID int64, resourceIDs []int64) (model.ResourcePoolBatchConvertResult, error)
}

type resourcePoolService struct {
	repo         repository.ResourcePoolRepository
	customerSvc  CustomerService
	customerRepo repository.CustomerRepository
	provider     resourcePoolSearchProvider
}

type resourcePoolSearchProvider interface {
	SearchNearby(ctx context.Context, input model.ResourcePoolSearchInput) (centerLat float64, centerLng float64, leads []resourceLead, err error)
}

type resourceLead struct {
	Name      string
	Phone     string
	Address   string
	Province  string
	City      string
	Area      string
	Latitude  float64
	Longitude float64
	SourceUID string
}

type baiduSearchProvider struct {
	ak      string
	baseURL string
	client  *http.Client
}

type baiduPlaceSearchResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Results []struct {
		Name      string `json:"name"`
		Address   string `json:"address"`
		Province  string `json:"province"`
		City      string `json:"city"`
		Area      string `json:"area"`
		Telephone string `json:"telephone"`
		UID       string `json:"uid"`
		Location  struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"location"`
	} `json:"results"`
}

func NewResourcePoolService(
	repo repository.ResourcePoolRepository,
	customerSvc CustomerService,
	customerRepo repository.CustomerRepository,
	baiduAK string,
	baiduBaseURL string,
) ResourcePoolService {
	return &resourcePoolService{
		repo:         repo,
		customerSvc:  customerSvc,
		customerRepo: customerRepo,
		provider: &baiduSearchProvider{
			ak:      strings.TrimSpace(baiduAK),
			baseURL: normalizeBaiduBaseURL(baiduBaseURL),
			client: &http.Client{
				Timeout: 8 * time.Second,
			},
		},
	}
}

func (s *resourcePoolService) List(ctx context.Context, filter model.ResourcePoolListFilter) (model.ResourcePoolListResult, error) {
	return s.repo.List(ctx, filter)
}

func (s *resourcePoolService) SearchAndStore(ctx context.Context, operatorUserID int64, input model.ResourcePoolSearchInput) (model.ResourcePoolSearchResult, error) {
	normalized, err := normalizeResourcePoolSearchInput(input)
	if err != nil {
		return model.ResourcePoolSearchResult{}, err
	}

	centerLat, centerLng, leads, err := s.provider.SearchNearby(ctx, normalized)
	if err != nil {
		return model.ResourcePoolSearchResult{}, err
	}

	upsertItems := make([]model.ResourcePoolItemUpsertInput, 0, len(leads))
	for _, lead := range leads {
		upsertItems = append(upsertItems, model.ResourcePoolItemUpsertInput{
			Name:            lead.Name,
			Phone:           lead.Phone,
			Address:         lead.Address,
			Province:        lead.Province,
			City:            lead.City,
			Area:            lead.Area,
			Latitude:        lead.Latitude,
			Longitude:       lead.Longitude,
			Source:          "baidu",
			SourceUID:       buildSourceUID(lead),
			SearchKeyword:   normalized.Keyword,
			SearchRadius:    normalized.Radius,
			SearchRegion:    normalized.Region,
			QueryAddress:    normalized.Address,
			CenterLatitude:  centerLat,
			CenterLongitude: centerLng,
			CreatedBy:       operatorUserID,
		})
	}

	savedItems, err := s.repo.UpsertBatch(ctx, upsertItems)
	if err != nil {
		return model.ResourcePoolSearchResult{}, err
	}

	return model.ResourcePoolSearchResult{
		CenterLatitude:  centerLat,
		CenterLongitude: centerLng,
		TotalFetched:    len(leads),
		TotalSaved:      len(savedItems),
		Items:           savedItems,
	}, nil
}

func (s *resourcePoolService) ConvertToCustomer(
	ctx context.Context,
	operatorUserID int64,
	resourceID int64,
) (model.ResourcePoolConvertResult, error) {
	item, err := s.repo.GetByID(ctx, resourceID)
	if err != nil {
		if errors.Is(err, repository.ErrResourcePoolNotFound) {
			return model.ResourcePoolConvertResult{}, ErrResourcePoolItemNotFound
		}
		return model.ResourcePoolConvertResult{}, err
	}

	if item.Converted {
		customerID := int64(0)
		if item.ConvertedCustomerID != nil {
			customerID = *item.ConvertedCustomerID
		}
		return model.ResourcePoolConvertResult{
			ResourceID:    item.ID,
			CustomerID:    customerID,
			AlreadyLinked: true,
		}, nil
	}

	phone := extractConvertiblePhone(item.Phone)
	if phone == "" {
		return model.ResourcePoolConvertResult{}, ErrResourcePoolNoConvertiblePhone
	}

	name := strings.TrimSpace(item.Name)
	if name == "" {
		name = fmt.Sprintf("地图资源线索#%d", item.ID)
	}
	ownerID := operatorUserID
	createInput := model.CustomerCreateInput{
		Name:           name,
		LegalName:      name,
		ContactName:    "",
		Weixin:         "",
		Email:          "",
		Province:       0,
		City:           0,
		Area:           0,
		DetailAddress:  strings.TrimSpace(item.Address),
		Remark:         buildResourcePoolCustomerRemark(item),
		Status:         model.CustomerStatusOwned,
		OwnerUserID:    &ownerID,
		OperatorUserID: operatorUserID,
		Phones: []model.CustomerPhoneInput{
			{
				Phone:      phone,
				PhoneLabel: "地图资源",
				IsPrimary:  true,
			},
		},
	}

	customer, err := s.customerSvc.CreateCustomer(ctx, createInput)
	if err != nil {
		if errors.Is(err, ErrCustomerNameExists) || errors.Is(err, ErrCustomerLegalExists) {
			retryInput := createInput
			retryInput.Name = fmt.Sprintf("%s-线索%d", name, item.ID)
			retryInput.LegalName = retryInput.Name
			customer, err = s.customerSvc.CreateCustomer(ctx, retryInput)
		}
		if err != nil && errors.Is(err, ErrCustomerPhoneExists) {
			existingID, findErr := s.customerRepo.FindCustomerIDByPhone(ctx, phone)
			if findErr != nil {
				return model.ResourcePoolConvertResult{}, findErr
			}
			if existingID > 0 {
				if err := s.repo.MarkConverted(ctx, item.ID, existingID, operatorUserID); err != nil &&
					!errors.Is(err, repository.ErrResourcePoolAlreadyConverted) {
					if errors.Is(err, repository.ErrResourcePoolNotFound) {
						return model.ResourcePoolConvertResult{}, ErrResourcePoolItemNotFound
					}
					return model.ResourcePoolConvertResult{}, err
				}
				return model.ResourcePoolConvertResult{
					ResourceID:    item.ID,
					CustomerID:    existingID,
					AlreadyLinked: false,
				}, nil
			}
		}
		if err != nil {
			if errors.Is(err, ErrInvalidPhoneFormat) {
				return model.ResourcePoolConvertResult{}, ErrResourcePoolNoConvertiblePhone
			}
			return model.ResourcePoolConvertResult{}, fmt.Errorf("%w: %v", ErrResourcePoolConvertFailed, err)
		}
	}

	if customer == nil || customer.ID == 0 {
		return model.ResourcePoolConvertResult{}, ErrResourcePoolConvertFailed
	}

	if err := s.repo.MarkConverted(ctx, item.ID, customer.ID, operatorUserID); err != nil {
		if errors.Is(err, repository.ErrResourcePoolNotFound) {
			return model.ResourcePoolConvertResult{}, ErrResourcePoolItemNotFound
		}
		if errors.Is(err, repository.ErrResourcePoolAlreadyConverted) {
			latest, latestErr := s.repo.GetByID(ctx, item.ID)
			if latestErr == nil && latest != nil && latest.ConvertedCustomerID != nil {
				return model.ResourcePoolConvertResult{
					ResourceID:    item.ID,
					CustomerID:    *latest.ConvertedCustomerID,
					AlreadyLinked: true,
				}, nil
			}
			return model.ResourcePoolConvertResult{
				ResourceID:    item.ID,
				CustomerID:    customer.ID,
				AlreadyLinked: true,
			}, nil
		}
		return model.ResourcePoolConvertResult{}, err
	}

	return model.ResourcePoolConvertResult{
		ResourceID:    item.ID,
		CustomerID:    customer.ID,
		AlreadyLinked: false,
	}, nil
}

func (s *resourcePoolService) ConvertBatchToCustomer(
	ctx context.Context,
	operatorUserID int64,
	resourceIDs []int64,
) (model.ResourcePoolBatchConvertResult, error) {
	ids := uniquePositiveResourceIDs(resourceIDs)
	if len(ids) == 0 {
		return model.ResourcePoolBatchConvertResult{}, ErrResourcePoolInvalidInput
	}

	result := model.ResourcePoolBatchConvertResult{
		Total: len(ids),
		Items: make([]model.ResourcePoolBatchConvertItemResult, 0, len(ids)),
	}
	for _, id := range ids {
		converted, err := s.ConvertToCustomer(ctx, operatorUserID, id)
		if err != nil {
			result.Failed++
			result.Items = append(result.Items, model.ResourcePoolBatchConvertItemResult{
				ResourceID: id,
				Success:    false,
				Error:      humanizeBatchConvertError(err),
			})
			continue
		}

		result.Success++
		result.Items = append(result.Items, model.ResourcePoolBatchConvertItemResult{
			ResourceID:    converted.ResourceID,
			CustomerID:    converted.CustomerID,
			AlreadyLinked: converted.AlreadyLinked,
			Success:       true,
		})
	}

	return result, nil
}

func normalizeResourcePoolSearchInput(input model.ResourcePoolSearchInput) (model.ResourcePoolSearchInput, error) {
	region := strings.TrimSpace(input.Region)
	address := strings.TrimSpace(input.Address)
	if region == "" && address == "" && (input.CenterLatitude == nil || input.CenterLongitude == nil) {
		return model.ResourcePoolSearchInput{}, ErrResourcePoolInvalidInput
	}

	radius := input.Radius
	if radius <= 0 {
		radius = 3000
	}
	if radius > 50000 {
		radius = 50000
	}

	keyword := strings.TrimSpace(input.Keyword)
	if keyword == "" {
		keyword = "公司"
	}

	return model.ResourcePoolSearchInput{
		Region:          region,
		Address:         address,
		Radius:          radius,
		Keyword:         keyword,
		CenterLatitude:  input.CenterLatitude,
		CenterLongitude: input.CenterLongitude,
	}, nil
}

func normalizeBaiduBaseURL(raw string) string {
	base := strings.TrimSpace(raw)
	if base == "" {
		base = "https://api.map.baidu.com"
	}
	return strings.TrimRight(base, "/")
}

func (p *baiduSearchProvider) SearchNearby(ctx context.Context, input model.ResourcePoolSearchInput) (float64, float64, []resourceLead, error) {
	if p.ak == "" {
		return 0, 0, nil, ErrResourcePoolProviderNotConfigured
	}

	centerLat, centerLng := 0.0, 0.0
	if input.CenterLatitude != nil && input.CenterLongitude != nil {
		centerLat = *input.CenterLatitude
		centerLng = *input.CenterLongitude
	} else {
		lat, lng, err := p.resolveCenterLocation(ctx, input)
		if err != nil {
			return 0, 0, nil, err
		}
		centerLat, centerLng = lat, lng
	}

	results, err := p.searchNearbyByLocation(ctx, input.Keyword, centerLat, centerLng, input.Radius, input.Region)
	if err != nil {
		return 0, 0, nil, err
	}
	leads := make([]resourceLead, 0, len(results))
	for _, item := range results {
		leads = append(leads, resourceLead{
			Name:      strings.TrimSpace(item.Name),
			Phone:     strings.TrimSpace(item.Telephone),
			Address:   strings.TrimSpace(item.Address),
			Province:  strings.TrimSpace(item.Province),
			City:      strings.TrimSpace(item.City),
			Area:      strings.TrimSpace(item.Area),
			Latitude:  item.Location.Lat,
			Longitude: item.Location.Lng,
			SourceUID: strings.TrimSpace(item.UID),
		})
	}

	return centerLat, centerLng, leads, nil
}

func (p *baiduSearchProvider) resolveCenterLocation(ctx context.Context, input model.ResourcePoolSearchInput) (float64, float64, error) {
	query := input.Address
	if query == "" {
		query = "中心"
	}

	params := url.Values{}
	params.Set("query", query)
	if input.Region != "" {
		params.Set("region", input.Region)
	}
	params.Set("page_size", "1")
	params.Set("output", "json")
	params.Set("ak", p.ak)

	endpoint := p.baseURL + "/place/v2/search?" + params.Encode()
	resp, err := p.doBaiduRequest(ctx, endpoint)
	if err != nil {
		return 0, 0, err
	}
	if len(resp.Results) == 0 {
		return 0, 0, ErrResourcePoolLocationNotFound
	}

	return resp.Results[0].Location.Lat, resp.Results[0].Location.Lng, nil
}

func (p *baiduSearchProvider) searchNearbyByLocation(
	ctx context.Context,
	keyword string,
	lat float64,
	lng float64,
	radius int,
	region string,
) ([]struct {
	Name      string `json:"name"`
	Address   string `json:"address"`
	Province  string `json:"province"`
	City      string `json:"city"`
	Area      string `json:"area"`
	Telephone string `json:"telephone"`
	UID       string `json:"uid"`
	Location  struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"location"`
}, error) {
	params := url.Values{}
	params.Set("query", keyword)
	params.Set("location", strconv.FormatFloat(lat, 'f', 6, 64)+","+strconv.FormatFloat(lng, 'f', 6, 64))
	params.Set("radius", strconv.Itoa(radius))
	params.Set("scope", "2")
	params.Set("page_size", "20")
	if region != "" {
		params.Set("region", region)
	}
	params.Set("output", "json")
	params.Set("ak", p.ak)

	endpoint := p.baseURL + "/place/v2/search?" + params.Encode()
	resp, err := p.doBaiduRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	return resp.Results, nil
}

func (p *baiduSearchProvider) doBaiduRequest(ctx context.Context, requestURL string) (*baiduPlaceSearchResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}
	httpResp, err := p.client.Do(req)
	if err != nil {
		return nil, ErrResourcePoolSearchFailed
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, ErrResourcePoolSearchFailed
	}
	if httpResp.StatusCode < 200 || httpResp.StatusCode > 299 {
		return nil, ErrResourcePoolSearchFailed
	}

	var parsed baiduPlaceSearchResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, ErrResourcePoolSearchFailed
	}
	if parsed.Status != 0 {
		return nil, fmt.Errorf("%w: baidu status=%d message=%s", ErrResourcePoolSearchFailed, parsed.Status, parsed.Message)
	}
	return &parsed, nil
}

func buildSourceUID(lead resourceLead) string {
	uid := strings.TrimSpace(lead.SourceUID)
	if uid != "" {
		return uid
	}

	raw := strings.Join([]string{
		strings.TrimSpace(lead.Name),
		strings.TrimSpace(lead.Address),
		strings.TrimSpace(lead.Phone),
		strconv.FormatFloat(lead.Latitude, 'f', 6, 64),
		strconv.FormatFloat(lead.Longitude, 'f', 6, 64),
	}, "|")
	sum := sha1.Sum([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func extractConvertiblePhone(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}

	splitter := func(r rune) bool {
		switch r {
		case ',', '，', ';', '；', '/', '\\', '|', '、', ' ':
			return true
		default:
			return false
		}
	}
	parts := strings.FieldsFunc(raw, splitter)
	for _, part := range parts {
		candidate := strings.TrimSpace(part)
		candidate = strings.ReplaceAll(candidate, "-", "")
		candidate = strings.ReplaceAll(candidate, "(", "")
		candidate = strings.ReplaceAll(candidate, ")", "")
		candidate = strings.ReplaceAll(candidate, "（", "")
		candidate = strings.ReplaceAll(candidate, "）", "")
		if resourcePoolMobileRegex.MatchString(candidate) {
			return candidate
		}
	}
	return ""
}

func buildResourcePoolCustomerRemark(item *model.ResourcePoolItem) string {
	parts := []string{
		"来源: 地图资源",
		"渠道: 百度地图",
		fmt.Sprintf("资源ID: %d", item.ID),
	}
	if keyword := strings.TrimSpace(item.SearchKeyword); keyword != "" {
		parts = append(parts, fmt.Sprintf("关键词: %s", keyword))
	}
	if region := strings.TrimSpace(item.SearchRegion); region != "" {
		parts = append(parts, fmt.Sprintf("区域: %s", region))
	}
	if addr := strings.TrimSpace(item.QueryAddress); addr != "" {
		parts = append(parts, fmt.Sprintf("查询地址: %s", addr))
	}
	return strings.Join(parts, " | ")
}

func uniquePositiveResourceIDs(ids []int64) []int64 {
	if len(ids) == 0 {
		return []int64{}
	}

	seen := make(map[int64]struct{}, len(ids))
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func humanizeBatchConvertError(err error) string {
	switch {
	case errors.Is(err, ErrResourcePoolItemNotFound):
		return "地图资源不存在"
	case errors.Is(err, ErrResourcePoolNoConvertiblePhone):
		return "电话不可用于创建客户"
	case errors.Is(err, ErrResourcePoolConvertFailed):
		return "客户信息冲突，转客户失败"
	default:
		return "转客户失败"
	}
}
