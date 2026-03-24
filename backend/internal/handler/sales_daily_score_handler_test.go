package handler

import "testing"

func TestCanViewSalesDailyScoresSupportsAdminFinanceAndSalesRoles(t *testing.T) {
	t.Parallel()

	roles := []string{
		"admin",
		"管理员",
		"finance_manager",
		"财务经理",
		"sales_director",
		"销售总监",
		"sales_manager",
		"销售经理",
		"sales_staff",
		"销售员工",
		"sales_outside",
		"outside销售",
	}

	for _, role := range roles {
		if !canViewSalesDailyScores(role) {
			t.Fatalf("expected role %q to view sales daily scores", role)
		}
	}
}

func TestCanViewSalesDailyScoresRejectsInsideSalesRoles(t *testing.T) {
	t.Parallel()

	roles := []string{
		"sales_inside",
		"sale_inside",
		"销售",
		"Inside销售",
		"inside销售",
		"电销员工",
	}

	for _, role := range roles {
		if canViewSalesDailyScores(role) {
			t.Fatalf("expected role %q to be rejected from sales daily scores", role)
		}
	}
}
