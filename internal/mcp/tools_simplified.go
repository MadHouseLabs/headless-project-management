package mcp

import (
	"encoding/json"
	"fmt"
	"github.com/headless-pm/headless-project-management/internal/models"
)

// Project CRUD operations
func (s *EnhancedMCPServer) getProject(args []byte) (*ToolResponse, error) {
	var input struct {
		ProjectID uint `json:"project_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	var project models.Project
	if err := s.db.First(&project, input.ProjectID).Error; err != nil {
		return ErrorResponse(fmt.Errorf("project not found: %w", err)), nil
	}

	return SuccessResponse(project), nil
}

func (s *EnhancedMCPServer) deleteProject(args []byte) (*ToolResponse, error) {
	var input struct {
		ProjectID uint `json:"project_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	if err := s.db.Delete(&models.Project{}, input.ProjectID).Error; err != nil {
		return ErrorResponse(fmt.Errorf("failed to delete project: %w", err)), nil
	}

	return SuccessResponse(map[string]string{"status": "deleted"}), nil
}

// Task CRUD operations
func (s *EnhancedMCPServer) getTask(args []byte) (*ToolResponse, error) {
	var input struct {
		TaskID uint `json:"task_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	var task models.Task
	if err := s.db.Preload("Labels").Preload("Assignee").First(&task, input.TaskID).Error; err != nil {
		return ErrorResponse(fmt.Errorf("task not found: %w", err)), nil
	}

	return SuccessResponse(task), nil
}

func (s *EnhancedMCPServer) deleteTask(args []byte) (*ToolResponse, error) {
	var input struct {
		TaskID uint `json:"task_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	if err := s.db.Delete(&models.Task{}, input.TaskID).Error; err != nil {
		return ErrorResponse(fmt.Errorf("failed to delete task: %w", err)), nil
	}

	return SuccessResponse(map[string]string{"status": "deleted"}), nil
}

// Epic CRUD operations
func (s *EnhancedMCPServer) createEpic(args []byte) (*ToolResponse, error) {
	var input struct {
		ProjectID   uint   `json:"project_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	epic := &models.Epic{
		ProjectID:   input.ProjectID,
		Name:        input.Name,
		Description: input.Description,
		Status:      "open",
	}

	if err := s.db.Create(epic).Error; err != nil {
		return ErrorResponse(fmt.Errorf("failed to create epic: %w", err)), nil
	}

	return SuccessResponse(epic), nil
}

func (s *EnhancedMCPServer) getEpic(args []byte) (*ToolResponse, error) {
	var input struct {
		EpicID uint `json:"epic_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	var epic models.Epic
	if err := s.db.Preload("Tasks").First(&epic, input.EpicID).Error; err != nil {
		return ErrorResponse(fmt.Errorf("epic not found: %w", err)), nil
	}

	return SuccessResponse(epic), nil
}

func (s *EnhancedMCPServer) updateEpic(args []byte) (*ToolResponse, error) {
	var input struct {
		EpicID      uint   `json:"epic_id"`
		Name        string `json:"name,omitempty"`
		Description string `json:"description,omitempty"`
		Status      string `json:"status,omitempty"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	updates := make(map[string]interface{})
	if input.Name != "" {
		updates["name"] = input.Name
	}
	if input.Description != "" {
		updates["description"] = input.Description
	}
	if input.Status != "" {
		updates["status"] = input.Status
	}

	if err := s.db.Model(&models.Epic{}).Where("id = ?", input.EpicID).Updates(updates).Error; err != nil {
		return ErrorResponse(fmt.Errorf("failed to update epic: %w", err)), nil
	}

	var epic models.Epic
	s.db.First(&epic, input.EpicID)
	return SuccessResponse(epic), nil
}

func (s *EnhancedMCPServer) deleteEpic(args []byte) (*ToolResponse, error) {
	var input struct {
		EpicID uint `json:"epic_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	if err := s.db.Delete(&models.Epic{}, input.EpicID).Error; err != nil {
		return ErrorResponse(fmt.Errorf("failed to delete epic: %w", err)), nil
	}

	return SuccessResponse(map[string]string{"status": "deleted"}), nil
}

func (s *EnhancedMCPServer) listEpics(args []byte) (*ToolResponse, error) {
	var input struct {
		ProjectID uint `json:"project_id,omitempty"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	var epics []models.Epic
	query := s.db.Preload("Tasks")

	if input.ProjectID > 0 {
		query = query.Where("project_id = ?", input.ProjectID)
	}

	if err := query.Find(&epics).Error; err != nil {
		return ErrorResponse(fmt.Errorf("failed to list epics: %w", err)), nil
	}

	return SuccessResponse(epics), nil
}

// Label operations
func (s *EnhancedMCPServer) listLabels(args []byte) (*ToolResponse, error) {
	var input struct {
		ProjectID uint `json:"project_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	var labels []models.Label
	if err := s.db.Where("project_id = ?", input.ProjectID).Find(&labels).Error; err != nil {
		return ErrorResponse(fmt.Errorf("failed to list labels: %w", err)), nil
	}

	return SuccessResponse(labels), nil
}

// Assignee operations
func (s *EnhancedMCPServer) assignTask(args []byte) (*ToolResponse, error) {
	var input struct {
		TaskID     uint `json:"task_id"`
		AssigneeID uint `json:"assignee_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	if err := s.db.Model(&models.Task{}).Where("id = ?", input.TaskID).Update("assignee_id", input.AssigneeID).Error; err != nil {
		return ErrorResponse(fmt.Errorf("failed to assign task: %w", err)), nil
	}

	var task models.Task
	s.db.Preload("Assignee").First(&task, input.TaskID)
	return SuccessResponse(task), nil
}

func (s *EnhancedMCPServer) listAssignees(args []byte) (*ToolResponse, error) {
	var input struct {
		ProjectID uint `json:"project_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	// Get all users who have tasks in this project
	var users []models.User
	if err := s.db.Distinct("users.*").
		Joins("JOIN tasks ON tasks.assignee_id = users.id").
		Where("tasks.project_id = ?", input.ProjectID).
		Find(&users).Error; err != nil {
		return ErrorResponse(fmt.Errorf("failed to list assignees: %w", err)), nil
	}

	return SuccessResponse(users), nil
}