package promexporter

import (
	"testing"
)

func TestSetAuthlogPath(t *testing.T) {
	var (
		testLog = "/test_data/auth.log"
	)
	SetAuthlogPath(testLog)
	if testLog != authlogPath {
		t.Errorf("Variables do not match: %s,\nwant: %s", testLog, authlogPath)
	}
}
