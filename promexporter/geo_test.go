package promexporter

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestSetGeodbPath(t *testing.T) {
	type args struct {
		geoType    string
		filePath   string
		outputLang string
		url        string
		timeout    int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"defaultGeodbType",
			args{"", "", "", "", 2},
			"[INFO] GeoIP database is not use",
		},
		{"dbGeodbTypePathEmpty",
			args{"db", "", "", "", 2},
			"[ERROR] Flag geo.db is not set",
		},
		{"dbGeodbTypePathNotEmpty",
			args{"db", "test.file", "", "", 2},
			"[INFO] Use GeoIp database file",
		},
		{"urlGeodbType",
			args{"url", "", "", "http://test", 2},
			"[INFO] Use GeoIp database url",
		},
		{"badGeodbType",
			args{"test", "", "", "", 2},
			"[ERROR] Flag geo.type is incorrect",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer func() {
				log.SetOutput(os.Stdout)
			}()
			SetGeodbPath(
				tt.args.geoType,
				tt.args.filePath,
				tt.args.outputLang,
				tt.args.url,
				tt.args.timeout)
			if got := buf.String(); !strings.Contains(got, tt.want) {
				t.Errorf("\nSetGeodbPath() log output:\n%s\nnot containt want:\n%s", got, tt.want)
			}
		})
	}
}

func TestGetMap(t *testing.T) {
	type args struct {
		data map[string]interface{}
		key  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"existKey",
			args{
				map[string]interface{}{"country_code": "US"},
				"country_code",
			},
			"US"},
		{"notExistKey",
			args{
				map[string]interface{}{"country_code": "US"},
				"city",
			},
			""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMap(tt.args.data, tt.args.key); got != tt.want {
				t.Errorf("\ngetMap() =\n%v,\nwant=\n%v", got, tt.want)
			}
		})
	}
}

func TestGetIPDetailsFromURL(t *testing.T) {
	srv := serverMock()
	defer srv.Close()
	type args struct {
		returnValues *geoInfo
		ipAddres     string
	}

	tests := []struct {
		name       string
		args       args
		want       *geoInfo
		testGeoURL string
	}{
		{"getIPValidResponse",
			args{
				&geoInfo{"", "", ""},
				"12.123.12.123",
			},
			&geoInfo{"US", "United States", ""},
			srv.URL + "/json/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			geoURL = tt.testGeoURL
			getIPDetailsFromURL(tt.args.returnValues, tt.args.ipAddres)
			if !reflect.DeepEqual(tt.args.returnValues, tt.want) {
				t.Errorf("\ngetIPDetailsFromURL() =\n%v,\nwant=\n%v", tt.args.returnValues, tt.want)
			}
		})
	}
}

func TestGetIPDetailsFromURLErrors(t *testing.T) {
	srv := serverMock()
	defer srv.Close()
	type args struct {
		returnValues *geoInfo
		ipAddres     string
	}
	reqArgs := args{
		&geoInfo{"", "", ""},
		"12.123.12.123",
	}
	tests := []struct {
		name       string
		args       args
		testGeoURL string
		testText   string
	}{
		{"getIPGetError",
			reqArgs,
			"http://test",
			"[ERROR] Error getting GeoIp URL",
		},
		{"getIPNoBody",
			reqArgs,
			srv.URL + "/nobody/",
			"[ERROR] Error getting body from GeoIp URL",
		},
		{"getIPParseBodyError",
			reqArgs,
			srv.URL + "/badbody/",
			"[ERROR] Error parsing json-encoded body from GeoIp URL",
		},
		{"getIPGetNoResponse",
			reqArgs,
			srv.URL + "/longresponse/",
			"[ERROR] Error getting GeoIp URL",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			geoURL = tt.testGeoURL
			out := &bytes.Buffer{}
			log.SetOutput(out)
			getIPDetailsFromURL(tt.args.returnValues, tt.args.ipAddres)
			if !strings.Contains(out.String(), tt.testText) {
				t.Errorf("\nVariable do not match:\n%s\nwant:\n%s", tt.testText, out.String())
			}
		})
	}
}

func TestGetIPDetailsFromLocalDB(t *testing.T) {
	type args struct {
		returnValues *geoInfo
		ipAddres     string
	}
	geoLang = "en"
	tests := []struct {
		name        string
		args        args
		want        *geoInfo
		testGeoFile string
	}{
		{"getIPDetailValid",
			args{
				&geoInfo{"", "", ""},
				"12.123.12.123",
			},
			&geoInfo{"US", "United States", ""},
			"../test_data/geolite2_test.mmdb",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			geodbPath = getFullPath(tt.testGeoFile)
			getIPDetailsFromLocalDB(tt.args.returnValues, tt.args.ipAddres)
			if !reflect.DeepEqual(tt.args.returnValues, tt.want) {
				t.Errorf("\ngetIPDetailsFromURL() =\n%v,\nwant=\n%v", tt.args.returnValues, tt.want)
			}
		})
	}
}

func TestGetIPDetailsFromLocalDBErrors(t *testing.T) {
	type args struct {
		returnValues *geoInfo
		ipAddres     string
	}
	tests := []struct {
		name        string
		args        args
		testGeoFile string
		testText    string
	}{
		{"getIPDetailErrorOpenDB",
			args{
				&geoInfo{"", "", ""},
				"12.123.12.123",
			},
			"../test_data/GeoLite2-City-Missing.mmdb",
			"[ERROR] Error opening GeoIp database file",
		},
		{"getIPDetailErrorParseIP",
			args{
				&geoInfo{"", "", ""},
				"12.123.12.",
			},
			"../test_data/geolite2_test.mmdb",
			"[ERROR] Error parsing ip address",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			geodbPath = getFullPath(tt.testGeoFile)
			out := &bytes.Buffer{}
			log.SetOutput(out)
			getIPDetailsFromLocalDB(tt.args.returnValues, tt.args.ipAddres)
			if !strings.Contains(out.String(), tt.testText) {
				t.Errorf("\nVariable do not match:\n%s\nwant:\n%s", tt.testText, out.String())
			}
		})
	}
}

func serverMock() *httptest.Server {
	handler := http.NewServeMux()
	testIP := "12.123.12.123"
	handler.HandleFunc("/json/"+testIP,
		func(rw http.ResponseWriter, req *http.Request) {
			rw.Header().Set("Content-Type", "application/json")
			rw.Write([]byte(`{"ip":"12.123.12.123","country_code":"US","country_name":"United States","region_code":"","region_name":"","city":"","zip_code":"","time_zone":"America/Chicago","latitude":37.751,"longitude":-97.822,"metro_code":0}`))
		})
	handler.HandleFunc("/nobody/"+testIP,
		func(rw http.ResponseWriter, req *http.Request) {
			//rw.WriteHeader(http.StatusInternalServerError)
			rw.Header().Set("Content-Length", "1")
		})
	handler.HandleFunc("/badbody/"+testIP,
		func(rw http.ResponseWriter, req *http.Request) {
			rw.Write([]byte(`{"ip":"12.123.12.123",`))
		})
	handler.HandleFunc("/longresponse/"+testIP,
		func(rw http.ResponseWriter, req *http.Request) {
			time.Sleep(3 * time.Second)
		})
	srv := httptest.NewServer(handler)
	return srv
}

func getFullPath(relativeFilePath string) string {
	absPath, err := filepath.Abs(relativeFilePath)
	if err != nil {
		panic(err)
	}
	return absPath
}
