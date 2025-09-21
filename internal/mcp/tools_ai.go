package mcp

import (
	"fmt"

	"github.com/headless-pm/headless-project-management/internal/models"
)

// AI and Search Tool Implementations

func (s *EnhancedMCPServer) semanticSearch(args []byte) (*ToolResponse, error) {
	if s.vectorService == nil {
		return ErrorResponse(ErrEmbeddingsNotConfigured), nil
	}

	var input SearchInput
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	// Set defaults
	if input.EntityType == "" {
		input.EntityType = "task"
	}
	if input.Limit == 0 {
		input.Limit = 10
	}

	// Validate entity type
	validTypes := map[string]bool{
		"project":  true,
		"task":     true,
		"document": true,
	}
	if !validTypes[input.EntityType] {
		return ErrorResponse(fmt.Errorf("invalid entity type: %s", input.EntityType)), nil
	}

	results, err := s.vectorService.SemanticSearch(input.Query, input.EntityType, input.Limit)
	if err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(results), nil
}

func (s *EnhancedMCPServer) findSimilarTasks(args []byte) (*ToolResponse, error) {
	if s.vectorService == nil {
		return ErrorResponse(ErrEmbeddingsNotConfigured), nil
	}

	var input struct {
		TaskID uint `json:"task_id"`
		Limit  int  `json:"limit"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	if input.Limit == 0 {
		input.Limit = 5
	}

	// Validate task exists
	if _, err := s.db.GetTask(input.TaskID); err != nil {
		return ErrorResponse(ErrTaskNotFound), nil
	}

	tasks, err := s.vectorService.FindSimilarTasks(input.TaskID, input.Limit)
	if err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(tasks), nil
}

func (s *EnhancedMCPServer) recommendTasks(args []byte) (*ToolResponse, error) {
	if s.vectorService == nil {
		return ErrorResponse(ErrEmbeddingsNotConfigured), nil
	}

	var input struct {
		UserID uint `json:"user_id"`
		Limit  int  `json:"limit"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	if input.Limit == 0 {
		input.Limit = 10
	}

	// Validate user exists
	var user models.User
	if err := s.db.First(&user, input.UserID).Error; err != nil {
		return ErrorResponse(ErrUserNotFound), nil
	}

	tasks, err := s.vectorService.RecommendTasks(input.UserID, input.Limit)
	if err != nil {
		return ErrorResponse(err), nil
	}

	// If no recommendations based on history, return high-priority unassigned tasks
	if len(tasks) == 0 {
		if err := s.db.
			Where("assignee_id IS NULL OR assignee_id = ?", input.UserID).
			Where("status IN ?", []string{
				string(models.TaskStatusTodo),
				string(models.TaskStatusInProgress),
			}).
			Order("priority DESC, created_at ASC").
			Limit(input.Limit).
			Find(&tasks).Error; err != nil {
			return ErrorResponse(err), nil
		}
	}

	return SuccessResponse(tasks), nil
}