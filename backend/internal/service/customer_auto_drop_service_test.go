package service

import "testing"

func TestAutoDropReasonUsesSalesAssignDealRule(t *testing.T) {
	t.Parallel()

	triggerType, reason := autoDropReason(30, 60, 30, false, false, true)
	if triggerType != 4 {
		t.Fatalf("expected trigger type 4, got %d", triggerType)
	}
	expected := "系统自动掉库：电销分配后超过30天未签单"
	if reason != expected {
		t.Fatalf("expected reason %q, got %q", expected, reason)
	}
}

func TestAutoDropReasonUsesCombinedSalesAssignRule(t *testing.T) {
	t.Parallel()

	triggerType, reason := autoDropReason(7, 60, 30, true, false, true)
	if triggerType != 4 {
		t.Fatalf("expected trigger type 4, got %d", triggerType)
	}
	expected := "系统自动掉库：超过7天未跟进且电销分配后超过30天未签单"
	if reason != expected {
		t.Fatalf("expected reason %q, got %q", expected, reason)
	}
}
