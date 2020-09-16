package promexporter

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	promPort     string
	promEndpoint string
)

// SetPromPortandPath sets HTTP endpoint parameters from command line arguments 'port' and 'endpoint'
func SetPromPortandPath(port, endpoint string) {
	promPort = port
	promEndpoint = endpoint
	log.Printf("Use port %s and HTTP endpoint %s", promPort, promEndpoint)
}

func startPromEndpoint() {
	go func() {
		http.Handle(promEndpoint, promhttp.Handler())
		log.Fatalln(http.ListenAndServe(":"+promPort, nil))
	}()
}
