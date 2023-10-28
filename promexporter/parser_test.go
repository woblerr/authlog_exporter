package promexporter

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/nxadm/tail"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

func TestGetMatches(t *testing.T) {
	type args struct {
		line string
		re   *regexp.Regexp
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{"authAccepted",
			args{
				"Aug 30 15:11:36 hostname sshd[11944]: Accepted publickey for testuser from 123.123.12.12 port 46680",
				authLineRegexps["authAccepted"],
			},
			map[string]string{
				"":          "publickey",
				"date":      "Aug 30",
				"host":      "hostname",
				"ident":     "sshd",
				"ipAddress": "123.123.12.12",
				"pid":       "11944",
				"time":      "15:11:36",
				"user":      "testuser",
			},
		},
		{"authFailed_1",
			args{
				"Aug 30 15:11:36 hostname sshd[11944]: Failed password for invalid user root from 123.123.12.12 port 46680",
				authLineRegexps["authFailed"],
			},
			map[string]string{
				"":          "invalid user ",
				"date":      "Aug 30",
				"host":      "hostname",
				"ident":     "sshd",
				"ipAddress": "123.123.12.12",
				"pid":       "11944",
				"time":      "15:11:36",
				"user":      "root",
			},
		},
		{"authFailed_2",
			args{
				"Aug 30 15:11:36 hostname sshd[11944]: Failed password for root from 123.123.12.12 port 46680",
				authLineRegexps["authFailed"],
			},
			map[string]string{
				"":          "",
				"date":      "Aug 30",
				"host":      "hostname",
				"ident":     "sshd",
				"ipAddress": "123.123.12.12",
				"pid":       "11944",
				"time":      "15:11:36",
				"user":      "root",
			},
		},
		{"invalidUser",
			args{
				"Aug 30 15:11:36 hostname sshd[11944]: Invalid user root from 123.123.12.12 port 46680",
				authLineRegexps["invalidUser"],
			},
			map[string]string{
				"":          "[11944]",
				"date":      "Aug 30",
				"host":      "hostname",
				"ident":     "sshd",
				"ipAddress": "123.123.12.12",
				"pid":       "11944",
				"time":      "15:11:36",
				"user":      "root",
			},
		},
		{"notAllowedUser",
			args{
				"Aug 30 15:11:36 hostname sshd[11944]: User root from 123.123.12.12 not allowed because not listed in AllowUsers",
				authLineRegexps["notAllowedUser"],
			},
			map[string]string{
				"":          "[11944]",
				"date":      "Aug 30",
				"host":      "hostname",
				"ident":     "sshd",
				"ipAddress": "123.123.12.12",
				"pid":       "11944",
				"time":      "15:11:36",
				"user":      "root",
			},
		},
		{"connectionClosed",
			args{
				"Aug 30 15:11:36 hostname sshd[11944]: Connection closed by authenticating user root 123.123.12.12 port 46680",
				authLineRegexps["connectionClosed"],
			},
			map[string]string{
				"":          "[11944]",
				"date":      "Aug 30",
				"host":      "hostname",
				"ident":     "sshd",
				"ipAddress": "123.123.12.12",
				"pid":       "11944",
				"time":      "15:11:36",
				"user":      "root",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMatches(tt.args.line, tt.args.re); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("\ngetMatches():\n%v,\nwant:\n%v", got, tt.want)
			}
		})
	}
}

func TestHideValue(t *testing.T) {
	type args struct {
		value     string
		boolValue bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"hideTrue",
			args{
				value:     "test",
				boolValue: true,
			},
			"",
		},
		{
			"hideFalse",
			args{
				value:     "test",
				boolValue: false,
			},
			"test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hideValue(tt.args.boolValue, tt.args.value); got != tt.want {
				t.Errorf("hideValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseLine(t *testing.T) {
	type args struct {
		lineTest    *tail.Line
		geodbIsTest bool
	}
	tests := []struct {
		name     string
		args     args
		testText string
	}{
		{
			"ParseLineExist",
			args{
				valToPtr(tail.Line{
					Text:     "Aug 30 12:02:00 hostname sshd[17917]: User root from 12.123.12.123 not allowed because not listed in AllowUsers",
					Num:      3,
					SeekInfo: tail.SeekInfo{Offset: 305, Whence: 0},
					Time:     time.Unix(1693560000, 0),
					Err:      nil,
				}),
				false,
			},
			`# HELP authlog_events_total The total number of auth events.
# TYPE authlog_events_total counter
authlog_events_total{cityName="",countryName="",countyISOCode="",eventType="notAllowedUser",ipAddress="12.123.12.123",user="root"} 1
`,
		},
		{
			"ParseLineNotExist",
			args{
				valToPtr(tail.Line{
					Text:     "Some text",
					Num:      3,
					SeekInfo: tail.SeekInfo{Offset: 305, Whence: 0},
					Time:     time.Unix(1693560000, 0),
					Err:      nil,
				}),
				false,
			},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authVentsMetric.Reset()
			geodbIs = tt.args.geodbIsTest
			parseLine(tt.args.lineTest, logger)
			reg := prometheus.NewRegistry()
			reg.MustRegister(
				authVentsMetric,
			)
			metricFamily, err := reg.Gather()
			if err != nil {
				fmt.Println(err)
			}
			out := &bytes.Buffer{}
			for _, mf := range metricFamily {
				if _, err := expfmt.MetricFamilyToText(out, mf); err != nil {
					panic(err)
				}
			}
			if tt.testText != out.String() {
				t.Errorf("\nVariables do not match, metrics:\n%s\nwant:\n%s", tt.testText, out.String())
			}
		})
	}
}
