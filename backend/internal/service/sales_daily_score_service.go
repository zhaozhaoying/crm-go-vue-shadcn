package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/scoring"
	"context"
	"errors"
	"strings"
	"time"
)

var ErrSalesDailyScoreNotFound = errors.New("sales daily score not found")

type SyncSalesDailyScoreResult struct {
	ScoreDate   string                  `json:"scoreDate"`
	TotalSales  int                     `json:"totalSales"`
	TotalSaved  int                     `json:"totalSaved"`
	ScoredSales int                     `json:"scoredSales"`
	Items       []model.SalesDailyScore `json:"items"`
}

type SalesDailyScoreService interface {
	SyncDailyScores(ctx context.Context, scoreDate string) (SyncSalesDailyScoreResult, error)
	ListDailyRankings(ctx context.Context, scoreDate string, actorUserID int64, actorRole string) (model.SalesDailyScoreRankingListResult, error)
	GetDailyScoreDetail(ctx context.Context, scoreDate string, userID int64, actorUserID int64, actorRole string) (model.SalesDailyScoreDetail, error)
}

type salesDailyScoreService struct {
	repo repository.SalesDailyScoreRepository
}

func NewSalesDailyScoreService(repo repository.SalesDailyScoreRepository) SalesDailyScoreService {
	return &salesDailyScoreService{repo: repo}
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
