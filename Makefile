include ./build/Makefile

.PHONY: build
build: ## build the application
	@docker-compose --env-file ./.env build weather-api

.PHONY: run
run: ## Runs the application (containerised)
	@docker-compose --env-file ./.env up weather-api