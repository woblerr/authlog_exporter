package promexporter

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/nxadm/tail"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/exporter-toolkit/web"
)

var (
	authlogPath    string
	webFlagsConfig web.FlagConfig
	webEndpoint    string
	metricHideIP   bool
	metricHideUser bool
)

// SetExporterParams sets path for 'auth.log' from command line argument 'auth.log',
// HTTP endpoint parameters from command line arguments:
// 'web.telemetry-path',
// 'web.listen-address',
// 'web.config.file',
// 'web.systemd-socket' (Linux only)'.
func SetExporterParams(filePath, endpoint string, flagsConfig web.FlagConfig, hideIP, hideUser bool) {
	authlogPath = filePath
	webFlagsConfig = flagsConfig
	webEndpoint = endpoint
	metricHideIP = hideIP
	metricHideUser = hideUser
}

// Start runs promhttp endpoint and parsing log process.
func Start(version string, logger *slog.Logger) {
	go func(logger *slog.Logger) {
		if webEndpoint == "" {
			logger.Error("Metric endpoint is empty", "endpoint", webEndpoint)
		}
		http.Handle(webEndpoint, promhttp.Handler())
		if webEndpoint != "/" {
			landingConfig := web.LandingConfig{
				Name:        "AuthLog exporter",
				Description: "Prometheus exporter for AuthLog",
				HeaderColor: "#476b6b",
				Version:     version,
				Links: []web.LandingLinks{
					{
						Address: webEndpoint,
						Text:    "Metrics",
					},
				},
			}
			landingPage, err := web.NewLandingPage(landingConfig)
			if err != nil {
				logger.Error("Error creating landing page", "err", err)
				os.Exit(1)
			}
			http.Handle("/", landingPage)
		}
		server := &http.Server{
			ReadHeaderTimeout: 5 * time.Second,
		}
		if err := web.ListenAndServe(server, &webFlagsConfig, logger); err != nil {
			logger.Error("Run web endpoint failed", "err", err)
			os.Exit(1)
		}
	}(logger)
	t, err := tail.TailFile(authlogPath, tail.Config{
		Follow:    true,
		ReOpen:    true,
		MustExist: true})
	if err != nil {
		logger.Error("Open log file failed", "err", err)
		os.Exit(1)
	}
	for line := range t.Lines {
		parseLine(line, logger)
	}
}
