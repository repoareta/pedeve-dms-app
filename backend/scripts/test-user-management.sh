#!/bin/bash

# Script untuk test User Management Multi-Level
# Usage: ./backend/scripts/test-user-management.sh

set -e

BASE_URL="http://localhost:8080"
API_URL="${BASE_URL}/api/v1"

echo "=========================================="
echo "Testing User Management Multi-Level"
echo "=========================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

# Test 1: Get CSRF Token
echo "1. Getting CSRF Token..."
CSRF_RESPONSE=$(curl -s "${API_URL}/csrf-token")
CSRF_TOKEN=$(echo $CSRF_RESPONSE | jq -r '.csrf_token // empty')
if [ -z "$CSRF_TOKEN" ]; then
    print_error "Failed to get CSRF token"
    exit 1
fi
print_success "CSRF Token: ${CSRF_TOKEN:0:20}..."
echo ""

# Test 2: Login as Superadmin
echo "2. Logging in as superadmin..."
LOGIN_RESPONSE=$(curl -s -X POST "${API_URL}/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"superadmin","password":"Pedeve123"}' \
  -c /tmp/dms_cookies.txt)

AUTH_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token // empty')
if [ -z "$AUTH_TOKEN" ]; then
    print_error "Failed to login"
    echo "Response: $LOGIN_RESPONSE"
    exit 1
fi
print_success "Logged in successfully"
echo "Token: ${AUTH_TOKEN:0:30}..."
echo ""

# Test 3: Get Profile
echo "3. Getting user profile..."
PROFILE_RESPONSE=$(curl -s -X GET "${API_URL}/auth/profile" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -b /tmp/dms_cookies.txt)

USER_ID=$(echo $PROFILE_RESPONSE | jq -r '.id // empty')
ROLE=$(echo $PROFILE_RESPONSE | jq -r '.role // empty')
if [ -z "$USER_ID" ]; then
    print_error "Failed to get profile"
    exit 1
fi
print_success "Profile retrieved - Role: $ROLE"
echo ""

# Test 4: Get All Roles (untuk mendapatkan role_id)
echo "4. Getting roles..."
# Note: Endpoint roles belum ada, jadi kita skip dulu
print_info "Skipping role check (endpoint belum ada)"
echo ""

# Test 5: Create Company (Root Level)
echo "5. Creating root company..."
COMPANY_RESPONSE=$(curl -s -X POST "${API_URL}/companies" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  -b /tmp/dms_cookies.txt \
  -d '{
    "name": "Test Holding Company",
    "code": "TEST-HOLDING-001",
    "description": "Test holding company for testing"
  }')

COMPANY_ID=$(echo $COMPANY_RESPONSE | jq -r '.id // empty')
if [ -z "$COMPANY_ID" ]; then
    print_error "Failed to create company"
    echo "Response: $COMPANY_RESPONSE"
    exit 1
fi
print_success "Company created - ID: $COMPANY_ID"
echo ""

# Test 6: Get All Companies
echo "6. Getting all companies..."
COMPANIES_RESPONSE=$(curl -s -X GET "${API_URL}/companies" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -b /tmp/dms_cookies.txt)

COMPANIES_COUNT=$(echo $COMPANIES_RESPONSE | jq '. | length')
print_success "Found $COMPANIES_COUNT companies"
echo ""

# Test 7: Get Company by ID
echo "7. Getting company by ID..."
COMPANY_DETAIL=$(curl -s -X GET "${API_URL}/companies/$COMPANY_ID" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -b /tmp/dms_cookies.txt)

COMPANY_NAME=$(echo $COMPANY_DETAIL | jq -r '.name // empty')
if [ -z "$COMPANY_NAME" ]; then
    print_error "Failed to get company"
    exit 1
fi
print_success "Company retrieved - Name: $COMPANY_NAME"
echo ""

# Test 8: Create Sub-Company
echo "8. Creating sub-company..."
SUB_COMPANY_RESPONSE=$(curl -s -X POST "${API_URL}/companies" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  -b /tmp/dms_cookies.txt \
  -d "{
    \"name\": \"Test Subsidiary 1\",
    \"code\": \"TEST-SUB-001\",
    \"description\": \"Test subsidiary\",
    \"parent_id\": \"$COMPANY_ID\"
  }")

SUB_COMPANY_ID=$(echo $SUB_COMPANY_RESPONSE | jq -r '.id // empty')
if [ -z "$SUB_COMPANY_ID" ]; then
    print_error "Failed to create sub-company"
    echo "Response: $SUB_COMPANY_RESPONSE"
    exit 1
fi
print_success "Sub-company created - ID: $SUB_COMPANY_ID"
echo ""

# Test 9: Get Company Children
echo "9. Getting company children..."
CHILDREN_RESPONSE=$(curl -s -X GET "${API_URL}/companies/$COMPANY_ID/children" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -b /tmp/dms_cookies.txt)

CHILDREN_COUNT=$(echo $CHILDREN_RESPONSE | jq '. | length')
print_success "Found $CHILDREN_COUNT children"
echo ""

# Test 10: Get All Users
echo "10. Getting all users..."
USERS_RESPONSE=$(curl -s -X GET "${API_URL}/users" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -b /tmp/dms_cookies.txt)

USERS_COUNT=$(echo $USERS_RESPONSE | jq '. | length')
print_success "Found $USERS_COUNT users"
echo ""

echo "=========================================="
print_success "All basic tests passed!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "1. Test create user dengan company assignment"
echo "2. Test access control (login sebagai admin company)"
echo "3. Test update/delete operations"
echo ""

