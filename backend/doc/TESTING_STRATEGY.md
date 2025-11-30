# Automated Testing Strategy - Planning & Implementation

## 📋 Executive Summary

Dokumen ini menjelaskan strategi, planning, dan implementasi automated testing untuk Pedeve DMS App. Tujuan utama adalah:
1. **Memastikan logika bisnis berjalan sesuai harapan**
2. **Mencegah regresi** saat ada perbaikan bug dan penambahan fitur
3. **Memungkinkan refactoring dengan aman** - test akan memberitahu jika refactoring merusak fungsionalitas

---

## 🎯 Tujuan Testing

### Primary Goals
- ✅ **Regression Prevention**: Mencegah bug yang sudah diperbaiki muncul lagi
- ✅ **Business Logic Validation**: Memastikan logika bisnis (hierarchy, RBAC, leveling) berjalan benar
- ✅ **Safe Refactoring**: Memungkinkan refactoring tanpa takut merusak fungsionalitas
- ✅ **Documentation**: Test sebagai dokumentasi hidup tentang bagaimana sistem bekerja
- ✅ **Confidence**: Memberikan confidence saat deploy ke production

### Critical Areas to Test
1. **Company Hierarchy** (Jantung aplikasi)
   - Level calculation (holding = 0, anak = 1, cucu = 2, dst)
   - GetDescendants (recursive CTE)
   - GetAncestors
   - Parent-child relationships
   - Holding company protection (code = "PDV")

2. **RBAC (Role-Based Access Control)**
   - Superadmin vs Admin vs Regular user access
   - Company-based access control
   - Permission validation
   - Access to descendants

3. **Company CRUD Operations**
   - Create company (with/without parent)
   - Update company (level recalculation)
   - Delete company (soft delete)
   - Duplicate prevention

4. **User Management**
   - Create user with company assignment
   - Update user
   - Delete user
   - Company hierarchy access

5. **Audit Logging**
   - Log creation for all actions
   - Permanent vs temporary logs
   - Log filtering and retrieval

---

## 🏗️ Testing Pyramid

```
        /\
       /  \      E2E Tests (5%)
      /____\     - Full integration tests
     /      \    - Critical user flows
    /________\   Integration Tests (25%)
   /          \  - API endpoint tests
  /____________\ Unit Tests (70%)
                 - Business logic
                 - Helper functions
                 - Repository methods
```

### Distribution
- **70% Unit Tests**: Fast, isolated, test business logic
- **25% Integration Tests**: Test API endpoints with test database
- **5% E2E Tests**: Test critical user flows end-to-end

---

## 📦 Testing Stack

### Tools & Libraries
1. **Go Testing Package** (built-in)
   - Standard Go testing framework
   - `testing` package untuk unit tests
   - `testing/quick` untuk property-based testing (optional)

2. **Testify** (already in dependencies)
   - `github.com/stretchr/testify/assert` - Assertions
   - `github.com/stretchr/testify/mock` - Mocking
   - `github.com/stretchr/testify/suite` - Test suites
   - `github.com/stretchr/testify/require` - Required assertions

3. **Test Database**
   - **PostgreSQL** (same as production)
   - **Test Containers** (optional, untuk isolated test DB)
   - **SQLite in-memory** (untuk fast unit tests)

4. **HTTP Testing**
   - `net/http/httptest` (built-in)
   - Fiber test helpers

5. **Code Coverage**
   - `go test -cover`
   - `go test -coverprofile=coverage.out`
   - `go tool cover -html=coverage.out`

---

## 📁 Test Structure

### Directory Layout
```
backend/
├── internal/
│   ├── usecase/
│   │   ├── company_usecase.go
│   │   ├── company_usecase_test.go      # Unit tests
│   │   ├── user_management_usecase.go
│   │   └── user_management_usecase_test.go
│   │
│   ├── repository/
│   │   ├── company_repository.go
│   │   ├── company_repository_test.go    # Repository tests
│   │   └── ...
│   │
│   └── delivery/
│       └── http/
│           ├── company_handler.go
│           ├── company_handler_test.go   # Integration tests
│           └── ...
│
├── test/
│   ├── fixtures/                          # Test data
│   │   ├── companies.json
│   │   └── users.json
│   │
│   ├── helpers/                           # Test helpers
│   │   ├── database.go                    # Test DB setup
│   │   ├── auth.go                        # Auth helpers
│   │   └── assertions.go                  # Custom assertions
│   │
│   └── integration/                       # Integration tests
│       ├── company_api_test.go
│       └── user_api_test.go
│
└── cmd/
    └── test-runner/                       # Test runner script
        └── main.go
```

### Naming Convention
- Test files: `*_test.go`
- Test functions: `TestFunctionName` atau `TestFunctionName_Scenario`
- Test tables: `TestFunctionName_TableDriven`
- Benchmarks: `BenchmarkFunctionName`

---

## 🧪 Test Types & Examples

### 1. Unit Tests (70%)

#### 1.1 UseCase Tests
**Location**: `internal/usecase/*_test.go`

**Example**: `company_usecase_test.go`
```go
func TestCompanyUseCase_CreateCompany(t *testing.T) {
    // Test business logic without database
    // Mock repository
    // Test level calculation
    // Test validation
}

func TestCompanyUseCase_UpdateCompanyFull_LevelCalculation(t *testing.T) {
    // Test level recalculation
    // Test holding protection
    // Test parent change
}

func TestCompanyUseCase_GetCompanyDescendants_NoDuplicates(t *testing.T) {
    // Test no duplicate in descendants
    // Test hierarchy depth
}
```

#### 1.2 Repository Tests
**Location**: `internal/repository/*_test.go`

**Example**: `company_repository_test.go`
```go
func TestCompanyRepository_GetDescendants(t *testing.T) {
    // Test with real database (test DB)
    // Test recursive CTE
    // Test depth limit
    // Test duplicate prevention
}

func TestCompanyRepository_GetAll_ActiveOnly(t *testing.T) {
    // Test filtering
    // Test soft delete
}
```

#### 1.3 Helper Function Tests
**Location**: `internal/usecase/*_helper_test.go`

**Example**: `company_usecase_helper_test.go`
```go
func TestUpdateDescendantsLevel(t *testing.T) {
    // Test level update for all descendants
    // Test holding company protection
    // Test circular reference prevention
}
```

### 2. Integration Tests (25%)

#### 2.1 API Handler Tests
**Location**: `internal/delivery/http/*_test.go`

**Example**: `company_handler_test.go`
```go
func TestCompanyHandler_GetAllCompanies_RBAC(t *testing.T) {
    // Test superadmin sees all
    // Test admin sees only their company + descendants
    // Test access control
}

func TestCompanyHandler_UpdateCompanyFull_AdminBlockedFromHolding(t *testing.T) {
    // Test admin cannot update holding
    // Test superadmin can update holding
}

func TestCompanyHandler_GetAllCompanies_NoDuplicates(t *testing.T) {
    // Test no duplicate in response
    // Test deduplication works
}
```

### 3. E2E Tests (5%)

#### 3.1 Critical User Flows
**Location**: `test/integration/*_e2e_test.go`

**Example**: `company_e2e_test.go`
```go
func TestE2E_AdminEditSubsidiary_NoDuplicates(t *testing.T) {
    // Full flow:
    // 1. Login as admin
    // 2. Get companies list
    // 3. Edit subsidiary
    // 4. Verify no duplicates
    // 5. Verify level correct
}

func TestE2E_SuperadminEditHolding_Success(t *testing.T) {
    // Full flow:
    // 1. Login as superadmin
    // 2. Edit holding
    // 3. Verify saved
}
```

---

## 🎯 Test Coverage Goals

### Minimum Coverage
- **Critical Business Logic**: 90%+
  - Company hierarchy (level calculation, descendants, ancestors)
  - RBAC (access control, permissions)
  - Company CRUD operations

- **Repository Layer**: 80%+
  - Database queries
  - Data filtering
  - Error handling

- **UseCase Layer**: 85%+
  - Business logic
  - Validation
  - Error handling

- **Handler Layer**: 70%+
  - Request/response handling
  - Access control
  - Error responses

### Overall Coverage Target
- **Minimum**: 70%
- **Target**: 80%
- **Ideal**: 90%+

---

## ⚙️ Test Execution Strategy

### 1. Local Development
**When**: Before commit
**Command**: `make test` atau `go test ./... -v`
**Purpose**: Quick feedback saat development

**Script**: `Makefile`
```makefile
test:
	go test ./... -v -cover

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

test-unit:
	go test ./internal/usecase/... -v

test-integration:
	go test ./test/integration/... -v
```

### 2. Pre-Commit Hook (Optional)
**When**: Before git commit
**Tool**: Git hooks atau pre-commit framework
**Purpose**: Prevent committing broken code

**Script**: `.git/hooks/pre-commit`
```bash
#!/bin/bash
go test ./... -short
if [ $? -ne 0 ]; then
    echo "Tests failed. Commit aborted."
    exit 1
fi
```

### 3. CI/CD Pipeline
**When**: On every push to `development` or `main`
**Location**: `.github/workflows/ci-cd.yml`
**Purpose**: Automated testing before deployment

**Current Step** (line 105-107):
```yaml
- name: Test backend
  working-directory: backend
  run: go test ./... -v
```

**Enhanced Step** (proposed):
```yaml
- name: Test backend
  working-directory: backend
  run: |
    go test ./... -v -coverprofile=coverage.out
    go tool cover -func=coverage.out
    # Fail if coverage < 70%
    go tool cover -func=coverage.out | grep total | awk '{if ($3+0 < 70) exit 1}'

- name: Upload coverage to Codecov (optional)
  uses: codecov/codecov-action@v3
  with:
    file: ./backend/coverage.out
```

### 4. Nightly/Weekly Full Test Suite
**When**: Scheduled (e.g., every night at 2 AM)
**Purpose**: Run full test suite including slow integration tests
**Tool**: GitHub Actions scheduled workflow

---

## 📊 Test Data Management

### Test Fixtures
**Location**: `test/fixtures/`

**Structure**:
```json
// test/fixtures/companies.json
{
  "holding": {
    "id": "test-holding-id",
    "code": "PDV",
    "name": "Pedeve Pertamina",
    "level": 0,
    "parent_id": null
  },
  "child_companies": [
    {
      "id": "test-child-1",
      "code": "CHILD1",
      "name": "Child Company 1",
      "level": 1,
      "parent_id": "test-holding-id"
    }
  ]
}
```

### Test Database Setup
**Location**: `test/helpers/database.go`

**Features**:
- In-memory SQLite untuk fast unit tests
- PostgreSQL test container untuk integration tests
- Automatic migration
- Test data seeding
- Cleanup after tests

---

## 🚀 Implementation Phases

### Phase 1: Foundation (Week 1)
**Priority**: Critical
- [ ] Setup test infrastructure (helpers, fixtures)
- [ ] Create test database setup
- [ ] Write unit tests for `company_usecase.go` (critical functions)
  - Level calculation
  - Holding protection
  - Descendants calculation
- [ ] Write unit tests for `company_usecase_helper.go`
  - `updateDescendantsLevel`
- [ ] Target: 50% coverage for company usecase

### Phase 2: Repository Tests (Week 2)
**Priority**: High
- [ ] Write repository tests for `company_repository.go`
  - `GetDescendants` (recursive CTE)
  - `GetAll` (filtering, deduplication)
  - `GetByID`, `GetByCode`
- [ ] Write repository tests for `user_repository.go`
- [ ] Target: 70% coverage for repositories

### Phase 3: Integration Tests (Week 3)
**Priority**: High
- [ ] Write integration tests for `company_handler.go`
  - RBAC (superadmin vs admin)
  - Duplicate prevention
  - Holding protection
- [ ] Write integration tests for `user_management_handler.go`
- [ ] Target: 60% coverage for handlers

### Phase 4: E2E Tests (Week 4)
**Priority**: Medium
- [ ] Write E2E tests for critical flows
  - Admin edit subsidiary (no duplicates)
  - Superadmin edit holding (success)
  - Company hierarchy access
- [ ] Target: 5 critical E2E tests

### Phase 5: Coverage & Optimization (Week 5)
**Priority**: Medium
- [ ] Increase coverage to 80%+
- [ ] Optimize slow tests
- [ ] Add test documentation
- [ ] Setup coverage reporting

---

## 📝 Test Examples (Detailed)

### Example 1: Unit Test - Level Calculation
```go
// internal/usecase/company_usecase_test.go
func TestCompanyUseCase_UpdateCompanyFull_LevelCalculation(t *testing.T) {
    tests := []struct {
        name           string
        companyCode    string
        parentLevel    int
        expectedLevel  int
        shouldBlock    bool
    }{
        {
            name:          "Holding company must be level 0",
            companyCode:   "PDV",
            parentLevel:   -1, // NULL parent
            expectedLevel: 0,
            shouldBlock:   false,
        },
        {
            name:          "Child of holding is level 1",
            companyCode:   "CHILD1",
            parentLevel:   0,
            expectedLevel: 1,
            shouldBlock:   false,
        },
        {
            name:          "Admin cannot set parent_id for holding",
            companyCode:   "PDV",
            parentLevel:   1, // Trying to set parent
            expectedLevel: 0,
            shouldBlock:   true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup test
            // Execute
            // Assert
        })
    }
}
```

### Example 2: Integration Test - RBAC
```go
// internal/delivery/http/company_handler_test.go
func TestCompanyHandler_GetAllCompanies_RBAC(t *testing.T) {
    tests := []struct {
        name           string
        userRole       string
        userCompanyID  string
        expectedCount  int
        shouldContain  []string // Company IDs that should be in response
        shouldNotContain []string // Company IDs that should NOT be in response
    }{
        {
            name:          "Superadmin sees all companies",
            userRole:      "superadmin",
            userCompanyID: "",
            expectedCount: 10, // All companies in test DB
            shouldContain: []string{"holding-id", "child-1", "child-2"},
        },
        {
            name:          "Admin sees only their company + descendants",
            userRole:      "admin",
            userCompanyID: "holding-id",
            expectedCount: 5, // Holding + 4 descendants
            shouldContain: []string{"holding-id", "child-1"},
            shouldNotContain: []string{"other-company-id"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup test DB
            // Create test companies
            // Create test user with role
            // Make API request
            // Assert response
        })
    }
}
```

### Example 3: E2E Test - No Duplicates
```go
// test/integration/company_e2e_test.go
func TestE2E_AdminEditSubsidiary_NoDuplicates(t *testing.T) {
    // 1. Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    // 2. Create test data
    holding := createTestHolding(t, db)
    child := createTestChild(t, db, holding.ID)

    // 3. Create admin user
    admin := createTestAdmin(t, db, holding.ID)

    // 4. Login as admin
    token := loginAsUser(t, admin.Username, "password")

    // 5. Get companies list (before edit)
    companiesBefore := getCompanies(t, token)
    initialCount := len(companiesBefore)

    // 6. Edit subsidiary
    updateCompany(t, token, child.ID, map[string]interface{}{
        "name": "Updated Child Company",
    })

    // 7. Get companies list (after edit)
    companiesAfter := getCompanies(t, token)
    afterCount := len(companiesAfter)

    // 8. Assertions
    assert.Equal(t, initialCount, afterCount, "Company count should not change")
    
    // Check for duplicates
    companyIDs := make(map[string]int)
    for _, c := range companiesAfter {
        companyIDs[c.ID]++
    }
    for id, count := range companyIDs {
        assert.Equal(t, 1, count, "Company %s should appear only once, but appears %d times", id, count)
    }

    // 9. Verify level is correct
    updatedChild := findCompanyByID(t, companiesAfter, child.ID)
    assert.Equal(t, 1, updatedChild.Level, "Child company level should be 1")
}
```

---

## 🔍 Test Quality Checklist

### Before Writing Tests
- [ ] Understand the business logic
- [ ] Identify edge cases
- [ ] Identify error cases
- [ ] Plan test data

### While Writing Tests
- [ ] Test happy path
- [ ] Test error cases
- [ ] Test edge cases
- [ ] Test boundary conditions
- [ ] Use table-driven tests where possible
- [ ] Keep tests isolated (no dependencies between tests)
- [ ] Use descriptive test names
- [ ] Add comments for complex test logic

### After Writing Tests
- [ ] Run tests locally
- [ ] Check coverage
- [ ] Review test readability
- [ ] Ensure tests are fast (< 1 second per test)
- [ ] Ensure tests are deterministic (same result every time)

---

## 📈 Metrics & Reporting

### Coverage Metrics
- **Line Coverage**: Percentage of lines executed
- **Branch Coverage**: Percentage of branches executed
- **Function Coverage**: Percentage of functions called

### Test Metrics
- **Test Count**: Total number of tests
- **Test Duration**: Total time to run all tests
- **Pass Rate**: Percentage of passing tests
- **Flaky Tests**: Tests that sometimes pass, sometimes fail

### Reporting Tools
- **Built-in**: `go test -cover`
- **HTML Report**: `go tool cover -html=coverage.out`
- **External**: Codecov, Coveralls (optional)

---

## 🎓 Best Practices

### 1. Test Organization
- Group related tests together
- Use test tables for similar test cases
- Use subtests (`t.Run()`) for organization

### 2. Test Data
- Use fixtures for complex data
- Create helper functions for common setup
- Clean up test data after tests

### 3. Assertions
- Use `require` for critical assertions (stops test on failure)
- Use `assert` for non-critical assertions (continues test on failure)
- Provide clear error messages

### 4. Mocking
- Mock external dependencies (database, external APIs)
- Don't mock what you're testing
- Use interfaces for testability

### 5. Performance
- Keep unit tests fast (< 100ms each)
- Use test database for integration tests
- Run slow tests separately

---

## 🚨 Common Pitfalls to Avoid

1. **Testing Implementation Details**
   - ❌ Test internal function names
   - ✅ Test behavior and outcomes

2. **Over-Mocking**
   - ❌ Mock everything
   - ✅ Mock only external dependencies

3. **Flaky Tests**
   - ❌ Tests that depend on timing or random data
   - ✅ Deterministic tests with fixed data

4. **Slow Tests**
   - ❌ Tests that take minutes to run
   - ✅ Fast tests (< 1 second each)

5. **Test Duplication**
   - ❌ Copy-paste test code
   - ✅ Use helper functions and test tables

---

## 📚 Resources

### Documentation
- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Go Testing Best Practices](https://golang.org/doc/effective_go#testing)

### Tools
- [Go Test Coverage](https://go.dev/blog/cover)
- [Test Containers](https://golang.testcontainers.org/) (optional)

---

## ✅ Next Steps

1. **Review this document** with the team
2. **Approve the strategy** and timeline
3. **Start Phase 1** implementation
4. **Iterate** based on feedback

---

**Last Updated**: 2025-01-XX
**Version**: 1.0
**Status**: Planning Phase

