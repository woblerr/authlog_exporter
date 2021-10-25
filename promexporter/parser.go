package promexporter

import (
	"regexp"

	"github.com/hpcloud/tail"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	authLinePrefix  = "^(?P<date>[A-Z][a-z]{2}\\s+\\d{1,2}) (?P<time>(\\d{2}:?){3}) (?P<host>[a-zA-Z0-9_\\-\\.]+) (?P<ident>[a-zA-Z0-9_\\-]+)(\\[(?P<pid>\\d+)\\])?: "
	authLineRegexps = map[string]*regexp.Regexp{
		"authAccepted":     regexp.MustCompile(authLinePrefix + "Accepted (password|publickey) for (?P<user>.*) from (?P<ipAddress>.*) port"),
		"authFailed":       regexp.MustCompile(authLinePrefix + "Failed (password|publickey) for (invalid user )?(?P<user>.*) from (?P<ipAddress>.*) port"),
		"invalidUser":      regexp.MustCompile(authLinePrefix + "Invalid user (?P<user>.*) from (?P<ipAddress>.*) port"),
		"notAllowedUser":   regexp.MustCompile(authLinePrefix + "User (?P<user>.*) from (?P<ipAddress>.*) not allowed because"),
		"connectionClosed": regexp.MustCompile(authLinePrefix + "Connection closed by authenticating user (?P<user>.*) (?P<ipAddress>.*) port"),
	}
	authVentsMetric = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "authlog_exporter_auth_events",
		Help: "The total number of auth events by user and IP addresses",
	},
		[]string{"eventType", "user", "ipAddress", "countyISOCode", "countryName", "cityName"})
)

type authLogLine struct {
	Type      string
	Username  string
	IPAddress string
}

func parseLine(line *tail.Line) {
	parsedLog := &authLogLine{}
	matches := make(map[string]string)
	// Find the type of log and parse it
	for t, re := range authLineRegexps {
		if re.MatchString(line.Text) {
			parsedLog.Type = t
			matches = getMatches(line.Text, re)
			continue
		}
	}
	// Skip if not matching.
	if len(matches) == 0 {
		return
	}
	geoIPData := &geoInfo{}
	parsedLog.Username = matches["user"]
	parsedLog.IPAddress = matches["ipAddress"]
	// Get geo information
	if geodbIs {
		if geodbType == "db" {
			getIPDetailsFromLocalDB(geoIPData, parsedLog.IPAddress)
		} else {
			getIPDetailsFromURL(geoIPData, parsedLog.IPAddress)
		}
	}
	// Add metric.
	authVentsMetric.WithLabelValues(
		parsedLog.Type,
		parsedLog.Username,
		parsedLog.IPAddress,
		geoIPData.countyISOCode,
		geoIPData.countryName,
		geoIPData.cityName,
	).Inc()
}

func getMatches(line string, re *regexp.Regexp) map[string]string {
	matches := re.FindStringSubmatch(line)
	results := make(map[string]string)
	// Get the basic information out of the log
	for i, name := range re.SubexpNames() {
		if i != 0 && len(matches) > i {
			results[name] = matches[i]
		}
	}
	return results
}
