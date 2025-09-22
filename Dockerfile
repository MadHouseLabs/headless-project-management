# Build stage
FROM golang:1.23-bookworm AS builder

WORKDIR /app

# Install build dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    sqlite3 \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64
RUN go build -o server cmd/server/main.go

# Final stage
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Copy templates and static files
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/web/dist ./web/dist

# Create data directories
RUN mkdir -p /data/db /data/uploads

# Expose port
EXPOSE 8080

# Run the server
CMD ["./server"]