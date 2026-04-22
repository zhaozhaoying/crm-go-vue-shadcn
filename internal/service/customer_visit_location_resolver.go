package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var ErrCustomerVisitLocationResolveFailed = errors.New("customer visit location resolve failed")

type CustomerVisitResolvedLocation struct {
	Province      string
	City          string
	Area          string
	DetailAddress string
}

type CustomerVisitLocationResolver interface {
	Resolve(ctx context.Context, lat, lng float64) (CustomerVisitResolvedLocation, error)
}

type nominatimReverseGeocoder struct {
	client    *http.Client
	baseURL   string
	userAgent string
}

type nominatimReverseResponse struct {
	DisplayName string `json:"display_name"`
	Address     struct {
		Province      string `json:"province"`
		State         string `json:"state"`
		Region        string `json:"region"`
		StateDistrict string `json:"state_district"`
		City          string `json:"city"`
		Town          string `json:"town"`
		Municipality  string `json:"municipality"`
		County        string `json:"county"`
		District      string `json:"district"`
		CityDistrict  string `json:"city_district"`
		Suburb        string `json:"suburb"`
		Borough       string `json:"borough"`
		Road          string `json:"road"`
		HouseNumber   string `json:"house_number"`
		Neighbourhood string `json:"neighbourhood"`
		Quarter       string `json:"quarter"`
		Village       string `json:"village"`
		Hamlet        string `json:"hamlet"`
		Amenity       string `json:"amenity"`
		Building      string `json:"building"`
	} `json:"address"`
}

func NewNominatimReverseGeocoder(baseURL, userAgent string) CustomerVisitLocationResolver {
	trimmedBaseURL := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if trimmedBaseURL == "" {
		trimmedBaseURL = "https://nominatim.openstreetmap.org"
	}
	trimmedUserAgent := strings.TrimSpace(userAgent)
	if trimmedUserAgent == "" {
		trimmedUserAgent = "crm-go-vue-shadcn/1.0"
	}

	return &nominatimReverseGeocoder{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL:   trimmedBaseURL,
		userAgent: trimmedUserAgent,
	}
}

func (g *nominatimReverseGeocoder) Resolve(
	ctx context.Context,
	lat, lng float64,
) (CustomerVisitResolvedLocation, error) {
	params := url.Values{}
	params.Set("format", "json")
	params.Set("lat", strconv.FormatFloat(lat, 'f', -1, 64))
	params.Set("lon", strconv.FormatFloat(lng, 'f', -1, 64))

	requestURL := g.baseURL + "/reverse?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return CustomerVisitResolvedLocation{}, fmt.Errorf("%w: %v", ErrCustomerVisitLocationResolveFailed, err)
	}
	req.Header.Set("User-Agent", g.userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")

	resp, err := g.client.Do(req)
	if err != nil {
		return CustomerVisitResolvedLocation{}, fmt.Errorf("%w: %v", ErrCustomerVisitLocationResolveFailed, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CustomerVisitResolvedLocation{}, fmt.Errorf("%w: %v", ErrCustomerVisitLocationResolveFailed, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return CustomerVisitResolvedLocation{}, fmt.Errorf("%w: status=%d body=%s", ErrCustomerVisitLocationResolveFailed, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed nominatimReverseResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return CustomerVisitResolvedLocation{}, fmt.Errorf("%w: %v", ErrCustomerVisitLocationResolveFailed, err)
	}

	province := strings.TrimSpace(firstNonEmpty(
		parsed.Address.Province,
		parsed.Address.State,
		parsed.Address.Region,
	))
	cityCandidate := strings.TrimSpace(firstNonEmpty(
		parsed.Address.City,
		parsed.Address.Municipality,
		parsed.Address.Town,
		parsed.Address.StateDistrict,
		parsed.Address.Region,
	))
	areaCandidate := strings.TrimSpace(firstNonEmpty(
		parsed.Address.County,
		parsed.Address.District,
		parsed.Address.CityDistrict,
		parsed.Address.Borough,
	))
	suburbCandidate := strings.TrimSpace(firstNonEmpty(
		parsed.Address.Suburb,
		parsed.Address.Neighbourhood,
		parsed.Address.Quarter,
	))

	city := cityCandidate
	area := areaCandidate
	if isDirectAdminMunicipality(province) {
		city = province
		if area == "" && looksLikeChineseDistrict(cityCandidate) {
			area = cityCandidate
		}
		if area == "" && looksLikeChineseDistrict(suburbCandidate) {
			area = suburbCandidate
		}
	} else if area == "" {
		area = suburbCandidate
	}

	result := CustomerVisitResolvedLocation{
		Province:      province,
		City:          city,
		Area:          area,
		DetailAddress: strings.TrimSpace(buildResolvedDetailAddress(parsed)),
	}

	if result.DetailAddress == "" {
		result.DetailAddress = strings.TrimSpace(parsed.DisplayName)
	}
	if result.Province == "" && result.City == "" && result.Area == "" && result.DetailAddress == "" {
		return CustomerVisitResolvedLocation{}, fmt.Errorf("%w: empty address", ErrCustomerVisitLocationResolveFailed)
	}

	return result, nil
}

func buildResolvedDetailAddress(resp nominatimReverseResponse) string {
	parts := []string{
		strings.TrimSpace(resp.Address.Suburb),
		strings.TrimSpace(resp.Address.Neighbourhood),
		strings.TrimSpace(resp.Address.Quarter),
		strings.TrimSpace(resp.Address.Borough),
		strings.TrimSpace(resp.Address.Road),
		strings.TrimSpace(resp.Address.HouseNumber),
		strings.TrimSpace(resp.Address.Village),
		strings.TrimSpace(resp.Address.Hamlet),
		strings.TrimSpace(resp.Address.Amenity),
		strings.TrimSpace(resp.Address.Building),
	}

	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		filtered = append(filtered, part)
	}
	return strings.Join(filtered, "")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func isDirectAdminMunicipality(value string) bool {
	switch strings.TrimSpace(value) {
	case "北京市", "天津市", "上海市", "重庆市":
		return true
	default:
		return false
	}
}

func looksLikeChineseDistrict(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	return strings.HasSuffix(trimmed, "区") || strings.HasSuffix(trimmed, "县") || strings.HasSuffix(trimmed, "旗")
}
