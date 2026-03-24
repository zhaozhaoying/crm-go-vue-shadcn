package service

import (
	"backend/internal/model"
	"context"
	"errors"
	"testing"
	"time"
)

type salesDailyScoreRepoStub struct {
	users             []model.SalesDailyScoreUser
	callMetrics       []model.DailySalesCallMetric
	visitCounts       map[int64]int
	newCustomerCounts map[int64]int
	upserts           []model.SalesDailyScoreUpsertInput
	listByDateItems   []model.SalesDailyScore
}

func (s *salesDailyScoreRepoStub) ListEnabledSalesUsers(context.Context) ([]model.SalesDailyScoreUser, error) {
	return s.users, nil
}

func (s *salesDailyScoreRepoStub) ListDailyCallMetrics(context.Context, string) ([]model.DailySalesCallMetric, error) {
	return s.callMetrics, nil
}

func (s *salesDailyScoreRepoStub) CountVisitByUserOnDate(context.Context, string) (map[int64]int, error) {
	return s.visitCounts, nil
}

func (s *salesDailyScoreRepoStub) CountNewCustomersByUserBetween(context.Context, time.Time, time.Time) (map[int64]int, error) {
	return s.newCustomerCounts, nil
}

func (s *salesDailyScoreRepoStub) UpsertBatch(
	_ context.Context,
	items []model.SalesDailyScoreUpsertInput,
) ([]model.SalesDailyScore, error) {
	s.upserts = append([]model.SalesDailyScoreUpsertInput(nil), items...)

	result := make([]model.SalesDailyScore, 0, len(items))
	for _, item := range items {
		result = append(result, model.SalesDailyScore{
			ScoreDate:           item.ScoreDate,
			UserID:              item.UserID,
			UserName:            item.UserName,
			RoleName:            item.RoleName,
			CallNum:             item.CallNum,
			CallDurationSecond:  item.CallDurationSecond,
			CallScoreByCount:    item.CallScoreByCount,
			CallScoreByDuration: item.CallScoreByDuration,
			CallScoreType:       item.CallScoreType,
			CallScore:           item.CallScore,
			VisitCount:          item.VisitCount,
			VisitScore:          item.VisitScore,
			NewCustomerCount:    item.NewCustomerCount,
			NewCustomerScore:    item.NewCustomerScore,
			TotalScore:          item.TotalScore,
		})
	}
	return result, nil
}

func (s *salesDailyScoreRepoStub) ListByDate(context.Context, string, int64, string) ([]model.SalesDailyScore, error) {
	return append([]model.SalesDailyScore(nil), s.listByDateItems...), nil
}

func TestSyncDailyScoresChoosesHigherCallScoreAndAccumulates(t *testing.T) {
	t.Parallel()

	repo := &salesDailyScoreRepoStub{
		users: []model.SalesDailyScoreUser{
			{UserID: 1, UserName: "张三", RoleName: "销售员工"},
			{UserID: 2, UserName: "李四", RoleName: "销售员工"},
		},
		callMetrics: []model.DailySalesCallMetric{
			{UserID: 1, CallNum: 160, CallDurationSecond: 3200},
			{UserID: 2, CallNum: 190, CallDurationSecond: 1200},
		},
		visitCounts: map[int64]int{
			1: 5,
			2: 3,
		},
		newCustomerCounts: map[int64]int{
			1: 2,
			2: 3,
		},
	}

	svc := NewSalesDailyScoreService(repo)
	result, err := svc.SyncDailyScores(context.Background(), "2026-03-20")
	if err != nil {
		t.Fatalf("SyncDailyScores returned error: %v", err)
	}

	if result.ScoreDate != "2026-03-20" {
		t.Fatalf("unexpected score date: %s", result.ScoreDate)
	}
	if result.TotalSales != 2 || result.TotalSaved != 2 || result.ScoredSales != 2 {
		t.Fatalf("unexpected summary: %+v", result)
	}
	if len(repo.upserts) != 2 {
		t.Fatalf("expected 2 upserts, got %d", len(repo.upserts))
	}

	first := repo.upserts[0]
	if first.CallScoreByCount != 50 || first.CallScoreByDuration != 70 {
		t.Fatalf("unexpected first call score candidates: %+v", first)
	}
	if first.CallScoreType != model.SalesDailyScoreCallScoreTypeDuration || first.CallScore != 70 {
		t.Fatalf("expected duration call score for first user, got %+v", first)
	}
	if first.VisitScore != 60 || first.NewCustomerScore != 0 || first.TotalScore != 130 {
		t.Fatalf("unexpected first total score: %+v", first)
	}

	second := repo.upserts[1]
	if second.CallScoreByCount != 70 || second.CallScoreByDuration != 30 {
		t.Fatalf("unexpected second call score candidates: %+v", second)
	}
	if second.CallScoreType != model.SalesDailyScoreCallScoreTypeCallNum || second.CallScore != 70 {
		t.Fatalf("expected call-num score for second user, got %+v", second)
	}
	if second.VisitScore != 40 || second.NewCustomerScore != 10 || second.TotalScore != 120 {
		t.Fatalf("unexpected second total score: %+v", second)
	}
}

func TestSyncDailyScoresUsesProgressiveScoring(t *testing.T) {
	t.Parallel()

	repo := &salesDailyScoreRepoStub{
		users: []model.SalesDailyScoreUser{
			{UserID: 3, UserName: "王五", RoleName: "销售员工"},
		},
		callMetrics: []model.DailySalesCallMetric{
			{UserID: 3, CallNum: 90, CallDurationSecond: 24 * 60},
		},
		visitCounts: map[int64]int{
			3: 2,
		},
		newCustomerCounts: map[int64]int{
			3: 6,
		},
	}

	svc := NewSalesDailyScoreService(repo)
	_, err := svc.SyncDailyScores(context.Background(), "2026-03-20")
	if err != nil {
		t.Fatalf("SyncDailyScores returned error: %v", err)
	}
	if len(repo.upserts) != 1 {
		t.Fatalf("expected 1 upsert, got %d", len(repo.upserts))
	}

	item := repo.upserts[0]
	if item.CallScoreByCount != 30 || item.CallScoreByDuration != 40 {
		t.Fatalf("unexpected progressive call scores: %+v", item)
	}
	if item.CallScoreType != model.SalesDailyScoreCallScoreTypeDuration || item.CallScore != 40 {
		t.Fatalf("unexpected chosen call score: %+v", item)
	}
	if item.VisitScore != 20 || item.NewCustomerScore != 20 || item.TotalScore != 80 {
		t.Fatalf("unexpected progressive totals: %+v", item)
	}
}

func TestListDailyRankingsBuildsRanks(t *testing.T) {
	t.Parallel()

	repo := &salesDailyScoreRepoStub{
		listByDateItems: []model.SalesDailyScore{
			{UserID: 2, UserName: "李四", TotalScore: 120, CallScore: 70, VisitScore: 40, NewCustomerScore: 10},
			{UserID: 1, UserName: "张三", TotalScore: 100, CallScore: 50, VisitScore: 40, NewCustomerScore: 10},
		},
	}

	svc := NewSalesDailyScoreService(repo)
	result, err := svc.ListDailyRankings(context.Background(), "2026-03-20", 1, "admin")
	if err != nil {
		t.Fatalf("ListDailyRankings returned error: %v", err)
	}

	if result.Total != 2 || len(result.Items) != 2 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.Items[0].Rank != 1 || result.Items[0].UserID != 2 {
		t.Fatalf("unexpected first rank item: %+v", result.Items[0])
	}
	if result.Items[1].Rank != 2 || result.Items[1].UserID != 1 {
		t.Fatalf("unexpected second rank item: %+v", result.Items[1])
	}
}

func TestGetDailyScoreDetailReturnsNotFound(t *testing.T) {
	t.Parallel()

	repo := &salesDailyScoreRepoStub{}
	svc := NewSalesDailyScoreService(repo)

	_, err := svc.GetDailyScoreDetail(context.Background(), "2026-03-20", 99, 1, "admin")
	if !errors.Is(err, ErrSalesDailyScoreNotFound) {
		t.Fatalf("expected ErrSalesDailyScoreNotFound, got %v", err)
	}
}
