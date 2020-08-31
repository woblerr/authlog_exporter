package promexporter

import (
	"testing"
)

func TestSetPromPortandPath(t *testing.T) {
	var (
		testPort    = "9991"
		testEndpoit = "/metrics"
	)
	SetPromPortandPath(testPort, testEndpoit)

	if testPort != promPort || testEndpoit != promEndpoint {
		t.Errorf("Variables do not match: %s, want: %s; %s, want: %s", testPort, promPort, testEndpoit, promEndpoint)
	}
}
