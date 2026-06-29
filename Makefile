SHELL := /bin/bash

SERVICES := api-gateway identity-service logistics-service mobility-service payment-service operations-service mcp-server
SERVICE_DIRS := $(addprefix services/,$(SERVICES))
MIGRATE ?= migrate
GO ?= go
DOCKER_COMPOSE ?= docker compose
PROTOC ?= protoc
ROOT_DIR := $(CURDIR)
export GOCACHE := $(ROOT_DIR)/.cache/go-build

.PHONY: dev build test test-kafka-integration test-integration test-contracts test-e2e test-load test-chaos test-migrations lint migrate-up migrate-down proto proto-check-tools proto-lint proto-verify docker-build clean

dev:
	$(DOCKER_COMPOSE) up --build

build:
	mkdir -p bin
	for service in $(SERVICES); do \
		$(GO) build -o bin/$$service ./services/$$service/cmd/server; \
	done

test:
	for dir in pkg/* services/*; do \
		if [ -f "$$dir/go.mod" ]; then \
			(cd "$$dir" && $(GO) test ./...); \
		fi; \
	done
	(cd tests && $(GO) test ./...)

test-kafka-integration:
	cd pkg/kafka && KAFKA_INTEGRATION_BROKERS=$${KAFKA_INTEGRATION_BROKERS:-localhost:9092} $(GO) test -run TestKafkaRealBrokerIntegration ./...

test-integration:
	cd tests && TEST_INTEGRATION=1 $(GO) test -run TestInfrastructureIntegration ./...

test-contracts:
	cd tests && TEST_INTEGRATION=1 $(GO) test -run TestHTTPContracts_CreateOrderAndInitializePayment ./...

test-e2e:
	cd tests && TEST_INTEGRATION=1 $(GO) test -run TestOrderToPaymentWorkflowE2E ./...

test-load:
	cd tests && TEST_INTEGRATION=1 $(GO) test -run TestGatewayHealthLoad ./...

test-chaos:
	cd tests && TEST_INTEGRATION=1 $(GO) test -run TestKafkaPoisonMessageDLQ ./...

test-migrations:
	cd tests && TEST_INTEGRATION=1 $(GO) test -run TestMigrationsRollbackInTransaction ./...

lint:
	golangci-lint run ./...

migrate-up:
	for schema in identity logistics mobility payment operations; do \
		$(MIGRATE) -path infra/migrations/$$schema -database "$${MIGRATE_DATABASE_URL:?set MIGRATE_DATABASE_URL}" up; \
	done

migrate-down:
	for schema in operations payment mobility logistics identity; do \
		$(MIGRATE) -path infra/migrations/$$schema -database "$${MIGRATE_DATABASE_URL:?set MIGRATE_DATABASE_URL}" down 1; \
	done

proto:
	bash scripts/proto-generate.sh

proto-check-tools:
	@command -v $(PROTOC) >/dev/null 2>&1 || (echo "missing protoc" && exit 1)
	@command -v protoc-gen-go >/dev/null 2>&1 || (echo "missing protoc-gen-go" && exit 1)
	@command -v protoc-gen-go-grpc >/dev/null 2>&1 || (echo "missing protoc-gen-go-grpc" && exit 1)

proto-lint:
	bash scripts/proto-lint.sh

proto-verify:
	bash scripts/proto-verify.sh

docker-build:
	for service in $(SERVICES); do \
		docker build -t logistics-platform/$$service:latest -f services/$$service/Dockerfile .; \
	done

clean:
	rm -rf bin
	rm -rf .cache
	find . -name '.air.toml' -prune -o -name '*.test' -delete
