package service

import (
	"backend/internal/model"
	"context"
	"errors"
	"testing"
)

type customerVisitRepoStub struct {
	lastCreateInput model.CustomerVisitCreateInput
	existsSameDay   bool
	existsErr       error
}

func (s *customerVisitRepoStub) Create(input model.CustomerVisitCreateInput) (int64, error) {
	s.lastCreateInput = input
	return 1, nil
}

func (s *customerVisitRepoStub) ExistsSameDayVisitByUserCompanyIP(int64, string, string, string) (bool, error) {
	if s.existsErr != nil {
		return false, s.existsErr
	}
	return s.existsSameDay, nil
}

func (s *customerVisitRepoStub) List(model.CustomerVisitListFilter) (model.CustomerVisitListResult, error) {
	return model.CustomerVisitListResult{}, nil
}

type customerVisitLocationResolverStub struct {
	location CustomerVisitResolvedLocation
	err      error
}

func (s *customerVisitLocationResolverStub) Resolve(context.Context, float64, float64) (CustomerVisitResolvedLocation, error) {
	if s.err != nil {
		return CustomerVisitResolvedLocation{}, s.err
	}
	return s.location, nil
}

func TestCustomerVisitServiceCreateResolvesChineseLocation(t *testing.T) {
	t.Parallel()

	repo := &customerVisitRepoStub{}
	resolver := &customerVisitLocationResolverStub{
		location: CustomerVisitResolvedLocation{
			Province:      "天津市",
			City:          "天津市",
			Area:          "和平区",
			DetailAddress: "南京路189号",
		},
	}
	svc := NewCustomerVisitService(repo, resolver)

	_, err := svc.Create(context.Background(), model.CustomerVisitCreateInput{
		OperatorUserID: 1,
		CheckInLat:     39.15698464685517,
		CheckInLng:     117.23585571431285,
		VisitDate:      "2026-03-20",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if repo.lastCreateInput.Province != "天津市" {
		t.Fatalf("expected province 天津市, got %q", repo.lastCreateInput.Province)
	}
	if repo.lastCreateInput.City != "天津市" {
		t.Fatalf("expected city 天津市, got %q", repo.lastCreateInput.City)
	}
	if repo.lastCreateInput.Area != "和平区" {
		t.Fatalf("expected area 和平区, got %q", repo.lastCreateInput.Area)
	}
	if repo.lastCreateInput.DetailAddress != "南京路189号" {
		t.Fatalf("expected detailAddress 南京路189号, got %q", repo.lastCreateInput.DetailAddress)
	}
}

func TestCustomerVisitServiceCreateReturnsErrorWhenResolveFailsWithoutFallback(t *testing.T) {
	t.Parallel()

	repo := &customerVisitRepoStub{}
	resolver := &customerVisitLocationResolverStub{
		err: errors.New("resolve failed"),
	}
	svc := NewCustomerVisitService(repo, resolver)

	_, err := svc.Create(context.Background(), model.CustomerVisitCreateInput{
		OperatorUserID: 1,
		CheckInLat:     39.15698464685517,
		CheckInLng:     117.23585571431285,
		VisitDate:      "2026-03-20",
	})
	if err == nil {
		t.Fatal("expected error when location resolve fails without fallback fields")
	}
}

func TestCustomerVisitServiceCreateSkipsResolveWhenFrontendAddressProvided(t *testing.T) {
	t.Parallel()

	repo := &customerVisitRepoStub{}
	resolver := &customerVisitLocationResolverStub{
		location: CustomerVisitResolvedLocation{
			Province:      "不应覆盖的省",
			City:          "不应覆盖的市",
			Area:          "不应覆盖的区",
			DetailAddress: "不应覆盖的地址",
		},
	}
	svc := NewCustomerVisitService(repo, resolver)

	_, err := svc.Create(context.Background(), model.CustomerVisitCreateInput{
		OperatorUserID: 1,
		CustomerName:   "测试公司",
		CheckInIP:      "127.0.0.1",
		CheckInLat:     39.15698464685517,
		CheckInLng:     117.23585571431285,
		Province:       "天津市",
		City:           "天津市",
		Area:           "河北区",
		DetailAddress:  "江都路街道增产道",
		VisitDate:      "2026-03-20",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if repo.lastCreateInput.Province != "天津市" || repo.lastCreateInput.City != "天津市" || repo.lastCreateInput.Area != "河北区" || repo.lastCreateInput.DetailAddress != "江都路街道增产道" {
		t.Fatalf("expected frontend provided address to be preserved, got %#v", repo.lastCreateInput)
	}
}

func TestCustomerVisitServiceCreateRejectsDuplicateSameDayVisit(t *testing.T) {
	t.Parallel()

	repo := &customerVisitRepoStub{existsSameDay: true}
	svc := NewCustomerVisitService(repo, nil)

	_, err := svc.Create(context.Background(), model.CustomerVisitCreateInput{
		OperatorUserID: 1,
		CustomerName:   "测试公司",
		CheckInIP:      "127.0.0.1",
		VisitDate:      "2026-03-20",
	})
	if !errors.Is(err, ErrCustomerVisitAlreadyCheckedInToday) {
		t.Fatalf("expected ErrCustomerVisitAlreadyCheckedInToday, got %v", err)
	}
}
