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
	promPort    = flag.String("port", "9991", "Port for prometheus metrics to listen on")
	promPath    = flag.String("endpoint", "/metrics", "Endpoint used for metrics")
	authlogPath = flag.String("auth.log", "/var/log/auth.log", "Path to auth.log")
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
		log.Printf("RECEIVED SIGNAL: %s", s)
		log.Printf("Stopping : %s", filepath.Base(os.Args[0]))
		os.Exit(0)
	}()

	log.Printf("Starting %s", filepath.Base(os.Args[0]))
	log.Printf("Version: %s", version)

	// Setup parameters for exporter
	promexporter.SetPromPortandPath(*promPort, *promPath)
	promexporter.SetAuthlogPath(*authlogPath)

	// Start exporter
	promexporter.Start()
}
