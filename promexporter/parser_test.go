package promexporter

import (
	"reflect"
	"regexp"
	"testing"
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
