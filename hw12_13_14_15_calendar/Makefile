BIN := "./bin/calendar"
BIN_SCHEDULER := "./bin/calendar_scheduler"
BIN_SENDER := "./bin/calendar_sender"
DOCKER_IMG="calendar:develop"
DOCKER_SCHEDULER_IMG="scheduler:develop"
DOCKER_SENDER_IMG="sender:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

DB_STRING := "user=cuser password=cpassword dbname=calendar host=localhost port=5432"
MIGRATIONS_DIR := migrations

RABBIT_NAME := "calendar_db"
POSTGRES_NAME := "calendar_mq"

build_calendar:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar

build_scheduler:
	go build -v -o $(BIN_SCHEDULER) -ldflags "$(LDFLAGS)" ./cmd/calendar_scheduler

build_sender:
	go build -v -o $(BIN_SENDER) -ldflags "$(LDFLAGS)" ./cmd/calendar_sender

build: build_calendar build_scheduler build_sender

run: build_calendar
	$(BIN) -config ./configs/config.yaml

run_scheduler: build_scheduler
	$(BIN_SCHEDULER) -config ./configs/config_scheduler.yaml

run_sender: build_sender
	$(BIN_SENDER) -config ./configs/config_sender.yaml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/package/calendar/Dockerfile .
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_SCHEDULER_IMG) \
		-f build/package/scheduler/Dockerfile .
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_SENDER_IMG) \
		-f build/package/sender/Dockerfile .

run-img: build-img
	docker run -d $(DOCKER_IMG)
	docker run -d $(DOCKER_SCHEDULER_IMG)
	docker run -d $(DOCKER_SENDER_IMG)

version: build
	$(BIN) version

test:
	go test -race ./internal/... # ./pkg/...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.52.2

lint: install-lint-deps
	golangci-lint run --timeout 2m0s ./...

install-goose:
	go install github.com/pressly/goose/v3/cmd/goose@latest

migrate: install-goose
	goose -dir $(MIGRATIONS_DIR) postgres $(DB_STRING) up

generate:
	protoc --proto_path=./api/ --go_out=./internal/server/grpc --go-grpc_out=./internal/server/grpc calendar_service.proto

run_mq:
	docker run -d \
		--name $(RABBIT_NAME) \
		-p 15672:15672 \
		-p 5672:5672 \
		rabbitmq:3-management

stop_mq:
	docker stop $(RABBIT_NAME)
	docker rm $(RABBIT_NAME)

run_db:
	docker run -d \
		--name $(POSTGRES_NAME) \
		-e POSTGRES_PASSWORD=postgres \
		-e PGDATA=/var/lib/postgresql/data/pgdata \
		-v pg_data:/var/lib/postgresql/data \
		-p 5432:5432 \
		postgres
	docker exec -it $(POSTGRES_NAME) psql -Upostgres -dpostgres \
    	-c "CREATE DATABASE calendar;" \
    	-c "CREATE USER cuser WITH ENCRYPTED PASSWORD 'cpassword';" \
    	-c "GRANT ALL ON DATABASE calendar TO cuser;" \
    	-c "ALTER DATABASE calendar OWNER TO cuser;" \
    	-c "GRANT USAGE, CREATE ON SCHEMA PUBLIC TO cuser;"

stop_db:
	docker stop $(POSTGRES_NAME)
	docker rm $(POSTGRES_NAME)

up:
	docker compose -f deployments/docker-compose.yaml -p calendar up -d

up-build:
	docker compose -f deployments/docker-compose.yaml -p calendar up -d --build

down:
	docker compose -f deployments/docker-compose.yaml -p calendar down

integration-tests-build:
	docker compose -f deployments/docker-compose.test.yaml -p calendar-test build

integration-tests-build_tests:
	docker compose -f deployments/docker-compose.test.yaml -p calendar-test build tests

integration-tests:
	docker compose -f deployments/docker-compose.test.yaml -p calendar-test up --exit-code-from tests --attach tests && \
	EXIT_CODE=$$? &&\
	docker compose -f deployments/docker-compose.test.yaml -p calendar-test down && \
    echo "command exited with $$EXIT_CODE" && \
    exit $$EXIT_CODE

.PHONY: build build_calendar build_scheduler build_sender run run_scheduler run_sender build-img run-img version test lint migrate generate run_mq stop_mq run_db stop_db up up-build down integration-tests-build integration-tests integration-tests-build_tests
