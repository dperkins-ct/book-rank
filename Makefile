.PHONY: build run test clean docker-up docker-down docker-build lint help

# Variables
APP_NAME := bookrank
BINARY := bookrank
GO_VERSION := 1.21
DOCKER_COMPOSE := docker-compose

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

# Development targets
build: ## Build the application
	@echo "Building $(APP_NAME)..."
	go build -o $(BINARY) cmd/server/main.go

run: build ## Build and run the application locally
	@echo "Running $(APP_NAME)..."
	./$(BINARY)

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

# Docker targets
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):latest .

docker-up: ## Start all services with Docker Compose
	@echo "Starting services..."
	$(DOCKER_COMPOSE) up -d

docker-up-dev: ## Start development services (PostgreSQL only)
	@echo "Starting development services..."
	$(DOCKER_COMPOSE) up -d postgres redis

docker-down: ## Stop all services
	@echo "Stopping services..."
	$(DOCKER_COMPOSE) down

docker-logs: ## Show Docker logs
	@echo "Showing logs..."
	$(DOCKER_COMPOSE) logs -f

docker-clean: ## Clean Docker containers and images
	@echo "Cleaning Docker resources..."
	$(DOCKER_COMPOSE) down -v --remove-orphans
	docker system prune -f

# Database targets
db-migrate: ## Run database migrations
	@echo "Running database migrations..."
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/bookrank?sslmode=disable" up

db-rollback: ## Rollback database migrations
	@echo "Rolling back database migrations..."
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/bookrank?sslmode=disable" down

db-reset: ## Reset database (drop and recreate)
	@echo "Resetting database..."
	$(DOCKER_COMPOSE) exec postgres psql -U postgres -c "DROP DATABASE IF EXISTS bookrank;"
	$(DOCKER_COMPOSE) exec postgres psql -U postgres -c "CREATE DATABASE bookrank;"

# Code quality targets
fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

vet: ## Vet code
	@echo "Vetting code..."
	go vet ./...

mod-tidy: ## Tidy Go modules
	@echo "Tidying Go modules..."
	go mod tidy

mod-download: ## Download Go modules
	@echo "Downloading Go modules..."
	go mod download

# Utility targets
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY)
	rm -f coverage.out coverage.html
	go clean

deps: ## Install development dependencies
	@echo "Installing development dependencies..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

setup: deps docker-up-dev ## Setup development environment
	@echo "Development environment setup complete!"
	@echo "Run 'make run' to start the application"

start: ## Start everything (database + backend + frontend) - One command setup
	@echo "🚀 Starting BookRank full-stack development environment..."
	@echo "1. Starting Colima..."
	@colima start 2>/dev/null || echo "   Colima already running ✅"
	@echo "2. Starting database services..."
	@$(DOCKER_COMPOSE) up -d postgres redis
	@echo "3. Waiting for database to be ready..."
	@sleep 5
	@echo "4. Installing frontend dependencies..."
	@cd frontend && npm install --silent
	@echo "5. Starting backend API..."
	@go build ./cmd/server && ./server &
	@echo "6. Starting React frontend..."
	@cd frontend && npm run dev &
	@echo ""
	@echo "🎉 BookRank is now running!"
	@echo "   📱 Frontend (React): http://localhost:3000"
	@echo "   🔧 Backend API: http://localhost:8080"
	@echo "   ❤️ Health check: http://localhost:8080/health"
	@echo ""
	@echo "Press Ctrl+C to stop (you may need to run 'make stop-all' to cleanup)"

start-backend-only: ## Start just backend (database + API)
	@echo "🚀 Starting BookRank backend..."
	@colima start 2>/dev/null || echo "Colima already running ✅"
	@$(DOCKER_COMPOSE) up -d postgres redis
	@sleep 5
	@$(MAKE) run

start-frontend-only: ## Start just frontend (assumes backend is running)
	@echo "🚀 Starting BookRank frontend..."
	@$(MAKE) frontend-dev

stop-all: ## Stop everything (application + database + colima)
	@echo "🛑 Stopping BookRank development environment..."
	@echo "1. Stopping database services..."
	@$(DOCKER_COMPOSE) down
	@echo "2. Optionally stopping colima (uncomment if desired)..."
	@# colima stop
	@echo "✅ Environment stopped"

restart: stop-all start ## Restart everything

dev: start ## Alias for start (shorter command)

# Production targets
build-prod: ## Build for production
	@echo "Building for production..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o $(BINARY) cmd/server/main.go

docker-prod: ## Build and tag production Docker image
	@echo "Building production Docker image..."
	docker build -f Dockerfile.production -t $(APP_NAME):latest -t $(APP_NAME):$$(git rev-parse --short HEAD) .

# Deployment targets
deploy-staging: ## Deploy to staging environment
	@echo "Deploying to staging..."
	./scripts/deployment/deploy.sh -e staging

deploy-production: ## Deploy to production environment
	@echo "Deploying to production..."
	./scripts/deployment/deploy.sh -e production

deploy-infra-staging: ## Deploy only infrastructure to staging
	@echo "Deploying infrastructure to staging..."
	./scripts/deployment/deploy.sh -e staging --skip-docker

deploy-infra-production: ## Deploy only infrastructure to production
	@echo "Deploying infrastructure to production..."
	./scripts/deployment/deploy.sh -e production --skip-docker

destroy-staging: ## Destroy staging environment (careful!)
	@echo "Destroying staging environment..."
	./scripts/deployment/deploy.sh -e staging --destroy

destroy-production: ## Destroy production environment (careful!)
	@echo "Destroying production environment..."
	./scripts/deployment/deploy.sh -e production --destroy

# Terraform targets
tf-init: ## Initialize Terraform
	@echo "Initializing Terraform..."
	cd deployment/terraform && terraform init

tf-plan-staging: ## Plan Terraform changes for staging
	@echo "Planning Terraform changes for staging..."
	cd deployment/terraform && terraform workspace select staging && terraform plan -var-file="environments/staging.tfvars"

tf-plan-production: ## Plan Terraform changes for production
	@echo "Planning Terraform changes for production..."
	cd deployment/terraform && terraform workspace select production && terraform plan -var-file="environments/production.tfvars"

tf-apply-staging: ## Apply Terraform changes for staging
	@echo "Applying Terraform changes for staging..."
	cd deployment/terraform && terraform workspace select staging && terraform apply -var-file="environments/staging.tfvars"

tf-apply-production: ## Apply Terraform changes for production
	@echo "Applying Terraform changes for production..."
	cd deployment/terraform && terraform workspace select production && terraform apply -var-file="environments/production.tfvars"

# Frontend targets
frontend-install: ## Install frontend dependencies
	@echo "Installing frontend dependencies..."
	cd frontend && npm ci

frontend-build: ## Build frontend for production
	@echo "Building frontend..."
	cd frontend && npm run build

frontend-dev: ## Start frontend development server
	@echo "Starting frontend development server..."
	cd frontend && npm run dev

# Database backup targets
backup-db: ## Create database backup
	@echo "Creating database backup..."
	./scripts/deployment/backup.sh

# Health check
health: ## Check application health
	@echo "Checking application health..."
	curl -f http://localhost:8080/health || exit 1

health-detailed: ## Check detailed application health
	@echo "Checking detailed application health..."
	curl -s http://localhost:8080/health | jq .

# Monitoring targets
logs-staging: ## View staging logs
	@echo "Viewing staging logs..."
	aws logs tail /ecs/bookrank-staging --follow --region us-west-2

logs-production: ## View production logs
	@echo "Viewing production logs..."
	aws logs tail /ecs/bookrank-production --follow --region us-west-2

dashboard-staging: ## Open staging CloudWatch dashboard
	@echo "Opening staging dashboard..."
	open "https://us-west-2.console.aws.amazon.com/cloudwatch/home?region=us-west-2#dashboards:name=bookrank-staging-dashboard"

dashboard-production: ## Open production CloudWatch dashboard
	@echo "Opening production dashboard..."
	open "https://us-west-2.console.aws.amazon.com/cloudwatch/home?region=us-west-2#dashboards:name=bookrank-production-dashboard"