package errmsg

import "testing"

func TestNormalizeMySQLColumnCannotBeNull(t *testing.T) {
	got := Normalize("Error 1048 (23000): Column 'next_time' cannot be null")
	if got != "下次跟进时间不能为空" {
		t.Fatalf("unexpected normalized message: %q", got)
	}
}
