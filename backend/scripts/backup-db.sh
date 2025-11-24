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

# Ambil konfigurasi dari environment atau gunakan default
DB_HOST="${POSTGRES_HOST:-localhost}"
DB_PORT="${POSTGRES_PORT:-5432}"
DB_NAME="${POSTGRES_DB:-db_dms_pedeve}"
DB_USER="${POSTGRES_USER:-postgres}"
DB_PASSWORD="${POSTGRES_PASSWORD:-dms_password}"

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

