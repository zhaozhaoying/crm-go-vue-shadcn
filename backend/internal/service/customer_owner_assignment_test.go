package service

import (
	"context"
	"testing"
)

type customerOwnerAssignmentRepoStub struct {
	roles        map[int64]string
	parents      map[int64]int64
	enabledUsers map[int64]bool
	subordinates map[int64][]int64
	ownedCount   map[int64]int64
}

func (s *customerOwnerAssignmentRepoStub) GetUserRoleName(_ context.Context, userID int64) (string, error) {
	return s.roles[userID], nil
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

func TestPickBalancedSalesOwnerUserIDKeepsTeamBoundary(t *testing.T) {
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
	if ownerUserID != 3 {
		t.Fatalf("expected owner 3 from same director team, got %d", ownerUserID)
	}
}

func TestPickBalancedSalesOwnerUserIDDistributesEvenlyByLoad(t *testing.T) {
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
		ownedCount: map[int64]int64{
			1: 0,
			3: 0,
			4: 0,
		},
	}

	got := make([]int64, 0, 3)
	for i := 0; i < 3; i++ {
		ownerUserID, err := pickBalancedSalesOwnerUserID(context.Background(), repo, 2)
		if err != nil {
			t.Fatalf("pickBalancedSalesOwnerUserID returned error: %v", err)
		}
		got = append(got, ownerUserID)
		repo.ownedCount[ownerUserID]++
	}

	want := []int64{1, 3, 4}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected evenly distributed owners %v, got %v", want, got)
		}
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
