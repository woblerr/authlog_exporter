package promexporter

import (
	"encoding/json"
	"io/ioutil"
	"log"
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
func SetGeodbPath(geoType, filePath, outputLang, url string, timeout int) {
	geodbType = geoType
	geodbPath = filePath
	geoLang = outputLang
	geoURL = url
	geoTimeout = timeout
	checkGeoDBFlags()
}

func checkGeoDBFlags() {
	switch geodbType {
	case "":
		log.Println("[INFO] GeoIP database is not use")
	case "db":
		if geodbPath == "" {
			log.Println("[ERROR] Flag geo.db is not set", geodbPath)
			log.Println("[ERROR] GeoIP database is not set")
		} else {
			geodbIs = true
			log.Println("[INFO] Use GeoIp database file", geodbPath)
		}
	case "url":
		geodbIs = true
		log.Println("[INFO] Use GeoIp database url", geoURL)
	default:
		log.Println("[ERROR] Flag geo.type is incorrect", geodbType)
		log.Println("[ERROR] GeoIP database is not set")
	}
}

func getIPDetailsFromLocalDB(returnValues *geoInfo, ipAddres string) {
	geodb, err := geoip2.Open(geodbPath)
	if err != nil {
		log.Println("[ERROR] Error opening GeoIp database file", err)
		return
	}
	defer geodb.Close()
	ip := net.ParseIP(ipAddres)
	if ip == nil {
		log.Println("[ERROR] Error parsing ip address", ipAddres)
		return
	}
	record, err := geodb.City(ip)
	if err != nil {
		log.Println("[ERROR] Error getting location details", err)
		return
	}
	returnValues.countyISOCode = record.Country.IsoCode
	returnValues.countryName = record.Country.Names[geoLang]
	returnValues.cityName = record.City.Names[geoLang]
}

func getIPDetailsFromURL(returnValues *geoInfo, ipAddres string) {
	// Timeout for get and read response body.
	client := http.Client{
		Timeout: time.Duration(geoTimeout) * time.Second,
	}
	response, err := client.Get(geoURL + ipAddres)
	if err != nil {
		log.Println("[ERROR] Error getting GeoIp URL", err)
		return
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("[ERROR] Error getting body from GeoIp URL", err)
		return
	}
	var parseData map[string]interface{}
	err = json.Unmarshal(body, &parseData)
	if err != nil {
		log.Println("[ERROR] Error parsing json-encoded body from GeoIp URL", err)
		return
	}
	returnValues.countyISOCode = getMap(parseData, "country_code")
	returnValues.countryName = getMap(parseData, "country_name")
	returnValues.cityName = getMap(parseData, "city")
}

func getMap(data map[string]interface{}, key string) string {
	str, ok := data[key].(string)
	if !ok {
		log.Printf("[ERROR] Error for key %s value %s is not a string", key, str)
	}
	return str
}
