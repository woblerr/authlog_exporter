package promexporter

import (
	"log/slog"
	"os"
	"testing"

	"github.com/prometheus/exporter-toolkit/web"
)

var logger = getLogger()

func TestSetExporterParams(t *testing.T) {
	var (
		testLog         = "/test_data/auth.log"
		testFlagsConfig = web.FlagConfig{
			WebListenAddresses: &([]string{":9991"}),
			WebSystemdSocket:   valToPtr(bool(false)),
			WebConfigFile:      valToPtr(string("")),
		}
		testEndpoint = "/metrics"
		testHideIP   = false
		testHideUser = false
	)
	SetExporterParams(testLog, testEndpoint, testFlagsConfig, testHideIP, testHideUser)
	if testFlagsConfig.WebListenAddresses != webFlagsConfig.WebListenAddresses ||
		testFlagsConfig.WebSystemdSocket != webFlagsConfig.WebSystemdSocket ||
		testFlagsConfig.WebConfigFile != webFlagsConfig.WebConfigFile ||
		testEndpoint != webEndpoint ||
		testHideIP != metricHideIP ||
		testHideUser != metricHideUser {
		t.Errorf("\nVariables do not match,\nlistenAddresses: %v, want: %v;\n"+
			"systemSocket: %v, want: %v;\nwebConfig: %v, want: %v;\nendpoint: %s, want: %s"+
			"\nhideIP: %v, want: %v;\nhideUser: %v, want: %v",
			ptrToVal(testFlagsConfig.WebListenAddresses), ptrToVal(webFlagsConfig.WebListenAddresses),
			ptrToVal(testFlagsConfig.WebSystemdSocket), ptrToVal(webFlagsConfig.WebSystemdSocket),
			ptrToVal(testFlagsConfig.WebConfigFile), ptrToVal(webFlagsConfig.WebConfigFile),
			testEndpoint, webEndpoint,
			testHideIP, metricHideIP,
			testHideUser, metricHideUser,
		)
	}
}

func getLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func ptrToVal[T any](v *T) T {
	return *v
}

func valToPtr[T any](v T) *T {
	return &v
}
