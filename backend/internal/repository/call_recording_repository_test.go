package repository

import (
	"backend/internal/model"
	"context"
	"testing"
)

func TestDedupeCallRecordingUpsertInputsUsesBusinessKey(t *testing.T) {
	t.Parallel()

	items := []model.CallRecordingUpsertInput{
		{
			ID:        "record-1",
			Mobile:    "13800000000",
			Phone:     "13900000000",
			TelA:      "13800000000",
			TelB:      "13900000000",
			StartTime: 1710000000000,
			CallType:  1,
			Duration:  120,
		},
		{
			ID:        "record-2",
			Mobile:    "13800000000",
			Phone:     "13900000000",
			TelA:      "13800000000",
			TelB:      "13900000000",
			StartTime: 1710000000000,
			CallType:  1,
			Duration:  120,
		},
	}

	got := dedupeCallRecordingUpsertInputs(items)
	if len(got) != 1 {
		t.Fatalf("expected 1 item after dedupe, got %d", len(got))
	}
	if got[0].ID != "record-2" {
		t.Fatalf("expected latest duplicate to be kept, got %q", got[0].ID)
	}
}

func TestBuildCallRecordingDedupeKeyStable(t *testing.T) {
	t.Parallel()

	item := model.CallRecordingUpsertInput{
		ID:        "record-1",
		Mobile:    " 13800000000 ",
		Phone:     " 13900000000 ",
		TelA:      " 13800000000 ",
		TelB:      " 13900000000 ",
		StartTime: 1710000000000,
		CallType:  1,
		Duration:  120,
	}

	got := buildCallRecordingDedupeKey(item)
	want := "1710000000000|13800000000|13900000000|13800000000|13900000000|1|120"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestFindExistingIDsByDedupeKeysEmptyInput(t *testing.T) {
	t.Parallel()

	repo := &gormCallRecordingRepository{}
	got, err := repo.findExistingIDsByDedupeKeys(context.Background(), nil)
	if err != nil {
		t.Fatalf("findExistingIDsByDedupeKeys returned error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty map, got %#v", got)
	}
}
