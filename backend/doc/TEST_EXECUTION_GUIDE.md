# Test Execution Guide - Kapan Test Berjalan?

## 📋 Overview

Dokumen ini menjelaskan **kapan** dan **bagaimana** automated tests berjalan, baik secara **otomatis** maupun **manual**.

---

## 🤖 Automated Test Execution (CI/CD)

### ✅ **YA, Test Berjalan Otomatis!**

**Trigger**: Setiap kali push ke branch `development` atau `main`

**Kapan Berjalan**:
- ✅ **Otomatis** saat push ke `development` branch
- ✅ **Otomatis** saat push ke `main` branch
- ✅ **Otomatis** saat create tag version (v*.*.*)
- ✅ **Manual trigger** via GitHub Actions UI (workflow_dispatch)

**Lokasi**: `.github/workflows/ci-cd.yml` (line 105-117)

**Command yang Dijalankan**:
```bash
go test ./... -v -coverprofile=coverage.out
```

**Hasil**:
- ✅ Jika test **PASS** → Deployment lanjut
- ❌ Jika test **FAIL** → **Deployment DIBLOKIR**, developer harus fix

---

## 📊 CI/CD Pipeline Flow

```
Push ke development/main
    ↓
[1] Checkout code
    ↓
[2] Setup Go environment
    ↓
[3] Download dependencies
    ↓
[4] Build packages
    ↓
[5] Lint backend (golangci-lint)
    ↓
[6] 🧪 TEST BACKEND (AUTOMATED) ← INI YANG PENTING!
    ├─ Run: go test ./... -v -coverprofile=coverage.out
    ├─ Generate coverage report
    └─ ✅ PASS → Continue
    └─ ❌ FAIL → STOP, block deployment
    ↓
[7] Build Docker image
    ↓
[8] Security scan (Trivy)
    ↓
[9] Push to GHCR
    ↓
[10] Deploy to GCP (jika branch = development)
```

---

## 🔍 Detail Test Step di CI/CD

**File**: `.github/workflows/ci-cd.yml` (line 105-117)

```yaml
- name: Test backend
  working-directory: backend
  run: |
    echo "🧪 Running backend tests..."
    go test ./... -v -coverprofile=coverage.out
    echo "📊 Test coverage:"
    go tool cover -func=coverage.out | grep total || echo "No coverage data"
    # Fail if critical tests fail
    if [ $? -ne 0 ]; then
      echo "❌ Tests failed!"
      exit 1
    fi
    echo "✅ All tests passed!"
```

**Yang Dilakukan**:
1. ✅ Run semua tests (`go test ./...`)
2. ✅ Generate coverage report (`coverage.out`)
3. ✅ Show coverage summary
4. ✅ **FAIL pipeline jika ada test yang fail**

---

## 🖐️ Manual Test Execution

### Kapan Perlu Jalankan Test Manual?

#### 1. **Sebelum Commit (Recommended)**
**Kapan**: Setelah selesai develop fitur baru atau fix bug

**Command**:
```bash
cd backend
make test
# atau
go test ./internal/repository/... ./internal/usecase/... -v
```

**Tujuan**: 
- Memastikan code yang akan di-commit tidak break existing tests
- Fast feedback sebelum push

---

#### 2. **Sebelum Push ke Development**
**Kapan**: Sebelum push ke branch `development`

**Command**:
```bash
cd backend
make test-coverage
# atau
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

**Tujuan**:
- Memastikan semua tests pass
- Check coverage tidak turun
- Prevent CI/CD failure

---

#### 3. **Saat Development Fitur Baru**
**Kapan**: Saat sedang develop fitur baru

**Command**:
```bash
# Run specific test
go test ./internal/usecase/... -run TestCompanyUseCase_CreateCompany -v

# Run specific package
go test ./internal/repository/company_repository_test.go ./internal/repository/company_repository.go -v
```

**Tujuan**:
- Quick feedback saat development
- Test fitur yang sedang dikerjakan
- Debug test failures

---

#### 4. **Sebelum Deploy ke Production**
**Kapan**: Sebelum merge ke `main` atau create release tag

**Command**:
```bash
cd backend
./scripts/run-full-tests.sh
# atau
make test-coverage-html
```

**Tujuan**:
- Full test suite dengan coverage report
- Memastikan semua critical tests pass
- Generate HTML coverage report untuk review

---

#### 5. **Saat Fix Bug**
**Kapan**: Setelah fix bug, sebelum commit

**Command**:
```bash
# Run test yang related dengan bug
go test ./internal/usecase/... -run TestCompanyUseCase_UpdateCompanyFull -v
```

**Tujuan**:
- Verify bug fix bekerja
- Memastikan tidak ada regression

---

## 📋 Test Execution Summary

| Scenario | Execution | Trigger | Mandatory? |
|----------|-----------|---------|------------|
| **Push ke development** | ✅ **Otomatis** (CI/CD) | Git push | ✅ **YA** (block deployment jika fail) |
| **Push ke main** | ✅ **Otomatis** (CI/CD) | Git push | ✅ **YA** (block deployment jika fail) |
| **Create version tag** | ✅ **Otomatis** (CI/CD) | Git tag | ✅ **YA** (block release jika fail) |
| **Sebelum commit** | 🖐️ Manual | Developer | ⚠️ Recommended (tidak mandatory) |
| **Saat development** | 🖐️ Manual | Developer | ⚠️ Recommended (tidak mandatory) |
| **Sebelum production** | 🖐️ Manual | Developer/QA | ✅ **YA** (best practice) |

---

## 🎯 Best Practices

### ✅ **DO (Lakukan)**

1. **Run test sebelum push ke development**
   ```bash
   make test
   ```

2. **Run test saat fix bug**
   ```bash
   go test ./... -run TestRelatedToBug -v
   ```

3. **Check coverage sebelum major release**
   ```bash
   make test-coverage-html
   ```

4. **Trust CI/CD sebagai safety net**
   - Jika lupa run test lokal, CI/CD akan catch
   - Tapi lebih baik run lokal dulu untuk fast feedback

---

### ❌ **DON'T (Jangan)**

1. **Jangan push tanpa test jika ada perubahan critical**
   - Level calculation
   - RBAC logic
   - Holding protection

2. **Jangan ignore test failures di CI/CD**
   - Fix dulu sebelum push lagi
   - Test failures berarti ada bug

3. **Jangan skip test untuk "save time"**
   - Test failures di production lebih mahal
   - Fix di development lebih cepat

---

## 🔄 Workflow Lengkap

### Scenario 1: Develop Fitur Baru

```
1. Developer create branch
   ↓
2. Develop fitur baru
   ↓
3. [OPTIONAL] Run test lokal: make test
   ↓
4. Commit & push ke development
   ↓
5. [AUTOMATIC] CI/CD run test
   ├─ ✅ PASS → Deploy to GCP
   └─ ❌ FAIL → Developer fix → push lagi
```

---

### Scenario 2: Fix Bug

```
1. Developer identify bug
   ↓
2. Fix bug
   ↓
3. [RECOMMENDED] Run test related: go test -run TestBugFix -v
   ↓
4. Commit & push ke development
   ↓
5. [AUTOMATIC] CI/CD run test
   ├─ ✅ PASS → Deploy to GCP
   └─ ❌ FAIL → Developer fix → push lagi
```

---

### Scenario 3: Deploy ke Production

```
1. Merge development → main
   ↓
2. [AUTOMATIC] CI/CD run test
   ├─ ✅ PASS → Continue
   └─ ❌ FAIL → Block release
   ↓
3. Create version tag (v1.0.0)
   ↓
4. [AUTOMATIC] CI/CD run test lagi
   ├─ ✅ PASS → Create GitHub Release
   └─ ❌ FAIL → Block release
```

---

## 🚨 What Happens If Test Fails?

### Di CI/CD (Otomatis)

**Jika test FAIL**:
1. ❌ **Deployment DIBLOKIR**
2. ❌ Pipeline berhenti (exit 1)
3. 📧 GitHub akan notify developer (jika configured)
4. 🔍 Developer harus:
   - Check test output di GitHub Actions
   - Fix bug
   - Push lagi
   - Test akan run otomatis lagi

**Contoh Output**:
```
🧪 Running backend tests...
=== RUN   TestCompanyUseCase_UpdateCompanyFull
--- FAIL: TestCompanyUseCase_UpdateCompanyFull (0.01s)
    Error: Expected level 0, got 1
❌ Tests failed!
Error: Process completed with exit code 1.
```

---

### Di Local (Manual)

**Jika test FAIL**:
1. ⚠️ Developer dapat feedback langsung
2. 🔍 Fix bug sebelum commit
3. ✅ Run test lagi sampai pass
4. ✅ Baru commit & push

---

## 📊 Test Coverage di CI/CD

**Current**: Coverage report di-generate tapi **tidak fail pipeline** jika coverage rendah

**Command**:
```bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total
```

**Output**:
```
total: (statements) 75.5%
```

**Note**: Coverage check adalah **optional** untuk sekarang. Bisa di-enforce di masa depan jika diperlukan.

---

## 🎯 Summary

### ✅ **Automated Tests (CI/CD)**

**Kapan**: 
- ✅ Setiap push ke `development`
- ✅ Setiap push ke `main`
- ✅ Setiap create version tag

**Trigger**: Git push/tag
**Mandatory**: ✅ **YA** (block deployment jika fail)

---

### 🖐️ **Manual Tests**

**Kapan**:
- ⚠️ Sebelum commit (recommended)
- ⚠️ Saat development (recommended)
- ✅ Sebelum production deploy (best practice)

**Trigger**: Developer decision
**Mandatory**: ⚠️ **Recommended** (tidak block, tapi best practice)

---

## 📝 Quick Reference

```bash
# Run semua tests (local)
make test

# Run dengan coverage (local)
make test-coverage

# Run specific test (local)
go test ./internal/usecase/... -run TestName -v

# CI/CD akan run otomatis saat push
# Tidak perlu command manual
```

---

**Last Updated**: 2025-01-XX
**Status**: ✅ **ACTIVE** - Tests run otomatis di CI/CD

