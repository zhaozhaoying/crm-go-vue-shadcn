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
	roleUsers                            map[string][]int64
	userRoles                            map[int64]string
	displayNames                         map[int64]string
	subordinates                         map[int64][]int64
	parents                              map[int64]int64
	rankedUserIDs                        []int64
	recentContractExemptUserIDs          []int64
	enabledUserIDsByNickname             map[string]int64
	lastRecentContractSince              time.Time
	lastRankReferenceDate                string
	latestAutoAssignOwnerUserID          *int64
	activeBlockedUntilByDepartmentAnchor map[int64]time.Time
	findByID                             *model.Customer
	findByIDErr                          error
	createResult                         *model.Customer
	createErr                            error
	lastCreateInput                      model.CustomerCreateInput
	claimResult                          *model.Customer
	claimErr                             error
	claimCalled                          bool
	claimOwnerUserID                     int64
	claimOperatorUserID                  int64
	claimInsideSalesUserID               *int64
	convertResult                        *model.Customer
	convertErr                           error
	convertCalled                        bool
	convertOwner                         int64
	convertBy                            int64
	transferInputs                       []model.CustomerTransferInput
}

func (s *customerScopeRepoStub) List(ctx context.Context, filter model.CustomerListFilter) (model.CustomerListResult, error) {
	return model.CustomerListResult{}, nil
}

func (s *customerScopeRepoStub) ListAssignments(ctx context.Context, filter model.CustomerAssignmentListFilter) (model.CustomerAssignmentListResult, error) {
	return model.CustomerAssignmentListResult{}, nil
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

func (s *customerScopeRepoStub) GetUserDisplayName(ctx context.Context, userID int64) (string, error) {
	if s.displayNames != nil {
		return s.displayNames[userID], nil
	}
	return "", nil
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

func (s *customerScopeRepoStub) GetActiveBlockedUntilByDepartmentAnchor(ctx context.Context, customerID, departmentAnchorUserID int64, now time.Time) (*time.Time, error) {
	if s.activeBlockedUntilByDepartmentAnchor == nil {
		return nil, nil
	}
	blockedUntil, ok := s.activeBlockedUntilByDepartmentAnchor[departmentAnchorUserID]
	if !ok || !blockedUntil.After(now) {
		return nil, nil
	}
	result := blockedUntil
	return &result, nil
}

func (s *customerScopeRepoStub) CountOwnedActiveByOwner(ctx context.Context, ownerUserID int64) (int64, error) {
	return 0, nil
}

func (s *customerScopeRepoStub) ListAutoAssignRankedOwnerScores(ctx context.Context, referenceDate string, userIDs []int64) ([]model.SalesDailyScore, error) {
	s.lastRankReferenceDate = referenceDate
	if len(s.rankedUserIDs) == 0 {
		return []model.SalesDailyScore{}, nil
	}
	allowed := make(map[int64]struct{}, len(userIDs))
	for _, userID := range userIDs {
		allowed[userID] = struct{}{}
	}
	result := make([]model.SalesDailyScore, 0, len(s.rankedUserIDs))
	for _, userID := range s.rankedUserIDs {
		if _, ok := allowed[userID]; ok {
			result = append(result, model.SalesDailyScore{UserID: userID, TotalScore: autoAssignMinimumDailyScore})
		}
	}
	return result, nil
}

func (s *customerScopeRepoStub) ListRecentContractExemptOwnerUserIDs(ctx context.Context, since time.Time, userIDs []int64) ([]int64, error) {
	s.lastRecentContractSince = since
	if len(s.recentContractExemptUserIDs) == 0 {
		return []int64{}, nil
	}
	allowed := make(map[int64]struct{}, len(userIDs))
	for _, userID := range userIDs {
		allowed[userID] = struct{}{}
	}
	result := make([]int64, 0, len(s.recentContractExemptUserIDs))
	for _, userID := range s.recentContractExemptUserIDs {
		if _, ok := allowed[userID]; ok {
			result = append(result, userID)
		}
	}
	return result, nil
}

func (s *customerScopeRepoStub) FindEnabledUserIDByNickname(ctx context.Context, nickname string) (int64, error) {
	if s.enabledUserIDsByNickname == nil {
		return 0, nil
	}
	return s.enabledUserIDsByNickname[nickname], nil
}

func (s *customerScopeRepoStub) FindLatestAutoAssignOwnerUserID(ctx context.Context, ownerUserIDs []int64, since time.Time) (*int64, error) {
	if s.latestAutoAssignOwnerUserID == nil {
		return nil, nil
	}
	for _, ownerUserID := range ownerUserIDs {
		if ownerUserID == *s.latestAutoAssignOwnerUserID {
			return s.latestAutoAssignOwnerUserID, nil
		}
	}
	return nil, nil
}

func (s *customerScopeRepoStub) Create(ctx context.Context, input model.CustomerCreateInput) (*model.Customer, error) {
	s.lastCreateInput = input
	if s.createErr != nil {
		return nil, s.createErr
	}
	if s.createResult != nil {
		customer := *s.createResult
		return &customer, nil
	}
	return nil, nil
}

func (s *customerScopeRepoStub) Update(ctx context.Context, customerID int64, input model.CustomerUpdateInput) (*model.Customer, error) {
	return nil, nil
}

func (s *customerScopeRepoStub) CheckUnique(ctx context.Context, input model.CustomerUniqueCheckInput) (model.CustomerUniqueCheckResult, error) {
	return model.CustomerUniqueCheckResult{}, nil
}

func (s *customerScopeRepoStub) Claim(ctx context.Context, customerID, ownerUserID, operatorUserID int64, insideSalesUserID *int64) (*model.Customer, error) {
	s.claimCalled = true
	s.claimOwnerUserID = ownerUserID
	s.claimOperatorUserID = operatorUserID
	s.claimInsideSalesUserID = insideSalesUserID
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
	s.transferInputs = append(s.transferInputs, input)
	return &model.Customer{
		ID:          input.CustomerID,
		OwnerUserID: &input.ToOwnerUserID,
	}, nil
}

func (s *customerScopeRepoStub) Convert(ctx context.Context, customerID, ownerUserID, operatorUserID int64) (*model.Customer, error) {
	s.convertCalled = true
	s.convertOwner = ownerUserID
	s.convertBy = operatorUserID
	if s.convertErr != nil {
		return nil, s.convertErr
	}
	if s.convertResult != nil {
		customer := *s.convertResult
		return &customer, nil
	}
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

func TestApplyMyCustomerScopeByRoleForInsideSalesDepartment(t *testing.T) {
	repoStub := &customerScopeRepoStub{
		roleUsers: map[string][]int64{
			"sales_inside": {21, 22},
		},
		userRoles: map[int64]string{
			1:  "admin",
			10: "sales_manager",
			21: "sales_inside",
			22: "sales_inside",
			30: "sales_staff",
		},
		subordinates: map[int64][]int64{
			10: {21, 30},
		},
	}

	svc := &customerService{repo: repoStub}

	adminFilter, err := svc.applyMyCustomerScopeByRole(context.Background(), model.CustomerListFilter{
		Category:       "my",
		OwnershipScope: "inside_sales",
		ViewerID:       1,
		HasViewer:      true,
		ActorRole:      "admin",
	})
	if err != nil {
		t.Fatalf("applyMyCustomerScopeByRole(admin) returned error: %v", err)
	}
	assertSameIDs(t, adminFilter.AllowedInsideSalesUserIDs, []int64{21, 22})
	if !adminFilter.SkipViewerOwnerLimit {
		t.Fatalf("expected admin inside-sales scope to skip owner limit")
	}
	if !adminFilter.IncludePoolInMyScope {
		t.Fatalf("expected inside-sales scope to include pending pool customers for admin")
	}
	if !adminFilter.IncludePendingConvertScope {
		t.Fatalf("expected inside-sales scope to include pending conversion records for admin")
	}
	if !adminFilter.RequireInsideSalesAssociation {
		t.Fatalf("expected admin inside-sales scope to require inside-sales association")
	}

	managerFilter, err := svc.applyMyCustomerScopeByRole(context.Background(), model.CustomerListFilter{
		Category:       "my",
		OwnershipScope: "inside_sales",
		ViewerID:       10,
		HasViewer:      true,
		ActorRole:      "sales_manager",
	})
	if err != nil {
		t.Fatalf("applyMyCustomerScopeByRole(manager) returned error: %v", err)
	}
	assertSameIDs(t, managerFilter.AllowedInsideSalesUserIDs, []int64{21})
	assertSameIDs(t, managerFilter.AllowedOwnerUserIDs, []int64{10, 21, 30})
	if !managerFilter.RequireInsideSalesAssociation {
		t.Fatalf("expected manager inside-sales scope to require inside-sales association")
	}
}

func TestApplyMyCustomerScopeByRoleForInsideSalesDepartmentKeepsSalesOwnerScopeWhenNoInsideSalesSubordinates(t *testing.T) {
	repoStub := &customerScopeRepoStub{
		roleUsers: map[string][]int64{
			"sales_inside": {21, 22},
		},
		userRoles: map[int64]string{
			31: "sales_staff",
		},
	}

	svc := &customerService{repo: repoStub}
	filter, err := svc.applyMyCustomerScopeByRole(context.Background(), model.CustomerListFilter{
		Category:       "my",
		OwnershipScope: "inside_sales",
		ViewerID:       31,
		HasViewer:      true,
		ActorRole:      "sales_staff",
	})
	if err != nil {
		t.Fatalf("applyMyCustomerScopeByRole(sales-staff inside-sales scope) returned error: %v", err)
	}
	assertSameIDs(t, filter.AllowedOwnerUserIDs, []int64{31})
	assertSameIDs(t, filter.AllowedInsideSalesUserIDs, []int64{})
	if !filter.RequireInsideSalesAssociation {
		t.Fatalf("expected sales-staff inside-sales scope to require inside-sales association")
	}
}

func TestApplyMyCustomerScopeByRoleForInsideSalesSelfOnly(t *testing.T) {
	repoStub := &customerScopeRepoStub{
		userRoles: map[int64]string{
			21: "sales_inside",
		},
	}

	svc := &customerService{repo: repoStub}
	filter, err := svc.applyMyCustomerScopeByRole(context.Background(), model.CustomerListFilter{
		Category:       "my",
		OwnershipScope: "all",
		ViewerID:       21,
		HasViewer:      true,
		ActorRole:      "sales_inside",
	})
	if err != nil {
		t.Fatalf("applyMyCustomerScopeByRole(inside-sales) returned error: %v", err)
	}
	assertSameIDs(t, filter.AllowedOwnerUserIDs, []int64{21})
	assertSameIDs(t, filter.AllowedInsideSalesUserIDs, []int64{21})
	if !filter.IncludePendingConvertScope {
		t.Fatalf("expected inside-sales self scope to include pending conversion records")
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
	blockedUntil := time.Now().UTC().Add(48 * time.Hour)
	repoStub := &customerScopeRepoStub{
		userRoles: map[int64]string{
			1: "sales_staff",
			4: "sales_director",
		},
		parents: map[int64]int64{
			1: 4,
		},
		activeBlockedUntilByDepartmentAnchor: map[int64]time.Time{4: blockedUntil},
		findByID: &model.Customer{
			ID:         1002,
			Name:       "测试客户2",
			IsInPool:   true,
			DropUserID: int64Ptr(9),
		},
	}

	svc := &customerService{repo: repoStub}

	_, err := svc.ClaimCustomer(context.Background(), 1002, 1)
	var freezeErr *CustomerClaimFreezeError
	if !errors.As(err, &freezeErr) {
		t.Fatalf("expected CustomerClaimFreezeError, got %v", err)
	}
	if freezeErr.BlockType != customerClaimFreezeBlockTypeDepartment {
		t.Fatalf("expected department freeze error, got %q", freezeErr.BlockType)
	}
	if freezeErr.Remaining <= 0 {
		t.Fatalf("expected positive remaining duration, got %v", freezeErr.Remaining)
	}
	if repoStub.claimCalled {
		t.Fatalf("claim repository should not be called when same-department customer is forbidden")
	}
}

func TestClaimCustomerReturnsHistoricalDepartmentForbidden(t *testing.T) {
	blockedUntil := time.Now().UTC().Add(72 * time.Hour)
	repoStub := &customerScopeRepoStub{
		userRoles: map[int64]string{
			1: "sales_staff",
			2: "sales_staff",
			4: "sales_director",
			8: "sales_staff",
			9: "sales_staff",
			7: "sales_director",
		},
		parents: map[int64]int64{
			1: 4,
			2: 4,
			8: 7,
			9: 7,
		},
		activeBlockedUntilByDepartmentAnchor: map[int64]time.Time{
			7: blockedUntil,
			4: blockedUntil,
		},
		findByID: &model.Customer{
			ID:         1003,
			Name:       "测试客户3",
			IsInPool:   true,
			DropUserID: int64Ptr(9),
		},
	}

	svc := &customerService{repo: repoStub}

	_, err := svc.ClaimCustomer(context.Background(), 1003, 2)
	var freezeErr *CustomerClaimFreezeError
	if !errors.As(err, &freezeErr) {
		t.Fatalf("expected CustomerClaimFreezeError, got %v", err)
	}
	if freezeErr.BlockType != customerClaimFreezeBlockTypeDepartment {
		t.Fatalf("expected department freeze error, got %q", freezeErr.BlockType)
	}
	if freezeErr.FrozenUntil.Before(time.Now().UTC()) {
		t.Fatalf("expected future frozen until, got %v", freezeErr.FrozenUntil)
	}
	if repoStub.claimCalled {
		t.Fatalf("claim repository should not be called when historical same-department customer is forbidden")
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

func TestCreateCustomerInsideSalesKeepsCustomerOnInsideSalesWhenNoScores(t *testing.T) {
	repoStub := &customerScopeRepoStub{
		roleUsers: map[string][]int64{
			"sales_director": {1},
			"sales_staff":    {3, 4},
		},
		userRoles: map[int64]string{
			1: "sales_director",
			3: "sales_staff",
			4: "sales_staff",
			9: "sales_inside",
		},
		parents: map[int64]int64{
			9: 1,
			3: 1,
			4: 1,
		},
		subordinates: map[int64][]int64{
			1: {9, 3, 4},
		},
		createResult: &model.Customer{
			ID:   2001,
			Name: "自动转化客户",
		},
	}
	svc := &customerService{repo: repoStub}

	_, err := svc.CreateCustomer(context.Background(), model.CustomerCreateInput{
		Name:           "自动转化客户",
		LegalName:      "张三",
		ContactName:    "李四",
		Status:         model.CustomerStatusPool,
		OperatorUserID: 9,
		Phones: []model.CustomerPhoneInput{{
			Phone:     "13800138000",
			IsPrimary: true,
		}},
	})
	if err != nil {
		t.Fatalf("CreateCustomer returned error: %v", err)
	}
	if repoStub.lastCreateInput.Status != model.CustomerStatusOwned {
		t.Fatalf("expected inside-sales create status %q, got %q", model.CustomerStatusOwned, repoStub.lastCreateInput.Status)
	}
	if repoStub.lastCreateInput.OwnerUserID == nil || *repoStub.lastCreateInput.OwnerUserID != 9 {
		t.Fatalf("expected inside-sales create owner 9, got %v", repoStub.lastCreateInput.OwnerUserID)
	}
	if repoStub.lastCreateInput.InsideSalesUserID == nil || *repoStub.lastCreateInput.InsideSalesUserID != 9 {
		t.Fatalf("expected inside-sales create to bind insideSalesUserID 9, got %v", repoStub.lastCreateInput.InsideSalesUserID)
	}
	if repoStub.lastCreateInput.ConvertedAt != nil {
		t.Fatalf("expected inside-sales create convertedAt to stay nil without scores")
	}
}

func TestClaimCustomerInsideSalesKeepsCustomerOnInsideSalesWhenNoScores(t *testing.T) {
	repoStub := &customerScopeRepoStub{
		userRoles: map[int64]string{
			1: "sales_director",
			2: "sales_inside",
			3: "sales_staff",
			4: "sales_staff",
		},
		parents: map[int64]int64{
			2: 1,
			3: 1,
			4: 1,
		},
		subordinates: map[int64][]int64{
			1: {2, 3, 4},
		},
		roleUsers: map[string][]int64{
			"sales_director": {1},
			"sales_staff":    {3, 4},
		},
		findByID: &model.Customer{
			ID:       2101,
			Name:     "公海客户",
			IsInPool: true,
		},
		claimResult: &model.Customer{
			ID:       2101,
			Name:     "公海客户",
			IsInPool: false,
		},
	}
	svc := &customerService{repo: repoStub}

	_, err := svc.ClaimCustomer(context.Background(), 2101, 2)
	if err != nil {
		t.Fatalf("ClaimCustomer returned error: %v", err)
	}
	if !repoStub.claimCalled {
		t.Fatalf("expected Claim repository method to be called")
	}
	if repoStub.claimOwnerUserID != 2 {
		t.Fatalf("expected inside-sales claim owner 2, got %d", repoStub.claimOwnerUserID)
	}
	if repoStub.claimOperatorUserID != 2 {
		t.Fatalf("expected inside-sales claim operator 2, got %d", repoStub.claimOperatorUserID)
	}
	if repoStub.claimInsideSalesUserID == nil || *repoStub.claimInsideSalesUserID != 2 {
		t.Fatalf("expected inside-sales claim to bind insideSalesUserID 2, got %v", repoStub.claimInsideSalesUserID)
	}
}

func TestCreateCustomerInsideSalesAutoAssignsTopRankedOwner(t *testing.T) {
	repoStub := &customerScopeRepoStub{
		roleUsers: map[string][]int64{
			"sales_director": {1},
			"sales_staff":    {3, 4},
		},
		userRoles: map[int64]string{
			1: "sales_director",
			3: "sales_staff",
			4: "sales_staff",
			9: "sales_inside",
		},
		parents: map[int64]int64{
			9: 1,
			3: 1,
			4: 1,
		},
		subordinates: map[int64][]int64{
			1: {9, 3, 4},
		},
		rankedUserIDs: []int64{4, 3, 1},
		createResult: &model.Customer{
			ID:   2002,
			Name: "排名自动分配客户",
		},
	}
	svc := &customerService{repo: repoStub}

	_, err := svc.CreateCustomer(context.Background(), model.CustomerCreateInput{
		Name:           "排名自动分配客户",
		LegalName:      "张三",
		ContactName:    "李四",
		Status:         model.CustomerStatusPool,
		OperatorUserID: 9,
		Phones: []model.CustomerPhoneInput{{
			Phone:     "13800138001",
			IsPrimary: true,
		}},
	})
	if err != nil {
		t.Fatalf("CreateCustomer returned error: %v", err)
	}
	if repoStub.lastCreateInput.OwnerUserID == nil || *repoStub.lastCreateInput.OwnerUserID != 4 {
		t.Fatalf("expected ranked inside-sales create owner 4, got %v", repoStub.lastCreateInput.OwnerUserID)
	}
	if repoStub.lastRankReferenceDate != previousAutoAssignScoreDate() {
		t.Fatalf("expected yesterday reference date %q, got %q", previousAutoAssignScoreDate(), repoStub.lastRankReferenceDate)
	}
}

func TestCreateCustomerInsideSalesAutoAssignsRankedSaleOutsideAliasOwner(t *testing.T) {
	repoStub := &customerScopeRepoStub{
		roleUsers: map[string][]int64{
			"sale_outside": {4},
		},
		userRoles: map[int64]string{
			1: "sales_director",
			4: "sale_outside",
			9: "sale_inside",
		},
		parents: map[int64]int64{
			9: 1,
			4: 1,
		},
		subordinates: map[int64][]int64{
			1: {9, 4},
		},
		rankedUserIDs: []int64{4},
		createResult: &model.Customer{
			ID:   20021,
			Name: "别名外销自动分配客户",
		},
	}
	svc := &customerService{repo: repoStub}

	_, err := svc.CreateCustomer(context.Background(), model.CustomerCreateInput{
		Name:           "别名外销自动分配客户",
		LegalName:      "张三",
		ContactName:    "李四",
		Status:         model.CustomerStatusPool,
		OperatorUserID: 9,
		Phones: []model.CustomerPhoneInput{{
			Phone:     "13800138021",
			IsPrimary: true,
		}},
	})
	if err != nil {
		t.Fatalf("CreateCustomer returned error: %v", err)
	}
	if repoStub.lastCreateInput.OwnerUserID == nil || *repoStub.lastCreateInput.OwnerUserID != 4 {
		t.Fatalf("expected ranked sale_outside alias owner 4, got %v", repoStub.lastCreateInput.OwnerUserID)
	}
}

func TestCreateCustomerInsideSalesFallsBackToRecentContractExemptOwner(t *testing.T) {
	repoStub := &customerScopeRepoStub{
		roleUsers: map[string][]int64{
			"sales_director": {1},
			"sales_staff":    {3, 4},
		},
		userRoles: map[int64]string{
			1: "sales_director",
			3: "sales_staff",
			4: "sales_staff",
			9: "sales_inside",
		},
		parents: map[int64]int64{
			9: 1,
			3: 1,
			4: 1,
		},
		subordinates: map[int64][]int64{
			1: {9, 3, 4},
		},
		recentContractExemptUserIDs: []int64{3, 4},
		createResult: &model.Customer{
			ID:   2003,
			Name: "签单豁免客户",
		},
	}
	svc := &customerService{repo: repoStub}

	_, err := svc.CreateCustomer(context.Background(), model.CustomerCreateInput{
		Name:           "签单豁免客户",
		LegalName:      "张三",
		ContactName:    "李四",
		Status:         model.CustomerStatusPool,
		OperatorUserID: 9,
		Phones: []model.CustomerPhoneInput{{
			Phone:     "13800138011",
			IsPrimary: true,
		}},
	})
	if err != nil {
		t.Fatalf("CreateCustomer returned error: %v", err)
	}
	if repoStub.lastCreateInput.OwnerUserID == nil || *repoStub.lastCreateInput.OwnerUserID != 3 {
		t.Fatalf("expected recent contract exempt owner 3, got %v", repoStub.lastCreateInput.OwnerUserID)
	}
	if repoStub.lastCreateInput.ConvertedAt == nil {
		t.Fatalf("expected recent contract exempt assignment to mark convertedAt")
	}
}

func TestCreateCustomerInsideSalesFallsBackToCounterpartDirectorWhenLiTeamHasNoThresholdOrContracts(t *testing.T) {
	repoStub := &customerScopeRepoStub{
		roleUsers: map[string][]int64{
			"sales_director": {1, 7},
			"sales_staff":    {3, 4},
		},
		userRoles: map[int64]string{
			1: "sales_director",
			3: "sales_staff",
			4: "sales_staff",
			7: "sales_director",
			9: "sales_inside",
		},
		displayNames: map[int64]string{
			1: "李龙泉",
			7: "葛鹏辉",
		},
		enabledUserIDsByNickname: map[string]int64{
			"葛鹏辉": 7,
		},
		parents: map[int64]int64{
			9: 1,
			3: 1,
			4: 1,
		},
		subordinates: map[int64][]int64{
			1: {9, 3, 4},
		},
		createResult: &model.Customer{
			ID:   2004,
			Name: "跨部门兜底客户",
		},
	}
	svc := &customerService{repo: repoStub}

	_, err := svc.CreateCustomer(context.Background(), model.CustomerCreateInput{
		Name:           "跨部门兜底客户",
		LegalName:      "张三",
		ContactName:    "李四",
		Status:         model.CustomerStatusPool,
		OperatorUserID: 9,
		Phones: []model.CustomerPhoneInput{{
			Phone:     "13800138012",
			IsPrimary: true,
		}},
	})
	if err != nil {
		t.Fatalf("CreateCustomer returned error: %v", err)
	}
	if repoStub.lastCreateInput.OwnerUserID == nil || *repoStub.lastCreateInput.OwnerUserID != 7 {
		t.Fatalf("expected counterpart director owner 7, got %v", repoStub.lastCreateInput.OwnerUserID)
	}
	if repoStub.lastCreateInput.ConvertedAt == nil {
		t.Fatalf("expected counterpart fallback assignment to mark convertedAt")
	}
}

func TestClaimCustomerInsideSalesAutoAssignsTopRankedOwner(t *testing.T) {
	repoStub := &customerScopeRepoStub{
		userRoles: map[int64]string{
			1: "sales_director",
			2: "sales_inside",
			3: "sales_staff",
			4: "sales_staff",
		},
		parents: map[int64]int64{
			2: 1,
			3: 1,
			4: 1,
		},
		subordinates: map[int64][]int64{
			1: {2, 3, 4},
		},
		roleUsers: map[string][]int64{
			"sales_director": {1},
			"sales_staff":    {3, 4},
		},
		rankedUserIDs: []int64{4, 3, 1},
		findByID: &model.Customer{
			ID:       2102,
			Name:     "排名公海客户",
			IsInPool: true,
		},
		claimResult: &model.Customer{
			ID:       2102,
			Name:     "排名公海客户",
			IsInPool: false,
		},
	}
	svc := &customerService{repo: repoStub}

	_, err := svc.ClaimCustomer(context.Background(), 2102, 2)
	if err != nil {
		t.Fatalf("ClaimCustomer returned error: %v", err)
	}
	if repoStub.claimOwnerUserID != 4 {
		t.Fatalf("expected ranked inside-sales claim owner 4, got %d", repoStub.claimOwnerUserID)
	}
	if repoStub.lastRankReferenceDate != previousAutoAssignScoreDate() {
		t.Fatalf("expected yesterday reference date %q, got %q", previousAutoAssignScoreDate(), repoStub.lastRankReferenceDate)
	}
}

func TestClaimCustomerInsideSalesAutoAssignsRankedSaleOutsideAliasOwner(t *testing.T) {
	repoStub := &customerScopeRepoStub{
		userRoles: map[int64]string{
			1: "sales_director",
			2: "sale_inside",
			4: "sale_outside",
		},
		parents: map[int64]int64{
			2: 1,
			4: 1,
		},
		subordinates: map[int64][]int64{
			1: {2, 4},
		},
		roleUsers: map[string][]int64{
			"sale_outside": {4},
		},
		rankedUserIDs: []int64{4},
		findByID: &model.Customer{
			ID:       21021,
			Name:     "别名外销公海客户",
			IsInPool: true,
		},
		claimResult: &model.Customer{
			ID:       21021,
			Name:     "别名外销公海客户",
			IsInPool: false,
		},
	}
	svc := &customerService{repo: repoStub}

	_, err := svc.ClaimCustomer(context.Background(), 21021, 2)
	if err != nil {
		t.Fatalf("ClaimCustomer returned error: %v", err)
	}
	if repoStub.claimOwnerUserID != 4 {
		t.Fatalf("expected ranked sale_outside alias claim owner 4, got %d", repoStub.claimOwnerUserID)
	}
}

func TestClaimCustomerInsideSalesFallsBackToRecentContractExemptOwner(t *testing.T) {
	repoStub := &customerScopeRepoStub{
		userRoles: map[int64]string{
			1: "sales_director",
			2: "sales_inside",
			3: "sales_staff",
			4: "sales_staff",
		},
		parents: map[int64]int64{
			2: 1,
			3: 1,
			4: 1,
		},
		subordinates: map[int64][]int64{
			1: {2, 3, 4},
		},
		roleUsers: map[string][]int64{
			"sales_director": {1},
			"sales_staff":    {3, 4},
		},
		recentContractExemptUserIDs: []int64{4, 3},
		findByID: &model.Customer{
			ID:       2104,
			Name:     "签单豁免公海客户",
			IsInPool: true,
		},
		claimResult: &model.Customer{
			ID:       2104,
			Name:     "签单豁免公海客户",
			IsInPool: false,
		},
	}
	svc := &customerService{repo: repoStub}

	_, err := svc.ClaimCustomer(context.Background(), 2104, 2)
	if err != nil {
		t.Fatalf("ClaimCustomer returned error: %v", err)
	}
	if repoStub.claimOwnerUserID != 4 {
		t.Fatalf("expected recent contract exempt claim owner 4, got %d", repoStub.claimOwnerUserID)
	}
}

func TestClaimCustomerInsideSalesRoundsRobinAcrossRankedOwnersExcludingLast(t *testing.T) {
	latestOwnerUserID := int64(4)
	repoStub := &customerScopeRepoStub{
		userRoles: map[int64]string{
			1: "sales_director",
			2: "sales_inside",
			3: "sales_staff",
			4: "sales_staff",
		},
		parents: map[int64]int64{
			2: 1,
			3: 1,
			4: 1,
		},
		subordinates: map[int64][]int64{
			1: {2, 3, 4},
		},
		roleUsers: map[string][]int64{
			"sales_director": {1},
			"sales_staff":    {3, 4},
		},
		rankedUserIDs:               []int64{4, 3, 1},
		latestAutoAssignOwnerUserID: &latestOwnerUserID,
		findByID: &model.Customer{
			ID:       2103,
			Name:     "轮转公海客户",
			IsInPool: true,
		},
		claimResult: &model.Customer{
			ID:       2103,
			Name:     "轮转公海客户",
			IsInPool: false,
		},
	}
	svc := &customerService{repo: repoStub}

	_, err := svc.ClaimCustomer(context.Background(), 2103, 2)
	if err != nil {
		t.Fatalf("ClaimCustomer returned error: %v", err)
	}
	if repoStub.claimOwnerUserID != 3 {
		t.Fatalf("expected ranked round-robin claim owner 3, got %d", repoStub.claimOwnerUserID)
	}
}

func TestConvertCustomerInsideSalesKeepsCustomerOnInsideSalesWhenNoScores(t *testing.T) {
	repoStub := &customerScopeRepoStub{
		userRoles: map[int64]string{
			1: "sales_director",
			2: "sales_inside",
			3: "sales_staff",
			4: "sales_staff",
		},
		parents: map[int64]int64{
			2: 1,
			3: 1,
			4: 1,
		},
		subordinates: map[int64][]int64{
			1: {2, 3, 4},
		},
		roleUsers: map[string][]int64{
			"sales_director": {1},
			"sales_staff":    {3, 4},
		},
		findByID: &model.Customer{
			ID:           3001,
			Name:         "电销待转化客户",
			CreateUserID: 2,
			IsInPool:     true,
		},
		convertResult: &model.Customer{
			ID:           3001,
			Name:         "电销待转化客户",
			CreateUserID: 2,
			IsInPool:     false,
		},
	}
	svc := &customerService{repo: repoStub}

	_, err := svc.ConvertCustomer(context.Background(), 3001, 2)
	if err != nil {
		t.Fatalf("ConvertCustomer returned error: %v", err)
	}
	if !repoStub.convertCalled {
		t.Fatalf("expected Convert repository method to be called")
	}
	if repoStub.convertOwner != 2 {
		t.Fatalf("expected inside-sales owner 2, got %d", repoStub.convertOwner)
	}
}

func TestConvertCustomerInsideSalesFallsBackToRecentContractExemptOwner(t *testing.T) {
	repoStub := &customerScopeRepoStub{
		userRoles: map[int64]string{
			1: "sales_director",
			2: "sales_inside",
			3: "sales_staff",
			4: "sales_staff",
		},
		parents: map[int64]int64{
			2: 1,
			3: 1,
			4: 1,
		},
		subordinates: map[int64][]int64{
			1: {2, 3, 4},
		},
		roleUsers: map[string][]int64{
			"sales_director": {1},
			"sales_staff":    {3, 4},
		},
		recentContractExemptUserIDs: []int64{3, 4},
		findByID: &model.Customer{
			ID:           3004,
			Name:         "签单豁免待转化客户",
			CreateUserID: 2,
			IsInPool:     true,
		},
		convertResult: &model.Customer{
			ID:           3004,
			Name:         "签单豁免待转化客户",
			CreateUserID: 2,
			IsInPool:     false,
		},
	}
	svc := &customerService{repo: repoStub}

	_, err := svc.ConvertCustomer(context.Background(), 3004, 2)
	if err != nil {
		t.Fatalf("ConvertCustomer returned error: %v", err)
	}
	if repoStub.convertOwner != 3 {
		t.Fatalf("expected recent contract exempt convert owner 3, got %d", repoStub.convertOwner)
	}
}

func TestReassignCustomersByYesterdayRankingUsesDepartmentRanking(t *testing.T) {
	repoStub := &customerScopeRepoStub{
		roleUsers: map[string][]int64{
			"sales_director": {1},
			"sales_staff":    {3, 4},
		},
		userRoles: map[int64]string{
			1: "sales_director",
			2: "sales_staff",
			3: "sales_staff",
			4: "sales_staff",
		},
		parents: map[int64]int64{
			2: 1,
			3: 1,
			4: 1,
		},
		subordinates: map[int64][]int64{
			1: {2, 3, 4},
		},
		rankedUserIDs: []int64{4, 3, 1},
		findByID: &model.Customer{
			ID:          3101,
			Name:        "待重分配客户",
			OwnerUserID: int64Ptr(2),
		},
	}
	svc := &customerService{repo: repoStub}

	result, err := svc.ReassignCustomersByYesterdayRanking(context.Background(), model.CustomerBatchRankedReassignInput{
		CustomerIDs:    []int64{3101},
		OperatorUserID: 2,
	})
	if err != nil {
		t.Fatalf("ReassignCustomersByYesterdayRanking returned error: %v", err)
	}
	if result.SuccessCount != 1 || result.FailedCount != 0 {
		t.Fatalf("unexpected result counts: %+v", result)
	}
	if len(repoStub.transferInputs) != 1 {
		t.Fatalf("expected 1 transfer, got %d", len(repoStub.transferInputs))
	}
	if repoStub.transferInputs[0].ToOwnerUserID != 4 {
		t.Fatalf("expected reassigned owner 4, got %d", repoStub.transferInputs[0].ToOwnerUserID)
	}
	if repoStub.lastRankReferenceDate != previousAutoAssignScoreDate() {
		t.Fatalf("expected yesterday reference date %q, got %q", previousAutoAssignScoreDate(), repoStub.lastRankReferenceDate)
	}
}

func TestReassignCustomersByYesterdayRankingStartsFromTopRankInsteadOfContinuingRotation(t *testing.T) {
	latestOwnerUserID := int64(3)
	repoStub := &customerScopeRepoStub{
		roleUsers: map[string][]int64{
			"sales_director": {1},
			"sales_staff":    {3, 4},
		},
		userRoles: map[int64]string{
			1: "sales_director",
			2: "sales_staff",
			3: "sales_staff",
			4: "sales_staff",
		},
		parents: map[int64]int64{
			2: 1,
			3: 1,
			4: 1,
		},
		subordinates: map[int64][]int64{
			1: {2, 3, 4},
		},
		rankedUserIDs:               []int64{4, 3, 1},
		latestAutoAssignOwnerUserID: &latestOwnerUserID,
		findByID: &model.Customer{
			ID:          3102,
			Name:        "按排序重新分配客户",
			OwnerUserID: int64Ptr(2),
		},
	}
	svc := &customerService{repo: repoStub}

	result, err := svc.ReassignCustomersByYesterdayRanking(context.Background(), model.CustomerBatchRankedReassignInput{
		CustomerIDs:    []int64{3102},
		OperatorUserID: 2,
	})
	if err != nil {
		t.Fatalf("ReassignCustomersByYesterdayRanking returned error: %v", err)
	}
	if result.SuccessCount != 1 || result.FailedCount != 0 {
		t.Fatalf("unexpected result counts: %+v", result)
	}
	if len(repoStub.transferInputs) != 1 {
		t.Fatalf("expected 1 transfer, got %d", len(repoStub.transferInputs))
	}
	if repoStub.transferInputs[0].ToOwnerUserID != 4 {
		t.Fatalf("expected reassigned owner 4, got %d", repoStub.transferInputs[0].ToOwnerUserID)
	}
}

func TestConvertCustomerRejectsAlreadyConvertedLead(t *testing.T) {
	repoStub := &customerScopeRepoStub{
		userRoles: map[int64]string{
			2: "sales_inside",
		},
		findByID: &model.Customer{
			ID:           3002,
			Name:         "已转化客户",
			CreateUserID: 2,
			IsInPool:     true,
			ConvertedAt:  timePtr(time.Now().UTC()),
		},
	}
	svc := &customerService{repo: repoStub}

	_, err := svc.ConvertCustomer(context.Background(), 3002, 2)
	if !errors.Is(err, ErrCustomerConvertForbidden) {
		t.Fatalf("expected ErrCustomerConvertForbidden, got %v", err)
	}
	if repoStub.convertCalled {
		t.Fatalf("convert repository should not be called for already converted lead")
	}
}

func TestConvertCustomerAllowsAdminToConvertInsideSalesLead(t *testing.T) {
	repoStub := &customerScopeRepoStub{
		userRoles: map[int64]string{
			1: "sales_director",
			2: "sales_inside",
			9: "admin",
		},
		parents: map[int64]int64{
			2: 1,
		},
		subordinates: map[int64][]int64{
			1: {2},
		},
		roleUsers: map[string][]int64{
			"sales_director": {1},
		},
		findByID: &model.Customer{
			ID:           3003,
			Name:         "管理员代转化客户",
			CreateUserID: 2,
			IsInPool:     true,
		},
		convertResult: &model.Customer{
			ID:           3003,
			Name:         "管理员代转化客户",
			CreateUserID: 2,
			IsInPool:     false,
		},
	}
	svc := &customerService{repo: repoStub}

	_, err := svc.ConvertCustomer(context.Background(), 3003, 9)
	if err != nil {
		t.Fatalf("ConvertCustomer(admin) returned error: %v", err)
	}
	if !repoStub.convertCalled {
		t.Fatalf("expected admin convert to call repository")
	}
	if repoStub.convertOwner != 2 {
		t.Fatalf("expected admin convert to keep customer on creator 2 without scores, got %d", repoStub.convertOwner)
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

func timePtr(value time.Time) *time.Time {
	return &value
}
