package service

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestNominatimReverseGeocoderResolveMapsDirectAdminDistrict(t *testing.T) {
	t.Parallel()

	const expectedLat = "39.15698464685517"
	const expectedLon = "117.23585571431285"

	resolver := &nominatimReverseGeocoder{
		client: &http.Client{
			Transport: nominatimRoundTripFunc(func(r *http.Request) (*http.Response, error) {
				if got := r.URL.Path; got != "/reverse" {
					t.Fatalf("expected path /reverse, got %q", got)
				}
				if got := r.URL.Query().Get("format"); got != "json" {
					t.Fatalf("expected format=json, got %q", got)
				}
				if got := r.URL.Query().Get("lat"); got != expectedLat {
					t.Fatalf("expected lat %q, got %q", expectedLat, got)
				}
				if got := r.URL.Query().Get("lon"); got != expectedLon {
					t.Fatalf("expected lon %q, got %q", expectedLon, got)
				}
				if got := r.Header.Get("Accept-Language"); got != "zh-CN,zh;q=0.9" {
					t.Fatalf("expected Accept-Language header, got %q", got)
				}

				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body: io.NopCloser(strings.NewReader(`{
			"place_id":398956488,
			"display_name":"增产道, 江都路街道, 河北区, 天津市, 300151, 中国",
			"address":{
				"road":"增产道",
				"suburb":"江都路街道",
				"city":"河北区",
				"state":"天津市",
				"postcode":"300151",
				"country":"中国",
				"country_code":"cn"
			}
		}`)),
					Request: r,
				}, nil
			}),
		},
		baseURL:   "https://nominatim.openstreetmap.org",
		userAgent: "crm-go-vue-shadcn-test",
	}
	location, err := resolver.Resolve(context.Background(), 39.15698464685517, 117.23585571431285)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	if location.Province != "天津市" {
		t.Fatalf("expected province 天津市, got %q", location.Province)
	}
	if location.City != "天津市" {
		t.Fatalf("expected city 天津市, got %q", location.City)
	}
	if location.Area != "河北区" {
		t.Fatalf("expected area 河北区, got %q", location.Area)
	}
	if location.DetailAddress != "江都路街道增产道" {
		t.Fatalf("expected detail address 江都路街道增产道, got %q", location.DetailAddress)
	}
}

type nominatimRoundTripFunc func(*http.Request) (*http.Response, error)

func (f nominatimRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
