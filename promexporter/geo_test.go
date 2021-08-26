package promexporter

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

func TestSetGeodbPath(t *testing.T) {
	type args struct {
		geoType    string
		filePath   string
		outputLang string
		url        string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"defaultGeodbType",
			args{"", "", "", ""},
			"[INFO] GeoIP database is not use",
		},
		{"dbGeodbTypePathEmpty",
			args{"db", "", "", ""},
			"[ERROR] Flag geo.db is not set",
		},
		{"dbGeodbTypePathNotEmpty",
			args{"db", "test.file", "", ""},
			"[INFO] Use GeoIp database file",
		},
		{"urlGeodbType",
			args{"url", "", "", "http://test"},
			"[INFO] Use GeoIp database url",
		},
		{"badGeodbType",
			args{"test", "", "", ""},
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
			SetGeodbPath(tt.args.geoType, tt.args.filePath, tt.args.outputLang, tt.args.url)
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
