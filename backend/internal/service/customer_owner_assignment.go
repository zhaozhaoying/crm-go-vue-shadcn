package service

import (
	"context"
	"sort"
)

var assignableSalesOwnerRoleNames = []string{
	"sales_director", "sales_manager", "sales_staff", roleSalesOutside,
	"销售总监", "销售经理", "销售员工", "销售", "Outside销售", "outside销售",
}

type customerOwnerAssignmentRepo interface {
	GetUserRoleName(ctx context.Context, userID int64) (string, error)
	GetParentUserID(ctx context.Context, userID int64) (int64, error)
	ListEnabledUserIDsByRoleNames(ctx context.Context, roleNames []string) ([]int64, error)
	ListDirectSubordinateUserIDsByRoleNames(ctx context.Context, parentIDs []int64, roleNames []string) ([]int64, error)
	CountOwnedActiveByOwner(ctx context.Context, ownerUserID int64) (int64, error)
}

func pickBalancedSalesOwnerUserID(ctx context.Context, repo customerOwnerAssignmentRepo, operatorUserID int64) (int64, error) {
	directorUserID, err := resolveSalesDirectorUserID(ctx, repo, operatorUserID)
	if err != nil {
		return 0, err
	}
	if directorUserID <= 0 {
		return 0, ErrCustomerNoOutsideSalesAvailable
	}

	candidateUserIDs, err := listAssignableSalesOwnerUserIDs(ctx, repo, directorUserID)
	if err != nil {
		return 0, err
	}
	if len(candidateUserIDs) == 0 {
		return 0, ErrCustomerNoOutsideSalesAvailable
	}

	return pickLeastLoadedOwnerUserID(ctx, repo, candidateUserIDs)
}

func resolveSalesDirectorUserID(ctx context.Context, repo customerOwnerAssignmentRepo, userID int64) (int64, error) {
	if userID <= 0 {
		return 0, nil
	}

	visited := map[int64]struct{}{}
	currentUserID := userID
	for currentUserID > 0 {
		if _, seen := visited[currentUserID]; seen {
			return 0, nil
		}
		visited[currentUserID] = struct{}{}

		roleName, err := repo.GetUserRoleName(ctx, currentUserID)
		if err != nil {
			return 0, err
		}
		if isRole(roleName, "sales_director", "销售总监") {
			return currentUserID, nil
		}

		parentUserID, err := repo.GetParentUserID(ctx, currentUserID)
		if err != nil {
			return 0, err
		}
		currentUserID = parentUserID
	}

	return 0, nil
}

func listAssignableSalesOwnerUserIDs(ctx context.Context, repo customerOwnerAssignmentRepo, directorUserID int64) ([]int64, error) {
	descendantUserIDs, err := listAllDescendantUserIDsByRepo(ctx, repo, directorUserID)
	if err != nil {
		return nil, err
	}

	teamUserIDs := uniquePositiveInt64(append([]int64{directorUserID}, descendantUserIDs...))
	if len(teamUserIDs) == 0 {
		return []int64{}, nil
	}

	assignableUserIDs, err := repo.ListEnabledUserIDsByRoleNames(ctx, assignableSalesOwnerRoleNames)
	if err != nil {
		return nil, err
	}

	return intersectPositiveInt64(teamUserIDs, assignableUserIDs), nil
}

func listAllDescendantUserIDsByRepo(ctx context.Context, repo customerOwnerAssignmentRepo, rootUserID int64) ([]int64, error) {
	if rootUserID <= 0 {
		return []int64{}, nil
	}

	visited := map[int64]struct{}{rootUserID: {}}
	queue := []int64{rootUserID}
	result := make([]int64, 0, 8)

	for len(queue) > 0 {
		nextLevel, err := repo.ListDirectSubordinateUserIDsByRoleNames(ctx, queue, nil)
		if err != nil {
			return nil, err
		}
		if len(nextLevel) == 0 {
			break
		}

		nextQueue := make([]int64, 0, len(nextLevel))
		for _, id := range nextLevel {
			if id <= 0 {
				continue
			}
			if _, seen := visited[id]; seen {
				continue
			}
			visited[id] = struct{}{}
			result = append(result, id)
			nextQueue = append(nextQueue, id)
		}
		queue = nextQueue
	}

	return result, nil
}

func pickLeastLoadedOwnerUserID(ctx context.Context, repo customerOwnerAssignmentRepo, candidateUserIDs []int64) (int64, error) {
	candidateUserIDs = uniquePositiveInt64(candidateUserIDs)
	if len(candidateUserIDs) == 0 {
		return 0, ErrCustomerNoOutsideSalesAvailable
	}

	sort.Slice(candidateUserIDs, func(i, j int) bool {
		return candidateUserIDs[i] < candidateUserIDs[j]
	})

	var selectedUserID int64
	var selectedOwnedCount int64
	found := false

	for _, userID := range candidateUserIDs {
		ownedCount, err := repo.CountOwnedActiveByOwner(ctx, userID)
		if err != nil {
			return 0, err
		}
		if !found || ownedCount < selectedOwnedCount || (ownedCount == selectedOwnedCount && userID < selectedUserID) {
			selectedUserID = userID
			selectedOwnedCount = ownedCount
			found = true
		}
	}

	if !found {
		return 0, ErrCustomerNoOutsideSalesAvailable
	}
	return selectedUserID, nil
}
