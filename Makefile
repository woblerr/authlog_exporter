APP_NAME = auth_exporter
SERVICE_CONF_DIR = /etc/systemd/system/
ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

all: run-test

.PHONY: test, build, run-test, prepare-service, install-service, remove-service

test:
	@echo "Run tests for $(APP_NAME)"
	go get -v ./...
	go test -v ./...

build:
	@echo "Building $(APP_NAME)"
	@make test
	GOOS=linux go build -o $(APP_NAME) $(APP_NAME).go
	./$(APP_NAME) -h

run-test:
	@echo "Run $(APP_NAME) for test log: ./test/auth.log"
	@make build
	$(call run-test-log)

prepare-service:
	@echo "Prepare confi file $(APP_NAME).service for systemd"
	cp $(ROOT_DIR)/$(APP_NAME).service.template $(ROOT_DIR)/$(APP_NAME).service
	sed -i "s|{PATH_TO_FILE}|$(ROOT_DIR)|g" auth_exporter.service

install-service:
	@echo "Installing $(APP_NAME) as systemd service"
	$(call service-install)

remove-service:
	@echo "Deleting $(APP_NAME) systemd service"
	$(call service-remove)

define run-test-log
	./auth_exporter -auth.log ./test_log/auth.log&
	sleep 2
	curl -s "http://localhost:9991/metrics"| grep "^auth_exporter_auth_events"
	pkill -f auth_exporter
endef

define service-install
	cp $(ROOT_DIR)/$(APP_NAME).service $(SERVICE_CONF_DIR)/$(APP_NAME).service
	systemctl daemon-reload
	systemctl enable $(APP_NAME)
	systemctl restart $(APP_NAME)
	systemctl status $(APP_NAME)
endef

define service-remove
	systemctl stop $(APP_NAME)
	systemctl disable $(APP_NAME)
	rm $(SERVICE_CONF_DIR)$(APP_NAME).service
	systemctl daemon-reload
	systemctl reset-failed
endef