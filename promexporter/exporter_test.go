package promexporter

import (
	"testing"
)

func TestSetAuthlogPath(t *testing.T) {
	var (
		testLog = "/test_log/auth.log"
	)
	SetAuthlogPath(testLog)

	if testLog != authlogPath {
		t.Errorf("Variables do not match: %s, want: %s", testLog, authlogPath)
	}
}
