APP_NAME = authlog_exporter
BRANCH_FULL=$(shell git rev-parse --abbrev-ref HEAD)
BRANCH=$(subst /,-,$(BRANCH_FULL))
GIT_REV=$(shell git describe --abbrev=7 --always)
SERVICE_CONF_DIR = /etc/systemd/system
HTTP_PORT = 9991
ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

all: run-test docker-run-test

.PHONY: test
test:
	@echo "Run tests for $(APP_NAME)"
	go test -mod=vendor -timeout=60s -count 1  ./...

.PHONY: build
build:
	@echo "Build $(APP_NAME)"
	@make test
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -trimpath -ldflags "-X main.version=$(BRANCH)-$(GIT_REV)" -o $(APP_NAME) $(APP_NAME).go

.PHONY: build-darwin
build-darwin:
	@echo "Build $(APP_NAME)"
	@make test
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -mod=vendor -trimpath -ldflags "-X main.version=$(BRANCH)-$(GIT_REV)" -o $(APP_NAME) $(APP_NAME).go

.PHONY: dist
dist:
	- @mkdir -p dist
	docker build -f Dockerfile.artifacts --progress=plain -t $(APP_NAME)_dist .
	- @docker rm -f $(APP_NAME)_dist 2>/dev/null || exit 0
	docker run -d --name=$(APP_NAME)_dist $(APP_NAME)_dist
	docker cp $(APP_NAME)_dist:/artifacts dist/
	docker rm -f $(APP_NAME)_dist

.PHONY: run-test
run-test:
	@echo "Run $(APP_NAME) for test log: ./test/auth.log"
	@make build
	./$(APP_NAME) --auth.log ./test_data/auth.log &
	$(call http-test)
	pkill -f $(APP_NAME)

.PHONY: run-test-darwin
run-test-darwin:
	@echo "Run $(APP_NAME) for test log: ./test/auth.log"
	@make build-darwin
	./$(APP_NAME) --auth.log ./test_data/auth.log &
	$(call http-test)
	pkill -f $(APP_NAME)

.PHONY: prepare-service
prepare-service:
	@echo "Prepare config file $(APP_NAME).service for systemd"
	cp $(ROOT_DIR)/$(APP_NAME).service.template $(ROOT_DIR)/$(APP_NAME).service
	sed -i.bak "s|/usr/bin|$(ROOT_DIR)|g" $(APP_NAME).service
	rm $(APP_NAME).service.bak

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
	@echo "Version $(BRANCH)-$(GIT_REV)"
	docker build --pull -f Dockerfile --build-arg REPO_BUILD_TAG=$(BRANCH)-$(GIT_REV) -t $(APP_NAME) .

.PHONY: docker-run
docker-run:
	@echo "Run $(APP_NAME) docker container"
	$(call run-container, /var/log/auth.log:/log/auth.log:ro)

.PHONY: docker-run-test
docker-run-test:
	@echo "Run $(APP_NAME) docker container for test log: ./test/auth.log"
	$(call run-container, $(PWD)/test_data/auth.log:/log/auth.log:ro)
	$(call http-test)
	@make docker-remove

.PHONY: docker-remove
docker-remove:
	@echo "Stop and delete $(APP_NAME) docker container"
	docker rm -f $(APP_NAME)

.PHONY: docker-build-test-db
docker-build-test-db:
	@echo "Build custom Maxmind database for tests"
	docker build -f test_data/Dockerfile.testdb --build-arg SCRIPT_PATH="test_data/create_test_database.pl" --progress=plain -t $(APP_NAME)_build_test_db .
	- @docker rm -f $(APP_NAME)_build_test_db 2>/dev/null || exit 0
	docker run -d --name=$(APP_NAME)_build_test_db $(APP_NAME)_build_test_db
	docker cp $(APP_NAME)_build_test_db:/db/geolite2_test.mmdb test_data/
	docker rm -f $(APP_NAME)_build_test_db

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
		--auth.log /log/auth.log
endef

define http-test
	sleep 2
	curl -s "http://localhost:9991/metrics"| grep "^authlog_events_total"
endef