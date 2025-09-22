package mcp

import (
	"time"

	"github.com/headless-pm/headless-project-management/internal/models"
)

// Analytics Tool Implementations

func (s *EnhancedMCPServer) getBurndownChart(args []byte) (*ToolResponse, error) {
	var input struct {
		SprintID uint `json:"sprint_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	var sprint models.Sprint
	if err := s.db.First(&sprint, input.SprintID).Error; err != nil {
		return ErrorResponse(ErrSprintNotFound), nil
	}

	// Get total tasks in sprint
	var totalTasks int64
	s.db.Model(&models.Task{}).Where("sprint_id = ?", input.SprintID).Count(&totalTasks)

	// Calculate ideal burndown
	totalDays := int(sprint.EndDate.Sub(sprint.StartDate).Hours() / 24)
	idealBurnRate := float64(totalTasks) / float64(totalDays)

	// Get daily progress
	type DailyProgress struct {
		Date      string  `json:"date"`
		Remaining int     `json:"remaining"`
		Completed int     `json:"completed"`
		Ideal     float64 `json:"ideal"`
	}

	dailyProgressSlice := make([]DailyProgress, 0)
	currentDate := sprint.StartDate
	dayIndex := 0

	for !currentDate.After(sprint.EndDate) && !currentDate.After(time.Now()) {
		var completed int64
		s.db.Model(&models.Task{}).
			Where("sprint_id = ? AND status = ? AND updated_at <= ?",
				input.SprintID, models.TaskStatusDone,
				currentDate.Add(24*time.Hour)).
			Count(&completed)

		idealRemaining := float64(totalTasks) - (idealBurnRate * float64(dayIndex))
		if idealRemaining < 0 {
			idealRemaining = 0
		}

		dailyProgressSlice = append(dailyProgressSlice, DailyProgress{
			Date:      currentDate.Format("2006-01-02"),
			Completed: int(completed),
			Remaining: int(totalTasks - completed),
			Ideal:     idealRemaining,
		})

		currentDate = currentDate.AddDate(0, 0, 1)
		dayIndex++
	}

	// Calculate velocity and prediction
	// Convert to anonymous struct slice for helper function
	progressForVelocity := make([]struct {
		Date      string  `json:"date"`
		Remaining int     `json:"remaining"`
		Completed int     `json:"completed"`
		Ideal     float64 `json:"ideal"`
	}, len(dailyProgressSlice))
	for i, p := range dailyProgressSlice {
		progressForVelocity[i] = struct {
			Date      string  `json:"date"`
			Remaining int     `json:"remaining"`
			Completed int     `json:"completed"`
			Ideal     float64 `json:"ideal"`
		}(p)
	}

	velocity := calculateSprintVelocity(progressForVelocity)
	predictedCompletion := predictSprintCompletion(sprint.EndDate, totalTasks, velocity)

	return SuccessResponse(map[string]interface{}{
		"sprint": map[string]interface{}{
			"id":         sprint.ID,
			"name":       sprint.Name,
			"start_date": sprint.StartDate,
			"end_date":   sprint.EndDate,
			"status":     sprint.Status,
		},
		"total_tasks":          totalTasks,
		"daily_progress":       dailyProgressSlice,
		"current_velocity":     velocity,
		"predicted_completion": predictedCompletion,
		"days_remaining":       calculateDaysRemaining(sprint.EndDate),
	}), nil
}

func (s *EnhancedMCPServer) getActivityFeed(args []byte) (*ToolResponse, error) {
	var input struct {
		ProjectID uint `json:"project_id"`
		UserID    uint `json:"user_id"`
		Limit     int  `json:"limit"`
	}
	_ = UnmarshalArgs(args, &input)

	if input.Limit == 0 || input.Limit > 100 {
		input.Limit = 50
	}

	query := s.db.DB
	if input.ProjectID > 0 {
		query = query.Where("(entity_type = 'project' AND entity_id = ?) OR "+
			"(entity_type = 'task' AND entity_id IN (SELECT id FROM tasks WHERE project_id = ?))",
			input.ProjectID, input.ProjectID)
	}
	if input.UserID > 0 {
		query = query.Where("user_id = ?", input.UserID)
	}

	var activities []models.Activity
	if err := query.
		Order("created_at DESC").
		Limit(input.Limit).
		Find(&activities).Error; err != nil {
		return ErrorResponse(err), nil
	}

	// Enrich activities with entity details
	enrichedActivities := make([]map[string]interface{}, len(activities))
	for i, activity := range activities {
		enriched := map[string]interface{}{
			"id":          activity.ID,
			"user_id":     activity.UserID,
			"action":      activity.Action,
			"entity_type": activity.EntityType,
			"entity_id":   activity.EntityID,
			"details":     activity.Details,
			"created_at":  activity.CreatedAt,
		}

		// Add entity name based on type
		switch activity.EntityType {
		case "task":
			var task models.Task
			if err := s.db.Select("title").First(&task, activity.EntityID).Error; err == nil {
				enriched["entity_name"] = task.Title
			}
		case "project":
			var project models.Project
			if err := s.db.Select("name").First(&project, activity.EntityID).Error; err == nil {
				enriched["entity_name"] = project.Name
			}
		}

		// Add user info
		var user models.User
		if err := s.db.Select("username").First(&user, activity.UserID).Error; err == nil {
			enriched["username"] = user.Username
		}

		enrichedActivities[i] = enriched
	}

	return SuccessResponse(enrichedActivities), nil
}

// Helper functions for analytics

func calculateSprintVelocity(progress []struct {
	Date      string  `json:"date"`
	Remaining int     `json:"remaining"`
	Completed int     `json:"completed"`
	Ideal     float64 `json:"ideal"`
}) float64 {
	if len(progress) < 2 {
		return 0
	}

	// Calculate average daily completion rate from last 3 days
	recentDays := 3
	if len(progress) < recentDays {
		recentDays = len(progress)
	}

	totalCompleted := 0
	for i := len(progress) - recentDays; i < len(progress); i++ {
		if i > 0 {
			dailyCompleted := progress[i].Completed - progress[i-1].Completed
			if dailyCompleted > 0 {
				totalCompleted += dailyCompleted
			}
		}
	}

	return float64(totalCompleted) / float64(recentDays)
}

func predictSprintCompletion(endDate time.Time, totalTasks int64, velocity float64) map[string]interface{} {
	if velocity <= 0 {
		return map[string]interface{}{
			"on_track":       false,
			"predicted_date": nil,
			"confidence":     "low",
		}
	}

	daysRemaining := calculateDaysRemaining(endDate)
	predictedCompletion := int(velocity * float64(daysRemaining))

	return map[string]interface{}{
		"on_track":             predictedCompletion >= int(totalTasks),
		"predicted_completion": predictedCompletion,
		"days_needed":          int(float64(totalTasks) / velocity),
		"confidence":           calculateConfidence(velocity),
	}
}

func calculateDaysRemaining(endDate time.Time) int {
	remaining := endDate.Sub(time.Now()).Hours() / 24
	if remaining < 0 {
		return 0
	}
	return int(remaining)
}

func calculateConfidence(velocity float64) string {
	if velocity > 5 {
		return "high"
	} else if velocity > 2 {
		return "medium"
	}
	return "low"
}