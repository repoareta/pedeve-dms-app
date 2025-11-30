# Makefile untuk Pedeve DMS Backend
# Automated Testing Commands

.PHONY: test test-unit test-integration test-coverage test-coverage-html help

# Run all tests
test:
	@echo "🧪 Running all tests..."
	go test ./... -v

# Run unit tests only (fast)
test-unit:
	@echo "⚡ Running unit tests (fast)..."
	go test ./internal/usecase/... -v
	go test ./internal/repository/... -v

# Run integration tests only
test-integration:
	@echo "🔗 Running integration tests..."
	go test ./internal/delivery/... -v
	go test ./test/integration/... -v

# Run tests with coverage
test-coverage:
	@echo "📊 Running tests with coverage..."
	go test ./... -coverprofile=coverage.out
	@echo "✅ Coverage report generated: coverage.out"
	@go tool cover -func=coverage.out | grep total

# Generate HTML coverage report
test-coverage-html: test-coverage
	@echo "📊 Generating HTML coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ HTML coverage report generated: coverage.html"
	@echo "🌐 Open coverage.html in your browser to view coverage"

# Run tests with verbose output
test-verbose:
	@echo "🔍 Running tests with verbose output..."
	go test ./... -v -cover

# Run tests for specific package
test-package:
	@echo "📦 Usage: make test-package PKG=./internal/usecase"
	@if [ -z "$(PKG)" ]; then \
		echo "❌ Please specify PKG, e.g., make test-package PKG=./internal/usecase"; \
		exit 1; \
	fi
	go test $(PKG) -v

# Clean test artifacts
clean:
	@echo "🧹 Cleaning test artifacts..."
	rm -f coverage.out coverage.html
	@echo "✅ Cleaned"

# Help
help:
	@echo "📚 Available commands:"
	@echo "  make test              - Run all tests"
	@echo "  make test-unit         - Run unit tests only (fast)"
	@echo "  make test-integration  - Run integration tests only"
	@echo "  make test-coverage     - Run tests with coverage report"
	@echo "  make test-coverage-html - Generate HTML coverage report"
	@echo "  make test-verbose     - Run tests with verbose output"
	@echo "  make test-package PKG=./path - Run tests for specific package"
	@echo "  make clean             - Clean test artifacts"
	@echo "  make help              - Show this help message"

