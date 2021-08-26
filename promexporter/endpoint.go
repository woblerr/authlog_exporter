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

// SetPromPortandPath sets HTTP endpoint parameters from command line arguments 'port' and 'endpoint'.
func SetPromPortandPath(port, endpoint string) {
	promPort = port
	promEndpoint = endpoint
}

func startPromEndpoint() {
	go func() {
		http.Handle(promEndpoint, promhttp.Handler())
		log.Fatalf("[ERROR] Run HTTP endpoint failed, %v", http.ListenAndServe(":"+promPort, nil))
	}()
}
