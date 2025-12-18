#!/bin/bash

# Script untuk menjalankan full test suite dengan coverage report
# Usage: ./scripts/run-full-tests.sh
# atau: bash backend/scripts/run-full-tests.sh

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BACKEND_DIR="$(dirname "$SCRIPT_DIR")"

cd "$BACKEND_DIR"

echo "=========================================="
echo "🧪 Pedeve DMS Backend - Full Test Suite"
echo "=========================================="
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed. Please install Go first.${NC}"
    exit 1
fi

echo -e "${BLUE}📋 Go Version:${NC}"
go version
echo ""

# Step 1: Run all tests
echo -e "${YELLOW}Step 1: Running all tests...${NC}"
if go test ./... -v; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
else
    echo -e "${RED}❌ Some tests failed!${NC}"
    exit 1
fi
echo ""

# Step 2: Generate coverage report
echo -e "${YELLOW}Step 2: Generating coverage report...${NC}"
if go test ./... -coverprofile=coverage.out; then
    echo -e "${GREEN}✅ Coverage report generated: coverage.out${NC}"
else
    echo -e "${RED}❌ Failed to generate coverage report!${NC}"
    exit 1
fi
echo ""

# Step 3: Show coverage summary
echo -e "${YELLOW}Step 3: Coverage Summary:${NC}"
go tool cover -func=coverage.out | grep total || echo "No coverage data"
echo ""

# Step 4: Generate HTML coverage report
echo -e "${YELLOW}Step 4: Generating HTML coverage report...${NC}"
if go tool cover -html=coverage.out -o coverage.html; then
    echo -e "${GREEN}✅ HTML coverage report generated: coverage.html${NC}"
    echo -e "${BLUE}💡 Open coverage.html in your browser to view detailed coverage${NC}"
else
    echo -e "${RED}❌ Failed to generate HTML coverage report!${NC}"
    exit 1
fi
echo ""

# Step 5: Check coverage threshold (optional)
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
THRESHOLD=70

if (( $(echo "$COVERAGE >= $THRESHOLD" | bc -l) )); then
    echo -e "${GREEN}✅ Coverage: ${COVERAGE}% (meets threshold: ${THRESHOLD}%)${NC}"
else
    echo -e "${YELLOW}⚠️  Coverage: ${COVERAGE}% (below threshold: ${THRESHOLD}%)${NC}"
    echo -e "${YELLOW}   Consider adding more tests to improve coverage${NC}"
fi
echo ""

echo "=========================================="
echo -e "${GREEN}✅ Full test suite completed!${NC}"
echo "=========================================="
echo ""
echo "📊 Files generated:"
echo "  - coverage.out  (coverage data)"
echo "  - coverage.html (HTML coverage report)"
echo ""
echo "💡 Next steps:"
echo "  - Review coverage.html for detailed coverage"
echo "  - Add more tests if coverage is below threshold"
echo "  - Run 'make clean' to remove test artifacts"

