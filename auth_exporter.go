package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/woblerr/prom_authlog_exporter/promexporter"
)

var (
	promPort    = flag.String("prom.port", "9991", "Port for prometheus metrics to listen on")
	promPath    = flag.String("prom.endpoint", "/metrics", "Endpoint used for metrics")
	authlogPath = flag.String("auth.log", "/var/log/auth.log", "Path to auth.log")
	geodbPath   = flag.String("geo.db", "", "Path to geoIP database file")
	geodbLang   = flag.String("geo.lang", "en", "Output language format")
	geodbURL    = flag.String("geo.url", "https://freegeoip.live/json/", "URL for geoIP database API ")
	geodbType   = flag.String("geo.type", "", "Type of geoIP database: db, url")
	version     = "development"
)

func main() {
	// Load command line arguments
	flag.Parse()
	// Setup signal catching
	sigs := make(chan os.Signal, 1)
	// Catch  listed signals
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	// Method invoked upon seeing signal
	go func() {
		s := <-sigs
		log.Printf("RECEIVED SIGNAL %s", s)
		log.Printf("Stopping  %s", filepath.Base(os.Args[0]))
		os.Exit(0)
	}()
	log.Printf("Starting %s", filepath.Base(os.Args[0]))
	log.Printf("Version %s", version)
	// Setup parameters for exporter
	promexporter.SetPromPortandPath(*promPort, *promPath)
	promexporter.SetAuthlogPath(*authlogPath)
	promexporter.SetGeodbPath(*geodbType, *geodbPath, *geodbLang, *geodbURL)
	// Start exporter
	promexporter.Start()
}
