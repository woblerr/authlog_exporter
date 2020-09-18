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
			"GeoIP database is not set and not use",
		},
		{"dbGeodbTypePathEmpty",
			args{"db", "", "", ""},
			"Error geo.db flag is not set",
		},
		{"dbGeodbTypePathNotEmpty",
			args{"db", "test.file", "", ""},
			"Use GeoIp database file",
		},
		{"urlGeodbType",
			args{"url", "", "", "http://test"},
			"Use GeoIp database url",
		},
		{"badGeodbType",
			args{"test", "", "", ""},
			"Error geo.type flag value is incorect",
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
