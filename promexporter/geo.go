package promexporter

import (
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/oschwald/geoip2-golang"
)

var (
	geodbType  string
	geodbPath  string
	geoLang    string
	geoURL     string
	geoTimeout int
	geodbIs    = false
)

type geoInfo struct {
	countyISOCode string
	countryName   string
	cityName      string
}

// SetGeodbPath sets geoIP database parameters from command line argument
// geo.type, geo.db and geo.lang, geo.url or geo.timeout.
func SetGeodbPath(geoType, filePath, outputLang, url string, timeout int, logger *slog.Logger) {
	geodbType = geoType
	geodbPath = filePath
	geoLang = outputLang
	geoURL = url
	geoTimeout = timeout
	checkGeoDBFlags(logger)
}

func checkGeoDBFlags(logger *slog.Logger) {
	switch geodbType {
	case "":
		logger.Info("GeoIP database is not use")
	case "db":
		if geodbPath == "" {
			logger.Error("Flag geo.db is not set", "file", geodbPath)
		} else {
			geodbIs = true
			logger.Info("Use GeoIp database file", "file", geodbPath)
		}
	case "url":
		geodbIs = true
		logger.Info("Use GeoIp database url", "url", geoURL)
	default:
		logger.Error("Flag geo.type is incorrect", "type", geodbType)
	}
}

func getIPDetailsFromLocalDB(returnValues *geoInfo, ipAddress string, logger *slog.Logger) {
	geodb, err := geoip2.Open(geodbPath)
	if err != nil {
		logger.Error("Error opening GeoIp database file", "err", err)
		return
	}
	defer geodb.Close()
	logger.Debug("Parse IP", "ip", ipAddress)
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		logger.Error("Error parsing IP address", "ip", ipAddress)
		return
	}
	record, err := geodb.City(ip)
	if err != nil {
		logger.Error("Error getting location details", "err", err)
		return
	}
	returnValues.countyISOCode = record.Country.IsoCode
	returnValues.countryName = record.Country.Names[geoLang]
	returnValues.cityName = record.City.Names[geoLang]
}

func getIPDetailsFromURL(returnValues *geoInfo, ipAddress string, logger *slog.Logger) {
	// Timeout for get and read response body.
	client := http.Client{
		Timeout: time.Duration(geoTimeout) * time.Second,
	}
	logger.Debug("Get IP details from url", "url", geoURL+ipAddress)
	response, err := client.Get(geoURL + ipAddress)
	if err != nil {
		logger.Error("Error getting GeoIp URL", "err", err)
		return
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	logger.Debug("Response body", "body", string(body))
	if err != nil {
		logger.Error("Error getting body from GeoIp URL", "err", err)
		return
	}
	var parseData map[string]interface{}
	err = json.Unmarshal(body, &parseData)
	if err != nil {
		logger.Error("Error parsing json-encoded body from GeoIp URL", "err", err)
		return
	}
	returnValues.countyISOCode = getMap(parseData, "country_code", logger)
	returnValues.countryName = getMap(parseData, "country_name", logger)
	returnValues.cityName = getMap(parseData, "city", logger)
}

func getMap(data map[string]interface{}, key string, logger *slog.Logger) string {
	str, ok := data[key].(string)
	if !ok {
		logger.Error("Is not a string", "key", key, "value", str)
	}
	return str
}
