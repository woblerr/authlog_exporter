package promexporter

import (
	"log"

	"github.com/hpcloud/tail"
)

var (
	authlogPath string
)

// SetAuthlogPath sets path for 'auth.log' from command line argument 'auth.log'
func SetAuthlogPath(filePath string) {
	authlogPath = filePath
	log.Printf("Log for parsing %s", authlogPath)
}

func startParserAuthlog(filePath string) {
	t, err := tail.TailFile(filePath, tail.Config{
		Follow:    true,
		ReOpen:    true,
		MustExist: true})
	if err != nil {
		log.Fatal(err)
	}
	for line := range t.Lines {
		parseLine(line)
	}
}

// Start runs promhttp endpoind and parsing log process
func Start() {
	startPromEndpoint()
	startParserAuthlog(authlogPath)
}
