#!/bin/bash
# Database Backup Script for BookRank

set -euo pipefail

# Configuration
BACKUP_DIR="/backups"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="bookrank_backup_${TIMESTAMP}.sql"
RETENTION_DAYS=${BACKUP_RETENTION_DAYS:-7}

# Database connection variables (set via environment)
DB_HOST=${PGHOST:-localhost}
DB_PORT=${PGPORT:-5432}
DB_NAME=${PGDATABASE:-bookrank}
DB_USER=${PGUSER:-bookrank}

# Logging function
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >&2
}

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

log "Starting database backup for $DB_NAME"

# Create database backup
if pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
   --verbose --clean --if-exists --create \
   --format=custom --compress=9 \
   --file="$BACKUP_DIR/$BACKUP_FILE"; then
    log "Backup completed successfully: $BACKUP_FILE"

    # Compress the backup (if not already compressed by pg_dump)
    if [[ "$BACKUP_FILE" != *.gz ]]; then
        gzip "$BACKUP_DIR/$BACKUP_FILE"
        BACKUP_FILE="${BACKUP_FILE}.gz"
        log "Backup compressed: $BACKUP_FILE"
    fi

    # Verify backup integrity
    if [[ -f "$BACKUP_DIR/$BACKUP_FILE" ]] && [[ -s "$BACKUP_DIR/$BACKUP_FILE" ]]; then
        log "Backup verification successful"

        # Upload to S3 if AWS CLI is available and S3_BUCKET is set
        if command -v aws &> /dev/null && [[ -n "${S3_BUCKET:-}" ]]; then
            log "Uploading backup to S3: s3://$S3_BUCKET/database-backups/$BACKUP_FILE"
            if aws s3 cp "$BACKUP_DIR/$BACKUP_FILE" "s3://$S3_BUCKET/database-backups/$BACKUP_FILE" \
               --storage-class STANDARD_IA \
               --server-side-encryption AES256; then
                log "S3 upload completed successfully"
            else
                log "ERROR: S3 upload failed"
                exit 1
            fi
        fi

    else
        log "ERROR: Backup verification failed - file is empty or missing"
        exit 1
    fi
else
    log "ERROR: Database backup failed"
    exit 1
fi

# Clean up old local backups
log "Cleaning up backups older than $RETENTION_DAYS days"
find "$BACKUP_DIR" -name "bookrank_backup_*.sql*" -type f -mtime +$RETENTION_DAYS -delete

# Clean up old S3 backups if AWS CLI is available
if command -v aws &> /dev/null && [[ -n "${S3_BUCKET:-}" ]]; then
    log "Cleaning up old S3 backups"
    # List and delete backups older than retention period
    OLD_DATE=$(date -d "$RETENTION_DAYS days ago" +%Y%m%d)
    aws s3 ls "s3://$S3_BUCKET/database-backups/" | while read -r line; do
        BACKUP_DATE=$(echo "$line" | awk '{print $4}' | grep -o '[0-9]\{8\}' | head -1)
        if [[ -n "$BACKUP_DATE" ]] && [[ "$BACKUP_DATE" -lt "$OLD_DATE" ]]; then
            FILENAME=$(echo "$line" | awk '{print $4}')
            log "Deleting old S3 backup: $FILENAME"
            aws s3 rm "s3://$S3_BUCKET/database-backups/$FILENAME"
        fi
    done
fi

log "Backup process completed successfully"