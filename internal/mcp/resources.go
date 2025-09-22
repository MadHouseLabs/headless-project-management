package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/headless-pm/headless-project-management/internal/models"
)

// ListResources returns the list of available MCP resources
func (s *EnhancedMCPServer) ListResources() []Resource {
	return []Resource{
		{
			URI:         "projects://list",
			Name:        "List of Projects",
			Description: "Get all projects with their status and metadata",
			MimeType:    "application/json",
		},
		{
			URI:         "tasks://overdue",
			Name:        "Overdue Tasks",
			Description: "Tasks that are past their due date",
			MimeType:    "application/json",
		},
		{
			URI:         "tasks://high-priority",
			Name:        "High Priority Tasks",
			Description: "Tasks marked as high priority across all projects",
			MimeType:    "application/json",
		},
		{
			URI:         "epics://active",
			Name:        "Active Epics",
			Description: "Currently active epics with their progress",
			MimeType:    "application/json",
		},
		{
			URI:         "labels://all",
			Name:        "All Labels",
			Description: "All labels across projects",
			MimeType:    "application/json",
		},
	}
}

// GetResource retrieves a specific resource by URI
func (s *EnhancedMCPServer) GetResource(ctx context.Context, uri string) (*ResourceContent, error) {
	switch uri {
	case "projects://list":
		return s.getProjectsList()
	case "tasks://overdue":
		return s.getOverdueTasks()
	case "tasks://high-priority":
		return s.getHighPriorityTasks()
	case "epics://active":
		return s.getActiveEpics()
	case "labels://all":
		return s.getAllLabels()
	default:
		return nil, fmt.Errorf("resource not found: %s", uri)
	}
}

func (s *EnhancedMCPServer) getProjectsList() (*ResourceContent, error) {
	var projects []models.Project
	if err := s.db.Find(&projects).Error; err != nil {
		return nil, err
	}

	projectInfo := make([]map[string]interface{}, 0, len(projects))
	for _, project := range projects {
		var taskCount int64
		s.db.Model(&models.Task{}).Where("project_id = ?", project.ID).Count(&taskCount)

		var completedTasks int64
		s.db.Model(&models.Task{}).Where("project_id = ? AND status = ?", project.ID, models.TaskStatusDone).Count(&completedTasks)

		info := map[string]interface{}{
			"id":               project.ID,
			"name":             project.Name,
			"description":      project.Description,
			"status":           project.Status,
			"total_tasks":      taskCount,
			"completed_tasks":  completedTasks,
			"created_at":       project.CreatedAt,
		}
		projectInfo = append(projectInfo, info)
	}

	return &ResourceContent{
		URI:      "projects://list",
		MimeType: "application/json",
		Content:  projectInfo,
	}, nil
}

func (s *EnhancedMCPServer) getOverdueTasks() (*ResourceContent, error) {
	var tasks []models.Task
	now := time.Now()
	if err := s.db.Where("due_date < ? AND status != ?", now, models.TaskStatusDone).
		Preload("Project").
		Preload("Labels").
		Order("due_date ASC").
		Find(&tasks).Error; err != nil {
		return nil, err
	}

	taskInfo := make([]map[string]interface{}, 0, len(tasks))
	for _, task := range tasks {
		info := map[string]interface{}{
			"id":          task.ID,
			"title":       task.Title,
			"description": task.Description,
			"due_date":    task.DueDate,
			"status":      task.Status,
			"priority":    task.Priority,
		}
		if task.Project != nil {
			info["project_name"] = task.Project.Name
		}
		if task.DueDate != nil {
			info["days_overdue"] = int(now.Sub(*task.DueDate).Hours() / 24)
		}
		taskInfo = append(taskInfo, info)
	}

	return &ResourceContent{
		URI:      "tasks://overdue",
		MimeType: "application/json",
		Content:  taskInfo,
	}, nil
}

func (s *EnhancedMCPServer) getHighPriorityTasks() (*ResourceContent, error) {
	var tasks []models.Task
	if err := s.db.Where("priority = ? AND status != ?", models.TaskPriorityHigh, models.TaskStatusDone).
		Preload("Project").
		Preload("Labels").
		Order("created_at DESC").
		Find(&tasks).Error; err != nil {
		return nil, err
	}

	taskInfo := make([]map[string]interface{}, 0, len(tasks))
	for _, task := range tasks {
		info := map[string]interface{}{
			"id":          task.ID,
			"title":       task.Title,
			"description": task.Description,
			"status":      task.Status,
			"priority":    task.Priority,
			"created_at":  task.CreatedAt,
		}
		if task.Project != nil {
			info["project_name"] = task.Project.Name
		}
		if task.DueDate != nil {
			info["due_date"] = task.DueDate
		}
		taskInfo = append(taskInfo, info)
	}

	return &ResourceContent{
		URI:      "tasks://high-priority",
		MimeType: "application/json",
		Content:  taskInfo,
	}, nil
}

func (s *EnhancedMCPServer) getActiveEpics() (*ResourceContent, error) {
	var epics []models.Epic
	if err := s.db.Where("status = ?", models.EpicStatusActive).
		Preload("Project").
		Preload("Tasks").
		Find(&epics).Error; err != nil {
		return nil, err
	}

	epicInfo := make([]map[string]interface{}, 0, len(epics))
	for _, epic := range epics {
		completedTasks := 0
		for _, task := range epic.Tasks {
			if task.Status == models.TaskStatusDone {
				completedTasks++
			}
		}

		progress := 0
		if len(epic.Tasks) > 0 {
			progress = (completedTasks * 100) / len(epic.Tasks)
		}

		info := map[string]interface{}{
			"id":              epic.ID,
			"name":            epic.Name,
			"description":     epic.Description,
			"status":          epic.Status,
			"progress":        progress,
			"total_tasks":     len(epic.Tasks),
			"completed_tasks": completedTasks,
		}
		if epic.Project != nil {
			info["project_name"] = epic.Project.Name
		}
		epicInfo = append(epicInfo, info)
	}

	return &ResourceContent{
		URI:      "epics://active",
		MimeType: "application/json",
		Content:  epicInfo,
	}, nil
}

func (s *EnhancedMCPServer) getAllLabels() (*ResourceContent, error) {
	var labels []models.Label
	if err := s.db.Find(&labels).Error; err != nil {
		return nil, err
	}

	labelInfo := make([]map[string]interface{}, 0, len(labels))
	for _, label := range labels {
		var taskCount int64
		s.db.Table("task_labels").Where("label_id = ?", label.ID).Count(&taskCount)

		info := map[string]interface{}{
			"id":         label.ID,
			"name":       label.Name,
			"color":      label.Color,
			"project_id": label.ProjectID,
			"task_count": taskCount,
		}
		labelInfo = append(labelInfo, info)
	}

	return &ResourceContent{
		URI:      "labels://all",
		MimeType: "application/json",
		Content:  labelInfo,
	}, nil
}