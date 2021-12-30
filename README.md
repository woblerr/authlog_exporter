# authlog_exporter

[![Actions Status](https://github.com/woblerr/authlog_exporter/workflows/build/badge.svg)](https://github.com/woblerr/authlog_exporter/actions)
[![Coverage Status](https://coveralls.io/repos/github/woblerr/authlog_exporter/badge.svg?branch=master)](https://coveralls.io/github/woblerr/authlog_exporter?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/woblerr/authlog_exporter)](https://goreportcard.com/report/github.com/woblerr/authlog_exporter)

Prometheus exporter for collecting metrics from linux `auth.log` file.

## Collected metrics

The client provides a metric `authlog_events_total` which contains the number of auth events group by `event type`, `user` and `ip address`. Client also could analyze the location of IP addresses found in `auth.log` if geoIP database is specified.

### Metric description

**Example metrics:**

```
# HELP authlog_events_total The total number of auth events by user and IP addresses
# TYPE authlog_events_total counter
authlog_events_total{cityName="",countryName="",countyISOCode="",eventType="invalidUser",ipAddress="12.123.12.123",user="support"} 1
authlog_events_total{cityName="",countryName="",countyISOCode="",eventType="notAllowedUser",ipAddress="12.123.12.123",user="root"} 1
authlog_events_total{cityName="",countryName="",countyISOCode="",eventType="notAllowedUser",ipAddress="12.123.123.1",user="root"} 5
authlog_events_total{cityName="",countryName="",countyISOCode="",eventType="authAccepted",ipAddress="123.123.12.12",user="testuser"} 2
authlog_events_total{cityName="",countryName="",countyISOCode="",eventType="authFailed",ipAddress="123.123.12.12",user="root"} 1
authlog_events_total{cityName="",countryName="",countyISOCode="",eventType="authFailed",ipAddress="123.123.12.123",user="root"} 1
authlog_events_total{cityName="",countryName="",countyISOCode="",eventType="connectionClosed",ipAddress="123.123.12.12",user="testuser"} 1
```

If geoIP database is specified:

```
# HELP authlog_events_total The total number of auth events by user and IP addresses
# TYPE authlog_events_total counter
authlog_events_total{cityName="",countryName="United States",countyISOCode="US",eventType="invalidUser",ipAddress="12.123.12.123",user="support"} 1
authlog_events_total{cityName="",countryName="United States",countyISOCode="US",eventType="notAllowedUser",ipAddress="12.123.12.123",user="root"} 1
authlog_events_total{cityName="",countryName="United States",countyISOCode="US",eventType="notAllowedUser",ipAddress="12.123.123.1",user="root"} 5
authlog_events_total{cityName="Beijing",countryName="China",countyISOCode="CN",eventType="authAccepted",ipAddress="123.123.12.12",user="testuser"} 2
authlog_events_total{cityName="Beijing",countryName="China",countyISOCode="CN",eventType="authFailed",ipAddress="123.123.12.12",user="root"} 1
authlog_events_total{cityName="Beijing",countryName="China",countyISOCode="CN",eventType="authFailed",ipAddress="123.123.12.123",user="root"} 1
authlog_events_total{cityName="Beijing",countryName="China",countyISOCode="CN",eventType="connectionClosed",ipAddress="123.123.12.12",user="testuser"} 1
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

```bash
git clone https://github.com/woblerr/authlog_exporter.git
cd authlog_exporter
make build
./authlog_exporter <flags>
```

By default, metrics will be collecting from `/var/log/auth.log` and will be available at http://localhost:9991/metrics. This means that the user who runs `authlog_exporter` should have read permission to file `/var/log/auth.log`. You can changed logfile location, port and endpoint by using the`--auth.log`, `--prom.port` and `--prom.endpoint` flags.

For geoIP analyze you need to specify `--geo.type` flag:
* `db` - for local geoIP database file,
* `url` - for geoIP database API.

For local geoIP database usage you also need specify `--geo.db` flag (path to geoIP database file).

The flag `--prom.web-config` allows to specify the path to the configuration for TLS and/or basic authentication.<br>
The description of TLS configuration and basic authentication can be found at [exporter-toolkit/web](https://github.com/prometheus/exporter-toolkit/blob/v0.7.1/docs/web-configuration.md).

Available configuration flags:

```bash
./authlog_exporter --help
usage: authlog_exporter [<flags>]

Flags:
  -h, --help                Show context-sensitive help (also try --help-long and --help-man).
      --auth.log="/var/log/auth.log"  
                            Path to auth.log.
      --prom.endpoint="/metrics"  
                            Endpoint used for metrics.
      --prom.port="9991"    Port for prometheus metrics to listen on.
      --prom.web-config=""  [EXPERIMENTAL] Path to config yaml file that can enable TLS or authentication.
      --geo.db=""           Path to geoIP database file.
      --geo.lang="en"       Output language format.
      --geo.timeout=2       Timeout in seconds for waiting response from geoIP database API.
      --geo.type=""         Type of geoIP database: db, url.
      --geo.url="https://freegeoip.live/json/"  
                            URL for geoIP database API.
      --log.level=info      Only log messages with the given severity or above. One of: [debug, info, warn, error]
      --log.format=logfmt   Output format of log messages. One of: [logfmt, json]
```

### geoIP

#### Local geoIP database

To analyze IP addresses location found in the log from local geoIP database you need to free download: [GeoLite2-City](https://dev.maxmind.com/geoip/geoip2/geolite2/).

The library [geoip2-golang](https://github.com/oschwald/geoip2-golang) is used for reading the GeoLite2 database.

```bash
./authlog_exporter --geo.type db --geo.db /path/to/GeoLite2-City.mmdb
```

Уou can specify output language (default `en`):

```bash
./authlog_exporter --geo.type db --geo.db /path/to/GeoLite2-City.mmdb --geo.lang ru
```

Metric example:

```
authlog_events_total{cityName="Пекин",countryName="Китай",countyISOCode="CN",eventType="authAccepted",ipAddress="123.123.12.12",user="testuser"} 2
```

#### geoIP database API

To analyze IP addresses location using external API https://freegeoip.live:

```bash
./authlog_exporter --geo.type url
```

Be aware that API has a limit of **10K requests per hour**.

### Running tests

```bash
make test
```

For bulding and running on test log:

```bash
make run-test
```

### Running as systemd service

* Register `authlog_exporter` (already builded, if not - exec `make build` before) as a systemd service:

```bash
make prepare-service
```

Validate prepared file `authlog_exporter.service` and run:

```bash
sudo make install-service
```

* View service logs:

```bash
journalctl -u authlog_exporter.service
```

* Delete systemd service:

```bash
sudo make remove-service
```

---
Manual register systemd service:

```bash
cp authlog_exporter.service.template authlog_exporter.service
```

In file `authlog_exporter.service` replace ***{PATH_TO_FILE}*** to full path to `authlog_exporter`.

```bash
sudo cp authlog_exporter.service /etc/systemd/system/authlog_exporter.service
sudo systemctl daemon-reload
sudo systemctl enable authlog_exporter.service
sudo systemctl restart authlog_exporter.service
systemctl -l status authlog_exporter.service
```

---

### Running as docker container

Be aware that user who runs docker container should have read permission to file `/var/log/auth.log`. Otherwise, the container won't start.

* Build container:

```bash
make docker-build
```

or manual:

```bash
docker build  -f Dockerfile  -t authlog_exporter.
```

* Run container

```bash
make docker-run
```

or manual:

```bash
docker run -d --restart=always \
  --name authlog_exporter \
  -p 9991:9991 \
  -v /var/log/auth.log:/log/auth.log:ro \
  -u $(id -u):$(id -g) \
  authlog_exporter \
  --auth.log /log/auth.log
```
