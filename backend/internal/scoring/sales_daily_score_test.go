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
	if breakdown.NewCustomerScore != 10 {
		t.Fatalf("expected new customer score 10, got %d", breakdown.NewCustomerScore)
	}
	if breakdown.TotalScore != 70 {
		t.Fatalf("expected total score 70, got %d", breakdown.TotalScore)
	}
}

func TestBuildDailyTelemarketingScoreBreakdownUsesAnsweredCallCount(t *testing.T) {
	t.Parallel()

	breakdown := BuildDailyTelemarketingScoreBreakdown(120, 10*60, 2, 3)

	if breakdown.CallScoreByCount != 40 {
		t.Fatalf("expected answered-call score 40, got %d", breakdown.CallScoreByCount)
	}
	if breakdown.CallScoreByDuration != 10 {
		t.Fatalf("expected duration score 10, got %d", breakdown.CallScoreByDuration)
	}
	if breakdown.CallScoreType != model.SalesDailyScoreCallScoreTypeCallNum || breakdown.CallScore != 40 {
		t.Fatalf("unexpected chosen telemarketing call score: %+v", breakdown)
	}
	if breakdown.InvitationScore != 20 {
		t.Fatalf("expected invitation score 20, got %d", breakdown.InvitationScore)
	}
	if breakdown.NewCustomerScore != 10 {
		t.Fatalf("expected new customer score 10, got %d", breakdown.NewCustomerScore)
	}
	if breakdown.TotalScore != 70 {
		t.Fatalf("expected total score 70, got %d", breakdown.TotalScore)
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

func TestDailySalesScoresReachConfiguredCaps(t *testing.T) {
	t.Parallel()

	if score := CallCountScore(180); score != 70 {
		t.Fatalf("expected call count score 70, got %d", score)
	}
	if score := CallCountScore(260); score != 70 {
		t.Fatalf("expected capped call count score 70, got %d", score)
	}
	if score := CallDurationScore(50 * 60); score != 70 {
		t.Fatalf("expected call duration score 70, got %d", score)
	}
	if score := CallDurationScore(90 * 60); score != 70 {
		t.Fatalf("expected capped call duration score 70, got %d", score)
	}
	if score := VisitScore(5); score != 60 {
		t.Fatalf("expected visit score 60, got %d", score)
	}
	if score := VisitScore(10); score != 60 {
		t.Fatalf("expected capped visit score 60, got %d", score)
	}
	if score := NewCustomerScore(3); score != 10 {
		t.Fatalf("expected new customer score 10, got %d", score)
	}
	if score := NewCustomerScore(9); score != 10 {
		t.Fatalf("expected capped new customer score 10, got %d", score)
	}
}

func TestNewCustomerScoreRequiresThreeCustomers(t *testing.T) {
	t.Parallel()

	if score := NewCustomerScore(1); score != 0 {
		t.Fatalf("expected new customer score 0 for 1 customer, got %d", score)
	}
	if score := NewCustomerScore(2); score != 0 {
		t.Fatalf("expected new customer score 0 for 2 customers, got %d", score)
	}
	if score := NewCustomerScore(3); score != 10 {
		t.Fatalf("expected new customer score 10 for 3 customers, got %d", score)
	}
}
