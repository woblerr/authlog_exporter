# prom_authlog_exporter

[![Build Status](https://travis-ci.com/woblerr/prom_authlog_exporter.svg?branch=master)](https://travis-ci.com/woblerr/prom_authlog_exporter)

Prometheus exporter for collecting metrics from linux `auth.log` file.

## Collected metrics

The client provides a metric `auth_exporter_auth_events` which contains the number of auth events group by `event type`, `user` and `ip address`.

### Metric description

**Example metrics:**

```
# HELP auth_exporter_auth_events The total number of auth events by user and IP addresses
# TYPE auth_exporter_auth_events counter
auth_exporter_auth_events{eventType="authAccepted",ipAddress="123.123.12.12",user="testuser"} 2
auth_exporter_auth_events{eventType="authFailed",ipAddress="123.123.12.12",user="root"} 1
auth_exporter_auth_events{eventType="authFailed",ipAddress="123.123.12.123",user="root"} 1
auth_exporter_auth_events{eventType="connectionClosed",ipAddress="123.123.12.12",user="testuser"} 1
auth_exporter_auth_events{eventType="invalidUser",ipAddress="12.123.12.123",user="support"} 1
auth_exporter_auth_events{eventType="invalidUser",ipAddress="123.123.123.123",user="postgres"} 1
auth_exporter_auth_events{eventType="notAllowedUser",ipAddress="12.123.123.1",user="root"} 5
auth_exporter_auth_events{eventType="notAllowedUser",ipAddress="123.123.123.123",user="root"} 1
```

**Prefix regexp:**

```
^(?P<date>[A-Z][a-z]{2}\\s+\\d{1,2}) (?P<time>(\\d{2}:?){3}) (?P<host>[a-zA-Z0-9_\\-\\.]+) (?P<ident>[a-zA-Z0-9_\\-]+)(\\[(?P<pid>\\d+)\\])?: 
```

**Collecting events:**
|Event type|Regexp for search event|
|---|---|
|authAccepted|`Accepted (password\|publickey) for (?P<user>.*) from (?P<ipAddress>.*) port`|
|authFailed|`Failed (password\|publickey) for (invalid user )?(?P<user>.*) from (?P<ipAddress>.*) port`|
|invalidUser|`Invalid user (?P<user>.*) from (?P<ipAddress>.*) port`|
|notAllowedUser|`User (?P<user>.*) from (?P<ipAddress>.*) not allowed because`|
|connectionClosed|`Connection closed by authenticating user (?P<user>.*) (?P<ipAddress>.*) port`|

## Getting Started

### Building and running

Requirement:

* [Go compiler](https://golang.org/dl/)

```bash
go get github.com/woblerr/prom_authlog_exporter
cd ${GOPATH-$HOME/go}/src/github.com/woblerr/prom_authlog_exporter
make build
./auth_exporter <flags>
```

By default, metrics will be collecting from `/var/log/auth.log` and will be available at http://localhost:9090/metrics. This means that the user who runs `auth_exporter` should have read permission to file `/var/log/auth.log`. You can changed logfile location, port and endpoint by using the`-auth.log`, `-port` and `-endpoint` flags.

Available configuration flags:

```bash
./auth_exporter -h

Usage of ./auth_exporter:
  -auth.log string
        Path to auth.log (default "/var/log/auth.log")
  -endpoint string
        Endpoint used for metrics (default "/metrics")
  -port string
        Port for prometheus metrics to listen on (default "9991")
```

### Running tests

```bash
make test
```

For bulding and running on test log:

```bash
make run-test
```

### Running as systemd service

* Register `auth_exporter` (already builded, if not - exec `make build` before) as a systemd service:

```bash
cd ${GOPATH-$HOME/go}/src/github.com/woblerr/prom_authlog_exporter
make prepare-service
```

Validate prepared file `auth_exporter.service` and run:

```bash
sudo make install-service
```

* View service logs:

```bash
journalctl -u auth_exporter.service
```

* Delete systemd service:

```bash
cd ${GOPATH-$HOME/go}/src/github.com/woblerr/prom_authlog_exporter
sudo make remove-service
```

---
Manual register systemd service:

```bash
cd ${GOPATH-$HOME/go}/src/github.com/woblerr/prom_authlog_exporter
cp auth_exporter.service.template auth_exporter.service
```

In file `auth_exporter.service` replace ***{PATH_TO_FILE}*** to full path to `auth_exporter`.

```bash
sudo cp auth_exporter.service /etc/systemd/system/auth_exporter.service
sudo systemctl daemon-reload
sudo systemctl enable auth_exporter.service
sudo systemctl restart auth_exporter.service
systemctl -l status auth_exporter.service
```

---
