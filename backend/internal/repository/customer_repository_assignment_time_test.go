package repository

import (
	"testing"
	"time"
)

func TestAssignedSalesUnix(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_711_111_111, 0).UTC()
	insideSalesUserID := int64(11)

	if got := assignedSalesUnix(&insideSalesUserID, insideSalesUserID, now); got != nil {
		t.Fatalf("expected nil assign time when customer remains with inside sales, got %v", *got)
	}

	got := assignedSalesUnix(&insideSalesUserID, 22, now)
	if got == nil {
		t.Fatal("expected assign time when customer is assigned to sales")
	}
	if *got != now.Unix() {
		t.Fatalf("expected assign time %d, got %d", now.Unix(), *got)
	}
}

func TestCollectTimeUnixOrNow(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_711_111_111, 0).UTC()
	existing := time.Unix(1_700_000_000, 0).UTC()

	if got := collectTimeUnixOrNow(&existing, now); got != existing.Unix() {
		t.Fatalf("expected existing collect time %d, got %d", existing.Unix(), got)
	}

	if got := collectTimeUnixOrNow(nil, now); got != now.Unix() {
		t.Fatalf("expected fallback collect time %d, got %d", now.Unix(), got)
	}
}
