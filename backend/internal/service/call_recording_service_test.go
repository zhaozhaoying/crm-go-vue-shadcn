package service

import (
	"backend/internal/model"
	"context"
	"testing"
)

type callRecordingMaskRepoStub struct {
	listResult  model.CallRecordingListResult
	listErr     error
	findItem    *model.CallRecording
	findErr     error
	upsertItems []model.CallRecording
	upsertErr   error
}

func (s *callRecordingMaskRepoStub) List(
	_ context.Context,
	_ model.CallRecordingListFilter,
) (model.CallRecordingListResult, error) {
	return s.listResult, s.listErr
}

func (s *callRecordingMaskRepoStub) FindByID(
	_ context.Context,
	_ string,
	_ bool,
	_ string,
) (*model.CallRecording, error) {
	return s.findItem, s.findErr
}

func (s *callRecordingMaskRepoStub) GetLatestStartTime(_ context.Context) (int64, error) {
	return 0, nil
}

func (s *callRecordingMaskRepoStub) UpsertBatch(
	_ context.Context,
	_ []model.CallRecordingUpsertInput,
) ([]model.CallRecording, error) {
	return s.upsertItems, s.upsertErr
}

func TestCallRecordingServiceListMasksPhoneFields(t *testing.T) {
	t.Parallel()

	repo := &callRecordingMaskRepoStub{
		listResult: model.CallRecordingListResult{
			Items: []model.CallRecording{{
				ID:       "rec-1",
				RealName: "静欣",
				Mobile:   "13302002373",
				Phone:    "13313263399",
				TelA:     "13302002373",
				TelB:     "13313263399",
				TelX:     "03184669360",
			}},
			Total:    1,
			Page:     1,
			PageSize: 20,
		},
	}

	svc := NewCallRecordingService(repo)
	result, err := svc.List(context.Background(), model.CallRecordingListFilter{})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}

	item := result.Items[0]
	if item.RealName != "**" {
		t.Fatalf("expected masked realName, got %q", item.RealName)
	}
	if item.Mobile != "133****2373" {
		t.Fatalf("expected masked mobile, got %q", item.Mobile)
	}
	if item.Phone != "133****3399" {
		t.Fatalf("expected masked phone, got %q", item.Phone)
	}
	if item.TelA != "133****2373" {
		t.Fatalf("expected masked telA, got %q", item.TelA)
	}
	if item.TelB != "133****3399" {
		t.Fatalf("expected masked telB, got %q", item.TelB)
	}
	if item.TelX != "031****9360" {
		t.Fatalf("expected masked telX, got %q", item.TelX)
	}
}

func TestCallRecordingServiceGetByIDMasksPhoneFields(t *testing.T) {
	t.Parallel()

	repo := &callRecordingMaskRepoStub{
		findItem: &model.CallRecording{
			ID:       "rec-2",
			RealName: "王永刚",
			Mobile:   "13302002373",
			Phone:    "15122068960",
			TelA:     "13302002373",
			TelB:     "15122068960",
		},
	}

	svc := NewCallRecordingService(repo)
	item, err := svc.GetByID(context.Background(), "rec-2", true, "")
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if item == nil {
		t.Fatal("expected item, got nil")
	}
	if item.RealName != "**" {
		t.Fatalf("expected masked realName, got %q", item.RealName)
	}
	if item.Mobile != "133****2373" || item.Phone != "151****8960" {
		t.Fatalf("expected masked numbers, got %+v", item)
	}
	if item.TelA != "133****2373" || item.TelB != "151****8960" {
		t.Fatalf("expected masked tel fields, got %+v", item)
	}
}

func TestCallRecordingServiceUpsertBatchMasksReturnedPhoneFields(t *testing.T) {
	t.Parallel()

	repo := &callRecordingMaskRepoStub{
		upsertItems: []model.CallRecording{{
			ID:       "rec-3",
			RealName: "李龙泉",
			Mobile:   "13820039829",
			Phone:    "03184669360",
			TelA:     "13820039829",
			TelB:     "03184669360",
		}},
	}

	svc := NewCallRecordingService(repo)
	items, err := svc.UpsertBatch(context.Background(), []model.CallRecordingUpsertInput{{ID: "rec-3"}})
	if err != nil {
		t.Fatalf("UpsertBatch returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].RealName != "**" {
		t.Fatalf("expected masked realName, got %q", items[0].RealName)
	}
	if items[0].Mobile != "138****9829" || items[0].Phone != "031****9360" {
		t.Fatalf("expected masked upsert result, got %+v", items[0])
	}
}
