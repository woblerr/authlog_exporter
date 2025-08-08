package promexporter

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
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
			"GeoIP database is not use",
		},
		{"dbGeodbTypePathEmpty",
			args{"db", "", "", "", 2},
			"Flag geo.db is not set",
		},
		{"dbGeodbTypePathNotEmpty",
			args{"db", "test.file", "", "", 2},
			"Use GeoIp database file",
		},
		{"urlGeodbType",
			args{"url", "", "", "http://test", 2},
			"Use GeoIp database url",
		},
		{"badGeodbType",
			args{"test", "", "", "", 2},
			"Flag geo.type is incorrect",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			lc := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug}))
			SetGeodbPath(
				tt.args.geoType,
				tt.args.filePath,
				tt.args.outputLang,
				tt.args.url,
				tt.args.timeout,
				lc)
			if !strings.Contains(out.String(), tt.want) {
				t.Errorf("\nVariable do not match:\n%s\nwant:\n%s", tt.want, out.String())
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
		{"existKeyButNotString",
			args{
				map[string]interface{}{"count": 123},
				"count",
			},
			""},
		{"existKeyButNilValue",
			args{
				map[string]interface{}{"value": nil},
				"value",
			},
			""},
		{"existKeyButBoolValue",
			args{
				map[string]interface{}{"flag": true},
				"flag",
			},
			""},
		{"existKeyButFloatValue",
			args{
				map[string]interface{}{"rate": 3.14},
				"rate",
			},
			""},
		{"existKeyButSliceValue",
			args{
				map[string]interface{}{"items": []string{"a", "b"}},
				"items",
			},
			""},
		{"existKeyButMapValue",
			args{
				map[string]interface{}{"nested": map[string]string{"key": "value"}},
				"nested",
			},
			""},
		{"emptyStringValue",
			args{
				map[string]interface{}{"empty": ""},
				"empty",
			},
			""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMap(tt.args.data, tt.args.key, logger); got != tt.want {
				t.Errorf("\ngetMap() =\n%v,\nwant=\n%v", got, tt.want)
			}
		})
	}
}

func TestGetMapErrorLogging(t *testing.T) {
	defaultErrorText := "Is not a string"
	type args struct {
		data      map[string]interface{}
		key       string
		wantError bool
		wantText  string
	}
	tests := []struct {
		name string
		args args
	}{
		{"existKeyButNotStringLogsError",
			args{
				map[string]interface{}{"count": 123},
				"count",
				true,
				defaultErrorText,
			}},
		{"existKeyButNilValueLogsError",
			args{
				map[string]interface{}{"value": nil},
				"value",
				true,
				defaultErrorText,
			}},
		{"validStringNoError",
			args{
				map[string]interface{}{"country": "US"},
				"country",
				false,
				"",
			}},
		{"nonExistentKeyNoError",
			args{
				map[string]interface{}{"country": "US"},
				"city",
				false,
				"",
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			lc := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug}))
			result := getMap(tt.args.data, tt.args.key, lc)
			if tt.args.wantError {
				if !strings.Contains(out.String(), tt.args.wantText) {
					t.Errorf("\nError message not found:\ngot: %s\nwant: %s", out.String(), tt.args.wantText)
				}
				if result != "" {
					t.Errorf("\nEmpty string when error occurs, got: %v", result)
				}
			} else if strings.Contains(out.String(), defaultErrorText) {
				t.Errorf("\nUnwanted error logged: %s", out.String())
			}
		})
	}
}

func TestGetMapDebugLogging(t *testing.T) {
	defaultDebugText := "Key not found"
	type args struct {
		data     map[string]interface{}
		key      string
		wantText string
	}
	tests := []struct {
		name string
		args args
	}{
		{"nonExistentKeyLogsDebug",
			args{
				map[string]interface{}{"country": "US"},
				"city",
				defaultDebugText,
			}},
		{"anotherNonExistentKey",
			args{
				map[string]interface{}{"ip": "8.8.8.8", "country_code": "US"},
				"region_name",
				defaultDebugText,
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			lc := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug}))
			result := getMap(tt.args.data, tt.args.key, lc)
			if !strings.Contains(out.String(), tt.args.wantText) {
				t.Errorf("\nMessage not found:\ngot: %s\nwant: %s", out.String(), tt.args.wantText)
			}
			if result != "" {
				t.Errorf("\nExpected empty string for missing key, got: %v", result)
			}
		})
	}
}

func TestGetIPDetailsFromURL(t *testing.T) {
	srv := serverMock()
	defer srv.Close()
	type args struct {
		returnValues *geoInfo
		ipAddress    string
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
			getIPDetailsFromURL(tt.args.returnValues, tt.args.ipAddress, logger)
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
		ipAddress    string
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
			"Error getting GeoIp URL",
		},
		{"getIPNoBody",
			reqArgs,
			srv.URL + "/nobody/",
			"Error getting body from GeoIp URL",
		},
		{"getIPParseBodyError",
			reqArgs,
			srv.URL + "/badbody/",
			"Error parsing json-encoded body from GeoIp URL",
		},
		{"getIPGetNoResponse",
			reqArgs,
			srv.URL + "/longresponse/",
			"Error getting GeoIp URL",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			geoURL = tt.testGeoURL
			out := &bytes.Buffer{}
			lc := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug}))
			getIPDetailsFromURL(tt.args.returnValues, tt.args.ipAddress, lc)
			if !strings.Contains(out.String(), tt.testText) {
				t.Errorf("\nVariable do not match:\n%s\nwant:\n%s", tt.testText, out.String())
			}
		})
	}
}

func TestGetIPDetailsFromLocalDB(t *testing.T) {
	type args struct {
		returnValues *geoInfo
		ipAddress    string
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
			getIPDetailsFromLocalDB(tt.args.returnValues, tt.args.ipAddress, logger)
			if !reflect.DeepEqual(tt.args.returnValues, tt.want) {
				t.Errorf("\ngetIPDetailsFromURL() =\n%v,\nwant=\n%v", tt.args.returnValues, tt.want)
			}
		})
	}
}

func TestGetIPDetailsFromLocalDBErrors(t *testing.T) {
	type args struct {
		returnValues *geoInfo
		ipAddress    string
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
			"Error opening GeoIp database file",
		},
		{"getIPDetailErrorParseIP",
			args{
				&geoInfo{"", "", ""},
				"12.123.12.",
			},
			"../test_data/geolite2_test.mmdb",
			"Error parsing IP address",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			geodbPath = getFullPath(tt.testGeoFile)
			out := &bytes.Buffer{}
			lc := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug}))
			getIPDetailsFromLocalDB(tt.args.returnValues, tt.args.ipAddress, lc)
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
