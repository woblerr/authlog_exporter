package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/woblerr/authlog_exporter/promexporter"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var version = "unknown"

func main() {
	var (
		authlogPath = kingpin.Flag(
			"auth.log",
			"Path to auth.log.",
		).Default("/var/log/auth.log").String()
		promPath = kingpin.Flag(
			"prom.endpoint",
			"Endpoint used for metrics.",
		).Default("/metrics").String()
		promPort = kingpin.Flag(
			"prom.port",
			"Port for prometheus metrics to listen on.",
		).Default("9991").String()
		promTLSConfigFile = kingpin.Flag(
			"prom.web-config",
			"[EXPERIMENTAL] Path to config yaml file that can enable TLS or authentication.",
		).Default("").String()
		geodbPath = kingpin.Flag(
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
		).Default("https://freegeoip.live/json/").String()
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
	promlogConfig := &promlog.Config{}
	// Add flags log.level and log.format from promlog package.
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	// Add short help flag.
	kingpin.HelpFlag.Short('h')
	// Load command line arguments.
	kingpin.Parse()
	// Setup signal catching.
	sigs := make(chan os.Signal, 1)
	// Catch  listed signals.
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// Set logger.
	logger := promlog.New(promlogConfig)
	// Method invoked upon seeing signal.
	go func(logger log.Logger) {
		s := <-sigs
		level.Warn(logger).Log(
			"msg", "Stopping exporter",
			"name", filepath.Base(os.Args[0]),
			"signal", s)
		os.Exit(1)
	}(logger)
	level.Info(logger).Log(
		"msg", "Starting exporter",
		"name", filepath.Base(os.Args[0]),
		"version", version,
	)
	// Setup parameters for exporter.
	promexporter.SetExporterParams(*authlogPath, *promPort, *promPath, *promTLSConfigFile, *metricHideIP, *metricHideUser)
	level.Info(logger).Log(
		"authlog", *authlogPath,
		"mgs", "Use port and HTTP endpoint",
		"port", *promPort,
		"endpoint", *promPath,
		"web-config", *promTLSConfigFile,
	)
	promexporter.SetGeodbPath(*geodbType, *geodbPath, *geodbLang, *geodbURL, *geodbTimeout, logger)
	// Start exporter.
	promexporter.Start(logger)
}
