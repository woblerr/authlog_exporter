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
		t.Errorf("\nVariables do not match: %s,\nwant: %s;\nendpoint: %s,\nwant: %s", testPort, promPort, testEndpoit, promEndpoint)
	}
}
