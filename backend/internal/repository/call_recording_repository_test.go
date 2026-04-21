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

func TestDedupeTelemarketingRecordingUpsertInputsKeepsReturnedIDsUnique(t *testing.T) {
	t.Parallel()

	items := []model.TelemarketingRecordingUpsertInput{
		{
			ID:       "same-id",
			CCNumber: "cc-001",
		},
		{
			ID:       "same-id",
			CCNumber: "cc-002",
		},
	}

	got := dedupeTelemarketingRecordingUpsertInputs(items)
	if len(got) != 2 {
		t.Fatalf("expected 2 items after dedupe, got %d", len(got))
	}
	if got[0].ID != "same-id" {
		t.Fatalf("expected first item to keep upstream id, got %q", got[0].ID)
	}
	if got[1].ID != "cc-002" {
		t.Fatalf("expected duplicate upstream id to fall back to cc_number, got %q", got[1].ID)
	}
	if got[0].ID == got[1].ID {
		t.Fatalf("expected telemarketing recording ids to be unique, got duplicate %q", got[0].ID)
	}
}

func TestDedupeTelemarketingRecordingUpsertInputsUsesFallbackWhenIDIsEmpty(t *testing.T) {
	t.Parallel()

	items := []model.TelemarketingRecordingUpsertInput{
		{
			ID:       "",
			CCNumber: "cc-003",
		},
	}

	got := dedupeTelemarketingRecordingUpsertInputs(items)
	if len(got) != 1 {
		t.Fatalf("expected 1 item after dedupe, got %d", len(got))
	}
	if got[0].ID != "cc-003" {
		t.Fatalf("expected empty upstream id to fall back to cc_number, got %q", got[0].ID)
	}
}
