package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/headless-pm/headless-project-management/internal/database"
	"github.com/headless-pm/headless-project-management/internal/models"
)

// MCPServer is the legacy MCP server implementation
// kept for backward compatibility
type MCPServer struct {
	db *database.Database
}

// NewMCPServer creates a legacy MCP server
func NewMCPServer(db *database.Database) *MCPServer {
	return &MCPServer{db: db}
}

func (s *MCPServer) ListTools() []Tool {
	return []Tool{
		{
			Name:        "create_project",
			Description: "Create a new project",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Project name",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Project description",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "list_projects",
			Description: "List all projects",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"status": map[string]interface{}{
						"type":        "string",
						"description": "Filter by status (active, archived, draft)",
						"enum":        []string{"active", "archived", "draft"},
					},
				},
			},
		},
		{
			Name:        "get_project",
			Description: "Get project details by ID",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "integer",
						"description": "Project ID",
					},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "create_task",
			Description: "Create a new task",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id": map[string]interface{}{
						"type":        "integer",
						"description": "Project ID",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Task title",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Task description",
					},
					"priority": map[string]interface{}{
						"type":        "string",
						"description": "Task priority",
						"enum":        []string{"low", "medium", "high", "urgent"},
					},
					"assignee": map[string]interface{}{
						"type":        "string",
						"description": "Task assignee",
					},
				},
				"required": []string{"project_id", "title"},
			},
		},
		{
			Name:        "list_tasks",
			Description: "List tasks with optional filters",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id": map[string]interface{}{
						"type":        "integer",
						"description": "Filter by project ID",
					},
					"status": map[string]interface{}{
						"type":        "string",
						"description": "Filter by status",
						"enum":        []string{"todo", "in_progress", "review", "done", "cancelled"},
					},
				},
			},
		},
		{
			Name:        "update_task_status",
			Description: "Update task status",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "integer",
						"description": "Task ID",
					},
					"status": map[string]interface{}{
						"type":        "string",
						"description": "New status",
						"enum":        []string{"todo", "in_progress", "review", "done", "cancelled"},
					},
				},
				"required": []string{"id", "status"},
			},
		},
		{
			Name:        "add_comment",
			Description: "Add a comment to a task",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]interface{}{
						"type":        "integer",
						"description": "Task ID",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Comment content",
					},
					"author": map[string]interface{}{
						"type":        "string",
						"description": "Comment author",
					},
				},
				"required": []string{"task_id", "content", "author"},
			},
		},
	}
}

func (s *MCPServer) ExecuteTool(ctx context.Context, call ToolCall) (*ToolResponse, error) {
	switch call.Name {
	case "create_project":
		return s.createProject(call.Arguments)
	case "list_projects":
		return s.listProjects(call.Arguments)
	case "get_project":
		return s.getProject(call.Arguments)
	case "create_task":
		return s.createTask(call.Arguments)
	case "list_tasks":
		return s.listTasks(call.Arguments)
	case "update_task_status":
		return s.updateTaskStatus(call.Arguments)
	case "add_comment":
		return s.addComment(call.Arguments)
	default:
		return &ToolResponse{
			Content: fmt.Sprintf("Unknown tool: %s", call.Name),
			IsError: true,
		}, nil
	}
}

func (s *MCPServer) createProject(args json.RawMessage) (*ToolResponse, error) {
	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return &ToolResponse{Content: err.Error(), IsError: true}, nil
	}

	project := &models.Project{
		Name:        input.Name,
		Description: input.Description,
		Status:      models.ProjectStatusActive,
	}

	if err := s.db.CreateProject(project); err != nil {
		return &ToolResponse{Content: err.Error(), IsError: true}, nil
	}

	return &ToolResponse{Content: project}, nil
}

func (s *MCPServer) listProjects(args json.RawMessage) (*ToolResponse, error) {
	var input struct {
		Status string `json:"status"`
	}
	json.Unmarshal(args, &input)

	var status *models.ProjectStatus
	if input.Status != "" {
		s := models.ProjectStatus(input.Status)
		status = &s
	}

	projects, err := s.db.ListProjects(status)
	if err != nil {
		return &ToolResponse{Content: err.Error(), IsError: true}, nil
	}

	return &ToolResponse{Content: projects}, nil
}

func (s *MCPServer) getProject(args json.RawMessage) (*ToolResponse, error) {
	var input struct {
		ID uint `json:"id"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return &ToolResponse{Content: err.Error(), IsError: true}, nil
	}

	project, err := s.db.GetProject(input.ID)
	if err != nil {
		return &ToolResponse{Content: err.Error(), IsError: true}, nil
	}

	return &ToolResponse{Content: project}, nil
}

func (s *MCPServer) createTask(args json.RawMessage) (*ToolResponse, error) {
	var input struct {
		ProjectID   uint   `json:"project_id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Priority    string `json:"priority"`
		Assignee    string `json:"assignee"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return &ToolResponse{Content: err.Error(), IsError: true}, nil
	}

	task := &models.Task{
		ProjectID:   input.ProjectID,
		Title:       input.Title,
		Description: input.Description,
		Status:      models.TaskStatusTodo,
		Assignee:    input.Assignee,
	}

	if input.Priority != "" {
		task.Priority = models.TaskPriority(input.Priority)
	} else {
		task.Priority = models.TaskPriorityMedium
	}

	if err := s.db.CreateTask(task); err != nil {
		return &ToolResponse{Content: err.Error(), IsError: true}, nil
	}

	return &ToolResponse{Content: task}, nil
}

func (s *MCPServer) listTasks(args json.RawMessage) (*ToolResponse, error) {
	var input struct {
		ProjectID uint   `json:"project_id"`
		Status    string `json:"status"`
	}
	json.Unmarshal(args, &input)

	var projectID *uint
	if input.ProjectID > 0 {
		projectID = &input.ProjectID
	}

	var status *models.TaskStatus
	if input.Status != "" {
		s := models.TaskStatus(input.Status)
		status = &s
	}

	tasks, err := s.db.ListTasks(projectID, status)
	if err != nil {
		return &ToolResponse{Content: err.Error(), IsError: true}, nil
	}

	return &ToolResponse{Content: tasks}, nil
}

func (s *MCPServer) updateTaskStatus(args json.RawMessage) (*ToolResponse, error) {
	var input struct {
		ID     uint   `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return &ToolResponse{Content: err.Error(), IsError: true}, nil
	}

	task, err := s.db.GetTask(input.ID)
	if err != nil {
		return &ToolResponse{Content: err.Error(), IsError: true}, nil
	}

	task.Status = models.TaskStatus(input.Status)
	if err := s.db.UpdateTask(task); err != nil {
		return &ToolResponse{Content: err.Error(), IsError: true}, nil
	}

	return &ToolResponse{Content: task}, nil
}

func (s *MCPServer) addComment(args json.RawMessage) (*ToolResponse, error) {
	var input struct {
		TaskID  uint   `json:"task_id"`
		Content string `json:"content"`
		Author  string `json:"author"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return &ToolResponse{Content: err.Error(), IsError: true}, nil
	}

	comment := &models.Comment{
		TaskID:  input.TaskID,
		Content: input.Content,
		Author:  input.Author,
	}

	if err := s.db.AddComment(comment); err != nil {
		return &ToolResponse{Content: err.Error(), IsError: true}, nil
	}

	return &ToolResponse{Content: comment}, nil
}