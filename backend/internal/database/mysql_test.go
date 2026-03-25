package database

import (
	"testing"
	"time"
)

func TestMySQLSessionTimeZoneValueUsesPositiveOffset(t *testing.T) {
	location := time.FixedZone("Asia/Shanghai", 8*60*60)

	if got := mysqlSessionTimeZoneValue(location); got != "+08:00" {
		t.Fatalf("expected +08:00, got %q", got)
	}
}

func TestMySQLSessionTimeZoneValueUsesNegativeOffset(t *testing.T) {
	location := time.FixedZone("America/Los_Angeles", -7*60*60)

	if got := mysqlSessionTimeZoneValue(location); got != "-07:00" {
		t.Fatalf("expected -07:00, got %q", got)
	}
}
