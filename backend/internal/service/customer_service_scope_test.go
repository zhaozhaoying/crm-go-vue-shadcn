package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"context"
	"errors"
	"testing"
	"time"
)

type customerSettingReaderStub struct {
	values map[string]string
}

func (s *customerSettingReaderStub) GetSetting(key string) (*model.SystemSetting, error) {
	if s == nil {
		return nil, nil
	}
	value, ok := s.values[key]
	if !ok {
		return nil, nil
	}
	return &model.SystemSetting{
		Key:   key,
		Value: value,
	}, nil
}

type customerScopeRepoStub struct {
	roleUsers    map[string][]int64
	userRoles    map[int64]string
	subordinates map[int64][]int64
	parents      map[int64]int64
	findByID     *model.Customer
	findByIDErr  error
	claimResult  *model.Customer
	claimErr     error
	claimCalled  bool
}

func (s *customerScopeRepoStub) List(ctx context.Context, filter model.CustomerListFilter) (model.CustomerListResult, error) {
	return model.CustomerListResult{}, nil
}

func (s *customerScopeRepoStub) FindByID(ctx context.Context, customerID int64) (*model.Customer, error) {
	if s.findByIDErr != nil {
		return nil, s.findByIDErr
	}
	if s.findByID != nil {
		customer := *s.findByID
		return &customer, nil
	}
	return nil, repository.ErrCustomerNotFound
}

func (s *customerScopeRepoStub) ListUserIDsByRoleNames(ctx context.Context, roleNames []string) ([]int64, error) {
	result := make([]int64, 0)
	for _, roleName := range roleNames {
		result = append(result, s.roleUsers[roleName]...)
	}
	return uniquePositiveInt64(result), nil
}

func (s *customerScopeRepoStub) ListEnabledUserIDsByRoleNames(ctx context.Context, roleNames []string) ([]int64, error) {
	return s.ListUserIDsByRoleNames(ctx, roleNames)
}

func (s *customerScopeRepoStub) ListDirectSubordinateUserIDsByRoleNames(ctx context.Context, parentIDs []int64, roleNames []string) ([]int64, error) {
	roleFilter := map[string]struct{}{}
	for _, roleName := range roleNames {
		if roleName == "" {
			continue
		}
		roleFilter[roleName] = struct{}{}
	}

	result := make([]int64, 0)
	for _, parentID := range parentIDs {
		for _, userID := range s.subordinates[parentID] {
			if len(roleFilter) > 0 {
				if _, ok := roleFilter[s.userRoles[userID]]; !ok {
					continue
				}
			}
			result = append(result, userID)
		}
	}
	return uniquePositiveInt64(result), nil
}

func (s *customerScopeRepoStub) GetUserRoleName(ctx context.Context, userID int64) (string, error) {
	return s.userRoles[userID], nil
}

func (s *customerScopeRepoStub) GetParentUserID(ctx context.Context, userID int64) (int64, error) {
	if s.parents == nil {
		return 0, nil
	}
	return s.parents[userID], nil
}

func (s *customerScopeRepoStub) ResolveDepartmentAnchorUserID(ctx context.Context, userID int64) (int64, error) {
	return 0, nil
}

func (s *customerScopeRepoStub) CountOwnedActiveByOwner(ctx context.Context, ownerUserID int64) (int64, error) {
	return 0, nil
}

func (s *customerScopeRepoStub) Create(ctx context.Context, input model.CustomerCreateInput) (*model.Customer, error) {
	return nil, nil
}

func (s *customerScopeRepoStub) Update(ctx context.Context, customerID int64, input model.CustomerUpdateInput) (*model.Customer, error) {
	return nil, nil
}

func (s *customerScopeRepoStub) CheckUnique(ctx context.Context, input model.CustomerUniqueCheckInput) (model.CustomerUniqueCheckResult, error) {
	return model.CustomerUniqueCheckResult{}, nil
}

func (s *customerScopeRepoStub) Claim(ctx context.Context, customerID, operatorUserID int64) (*model.Customer, error) {
	s.claimCalled = true
	if s.claimErr != nil {
		return nil, s.claimErr
	}
	if s.claimResult != nil {
		customer := *s.claimResult
		return &customer, nil
	}
	return nil, nil
}

func (s *customerScopeRepoStub) Release(ctx context.Context, customerID, operatorUserID int64) (*model.Customer, error) {
	return nil, nil
}

func (s *customerScopeRepoStub) Transfer(ctx context.Context, input model.CustomerTransferInput) (*model.Customer, error) {
	return nil, nil
}

func (s *customerScopeRepoStub) AddPhone(ctx context.Context, phone *model.CustomerPhone) error {
	return nil
}

func (s *customerScopeRepoStub) ListPhones(ctx context.Context, customerID int64) ([]model.CustomerPhone, error) {
	return nil, nil
}

func (s *customerScopeRepoStub) UpdatePhone(ctx context.Context, phone *model.CustomerPhone) error {
	return nil
}

func (s *customerScopeRepoStub) DeletePhone(ctx context.Context, customerID, phoneID int64) error {
	return nil
}

func (s *customerScopeRepoStub) GetPhone(ctx context.Context, phoneID int64) (*model.CustomerPhone, error) {
	return nil, nil
}

func (s *customerScopeRepoStub) FindCustomerIDByPhone(ctx context.Context, phone string) (int64, error) {
	return 0, nil
}

func (s *customerScopeRepoStub) CreateStatusLog(ctx context.Context, log *model.CustomerStatusLog) error {
	return nil
}

func (s *customerScopeRepoStub) ListStatusLogs(ctx context.Context, customerID int64, page, pageSize int) ([]model.CustomerStatusLog, error) {
	return nil, nil
}

func TestApplyPartnerCustomerScopeByRole(t *testing.T) {
	repoStub := &customerScopeRepoStub{
		roleUsers: map[string][]int64{
			"sales_director": {10},
			"sales_manager":  {11},
			"sales_staff":    {12, 13},
			"ops_manager":    {30},
			"ops_staff":      {31, 32},
		},
		userRoles: map[int64]string{
			10: "sales_director",
			11: "sales_manager",
			12: "sales_staff",
			13: "sales_staff",
			30: "ops_manager",
			31: "ops_staff",
			32: "ops_staff",
			50: "sales_staff",
		},
		subordinates: map[int64][]int64{
			10: {11, 12},
			11: {13},
		},
	}

	svc := &customerService{repo: repoStub}
	ctx := context.Background()

	testCases := []struct {
		name                   string
		viewerID               int64
		role                   string
		wantOwnerUserIDs       []int64
		wantServiceUserIDs     []int64
		wantForceServiceUserID *int64
	}{
		{
			name:             "admin sees all sales partner customers",
			viewerID:         1,
			role:             "admin",
			wantOwnerUserIDs: []int64{10, 11, 12, 13},
		},
		{
			name:             "finance manager sees all sales partner customers",
			viewerID:         2,
			role:             "finance_manager",
			wantOwnerUserIDs: []int64{10, 11, 12, 13},
		},
		{
			name:             "sales director sees self and team partner customers",
			viewerID:         10,
			role:             "sales_director",
			wantOwnerUserIDs: []int64{10, 11, 12, 13},
		},
		{
			name:             "sales manager sees self and descendants partner customers",
			viewerID:         11,
			role:             "sales_manager",
			wantOwnerUserIDs: []int64{11, 13},
		},
		{
			name:               "ops manager sees all assigned ops partner customers",
			viewerID:           30,
			role:               "ops_manager",
			wantServiceUserIDs: []int64{30, 31, 32},
		},
		{
			name:                   "ops staff sees only own assigned partner customers",
			viewerID:               31,
			role:                   "ops_staff",
			wantForceServiceUserID: int64Ptr(31),
		},
		{
			name:             "sales staff sees own partner customers",
			viewerID:         50,
			role:             "sales_staff",
			wantOwnerUserIDs: []int64{50},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filter, err := svc.applyPartnerCustomerScopeByRole(ctx, model.CustomerListFilter{
				Category:  "partner",
				ViewerID:  tc.viewerID,
				HasViewer: true,
				ActorRole: tc.role,
			})
			if err != nil {
				t.Fatalf("applyPartnerCustomerScopeByRole returned error: %v", err)
			}

			assertSameIDs(t, filter.AllowedOwnerUserIDs, tc.wantOwnerUserIDs)
			assertSameIDs(t, filter.AllowedServiceUserIDs, tc.wantServiceUserIDs)

			if tc.wantForceServiceUserID == nil {
				if filter.ForceServiceUserID != nil {
					t.Fatalf("expected nil ForceServiceUserID, got %d", *filter.ForceServiceUserID)
				}
			} else {
				if filter.ForceServiceUserID == nil {
					t.Fatalf("expected ForceServiceUserID %d, got nil", *tc.wantForceServiceUserID)
				}
				if *filter.ForceServiceUserID != *tc.wantForceServiceUserID {
					t.Fatalf("expected ForceServiceUserID %d, got %d", *tc.wantForceServiceUserID, *filter.ForceServiceUserID)
				}
			}
		})
	}
}

func TestClaimCustomerReturnsFreezeErrorForSelfDroppedCustomer(t *testing.T) {
	dropTime := time.Now().UTC().Add(-24 * time.Hour)
	repoStub := &customerScopeRepoStub{
		findByID: &model.Customer{
			ID:         1001,
			Name:       "测试客户",
			IsInPool:   true,
			DropUserID: int64Ptr(7),
			DropTime:   &dropTime,
		},
	}

	svc := &customerService{
		repo: repoStub,
		settingReader: &customerSettingReaderStub{
			values: map[string]string{
				claimFreezeSettingKey: "3",
			},
		},
	}

	_, err := svc.ClaimCustomer(context.Background(), 1001, 7)
	var freezeErr *CustomerClaimFreezeError
	if !errors.As(err, &freezeErr) {
		t.Fatalf("expected CustomerClaimFreezeError, got %v", err)
	}
	if freezeErr.FreezeDays != 3 {
		t.Fatalf("expected freeze days 3, got %d", freezeErr.FreezeDays)
	}
	if freezeErr.Remaining <= 0 {
		t.Fatalf("expected positive remaining freeze duration, got %v", freezeErr.Remaining)
	}
	if repoStub.claimCalled {
		t.Fatalf("claim repository should not be called when self-dropped customer is still frozen")
	}
}

func TestClaimCustomerReturnsSameDepartmentForbidden(t *testing.T) {
	repoStub := &customerScopeRepoStub{
		userRoles: map[int64]string{
			1: "sales_staff",
			9: "sales_staff",
			4: "sales_director",
		},
		parents: map[int64]int64{
			1: 4,
			9: 4,
		},
		findByID: &model.Customer{
			ID:         1002,
			Name:       "测试客户2",
			IsInPool:   true,
			DropUserID: int64Ptr(9),
		},
	}

	svc := &customerService{repo: repoStub}

	_, err := svc.ClaimCustomer(context.Background(), 1002, 1)
	if !errors.Is(err, ErrCustomerSameDepartmentClaimForbidden) {
		t.Fatalf("expected ErrCustomerSameDepartmentClaimForbidden, got %v", err)
	}
	if repoStub.claimCalled {
		t.Fatalf("claim repository should not be called when same-department customer is forbidden")
	}
}

func TestIsSameDepartmentUsesSalesDirectorScope(t *testing.T) {
	repoStub := &customerScopeRepoStub{
		userRoles: map[int64]string{
			100: "sales_director",
			101: "sales_staff",
			102: "sales_staff",
			200: "sales_director",
			201: "sales_staff",
			300: "admin",
		},
		parents: map[int64]int64{
			101: 100,
			102: 100,
			201: 200,
		},
	}

	svc := &customerService{repo: repoStub}
	ctx := context.Background()

	testCases := []struct {
		name    string
		leftID  int64
		rightID int64
		want    bool
	}{
		{
			name:    "same sales director subordinates are blocked",
			leftID:  101,
			rightID: 102,
			want:    true,
		},
		{
			name:    "sales director and subordinate are blocked",
			leftID:  100,
			rightID: 101,
			want:    true,
		},
		{
			name:    "different sales director teams are allowed",
			leftID:  101,
			rightID: 201,
			want:    false,
		},
		{
			name:    "non sales director role does not trigger team block",
			leftID:  101,
			rightID: 300,
			want:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := svc.isSameDepartment(ctx, tc.leftID, tc.rightID)
			if err != nil {
				t.Fatalf("isSameDepartment returned error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("unexpected result: got %v want %v", got, tc.want)
			}
		})
	}
}

func assertSameIDs(t *testing.T, got, want []int64) {
	t.Helper()
	got = uniquePositiveInt64(got)
	want = uniquePositiveInt64(want)
	if len(got) != len(want) {
		t.Fatalf("unexpected id count: got %v want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("unexpected ids: got %v want %v", got, want)
		}
	}
}

func int64Ptr(value int64) *int64 {
	return &value
}
