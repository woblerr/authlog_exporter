package promexporter

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/netip"
	"time"

	geoip2 "github.com/oschwald/geoip2-golang/v2"
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
		logger.Info("GeoIP database is not used")
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
	if ipAddress == "" {
		return
	}
	geodb, err := geoip2.Open(geodbPath)
	if err != nil {
		logger.Error("Error opening GeoIp database file", "err", err)
		return
	}
	defer geodb.Close()
	logger.Debug("Parse IP", "ip", ipAddress)
	ip, err := netip.ParseAddr(ipAddress)
	if err != nil {
		logger.Error("Error parsing IP address", "ip", ipAddress, "err", err)
		return
	}
	record, err := geodb.City(ip)
	if err != nil {
		logger.Error("Error getting location details", "err", err)
		return
	}
	returnValues.countyISOCode = record.Country.ISOCode
	returnValues.countryName = getNameByLang(record.Country.Names, geoLang)
	returnValues.cityName = getNameByLang(record.City.Names, geoLang)
}

// hugeParam: intentional pass-by-value for immutability guarantee
//
//nolint:gocritic
func getNameByLang(names geoip2.Names, lang string) string {
	switch lang {
	case "de":
		return names.German
	case "en":
		return names.English
	case "es":
		return names.Spanish
	case "fr":
		return names.French
	case "ja":
		return names.Japanese
	case "pt-BR":
		return names.BrazilianPortuguese
	case "ru":
		return names.Russian
	case "zh-CN":
		return names.SimplifiedChinese
	default:
		return names.English
	}
}

func getIPDetailsFromURL(returnValues *geoInfo, ipAddress string, logger *slog.Logger) {
	if ipAddress == "" {
		return
	}
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
	if err != nil {
		logger.Error("Error getting body from GeoIp URL", "err", err)
		return
	}
	logger.Debug("Response body", "body", string(body))
	if response.StatusCode != http.StatusOK {
		logger.Error("Unexpected status code from GeoIp URL", "status", response.StatusCode)
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
	value, exists := data[key]
	if !exists {
		logger.Debug("Key not found", "key", key)
		return ""
	}
	str, ok := value.(string)
	if !ok {
		logger.Error("Is not a string", "key", key, "value", value)
	}
	return str
}
