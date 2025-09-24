#!/bin/bash
set -e

# Check if /data exists, if not create it locally (fallback for when mount fails)
if [ ! -d "/data" ]; then
    echo "WARNING: /data directory does not exist, creating it locally"
    mkdir -p /data
fi

# Create necessary directories if they don't exist
echo "Creating directories..."
mkdir -p /data/db /data/uploads

# Set DATABASE_DIR to use the mounted directory directly
export DATABASE_DIR=/data

# Execute the server
echo "Starting server with DATABASE_DIR=$DATABASE_DIR..."
exec ./server