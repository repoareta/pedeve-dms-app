.PHONY: dev up down restart logs clean help

# Default target
.DEFAULT_GOAL := help

# Development - Run all services
dev:
	@echo "🚀 Starting DMS App Development Environment..."
	docker-compose -f docker-compose.dev.yml up --build

# Start services in background
up:
	@echo "📦 Starting services in background..."
	docker-compose -f docker-compose.dev.yml up -d --build

# Stop services
down:
	@echo "🛑 Stopping services..."
	docker-compose -f docker-compose.dev.yml down

# Restart services
restart:
	@echo "🔄 Restarting services..."
	docker-compose -f docker-compose.dev.yml restart

# View logs
logs:
	docker-compose -f docker-compose.dev.yml logs -f

# View backend logs only
logs-backend:
	docker-compose -f docker-compose.dev.yml logs -f backend

# View frontend logs only
logs-frontend:
	docker-compose -f docker-compose.dev.yml logs -f frontend

# Clean everything (containers, volumes, networks)
clean:
	@echo "🧹 Cleaning up..."
	docker-compose -f docker-compose.dev.yml down -v
	docker system prune -f

# Rebuild and restart
rebuild:
	@echo "🔨 Rebuilding and restarting..."
	docker-compose -f docker-compose.dev.yml up --build -d

# Status check
status:
	@echo "📊 Service Status:"
	docker-compose -f docker-compose.dev.yml ps

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
	@echo "  clean         - Stop and remove all containers, volumes, networks"
	@echo "  rebuild       - Rebuild and restart services"
	@echo "  status        - Show service status"
	@echo "  help          - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make dev      - Start development environment"
	@echo "  make up       - Start in background"
	@echo "  make logs     - View logs"
	@echo "  make down     - Stop services"

