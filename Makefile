.PHONY: help
help: ## Display this help text
	@grep -hE '^[A-Za-z0-9_ \-]*?:.*##.*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-35s\033[0m %s\n", $$1, $$2}'

.PHONY: env
env: ## Creates a default .env file if it doesn't exist
	@./build/env.sh

.PHONY: env-dockerized
env-dockerized: ## Runs env dockerised.
	@./build/docker/env.sh

.PHONY: deps
deps: ## Installs dependencies
	@./build/deps.sh

.PHONY: deps-dockerized
deps-dockerized: ## Runs deps dockerised.
	@./build/docker/deps.sh

.PHONY: deps-upgrade
deps-upgrade: ## Installs/upgrades all dependencies
	@./build/deps-upgrade.sh

.PHONY: deps-upgrade-dockerized
deps-upgrade-dockerized: ## Runs deps-upgrade dockerised.
	@./build/docker/deps-upgrade.sh

.PHONY: clean
clean: ## Removes build artifacts and vendor directories
	@./build/clean.sh

.PHONY: clean-dockerized
clean-dockerized: ## Runs clean dockerised.
	@./build/docker/clean.sh

.PHONY: generate-code
generate-code: ## Generates all generated code
	@./build/generate-code.sh

.PHONY: generate-code-dockerized
generate-code-dockerized: ## Runs generate-code dockerised.
	@./build/docker/generate-code.sh

.PHONY: docs
docs: ## Displays source code documentation.
	@./build/docs.sh

.PHONY: docs-dockerized
docs-dockerized: ## Runs docs dockerised.
	@./build/docker/docs.sh

.PHONY: lint
lint: ## Runs linting
	@./build/lint.sh

.PHONY: lint-dockerized
lint-dockerized: ## Runs lint dockerised.
	@./build/docker/lint.sh

.PHONY: lint-optional
lint-optional: ## Runs linting with more linters that are not required for CI to pass.
	@./build/lint-optional.sh

.PHONY: lint-optional-dockerized
lint-optional-dockerized: ## Runs lint-optional dockerised.
	@./build/docker/lint-optional.sh

.PHONY: test
test: ## Run all tests
	@./build/test.sh

.PHONY: test-dockerized
test-dockerized: ## Runs test dockerised.
	@./build/docker/test.sh

.PHONY: generate-test-coverage
generate-test-coverage: ## Generates test coverage reports
	@./build/generate-test-coverage.sh

.PHONY: generate-test-coverage-dockerized
generate-test-coverage-dockerized: ## Runs generate-test-coverage dockerised.
	@./build/docker/generate-test-coverage.sh

.PHONY: quality
quality: ## Runs all quality checks
	@./build/quality.sh

.PHONY: quality-dockerized
quality-dockerized: ## Runs quality dockerised.
	@./build/docker/quality.sh

.PHONY: run
run: ## Runs the application (containerised)
	@docker-compose --env-file ./.env up weather-api