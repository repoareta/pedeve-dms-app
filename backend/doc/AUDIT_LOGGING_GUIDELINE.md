# Audit Logging Guideline

## Overview

Aplikasi ini sudah memiliki sistem audit trail yang lengkap untuk mencatat semua aktivitas user. Dokumen ini menjelaskan bagaimana menggunakan sistem audit logging yang ada untuk memastikan semua operasi penting tercatat dengan benar.

## Konsep Audit Trail

Audit trail adalah catatan permanen dari semua aktivitas yang dilakukan user dalam sistem. Setiap operasi penting (create, update, delete) harus dicatat dengan informasi:
- **Siapa**: User ID dan Username
- **Apa**: Action yang dilakukan
- **Objek apa**: Resource dan Resource ID
- **Dari mana**: IP Address dan User Agent
- **Kapan**: Timestamp otomatis
- **Status**: Success, Failure, atau Error
- **Detail tambahan**: JSON dengan informasi spesifik (opsional)

## Struktur Audit Log

```go
type AuditLog struct {
    ID         string    // UUID
    UserID     string    // ID user yang melakukan aksi
    Username   string    // Username untuk kemudahan query
    Action     string    // Jenis aksi (create, update, delete, dll)
    Resource   string    // Tipe resource (user, company, file, dll)
    ResourceID string   // ID dari resource yang dioperasikan
    IPAddress  string    // IP address user
    UserAgent  string    // Browser/client info
    Details    string    // JSON string untuk detail tambahan
    Status     string    // success, failure, error
    LogType    string    // user_action atau technical_error
    CreatedAt  time.Time // Timestamp otomatis
}
```

## Cara Menggunakan Audit Logging

### 1. Import Package

```go
import (
    "github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/audit"
)
```

### 2. Fungsi LogAction

Gunakan fungsi `audit.LogAction()` untuk mencatat aktivitas:

```go
audit.LogAction(
    userID,           // string: ID user yang melakukan aksi
    username,         // string: Username user
    action,           // string: Jenis aksi (gunakan constants dari audit package)
    resource,         // string: Tipe resource (gunakan constants dari audit package)
    resourceID,       // string: ID dari resource yang dioperasikan
    ipAddress,        // string: IP address user (gunakan getClientIP(c))
    userAgent,        // string: User agent (gunakan c.Get("User-Agent"))
    status,           // string: success, failure, atau error (gunakan constants)
    details,          // map[string]interface{}: Detail tambahan (opsional, bisa nil)
)
```

### 3. Helper Function untuk IP Address

Gunakan helper function `getClientIP()` untuk mendapatkan IP address:

```go
func getClientIP(c *fiber.Ctx) string {
    ip := c.IP()
    if forwarded := c.Get("X-Forwarded-For"); forwarded != "" {
        ip = forwarded
    }
    return ip
}
```

## Constants yang Tersedia

### Action Constants

#### Authentication Actions
- `audit.ActionLogin` - User login
- `audit.ActionLogout` - User logout
- `audit.ActionRegister` - User registration
- `audit.ActionFailedLogin` - Login gagal
- `audit.ActionPasswordReset` - Reset password

#### Generic CRUD Actions
- `audit.ActionCreate` - Create operation (generic)
- `audit.ActionUpdate` - Update operation (generic)
- `audit.ActionDelete` - Delete operation (generic)
- `audit.ActionView` - View operation (generic)

#### User Management Actions
- `audit.ActionCreateUser` - Create user
- `audit.ActionUpdateUser` - Update user
- `audit.ActionDeleteUser` - Delete user

#### Company/Subsidiary Actions
- `audit.ActionCreateCompany` - Create company/subsidiary
- `audit.ActionUpdateCompany` - Update company/subsidiary
- `audit.ActionDeleteCompany` - Delete company/subsidiary

#### Document Actions
- `audit.ActionCreateDoc` - Create document
- `audit.ActionUpdateDoc` - Update document
- `audit.ActionDeleteDoc` - Delete document
- `audit.ActionViewDoc` - View document

#### File Management Actions (untuk modul File Management)
- `audit.ActionCreateFile` - Create file
- `audit.ActionUpdateFile` - Update file metadata
- `audit.ActionDeleteFile` - Delete file
- `audit.ActionDownloadFile` - Download file
- `audit.ActionViewFile` - View file
- `audit.ActionUploadFile` - Upload file

#### Report Management Actions (untuk modul Report Management)
- `audit.ActionGenerateReport` - Generate report
- `audit.ActionViewReport` - View report
- `audit.ActionExportReport` - Export report
- `audit.ActionDeleteReport` - Delete report

#### 2FA Actions
- `audit.ActionEnable2FA` - Enable 2FA
- `audit.ActionDisable2FA` - Disable 2FA

### Resource Constants

- `audit.ResourceUser` - User resource
- `audit.ResourceCompany` - Company/Subsidiary resource
- `audit.ResourceDocument` - Document resource
- `audit.ResourceAuth` - Authentication resource
- `audit.ResourceRole` - Role resource
- `audit.ResourcePermission` - Permission resource
- `audit.ResourceFile` - File resource (untuk modul File Management)
- `audit.ResourceReport` - Report resource (untuk modul Report Management)

### Status Constants

- `audit.StatusSuccess` - Operasi berhasil
- `audit.StatusFailure` - Operasi gagal (misalnya validasi error)
- `audit.StatusError` - Error sistem

## Contoh Implementasi

### Contoh 1: Create Company

```go
func (h *CompanyHandler) CreateCompany(c *fiber.Ctx) error {
    // ... validasi dan business logic ...
    
    company, err := h.companyUseCase.CreateCompany(req.Name, req.Code, req.Description, req.ParentID)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
            Error:   "creation_failed",
            Message: err.Error(),
        })
    }

    // Audit log untuk operasi berhasil
    userID := c.Locals("userID").(string)
    username := c.Locals("username").(string)
    audit.LogAction(
        userID, 
        username, 
        audit.ActionCreateCompany, 
        audit.ResourceCompany, 
        company.ID, 
        getClientIP(c), 
        c.Get("User-Agent"), 
        audit.StatusSuccess, 
        nil,
    )

    return c.Status(fiber.StatusCreated).JSON(company)
}
```

### Contoh 2: Update User dengan Detail

```go
func (h *UserManagementHandler) UpdateUser(c *fiber.Ctx) error {
    id := c.Params("id")
    // ... validasi dan business logic ...
    
    user, err := h.userUseCase.UpdateUser(id, req.Username, req.Email, req.CompanyID, req.RoleID)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
            Error:   "update_failed",
            Message: err.Error(),
        })
    }

    // Audit log dengan detail tambahan
    userID := c.Locals("userID").(string)
    username := c.Locals("username").(string)
    audit.LogAction(
        userID,
        username,
        audit.ActionUpdateUser,
        audit.ResourceUser,
        id,
        getClientIP(c),
        c.Get("User-Agent"),
        audit.StatusSuccess,
        map[string]interface{}{
            "updated_fields": []string{"username", "email", "company_id", "role_id"},
            "old_company_id": oldCompanyID,
            "new_company_id": req.CompanyID,
        },
    )

    return c.Status(fiber.StatusOK).JSON(user)
}
```

### Contoh 3: Delete dengan Error Handling

```go
func (h *CompanyHandler) DeleteCompany(c *fiber.Ctx) error {
    id := c.Params("id")
    userID := c.Locals("userID").(string)
    username := c.Locals("username").(string)
    
    // ... validasi access ...
    
    if err := h.companyUseCase.DeleteCompany(id); err != nil {
        // Audit log untuk operasi gagal
        audit.LogAction(
            userID,
            username,
            audit.ActionDeleteCompany,
            audit.ResourceCompany,
            id,
            getClientIP(c),
            c.Get("User-Agent"),
            audit.StatusFailure,
            map[string]interface{}{
                "error": err.Error(),
            },
        )
        
        return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
            Error:   "delete_failed",
            Message: err.Error(),
        })
    }

    // Audit log untuk operasi berhasil
    audit.LogAction(
        userID,
        username,
        audit.ActionDeleteCompany,
        audit.ResourceCompany,
        id,
        getClientIP(c),
        c.Get("User-Agent"),
        audit.StatusSuccess,
        nil,
    )

    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "message": "Company deleted successfully",
    })
}
```

### Contoh 4: File Management (untuk modul next)

```go
func (h *FileHandler) UploadFile(c *fiber.Ctx) error {
    // ... handle file upload ...
    
    file, err := h.fileUseCase.UploadFile(fileData, metadata)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
            Error:   "upload_failed",
            Message: err.Error(),
        })
    }

    // Audit log
    userID := c.Locals("userID").(string)
    username := c.Locals("username").(string)
    audit.LogAction(
        userID,
        username,
        audit.ActionUploadFile,
        audit.ResourceFile,
        file.ID,
        getClientIP(c),
        c.Get("User-Agent"),
        audit.StatusSuccess,
        map[string]interface{}{
            "filename": file.Filename,
            "file_size": file.Size,
            "file_type": file.Type,
        },
    )

    return c.Status(fiber.StatusCreated).JSON(file)
}
```

### Contoh 5: Report Management (untuk modul next)

```go
func (h *ReportHandler) GenerateReport(c *fiber.Ctx) error {
    // ... generate report ...
    
    report, err := h.reportUseCase.GenerateReport(req.ReportType, req.DateRange, req.Filters)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
            Error:   "generation_failed",
            Message: err.Error(),
        })
    }

    // Audit log
    userID := c.Locals("userID").(string)
    username := c.Locals("username").(string)
    audit.LogAction(
        userID,
        username,
        audit.ActionGenerateReport,
        audit.ResourceReport,
        report.ID,
        getClientIP(c),
        c.Get("User-Agent"),
        audit.StatusSuccess,
        map[string]interface{}{
            "report_type": req.ReportType,
            "date_range": req.DateRange,
            "filters": req.Filters,
        },
    )

    return c.Status(fiber.StatusOK).JSON(report)
}
```

## Best Practices

### 1. Selalu Log Operasi CRUD
Setiap operasi Create, Update, dan Delete **WAJIB** dicatat dalam audit log.

### 2. Log Setelah Operasi Berhasil
Log audit setelah operasi berhasil dilakukan, bukan sebelum. Ini memastikan kita hanya log operasi yang benar-benar terjadi.

### 3. Log Operasi Gagal (Opsional tapi Disarankan)
Untuk operasi yang gagal karena validasi atau business logic error, pertimbangkan untuk log dengan status `audit.StatusFailure` untuk tracking.

### 4. Gunakan Constants
Selalu gunakan constants dari package `audit` untuk action dan resource. Jangan hardcode string.

### 5. Tambahkan Detail yang Relevan
Gunakan parameter `details` untuk mencatat informasi tambahan yang berguna untuk audit, seperti:
- Field yang diubah (untuk update)
- Nilai lama vs baru (untuk update)
- Metadata file (untuk file operations)
- Parameter report (untuk report generation)

### 6. Jangan Log Informasi Sensitif
Jangan masukkan password, token, atau data sensitif lainnya ke dalam audit log.

### 7. Async Logging
Fungsi `audit.LogAction()` sudah menggunakan async logging (goroutine), jadi tidak akan memperlambat response. Namun, tetap pastikan error handling yang baik.

## Kapan Harus Menambahkan Constants Baru?

Tambahkan constants baru jika:
1. Ada action atau resource baru yang belum ada di constants
2. Action/resource tersebut akan digunakan di multiple tempat
3. Action/resource tersebut penting untuk tracking dan filtering

Contoh: Jika ada modul baru "Notification", tambahkan:
```go
// Di audit.go
const (
    ActionCreateNotification = "create_notification"
    ActionReadNotification = "read_notification"
    // ...
)

const (
    ResourceNotification = "notification"
    // ...
)
```

## Query Audit Logs

Untuk melihat audit logs, gunakan endpoint:
- `GET /api/v1/audit-logs` - List audit logs dengan pagination dan filter
- `GET /api/v1/audit-logs/stats` - Statistik audit logs

Filter yang tersedia:
- `action` - Filter berdasarkan action
- `resource` - Filter berdasarkan resource
- `status` - Filter berdasarkan status
- `logType` - Filter berdasarkan tipe log (user_action atau technical_error)
- `page` - Nomor halaman
- `pageSize` - Jumlah item per halaman (maksimal 100)

## Performance Considerations

1. **Async Logging**: Audit logging dilakukan secara async, jadi tidak akan memperlambat response time.

2. **Indexing**: Database sudah memiliki index pada field-field penting untuk performa query yang cepat.

3. **Retention Policy**: 
   - User actions: 90 hari (default)
   - Technical errors: 30 hari (default)
   - Dapat dikonfigurasi via environment variables

4. **Storage Optimization**: 
   - Gunakan `details` dengan bijak, jangan masukkan data yang terlalu besar
   - Detail yang besar bisa di-compress atau disimpan di storage terpisah

## Checklist untuk Developer

Saat menambahkan handler baru atau modifikasi handler yang ada:

- [ ] Apakah operasi ini termasuk Create, Update, atau Delete?
- [ ] Jika ya, sudahkah ditambahkan audit log?
- [ ] Apakah menggunakan constants yang benar untuk action dan resource?
- [ ] Apakah status yang digunakan sesuai (success/failure/error)?
- [ ] Apakah ada detail tambahan yang perlu dicatat?
- [ ] Apakah IP address dan user agent sudah diambil dengan benar?
- [ ] Apakah audit log ditambahkan setelah operasi berhasil?

## Troubleshooting

### Audit log tidak muncul
1. Pastikan `audit.InitAuditLogger()` sudah dipanggil saat aplikasi start
2. Pastikan database migration sudah berjalan (tabel `audit_logs` sudah ada)
3. Cek error log untuk melihat apakah ada error saat logging

### Query audit log lambat
1. Pastikan index sudah dibuat dengan benar
2. Gunakan filter yang tepat untuk mengurangi jumlah data
3. Gunakan pagination dengan pageSize yang wajar (10-50)

## Referensi

- File constants: `backend/internal/infrastructure/audit/audit.go`
- Repository implementation: `backend/internal/repository/audit_repository.go`
- Handler implementation: `backend/internal/delivery/http/audit_handler.go`
- Dokumentasi optimasi: `backend/doc/AUDIT_LOG_OPTIMIZATION.md`

