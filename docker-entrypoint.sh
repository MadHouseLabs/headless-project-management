#!/bin/bash
set -e

# Check if /data exists, if not create it locally (fallback for when mount fails)
if [ ! -d "/data" ]; then
    echo "WARNING: /data directory does not exist, creating it locally"
    mkdir -p /data
fi

# Create necessary directories if they don't exist
echo "Creating /data/db and /data/uploads directories..."
mkdir -p /data/db /data/uploads

# Check if directories were created successfully
if [ -d "/data/db" ] && [ -d "/data/uploads" ]; then
    echo "Directories created successfully"
    ls -la /data/
else
    echo "ERROR: Failed to create directories"
    exit 1
fi

# Execute the server
echo "Starting server..."
exec ./server