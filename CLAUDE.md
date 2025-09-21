# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A headless project management system built with Go, featuring REST APIs and MCP (Model Context Protocol) support for AI integration. The system uses SQLite with vector embeddings (sqlite-vec) for semantic search capabilities, integrated with Azure OpenAI.

## Key Commands

### Build and Run
```bash
task run          # Build and run server (production)
task dev          # Run with hot reload (development)
task build        # Build binary only
task clean        # Clean build artifacts and data
```

### Testing
```bash
task test         # Run all tests
task test:coverage # Run tests with coverage
task api:test     # Test API endpoints
```

### Code Quality
```bash
task fmt          # Format Go code
task lint         # Run golangci-lint
```

### Docker Operations
```bash
task docker:build # Build Docker image
task docker:run   # Run Docker container
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

Required `.env` file:
```bash
AZURE_OPENAI_ENDPOINT=<your-endpoint>
AZURE_OPENAI_API_KEY=<your-key>
AZURE_OPENAI_EMBEDDING_DEPLOYMENT=<deployment-name>
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
- Database stored in `./data/headless_pm.db`
- Uploads stored in `./data/uploads/`
- JWT auth required for most API endpoints (not MCP)
- MCP endpoints don't require authentication