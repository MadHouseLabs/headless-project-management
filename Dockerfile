# Optimized Dockerfile using pre-built binary
# This expects the Go binary to be built outside the container
FROM debian:bookworm-slim

# Install runtime dependencies only
RUN apt-get update && apt-get install -y \
    ca-certificates \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy pre-built binary (built in CI pipeline)
COPY server .

# Copy templates and static files
COPY templates ./templates
COPY web/dist ./web/dist

# Copy entrypoint script
COPY docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

# Create data directory
RUN mkdir -p /data

# Expose port
EXPOSE 8080

# Run the server
ENTRYPOINT ["/app/docker-entrypoint.sh"]