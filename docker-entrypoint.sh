#!/bin/bash
set -e

# Check if /data exists, if not create it locally (fallback for when mount fails)
if [ ! -d "/data" ]; then
    echo "WARNING: /data directory does not exist, creating it locally"
    mkdir -p /data
fi

# Create necessary directories if they don't exist
echo "Creating directories..."
mkdir -p /data/db /data/uploads /tmp/db

# If database exists in Azure Files, copy it to local temp for SQLite compatibility
if [ -f "/data/db/projects.db" ]; then
    echo "Copying database from Azure Files to local storage..."
    cp /data/db/projects.db /tmp/db/projects.db
    cp /data/db/projects.db-* /tmp/db/ 2>/dev/null || true
fi

# Set DATABASE_DIR to use local temp directory
export DATABASE_DIR=/tmp

# Function to sync database back to Azure Files
sync_database() {
    if [ -f "/tmp/db/projects.db" ]; then
        echo "Syncing database to Azure Files..."
        cp /tmp/db/projects.db /data/db/projects.db
        cp /tmp/db/projects.db-* /data/db/ 2>/dev/null || true
    fi
}

# Trap signals to sync database on shutdown
trap 'sync_database; exit' SIGTERM SIGINT

# Start background sync every 5 minutes
(
    while true; do
        sleep 300
        sync_database
    done
) &

# Execute the server
echo "Starting server with DATABASE_DIR=$DATABASE_DIR..."
exec ./server