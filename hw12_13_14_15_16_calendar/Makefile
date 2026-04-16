BIN := "./bin/calendar"
BIN_SCHEDULER := "./bin/calendar_scheduler"
BIN_STORER := "./bin/calendar_storer"
DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

generate:
	oapi-codegen -generate "types,std-http,spec" -package generated \
		-o internal/server/http/generated/api.gen.go \
		api/openapi.yaml

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar
	go build -v -o $(BIN_SCHEDULER) -ldflags "$(LDFLAGS)" ./cmd/calendar_scheduler
	go build -v -o $(BIN_STORER) -ldflags "$(LDFLAGS)" ./cmd/calendar_storer

run: build
	$(BIN) --config=./configs/config.yaml

run-scheduler: build
	$(BIN_SCHEDULER) --config=./configs/scheduler_config.yaml

run-storer: build
	$(BIN_STORER) --config=./configs/storer_config.yaml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

up:
	docker-compose -f deployments/docker-compose.yaml up -d --build

down:
	docker-compose -f deployments/docker-compose.yaml down

restart:
	docker-compose -f deployments/docker-compose.yaml restart

version: build
	$(BIN) version

test:
	go test -race ./internal/...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin latest

deps:
	go mod download

lint: install-lint-deps deps
	golangci-lint run ./...

.PHONY: generate build run run-scheduler run-storer build-img run-img version test lint deps up down logs ps restart
