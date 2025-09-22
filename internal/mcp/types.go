package mcp

import (
	"encoding/json"
)

// Tool represents an MCP tool definition
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ToolCall represents a request to execute a tool
type ToolCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ToolResponse represents the result of a tool execution
type ToolResponse struct {
	Content interface{} `json:"content"`
	IsError bool        `json:"isError,omitempty"`
}

// Resource represents an MCP resource
type Resource struct {
	URI         string                 `json:"uri"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	MimeType    string                 `json:"mimeType"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ResourceContent represents the content of a resource
type ResourceContent struct {
	URI      string      `json:"uri"`
	MimeType string      `json:"mimeType"`
	Content  interface{} `json:"content"`
}

// Common input structures for tools
type ProjectInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerID     uint   `json:"owner_id"`
}

type TaskInput struct {
	ProjectID   uint     `json:"project_id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Priority    string   `json:"priority"`
	AssigneeID  uint     `json:"assignee_id"`
	EpicID      uint     `json:"epic_id"`
	DueDate     string   `json:"due_date"`
	Labels      []string `json:"labels"`
}

// Helper function to create a success response
func SuccessResponse(content interface{}) *ToolResponse {
	return &ToolResponse{
		Content: content,
		IsError: false,
	}
}

// Helper function to create an error response
func ErrorResponse(err error) *ToolResponse {
	return &ToolResponse{
		Content: err.Error(),
		IsError: true,
	}
}

// Helper function to unmarshal arguments
func UnmarshalArgs(args json.RawMessage, target interface{}) error {
	return json.Unmarshal(args, target)
}