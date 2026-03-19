package repository

import (
	"backend/internal/model"
	"testing"
)

func TestBuildMyCustomerOwnershipConditionReturnsNoResultsForEmptyExplicitScope(t *testing.T) {
	condition, args := buildMyCustomerOwnershipCondition(model.CustomerListFilter{
		ViewerID:                  99,
		AllowedOwnerUserIDs:       []int64{-1},
		AllowedInsideSalesUserIDs: []int64{-1},
	})

	if condition != "1 = 0" {
		t.Fatalf("expected no-result condition, got %q", condition)
	}
	if args != nil {
		t.Fatalf("expected nil args, got %#v", args)
	}
}

func TestBuildMyCustomerOwnershipConditionFallsBackToViewerWithoutExplicitScope(t *testing.T) {
	condition, args := buildMyCustomerOwnershipCondition(model.CustomerListFilter{
		ViewerID: 99,
	})

	if condition != "c.owner_user_id = ?" {
		t.Fatalf("expected viewer fallback condition, got %q", condition)
	}
	if len(args) != 1 {
		t.Fatalf("expected one arg, got %#v", args)
	}
	viewerID, ok := args[0].(int64)
	if !ok || viewerID != 99 {
		t.Fatalf("expected viewer id 99, got %#v", args[0])
	}
}
