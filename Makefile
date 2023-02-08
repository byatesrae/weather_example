.PHONY: help
help: ## Display this help text
	@grep -hE '^[A-Za-z0-9_ \-]*?:.*##.*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: env
env: ## Creates a default .env file if it doesn't exist
	@./build/env.sh

.PHONY: deps
deps: ## Installs dependencies
	@./build/deps.sh

.PHONY: deps-upgrade
deps-upgrade: ## Installs/upgrades all dependencies
	@./build/deps-upgrade.sh

.PHONY: clean
clean: ## Removes build artifacts and vendor directories
	@./build/clean.sh

.PHONY: generate-code
generate-code: ## Generates all generated code
	@./build/generate-code.sh

.PHONY: doc
doc: ## Runs godoc documentation.
	@./build/doc.sh

.PHONY: lint
lint: ## Runs linting
	@./build/lint.sh

.PHONY: lint-optional
lint-optional: ## Runs linting with more linters that are not required for CI to pass.
	@./build/lint-optional.sh

.PHONY: test
test: ## Run all tests
	@./build/test.sh

.PHONY: generate-test-coverage
generate-test-coverage: ## Generates test coverage reports
	@./build/generate-test-coverage.sh

.PHONY: run
run: ## Runs the application (containerised)
	@docker-compose --env-file ./.env up weather-api

.PHONY: quality
quality: env clean lint test generate-test-coverage ## Runs all quality checks
	@echo Done