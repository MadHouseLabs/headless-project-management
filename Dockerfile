# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc g++ musl-dev sqlite-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with sqlite-vec (only for amd64 to avoid ARM cross-compilation issues)
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o server cmd/server/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Copy templates and static files
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/web/dist ./web/dist

# Copy default config if exists
COPY --from=builder /app/config.example.json ./config.json 2>/dev/null || true

# Create data directories
RUN mkdir -p data/db data/uploads

# Expose port
EXPOSE 8080

# Run the server
CMD ["./server", "-config", "config.json"]