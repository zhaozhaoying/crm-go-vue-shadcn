package config

import (
	"os"
	"testing"
	"time"
)

func TestApplyProcessTimezoneSetsLocalAndTZ(t *testing.T) {
	originalLocal := time.Local
	originalTZ, hadTZ := os.LookupEnv("TZ")
	t.Cleanup(func() {
		time.Local = originalLocal
		if hadTZ {
			_ = os.Setenv("TZ", originalTZ)
			return
		}
		_ = os.Unsetenv("TZ")
	})

	location := time.FixedZone("Asia/Shanghai", 8*60*60)
	ApplyProcessTimezone(location)

	if time.Local != location {
		t.Fatalf("expected time.Local to be updated")
	}
	if got := os.Getenv("TZ"); got != "Asia/Shanghai" {
		t.Fatalf("expected TZ to be Asia/Shanghai, got %q", got)
	}
}
