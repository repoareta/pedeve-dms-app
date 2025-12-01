# Refactoring untuk Dependency Injection - Detail

## 📋 Overview

Untuk membuat test bisa dijalankan dengan test database, kita perlu refactor repository dan usecase untuk support dependency injection. Saat ini, repository menggunakan `database.GetDB()` yang global, sehingga tidak bisa di-inject test database.

---

## 🎯 Tujuan Refactoring

**Before (Current)**:
```go
// Repository menggunakan global DB
func NewCompanyRepository() CompanyRepository {
    return &companyRepository{
        db: database.GetDB(), // ❌ Global, tidak bisa di-inject
    }
}
```

**After (Target)**:
```go
// Repository accept DB sebagai parameter
func NewCompanyRepository(db *gorm.DB) CompanyRepository {
    return &companyRepository{
        db: db, // ✅ Injected, bisa pakai test DB
    }
}

// Keep backward compatibility
func NewCompanyRepository() CompanyRepository {
    return NewCompanyRepositoryWithDB(database.GetDB())
}
```

---

## 📁 Files yang Perlu Di-Refactor

### 1. Repository Layer (9 files)

#### 1.1 Company Repository
**File**: `backend/internal/repository/company_repository.go`

**Current**:
```go
func NewCompanyRepository() CompanyRepository {
    return &companyRepository{
        db: database.GetDB(),
    }
}
```

**After**:
```go
// NewCompanyRepositoryWithDB creates repository with injected DB (for testing)
func NewCompanyRepositoryWithDB(db *gorm.DB) CompanyRepository {
    return &companyRepository{
        db: db,
    }
}

// NewCompanyRepository creates repository with default DB (backward compatibility)
func NewCompanyRepository() CompanyRepository {
    return NewCompanyRepositoryWithDB(database.GetDB())
}
```

**Impact**: ✅ Low (backward compatible)

---

#### 1.2 User Repository
**File**: `backend/internal/repository/user_repository.go`

**Current**:
```go
func NewUserRepository() UserRepository {
    return &userRepository{
        db: database.GetDB(),
    }
}
```

**After**: Same pattern as Company Repository

**Impact**: ✅ Low

---

#### 1.3 Shareholder Repository
**File**: `backend/internal/repository/shareholder_repository.go`

**Current**:
```go
func NewShareholderRepository() ShareholderRepository {
    return &shareholderRepository{
        db: database.GetDB(),
    }
}
```

**After**: Same pattern

**Impact**: ✅ Low

---

#### 1.4 Business Field Repository
**File**: `backend/internal/repository/business_field_repository.go`

**Current**:
```go
func NewBusinessFieldRepository() BusinessFieldRepository {
    return &businessFieldRepository{
        db: database.GetDB(),
    }
}
```

**After**: Same pattern

**Impact**: ✅ Low

---

#### 1.5 Director Repository
**File**: `backend/internal/repository/director_repository.go`

**Current**:
```go
func NewDirectorRepository() DirectorRepository {
    return &directorRepository{
        db: database.GetDB(),
    }
}
```

**After**: Same pattern

**Impact**: ✅ Low

---

#### 1.6 Role Repository
**File**: `backend/internal/repository/role_repository.go`

**Current**:
```go
func NewRoleRepository() RoleRepository {
    return &roleRepository{
        db: database.GetDB(),
    }
}
```

**After**: Same pattern

**Impact**: ✅ Low

---

#### 1.7 Permission Repository
**File**: `backend/internal/repository/permission_repository.go`

**Current**:
```go
func NewPermissionRepository() PermissionRepository {
    return &permissionRepository{
        db: database.GetDB(),
    }
}
```

**After**: Same pattern

**Impact**: ✅ Low

---

#### 1.8 User Company Assignment Repository
**File**: `backend/internal/repository/user_company_assignment_repository.go`

**Current**: (Need to check)

**After**: Same pattern

**Impact**: ✅ Low

---

#### 1.9 Audit Repository
**File**: `backend/internal/repository/audit_repository.go`

**Current**:
```go
func NewAuditLogger(db *gorm.DB) *AuditLogger {
    return &AuditLogger{db: db}
}
```

**After**: ✅ Already supports dependency injection!

**Impact**: ✅ None (already good)

---

### 2. UseCase Layer (5 files)

#### 2.1 Company UseCase
**File**: `backend/internal/usecase/company_usecase.go`

**Current**:
```go
func NewCompanyUseCase() CompanyUseCase {
    return &companyUseCase{
        companyRepo:       repository.NewCompanyRepository(),
        shareholderRepo:   repository.NewShareholderRepository(),
        businessFieldRepo: repository.NewBusinessFieldRepository(),
        directorRepo:      repository.NewDirectorRepository(),
    }
}
```

**After**:
```go
// NewCompanyUseCaseWithDB creates usecase with injected DB (for testing)
func NewCompanyUseCaseWithDB(db *gorm.DB) CompanyUseCase {
    return &companyUseCase{
        companyRepo:       repository.NewCompanyRepositoryWithDB(db),
        shareholderRepo:   repository.NewShareholderRepositoryWithDB(db),
        businessFieldRepo: repository.NewBusinessFieldRepositoryWithDB(db),
        directorRepo:      repository.NewDirectorRepositoryWithDB(db),
    }
}

// NewCompanyUseCase creates usecase with default DB (backward compatibility)
func NewCompanyUseCase() CompanyUseCase {
    return NewCompanyUseCaseWithDB(database.GetDB())
}
```

**Impact**: ⚠️ Medium (perlu update semua repository calls)

---

#### 2.2 User Management UseCase
**File**: `backend/internal/usecase/user_management_usecase.go`

**Current**:
```go
func NewUserManagementUseCase() UserManagementUseCase {
    return &userManagementUseCase{
        userRepo:     repository.NewUserRepository(),
        companyRepo:  repository.NewCompanyRepository(),
        roleRepo:     repository.NewRoleRepository(),
        // ... more repos
    }
}
```

**After**: Same pattern as Company UseCase

**Impact**: ⚠️ Medium

---

#### 2.3 Role Management UseCase
**File**: `backend/internal/usecase/role_management_usecase.go`

**Current**: (Need to check)

**After**: Same pattern

**Impact**: ⚠️ Medium

---

#### 2.4 Development UseCase
**File**: `backend/internal/usecase/development_usecase.go`

**Current**: (Need to check)

**After**: Same pattern

**Impact**: ⚠️ Medium

---

#### 2.5 TwoFA UseCase
**File**: `backend/internal/usecase/twofa_usecase.go`

**Current**: (Need to check)

**After**: Same pattern

**Impact**: ⚠️ Medium

---

### 3. Handler Layer (✅ Already Good!)

**Status**: ✅ **TIDAK PERLU REFACTOR**

**Reason**: Handler sudah menggunakan dependency injection!

**Example**:
```go
// company_handler.go
func NewCompanyHandler(companyUseCase usecase.CompanyUseCase) *CompanyHandler {
    return &CompanyHandler{
        companyUseCase: companyUseCase, // ✅ Already injected
    }
}
```

**Impact**: ✅ None

---

### 4. Main Application (1 file)

#### 4.1 Main.go / API Entry Point
**File**: `backend/cmd/api/main.go`

**Current**:
```go
// Handlers di-instantiate dengan usecase default
companyHandler := http.NewCompanyHandler(usecase.NewCompanyUseCase())
userHandler := http.NewUserManagementHandler(usecase.NewUserManagementUseCase())
```

**After**:
```go
// Handlers tetap sama (sudah menggunakan dependency injection)
// Tidak perlu perubahan karena NewCompanyUseCase() tetap bekerja
companyHandler := http.NewCompanyHandler(usecase.NewCompanyUseCase())
userHandler := http.NewUserManagementHandler(usecase.NewUserManagementUseCase())
```

**Impact**: ✅ None (backward compatible)

---

### 5. Helper Functions (1 file)

#### 5.1 Auth Helper
**File**: `backend/internal/usecase/auth_helper.go`

**Current**:
```go
func GetUserAuthInfo(userID string) (*string, string, *string, int, string, []string, error) {
    userRepo := repository.NewUserRepository()
    assignmentRepo := repository.NewUserCompanyAssignmentRepository()
    roleRepo := repository.NewRoleRepository()
    companyRepo := repository.NewCompanyRepository()
    // ...
}
```

**After**:
```go
// Option 1: Keep as is (use default DB)
// Option 2: Accept DB as parameter (if needed for testing)
func GetUserAuthInfoWithDB(userID string, db *gorm.DB) (*string, string, *string, int, string, []string, error) {
    userRepo := repository.NewUserRepositoryWithDB(db)
    // ...
}

// Keep backward compatibility
func GetUserAuthInfo(userID string) (*string, string, *string, int, string, []string, error) {
    return GetUserAuthInfoWithDB(userID, database.GetDB())
}
```

**Impact**: ⚠️ Medium (perlu update semua calls)

---

## 📊 Summary: Files to Refactor

| Layer | Files | Status | Impact |
|-------|-------|--------|--------|
| **Repository** | 9 files | ❌ Need refactor | ✅ Low (backward compatible) |
| **UseCase** | 5 files | ❌ Need refactor | ⚠️ Medium |
| **Handler** | 0 files | ✅ Already good | ✅ None |
| **Main** | 1 file | ✅ No change needed | ✅ None |
| **Helper** | 1 file | ⚠️ Optional | ⚠️ Medium |

**Total Files to Refactor**: ~15 files

---

## 🔧 Refactoring Pattern

### Pattern untuk Repository

```go
// ✅ NEW: With DB injection (for testing)
func NewXxxRepositoryWithDB(db *gorm.DB) XxxRepository {
    return &xxxRepository{
        db: db,
    }
}

// ✅ KEEP: Default (backward compatibility)
func NewXxxRepository() XxxRepository {
    return NewXxxRepositoryWithDB(database.GetDB())
}
```

### Pattern untuk UseCase

```go
// ✅ NEW: With DB injection (for testing)
func NewXxxUseCaseWithDB(db *gorm.DB) XxxUseCase {
    return &xxxUseCase{
        repo1: repository.NewRepo1WithDB(db),
        repo2: repository.NewRepo2WithDB(db),
        // ...
    }
}

// ✅ KEEP: Default (backward compatibility)
func NewXxxUseCase() XxxUseCase {
    return NewXxxUseCaseWithDB(database.GetDB())
}
```

---

## ⚠️ Breaking Changes

### ✅ NO BREAKING CHANGES!

**Reason**: 
- Semua fungsi lama tetap ada (backward compatible)
- Fungsi baru hanya untuk testing
- Production code tidak perlu diubah

**Example**:
```go
// Production code tetap bekerja
repo := repository.NewCompanyRepository() // ✅ Works

// Test code bisa pakai test DB
testDB := setupTestDB(t)
repo := repository.NewCompanyRepositoryWithDB(testDB) // ✅ Works for testing
```

---

## 📝 Step-by-Step Refactoring Plan

### Step 1: Repository Layer (Week 1, Day 1-2)
1. Refactor `company_repository.go`
2. Refactor `user_repository.go`
3. Refactor `shareholder_repository.go`
4. Refactor `business_field_repository.go`
5. Refactor `director_repository.go`
6. Refactor `role_repository.go`
7. Refactor `permission_repository.go`
8. Refactor `user_company_assignment_repository.go`
9. Test: Verify semua repository masih bekerja

### Step 2: UseCase Layer (Week 1, Day 3-4)
1. Refactor `company_usecase.go`
2. Refactor `user_management_usecase.go`
3. Refactor `role_management_usecase.go`
4. Refactor `development_usecase.go`
5. Refactor `twofa_usecase.go`
6. Test: Verify semua usecase masih bekerja

### Step 3: Helper Functions (Week 1, Day 5)
1. Refactor `auth_helper.go` (optional)
2. Test: Verify helper functions masih bekerja

### Step 4: Update Tests (Week 2)
1. Update `company_usecase_test.go` untuk pakai test DB
2. Create tests untuk repository
3. Create integration tests
4. Verify semua tests pass

---

## 🧪 Testing After Refactoring

### Test Production Code
```bash
# Verify production code masih bekerja
go build ./cmd/api
# Run application
# Test manual via Swagger/Postman
```

### Test with Test Database
```go
// Test bisa pakai test DB
func TestCompanyUseCase_UpdateCompanyFull(t *testing.T) {
    testDB := helpers.SetupTestDB(t)
    defer helpers.CleanupTestDB(t, testDB)
    
    uc := usecase.NewCompanyUseCaseWithDB(testDB)
    // ... test code
}
```

---

## 📈 Benefits After Refactoring

### ✅ Testability
- Bisa inject test database
- Tests isolated dan fast
- No dependency on production DB

### ✅ Maintainability
- Clear dependency flow
- Easy to mock (if needed)
- Better architecture

### ✅ Backward Compatibility
- Production code tidak perlu diubah
- No breaking changes
- Safe to deploy

---

## ⏱️ Estimated Time

| Phase | Time | Complexity |
|-------|------|------------|
| Repository Layer | 1-2 days | Low |
| UseCase Layer | 1-2 days | Medium |
| Helper Functions | 0.5 day | Medium |
| Update Tests | 1-2 days | Low |
| **Total** | **3-6 days** | Medium |

---

## 🎯 Decision Point

**Pilih salah satu**:

### Option A: Full Refactoring (3-6 days)
- ✅ Proper dependency injection
- ✅ Best for long-term
- ✅ Easy to test
- ❌ Time consuming

### Option B: Integration Tests Only (1 day)
- ✅ Fast to implement
- ✅ Test real database logic
- ✅ No refactoring needed
- ❌ Slower tests
- ❌ Need test database setup

### Option C: Hybrid (Recommended)
- ✅ Start with integration tests (fast)
- ✅ Refactor incrementally (long-term)
- ✅ Best of both worlds

---

## ✅ Checklist

### Before Refactoring
- [ ] Backup current code
- [ ] Create feature branch
- [ ] Review all repository files
- [ ] Review all usecase files
- [ ] Plan test strategy

### During Refactoring
- [ ] Refactor one repository at a time
- [ ] Test after each refactor
- [ ] Keep backward compatibility
- [ ] Update tests incrementally

### After Refactoring
- [ ] All tests pass
- [ ] Production code works
- [ ] Test coverage improved
- [ ] Documentation updated

---

**Last Updated**: 2025-01-XX
**Status**: Planning Phase

