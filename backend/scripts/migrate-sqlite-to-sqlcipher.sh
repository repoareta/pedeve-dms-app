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

# Validate encryption key format (alphanumeric, dash, underscore only, no SQL injection)
if [[ ! "$ENCRYPTION_KEY" =~ ^[a-zA-Z0-9_-]+$ ]]; then
    echo "❌ Error: Encryption key contains invalid characters. Only alphanumeric, dash, and underscore allowed."
    exit 1
fi

# Validate output database path (no path traversal)
if [[ "$OUTPUT_DB" =~ \.\./ ]] || [[ "$OUTPUT_DB" =~ ^/ ]]; then
    echo "❌ Error: Output database path contains invalid characters or absolute path."
    exit 1
fi

# Convert to SQLCipher using printf to safely escape the key
# Note: sqlcipher expects the key as a string literal, so we use printf %q for safe quoting
ENCRYPTION_KEY_ESCAPED=$(printf '%q' "$ENCRYPTION_KEY")
OUTPUT_DB_ESCAPED=$(printf '%q' "$OUTPUT_DB")

sqlcipher "$INPUT_DB" <<EOF
ATTACH DATABASE $OUTPUT_DB_ESCAPED AS encrypted KEY $ENCRYPTION_KEY_ESCAPED;
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
