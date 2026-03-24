package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type hanghangCRMDailyUserCallStatRepoStub struct {
	findErr error
	upserts []model.DailyUserCallStatUpsertInput
}

func (s *hanghangCRMDailyUserCallStatRepoStub) UpsertBatch(
	_ context.Context,
	items []model.DailyUserCallStatUpsertInput,
) ([]model.DailyUserCallStat, error) {
	s.upserts = append(s.upserts, items...)
	result := make([]model.DailyUserCallStat, 0, len(items))
	for idx, item := range items {
		result = append(result, model.DailyUserCallStat{
			ID:       int64(idx + 1),
			StatDate: item.StatDate,
			UserID:   item.UserID,
			RealName: item.RealName,
			Mobile:   item.Mobile,
			BindNum:  item.BindNum,
			CallNum:  item.CallNum,
		})
	}
	return result, nil
}

func (s *hanghangCRMDailyUserCallStatRepoStub) FindUserIDByNicknameAndHanghangCRMMobile(
	_ context.Context,
	_ string,
	_ string,
) (*int64, error) {
	if s.findErr != nil {
		return nil, s.findErr
	}
	return nil, nil
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestSyncDailyUserCallStatsUsesConfiguredTokenAndAllowsUnmatchedUser(t *testing.T) {
	t.Parallel()

	repoStub := &hanghangCRMDailyUserCallStatRepoStub{
		findErr: repository.ErrHanghangCRMUserNotMatched,
	}
	svc := &hanghangCRMDailyUserCallStatService{
		repo:       repoStub,
		cloudToken: "configured-token",
		client: &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if got := r.Header.Get("cloud-token"); got != "configured-token" {
					t.Fatalf("expected configured cloud-token, got %q", got)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body: io.NopCloser(strings.NewReader(
						`{"code":0,"data":{"records":[{"realName":"张三","mobile":"13800000000","bindNum":"3","callNum":"5","notConnected":1,"connectionRate":0.5,"timeTotal":"100","totalMinute":"1","totalSecond":"100","averageCallDuration":20,"averageCallSecond":20}],"current":1,"pages":1}}`,
					)),
				}, nil
			}),
		},
		endpoint: "https://hanghang-crm.test/sync",
	}

	result, err := svc.SyncDailyUserCallStats(context.Background(), SyncHanghangCRMDailyUserCallStatInput{
		StartTime: "2026-03-19",
		EndTime:   "2026-03-19",
	})
	if err != nil {
		t.Fatalf("SyncDailyUserCallStats returned error: %v", err)
	}

	if result.TotalFetched != 1 {
		t.Fatalf("expected TotalFetched=1, got %d", result.TotalFetched)
	}
	if result.TotalSaved != 1 {
		t.Fatalf("expected TotalSaved=1, got %d", result.TotalSaved)
	}
	if result.MatchedUserCount != 0 {
		t.Fatalf("expected MatchedUserCount=0, got %d", result.MatchedUserCount)
	}
	if result.UnmatchedUserCount != 1 {
		t.Fatalf("expected UnmatchedUserCount=1, got %d", result.UnmatchedUserCount)
	}
	if len(repoStub.upserts) != 1 {
		t.Fatalf("expected 1 upsert item, got %d", len(repoStub.upserts))
	}
	if repoStub.upserts[0].UserID != nil {
		t.Fatalf("expected unmatched user to be saved with nil user_id")
	}
}

func TestSyncDailyUserCallStatsRequiresTokenWhenNotConfigured(t *testing.T) {
	t.Parallel()

	svc := &hanghangCRMDailyUserCallStatService{
		repo: &hanghangCRMDailyUserCallStatRepoStub{},
	}

	_, err := svc.SyncDailyUserCallStats(context.Background(), SyncHanghangCRMDailyUserCallStatInput{
		StartTime: "2026-03-19",
		EndTime:   "2026-03-19",
	})
	if !errors.Is(err, ErrHanghangCRMCloudTokenRequired) {
		t.Fatalf("expected ErrHanghangCRMCloudTokenRequired, got %v", err)
	}
}
