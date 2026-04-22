package config

import (
	"os"
	"strings"
	"time"
)

// ApplyProcessTimezone makes the configured business timezone the process-local timezone,
// so all time.Now()/time.Local based day boundaries follow the expected business date.
func ApplyProcessTimezone(location *time.Location) {
	if location == nil {
		return
	}

	time.Local = location

	locationName := strings.TrimSpace(location.String())
	if locationName == "" || strings.EqualFold(locationName, "local") {
		return
	}
	_ = os.Setenv("TZ", locationName)
}
