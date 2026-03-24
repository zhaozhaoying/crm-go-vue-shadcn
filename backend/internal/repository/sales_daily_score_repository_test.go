package repository

import "testing"

func TestSalesDailyScoreRoleFiltersExcludeInsideSalesVariants(t *testing.T) {
	t.Parallel()

	excludedNames := []string{
		"sales_inside",
		"sale_inside",
	}
	for _, excluded := range excludedNames {
		if containsString(salesDailyScoreRoleNames, excluded) {
			t.Fatalf("expected salesDailyScoreRoleNames to exclude %q", excluded)
		}
	}

	excludedLabels := []string{
		"销售",
		"Inside销售",
		"电销员工",
	}
	for _, excluded := range excludedLabels {
		if containsString(salesDailyScoreRoleLabels, excluded) {
			t.Fatalf("expected salesDailyScoreRoleLabels to exclude %q", excluded)
		}
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
