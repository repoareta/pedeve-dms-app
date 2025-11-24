# Database Backup Scripts

## backup-db.sh

Script untuk backup database PostgreSQL secara manual atau otomatis.

### Usage

```bash
# Backup ke direktori default (./backups)
./scripts/backup-db.sh

# Backup ke direktori custom
./scripts/backup-db.sh /path/to/backup/directory
```

### Environment Variables

Script menggunakan environment variables berikut (dengan default values):

- `POSTGRES_HOST` (default: `localhost`)
- `POSTGRES_PORT` (default: `5432`)
- `POSTGRES_DB` (default: `db_dms_pedeve`)
- `POSTGRES_USER` (default: `postgres`)
- `POSTGRES_PASSWORD` (default: `dms_password`)

### Automated Backup (Cron)

Untuk menjalankan backup otomatis setiap hari jam 2 pagi:

```bash
# Edit crontab
crontab -e

# Tambahkan baris berikut:
0 2 * * * cd /path/to/dms-app/backend && ./scripts/backup-db.sh >> /var/log/db-backup.log 2>&1
```

### Backup Retention

Backup otomatis dihapus setelah 7 hari (configurable di script: `RETENTION_DAYS`).

### Restore Database

Untuk restore dari backup:

```bash
# Uncompress backup
gunzip backups/pedeve_db_backup_YYYYMMDD_HHMMSS.sql.gz

# Restore database
psql -h localhost -U postgres -d db_dms_pedeve < backups/pedeve_db_backup_YYYYMMDD_HHMMSS.sql
```

### Catatan

- Backup file di-compress dengan gzip untuk menghemat space
- Backup menggunakan `--clean --if-exists` untuk memastikan restore yang bersih
- Pastikan PostgreSQL client tools (`pg_dump`, `psql`) sudah terinstall

