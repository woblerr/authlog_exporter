package promexporter

import (
	"regexp"

	"github.com/go-kit/log"
	"github.com/nxadm/tail"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	authLinePrefix  = "^(?P<date>[A-Z][a-z]{2}\\s+\\d{1,2}) (?P<time>(\\d{2}:?){3}) (?P<host>[a-zA-Z0-9_\\-\\.]+) (?P<ident>[a-zA-Z0-9_\\-]+)(\\[(?P<pid>\\d+)\\])?: "
	authLineRegexps = map[string]*regexp.Regexp{
		"authAccepted":                  regexp.MustCompile(authLinePrefix + "Accepted (password|publickey) for (?P<user>.*) from (?P<ipAddress>.*) port"),
		"authFailed":                    regexp.MustCompile(authLinePrefix + "Failed (password|publickey) for (invalid user )?(?P<user>.*) from (?P<ipAddress>.*) port"),
		"invalidUser":                   regexp.MustCompile(authLinePrefix + "Invalid user (?P<user>.*) from (?P<ipAddress>.*) port"),
		"notAllowedUser":                regexp.MustCompile(authLinePrefix + "User (?P<user>.*) from (?P<ipAddress>.*) not allowed because"),
		"connectionClosed":              regexp.MustCompile(authLinePrefix + "Connection closed by authenticating user (?P<user>.*) (?P<ipAddress>.*) port"),
		"sudoIncorrectPasswordAttempts": regexp.MustCompile(authLinePrefix + "[ ]+(?P<user>.*) : (?P<attempts>\\d+) incorrect password attempts ; TTY=(?P<tty>[^ ]+) ; PWD=(?P<pwd>.+) ; USER=(?P<user_as>.*) ; COMMAND=(?P<command>.*)"),
		"sudoNotInSudoers":              regexp.MustCompile(authLinePrefix + "[ ]+(?P<user>.*) : user NOT in sudoers ; TTY=(?P<tty>[^ ]+) ; PWD=(?P<pwd>.+) ; USER=(?P<user_as>.*) ; COMMAND=(?P<command>.*)"),
		"sudoSucceeded":                 regexp.MustCompile(authLinePrefix + "[ ]+(?P<user>.*) : TTY=(?P<tty>[^ ]+) ; PWD=(?P<pwd>.+) ; USER=(?P<user_as>.*) ; COMMAND=(?P<command>.*)"),
		"suSucceeded":                   regexp.MustCompile(authLinePrefix + "\\(to (?P<user_as>.*)\\) (?P<user>.*) on (?P<tty>[^ ]+)"),
		"suFailed":                      regexp.MustCompile(authLinePrefix + "FAILED SU \\(to (?P<user_as>.*)\\) (?P<user>.*) on (?P<tty>[^ ]+)"),
	}
	authVentsMetric = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "authlog_events_total",
		Help: "The total number of auth events.",
	},
		[]string{"eventType", "user", "ipAddress", "countyISOCode", "countryName", "cityName"})
)

type authLogLine struct {
	Type      string
	Username  string
	IPAddress string
}

func parseLine(line *tail.Line, logger log.Logger) {
	parsedLog := authLogLine{}
	matches := make(map[string]string)
	// Find the type of log and parse it.
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
	parsedLog.Username = matches["user"]
	parsedLog.IPAddress = matches["ipAddress"]
	geoIPData := geoInfo{}
	// Get geo information.
	if geodbIs {
		if geodbType == "db" {
			getIPDetailsFromLocalDB(&geoIPData, parsedLog.IPAddress, logger)
		} else {
			getIPDetailsFromURL(&geoIPData, parsedLog.IPAddress, logger)
		}
	}
	// Add metric.
	authVentsMetric.WithLabelValues(
		parsedLog.Type,
		hideValue(metricHideUser, parsedLog.Username),
		hideValue(metricHideIP, parsedLog.IPAddress),
		geoIPData.countyISOCode,
		geoIPData.countryName,
		geoIPData.cityName,
	).Inc()
}

func getMatches(line string, re *regexp.Regexp) map[string]string {
	matches := re.FindStringSubmatch(line)
	results := make(map[string]string)
	// Get the basic information out of the log.
	for i, name := range re.SubexpNames() {
		if i != 0 && len(matches) > i {
			results[name] = matches[i]
		}
	}
	return results
}

func hideValue(boolValue bool, value string) string {
	if boolValue {
		return ""
	}
	return value
}
