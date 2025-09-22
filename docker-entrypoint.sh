#!/bin/bash
set -e

# Create necessary directories if they don't exist
mkdir -p /data/db /data/uploads

# Execute the server
exec ./server