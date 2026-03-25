package scoring

import (
	"backend/internal/model"
	"testing"
	"time"
)

func TestCalculateDailySalesScoreReachedAtUsesEarliestMomentFinalScoreWasReached(t *testing.T) {
	t.Parallel()

	loc := time.FixedZone("CST", 8*3600)
	reachedAt := CalculateDailySalesScoreReachedAt(
		BuildDailySalesScoreBreakdown(1, 50*60, 1, 3),
		[]model.DailySalesCallEvent{
			{
				UserID:         1,
				EventTime:      time.Date(2026, 3, 25, 8, 0, 0, 0, loc),
				DurationSecond: 50 * 60,
			},
		},
		[]time.Time{
			time.Date(2026, 3, 25, 7, 50, 0, 0, loc),
		},
		[]time.Time{
			time.Date(2026, 3, 25, 7, 45, 0, 0, loc),
			time.Date(2026, 3, 25, 7, 46, 0, 0, loc),
			time.Date(2026, 3, 25, 7, 47, 0, 0, loc),
			time.Date(2026, 3, 25, 9, 0, 0, 0, loc),
		},
	)

	if reachedAt == nil {
		t.Fatal("expected reachedAt to be calculated")
	}
	want := time.Date(2026, 3, 25, 8, 0, 0, 0, loc).UTC()
	if !reachedAt.Equal(want) {
		t.Fatalf("expected reachedAt %v, got %v", want, reachedAt)
	}
}

func TestCalculateDailySalesScoreReachedAtReturnsNilWhenNoScoreIsReached(t *testing.T) {
	t.Parallel()

	reachedAt := CalculateDailySalesScoreReachedAt(
		BuildDailySalesScoreBreakdown(0, 0, 0, 0),
		nil,
		nil,
		nil,
	)
	if reachedAt != nil {
		t.Fatalf("expected nil reachedAt for zero score, got %v", reachedAt)
	}
}
