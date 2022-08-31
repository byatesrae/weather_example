.PHONY: help
help: ## Display this help text
	@echo 'Perform common development tasks.'
	@echo 'Usage: make [TARGET]'
	@echo 'Targets:'
	@grep '^[a-zA-Z]' $(MAKEFILE_LIST) | awk -F ':.*?## ' 'NF==2 {printf "\033[36m  %-25s\033[0m %s\n", $$1, $$2}'

.PHONY: env
env: ## Creates a default .env file if it doesn't exist
	docker run --rm -v $(PWD):/src busybox:stable cp -n /src/.env.example /src/.env

.PHONY: deps
deps: env ## Installs dependencies
	./build/run-bash.sh "./build/deps.sh"

.PHONY: deps-upgrade
deps-upgrade: env ## Installs/upgrades all dependencies
	./build/run-bash.sh "./build/deps-upgrade.sh"

.PHONY: clean
clean: env ## Removes build artifacts and vendor directories
	./build/run-bash.sh "./build/clean.sh"

.PHONY: generate-code
generate-code: env ## Generates all generated code
	./build/run-bash.sh "./build/generate-code.sh"

.PHONY: lint
lint: env ## Runs linting
	./build/run-bash.sh "./build/lint.sh"

.PHONY: lint-optional
lint-optional: env ## Runs linting with more linters and always succeeds
	./build/run-bash.sh "./build/lint-optional.sh"

.PHONY: test
test: env ## Run all tests
	./build/run-bash.sh "./build/test.sh"

.PHONY: generate-test-coverage
generate-test-coverage: env ## Generates test coverage reports
	./build/run-bash.sh "./build/generate-test-coverage.sh"

.PHONY: run
run: env ## Runs the application (containerised)
	@docker-compose up weather-api

.PHONY: quality
quality: env clean lint test generate-test-coverage ## Runs all quality checks
	@echo Done
