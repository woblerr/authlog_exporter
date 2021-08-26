package main

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/woblerr/prom_authlog_exporter/promexporter"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var version = "unknown"

func main() {
	var (
		promPort = kingpin.Flag(
			"prom.port",
			"Port for prometheus metrics to listen on.",
		).Default("9991").String()
		promPath = kingpin.Flag(
			"prom.endpoint",
			"Endpoint used for metrics.",
		).Default("/metrics").String()
		authlogPath = kingpin.Flag(
			"auth.log",
			"Path to auth.log.",
		).Default("/var/log/auth.log").String()
		geodbPath = kingpin.Flag(
			"geo.db",
			"Path to geoIP database file.",
		).Default("").String()
		geodbLang = kingpin.Flag(
			"geo.lang",
			"Output language format.",
		).Default("en").String()
		geodbURL = kingpin.Flag(
			"geo.url",
			"URL for geoIP database API.",
		).Default("https://freegeoip.live/json/").String()
		geodbType = kingpin.Flag(
			"geo.type",
			"Type of geoIP database: db, url.",
		).Default("").String()
	)
	// Load command line arguments.
	kingpin.Parse()
	// Setup signal catching.
	sigs := make(chan os.Signal, 1)
	// Catch  listed signals.
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// Method invoked upon seeing signal.
	go func() {
		s := <-sigs
		log.Printf("[WARN] RECEIVED SIGNAL %s", s)
		log.Printf("[WARN] Stopping  %s", filepath.Base(os.Args[0]))
		os.Exit(1)
	}()
	log.Printf("[INFO] Starting %s", filepath.Base(os.Args[0]))
	log.Printf("[INFO] Version %s", version)
	// Setup parameters for exporter.
	promexporter.SetPromPortandPath(*promPort, *promPath)
	log.Printf("[INFO] Use port %s and HTTP endpoint %s", *promPort, *promPath)
	promexporter.SetAuthlogPath(*authlogPath)
	log.Printf("[INFO] Log for parsing %s", *authlogPath)
	promexporter.SetGeodbPath(*geodbType, *geodbPath, *geodbLang, *geodbURL)
	// Start exporter.
	promexporter.Start()
}
