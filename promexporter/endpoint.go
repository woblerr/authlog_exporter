package promexporter

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	promPort     = "9991"
	promEndpoint = "/metrics"
)

// SetPromPortandPath sets HTTP endpoint parameters from command line arguments 'port' and 'endpoint'
func SetPromPortandPath(port, endpoint string) {
	promPort = port
	promEndpoint = endpoint
}

func startPromEndpoint() {

	log.Printf("Use port: %s and HTTP endpoint: %s", promPort, promEndpoint)

	go func() {
		http.Handle(promEndpoint, promhttp.Handler())
		log.Fatalln(http.ListenAndServe(":"+promPort, nil))
	}()
}
