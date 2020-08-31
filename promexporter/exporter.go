package promexporter

import (
	"log"

	"github.com/hpcloud/tail"
)

var (
	authlogPath = "/var/log/auth.log"
)

// SetAuthlogPath sets path for 'auth.log' from command line argument 'auth.log'
func SetAuthlogPath(filePath string) {
	authlogPath = filePath
}

func startParserAuthlog(filePath string) {

	log.Printf("Log for parsing: %s", authlogPath)

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
