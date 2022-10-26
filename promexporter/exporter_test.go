package promexporter

import (
	"testing"

	"github.com/go-kit/log"
	"github.com/prometheus/common/promlog"
)

var logger = getLogger()

func TestSetExporterParams(t *testing.T) {
	var (
		testLog           = "/test_data/auth.log"
		testPort          = "9991"
		testEndpoit       = "/metrics"
		testTLSConfigPath = ""
		testHideIP        = false
		testHideUser      = false
	)
	SetExporterParams(testLog, testPort, testEndpoit, testTLSConfigPath, testHideIP, testHideUser)
	if testLog != authlogPath || testPort != promPort || testEndpoit != promEndpoint || testTLSConfigPath != promTLSConfigPath {
		t.Errorf("\nVariables do not match,\nlog: %s, want: %s;\nport: %s, want: %s;\nendpoint: %s, want: %s;\nconfig: %s, want: %s",
			testLog, authlogPath,
			testPort, promPort,
			testEndpoit, promEndpoint,
			testTLSConfigPath, promTLSConfigPath,
		)
	}
}

func getLogger() log.Logger {
	var err error
	logLevel := &promlog.AllowedLevel{}
	err = logLevel.Set("info")
	if err != nil {
		panic(err)
	}
	promlogConfig := &promlog.Config{}
	promlogConfig.Level = logLevel
	if err != nil {
		panic(err)
	}
	return promlog.New(promlogConfig)
}
