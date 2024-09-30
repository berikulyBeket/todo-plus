include .env
export

LOCAL_BIN:=$(CURDIR)/bin
PATH:=$(LOCAL_BIN):$(PATH)

compose-up: ### Run docker-compose
	docker-compose --env-file .env.dev --profile default up -d --build
.PHONY: compose-up

compose-up-services: ### Run docker-compose services
	docker-compose --env-file .env --profile services up -d --build
.PHONY: compose-up-services

compose-up-integration: ### Run docker-compose with integration test
	docker-compose --env-file .env.dev --profile integration up --build --abort-on-container-exit --exit-code-from integration
.PHONY: compose-up-integration-test

compose-down: ### Down docker-compose with a specific profile
	@if [ -z "$(profile)" ]; then \
		docker-compose --profile default down --remove-orphans; \
	else \
		docker-compose --profile $(profile) down --remove-orphans; \
	fi
.PHONY: compose-down

run: swag-v1 ### run
	go mod tidy && go mod download && \
	GIN_MODE=debug CGO_ENABLED=0 go run ./cmd/app
.PHONY: run

swag-v1: ### swag init
	swag init -g internal/controller/http/v1/handler.go --outputTypes json
.PHONY: swag-v1

linter-golangci: ### check by golangci linter
	golangci-lint run
.PHONY: linter-golangci

unit-test: ### run unit test
	go test -v ./internal/...
.PHONY: test

integration-test: ### run integration-test
	go clean -testcache && go test -v ./integration-test/...
.PHONY: integration-test

migrate-create:  ### create new migration
	migrate create -ext sql -dir migrations 'migrate_name'
.PHONY: migrate-create

migrate-up: ### migration up
	migrate -path migrations -database '$(PG_URL)' up
.PHONY: migrate-up

migrate-down: ### migration up
	migrate -path migrations -database '$(PG_URL)' down
.PHONY: migrate-down

generate-ssl-certs: ## Generate SSL certificates
	chmod +x generate_ssl_certs.sh
	./scripts/generate_ssl_certs.sh
.PHONY: generate-ssl-certs

generate-sentinel-conf: ## Generate sentinel.conf
	chmod +x ./scripts/generate_sentinel_conf.sh
	./scripts/generate_sentinel_conf.sh
.PHONY: generate-sentinel-conf

generate-redis-slave-conf: ## Generate redis_slave.conf
	chmod +x ./scripts/generate_redis_slave_conf.sh
	./scripts/generate_redis_slave_conf.sh
.PHONY: generate-sentinel-conf

bin-deps:
	GOBIN=$(LOCAL_BIN) go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
