package promexporter

import (
	"testing"
)

func TestSetExporterParams(t *testing.T) {
	var (
		testLog           = "/test_data/auth.log"
		testPort          = "9991"
		testEndpoit       = "/metrics"
		testTLSConfigPath = ""
	)
	SetExporterParams(testLog, testPort, testEndpoit, testTLSConfigPath)
	if testLog != authlogPath || testPort != promPort || testEndpoit != promEndpoint || testTLSConfigPath != promTLSConfigPath {
		t.Errorf("\nVariables do not match,\nlog: %s, want: %s;\nport: %s, want: %s;\nendpoint: %s, want: %s;\nconfig: %s, want: %s",
			testLog, authlogPath,
			testPort, promPort,
			testEndpoit, promEndpoint,
			testTLSConfigPath, promTLSConfigPath,
		)
	}
}
