package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/headless-pm/headless-project-management/internal/models"
)

// Resources use the types defined in types.go

// ListResources returns available resources for context
func (s *EnhancedMCPServer) ListResources() []Resource {
	return []Resource{
		{
			URI:         "project://overview",
			Name:        "Project Overview",
			Description: "Get an overview of all projects with their status and metrics",
			MimeType:    "application/json",
		},
		{
			URI:         "project://active",
			Name:        "Active Projects",
			Description: "List of all active projects with current tasks",
			MimeType:    "application/json",
		},
		{
			URI:         "project://team-dashboard",
			Name:        "Team Dashboard",
			Description: "Team performance metrics and current assignments",
			MimeType:    "application/json",
		},
		{
			URI:         "sprint://current",
			Name:        "Current Sprints",
			Description: "All currently active sprints with their progress",
			MimeType:    "application/json",
		},
		{
			URI:         "tasks://high-priority",
			Name:        "High Priority Tasks",
			Description: "All high priority and urgent tasks across projects",
			MimeType:    "application/json",
		},
		{
			URI:         "tasks://overdue",
			Name:        "Overdue Tasks",
			Description: "Tasks that are past their due date",
			MimeType:    "application/json",
		},
		{
			URI:         "milestones://upcoming",
			Name:        "Upcoming Milestones",
			Description: "Milestones due in the next 30 days",
			MimeType:    "application/json",
		},
		{
			URI:         "notifications://unread",
			Name:        "Unread Notifications",
			Description: "All unread notifications for active users",
			MimeType:    "application/json",
		},
		{
			URI:         "analytics://weekly",
			Name:        "Weekly Analytics",
			Description: "Project and team analytics for the current week",
			MimeType:    "application/json",
		},
		{
			URI:         "workflow://states",
			Name:        "Workflow States",
			Description: "All configured workflows and their states",
			MimeType:    "application/json",
		},
	}
}

// GetResource returns the content of a specific resource
func (s *EnhancedMCPServer) GetResource(ctx context.Context, uri string) (*ResourceContent, error) {
	switch uri {
	case "project://overview":
		return s.getProjectOverview()
	case "project://active":
		return s.getActiveProjects()
	case "project://team-dashboard":
		return s.getTeamDashboard()
	case "sprint://current":
		return s.getCurrentSprints()
	case "tasks://high-priority":
		return s.getHighPriorityTasks()
	case "tasks://overdue":
		return s.getOverdueTasks()
	case "milestones://upcoming":
		return s.getUpcomingMilestones()
	case "notifications://unread":
		return s.getUnreadNotifications()
	case "analytics://weekly":
		return s.getWeeklyAnalytics()
	case "workflow://states":
		return s.getWorkflowStates()
	default:
		return nil, fmt.Errorf("unknown resource URI: %s", uri)
	}
}

func (s *EnhancedMCPServer) getProjectOverview() (*ResourceContent, error) {
	var projects []models.Project
	if err := s.db.Preload("Tasks").Find(&projects).Error; err != nil {
		return nil, err
	}

	overview := make([]map[string]interface{}, 0, len(projects))
	for _, project := range projects {
		stats := map[string]int{
			"total": 0,
			"todo": 0,
			"in_progress": 0,
			"done": 0,
		}

		for _, task := range project.Tasks {
			stats["total"]++
			switch task.Status {
			case models.TaskStatusTodo:
				stats["todo"]++
			case models.TaskStatusInProgress:
				stats["in_progress"]++
			case models.TaskStatusDone:
				stats["done"]++
			}
		}

		overview = append(overview, map[string]interface{}{
			"id":          project.ID,
			"name":        project.Name,
			"description": project.Description,
			"status":      project.Status,
			"task_stats":  stats,
			"created_at":  project.CreatedAt,
			"updated_at":  project.UpdatedAt,
		})
	}

	return &ResourceContent{
		URI:      "project://overview",
		MimeType: "application/json",
		Content:  overview,
	}, nil
}

func (s *EnhancedMCPServer) getActiveProjects() (*ResourceContent, error) {
	var projects []models.Project
	if err := s.db.Where("status = ?", models.ProjectStatusActive).
		Preload("Tasks", "status != ?", models.TaskStatusDone).
		Preload("Team").
		Find(&projects).Error; err != nil {
		return nil, err
	}

	return &ResourceContent{
		URI:      "project://active",
		MimeType: "application/json",
		Content:  projects,
	}, nil
}

func (s *EnhancedMCPServer) getTeamDashboard() (*ResourceContent, error) {
	var teams []models.Team
	if err := s.db.Preload("Members.User").Find(&teams).Error; err != nil {
		return nil, err
	}

	dashboard := make([]map[string]interface{}, 0, len(teams))
	for _, team := range teams {
		// Get team statistics
		var activeProjects int64
		s.db.Model(&models.Project{}).Where("team_id = ? AND status = ?", team.ID, models.ProjectStatusActive).Count(&activeProjects)

		var activeTasks int64
		s.db.Table("tasks").
			Joins("JOIN projects ON projects.id = tasks.project_id").
			Where("projects.team_id = ? AND tasks.status IN ?", team.ID, []string{"todo", "in_progress"}).
			Count(&activeTasks)

		// Get team members with their details
		var members []models.TeamMember
		s.db.Where("team_id = ?", team.ID).Find(&members)

		memberInfo := make([]map[string]interface{}, 0, len(members))
		for _, member := range members {
			var assignedTasks int64
			s.db.Model(&models.Task{}).Where("assignee_id = ? AND status != ?", member.UserID, models.TaskStatusDone).Count(&assignedTasks)

			memberInfo = append(memberInfo, map[string]interface{}{
				"user_id":        member.UserID,
				"role":           member.Role,
				"assigned_tasks": assignedTasks,
			})
		}

		dashboard = append(dashboard, map[string]interface{}{
			"team_id":         team.ID,
			"team_name":       team.Name,
			"active_projects": activeProjects,
			"active_tasks":    activeTasks,
			"members":         memberInfo,
		})
	}

	return &ResourceContent{
		URI:      "project://team-dashboard",
		MimeType: "application/json",
		Content:  dashboard,
	}, nil
}

func (s *EnhancedMCPServer) getCurrentSprints() (*ResourceContent, error) {
	var sprints []models.Sprint
	if err := s.db.Where("status = ?", models.SprintStatusActive).
		Preload("Tasks").
		Preload("Project").
		Find(&sprints).Error; err != nil {
		return nil, err
	}

	sprintInfo := make([]map[string]interface{}, 0, len(sprints))
	for _, sprint := range sprints {
		taskStats := map[string]int{
			"total": len(sprint.Tasks),
			"completed": 0,
			"in_progress": 0,
			"todo": 0,
		}

		for _, task := range sprint.Tasks {
			switch task.Status {
			case models.TaskStatusDone:
				taskStats["completed"]++
			case models.TaskStatusInProgress:
				taskStats["in_progress"]++
			case models.TaskStatusTodo:
				taskStats["todo"]++
			}
		}

		sprintInfo = append(sprintInfo, map[string]interface{}{
			"sprint_id":    sprint.ID,
			"sprint_name":  sprint.Name,
			"project_name": sprint.Project.Name,
			"goal":         sprint.Goal,
			"start_date":   sprint.StartDate,
			"end_date":     sprint.EndDate,
			"task_stats":   taskStats,
			"progress":     float64(taskStats["completed"]) / float64(taskStats["total"]) * 100,
		})
	}

	return &ResourceContent{
		URI:      "sprint://current",
		MimeType: "application/json",
		Content:  sprintInfo,
	}, nil
}

func (s *EnhancedMCPServer) getHighPriorityTasks() (*ResourceContent, error) {
	var tasks []models.Task
	if err := s.db.Where("priority IN ? AND status != ?",
		[]string{string(models.TaskPriorityHigh), string(models.TaskPriorityUrgent)},
		models.TaskStatusDone).
		Preload("Project").
		Preload("Assignee").
		Order("priority DESC, created_at ASC").
		Find(&tasks).Error; err != nil {
		return nil, err
	}

	return &ResourceContent{
		URI:      "tasks://high-priority",
		MimeType: "application/json",
		Content:  tasks,
	}, nil
}

func (s *EnhancedMCPServer) getOverdueTasks() (*ResourceContent, error) {
	var tasks []models.Task
	if err := s.db.Where("due_date < NOW() AND status != ?", models.TaskStatusDone).
		Preload("Project").
		Preload("Assignee").
		Order("due_date ASC").
		Find(&tasks).Error; err != nil {
		return nil, err
	}

	return &ResourceContent{
		URI:      "tasks://overdue",
		MimeType: "application/json",
		Content:  tasks,
	}, nil
}

func (s *EnhancedMCPServer) getUpcomingMilestones() (*ResourceContent, error) {
	var milestones []models.Milestone
	if err := s.db.Where("due_date BETWEEN NOW() AND DATE_ADD(NOW(), INTERVAL 30 DAY) AND status != ?",
		models.MilestoneStatusCompleted).
		Preload("Project").
		Preload("Tasks").
		Order("due_date ASC").
		Find(&milestones).Error; err != nil {
		return nil, err
	}

	milestoneInfo := make([]map[string]interface{}, 0, len(milestones))
	for _, milestone := range milestones {
		completedTasks := 0
		for _, task := range milestone.Tasks {
			if task.Status == models.TaskStatusDone {
				completedTasks++
			}
		}

		milestoneInfo = append(milestoneInfo, map[string]interface{}{
			"milestone_id":   milestone.ID,
			"milestone_name": milestone.Name,
			"project_name":   milestone.Project.Name,
			"due_date":       milestone.DueDate,
			"total_tasks":    len(milestone.Tasks),
			"completed_tasks": completedTasks,
			"progress":       float64(completedTasks) / float64(len(milestone.Tasks)) * 100,
		})
	}

	return &ResourceContent{
		URI:      "milestones://upcoming",
		MimeType: "application/json",
		Content:  milestoneInfo,
	}, nil
}

func (s *EnhancedMCPServer) getUnreadNotifications() (*ResourceContent, error) {
	var notifications []models.Notification
	if err := s.db.Where("read = ?", false).
		Preload("User").
		Order("created_at DESC").
		Limit(100).
		Find(&notifications).Error; err != nil {
		return nil, err
	}

	return &ResourceContent{
		URI:      "notifications://unread",
		MimeType: "application/json",
		Content:  notifications,
	}, nil
}

func (s *EnhancedMCPServer) getWeeklyAnalytics() (*ResourceContent, error) {
	analytics := map[string]interface{}{}

	// Tasks created this week
	var tasksCreated int64
	s.db.Model(&models.Task{}).
		Where("created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)").
		Count(&tasksCreated)
	analytics["tasks_created"] = tasksCreated

	// Tasks completed this week
	var tasksCompleted int64
	s.db.Model(&models.Task{}).
		Where("status = ? AND updated_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)", models.TaskStatusDone).
		Count(&tasksCompleted)
	analytics["tasks_completed"] = tasksCompleted

	// Active users this week
	var activeUsers int64
	s.db.Table("activities").
		Select("COUNT(DISTINCT user_id)").
		Where("created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)").
		Scan(&activeUsers)
	analytics["active_users"] = activeUsers

	// Time logged this week
	var hoursLogged float64
	s.db.Model(&models.TimeEntry{}).
		Select("SUM(hours)").
		Where("date >= DATE_SUB(NOW(), INTERVAL 7 DAY)").
		Scan(&hoursLogged)
	analytics["hours_logged"] = hoursLogged

	// Top contributors
	type Contributor struct {
		UserID        uint    `json:"user_id"`
		Username      string  `json:"username"`
		TasksCompleted int    `json:"tasks_completed"`
		HoursLogged   float64 `json:"hours_logged"`
	}

	var contributors []Contributor
	s.db.Table("tasks").
		Select("assignee_id as user_id, users.username, COUNT(*) as tasks_completed").
		Joins("JOIN users ON users.id = tasks.assignee_id").
		Where("tasks.status = ? AND tasks.updated_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)", models.TaskStatusDone).
		Group("assignee_id").
		Order("tasks_completed DESC").
		Limit(5).
		Scan(&contributors)

	// Add hours logged for each contributor
	for i := range contributors {
		s.db.Model(&models.TimeEntry{}).
			Select("SUM(hours)").
			Where("user_id = ? AND date >= DATE_SUB(NOW(), INTERVAL 7 DAY)", contributors[i].UserID).
			Scan(&contributors[i].HoursLogged)
	}

	analytics["top_contributors"] = contributors

	// Project progress
	type ProjectProgress struct {
		ProjectID   uint   `json:"project_id"`
		ProjectName string `json:"project_name"`
		Progress    float64 `json:"progress"`
	}

	var projectProgress []ProjectProgress
	s.db.Table("projects").
		Select("projects.id as project_id, projects.name as project_name").
		Where("projects.status = ?", models.ProjectStatusActive).
		Scan(&projectProgress)

	for i := range projectProgress {
		var totalTasks, completedTasks int64
		s.db.Model(&models.Task{}).Where("project_id = ?", projectProgress[i].ProjectID).Count(&totalTasks)
		s.db.Model(&models.Task{}).Where("project_id = ? AND status = ?", projectProgress[i].ProjectID, models.TaskStatusDone).Count(&completedTasks)

		if totalTasks > 0 {
			projectProgress[i].Progress = float64(completedTasks) / float64(totalTasks) * 100
		}
	}

	analytics["project_progress"] = projectProgress

	return &ResourceContent{
		URI:      "analytics://weekly",
		MimeType: "application/json",
		Content:  analytics,
	}, nil
}

func (s *EnhancedMCPServer) getWorkflowStates() (*ResourceContent, error) {
	var workflows []models.Workflow
	if err := s.db.Where("is_active = ?", true).
		Preload("States").
		Preload("Project").
		Find(&workflows).Error; err != nil {
		return nil, err
	}

	workflowInfo := make([]map[string]interface{}, 0, len(workflows))
	for _, workflow := range workflows {
		states := make([]map[string]interface{}, 0, len(workflow.States))
		for _, state := range workflow.States {
			states = append(states, map[string]interface{}{
				"name":  state.Name,
				"type":  state.Type,
				"order": state.Order,
			})
		}

		workflowInfo = append(workflowInfo, map[string]interface{}{
			"workflow_id":   workflow.ID,
			"workflow_name": workflow.Name,
			"project_name":  workflow.Project.Name,
			"description":   workflow.Description,
			"states":        states,
		})
	}

	return &ResourceContent{
		URI:      "workflow://states",
		MimeType: "application/json",
		Content:  workflowInfo,
	}, nil
}

// SubscribeToResource subscribes to resource updates (for real-time updates)
func (s *EnhancedMCPServer) SubscribeToResource(ctx context.Context, uri string) (<-chan *ResourceContent, error) {
	// This would implement real-time subscriptions using channels
	// For now, returning a simple implementation
	ch := make(chan *ResourceContent)

	// In a real implementation, this would monitor database changes
	// and send updates through the channel
	go func() {
		defer close(ch)
		<-ctx.Done()
	}()

	return ch, nil
}

// Helper to format resources as JSON for MCP protocol
func (s *EnhancedMCPServer) FormatResourcesJSON() ([]byte, error) {
	resources := s.ListResources()
	return json.MarshalIndent(resources, "", "  ")
}

// Helper to get resource content as JSON
func (s *EnhancedMCPServer) GetResourceJSON(ctx context.Context, uri string) ([]byte, error) {
	content, err := s.GetResource(ctx, uri)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(content, "", "  ")
}