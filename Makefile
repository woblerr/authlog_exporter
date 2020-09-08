APP_NAME = auth_exporter
SERVICE_CONF_DIR = /etc/systemd/system
HTTP_PORT = 9991
ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

all: run-test docker-run-test

.PHONY: test
test:
	@echo "Run tests for $(APP_NAME)"
	go get -v ./...
	go get golang.org/x/tools/cmd/cover
	go test -v ./... -covermode=count -coverprofile=coverage.out

.PHONY: build
build:
	@echo "Build $(APP_NAME)"
	@make test
	CGO_ENABLED=0 GOOS=linux go build -trimpath -o $(APP_NAME) $(APP_NAME).go
	./$(APP_NAME) -h

.PHONY: run-test
run-test:
	@echo "Run $(APP_NAME) for test log: ./test/auth.log"
	@make build
	./auth_exporter -auth.log ./test_log/auth.log &
	$(call http-test)
	pkill -f auth_exporter

.PHONY: prepare-service
prepare-service:
	@echo "Prepare config file $(APP_NAME).service for systemd"
	cp $(ROOT_DIR)/$(APP_NAME).service.template $(ROOT_DIR)/$(APP_NAME).service
	sed -i "s|{PATH_TO_FILE}|$(ROOT_DIR)|g" auth_exporter.service

.PHONY: install-service
install-service:
	@echo "Install $(APP_NAME) as systemd service"
	$(call service-install)

.PHONY: remove-service
remove-service:
	@echo "Delete $(APP_NAME) systemd service"
	$(call service-remove)

.PHONY: docker-build
docker-build:
	@echo "Build $(APP_NAME) docker container"
	docker build -t $(APP_NAME) .

.PHONY: docker-run
docker-run:
	@echo "Run $(APP_NAME) docker container"
	$(call run-container, /var/log/auth.log:/log/auth.log:ro)

.PHONY: docker-run-test
docker-run-test:
	@echo "Run $(APP_NAME) docker container for test log: ./test/auth.log"
	$(call run-container, $(PWD)/test_log/auth.log:/log/auth.log:ro)
	$(call http-test)
	@make docker-remove

.PHONY: docker-remove
docker-remove:
	@echo "Stop and delete $(APP_NAME) docker container"
	docker rm -f $(APP_NAME)

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
	rm $(SERVICE_CONF_DIR)/$(APP_NAME).service
	systemctl daemon-reload
	systemctl reset-failed
endef

define run-container
	docker run -d --restart=always \
		--name $(APP_NAME) -p $(HTTP_PORT):9991 \
		-v ${1} \
		-u $(shell id -u):$(shell id -g) \
		$(APP_NAME) \
		-auth.log /log/auth.log
endef

define http-test
	sleep 2
	curl -s "http://localhost:9991/metrics"| grep "^auth_exporter_auth_events"
endef