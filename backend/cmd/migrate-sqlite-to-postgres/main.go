package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Get database URLs from environment or use defaults
	sqlitePath := os.Getenv("SQLITE_PATH")
	if sqlitePath == "" {
		sqlitePath = "dms.db"
	}

	postgresURL := os.Getenv("DATABASE_URL")
	if postgresURL == "" {
		postgresURL = "postgres://postgres:dms_password@localhost:5432/db_dms_pedeve?sslmode=disable"
	}

	fmt.Println("🚀 Starting SQLite to PostgreSQL Migration")
	fmt.Printf("📂 SQLite: %s\n", sqlitePath)
	fmt.Printf("🐘 PostgreSQL: %s\n", postgresURL)
	fmt.Println()

	// Connect to SQLite
	fmt.Println("📖 Connecting to SQLite...")
	sqliteDB, err := gorm.Open(sqlite.Open(sqlitePath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to SQLite: %v", err)
	}
	fmt.Println("✅ Connected to SQLite")

	// Connect to PostgreSQL
	fmt.Println("📖 Connecting to PostgreSQL...")
	postgresDB, err := gorm.Open(postgres.Open(postgresURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL: %v", err)
	}
	fmt.Println("✅ Connected to PostgreSQL")
	fmt.Println()

	// Get underlying sql.DB for raw queries
	sqliteSQL, _ := sqliteDB.DB()
	postgresSQL, _ := postgresDB.DB()

	// Migrate in order (respecting foreign keys)
	fmt.Println("🔄 Starting data migration...")
	fmt.Println()

	// 1. Roles (no dependencies)
	fmt.Println("1️⃣  Migrating roles...")
	if err := migrateTable(sqliteSQL, postgresSQL, "roles", []string{"id", "name", "description", "level", "is_system", "created_at", "updated_at"}); err != nil {
		log.Fatalf("❌ Failed to migrate roles: %v", err)
	}
	count := getCount(postgresSQL, "roles")
	fmt.Printf("   ✅ Migrated %d roles\n", count)
	fmt.Println()

	// 2. Permissions (no dependencies)
	fmt.Println("2️⃣  Migrating permissions...")
	if err := migrateTable(sqliteSQL, postgresSQL, "permissions", []string{"id", "name", "description", "resource", "action", "scope", "created_at", "updated_at"}); err != nil {
		log.Fatalf("❌ Failed to migrate permissions: %v", err)
	}
	count = getCount(postgresSQL, "permissions")
	fmt.Printf("   ✅ Migrated %d permissions\n", count)
	fmt.Println()

	// 3. Role Permissions (depends on roles and permissions)
	fmt.Println("3️⃣  Migrating role_permissions...")
	if err := migrateTable(sqliteSQL, postgresSQL, "role_permissions", []string{"role_id", "permission_id", "created_at"}); err != nil {
		log.Fatalf("❌ Failed to migrate role_permissions: %v", err)
	}
	count = getCount(postgresSQL, "role_permissions")
	fmt.Printf("   ✅ Migrated %d role_permissions\n", count)
	fmt.Println()

	// 4. Companies (no dependencies)
	fmt.Println("4️⃣  Migrating companies...")
	// Note: main_parent_company_id tidak ada di SQLite, akan NULL di PostgreSQL
	companyColumns := []string{
		"id", "name", "short_name", "code", "description", "npwp", "nib", "status",
		"logo", "phone", "fax", "email", "website", "address", "operational_address",
		"parent_id", "level", "is_active", "created_at", "updated_at",
	}
	if err := migrateTable(sqliteSQL, postgresSQL, "companies", companyColumns); err != nil {
		log.Fatalf("❌ Failed to migrate companies: %v", err)
	}
	count = getCount(postgresSQL, "companies")
	fmt.Printf("   ✅ Migrated %d companies\n", count)
	fmt.Println()

	// 5. Users (depends on roles and companies)
	fmt.Println("5️⃣  Migrating users...")
	userColumns := []string{
		"id", "username", "email", "password", "role", "company_id", "role_id", "is_active", "created_at", "updated_at",
	}
	if err := migrateTable(sqliteSQL, postgresSQL, "users", userColumns); err != nil {
		log.Fatalf("❌ Failed to migrate users: %v", err)
	}
	count = getCount(postgresSQL, "users")
	fmt.Printf("   ✅ Migrated %d users\n", count)
	fmt.Println()

	// 6. Two Factor Auths (depends on users)
	fmt.Println("6️⃣  Migrating two_factor_auths...")
	if err := migrateTable(sqliteSQL, postgresSQL, "two_factor_auths", []string{"id", "user_id", "secret", "enabled", "backup_codes", "created_at", "updated_at"}); err != nil {
		log.Fatalf("❌ Failed to migrate two_factor_auths: %v", err)
	}
	count = getCount(postgresSQL, "two_factor_auths")
	fmt.Printf("   ✅ Migrated %d two_factor_auths\n", count)
	fmt.Println()

	// 7. Audit Logs (depends on users, but user_id can be null)
	fmt.Println("7️⃣  Migrating audit_logs...")
	auditColumns := []string{
		"id", "user_id", "username", "action", "resource", "resource_id",
		"ip_address", "user_agent", "details", "status", "log_type", "created_at",
	}
	if err := migrateTable(sqliteSQL, postgresSQL, "audit_logs", auditColumns); err != nil {
		log.Fatalf("❌ Failed to migrate audit_logs: %v", err)
	}
	count = getCount(postgresSQL, "audit_logs")
	fmt.Printf("   ✅ Migrated %d audit_logs\n", count)
	fmt.Println()

	fmt.Println("🎉 Migration completed successfully!")
	fmt.Println()
	fmt.Println("📊 Summary:")
	fmt.Printf("   - Roles: %d\n", getCount(postgresSQL, "roles"))
	fmt.Printf("   - Permissions: %d\n", getCount(postgresSQL, "permissions"))
	fmt.Printf("   - Role Permissions: %d\n", getCount(postgresSQL, "role_permissions"))
	fmt.Printf("   - Companies: %d\n", getCount(postgresSQL, "companies"))
	fmt.Printf("   - Users: %d\n", getCount(postgresSQL, "users"))
	fmt.Printf("   - Two Factor Auths: %d\n", getCount(postgresSQL, "two_factor_auths"))
	fmt.Printf("   - Audit Logs: %d\n", getCount(postgresSQL, "audit_logs"))
}

func migrateTable(sqliteDB, postgresDB *sql.DB, tableName string, columns []string) error {
	// Check if table exists in SQLite
	var count int
	err := sqliteDB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count)
	if err != nil {
		return fmt.Errorf("table %s not found in SQLite: %v", tableName, err)
	}

	if count == 0 {
		fmt.Printf("   ⚠️  Table %s is empty, skipping...\n", tableName)
		return nil
	}

	// Build column list for SELECT
	columnList := ""
	for i, col := range columns {
		if i > 0 {
			columnList += ", "
		}
		columnList += col
	}

	// Build placeholders for INSERT
	placeholders := ""
	for i := range columns {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += fmt.Sprintf("$%d", i+1)
	}

	// Query data from SQLite
	rows, err := sqliteDB.Query(fmt.Sprintf("SELECT %s FROM %s", columnList, tableName))
	if err != nil {
		return fmt.Errorf("failed to query SQLite: %v", err)
	}
	defer rows.Close()

	// Prepare INSERT statement for PostgreSQL
	insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT DO NOTHING", tableName, columnList, placeholders)
	stmt, err := postgresDB.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare INSERT statement: %v", err)
	}
	defer stmt.Close()

	// Get column types for scanning
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return fmt.Errorf("failed to get column types: %v", err)
	}

	// Insert data into PostgreSQL
	inserted := 0
	for rows.Next() {
		// Create scan destination based on column types
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return fmt.Errorf("failed to scan row: %v", err)
		}

		// Convert values to appropriate types for PostgreSQL
		convertedValues := make([]interface{}, len(values))
		for i, val := range values {
			convertedValues[i] = convertValue(val, columnTypes[i].DatabaseTypeName())
		}

		// Execute INSERT
		_, err := stmt.Exec(convertedValues...)
		if err != nil {
			// Log error but continue (might be duplicate)
			fmt.Printf("   ⚠️  Warning inserting row: %v\n", err)
			continue
		}
		inserted++
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %v", err)
	}

	fmt.Printf("   📝 Inserted %d rows\n", inserted)
	return nil
}

func convertValue(val interface{}, dbType string) interface{} {
	if val == nil {
		return nil
	}

	// Handle time.Time conversion
	switch v := val.(type) {
	case time.Time:
		return v
	case string:
		// Try to parse as time if it looks like a timestamp
		if t, err := time.Parse("2006-01-02 15:04:05.999999999-07:00", v); err == nil {
			return t
		}
		if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
			return t
		}
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t
		}
		return v
	case []byte:
		return string(v)
	case int64:
		// SQLite stores booleans as integers (0/1), convert to bool for PostgreSQL
		if dbType == "BOOLEAN" || dbType == "numeric" {
			return v != 0
		}
		return v
	case int:
		// SQLite stores booleans as integers (0/1), convert to bool for PostgreSQL
		if dbType == "BOOLEAN" || dbType == "numeric" {
			return v != 0
		}
		return v
	case float64:
		// SQLite might store booleans as float
		if dbType == "BOOLEAN" || dbType == "numeric" {
			return v != 0
		}
		return v
	}

	return val
}

func getCount(db *sql.DB, tableName string) int {
	var count int
	err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

