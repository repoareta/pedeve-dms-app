.PHONY: dev up down restart logs clean help seed-companies test test-frontend lint lint-backend lint-frontend

# Default target
.DEFAULT_GOAL := help

# Development - Run all services
dev:
	@echo "🚀 Starting DMS App Development Environment..."
	@if docker ps > /dev/null 2>&1; then \
		docker-compose -f docker-compose.dev.yml up --build; \
	else \
		echo "⚠️  Docker daemon tidak berjalan."; \
		echo "💡 Untuk development tanpa Docker:"; \
		echo "   - Backend: cd backend && go run ./cmd/api/main.go"; \
		echo "   - Frontend: cd frontend && npm run dev"; \
		exit 1; \
	fi

# Start services in background
up:
	@echo "📦 Starting services in background..."
	@if docker ps > /dev/null 2>&1; then \
		docker-compose -f docker-compose.dev.yml up -d --build; \
	else \
		echo "⚠️  Docker daemon tidak berjalan. Silakan start Docker Desktop terlebih dahulu."; \
		exit 1; \
	fi

# Stop services
down:
	@echo "🛑 Stopping services..."
	@if docker ps > /dev/null 2>&1; then \
		docker-compose -f docker-compose.dev.yml down; \
	else \
		echo "⚠️  Docker daemon tidak berjalan. Skip docker-compose down."; \
		echo "💡 Jika Anda tidak menggunakan Docker, ini normal."; \
	fi

# Restart services
restart:
	@echo "🔄 Restarting services..."
	@if docker ps > /dev/null 2>&1; then \
		docker-compose -f docker-compose.dev.yml restart; \
	else \
		echo "⚠️  Docker daemon tidak berjalan. Silakan start Docker Desktop terlebih dahulu."; \
		exit 1; \
	fi

# View logs
logs:
	@if docker ps > /dev/null 2>&1; then \
		docker-compose -f docker-compose.dev.yml logs -f; \
	else \
		echo "⚠️  Docker daemon tidak berjalan. Silakan start Docker Desktop terlebih dahulu."; \
		exit 1; \
	fi

# View backend logs only
logs-backend:
	@if docker ps > /dev/null 2>&1; then \
		docker-compose -f docker-compose.dev.yml logs -f backend; \
	else \
		echo "⚠️  Docker daemon tidak berjalan. Silakan start Docker Desktop terlebih dahulu."; \
		exit 1; \
	fi

# View frontend logs only
logs-frontend:
	@if docker ps > /dev/null 2>&1; then \
		docker-compose -f docker-compose.dev.yml logs -f frontend; \
	else \
		echo "⚠️  Docker daemon tidak berjalan. Silakan start Docker Desktop terlebih dahulu."; \
		exit 1; \
	fi

# Restart backend only
restart-backend:
	@echo "🔄 Restarting backend only..."
	@if docker ps > /dev/null 2>&1; then \
		docker-compose -f docker-compose.dev.yml restart backend; \
	else \
		echo "⚠️  Docker daemon tidak berjalan. Silakan start Docker Desktop terlebih dahulu."; \
		exit 1; \
	fi

# Restart frontend only
restart-frontend:
	@echo "🔄 Restarting frontend only..."
	@if docker ps > /dev/null 2>&1; then \
		docker-compose -f docker-compose.dev.yml restart frontend; \
	else \
		echo "⚠️  Docker daemon tidak berjalan. Silakan start Docker Desktop terlebih dahulu."; \
		exit 1; \
	fi

# Clean everything (containers, volumes, networks)
clean:
	@echo "🧹 Cleaning up..."
	@if docker ps > /dev/null 2>&1; then \
		docker-compose -f docker-compose.dev.yml down -v; \
		docker system prune -f; \
	else \
		echo "⚠️  Docker daemon tidak berjalan. Skip cleanup."; \
		echo "💡 Jika Anda tidak menggunakan Docker, ini normal."; \
	fi

# Rebuild and restart
rebuild:
	@echo "🔨 Rebuilding and restarting..."
	docker-compose -f docker-compose.dev.yml up --build -d

# Status check
status:
	@echo "📊 Service Status:"
	docker-compose -f docker-compose.dev.yml ps

# Seed companies
seed-companies: ## Seed sample companies and users (10 subsidiaries with 3-layer hierarchy)
	@echo "🌱 Seeding Companies and Users..."
	@cd backend && DATABASE_URL="postgres://postgres:dms_password@localhost:5432/db_dms_pedeve?sslmode=disable" go run ./cmd/seed-companies

# Run tests (both backend and frontend)
test: ## Run all tests (backend + frontend)
	@echo "🧪 Running all tests..."
	@echo ""
	@BACKEND_FAILED=0; \
	FRONTEND_FAILED=0; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo "🔵 Backend Tests"; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	cd backend && make test || BACKEND_FAILED=1; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo "🟢 Frontend Tests"; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	cd ../frontend && npm run test:unit || FRONTEND_FAILED=1; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo "📊 Test Summary"; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	if [ $${BACKEND_FAILED} -eq 1 ] || [ $${FRONTEND_FAILED} -eq 1 ]; then \
		echo "❌ Some tests failed!"; \
		[ $${BACKEND_FAILED} -eq 1 ] && echo "   🔵 Backend tests failed"; \
		[ $${FRONTEND_FAILED} -eq 1 ] && echo "   🟢 Frontend tests failed"; \
		exit 1; \
	else \
		echo "✅ All tests passed!"; \
	fi

# Run frontend tests only
test-frontend: ## Run frontend unit tests only
	@echo "🧪 Running frontend tests..."
	@cd frontend && npm run test:unit

# Lint backend
lint-backend: ## Run backend linter
	@echo "🔍 Linting backend..."
	@cd backend && golangci-lint run

# Lint frontend
lint-frontend: ## Run frontend linter
	@echo "🔍 Linting frontend..."
	@cd frontend && npm run lint

# Lint all
lint: lint-backend lint-frontend ## Run linters for both backend and frontend

# Help
help:
	@echo "DMS App - Development Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  dev           - Start all services with hot reload (foreground)"
	@echo "  up            - Start all services in background"
	@echo "  down          - Stop all services"
	@echo "  restart       - Restart all services"
	@echo "  logs          - View logs from all services"
	@echo "  logs-backend  - View backend logs only"
	@echo "  logs-frontend - View frontend logs only"
	@echo "  restart-backend  - Restart backend service only"
	@echo "  restart-frontend - Restart frontend service only"
	@echo "  clean         - Stop and remove all containers, volumes, networks"
	@echo "  rebuild       - Rebuild and restart services"
	@echo "  status        - Show service status"
	@echo "  seed-companies - Seed sample companies and users (10 subsidiaries)"
	@echo "  test          - Run all tests (backend + frontend)"
	@echo "  test-frontend - Run frontend unit tests only"
	@echo "  lint          - Run linters for both backend and frontend"
	@echo "  lint-backend  - Run backend linter only"
	@echo "  lint-frontend - Run frontend linter only"
	@echo "  help          - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make dev      - Start development environment"
	@echo "  make up       - Start in background"
	@echo "  make logs     - View logs"
	@echo "  make down     - Stop services"

