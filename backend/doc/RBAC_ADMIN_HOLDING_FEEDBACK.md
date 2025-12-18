# Feedback: RBAC untuk Admin Holding Company

## 📋 Statement dari User

> "Untuk user dengan role admin, itu artinya dia bisa mengedit data perusahaannya sendiri. Contohnya, admin.pedeve@pedeve.com yang di assign sebagai admin di perusahaan holding, dia harusnya boleh mengedit perusahaannya sendiri. Yang tidak boleh ada lah menghapus perusahaannya dia sendiri, dan mengganti role dia sendiri menjadi non admin.
>
> Intinya siapapun admin nya, selama dia di assign di perusahaan holding sebagai admin, maka privilege dia mirip seperti superadmin, tapi bedanya tidak bisa mengakses fitur2 terkait development, seperti reset dan seeder data. Tapi user dengan role admin yang khusus di assign di perusahaan holding, harusnya di izinkan melihat data log activity user saja, bukan log teknis."

---

## ✅ Feedback & Analisis

### 1. **Admin Holding Bisa Edit Perusahaan Sendiri** ✅ **SETUJU**

**Current State:**
- Saat ini ada block di `company_handler.go` line 237-250 yang mencegah admin mengupdate holding company
- Block ini terlalu ketat - admin holding seharusnya bisa edit perusahaan mereka sendiri

**Rekomendasi:**
```go
// BUKAN: Block semua admin dari update holding
if targetCompany.Code == "PDV" {
    return forbidden
}

// TAPI: Allow admin holding untuk edit holding mereka sendiri
if targetCompany.Code == "PDV" && userCompanyID != targetCompany.ID {
    return forbidden // Admin holding lain tidak bisa edit holding
}
```

**Logic yang Benar:**
- ✅ Admin holding bisa edit holding mereka sendiri (company_id mereka = holding ID)
- ❌ Admin holding tidak bisa edit holding lain
- ❌ Admin non-holding tidak bisa edit holding

---

### 2. **Admin Holding Tidak Boleh Hapus Perusahaan Sendiri** ✅ **SETUJU**

**Current State:**
- Perlu dicek apakah ada block untuk delete company

**Rekomendasi:**
```go
// Block admin dari delete company mereka sendiri
if roleName == "admin" && userCompanyID == companyID {
    return forbidden // Admin tidak bisa delete company mereka sendiri
}
```

**Logic yang Benar:**
- ❌ Admin holding tidak bisa delete holding (perusahaan mereka sendiri)
- ❌ Admin tidak bisa delete company mereka sendiri
- ✅ Superadmin bisa delete (dengan pertimbangan)

---

### 3. **Admin Holding Tidak Boleh Ganti Role Sendiri** ✅ **SETUJU**

**Current State:**
- Perlu dicek di user management handler

**Rekomendasi:**
```go
// Block admin dari update role mereka sendiri
if roleName == "admin" && userID == targetUserID {
    // Check if trying to change role to non-admin
    if newRole != "admin" {
        return forbidden
    }
}
```

**Logic yang Benar:**
- ❌ Admin tidak bisa ganti role mereka sendiri menjadi non-admin
- ✅ Admin bisa update role user lain (dalam scope mereka)
- ✅ Superadmin bisa update role siapa saja

---

### 4. **Admin Holding = Mirip Superadmin (Kecuali Development Features)** ✅ **SETUJU**

**Current State:**
- Development features sudah di-block untuk non-superadmin (line 55-64 di `development_handler.go`)
- Ini sudah benar ✅

**Rekomendasi:**
- **Tidak perlu perubahan** - sudah benar
- Development features (reset, seeder) hanya untuk superadmin

**Logic yang Benar:**
- ✅ Admin holding bisa:
  - Edit perusahaan mereka sendiri
  - Lihat semua companies (holding + descendants)
  - Manage users di company mereka + descendants
  - Lihat user activity logs
- ❌ Admin holding tidak bisa:
  - Akses development features (reset, seeder)
  - Lihat technical logs
  - Delete perusahaan mereka sendiri
  - Ganti role mereka sendiri

---

### 5. **Admin Holding Bisa Lihat User Activity Logs (Bukan Technical Logs)** ✅ **SETUJU**

**Current State:**
- Audit logs handler (line 81-162 di `audit_handler.go`) membedakan:
  - User reguler: hanya lihat logs sendiri
  - Admin/superadmin: lihat semua logs
- **TAPI**: Tidak ada filter untuk membedakan user activity logs vs technical logs

**Rekomendasi:**
```go
// Di GetAuditLogsHandler atau GetUserActivityLogsHandler
if roleName == "admin" {
    // Admin holding hanya bisa lihat user activity logs (permanent)
    // Block technical logs (log_type = "technical_error")
    if logType == "technical_error" {
        return forbidden // atau filter out technical logs
    }
    
    // Hanya return user activity logs (permanent resources)
    // atau gunakan endpoint /user-activity-logs saja
}
```

**Logic yang Benar:**
- ✅ Admin holding bisa lihat:
  - User activity logs (permanent: report, document, company, user)
  - Logs dari company mereka + descendants
- ❌ Admin holding tidak bisa lihat:
  - Technical error logs
  - System logs
  - Logs dari company lain (non-descendants)

---

## 🔧 Implementation Plan

### Step 1: Update Company Handler
**File**: `backend/internal/delivery/http/company_handler.go`

**Change**: Allow admin holding to edit their own holding
```go
// CRITICAL: For admin, prevent updating holding company (EXCEPT their own)
if roleName != "superadmin" && companyID != nil {
    // ... existing code ...
    
    targetCompany, err := h.companyUseCase.GetCompanyByID(id)
    if err == nil && targetCompany != nil && targetCompany.Code == "PDV" {
        // Allow admin holding to edit their own holding
        if userCompanyID != targetCompany.ID {
            // Admin holding lain tidak bisa edit holding
            return forbidden
        }
        // Admin holding bisa edit holding mereka sendiri - continue
    }
}
```

### Step 2: Block Admin from Deleting Own Company
**File**: `backend/internal/delivery/http/company_handler.go`

**Change**: Add check in DeleteCompany handler
```go
// Block admin from deleting their own company
if roleName == "admin" && userCompanyID == id {
    return forbidden
}
```

### Step 3: Block Admin from Changing Own Role
**File**: `backend/internal/delivery/http/user_management_handler.go`

**Change**: Add check in UpdateUser handler
```go
// Block admin from changing their own role to non-admin
if roleName == "admin" && userID == id {
    if req.RoleID != nil {
        // Check if new role is non-admin
        newRole, err := roleRepo.GetByID(*req.RoleID)
        if err == nil && newRole.Name != "admin" {
            return forbidden
        }
    }
}
```

### Step 4: Filter Technical Logs for Admin Holding
**File**: `backend/internal/delivery/http/audit_handler.go`

**Change**: Filter out technical logs for admin
```go
// Admin holding hanya bisa lihat user activity logs
if roleName == "admin" {
    // Filter out technical logs
    if logType == "technical_error" {
        logType = "" // Don't show technical logs
    }
    
    // Or redirect to user-activity-logs endpoint only
}
```

---

## 📊 Privilege Matrix

| Action | Superadmin | Admin Holding | Admin Non-Holding | Regular User |
|--------|-----------|--------------|------------------|--------------|
| **Edit Own Company** | ✅ | ✅ | ✅ | ❌ |
| **Edit Holding** | ✅ | ✅ (own only) | ❌ | ❌ |
| **Delete Own Company** | ✅ | ❌ | ❌ | ❌ |
| **Delete Holding** | ✅ | ❌ | ❌ | ❌ |
| **Change Own Role** | ✅ | ❌ | ❌ | ❌ |
| **View All Companies** | ✅ | ✅ (holding + descendants) | ✅ (own + descendants) | ❌ |
| **View User Activity Logs** | ✅ | ✅ | ✅ (own company only) | ❌ (own only) |
| **View Technical Logs** | ✅ | ❌ | ❌ | ❌ |
| **Development Features** | ✅ | ❌ | ❌ | ❌ |
| **Manage Users** | ✅ | ✅ (own + descendants) | ✅ (own + descendants) | ❌ |

---

## ✅ Kesimpulan

**Statement user BENAR dan LOGIS** ✅

**Yang Perlu Diimplementasikan:**
1. ✅ Allow admin holding edit holding mereka sendiri
2. ✅ Block admin dari delete company mereka sendiri
3. ✅ Block admin dari ganti role mereka sendiri
4. ✅ Filter technical logs untuk admin (hanya user activity logs)
5. ✅ Development features tetap hanya untuk superadmin (sudah benar)

**Urutan Prioritas:**
1. **HIGH**: Allow admin holding edit holding mereka sendiri (fix bug yang ada)
2. **HIGH**: Block admin dari delete company mereka sendiri
3. **MEDIUM**: Block admin dari ganti role mereka sendiri
4. **MEDIUM**: Filter technical logs untuk admin

---

**Status**: ✅ **APPROVED** - Siap untuk implementasi

