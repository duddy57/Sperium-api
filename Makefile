SHELL := /bin/bash
COMPOSE := docker compose

.PHONY: help
help: ## List available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'

.PHONY: clean
clean:
	go clean
	rm -rf ./build

.PHONY: lint
lint:
	GOOS=linux golangci-lint run

.PHONY: format
format:
	GOOS=linux golangci-lint fmt

.PHONY: test
test:
ifeq ($(TEST_NAME),)
	go test -shuffle=on -count=1 -v ./...
else
	go test -shuffle=on -race -count=1 -v -run ^$(TEST_NAME)$$ ./...
endif

.PHONY: vet
vet:
	go vet ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: build
build: ## Compile all services
	go build ./...

## ----- Local stack (Docker Compose) -----
.PHONY: up
up: ## Build and start the full local stack
	$(COMPOSE) up -d
#	@echo "gateway on http://localhost:8080"
#	@echo "console on http://localhost:3001"
#	@echo "grafana on http://localhost:3000"

.PHONY: migrate
migrate:
	goose up

.PHONY: down
down: ## Stop the stack and remove volumes
	$(COMPOSE) down -v

.PHONY: logs
logs: ## Tail logs from every service
	$(COMPOSE) logs -f

.PHONY: ps
ps: ## Show stack status
	$(COMPOSE) ps
