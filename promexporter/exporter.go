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
// 'web.endpoint',
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

// Start runs promhttp endpoind and parsing log process.
func Start(logger *slog.Logger) {
	go func(logger *slog.Logger) {
		http.Handle(webEndpoint, promhttp.Handler())
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`<html>
			<head><title>AuthLog exporter</title></head>
			<body>
			<h1>AuthLog exporter</h1>
			<p><a href='` + webEndpoint + `'>Metrics</a></p>
			</body>
			</html>`))
		})
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
