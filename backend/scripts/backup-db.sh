#!/bin/bash

# Script backup database PostgreSQL untuk Pedeve App
# Usage: ./backup-db.sh [backup_directory]
# Default backup directory: ./backups

set -e

# Konfigurasi default
BACKUP_DIR="${1:-./backups}"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="${BACKUP_DIR}/pedeve_db_backup_${TIMESTAMP}.sql"
RETENTION_DAYS=7  # Hapus backup lebih dari 7 hari

# Ambil konfigurasi dari environment variables (required for security)
# Never hardcode database credentials in source code
DB_HOST="${POSTGRES_HOST}"
DB_PORT="${POSTGRES_PORT:-5432}"
DB_NAME="${POSTGRES_DB}"
DB_USER="${POSTGRES_USER}"
DB_PASSWORD="${POSTGRES_PASSWORD}"

# Validate required environment variables
if [ -z "$DB_HOST" ] || [ -z "$DB_NAME" ] || [ -z "$DB_USER" ] || [ -z "$DB_PASSWORD" ]; then
    echo "❌ Error: Required environment variables are not set:" >&2
    [ -z "$DB_HOST" ] && echo "   - POSTGRES_HOST" >&2
    [ -z "$DB_NAME" ] && echo "   - POSTGRES_DB" >&2
    [ -z "$DB_USER" ] && echo "   - POSTGRES_USER" >&2
    [ -z "$DB_PASSWORD" ] && echo "   - POSTGRES_PASSWORD" >&2
    echo "   Please set all required variables before running this script." >&2
    exit 1
fi

# Buat direktori backup jika belum ada
mkdir -p "$BACKUP_DIR"

echo "=========================================="
echo "Pedeve App - Database Backup"
echo "=========================================="
echo "Database: $DB_NAME"
echo "Host: $DB_HOST:$DB_PORT"
echo "Backup file: $BACKUP_FILE"
echo "=========================================="

# Export password untuk pg_dump (menghindari prompt)
export PGPASSWORD="$DB_PASSWORD"

# Jalankan backup
echo "Starting backup..."
pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
    --no-owner \
    --no-acl \
    --clean \
    --if-exists \
    --format=plain \
    --file="$BACKUP_FILE"

# Compress backup file
echo "Compressing backup..."
gzip "$BACKUP_FILE"
BACKUP_FILE="${BACKUP_FILE}.gz"

# Hapus password dari environment
unset PGPASSWORD

# Hapus backup lama (lebih dari RETENTION_DAYS)
echo "Cleaning up old backups (older than $RETENTION_DAYS days)..."
find "$BACKUP_DIR" -name "pedeve_db_backup_*.sql.gz" -type f -mtime +$RETENTION_DAYS -delete

# Hitung ukuran backup
BACKUP_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)

echo "=========================================="
echo "Backup completed successfully!"
echo "Backup file: $BACKUP_FILE"
echo "Size: $BACKUP_SIZE"
echo "=========================================="

