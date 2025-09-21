# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o server cmd/server/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Copy default config
COPY --from=builder /app/config.example.json ./config.json

# Create data directories
RUN mkdir -p data/db data/uploads

# Expose port
EXPOSE 8080

# Run the server
CMD ["./server", "-config", "config.json"]