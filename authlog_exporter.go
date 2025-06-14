package main

import (
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	version_collector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"github.com/woblerr/authlog_exporter/promexporter"
)

const exporterName = "authlog_exporter"

func main() {
	var (
		authlogPath = kingpin.Flag(
			"auth.log",
			"Path to auth.log.",
		).Default("/var/log/auth.log").String()
		webPath = kingpin.Flag(
			"web.telemetry-path",
			"Path under which to expose metrics.",
		).Default("/metrics").String()
		webAdditionalToolkitFlags = kingpinflag.AddFlags(kingpin.CommandLine, ":9991")
		geodbPath                 = kingpin.Flag(
			"geo.db",
			"Path to geoIP database file.",
		).Default("").String()
		geodbLang = kingpin.Flag(
			"geo.lang",
			"Output language format.",
		).Default("en").String()
		geodbTimeout = kingpin.Flag(
			"geo.timeout",
			"Timeout in seconds for waiting response from geoIP database API.",
		).Default("2").Int()
		geodbType = kingpin.Flag(
			"geo.type",
			"Type of geoIP database: db, url.",
		).Default("").String()
		geodbURL = kingpin.Flag(
			"geo.url",
			"URL for geoIP database API.",
		).Default("https://reallyfreegeoip.org/json/").String()
		metricHideIP = kingpin.Flag(
			"metric.hideip",
			"Set this flag to hide IPs in the output and therefore drastically reduce the amount of metrics published.",
		).Bool()
		metricHideUser = kingpin.Flag(
			"metric.hideuser",
			"Set this flag to hide user accounts in the output and therefore drastically reduce the amount of metrics published.",
		).Bool()
	)
	// Set logger config.
	promslogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promslogConfig)
	kingpin.Version(version.Print(exporterName))
	// Add short help flag.
	kingpin.HelpFlag.Short('h')
	// Load command line arguments.
	kingpin.Parse()
	// Setup signal catching.
	sigs := make(chan os.Signal, 1)
	// Catch  listed signals.
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// Set logger.
	logger := promslog.New(promslogConfig)
	// Method invoked upon seeing signal.
	go func(logger *slog.Logger) {
		s := <-sigs
		logger.Warn(
			"Stopping exporter",
			"name", filepath.Base(os.Args[0]),
			"signal", s)
		os.Exit(1)
	}(logger)
	logger.Info(
		"Starting exporter",
		"name", filepath.Base(os.Args[0]),
		"version", version.Info(),
	)
	logger.Info("Build context", "build_context", version.BuildContext())
	// Setup parameters for exporter.
	promexporter.SetExporterParams(*authlogPath, *webPath, *webAdditionalToolkitFlags, *metricHideIP, *metricHideUser)
	logger.Info(
		"Use exporter parameters",
		"authlog", *authlogPath,
		"endpoint", *webPath,
		"config.file", *webAdditionalToolkitFlags.WebConfigFile,
		"hideip", *metricHideIP,
		"hideuser", *metricHideUser,
	)
	// Exporter build info metric.
	prometheus.MustRegister(version_collector.NewCollector(exporterName))
	promexporter.SetGeodbPath(*geodbType, *geodbPath, *geodbLang, *geodbURL, *geodbTimeout, logger)
	// Start exporter.
	promexporter.Start(version.Info(), logger)
}
