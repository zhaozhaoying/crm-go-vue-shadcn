package scoring

import (
	"backend/internal/model"
	"testing"
)

func TestBuildDailySalesScoreBreakdownUsesProgressiveRules(t *testing.T) {
	t.Parallel()

	breakdown := BuildDailySalesScoreBreakdown(90, 24*60, 2, 6)

	if breakdown.CallScoreByCount != 30 {
		t.Fatalf("expected call-count score 30, got %d", breakdown.CallScoreByCount)
	}
	if breakdown.CallScoreByDuration != 40 {
		t.Fatalf("expected duration score 40, got %d", breakdown.CallScoreByDuration)
	}
	if breakdown.CallScoreType != model.SalesDailyScoreCallScoreTypeDuration || breakdown.CallScore != 40 {
		t.Fatalf("unexpected chosen call score: %+v", breakdown)
	}
	if breakdown.VisitScore != 20 {
		t.Fatalf("expected visit score 20, got %d", breakdown.VisitScore)
	}
	if breakdown.NewCustomerScore != 20 {
		t.Fatalf("expected new customer score 20, got %d", breakdown.NewCustomerScore)
	}
	if breakdown.TotalScore != 80 {
		t.Fatalf("expected total score 80, got %d", breakdown.TotalScore)
	}
}

func TestChooseCallScorePrefersHigherScore(t *testing.T) {
	t.Parallel()

	scoreType, score := ChooseCallScore(50, 70)
	if scoreType != model.SalesDailyScoreCallScoreTypeDuration || score != 70 {
		t.Fatalf("unexpected duration preference: %s %d", scoreType, score)
	}

	scoreType, score = ChooseCallScore(60, 60)
	if scoreType != model.SalesDailyScoreCallScoreTypeCallNum || score != 60 {
		t.Fatalf("unexpected tie preference: %s %d", scoreType, score)
	}
}

func TestCallAndVisitScoresReachConfiguredCaps(t *testing.T) {
	t.Parallel()

	if score := CallCountScore(180); score != 70 {
		t.Fatalf("expected call count score 70, got %d", score)
	}
	if score := CallDurationScore(50 * 60); score != 70 {
		t.Fatalf("expected call duration score 70, got %d", score)
	}
	if score := VisitScore(5); score != 60 {
		t.Fatalf("expected visit score 60, got %d", score)
	}
	if score := NewCustomerScore(3); score != 10 {
		t.Fatalf("expected new customer score 10, got %d", score)
	}
}
