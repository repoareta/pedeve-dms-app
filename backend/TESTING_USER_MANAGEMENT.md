# Panduan Testing Manual - User Management Multi-Level

## Prerequisites

1. Backend server running: `make dev`
2. Frontend running (opsional, untuk test via UI)
3. Superadmin credentials:
   - Email: `superadmin@pertamina.com`
   - Username: `superadmin`
   - Password: `Pedeve123`

## Tools untuk Testing

### Opsi 1: Swagger UI (Recommended)
- URL: `http://localhost:8080/swagger/index.html`
- Keuntungan: UI friendly, auto-generated docs, bisa test langsung

### Opsi 2: Postman/Insomnia
- Import dari Swagger atau buat collection manual

### Opsi 3: cURL (Command Line)
- Untuk automation atau quick test

---

## Test Flow Overview

```
1. Login sebagai Superadmin
2. Get CSRF Token
3. Create Company (Root Level)
4. Create Sub-Company
5. Get All Companies
6. Get Company by ID
7. Get Company Children
8. Create Role (opsional, karena sudah ada default roles)
9. Get All Roles
10. Get Role Permissions
11. Create User dengan Company Assignment
12. Get All Users
13. Get User by ID
14. Update User
15. Test Access Control (login sebagai admin company)
16. Update Company
17. Delete User
18. Delete Company (soft delete)
```

---

## Detailed Test Cases

### 1. Authentication & Setup

#### Test 1.1: Login sebagai Superadmin
**Endpoint:** `POST /api/v1/auth/login`

**Request:**
```json
{
  "username": "superadmin@pertamina.com",
  "password": "Pedeve123"
}
```

**Expected Response:**
- Status: `200 OK`
- Body: `{ "token": "...", "user": { ... } }`
- Cookie: `auth_token` ter-set (httpOnly)

**Check:**
- ✅ Token diterima
- ✅ User info lengkap (id, username, email, role)
- ✅ Cookie ter-set di browser

---

#### Test 1.2: Get CSRF Token
**Endpoint:** `GET /api/v1/csrf-token`

**Expected Response:**
- Status: `200 OK`
- Body: `{ "csrf_token": "..." }`

**Check:**
- ✅ CSRF token diterima
- ✅ Simpan token untuk request berikutnya

---

### 2. Company Management

#### Test 2.1: Create Root Company (Holding)
**Endpoint:** `POST /api/v1/companies`

**Headers:**
```
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
Content-Type: application/json
```

**Request:**
```json
{
  "name": "PT Holding Pertamina",
  "code": "HOLDING-001",
  "description": "Main holding company"
}
```

**Expected Response:**
- Status: `201 Created`
- Body: Company object dengan:
  - `id`: UUID
  - `name`: "PT Holding Pertamina"
  - `code`: "HOLDING-001"
  - `level`: 0
  - `parent_id`: null

**Check:**
- ✅ Company created
- ✅ Level = 0 (root)
- ✅ Parent ID = null
- ✅ Simpan `id` untuk test berikutnya

---

#### Test 2.2: Create Sub-Company
**Endpoint:** `POST /api/v1/companies`

**Request:**
```json
{
  "name": "PT Subsidiary 1",
  "code": "SUB-001",
  "description": "First subsidiary",
  "parent_id": "<company_id_dari_test_2.1>"
}
```

**Expected Response:**
- Status: `201 Created`
- Body: Company object dengan:
  - `level`: 1
  - `parent_id`: ID dari company parent

**Check:**
- ✅ Sub-company created
- ✅ Level = 1
- ✅ Parent ID sesuai
- ✅ Simpan `id` untuk test berikutnya

---

#### Test 2.3: Create Sub-Sub-Company (Level 2)
**Endpoint:** `POST /api/v1/companies`

**Request:**
```json
{
  "name": "PT Sub-Subsidiary 1",
  "code": "SUBSUB-001",
  "description": "Sub-subsidiary",
  "parent_id": "<sub_company_id_dari_test_2.2>"
}
```

**Expected Response:**
- Status: `201 Created`
- Body: Company dengan `level`: 2

**Check:**
- ✅ Hierarchy infinite nesting bekerja
- ✅ Level = 2

---

#### Test 2.4: Get All Companies
**Endpoint:** `GET /api/v1/companies`

**Headers:**
```
Authorization: Bearer <token>
```

**Expected Response:**
- Status: `200 OK`
- Body: Array of companies

**Check:**
- ✅ Semua companies ter-list
- ✅ Superadmin melihat semua companies

---

#### Test 2.5: Get Company by ID
**Endpoint:** `GET /api/v1/companies/:id`

**Expected Response:**
- Status: `200 OK`
- Body: Company object lengkap

**Check:**
- ✅ Company detail lengkap
- ✅ Semua field ada

---

#### Test 2.6: Get Company Children
**Endpoint:** `GET /api/v1/companies/:id/children`

**Expected Response:**
- Status: `200 OK`
- Body: Array of child companies

**Check:**
- ✅ Hanya children yang ter-list
- ✅ Tidak termasuk grandchildren

---

#### Test 2.7: Update Company
**Endpoint:** `PUT /api/v1/companies/:id`

**Headers:**
```
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
Content-Type: application/json
```

**Request:**
```json
{
  "name": "PT Holding Pertamina Updated",
  "description": "Updated description"
}
```

**Expected Response:**
- Status: `200 OK`
- Body: Updated company object

**Check:**
- ✅ Name updated
- ✅ Description updated
- ✅ Other fields tidak berubah

---

#### Test 2.8: Delete Company (Soft Delete)
**Endpoint:** `DELETE /api/v1/companies/:id`

**Headers:**
```
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
```

**Expected Response:**
- Status: `200 OK`
- Body: `{ "message": "Company deleted successfully" }`

**Check:**
- ✅ Company `is_active` = false (cek di database)
- ✅ Company tidak muncul di GET /companies (jika filter active)

---

### 3. Role Management

#### Test 3.1: Get All Roles
**Endpoint:** `GET /api/v1/roles`

**Expected Response:**
- Status: `200 OK`
- Body: Array of roles

**Check:**
- ✅ Default roles ada: superadmin, admin, manager, staff
- ✅ Semua roles ter-list
- ✅ Simpan `id` role "admin" untuk test berikutnya

---

#### Test 3.2: Get Role by ID
**Endpoint:** `GET /api/v1/roles/:id`

**Expected Response:**
- Status: `200 OK`
- Body: Role object lengkap

**Check:**
- ✅ Role detail lengkap
- ✅ Field: id, name, description, level, is_system

---

#### Test 3.3: Get Role Permissions
**Endpoint:** `GET /api/v1/roles/:id/permissions`

**Expected Response:**
- Status: `200 OK`
- Body: Array of permissions

**Check:**
- ✅ Permissions ter-list
- ✅ Permissions sesuai dengan role (admin punya permissions tertentu)

---

#### Test 3.4: Create Custom Role (Opsional)
**Endpoint:** `POST /api/v1/roles`

**Request:**
```json
{
  "name": "custom_role",
  "description": "Custom role for testing",
  "level": 5
}
```

**Expected Response:**
- Status: `201 Created`
- Body: Role object

**Check:**
- ✅ Custom role created
- ✅ `is_system` = false

---

#### Test 3.5: Assign Permission to Role
**Endpoint:** `POST /api/v1/roles/:id/permissions`

**Request:**
```json
{
  "permission_id": "<permission_id>"
}
```

**Expected Response:**
- Status: `200 OK`
- Body: `{ "message": "Permission assigned successfully" }`

**Check:**
- ✅ Permission ter-assign
- ✅ Cek dengan GET /roles/:id/permissions

---

#### Test 3.6: Revoke Permission from Role
**Endpoint:** `DELETE /api/v1/roles/:id/permissions`

**Request:**
```json
{
  "permission_id": "<permission_id>"
}
```

**Expected Response:**
- Status: `200 OK`

**Check:**
- ✅ Permission ter-revoke
- ✅ Cek dengan GET /roles/:id/permissions

---

### 4. Permission Management

#### Test 4.1: Get All Permissions
**Endpoint:** `GET /api/v1/permissions`

**Expected Response:**
- Status: `200 OK`
- Body: Array of permissions

**Check:**
- ✅ Semua permissions ter-list
- ✅ Default permissions ada

---

#### Test 4.2: Get Permissions by Resource
**Endpoint:** `GET /api/v1/permissions?resource=user`

**Expected Response:**
- Status: `200 OK`
- Body: Array of permissions dengan resource = "user"

**Check:**
- ✅ Hanya permissions dengan resource "user"
- ✅ Filter bekerja

---

#### Test 4.3: Get Permissions by Scope
**Endpoint:** `GET /api/v1/permissions?scope=company`

**Expected Response:**
- Status: `200 OK`
- Body: Array of permissions dengan scope = "company"

**Check:**
- ✅ Hanya permissions dengan scope "company"
- ✅ Filter bekerja

---

### 5. User Management

#### Test 5.1: Get All Roles (untuk mendapatkan role_id)
**Endpoint:** `GET /api/v1/roles`

**Action:** Simpan `id` dari role "admin" untuk test berikutnya

---

#### Test 5.2: Create User dengan Company Assignment
**Endpoint:** `POST /api/v1/users`

**Request:**
```json
{
  "username": "admin1",
  "email": "admin1@example.com",
  "password": "Password123!",
  "company_id": "<company_id_dari_test_2.1>",
  "role_id": "<admin_role_id>"
}
```

**Expected Response:**
- Status: `201 Created`
- Body: User object

**Check:**
- ✅ User created
- ✅ Company ID ter-assign
- ✅ Role ID ter-assign
- ✅ Password ter-hash (tidak muncul di response)
- ✅ Simpan `id` untuk test berikutnya

---

#### Test 5.3: Create User tanpa Company (harus error untuk non-superadmin)
**Endpoint:** `POST /api/v1/users`

**Request:**
```json
{
  "username": "user_no_company",
  "email": "user_no_company@example.com",
  "password": "Password123!",
  "role_id": "<staff_role_id>"
}
```

**Expected Response:**
- Status: `201 Created` (superadmin bisa)
- Atau `400 Bad Request` jika validation error

**Check:**
- ✅ Superadmin bisa create user tanpa company
- ✅ User lain harus specify company

---

#### Test 5.4: Get All Users
**Endpoint:** `GET /api/v1/users`

**Expected Response:**
- Status: `200 OK`
- Body: Array of users

**Check:**
- ✅ Semua users ter-list
- ✅ Superadmin melihat semua users
- ✅ Password tidak muncul di response

---

#### Test 5.5: Get User by ID
**Endpoint:** `GET /api/v1/users/:id`

**Expected Response:**
- Status: `200 OK`
- Body: User object lengkap

**Check:**
- ✅ User detail lengkap
- ✅ Company info (jika ada)
- ✅ Role info (jika ada)

---

#### Test 5.6: Update User
**Endpoint:** `PUT /api/v1/users/:id`

**Request:**
```json
{
  "username": "admin1_updated",
  "email": "admin1_updated@example.com",
  "company_id": "<company_id>",
  "role_id": "<role_id>"
}
```

**Expected Response:**
- Status: `200 OK`
- Body: Updated user object

**Check:**
- ✅ Username updated
- ✅ Email updated
- ✅ Company ID updated (jika diubah)
- ✅ Role ID updated (jika diubah)

---

#### Test 5.7: Delete User
**Endpoint:** `DELETE /api/v1/users/:id`

**Expected Response:**
- Status: `200 OK`
- Body: `{ "message": "User deleted successfully" }`

**Check:**
- ✅ User ter-delete dari database
- ✅ User tidak muncul di GET /users

---

### 6. Access Control Testing

#### Test 6.1: Login sebagai Admin Company
**Endpoint:** `POST /api/v1/auth/login`

**Request:**
```json
{
  "username": "admin1@example.com",
  "password": "Password123!"
}
```

**Expected Response:**
- Status: `200 OK`
- Body: Token + User info

**Check:**
- ✅ Login berhasil
- ✅ JWT claims berisi company_id
- ✅ JWT claims berisi role
- ✅ JWT claims berisi permissions

---

#### Test 6.2: Coba Akses Company Lain (harus 403)
**Endpoint:** `GET /api/v1/companies/:other_company_id`

**Headers:**
```
Authorization: Bearer <admin_company_token>
```

**Expected Response:**
- Status: `403 Forbidden`
- Body: `{ "error": "forbidden", "message": "You don't have access to this company" }`

**Check:**
- ✅ Access control bekerja
- ✅ Admin tidak bisa akses company lain

---

#### Test 6.3: Coba Akses Sub-Company (harus allowed)
**Endpoint:** `GET /api/v1/companies/:sub_company_id`

**Headers:**
```
Authorization: Bearer <admin_company_token>
```

**Expected Response:**
- Status: `200 OK` (jika sub-company adalah descendant)
- Atau `403 Forbidden` (jika bukan descendant)

**Check:**
- ✅ Admin bisa akses sub-company mereka
- ✅ Hierarchy access control bekerja

---

#### Test 6.4: Coba Create User di Company Lain (harus 403)
**Endpoint:** `POST /api/v1/users`

**Request:**
```json
{
  "username": "user_other_company",
  "email": "user_other@example.com",
  "password": "Password123!",
  "company_id": "<other_company_id>",
  "role_id": "<role_id>"
}
```

**Expected Response:**
- Status: `403 Forbidden`

**Check:**
- ✅ Admin tidak bisa create user di company lain
- ✅ Access control bekerja

---

### 7. Edge Cases & Error Handling

#### Test 7.1: Create Company dengan Code Duplicate
**Endpoint:** `POST /api/v1/companies`

**Request:**
```json
{
  "name": "Duplicate Company",
  "code": "HOLDING-001",
  "description": "Duplicate code"
}
```

**Expected Response:**
- Status: `400 Bad Request`
- Body: `{ "error": "creation_failed", "message": "company code already exists" }`

**Check:**
- ✅ Validation bekerja
- ✅ Error message jelas

---

#### Test 7.2: Create User dengan Username Duplicate
**Endpoint:** `POST /api/v1/users`

**Request:**
```json
{
  "username": "admin1",
  "email": "different@example.com",
  "password": "Password123!"
}
```

**Expected Response:**
- Status: `400 Bad Request`
- Body: `{ "error": "creation_failed", "message": "username already exists" }`

**Check:**
- ✅ Validation bekerja

---

#### Test 7.3: Create User dengan Email Duplicate
**Endpoint:** `POST /api/v1/users`

**Request:**
```json
{
  "username": "different_user",
  "email": "admin1@example.com",
  "password": "Password123!"
}
```

**Expected Response:**
- Status: `400 Bad Request`
- Body: `{ "error": "creation_failed", "message": "email already exists" }`

**Check:**
- ✅ Validation bekerja

---

#### Test 7.4: Update System Role (harus error)
**Endpoint:** `PUT /api/v1/roles/:superadmin_role_id`

**Request:**
```json
{
  "name": "superadmin_modified",
  "description": "Modified"
}
```

**Expected Response:**
- Status: `400 Bad Request`
- Body: `{ "error": "update_failed", "message": "system roles cannot be modified" }`

**Check:**
- ✅ System role protection bekerja

---

#### Test 7.5: Delete System Role (harus error)
**Endpoint:** `DELETE /api/v1/roles/:superadmin_role_id`

**Expected Response:**
- Status: `400 Bad Request`
- Body: `{ "error": "delete_failed", "message": "..." }`

**Check:**
- ✅ System role protection bekerja

---

## Checklist Testing

### Company Management
- [ ] Create root company
- [ ] Create sub-company
- [ ] Create sub-sub-company (level 2+)
- [ ] Get all companies
- [ ] Get company by ID
- [ ] Get company children
- [ ] Update company
- [ ] Delete company (soft delete)
- [ ] Access control (admin hanya bisa akses company mereka)

### Role Management
- [ ] Get all roles
- [ ] Get role by ID
- [ ] Get role permissions
- [ ] Create custom role
- [ ] Update role
- [ ] Delete role
- [ ] Assign permission to role
- [ ] Revoke permission from role
- [ ] System role protection

### Permission Management
- [ ] Get all permissions
- [ ] Get permission by ID
- [ ] Get permissions by resource
- [ ] Get permissions by scope
- [ ] Create permission
- [ ] Update permission
- [ ] Delete permission

### User Management
- [ ] Create user dengan company
- [ ] Create user tanpa company (superadmin only)
- [ ] Get all users
- [ ] Get user by ID
- [ ] Update user
- [ ] Delete user
- [ ] Access control (admin hanya bisa manage users di company mereka)

### Access Control
- [ ] Login sebagai admin company
- [ ] Coba akses company lain (403)
- [ ] Coba akses sub-company (200 jika descendant)
- [ ] Coba create user di company lain (403)
- [ ] JWT claims lengkap (company_id, role, permissions)

### Error Handling
- [ ] Duplicate company code
- [ ] Duplicate username
- [ ] Duplicate email
- [ ] System role modification
- [ ] System role deletion
- [ ] Invalid company ID
- [ ] Invalid role ID

---

## Tips Testing

1. **Gunakan Swagger UI** untuk kemudahan testing
2. **Simpan IDs** dari response untuk test berikutnya
3. **Test access control** dengan login sebagai user berbeda
4. **Cek database** untuk memastikan data tersimpan dengan benar
5. **Cek audit logs** untuk memastikan semua action tercatat
6. **Test hierarchy** dengan membuat company level 3+ untuk memastikan infinite nesting bekerja

---

## Expected Database State

Setelah testing lengkap, database harus berisi:

1. **Companies:**
   - 1 root company (level 0)
   - 1+ sub-companies (level 1)
   - 1+ sub-sub-companies (level 2+)

2. **Users:**
   - 1 superadmin (company_id = null)
   - 1+ admin users (dengan company_id)
   - 1+ regular users

3. **Roles:**
   - 4 system roles (superadmin, admin, manager, staff)
   - 0+ custom roles

4. **Permissions:**
   - Default permissions untuk setiap resource
   - Permissions ter-assign ke roles

---

## Troubleshooting

### Issue: Cannot login
- Cek email superadmin: `superadmin@pertamina.com`
- Cek password: `Pedeve123`
- Cek apakah user ada di database

### Issue: 403 Forbidden
- Cek JWT token valid
- Cek company_id di JWT claims
- Cek apakah user punya akses ke resource

### Issue: 400 Bad Request
- Cek request body format
- Cek required fields
- Cek validation rules

### Issue: 404 Not Found
- Cek ID yang digunakan
- Cek apakah resource ada di database

---

## Next Steps Setelah Testing

Jika semua test passed:
1. ✅ Modul User Management siap untuk production
2. ✅ Lanjutkan ke modul berikutnya (Dashboard, DMS, dll)
3. ✅ Buat frontend UI untuk User Management

Jika ada issues:
1. ✅ Fix bugs yang ditemukan
2. ✅ Re-test setelah fix
3. ✅ Update dokumentasi jika perlu

