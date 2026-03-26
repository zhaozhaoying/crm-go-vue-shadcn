package service

import (
	"backend/internal/model"
	"context"
	"errors"
	"testing"
	"time"
)

type runtimeCallStatServiceStub struct {
	calls     int
	lastInput SyncHanghangCRMDailyUserCallStatInput
	result    SyncHanghangCRMDailyUserCallStatResult
	err       error
}

func (s *runtimeCallStatServiceStub) SyncDailyUserCallStats(
	_ context.Context,
	input SyncHanghangCRMDailyUserCallStatInput,
) (SyncHanghangCRMDailyUserCallStatResult, error) {
	s.calls++
	s.lastInput = input
	if s.err != nil {
		return SyncHanghangCRMDailyUserCallStatResult{}, s.err
	}
	if s.result.StatDate == "" {
		s.result.StatDate = input.StartTime
	}
	return s.result, nil
}

type runtimeSalesDailyScoreServiceStub struct {
	calls         int
	lastScoreDate string
	result        SyncSalesDailyScoreResult
	err           error
}

func (s *runtimeSalesDailyScoreServiceStub) SyncDailyScores(
	_ context.Context,
	scoreDate string,
) (SyncSalesDailyScoreResult, error) {
	s.calls++
	s.lastScoreDate = scoreDate
	if s.err != nil {
		return SyncSalesDailyScoreResult{}, s.err
	}
	if s.result.ScoreDate == "" {
		s.result.ScoreDate = scoreDate
	}
	return s.result, nil
}

func (s *runtimeSalesDailyScoreServiceStub) ListDailyRankings(
	context.Context,
	string,
	int64,
	string,
) (result model.SalesDailyScoreRankingListResult, err error) {
	return result, nil
}

func (s *runtimeSalesDailyScoreServiceStub) GetDailyScoreDetail(
	context.Context,
	string,
	int64,
	int64,
	string,
) (detail model.SalesDailyScoreDetail, err error) {
	return detail, nil
}

func TestHanghangCRMDailyScoreRuntimeSkipsBeforeScheduledHour(t *testing.T) {
	t.Parallel()

	callStatStub := &runtimeCallStatServiceStub{}
	scoreStub := &runtimeSalesDailyScoreServiceStub{}
	runtime := NewHanghangCRMDailyScoreRuntime(callStatStub, scoreStub, time.Minute, time.Local)
	runtime.nowFunc = func() time.Time {
		return time.Date(2026, 3, 23, 20, 59, 0, 0, time.Local)
	}

	runtime.runOnce(context.Background())

	if callStatStub.calls != 0 {
		t.Fatalf("expected call-stat sync not to run before schedule hour, got %d calls", callStatStub.calls)
	}
	if scoreStub.calls != 0 {
		t.Fatalf("expected score sync not to run before schedule hour, got %d calls", scoreStub.calls)
	}
}

func TestHanghangCRMDailyScoreRuntimeRunsOncePerDayAfterScheduledHour(t *testing.T) {
	t.Parallel()

	callStatStub := &runtimeCallStatServiceStub{}
	scoreStub := &runtimeSalesDailyScoreServiceStub{}
	runtime := NewHanghangCRMDailyScoreRuntime(callStatStub, scoreStub, time.Minute, time.Local)
	runtime.nowFunc = func() time.Time {
		return time.Date(2026, 3, 23, 21, 0, 0, 0, time.Local)
	}

	runtime.runOnce(context.Background())
	runtime.runOnce(context.Background())

	if callStatStub.calls != 1 {
		t.Fatalf("expected call-stat sync to run once, got %d calls", callStatStub.calls)
	}
	if scoreStub.calls != 1 {
		t.Fatalf("expected score sync to run once, got %d calls", scoreStub.calls)
	}
	if callStatStub.lastInput.StartTime != "2026-03-23" || callStatStub.lastInput.EndTime != "2026-03-23" {
		t.Fatalf("expected runtime to sync current date, got %+v", callStatStub.lastInput)
	}
	if scoreStub.lastScoreDate != "2026-03-23" {
		t.Fatalf("expected score sync date 2026-03-23, got %s", scoreStub.lastScoreDate)
	}
}

func TestHanghangCRMDailyScoreRuntimeSkipsOnWeekends(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		now  time.Time
	}{
		{"Saturday", time.Date(2026, 3, 28, 21, 5, 0, 0, time.Local)},
		{"Sunday", time.Date(2026, 3, 29, 21, 5, 0, 0, time.Local)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callStatStub := &runtimeCallStatServiceStub{}
			scoreStub := &runtimeSalesDailyScoreServiceStub{}
			runtime := NewHanghangCRMDailyScoreRuntime(callStatStub, scoreStub, time.Minute, time.Local)
			runtime.nowFunc = func() time.Time { return tt.now }

			runtime.runOnce(context.Background())

			if callStatStub.calls != 0 {
				t.Fatalf("expected no call-stat sync on %s, got %d calls", tt.name, callStatStub.calls)
			}
			if scoreStub.calls != 0 {
				t.Fatalf("expected no score sync on %s, got %d calls", tt.name, scoreStub.calls)
			}
		})
	}
}

func TestHanghangCRMDailyScoreRuntimeSkipsWhenHolidayModeEnabled(t *testing.T) {
	t.Parallel()

	callStatStub := &runtimeCallStatServiceStub{}
	scoreStub := &runtimeSalesDailyScoreServiceStub{}
	runtime := NewHanghangCRMDailyScoreRuntime(
		callStatStub, scoreStub, time.Minute, time.Local,
		func() bool { return true },
	)
	runtime.nowFunc = func() time.Time {
		return time.Date(2026, 3, 25, 21, 5, 0, 0, time.Local) // Wednesday
	}

	runtime.runOnce(context.Background())

	if callStatStub.calls != 0 {
		t.Fatalf("expected no call-stat sync when holiday mode enabled, got %d calls", callStatStub.calls)
	}
	if scoreStub.calls != 0 {
		t.Fatalf("expected no score sync when holiday mode enabled, got %d calls", scoreStub.calls)
	}
}

func TestHanghangCRMDailyScoreRuntimeRetriesAfterFailure(t *testing.T) {
	t.Parallel()

	callStatStub := &runtimeCallStatServiceStub{
		err: errors.New("upstream failed"),
	}
	scoreStub := &runtimeSalesDailyScoreServiceStub{}
	runtime := NewHanghangCRMDailyScoreRuntime(callStatStub, scoreStub, time.Minute, time.Local)
	runtime.nowFunc = func() time.Time {
		return time.Date(2026, 3, 23, 21, 5, 0, 0, time.Local)
	}

	runtime.runOnce(context.Background())
	if runtime.lastSuccessfulDate != "" {
		t.Fatalf("expected failed run not to record success date, got %q", runtime.lastSuccessfulDate)
	}

	callStatStub.err = nil
	runtime.runOnce(context.Background())

	if callStatStub.calls != 2 {
		t.Fatalf("expected runtime to retry after failure, got %d calls", callStatStub.calls)
	}
	if scoreStub.calls != 1 {
		t.Fatalf("expected score sync to run after retry success, got %d calls", scoreStub.calls)
	}
	if runtime.lastSuccessfulDate != "2026-03-23" {
		t.Fatalf("expected success date to be recorded after retry, got %q", runtime.lastSuccessfulDate)
	}
}
