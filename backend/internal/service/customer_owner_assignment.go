package service

import (
	"backend/internal/model"
	"context"
	"os"
	"sort"
	"strings"
	"time"
)

var assignableSalesOwnerRoleNames = []string{
	"sales_director", "sales_manager", "sales_staff", roleSalesOutside, "sale_outside",
	"销售总监", "销售经理", "销售员工", "销售", "Outside销售", "outside销售",
}

const autoAssignMinimumDailyScore = 80

const (
	autoAssignDirectorFallbackNicknameGe = "葛鹏辉"
	autoAssignDirectorFallbackNicknameLi = "李龙泉"
)

type customerOwnerAssignmentRepo interface {
	GetUserRoleName(ctx context.Context, userID int64) (string, error)
	GetUserDisplayName(ctx context.Context, userID int64) (string, error)
	GetParentUserID(ctx context.Context, userID int64) (int64, error)
	ListEnabledUserIDsByRoleNames(ctx context.Context, roleNames []string) ([]int64, error)
	ListDirectSubordinateUserIDsByRoleNames(ctx context.Context, parentIDs []int64, roleNames []string) ([]int64, error)
	ListAutoAssignRankedOwnerScores(ctx context.Context, referenceDate string, userIDs []int64) ([]model.SalesDailyScore, error)
	ListRecentContractExemptOwnerUserIDs(ctx context.Context, since time.Time, userIDs []int64) ([]int64, error)
	FindEnabledUserIDByNickname(ctx context.Context, nickname string) (int64, error)
	FindLatestAutoAssignOwnerUserID(ctx context.Context, ownerUserIDs []int64, since time.Time) (*int64, error)
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

	rankedOwnerUserID, ok, err := pickRankedSalesOwnerUserID(ctx, repo, candidateUserIDs)
	if err != nil {
		return 0, err
	}
	if ok {
		return rankedOwnerUserID, nil
	}

	exemptOwnerUserID, ok, err := pickRecentContractExemptOwnerUserID(ctx, repo, candidateUserIDs)
	if err != nil {
		return 0, err
	}
	if ok {
		return exemptOwnerUserID, nil
	}

	crossDirectorUserID, ok, err := pickCrossDepartmentFallbackOwnerUserID(ctx, repo, directorUserID)
	if err != nil {
		return 0, err
	}
	if ok {
		return crossDirectorUserID, nil
	}
	return 0, nil
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

func pickRankedSalesOwnerUserID(ctx context.Context, repo customerOwnerAssignmentRepo, candidateUserIDs []int64) (int64, bool, error) {
	return pickRankedSalesOwnerUserIDByDate(ctx, repo, candidateUserIDs, previousAutoAssignScoreDate(), currentAutoAssignRotationStart())
}

func pickRankedSalesOwnerUserIDByDate(
	ctx context.Context,
	repo customerOwnerAssignmentRepo,
	candidateUserIDs []int64,
	referenceDate string,
	rotationStart time.Time,
) (int64, bool, error) {
	candidateUserIDs = uniquePositiveInt64(candidateUserIDs)
	if len(candidateUserIDs) == 0 {
		return 0, false, nil
	}

	rankedScores, err := repo.ListAutoAssignRankedOwnerScores(ctx, referenceDate, candidateUserIDs)
	if err != nil {
		return 0, false, err
	}
	eligibleRankedUserIDs := filterAutoAssignEligibleOwnerUserIDs(rankedScores)
	if len(eligibleRankedUserIDs) == 0 {
		return 0, false, nil
	}
	if len(eligibleRankedUserIDs) == 1 {
		return eligibleRankedUserIDs[0], true, nil
	}

	latestOwnerUserID, err := repo.FindLatestAutoAssignOwnerUserID(ctx, eligibleRankedUserIDs, rotationStart)
	if err != nil {
		return 0, false, err
	}
	if latestOwnerUserID == nil || *latestOwnerUserID <= 0 {
		return eligibleRankedUserIDs[0], true, nil
	}

	for idx, userID := range eligibleRankedUserIDs {
		if userID != *latestOwnerUserID {
			continue
		}
		return eligibleRankedUserIDs[(idx+1)%len(eligibleRankedUserIDs)], true, nil
	}

	return eligibleRankedUserIDs[0], true, nil
}

func previousAutoAssignScoreDate() string {
	location := autoAssignBusinessLocation()
	return time.Now().In(location).AddDate(0, 0, -1).Format("2006-01-02")
}

func pickRecentContractExemptOwnerUserID(ctx context.Context, repo customerOwnerAssignmentRepo, candidateUserIDs []int64) (int64, bool, error) {
	candidateUserIDs = uniquePositiveInt64(candidateUserIDs)
	if len(candidateUserIDs) == 0 {
		return 0, false, nil
	}

	exemptOwnerUserIDs, err := repo.ListRecentContractExemptOwnerUserIDs(ctx, recentAutoAssignContractExemptionSince(), candidateUserIDs)
	if err != nil {
		return 0, false, err
	}
	if len(exemptOwnerUserIDs) == 0 {
		return 0, false, nil
	}
	return exemptOwnerUserIDs[0], true, nil
}

func pickCrossDepartmentFallbackOwnerUserID(ctx context.Context, repo customerOwnerAssignmentRepo, directorUserID int64) (int64, bool, error) {
	if directorUserID <= 0 {
		return 0, false, nil
	}

	directorName, err := repo.GetUserDisplayName(ctx, directorUserID)
	if err != nil {
		return 0, false, err
	}

	counterpartNickname := ""
	switch strings.TrimSpace(directorName) {
	case autoAssignDirectorFallbackNicknameLi:
		counterpartNickname = autoAssignDirectorFallbackNicknameGe
	case autoAssignDirectorFallbackNicknameGe:
		counterpartNickname = autoAssignDirectorFallbackNicknameLi
	default:
		return 0, false, nil
	}

	counterpartUserID, err := repo.FindEnabledUserIDByNickname(ctx, counterpartNickname)
	if err != nil {
		return 0, false, err
	}
	if counterpartUserID <= 0 {
		return 0, false, nil
	}
	return counterpartUserID, true, nil
}

func filterAutoAssignEligibleOwnerUserIDs(rankedScores []model.SalesDailyScore) []int64 {
	if len(rankedScores) == 0 {
		return []int64{}
	}

	eligibleUserIDs := make([]int64, 0, len(rankedScores))
	for _, score := range rankedScores {
		if score.UserID <= 0 || score.TotalScore < autoAssignMinimumDailyScore {
			continue
		}
		eligibleUserIDs = append(eligibleUserIDs, score.UserID)
	}
	if len(eligibleUserIDs) == 0 {
		return []int64{}
	}

	// If everyone meets the minimum score and there are at least two ranked owners,
	// exclude the last-ranked owner.
	if len(eligibleUserIDs) == len(rankedScores) && len(eligibleUserIDs) > 1 {
		return eligibleUserIDs[:len(eligibleUserIDs)-1]
	}

	return eligibleUserIDs
}

func currentAutoAssignScoreDate() string {
	location := autoAssignBusinessLocation()
	return time.Now().In(location).Format("2006-01-02")
}

func currentAutoAssignRotationStart() time.Time {
	location := autoAssignBusinessLocation()
	now := time.Now().In(location)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location).UTC()
}

func recentAutoAssignContractExemptionSince() time.Time {
	location := autoAssignBusinessLocation()
	now := time.Now().In(location)
	weekdayOffset := (int(now.Weekday()) + 6) % 7
	weekStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location).AddDate(0, 0, -weekdayOffset)
	return weekStart.UTC()
}

func autoAssignBusinessLocation() *time.Location {
	locationName := strings.TrimSpace(os.Getenv("SCHEDULE_TIMEZONE"))
	if locationName == "" {
		locationName = "Asia/Shanghai"
	}
	location, err := time.LoadLocation(locationName)
	if err != nil {
		return time.Local
	}
	return location
}
