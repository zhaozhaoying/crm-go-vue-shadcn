package service

import (
	"backend/internal/model"
	"context"
	"testing"
	"time"
)

type callRecordingRepoStub struct {
	latestStartTime int64
	latestErr       error
}

func (s *callRecordingRepoStub) List(
	_ context.Context,
	_ model.CallRecordingListFilter,
) (model.CallRecordingListResult, error) {
	return model.CallRecordingListResult{}, nil
}

func (s *callRecordingRepoStub) FindByID(
	_ context.Context,
	_ string,
	_ bool,
	_ string,
) (*model.CallRecording, error) {
	return nil, ErrCallRecordingNotFound
}

func (s *callRecordingRepoStub) GetLatestStartTime(_ context.Context) (int64, error) {
	return s.latestStartTime, s.latestErr
}

func (s *callRecordingRepoStub) UpsertBatch(
	_ context.Context,
	_ []model.CallRecordingUpsertInput,
) ([]model.CallRecording, error) {
	return []model.CallRecording{}, nil
}

func TestCallRecordingSyncServiceNormalizeInputDefaultsToFullSyncWithoutExistingData(t *testing.T) {
	t.Parallel()

	svc := &CallRecordingSyncService{
		recordingService: NewCallRecordingService(&callRecordingRepoStub{}),
		cookie:           "cookie-token",
		nowFunc: func() time.Time {
			return time.Date(2026, 3, 24, 10, 0, 0, 0, time.FixedZone("CST", 8*3600))
		},
	}

	normalized, err := svc.normalizeInput(context.Background(), SyncCallRecordingsInput{})
	if err != nil {
		t.Fatalf("normalizeInput returned error: %v", err)
	}

	if normalized.StartTimeBegin != feigeCallRecordingFullSyncDate {
		t.Fatalf("expected StartTimeBegin=%q, got %q", feigeCallRecordingFullSyncDate, normalized.StartTimeBegin)
	}
	if normalized.StartTimeFinish != "2026-03-24" {
		t.Fatalf("expected StartTimeFinish=2026-03-24, got %q", normalized.StartTimeFinish)
	}
	if normalized.Cookie != "cookie-token" {
		t.Fatalf("expected configured cookie to be used, got %q", normalized.Cookie)
	}
}

func TestCallRecordingSyncServiceNormalizeInputUsesLatestRecordingDate(t *testing.T) {
	t.Parallel()

	loc := time.FixedZone("CST", 8*3600)
	latest := time.Date(2026, 3, 20, 23, 59, 59, 0, loc).UnixMilli()
	svc := &CallRecordingSyncService{
		recordingService: NewCallRecordingService(&callRecordingRepoStub{
			latestStartTime: latest,
		}),
		cookie: "cookie-token",
		nowFunc: func() time.Time {
			return time.Date(2026, 3, 24, 10, 0, 0, 0, loc)
		},
	}

	normalized, err := svc.normalizeInput(context.Background(), SyncCallRecordingsInput{})
	if err != nil {
		t.Fatalf("normalizeInput returned error: %v", err)
	}

	if normalized.StartTimeBegin != "2026-03-20" {
		t.Fatalf("expected StartTimeBegin=2026-03-20, got %q", normalized.StartTimeBegin)
	}
	if normalized.StartTimeFinish != "2026-03-24" {
		t.Fatalf("expected StartTimeFinish=2026-03-24, got %q", normalized.StartTimeFinish)
	}
}
