.PHONY: help build run test clean install dev-backend dev-frontend

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## Install all dependencies
	go mod download
	cd frontend && npm install

build: ## Build backend and frontend
	go build -o bin/librecov-server backend/cmd/server/main.go
	cd frontend && npm run build

run: ## Run the application (production mode)
	./bin/librecov-server

dev-backend: ## Run backend in development mode
	go run backend/cmd/server/main.go

dev-frontend: ## Run frontend in development mode
	cd frontend && npm run dev

test: ## Run all tests
	go test -v ./...
	cd frontend && npm run test

test-backend: ## Run backend tests
	go test -v ./...

test-coverage: ## Run backend tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run linters
	go fmt ./...
	go vet ./...

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf frontend/dist/
	rm -f coverage.out coverage.html

.DEFAULT_GOAL := help
