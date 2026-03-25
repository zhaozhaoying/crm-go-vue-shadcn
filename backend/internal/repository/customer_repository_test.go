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

func TestBuildCustomerAssignmentMeta(t *testing.T) {
	tests := []struct {
		name           string
		hasInsideSales bool
		reason         string
		isInPool       bool
		wantType       string
		wantLabel      string
	}{
		{
			name:      "self add",
			reason:    model.CustomerOwnerLogReasonCreateInitialAssign,
			wantType:  "self_add",
			wantLabel: "自己添加",
		},
		{
			name:      "import assign",
			reason:    model.CustomerOwnerLogReasonImportInitialAssign,
			wantType:  "import_assign",
			wantLabel: "导入分配",
		},
		{
			name:      "pool claim",
			reason:    model.CustomerOwnerLogReasonClaimFromPool,
			wantType:  "pool_claim",
			wantLabel: "公海领取",
		},
		{
			name:           "inside sales assign wins",
			hasInsideSales: true,
			reason:         model.CustomerOwnerLogReasonInsideSalesCreate,
			wantType:       "auto_assign",
			wantLabel:      "电销分配",
			isInPool:       false,
		},
		{
			name:      "manual transfer",
			reason:    model.CustomerOwnerLogReasonManualTransfer,
			wantType:  "manual_transfer",
			wantLabel: "手动转移",
		},
		{
			name:      "manual release in pool",
			reason:    model.CustomerOwnerLogReasonManualRelease,
			isInPool:  true,
			wantType:  "manual_release",
			wantLabel: "手动丢弃",
		},
		{
			name:      "fallback",
			reason:    "",
			wantType:  "",
			wantLabel: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotLabel := buildCustomerAssignmentMeta(
				tt.hasInsideSales,
				tt.reason,
				tt.isInPool,
			)
			if gotType != tt.wantType || gotLabel != tt.wantLabel {
				t.Fatalf("unexpected assignment meta: got (%q, %q), want (%q, %q)", gotType, gotLabel, tt.wantType, tt.wantLabel)
			}
		})
	}
}

func TestBuildCustomerListWhereIncludesOwnerUserID(t *testing.T) {
	where, args := buildCustomerListWhere(model.CustomerListFilter{
		Category:    "my",
		OwnerUserID: 42,
	})

	found := false
	for _, condition := range where {
		if condition == "c.owner_user_id = ?" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected owner_user_id condition, got %#v", where)
	}
	if len(args) == 0 {
		t.Fatalf("expected args to include owner user id, got %#v", args)
	}
	if ownerUserID, ok := args[0].(int64); !ok || ownerUserID != 42 {
		t.Fatalf("expected first arg to be owner user id 42, got %#v", args[0])
	}
}

func TestBuildCustomerListWhereUsesDropUserForPoolOwnerFilter(t *testing.T) {
	where, args := buildCustomerListWhere(model.CustomerListFilter{
		Category:    "pool",
		OwnerUserID: 88,
	})

	found := false
	for _, condition := range where {
		if condition == "du.id = ?" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected pool owner filter to use drop user join, got %#v", where)
	}
	if len(args) == 0 {
		t.Fatalf("expected args to include drop user id, got %#v", args)
	}
	if ownerUserID, ok := args[0].(int64); !ok || ownerUserID != 88 {
		t.Fatalf("expected first arg to be drop user id 88, got %#v", args[0])
	}
}
