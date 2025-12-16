#!/bin/bash

# Script untuk migrasi SQLite database ke SQLCipher encrypted database
# Usage: ./migrate-sqlite-to-sqlcipher.sh [input_db] [output_db] [encryption_key]

set -e

INPUT_DB="${1:-dms.db}"
OUTPUT_DB="${2:-dms_encrypted.db}"
ENCRYPTION_KEY="${3:-${SQLCIPHER_KEY}}"

if [ -z "$ENCRYPTION_KEY" ]; then
    echo "Error: Encryption key is required"
    echo "Usage: $0 [input_db] [output_db] [encryption_key]"
    echo "   or: SQLCIPHER_KEY=your-key $0 [input_db] [output_db]"
    exit 1
fi

if [ ! -f "$INPUT_DB" ]; then
    echo "Error: Input database file '$INPUT_DB' not found"
    exit 1
fi

# Check if sqlcipher is installed
if ! command -v sqlcipher &> /dev/null; then
    echo "Error: sqlcipher is not installed"
    echo "Install it with: brew install sqlcipher (macOS) or apt-get install sqlcipher (Ubuntu)"
    exit 1
fi

echo "Migrating SQLite database to SQLCipher encrypted format..."
echo "Input: $INPUT_DB"
echo "Output: $OUTPUT_DB"
echo ""

# Backup original database
BACKUP_DB="${INPUT_DB}.backup.$(date +%Y%m%d_%H%M%S)"
cp "$INPUT_DB" "$BACKUP_DB"
echo "Backup created: $BACKUP_DB"
echo ""

# Convert to SQLCipher
sqlcipher "$INPUT_DB" <<EOF
ATTACH DATABASE '$OUTPUT_DB' AS encrypted KEY '$ENCRYPTION_KEY';
SELECT sqlcipher_export('encrypted');
DETACH DATABASE encrypted;
.quit
EOF

if [ $? -eq 0 ]; then
    echo ""
    echo "Migration successful!"
    echo "Encrypted database: $OUTPUT_DB"
    echo ""
    echo "To use the encrypted database:"
    echo "  1. Backup original: mv $INPUT_DB ${INPUT_DB}.plain"
    echo "  2. Replace with encrypted: mv $OUTPUT_DB $INPUT_DB"
    echo "  3. Set environment variable: export SQLCIPHER_KEY='$ENCRYPTION_KEY'"
    echo "  4. Set encryption flag: export ENABLE_SQLCIPHER=true"
    echo "  5. Restart application"
else
    echo ""
    echo "Migration failed! Original database is safe at: $INPUT_DB"
    exit 1
fi
