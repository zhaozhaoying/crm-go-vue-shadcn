package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/scoring"
	"context"
	"crypto/rand"
	"encoding/hex"
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

const defaultMiHuaPageSize = 100

var (
	ErrSalesDailyScoreNotFound          = errors.New("sales daily score not found")
	ErrTelemarketingDailyScoreNotFound  = errors.New("telemarketing daily score not found")
	ErrRankingLeaderboardNotFound       = errors.New("ranking leaderboard detail not found")
	ErrRankingLeaderboardInvalidPeriod  = errors.New("ranking leaderboard invalid period")
	ErrRankingLeaderboardInvalidRange   = errors.New("ranking leaderboard invalid range")
	ErrMiHuaTelemarketingConfigRequired = errors.New("mihua telemarketing config required")
	ErrMiHuaTelemarketingRequestFailed  = errors.New("mihua telemarketing request failed")
)

type SyncSalesDailyScoreResult struct {
	ScoreDate   string                  `json:"scoreDate"`
	TotalSales  int                     `json:"totalSales"`
	TotalSaved  int                     `json:"totalSaved"`
	ScoredSales int                     `json:"scoredSales"`
	Items       []model.SalesDailyScore `json:"items"`
}

type SyncTelemarketingDailyScoreResult struct {
	ScoreDate string `json:"scoreDate"`
}

type SalesDailyScoreService interface {
	SyncDailyScores(ctx context.Context, scoreDate string) (SyncSalesDailyScoreResult, error)
	ListDailyRankings(ctx context.Context, scoreDate string, actorUserID int64, actorRole string) (model.SalesDailyScoreRankingListResult, error)
	GetDailyScoreDetail(ctx context.Context, scoreDate string, userID int64, actorUserID int64, actorRole string) (model.SalesDailyScoreDetail, error)
	ListTelemarketingDailyRankings(ctx context.Context, scoreDate string) (model.TelemarketingDailyScoreRankingListResult, error)
	GetTelemarketingDailyScoreDetail(ctx context.Context, scoreDate string, seatWorkNumber string) (model.TelemarketingDailyScoreDetail, error)
	ListRankingLeaderboard(ctx context.Context, period, startDate, endDate string) (model.RankingLeaderboardResult, error)
	GetRankingLeaderboardDetail(ctx context.Context, period, startDate, endDate, identityKey string) (model.RankingLeaderboardDetail, error)
	SyncTelemarketingDailyScores(ctx context.Context) (SyncTelemarketingDailyScoreResult, error)
}

type SalesDailyScoreServiceOption func(*salesDailyScoreService)

type salesDailyScoreService struct {
	repo                      repository.SalesDailyScoreRepository
	systemSettingReader       systemSettingValueReader
	miHuaSeatStatisticsURL    string
	miHuaSeatStatisticsToken  string
	miHuaSeatStatisticsOrigin string
	miHuaHTTPClient           *http.Client
}

type spxxjjMiHuaSeatStatisticsListResponse struct {
	Code int    `json:"code"`
	Info string `json:"info"`
	Data struct {
		List            []json.RawMessage `json:"list"`
		GroupTotalCount int               `json:"groupTotalCount"`
	} `json:"data"`
}

type spxxjjMiHuaSeatStatisticGroup struct {
	Name string `json:"name"`
}

type spxxjjMiHuaSeatStatisticPayload struct {
	ID                int64                           `json:"id"`
	DisplayName       string                          `json:"displayname"`
	WorkNumber        string                          `json:"work_number"`
	Number            string                          `json:"number"`
	IsMobileSeat      string                          `json:"is_mobile_seat"`
	SeatType          int                             `json:"seat_type"`
	Ccgeid            int64                           `json:"ccgeid"`
	SuccessCallCount  int                             `json:"success_call_count"`
	OutTotalSuccess   int                             `json:"out_total_success"`
	OutTotalCallCount int                             `json:"out_total_call_count"`
	CallTotalTimeS    int                             `json:"call_total_time_s"`
	CallValidTimeS    int                             `json:"call_valid_time_s"`
	OutCallTotalTimeS int                             `json:"out_call_total_time_s"`
	OutCallValidTimeS int                             `json:"out_call_valid_time_s"`
	LatestStateTime   int64                           `json:"latest_state_time"`
	LatestStateID     int                             `json:"latest_state_id"`
	Timestamp         int64                           `json:"timestamp"`
	EnterpriseName    string                          `json:"enterprise_name"`
	DepartmentName    string                          `json:"department_name"`
	SeatRealTimeState json.RawMessage                 `json:"seat_real_time_state"`
	Groups            []spxxjjMiHuaSeatStatisticGroup `json:"groups"`
}

type spxxjjMiHuaSeatStatisticRecord struct {
	payload               spxxjjMiHuaSeatStatisticPayload
	rawPayload            string
	seatRealTimeStateJSON string
	groupsJSON            string
	groupName             string
	scoreDate             string
	dataUpdatedAt         time.Time
	statTimestamp         time.Time
}

func WithMiHuaTelemarketingConfig(listURL, token, origin string) SalesDailyScoreServiceOption {
	return func(s *salesDailyScoreService) {
		s.miHuaSeatStatisticsURL = strings.TrimSpace(listURL)
		s.miHuaSeatStatisticsToken = strings.TrimSpace(token)
		s.miHuaSeatStatisticsOrigin = strings.TrimSpace(origin)
	}
}

func WithSalesDailyScoreHTTPClient(client *http.Client) SalesDailyScoreServiceOption {
	return func(s *salesDailyScoreService) {
		if client != nil {
			s.miHuaHTTPClient = client
		}
	}
}

func WithSalesDailyScoreSystemSettingReader(reader systemSettingValueReader) SalesDailyScoreServiceOption {
	return func(s *salesDailyScoreService) {
		s.systemSettingReader = reader
	}
}

func NewSalesDailyScoreService(repo repository.SalesDailyScoreRepository, options ...SalesDailyScoreServiceOption) SalesDailyScoreService {
	service := &salesDailyScoreService{
		repo: repo,
		miHuaHTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	for _, option := range options {
		if option != nil {
			option(service)
		}
	}
	return service
}

func (s *salesDailyScoreService) SyncDailyScores(
	ctx context.Context,
	scoreDate string,
) (SyncSalesDailyScoreResult, error) {
	normalizedDate, startUTC, endUTC, err := normalizeSalesScoreDate(scoreDate)
	if err != nil {
		return SyncSalesDailyScoreResult{}, err
	}

	users, err := s.repo.ListEnabledSalesUsers(ctx)
	if err != nil {
		return SyncSalesDailyScoreResult{}, err
	}

	callMetrics, err := s.repo.ListDailyCallMetrics(ctx, normalizedDate)
	if err != nil {
		return SyncSalesDailyScoreResult{}, err
	}
	callEventsByUser, err := s.repo.ListDailyCallEventsByUser(ctx, startUTC, endUTC)
	if err != nil {
		return SyncSalesDailyScoreResult{}, err
	}

	visitCounts, err := s.repo.CountVisitByUserOnDate(ctx, normalizedDate)
	if err != nil {
		return SyncSalesDailyScoreResult{}, err
	}
	visitEventTimesByUser, err := s.repo.ListVisitEventTimesByUserOnDate(ctx, normalizedDate)
	if err != nil {
		return SyncSalesDailyScoreResult{}, err
	}

	newCustomerCounts, err := s.repo.CountNewCustomersByUserBetween(ctx, startUTC, endUTC)
	if err != nil {
		return SyncSalesDailyScoreResult{}, err
	}
	newCustomerEventTimesByUser, err := s.repo.ListNewCustomerEventTimesByUserBetween(ctx, startUTC, endUTC)
	if err != nil {
		return SyncSalesDailyScoreResult{}, err
	}

	callMetricMap := make(map[int64]model.DailySalesCallMetric, len(callMetrics))
	for _, metric := range callMetrics {
		callMetricMap[metric.UserID] = metric
	}

	inputs := make([]model.SalesDailyScoreUpsertInput, 0, len(users))
	scoredSales := 0
	for _, user := range users {
		callMetric := callMetricMap[user.UserID]
		visitCount := visitCounts[user.UserID]
		newCustomerCount := newCustomerCounts[user.UserID]
		breakdown := scoring.BuildDailySalesScoreBreakdown(
			callMetric.CallNum,
			callMetric.CallDurationSecond,
			visitCount,
			newCustomerCount,
		)
		scoreReachedAt := scoring.CalculateDailySalesScoreReachedAt(
			breakdown,
			callEventsByUser[user.UserID],
			visitEventTimesByUser[user.UserID],
			newCustomerEventTimesByUser[user.UserID],
		)
		if breakdown.TotalScore > 0 {
			scoredSales++
		}

		inputs = append(inputs, model.SalesDailyScoreUpsertInput{
			ScoreDate:           normalizedDate,
			UserID:              user.UserID,
			UserName:            user.UserName,
			RoleName:            user.RoleName,
			CallNum:             callMetric.CallNum,
			CallDurationSecond:  callMetric.CallDurationSecond,
			CallScoreByCount:    breakdown.CallScoreByCount,
			CallScoreByDuration: breakdown.CallScoreByDuration,
			CallScoreType:       breakdown.CallScoreType,
			CallScore:           breakdown.CallScore,
			VisitCount:          visitCount,
			VisitScore:          breakdown.VisitScore,
			NewCustomerCount:    newCustomerCount,
			NewCustomerScore:    breakdown.NewCustomerScore,
			TotalScore:          breakdown.TotalScore,
			ScoreReachedAt:      scoreReachedAt,
		})
	}

	items, err := s.repo.UpsertBatch(ctx, inputs)
	if err != nil {
		return SyncSalesDailyScoreResult{}, err
	}

	return SyncSalesDailyScoreResult{
		ScoreDate:   normalizedDate,
		TotalSales:  len(users),
		TotalSaved:  len(items),
		ScoredSales: scoredSales,
		Items:       items,
	}, nil
}

func (s *salesDailyScoreService) ListDailyRankings(
	ctx context.Context,
	scoreDate string,
	actorUserID int64,
	actorRole string,
) (model.SalesDailyScoreRankingListResult, error) {
	normalizedDate, _, _, err := normalizeSalesScoreDate(scoreDate)
	if err != nil {
		return model.SalesDailyScoreRankingListResult{}, err
	}

	items, err := s.repo.ListByDate(ctx, normalizedDate, actorUserID, actorRole)
	if err != nil {
		return model.SalesDailyScoreRankingListResult{}, err
	}

	rankItems := make([]model.SalesDailyScoreRankingItem, 0, len(items))
	for idx, item := range items {
		rankItems = append(rankItems, model.SalesDailyScoreRankingItem{
			Rank:            idx + 1,
			SalesDailyScore: item,
		})
	}

	return model.SalesDailyScoreRankingListResult{
		ScoreDate: normalizedDate,
		Total:     len(rankItems),
		Items:     rankItems,
	}, nil
}

func (s *salesDailyScoreService) GetDailyScoreDetail(
	ctx context.Context,
	scoreDate string,
	userID int64,
	actorUserID int64,
	actorRole string,
) (model.SalesDailyScoreDetail, error) {
	result, err := s.ListDailyRankings(ctx, scoreDate, actorUserID, actorRole)
	if err != nil {
		return model.SalesDailyScoreDetail{}, err
	}

	for _, item := range result.Items {
		if item.UserID != userID {
			continue
		}
		return model.SalesDailyScoreDetail{
			ScoreDate:  result.ScoreDate,
			Rank:       item.Rank,
			TotalUsers: result.Total,
			HasData:    true,
			Score:      item.SalesDailyScore,
		}, nil
	}

	return model.SalesDailyScoreDetail{}, ErrSalesDailyScoreNotFound
}

func (s *salesDailyScoreService) ListTelemarketingDailyRankings(
	ctx context.Context,
	scoreDate string,
) (model.TelemarketingDailyScoreRankingListResult, error) {
	normalizedScoreDate, _, _, err := normalizeSalesScoreDate(scoreDate)
	if err != nil {
		return model.TelemarketingDailyScoreRankingListResult{}, err
	}

	if isTodayScoreDate(normalizedScoreDate) {
		syncedScoreDate, syncErr := s.spxxjjSyncTelemarketingDailyScores(ctx)
		if syncErr != nil {
			return model.TelemarketingDailyScoreRankingListResult{}, syncErr
		}
		if strings.TrimSpace(syncedScoreDate) != "" {
			normalizedScoreDate = syncedScoreDate
		}
	}

	items, err := s.repo.SpxxjjListTelemarketingDailyScoresByDate(ctx, normalizedScoreDate)
	if err != nil {
		return model.TelemarketingDailyScoreRankingListResult{}, err
	}

	rankItems := make([]model.TelemarketingDailyScoreRankingItem, 0, len(items))
	for idx, item := range items {
		rankItems = append(rankItems, model.TelemarketingDailyScoreRankingItem{
			Rank:                    idx + 1,
			TelemarketingDailyScore: item,
		})
	}

	return model.TelemarketingDailyScoreRankingListResult{
		ScoreDate: normalizedScoreDate,
		Total:     len(rankItems),
		Items:     rankItems,
	}, nil
}

func (s *salesDailyScoreService) SyncTelemarketingDailyScores(
	ctx context.Context,
) (SyncTelemarketingDailyScoreResult, error) {
	scoreDate, err := s.spxxjjSyncTelemarketingDailyScores(ctx)
	if err != nil {
		return SyncTelemarketingDailyScoreResult{}, err
	}
	return SyncTelemarketingDailyScoreResult{ScoreDate: scoreDate}, nil
}

func (s *salesDailyScoreService) ListRankingLeaderboard(
	ctx context.Context,
	period string,
	startDate string,
	endDate string,
) (model.RankingLeaderboardResult, error) {
	normalizedPeriod, normalizedStartDate, normalizedEndDate, shouldSyncDaily, err := resolveRankingLeaderboardRange(
		period,
		startDate,
		endDate,
		time.Now().In(time.Local),
	)
	if err != nil {
		return model.RankingLeaderboardResult{}, err
	}

	if shouldSyncDaily {
		syncedScoreDate, syncErr := s.spxxjjSyncTelemarketingDailyScores(ctx)
		if syncErr != nil {
			return model.RankingLeaderboardResult{}, syncErr
		}
		if strings.TrimSpace(syncedScoreDate) != "" {
			normalizedStartDate = syncedScoreDate
			normalizedEndDate = syncedScoreDate
		}
	}

	items, err := s.repo.ListRankingLeaderboard(ctx, normalizedStartDate, normalizedEndDate)
	if err != nil {
		return model.RankingLeaderboardResult{}, err
	}

	rankItems := make([]model.RankingLeaderboardItem, 0, len(items))
	for idx, item := range items {
		item.Rank = idx + 1
		rankItems = append(rankItems, item)
	}

	return model.RankingLeaderboardResult{
		Period:    normalizedPeriod,
		StartDate: normalizedStartDate,
		EndDate:   normalizedEndDate,
		Total:     len(rankItems),
		Items:     rankItems,
	}, nil
}

func (s *salesDailyScoreService) GetRankingLeaderboardDetail(
	ctx context.Context,
	period string,
	startDate string,
	endDate string,
	identityKey string,
) (model.RankingLeaderboardDetail, error) {
	result, err := s.ListRankingLeaderboard(ctx, period, startDate, endDate)
	if err != nil {
		return model.RankingLeaderboardDetail{}, err
	}

	targetIdentityKey := strings.TrimSpace(identityKey)
	for _, item := range result.Items {
		if strings.TrimSpace(item.IdentityKey) != targetIdentityKey {
			continue
		}
		return model.RankingLeaderboardDetail{
			Period:     result.Period,
			StartDate:  result.StartDate,
			EndDate:    result.EndDate,
			Rank:       item.Rank,
			TotalUsers: result.Total,
			HasData:    true,
			Score:      item,
		}, nil
	}

	return model.RankingLeaderboardDetail{}, ErrRankingLeaderboardNotFound
}

func (s *salesDailyScoreService) GetTelemarketingDailyScoreDetail(
	ctx context.Context,
	scoreDate string,
	seatWorkNumber string,
) (model.TelemarketingDailyScoreDetail, error) {
	result, err := s.ListTelemarketingDailyRankings(ctx, scoreDate)
	if err != nil {
		return model.TelemarketingDailyScoreDetail{}, err
	}

	targetSeatWorkNumber := strings.TrimSpace(seatWorkNumber)
	for _, item := range result.Items {
		if strings.TrimSpace(item.SeatWorkNumber) != targetSeatWorkNumber {
			continue
		}
		return model.TelemarketingDailyScoreDetail{
			ScoreDate:  result.ScoreDate,
			Rank:       item.Rank,
			TotalUsers: result.Total,
			HasData:    true,
			Score:      item.TelemarketingDailyScore,
		}, nil
	}

	return model.TelemarketingDailyScoreDetail{}, ErrTelemarketingDailyScoreNotFound
}

func (s *salesDailyScoreService) spxxjjSyncTelemarketingDailyScores(ctx context.Context) (string, error) {
	records, scoreDate, err := s.spxxjjFetchMiHuaSeatStatistics(ctx)
	if err != nil {
		return "", err
	}

	normalizedScoreDate, startUTC, endUTC, err := normalizeSalesScoreDate(scoreDate)
	if err != nil {
		return "", err
	}

	workNumbers := make([]string, 0, len(records))
	for _, record := range records {
		workNumbers = append(workNumbers, strings.TrimSpace(record.payload.WorkNumber))
	}

	telemarketingUsers, err := s.repo.SpxxjjListEnabledTelemarketingUsersByWorkNumbers(ctx, workNumbers)
	if err != nil {
		return "", err
	}

	rawInputs := make([]model.SpxxjjMiHuaSeatStatisticUpsertInput, 0, len(records))
	telemarketingUserIDs := make([]int64, 0, len(telemarketingUsers))
	inviterCandidatesBySeat := make(map[string][]string, len(records))
	inviterQueries := make([]string, 0, len(records)*4)
	for _, record := range records {
		workNumber := strings.TrimSpace(record.payload.WorkNumber)
		if workNumber == "" {
			continue
		}

		var matchedUserID *int64
		matchedUserName := ""
		roleName := "电销"
		candidates := uniqueNonEmptyStrings(
			workNumber,
			record.payload.DisplayName,
		)
		if localUser, exists := telemarketingUsers[workNumber]; exists {
			id := localUser.UserID
			matchedUserID = &id
			matchedUserName = firstNonEmptyTelemarketing(localUser.UserName, localUser.Nickname, localUser.Username)
			roleName = firstNonEmptyTelemarketing(localUser.RoleName, "电销")
			telemarketingUserIDs = append(telemarketingUserIDs, localUser.UserID)
			candidates = uniqueNonEmptyStrings(append(
				append([]string{}, candidates...),
				matchedUserName,
				localUser.Username,
				localUser.Nickname,
				localUser.WorkNumber,
			)...)
		}
		inviterCandidatesBySeat[workNumber] = candidates
		inviterQueries = append(inviterQueries, candidates...)

		rawInputs = append(rawInputs, model.SpxxjjMiHuaSeatStatisticUpsertInput{
			ScoreDate:              normalizedScoreDate,
			SeatID:                 record.payload.ID,
			SeatName:               strings.TrimSpace(record.payload.DisplayName),
			WorkNumber:             workNumber,
			ServiceNumber:          strings.TrimSpace(record.payload.Number),
			IsMobileSeat:           strings.TrimSpace(record.payload.IsMobileSeat),
			SeatType:               record.payload.SeatType,
			Ccgeid:                 record.payload.Ccgeid,
			SuccessCallCount:       record.payload.SuccessCallCount,
			OutTotalSuccess:        record.payload.OutTotalSuccess,
			OutTotalCallCount:      record.payload.OutTotalCallCount,
			CallTotalTimeSecond:    record.payload.CallTotalTimeS,
			CallValidTimeSecond:    record.payload.CallValidTimeS,
			OutCallTotalTimeSecond: record.payload.OutCallTotalTimeS,
			OutCallValidTimeSecond: record.payload.OutCallValidTimeS,
			LatestStateTime:        zeroTimeAsNil(record.dataUpdatedAt),
			LatestStateID:          record.payload.LatestStateID,
			StatTimestamp:          zeroTimeAsNil(record.statTimestamp),
			EnterpriseName:         strings.TrimSpace(record.payload.EnterpriseName),
			DepartmentName:         strings.TrimSpace(record.payload.DepartmentName),
			GroupName:              record.groupName,
			SeatRealTimeStateJSON:  record.seatRealTimeStateJSON,
			GroupsJSON:             record.groupsJSON,
			RawPayload:             record.rawPayload,
			MatchedUserID:          matchedUserID,
			MatchedUserName:        matchedUserName,
			RoleName:               roleName,
		})
	}

	if err := s.repo.SpxxjjUpsertMiHuaSeatStatistics(ctx, rawInputs); err != nil {
		return "", err
	}

	newCustomerCounts := make(map[int64]int)
	if len(telemarketingUserIDs) > 0 {
		newCustomerCounts, err = s.repo.CountNewCustomersByTelemarketingUserBetween(ctx, telemarketingUserIDs, startUTC, endUTC)
		if err != nil {
			return "", err
		}
	}

	invitationCountsByInviter := make(map[string]int)
	if len(inviterQueries) > 0 {
		invitationCountsByInviter, err = s.repo.CountCustomerVisitsByInvitersOnDate(ctx, normalizedScoreDate, inviterQueries)
		if err != nil {
			return "", err
		}
	}

	scoreInputs := make([]model.SpxxjjTelemarketingDailyScoreUpsertInput, 0, len(records))
	for _, record := range records {
		workNumber := strings.TrimSpace(record.payload.WorkNumber)
		if workNumber == "" {
			continue
		}

		var matchedUserID *int64
		matchedUserName := ""
		roleName := "电销"
		newCustomerCount := 0
		if localUser, exists := telemarketingUsers[workNumber]; exists {
			id := localUser.UserID
			matchedUserID = &id
			matchedUserName = firstNonEmptyTelemarketing(localUser.UserName, localUser.Nickname, localUser.Username)
			roleName = firstNonEmptyTelemarketing(localUser.RoleName, "电销")
			newCustomerCount = newCustomerCounts[localUser.UserID]
		}

		invitationCount := 0
		for _, inviter := range inviterCandidatesBySeat[workNumber] {
			invitationCount += invitationCountsByInviter[strings.TrimSpace(inviter)]
		}

		callNum := max(record.payload.OutTotalCallCount, 0)
		answeredCallCount := max(record.payload.OutTotalSuccess, 0)
		callDurationSecond := max(record.payload.OutCallTotalTimeS, 0)
		if callDurationSecond == 0 {
			callDurationSecond = max(record.payload.CallTotalTimeS, 0)
		}
		missedCallCount := max(callNum-answeredCallCount, 0)
		answerRate := 0.0
		if callNum > 0 {
			answerRate = float64(answeredCallCount) * 100 / float64(callNum)
		}

		breakdown := scoring.BuildDailyTelemarketingScoreBreakdown(
			answeredCallCount,
			callDurationSecond,
			invitationCount,
			newCustomerCount,
		)

		scoreInputs = append(scoreInputs, model.SpxxjjTelemarketingDailyScoreUpsertInput{
			ScoreDate:           normalizedScoreDate,
			SeatWorkNumber:      workNumber,
			SeatName:            strings.TrimSpace(record.payload.DisplayName),
			MatchedUserID:       matchedUserID,
			MatchedUserName:     matchedUserName,
			ServiceNumber:       strings.TrimSpace(record.payload.Number),
			GroupName:           record.groupName,
			RoleName:            roleName,
			CallNum:             callNum,
			AnsweredCallCount:   answeredCallCount,
			MissedCallCount:     missedCallCount,
			AnswerRate:          answerRate,
			CallDurationSecond:  callDurationSecond,
			NewCustomerCount:    newCustomerCount,
			InvitationCount:     invitationCount,
			CallScoreByCount:    breakdown.CallScoreByCount,
			CallScoreByDuration: breakdown.CallScoreByDuration,
			CallScoreType:       breakdown.CallScoreType,
			CallScore:           breakdown.CallScore,
			InvitationScore:     breakdown.InvitationScore,
			NewCustomerScore:    breakdown.NewCustomerScore,
			TotalScore:          breakdown.TotalScore,
			DataUpdatedAt:       zeroTimeAsNil(record.dataUpdatedAt),
		})
	}

	if _, err := s.repo.SpxxjjUpsertTelemarketingDailyScores(ctx, scoreInputs); err != nil {
		return "", err
	}
	return normalizedScoreDate, nil
}

func (s *salesDailyScoreService) spxxjjFetchMiHuaSeatStatistics(
	ctx context.Context,
) ([]spxxjjMiHuaSeatStatisticRecord, string, error) {
	token, err := s.resolveMiHuaSeatStatisticsToken()
	if err != nil {
		return nil, "", err
	}
	if strings.TrimSpace(s.miHuaSeatStatisticsURL) == "" ||
		strings.TrimSpace(s.miHuaSeatStatisticsOrigin) == "" {
		return nil, "", ErrMiHuaTelemarketingConfigRequired
	}

	baseURL, err := url.Parse(s.miHuaSeatStatisticsURL)
	if err != nil {
		return nil, "", fmt.Errorf("%w: invalid url: %v", ErrMiHuaTelemarketingRequestFailed, err)
	}
	originURL, err := url.Parse(s.miHuaSeatStatisticsOrigin)
	if err != nil || strings.TrimSpace(originURL.Scheme) == "" || strings.TrimSpace(originURL.Host) == "" {
		return nil, "", fmt.Errorf("%w: invalid source origin", ErrMiHuaTelemarketingRequestFailed)
	}
	origin := originURL.Scheme + "://" + originURL.Host
	referer := origin + "/"

	allRecords := make([]spxxjjMiHuaSeatStatisticRecord, 0, defaultMiHuaPageSize)
	scoreDate := ""
	total := 0
	for page := 1; ; page++ {
		pageURL := *baseURL
		query := pageURL.Query()
		query.Set("page", strconv.Itoa(page))
		query.Set("per_page", strconv.Itoa(defaultMiHuaPageSize))
		pageURL.RawQuery = query.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL.String(), nil)
		if err != nil {
			return nil, "", fmt.Errorf("%w: %v", ErrMiHuaTelemarketingRequestFailed, err)
		}
		req.Header.Set("token", token)
		req.Header.Set("accept", "application/json, text/plain, */*")
		req.Header.Set("origin", origin)
		req.Header.Set("referer", referer)
		req.Header.Set("source", "client.web")
		req.Header.Set("nonce", generateMiHuaRequestNonce())
		req.Header.Set("user-agent", "Mozilla/5.0")

		resp, err := s.miHuaHTTPClient.Do(req)
		if err != nil {
			return nil, "", fmt.Errorf("%w: %v", ErrMiHuaTelemarketingRequestFailed, err)
		}

		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, "", fmt.Errorf("%w: %v", ErrMiHuaTelemarketingRequestFailed, readErr)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, "", fmt.Errorf("%w: status=%d body=%s", ErrMiHuaTelemarketingRequestFailed, resp.StatusCode, strings.TrimSpace(string(respBody)))
		}

		var pageResp spxxjjMiHuaSeatStatisticsListResponse
		if err := json.Unmarshal(respBody, &pageResp); err != nil {
			return nil, "", fmt.Errorf("%w: %v", ErrMiHuaTelemarketingRequestFailed, err)
		}
		if pageResp.Code != http.StatusOK {
			message := strings.TrimSpace(pageResp.Info)
			if message == "" {
				message = "unknown upstream error"
			}
			return nil, "", fmt.Errorf("%w: %s", ErrMiHuaTelemarketingRequestFailed, message)
		}
		if page == 1 {
			total = max(pageResp.Data.GroupTotalCount, 0)
		}
		if len(pageResp.Data.List) == 0 {
			break
		}

		for _, rawItem := range pageResp.Data.List {
			var payload spxxjjMiHuaSeatStatisticPayload
			if err := json.Unmarshal(rawItem, &payload); err != nil {
				return nil, "", fmt.Errorf("%w: parse seat statistics item failed: %v", ErrMiHuaTelemarketingRequestFailed, err)
			}
			workNumber := strings.TrimSpace(payload.WorkNumber)
			if workNumber == "" {
				continue
			}

			groupsJSON := marshalMiHuaRawJSON(payload.Groups, "[]")
			seatRealTimeStateJSON := rawJSONText(payload.SeatRealTimeState, "{}")
			statTimestamp := parseMiHuaUnixSecond(payload.Timestamp)
			dataUpdatedAt := parseMiHuaUnixSecond(payload.LatestStateTime)
			if scoreDate == "" {
				scoreDate = resolveSpxxjjScoreDate(statTimestamp, dataUpdatedAt)
			}
			allRecords = append(allRecords, spxxjjMiHuaSeatStatisticRecord{
				payload:               payload,
				rawPayload:            rawJSONText(rawItem, "{}"),
				seatRealTimeStateJSON: seatRealTimeStateJSON,
				groupsJSON:            groupsJSON,
				groupName:             joinSpxxjjGroupNames(payload.Groups),
				scoreDate:             scoreDate,
				dataUpdatedAt:         dataUpdatedAt,
				statTimestamp:         statTimestamp,
			})
		}

		if total > 0 && len(allRecords) >= total {
			break
		}
		if len(pageResp.Data.List) < defaultMiHuaPageSize {
			break
		}
	}

	if scoreDate == "" {
		scoreDate = time.Now().In(time.Local).Format("2006-01-02")
	}
	return allRecords, scoreDate, nil
}

func (s *salesDailyScoreService) resolveMiHuaSeatStatisticsToken() (string, error) {
	if token := getTrimmedSystemSettingValue(s.systemSettingReader, model.SystemSettingKeyMiHuaCallRecordToken); token != "" {
		return token, nil
	}
	if token := strings.TrimSpace(s.miHuaSeatStatisticsToken); token != "" {
		return token, nil
	}
	return "", ErrMiHuaTelemarketingConfigRequired
}

func generateMiHuaRequestNonce() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return hex.EncodeToString(bytes)
}

func zeroTimeAsNil(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	cloned := value.UTC()
	return &cloned
}

func parseMiHuaUnixSecond(value int64) time.Time {
	if value <= 0 {
		return time.Time{}
	}
	return time.Unix(value, 0).UTC()
}

func resolveSpxxjjScoreDate(statTimestamp, updatedAt time.Time) string {
	switch {
	case !statTimestamp.IsZero():
		return statTimestamp.In(time.Local).Format("2006-01-02")
	case !updatedAt.IsZero():
		return updatedAt.In(time.Local).Format("2006-01-02")
	default:
		return time.Now().In(time.Local).Format("2006-01-02")
	}
}

func resolveRankingLeaderboardRange(
	period string,
	startDate string,
	endDate string,
	now time.Time,
) (string, string, string, bool, error) {
	normalizedPeriod := strings.ToLower(strings.TrimSpace(period))
	if normalizedPeriod == "" {
		normalizedPeriod = "day"
	}

	now = now.In(time.Local)
	defaultStartDate := ""
	defaultEndDate := now.Format("2006-01-02")

	switch normalizedPeriod {
	case "day":
		defaultStartDate = defaultEndDate
	case "week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		defaultStartDate = now.AddDate(0, 0, -(weekday - 1)).Format("2006-01-02")
	case "month":
		defaultStartDate = now.Format("2006-01") + "-01"
	case "all":
		defaultStartDate = "1970-01-01"
	default:
		return "", "", "", false, fmt.Errorf("%w: %s", ErrRankingLeaderboardInvalidPeriod, normalizedPeriod)
	}

	start := strings.TrimSpace(startDate)
	end := strings.TrimSpace(endDate)
	if start == "" && end == "" {
		start = defaultStartDate
		end = defaultEndDate
	} else {
		if start == "" {
			start = end
		}
		if end == "" {
			end = start
		}
	}

	normalizedStartDate, _, _, err := normalizeSalesScoreDate(start)
	if err != nil {
		return "", "", "", false, fmt.Errorf("%w: %v", ErrRankingLeaderboardInvalidRange, err)
	}
	normalizedEndDate, _, _, err := normalizeSalesScoreDate(end)
	if err != nil {
		return "", "", "", false, fmt.Errorf("%w: %v", ErrRankingLeaderboardInvalidRange, err)
	}
	if normalizedStartDate > normalizedEndDate {
		return "", "", "", false, fmt.Errorf("%w: startDate must be earlier than or equal to endDate", ErrRankingLeaderboardInvalidRange)
	}

	shouldSyncDaily := normalizedPeriod == "day" &&
		normalizedStartDate == normalizedEndDate &&
		isTodayScoreDate(normalizedStartDate)

	return normalizedPeriod, normalizedStartDate, normalizedEndDate, shouldSyncDaily, nil
}

func rawJSONText(raw json.RawMessage, fallback string) string {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return fallback
	}
	return trimmed
}

func marshalMiHuaRawJSON(value interface{}, fallback string) string {
	data, err := json.Marshal(value)
	if err != nil {
		return fallback
	}
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" || trimmed == "null" {
		return fallback
	}
	return trimmed
}

func joinSpxxjjGroupNames(groups []spxxjjMiHuaSeatStatisticGroup) string {
	names := make([]string, 0, len(groups))
	for _, group := range groups {
		if name := strings.TrimSpace(group.Name); name != "" {
			names = append(names, name)
		}
	}
	return strings.Join(uniqueNonEmptyStrings(names...), "、")
}

func normalizeSalesScoreDate(scoreDate string) (string, time.Time, time.Time, error) {
	trimmed := strings.TrimSpace(scoreDate)
	if trimmed == "" {
		trimmed = time.Now().In(time.Local).Format("2006-01-02")
	}

	localDay, err := time.ParseInLocation("2006-01-02", trimmed, time.Local)
	if err != nil {
		return "", time.Time{}, time.Time{}, err
	}
	startUTC := localDay.UTC()
	endUTC := localDay.AddDate(0, 0, 1).UTC()
	return localDay.Format("2006-01-02"), startUTC, endUTC, nil
}

func isTodayScoreDate(scoreDate string) bool {
	return strings.TrimSpace(scoreDate) == time.Now().In(time.Local).Format("2006-01-02")
}

func uniqueNonEmptyStrings(values ...string) []string {
	if len(values) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func firstNonEmptyTelemarketing(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
