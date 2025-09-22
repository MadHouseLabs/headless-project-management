package mcp

import (
	"context"
	"fmt"

	"github.com/headless-pm/headless-project-management/internal/database"
	"github.com/headless-pm/headless-project-management/internal/service"
	"github.com/headless-pm/headless-project-management/pkg/embeddings"
)

// EnhancedMCPServer implements the enhanced MCP server with all features
type EnhancedMCPServer struct {
	db                *database.Database
	embeddingProvider embeddings.EmbeddingProvider
	embeddingWorker   *service.EmbeddingWorker
}

// NewEnhancedMCPServer creates a new enhanced MCP server
func NewEnhancedMCPServer(db *database.Database, embeddingProvider embeddings.EmbeddingProvider, embeddingWorker *service.EmbeddingWorker) *EnhancedMCPServer {
	return &EnhancedMCPServer{
		db:                db,
		embeddingProvider: embeddingProvider,
		embeddingWorker:   embeddingWorker,
	}
}

// ListTools returns the list of available tools (enhanced version)
func (s *EnhancedMCPServer) ListTools() []Tool {
	return toolDefinitions()
}

// ExecuteTool executes a tool by name with the provided arguments
func (s *EnhancedMCPServer) ExecuteTool(ctx context.Context, call ToolCall) (*ToolResponse, error) {
	// Validate server state
	if s.db == nil {
		return ErrorResponse(ErrDatabaseNotConfigured), nil
	}

	// Route to appropriate handler based on tool name
	handlers := s.getToolHandlers()

	handler, exists := handlers[call.Name]
	if !exists {
		return ErrorResponse(fmt.Errorf("unknown tool: %s", call.Name)), nil
	}

	return handler(call.Arguments)
}

// getToolHandlers returns a map of tool names to their handler functions
func (s *EnhancedMCPServer) getToolHandlers() map[string]func([]byte) (*ToolResponse, error) {
	return map[string]func([]byte) (*ToolResponse, error){
		// Project Management (CRUD)
		"create_project": s.createProject,
		"get_project":    s.getProject,
		"update_project": s.updateProject,
		"delete_project": s.deleteProject,
		"list_projects":  s.listProjects,

		// Task Management (CRUD)
		"create_task": s.createTask,
		"get_task":    s.getTask,
		"update_task": s.updateTask,
		"delete_task": s.deleteTask,
		"list_tasks":  s.listTasks,

		// Epic Management (CRUD)
		"create_epic": s.createEpic,
		"get_epic":    s.getEpic,
		"update_epic": s.updateEpic,
		"delete_epic": s.deleteEpic,
		"list_epics":  s.listEpics,

		// Labels
		"create_label": s.createLabel,
		"assign_label": s.assignLabel,
		"list_labels":  s.listLabels,

		// Assignees
		"assign_task":    s.assignTask,
		"list_assignees": s.listAssignees,

		// Comments
		"add_comment":   s.addComment,
		"list_comments": s.listComments,

		// Task Dependencies
		"add_task_dependency":    s.addTaskDependency,
		"remove_task_dependency": s.removeTaskDependency,
		"list_task_dependencies": s.listTaskDependencies,
	}
}