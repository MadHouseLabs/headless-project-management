package mcp

import (
	"github.com/headless-pm/headless-project-management/internal/models"
)

// Project Management Tool Implementations

func (s *EnhancedMCPServer) createProject(args []byte) (*ToolResponse, error) {
	var input ProjectInput
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	project := &models.Project{
		Name:        input.Name,
		Description: input.Description,
		OwnerID:     input.OwnerID,
		Status:      models.ProjectStatusActive,
	}

	// Set TeamID if provided
	if input.TeamID > 0 {
		project.TeamID = &input.TeamID
	}

	if err := s.db.CreateProject(project); err != nil {
		return ErrorResponse(err), nil
	}

	// Queue embedding generation if worker is available
	if s.embedWorker != nil {
		s.embedWorker.QueueJob("project", project.ID)
	}

	return SuccessResponse(project), nil
}

func (s *EnhancedMCPServer) listProjects(args []byte) (*ToolResponse, error) {
	var input struct {
		Status string `json:"status"`
		TeamID uint   `json:"team_id"`
	}
	// Ignore unmarshal errors for optional parameters
	_ = UnmarshalArgs(args, &input)

	query := s.db.DB
	if input.Status != "" {
		query = query.Where("status = ?", input.Status)
	}
	if input.TeamID > 0 {
		query = query.Where("team_id = ?", input.TeamID)
	}

	var projects []models.Project
	if err := query.Find(&projects).Error; err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(projects), nil
}

func (s *EnhancedMCPServer) updateProject(args []byte) (*ToolResponse, error) {
	var input struct {
		ID          uint   `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	project, err := s.db.GetProject(input.ID)
	if err != nil {
		return ErrorResponse(err), nil
	}

	// Update only provided fields
	if input.Name != "" {
		project.Name = input.Name
	}
	if input.Description != "" {
		project.Description = input.Description
	}
	if input.Status != "" {
		project.Status = models.ProjectStatus(input.Status)
	}

	if err := s.db.UpdateProject(project); err != nil {
		return ErrorResponse(err), nil
	}

	// Queue embedding regeneration if worker is available
	if s.embedWorker != nil {
		s.embedWorker.QueueJob("project", project.ID)
	}

	return SuccessResponse(project), nil
}

func (s *EnhancedMCPServer) getProjectStats(args []byte) (*ToolResponse, error) {
	var input struct {
		ProjectID uint `json:"project_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	stats := make(map[string]interface{})

	// Task counts by status
	var taskStats []struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	}
	if err := s.db.Model(&models.Task{}).
		Select("status, count(*) as count").
		Where("project_id = ?", input.ProjectID).
		Group("status").
		Scan(&taskStats).Error; err != nil {
		return ErrorResponse(err), nil
	}
	stats["task_stats"] = taskStats

	// Total time logged
	var totalHours float64
	if err := s.db.Model(&models.TimeEntry{}).
		Joins("JOIN tasks ON tasks.id = time_entries.task_id").
		Where("tasks.project_id = ?", input.ProjectID).
		Select("COALESCE(sum(hours), 0)").
		Scan(&totalHours).Error; err != nil {
		return ErrorResponse(err), nil
	}
	stats["total_hours"] = totalHours

	// Team size
	var teamSize int64
	if err := s.db.Model(&models.TeamMember{}).
		Joins("JOIN projects ON projects.team_id = team_members.team_id").
		Where("projects.id = ?", input.ProjectID).
		Count(&teamSize).Error; err != nil {
		return ErrorResponse(err), nil
	}
	stats["team_size"] = teamSize

	// Active milestones
	var activeMilestones int64
	if err := s.db.Model(&models.Milestone{}).
		Where("project_id = ? AND status IN ?", input.ProjectID,
			[]string{string(models.MilestoneStatusPlanned), string(models.MilestoneStatusActive)}).
		Count(&activeMilestones).Error; err != nil {
		return ErrorResponse(err), nil
	}
	stats["active_milestones"] = activeMilestones

	return SuccessResponse(stats), nil
}