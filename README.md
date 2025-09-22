# Headless Project Management System

A lightweight, headless project management system built with Go, featuring REST APIs and MCP (Model Context Protocol) support for AI integration through a unified HTTP server.

## Features

- **Project Management**: Create, read, update, and delete projects
- **Task Management**: Hierarchical task structure with subtasks
- **Comments**: Add comments to tasks for collaboration
- **File Attachments**: Upload and manage file attachments for tasks
- **Labels**: Organize tasks with customizable labels
- **MCP Integration**: AI-powered project management through Model Context Protocol
- **SQLite Database**: Lightweight, file-based database for easy deployment
- **File System Storage**: Organized file storage for attachments

## Architecture

- **Backend**: Go with Gin web framework
- **Database**: SQLite for persistent storage
- **Storage**: Filesystem-based attachment storage
- **API**: RESTful endpoints for all operations
- **MCP Server**: JSON-RPC based Model Context Protocol server

## Installation

### Using Task (recommended)
```bash
# Clone the repository
git clone https://github.com/headless-pm/headless-project-management.git
cd headless-project-management

# Install Task runner (if not already installed)
# macOS: brew install go-task
# Linux: sh -c "$(curl -sL https://taskfile.dev/install.sh)"

# Build the server
task build
```

### Manual Build
```bash
# Install dependencies
go mod download

# Build the server
go build -o bin/server cmd/server/main.go
```

## Configuration

The application is configured using environment variables. Copy `.env.example` to `.env`:

```bash
cp .env.example .env
```

Key environment variables:
- `SERVER_HOST`: Server host (default: localhost)
- `SERVER_PORT`: Server port (default: 8080)
- `DATABASE_DIR`: Database directory (default: ./data)
- `UPLOAD_DIR`: Upload directory (default: ./data/uploads)
- `MCP_ENABLED`: Enable MCP server (default: true)
- `ADMIN_API_TOKEN`: Admin token for creating API tokens

## Running the Server

### Using Task
```bash
# Run in production mode
task run

# Run in development mode
task dev

# List all available tasks
task --list
```

### Manual Run
```bash
./bin/server
```

### Using Docker
```bash
# Build Docker image
task docker:build

# Run Docker container
task docker:run
```

The unified server will start on `http://localhost:8080` and provides:
- REST API endpoints at `/api`
- MCP endpoints at `/api/mcp`
- Health check at `/health`

## API Endpoints

### MCP (Model Context Protocol)
- `GET /api/mcp/tools` - List available MCP tools
- `POST /api/mcp/tools/call` - Execute an MCP tool

### Projects
- `POST /api/projects` - Create a new project
- `GET /api/projects` - List all projects
- `GET /api/projects/:id` - Get project details
- `PUT /api/projects/:id` - Update project
- `DELETE /api/projects/:id` - Delete project

### Tasks
- `POST /api/tasks` - Create a new task
- `GET /api/tasks` - List all tasks
- `GET /api/tasks/:id` - Get task details
- `PUT /api/tasks/:id` - Update task
- `DELETE /api/tasks/:id` - Delete task
- `POST /api/tasks/:id/comments` - Add comment to task
- `POST /api/tasks/:id/attachments` - Upload attachment to task

### Health Check
- `GET /health` - Check server health

## MCP Tools

The MCP HTTP endpoints provide the following tools for AI assistants:

- `create_project` - Create a new project
- `list_projects` - List all projects with optional status filter
- `get_project` - Get project details by ID
- `create_task` - Create a new task in a project
- `list_tasks` - List tasks with optional filters
- `update_task_status` - Update the status of a task
- `add_comment` - Add a comment to a task

## Development

### Project Structure
```
├── cmd/
│   └── server/       # Unified server entry point
├── internal/
│   ├── api/          # HTTP handlers and routes
│   ├── models/       # Data models
│   ├── database/     # Database layer
│   ├── storage/      # File storage
│   └── mcp/          # MCP server implementation
├── pkg/
│   └── config/       # Configuration management
├── data/
│   ├── db/           # SQLite database files
│   └── uploads/      # File attachments
├── migrations/       # Database migrations
├── Taskfile.yml      # Task automation
├── Dockerfile        # Container configuration
└── .env.example      # Example environment variables
```

### Running Tests
```bash
# Run all tests
task test

# Run tests with coverage
task test:coverage

# Test API endpoints
task api:test
```

### Building for Production
```bash
# Using Task
task build

# Manual build with optimizations
go build -ldflags="-s -w" -o bin/server cmd/server/main.go

# Docker build
task docker:build
```

## License

MIT