# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Headless Project Management System - A Go-based REST API and MCP server for project management with AI-powered semantic search capabilities using vector embeddings.

## Key Commands

### Build and Run
```bash
task run          # Build and run server (production)
task dev          # Run with hot reload (development)
task build:all    # Build server and web assets
task web:dev      # Run Vite dev server for CSS hot reload (separate terminal)
task clean        # Clean build artifacts and data
```

### Testing
```bash
task test         # Run all tests
go test -v -run TestName ./internal/service/  # Run single test
go test -v ./internal/database -run TestDatabaseConnection
```

### Code Quality
```bash
go fmt ./...      # Format Go code
go vet ./...      # Run Go vet
golangci-lint run # Run linter (if installed)
```

### Docker Operations
```bash
task docker:build # Build Docker image
source docker-remote-env.sh  # Configure Docker for remote VM
docker logs headless-pm       # View application logs on VM
```

## Architecture

### Service Layers

1. **Unified HTTP Server** (`cmd/server/main.go`)
   - Single server hosting both API and MCP endpoints
   - API routes at `/api`, MCP routes at `/mcp`
   - Initializes embedding provider and vector service
   - Sets up background embedding worker

2. **Database Layer** (`internal/database/`)
   - GORM-based ORM with SQLite backend
   - Vector embeddings via sqlite-vec extension
   - Automatic migrations on startup
   - Pointer types for optional foreign keys (TeamID, AssigneeID, etc.)

3. **MCP Implementation** (`internal/mcp/`)
   - `server_enhanced.go`: Main MCP server with 33 tools
   - `tools_*.go`: Tool implementations by domain (project, task, team, etc.)
   - `resources.go`: 10 MCP resources for contextual information
   - `types.go`: Centralized type definitions
   - Tools auto-generate embeddings via background worker

4. **Vector/AI Service** (`internal/service/`)
   - `vector_service.go`: Semantic search and similarity operations
   - `embedding_worker.go`: Background queue for embedding generation
   - Auto-indexes entities on create/update
   - Hybrid search combining semantic and keyword matching

5. **Embedding Provider** (`pkg/embeddings/`)
   - `azure_openai.go`: REST API client for Azure OpenAI embeddings
   - Configured via environment variables in `.env`
   - Falls back to local provider if Azure unavailable

### Key Design Patterns

- **Pointer Fields**: Optional relationships use pointers (e.g., `TeamID *uint`)
- **Background Processing**: Embedding generation happens asynchronously
- **Unified Server**: Single process serves both API and MCP protocols
- **Auto-indexing**: Embeddings generated automatically on data changes

## Environment Configuration

### Required for AI Features
```bash
AZURE_OPENAI_ENDPOINT=https://your-instance.openai.azure.com
AZURE_OPENAI_API_KEY=your-key
AZURE_OPENAI_EMBEDDING_DEPLOYMENT=your-embedding-model
```

### Authentication & Server
```bash
ADMIN_API_TOKEN=your-admin-token  # If not set, generates temporary token
DATABASE_DIR=/data                # Database location (default: ./data)
UPLOAD_DIR=/data/uploads          # File upload location
MCP_ENABLED=true                  # Enable MCP server (default: true)
SERVER_HOST=0.0.0.0               # Server bind address
SERVER_PORT=8080                  # Server port
```

## MCP Tools Available

The system provides 33 MCP tools including:
- Project/Task/Team management
- Semantic and hybrid search
- AI-powered task recommendations
- Intelligent task assignment
- Sprint and milestone tracking
- Workflow management

Access tools at `/mcp/tools` and resources at `/mcp/resources`.

## Deployment Architecture

### Critical: SQLite & Azure Files Incompatibility
SQLite's file locking is incompatible with Azure Files (SMB/CIFS). The Docker entrypoint implements a sync mechanism:
1. Database copied from Azure Files to `/tmp/db` on startup
2. Background sync every 5 minutes back to Azure Files
3. Sync on shutdown via signal trapping
4. Set `DATABASE_DIR=/tmp` in container

### GitHub Actions Deployment
Workflow (`.github/workflows/deploy.yml`):
1. Builds Docker image → `ghcr.io/madhouselabs/headless-project-management:latest`
2. Deploys to Azure VM at `pm-instance.khost.dev`
3. Required secrets: `DOCKER_CA_PEM`, `DOCKER_CERT_PEM`, `DOCKER_KEY_PEM`

## Common Issues & Solutions

### "unable to open database file"
**Cause**: SQLite file locking incompatibility with network filesystems
**Solution**: Ensure `DATABASE_DIR` points to local filesystem in container

### Vector search returns no results
1. Check embeddings: `SELECT COUNT(*) FROM embeddings`
2. Verify Azure OpenAI credentials
3. Check logs for "Starting embedding worker"

### GitHub Actions deployment fails
1. Docker image name must be lowercase
2. Verify GitHub secrets configured
3. Check VM accessibility at pm-instance.khost.dev:2376

## Common Development Tasks

### Adding a New MCP Tool
1. Implement handler in appropriate `internal/mcp/tools_*.go` file
2. Add tool definition in `internal/mcp/tool_definitions.go`
3. Register handler in `server_enhanced.go` `getToolHandlers()` map

### Adding a New API Endpoint
1. Create handler in `internal/api/`
2. Register route in `internal/api/router_extended.go`
3. Add authentication middleware if needed

### Working with Embeddings
- Embeddings auto-generate via background worker
- Manual queueing: `embeddingWorker.QueueJob("entity_type", entityID)`
- Vector service provides semantic search and similarity methods

## Important Notes

- Server runs on `localhost:8080` by default
- Database stored in `./data/db/projects.db` (note: different path in Docker)
- Uploads stored in `./data/uploads/`
- JWT auth required for API endpoints (`/api/*`)
- MCP endpoints (`/mcp/*`) require JWT authentication
- Admin endpoints (`/admin/*`) require admin token
- Web UI at root paths requires no authentication
- CGO required for sqlite-vec extension