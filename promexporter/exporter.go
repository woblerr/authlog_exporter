package promexporter

import (
	"net/http"
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/hpcloud/tail"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/exporter-toolkit/web"
)

var (
	authlogPath       string
	promPort          string
	promEndpoint      string
	promTLSConfigPath string
	metricHideIP      bool
	metricHideUser    bool
)

// SetExporterParams sets path for 'auth.log' from command line argument 'auth.log',
// HTTP endpoint parameters from command line arguments 'port', 'endpoint' and 'tlsConfigPath'.
func SetExporterParams(filePath, port, endpoint, tlsConfigPath string, hideIP, hideUser bool) {
	authlogPath = filePath
	promPort = port
	promEndpoint = endpoint
	promTLSConfigPath = tlsConfigPath
	metricHideIP = hideIP
	metricHideUser = hideUser
}

// Start runs promhttp endpoind and parsing log process.
func Start(logger log.Logger) {
	go func(logger log.Logger) {
		http.Handle(promEndpoint, promhttp.Handler())
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`<html>
			<head><title>AuthLog exporter</title></head>
			<body>
			<h1>AuthLog exporter</h1>
			<p><a href='` + promEndpoint + `'>Metrics</a></p>
			</body>
			</html>`))
		})
		server := &http.Server{Addr: ":" + promPort}
		if err := web.ListenAndServe(server, promTLSConfigPath, logger); err != nil {
			level.Error(logger).Log("msg", "Run web endpoint failed", "err", err)
			os.Exit(1)
		}
	}(logger)
	t, err := tail.TailFile(authlogPath, tail.Config{
		Follow:    true,
		ReOpen:    true,
		MustExist: true})
	if err != nil {
		level.Error(logger).Log("msg", "Open log file failed", "err", err)
		os.Exit(1)
	}
	for line := range t.Lines {
		parseLine(line, logger)
	}
}
