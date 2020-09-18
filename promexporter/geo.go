package promexporter

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"github.com/oschwald/geoip2-golang"
)

var (
	geodbType string
	geodbPath string
	geoLang   string
	geoURL    string
	geodbIs   = false
)

type geoInfo struct {
	countyISOCode string
	countryName   string
	cityName      string
}

// SetGeodbPath sets geoIP database parameters from command line argument geo.type, geo.db and geo.lang or geo.url
func SetGeodbPath(geoType, filePath, outputLang, url string) {
	geodbType = geoType
	geodbPath = filePath
	geoLang = outputLang
	geoURL = url
	checkGeoDBFlags()
}

func checkGeoDBFlags() {
	notSetLogMsg := "GeoIP database is not set and not use"
	switch geodbType {
	case "":
		log.Println(notSetLogMsg)
	case "db":
		if geodbPath == "" {
			log.Println("Error geo.db flag is not set", geodbPath)
			log.Println(notSetLogMsg)
		} else {
			geodbIs = true
			log.Println("Use GeoIp database file", geodbPath)
		}
	case "url":
		geodbIs = true
		log.Println("Use GeoIp database url", geoURL)
	default:
		log.Println("Error geo.type flag value is incorrect", geodbType)
		log.Println(notSetLogMsg)
	}
}

func getIPDetailsFromLocalDB(returnValues *geoInfo, ipAddres string) {
	geodb, err := geoip2.Open(geodbPath)
	if err != nil {
		log.Println("Error opening GeoIp database file", err)
		return
	}
	defer geodb.Close()
	ip := net.ParseIP(ipAddres)
	if ip == nil {
		log.Println("Error parsing ip address", ipAddres)
		return
	}
	record, err := geodb.City(ip)
	if err != nil {
		log.Println("Error getting location details", err)
		return
	}
	returnValues.countyISOCode = record.Country.IsoCode
	returnValues.countryName = record.Country.Names[geoLang]
	returnValues.cityName = record.City.Names[geoLang]
}

func getIPDetailsFromURL(returnValues *geoInfo, ipAddres string) {
	response, err := http.Get(geoURL + ipAddres)
	if err != nil {
		log.Println("Error getting GeoIp URL", err)
		return
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("Error getting body from GeoIp URL", err)
		return
	}
	var parseData map[string]interface{}
	err = json.Unmarshal([]byte(body), &parseData)
	if err != nil {
		log.Println("Error parsing json-encoded body from GeoIp URL", err)
		return
	}
	returnValues.countyISOCode = getMap(parseData, "country_code")
	returnValues.countryName = getMap(parseData, "country_name")
	returnValues.cityName = getMap(parseData, "city")
}

func getMap(data map[string]interface{}, key string) string {
	str, ok := data[key].(string)
	if !ok {
		log.Printf("Error for key %s value %s is not a string", key, str)
	}
	return str
}
