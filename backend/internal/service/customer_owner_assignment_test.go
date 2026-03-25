package service

import (
	"backend/internal/model"
	"context"
	"testing"
	"time"
)

type customerOwnerAssignmentRepoStub struct {
	roles                       map[int64]string
	displayNames                map[int64]string
	parents                     map[int64]int64
	enabledUsers                map[int64]bool
	subordinates                map[int64][]int64
	ownedCount                  map[int64]int64
	rankedScores                []model.SalesDailyScore
	recentContractExemptUserIDs []int64
	lastRecentContractSince     time.Time
	enabledUserIDsByNickname    map[string]int64
	latestAutoAssignOwnerUserID *int64
}

func (s *customerOwnerAssignmentRepoStub) GetUserRoleName(_ context.Context, userID int64) (string, error) {
	return s.roles[userID], nil
}

func (s *customerOwnerAssignmentRepoStub) GetUserDisplayName(_ context.Context, userID int64) (string, error) {
	if s.displayNames != nil {
		return s.displayNames[userID], nil
	}
	return "", nil
}

func (s *customerOwnerAssignmentRepoStub) GetParentUserID(_ context.Context, userID int64) (int64, error) {
	return s.parents[userID], nil
}

func (s *customerOwnerAssignmentRepoStub) ListEnabledUserIDsByRoleNames(_ context.Context, roleNames []string) ([]int64, error) {
	allowedRoles := make(map[string]struct{}, len(roleNames))
	for _, roleName := range roleNames {
		allowedRoles[roleName] = struct{}{}
	}

	result := make([]int64, 0)
	for userID, enabled := range s.enabledUsers {
		if !enabled {
			continue
		}
		if _, ok := allowedRoles[s.roles[userID]]; !ok {
			continue
		}
		result = append(result, userID)
	}
	return result, nil
}

func (s *customerOwnerAssignmentRepoStub) ListDirectSubordinateUserIDsByRoleNames(_ context.Context, parentIDs []int64, roleNames []string) ([]int64, error) {
	var roleFilter map[string]struct{}
	if len(roleNames) > 0 {
		roleFilter = make(map[string]struct{}, len(roleNames))
		for _, roleName := range roleNames {
			roleFilter[roleName] = struct{}{}
		}
	}

	result := make([]int64, 0)
	for _, parentID := range parentIDs {
		for _, userID := range s.subordinates[parentID] {
			if roleFilter != nil {
				if _, ok := roleFilter[s.roles[userID]]; !ok {
					continue
				}
			}
			result = append(result, userID)
		}
	}
	return result, nil
}

func (s *customerOwnerAssignmentRepoStub) CountOwnedActiveByOwner(_ context.Context, ownerUserID int64) (int64, error) {
	return s.ownedCount[ownerUserID], nil
}

func (s *customerOwnerAssignmentRepoStub) ListAutoAssignRankedOwnerScores(_ context.Context, _ string, userIDs []int64) ([]model.SalesDailyScore, error) {
	if len(s.rankedScores) == 0 {
		return []model.SalesDailyScore{}, nil
	}
	allowed := make(map[int64]struct{}, len(userIDs))
	for _, userID := range userIDs {
		allowed[userID] = struct{}{}
	}
	result := make([]model.SalesDailyScore, 0, len(s.rankedScores))
	for _, score := range s.rankedScores {
		if _, ok := allowed[score.UserID]; ok {
			result = append(result, score)
		}
	}
	return result, nil
}

func (s *customerOwnerAssignmentRepoStub) listRecentContractExemptOwnerUserIDs(userIDs []int64) []int64 {
	if len(s.recentContractExemptUserIDs) == 0 {
		return []int64{}
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
	return result
}

func (s *customerOwnerAssignmentRepoStub) ListRecentContractExemptOwnerUserIDs(_ context.Context, since time.Time, userIDs []int64) ([]int64, error) {
	s.lastRecentContractSince = since
	return s.listRecentContractExemptOwnerUserIDs(userIDs), nil
}

func (s *customerOwnerAssignmentRepoStub) FindEnabledUserIDByNickname(_ context.Context, nickname string) (int64, error) {
	if s.enabledUserIDsByNickname == nil {
		return 0, nil
	}
	return s.enabledUserIDsByNickname[nickname], nil
}

func (s *customerOwnerAssignmentRepoStub) FindLatestAutoAssignOwnerUserID(_ context.Context, ownerUserIDs []int64, _ time.Time) (*int64, error) {
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

func TestPickBalancedSalesOwnerUserIDReturnsZeroWithoutRankedScoresAcrossTeam(t *testing.T) {
	repo := &customerOwnerAssignmentRepoStub{
		roles: map[int64]string{
			1:  "sales_director",
			2:  "sales_inside",
			3:  "sales_staff",
			10: "sales_director",
			11: "sales_staff",
		},
		parents: map[int64]int64{
			2:  1,
			3:  1,
			11: 10,
		},
		enabledUsers: map[int64]bool{
			1:  true,
			2:  true,
			3:  true,
			10: true,
			11: true,
		},
		subordinates: map[int64][]int64{
			1:  {2, 3},
			10: {11},
		},
		ownedCount: map[int64]int64{
			1:  2,
			3:  0,
			10: 0,
			11: 0,
		},
	}

	ownerUserID, err := pickBalancedSalesOwnerUserID(context.Background(), repo, 2)
	if err != nil {
		t.Fatalf("pickBalancedSalesOwnerUserID returned error: %v", err)
	}
	if ownerUserID != 0 {
		t.Fatalf("expected no assignment without ranked scores, got %d", ownerUserID)
	}
}

func TestPickBalancedSalesOwnerUserIDReturnsZeroWithoutAnyScoreData(t *testing.T) {
	repo := &customerOwnerAssignmentRepoStub{
		roles: map[int64]string{
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
		enabledUsers: map[int64]bool{
			1: true,
			2: true,
			3: true,
			4: true,
		},
		subordinates: map[int64][]int64{
			1: {2, 3, 4},
		},
	}
	ownerUserID, err := pickBalancedSalesOwnerUserID(context.Background(), repo, 2)
	if err != nil {
		t.Fatalf("pickBalancedSalesOwnerUserID returned error: %v", err)
	}
	if ownerUserID != 0 {
		t.Fatalf("expected no assignment without any ranked scores, got %d", ownerUserID)
	}
}

func TestRecentAutoAssignContractExemptionSinceStartsAtCurrentWeekMonday(t *testing.T) {
	since := recentAutoAssignContractExemptionSince().In(autoAssignBusinessLocation())
	if since.Weekday() != time.Monday {
		t.Fatalf("expected monday start, got %s", since.Weekday())
	}
	if since.Hour() != 0 || since.Minute() != 0 || since.Second() != 0 || since.Nanosecond() != 0 {
		t.Fatalf("expected start of day, got %v", since)
	}
}

func TestPickBalancedSalesOwnerUserIDFallsBackToRecentContractExemptOwnerWhenNoOneReachesThreshold(t *testing.T) {
	repo := &customerOwnerAssignmentRepoStub{
		roles: map[int64]string{
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
		enabledUsers: map[int64]bool{
			1: true,
			2: true,
			3: true,
			4: true,
		},
		subordinates: map[int64][]int64{
			1: {2, 3, 4},
		},
		rankedScores: []model.SalesDailyScore{
			{UserID: 3, TotalScore: 79},
			{UserID: 4, TotalScore: 60},
		},
		recentContractExemptUserIDs: []int64{4, 3},
	}

	ownerUserID, err := pickBalancedSalesOwnerUserID(context.Background(), repo, 2)
	if err != nil {
		t.Fatalf("pickBalancedSalesOwnerUserID returned error: %v", err)
	}
	if ownerUserID != 4 {
		t.Fatalf("expected recent contract exempt owner 4, got %d", ownerUserID)
	}
	if !repo.lastRecentContractSince.Equal(recentAutoAssignContractExemptionSince()) {
		t.Fatalf("expected current-week exemption since %v, got %v", recentAutoAssignContractExemptionSince(), repo.lastRecentContractSince)
	}
}

func TestPickBalancedSalesOwnerUserIDUsesHighestRankedOwnerWhenScoresMeetThreshold(t *testing.T) {
	repo := &customerOwnerAssignmentRepoStub{
		roles: map[int64]string{
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
		enabledUsers: map[int64]bool{
			1: true,
			2: true,
			3: true,
			4: true,
		},
		subordinates: map[int64][]int64{
			1: {2, 3, 4},
		},
		rankedScores: []model.SalesDailyScore{
			{UserID: 4, TotalScore: 95},
			{UserID: 3, TotalScore: 82},
			{UserID: 1, TotalScore: 79},
		},
	}

	ownerUserID, err := pickBalancedSalesOwnerUserID(context.Background(), repo, 2)
	if err != nil {
		t.Fatalf("pickBalancedSalesOwnerUserID returned error: %v", err)
	}
	if ownerUserID != 4 {
		t.Fatalf("expected highest ranked owner 4, got %d", ownerUserID)
	}
}

func TestPickBalancedSalesOwnerUserIDUsesEarliestReachedAtWhenScoresTie(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	repo := &customerOwnerAssignmentRepoStub{
		roles: map[int64]string{
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
		enabledUsers: map[int64]bool{
			1: true,
			2: true,
			3: true,
			4: true,
		},
		subordinates: map[int64][]int64{
			1: {2, 3, 4},
		},
		rankedScores: []model.SalesDailyScore{
			{UserID: 3, TotalScore: 80, ScoreReachedAt: timePtr(time.Date(2026, 3, 25, 8, 0, 0, 0, loc).UTC())},
			{UserID: 4, TotalScore: 80, ScoreReachedAt: timePtr(time.Date(2026, 3, 25, 8, 1, 0, 0, loc).UTC())},
			{UserID: 1, TotalScore: 80, ScoreReachedAt: timePtr(time.Date(2026, 3, 25, 9, 0, 0, 0, loc).UTC())},
		},
	}

	ownerUserID, err := pickBalancedSalesOwnerUserID(context.Background(), repo, 2)
	if err != nil {
		t.Fatalf("pickBalancedSalesOwnerUserID returned error: %v", err)
	}
	if ownerUserID != 3 {
		t.Fatalf("expected earliest reached tied owner 3, got %d", ownerUserID)
	}
}

func TestPickBalancedSalesOwnerUserIDPrefersRankedPoolOverRecentContractExemption(t *testing.T) {
	repo := &customerOwnerAssignmentRepoStub{
		roles: map[int64]string{
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
		enabledUsers: map[int64]bool{
			1: true,
			2: true,
			3: true,
			4: true,
		},
		subordinates: map[int64][]int64{
			1: {2, 3, 4},
		},
		rankedScores: []model.SalesDailyScore{
			{UserID: 4, TotalScore: 92},
			{UserID: 3, TotalScore: 85},
		},
		recentContractExemptUserIDs: []int64{3},
	}

	ownerUserID, err := pickBalancedSalesOwnerUserID(context.Background(), repo, 2)
	if err != nil {
		t.Fatalf("pickBalancedSalesOwnerUserID returned error: %v", err)
	}
	if ownerUserID != 4 {
		t.Fatalf("expected ranked owner 4 to win before contract exemption, got %d", ownerUserID)
	}
}

func TestPickBalancedSalesOwnerUserIDFallsBackToCounterpartDirectorWhenLiTeamHasNoThresholdOrContracts(t *testing.T) {
	repo := &customerOwnerAssignmentRepoStub{
		roles: map[int64]string{
			1: "sales_director",
			2: "sales_inside",
			3: "sales_staff",
			7: "sales_director",
		},
		displayNames: map[int64]string{
			1: "李龙泉",
			7: "葛鹏辉",
		},
		parents: map[int64]int64{
			2: 1,
			3: 1,
		},
		enabledUsers: map[int64]bool{
			1: true,
			2: true,
			3: true,
			7: true,
		},
		subordinates: map[int64][]int64{
			1: {2, 3},
		},
		enabledUserIDsByNickname: map[string]int64{
			"葛鹏辉": 7,
		},
	}

	ownerUserID, err := pickBalancedSalesOwnerUserID(context.Background(), repo, 2)
	if err != nil {
		t.Fatalf("pickBalancedSalesOwnerUserID returned error: %v", err)
	}
	if ownerUserID != 7 {
		t.Fatalf("expected counterpart director 7, got %d", ownerUserID)
	}
}

func TestPickBalancedSalesOwnerUserIDFallsBackToCounterpartDirectorWhenGeTeamHasNoThresholdOrContracts(t *testing.T) {
	repo := &customerOwnerAssignmentRepoStub{
		roles: map[int64]string{
			1: "sales_director",
			2: "sales_inside",
			3: "sales_staff",
			7: "sales_director",
		},
		displayNames: map[int64]string{
			1: "葛鹏辉",
			7: "李龙泉",
		},
		parents: map[int64]int64{
			2: 1,
			3: 1,
		},
		enabledUsers: map[int64]bool{
			1: true,
			2: true,
			3: true,
			7: true,
		},
		subordinates: map[int64][]int64{
			1: {2, 3},
		},
		enabledUserIDsByNickname: map[string]int64{
			"李龙泉": 7,
		},
	}

	ownerUserID, err := pickBalancedSalesOwnerUserID(context.Background(), repo, 2)
	if err != nil {
		t.Fatalf("pickBalancedSalesOwnerUserID returned error: %v", err)
	}
	if ownerUserID != 7 {
		t.Fatalf("expected counterpart director 7, got %d", ownerUserID)
	}
}

func TestPickBalancedSalesOwnerUserIDRoundsRobinAcrossRankedOwnersExcludingLast(t *testing.T) {
	latestOwnerUserID := int64(4)
	repo := &customerOwnerAssignmentRepoStub{
		roles: map[int64]string{
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
		enabledUsers: map[int64]bool{
			1: true,
			2: true,
			3: true,
			4: true,
		},
		subordinates: map[int64][]int64{
			1: {2, 3, 4},
		},
		rankedScores: []model.SalesDailyScore{
			{UserID: 4, TotalScore: 95},
			{UserID: 3, TotalScore: 88},
			{UserID: 1, TotalScore: 81},
		},
		latestAutoAssignOwnerUserID: &latestOwnerUserID,
	}

	ownerUserID, err := pickBalancedSalesOwnerUserID(context.Background(), repo, 2)
	if err != nil {
		t.Fatalf("pickBalancedSalesOwnerUserID returned error: %v", err)
	}
	if ownerUserID != 3 {
		t.Fatalf("expected next eligible owner 3 after owner 4, got %d", ownerUserID)
	}
}

func TestPickBalancedSalesOwnerUserIDUsesQualifiedSubsetWhenOnlyPartOfTeamReachesThreshold(t *testing.T) {
	repo := &customerOwnerAssignmentRepoStub{
		roles: map[int64]string{
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
		enabledUsers: map[int64]bool{
			1: true,
			2: true,
			3: true,
			4: true,
		},
		subordinates: map[int64][]int64{
			1: {2, 3, 4},
		},
		rankedScores: []model.SalesDailyScore{
			{UserID: 3, TotalScore: 85},
			{UserID: 4, TotalScore: 79},
			{UserID: 1, TotalScore: 60},
		},
	}

	ownerUserID, err := pickBalancedSalesOwnerUserID(context.Background(), repo, 2)
	if err != nil {
		t.Fatalf("pickBalancedSalesOwnerUserID returned error: %v", err)
	}
	if ownerUserID != 3 {
		t.Fatalf("expected ranked owner 3 to win, got %d", ownerUserID)
	}
}

func TestPickBalancedSalesOwnerUserIDKeepsSingleRankedOwnerWhenAllReachThreshold(t *testing.T) {
	repo := &customerOwnerAssignmentRepoStub{
		roles: map[int64]string{
			1: "sales_director",
			2: "sales_inside",
		},
		parents: map[int64]int64{
			2: 1,
		},
		enabledUsers: map[int64]bool{
			1: true,
			2: true,
		},
		subordinates: map[int64][]int64{
			1: {2},
		},
		rankedScores: []model.SalesDailyScore{
			{UserID: 1, TotalScore: 88},
		},
	}

	ownerUserID, err := pickBalancedSalesOwnerUserID(context.Background(), repo, 2)
	if err != nil {
		t.Fatalf("pickBalancedSalesOwnerUserID returned error: %v", err)
	}
	if ownerUserID != 1 {
		t.Fatalf("expected single ranked owner 1 to remain assignable, got %d", ownerUserID)
	}
}

func TestPickBalancedSalesOwnerUserIDKeepsSingleDirectorTeamFallback(t *testing.T) {
	repo := &customerOwnerAssignmentRepoStub{
		roles: map[int64]string{
			1: "sales_director",
			2: "sales_inside",
		},
		parents: map[int64]int64{
			2: 1,
		},
		enabledUsers: map[int64]bool{
			1: true,
			2: true,
		},
		subordinates: map[int64][]int64{
			1: {2},
		},
	}

	ownerUserID, err := pickBalancedSalesOwnerUserID(context.Background(), repo, 2)
	if err != nil {
		t.Fatalf("pickBalancedSalesOwnerUserID returned error: %v", err)
	}
	if ownerUserID != 0 {
		t.Fatalf("expected no assignment for single-director team without scores, got %d", ownerUserID)
	}
}

func TestResolveSalesDirectorUserIDReturnsZeroWithoutDirectorAncestor(t *testing.T) {
	repo := &customerOwnerAssignmentRepoStub{
		roles: map[int64]string{
			2: "sales_inside",
			3: "sales_staff",
		},
		parents: map[int64]int64{
			2: 3,
		},
	}

	directorUserID, err := resolveSalesDirectorUserID(context.Background(), repo, 2)
	if err != nil {
		t.Fatalf("resolveSalesDirectorUserID returned error: %v", err)
	}
	if directorUserID != 0 {
		t.Fatalf("expected no director ancestor, got %d", directorUserID)
	}
}
