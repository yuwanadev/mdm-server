# Load environment variables from .env file
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Database connection string for goose
# Format: postgres://user:password@host:port/dbname?sslmode=disable
DB_URL=postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# Go settings
GO_BIN=$(shell which go)
GO_MOD_TIDY=$(GO_BIN) mod tidy
GO_BUILD=$(GO_BIN) build
GO_RUN=$(GO_BIN) run

# Paths
MIGRATIONS_DIR=migrations
MAIN_PATH=cmd/server/main.go
BINARY_NAME=server

.PHONY: help tidy build run migrate-up migrate-down migrate-status migrate-create

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  tidy            Run go mod tidy"
	@echo "  build           Build the application"
	@echo "  run             Run the application"
	@echo "  migrate-up      Run migrations up"
	@echo "  migrate-down    Run migrations down"
	@echo "  migrate-status  Show migration status"
	@echo "  migrate-create  Create a new migration (usage: make migrate-create name=migration_name)"

tidy:
	@echo "Tidying go modules..."
	@$(GO_MOD_TIDY)

build: tidy
	@echo "Building $(BINARY_NAME)..."
	@$(GO_BUILD) -o $(BINARY_NAME) $(MAIN_PATH)

run:
	@echo "Running application..."
	@$(GO_RUN) $(MAIN_PATH)

migrate-up:
	@echo "Running migrations up..."
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" up

migrate-down:
	@echo "Running migrations down..."
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" down

migrate-status:
	@echo "Checking migration status..."
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" status

migrate-create:
	@echo "Creating new migration: $(name)..."
	@goose -dir $(MIGRATIONS_DIR) create $(name) sql
