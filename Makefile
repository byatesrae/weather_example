SHELL := /bin/bash

.PHONY: help
help: ## Display this help text
	@echo 'Perform common development tasks.'
	@echo 'Usage: make [TARGET]'
	@echo 'Targets:'
	@grep '^[a-zA-Z]' $(MAKEFILE_LIST) | awk -F ':.*?## ' 'NF==2 {printf "\033[36m  %-25s\033[0m %s\n", $$1, $$2}'

.PHONY: deps
deps: ## Installs dependencies
	go mod tidy
	go mod vendor

.PHONY: deps-upgrade
deps-upgrade: ## Installs/upgrades all dependencies
	go get -u -t -d -v ./...
	go mod tidy
	go mod vendor

.PHONY: clean
clean: ## Removes build artifacts and vendor directories
	./build/run.sh "./build/clean.sh"

.PHONY: generate
generate: ## Generates all generated code
	./build/run.sh "go generate ./..."

.PHONY: lint
lint: ## Runs linting
	./build/run.sh "./build/lint.sh"

.PHONY: test
test: ## Run all tests
	./build/run.sh "./build/test.sh"

.PHONY: generate-test-coverage
generate-test-coverage: ## Generates test coverage reports
	./build/run.sh "./build/generate-test-coverage.sh"

.PHONY: run
run: ## Runs the application (containerised)
	@docker-compose up weather-api

.PHONY: quality
quality: clean lint test generate-test-coverage ## Runs all quality checks
	@echo Done
