include ./build/Makefile

.PHONY: run
run: ## Runs the application (containerised)
	@docker-compose --env-file ./.env up weather-api