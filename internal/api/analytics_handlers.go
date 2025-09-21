package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/headless-pm/headless-project-management/internal/database"
	"github.com/headless-pm/headless-project-management/internal/models"
)

type AnalyticsHandler struct {
	db *database.Database
}

func NewAnalyticsHandler(db *database.Database) *AnalyticsHandler {
	return &AnalyticsHandler{db: db}
}

func (h *AnalyticsHandler) GetProjectStats(c *gin.Context) {
	projectID := c.Param("id")

	stats := gin.H{}

	var taskCount int64
	h.db.Model(&models.Task{}).Where("project_id = ?", projectID).Count(&taskCount)
	stats["total_tasks"] = taskCount

	var statusCounts []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	h.db.Model(&models.Task{}).
		Select("status, COUNT(*) as count").
		Where("project_id = ?", projectID).
		Group("status").
		Scan(&statusCounts)
	stats["tasks_by_status"] = statusCounts

	var priorityCounts []struct {
		Priority string `json:"priority"`
		Count    int64  `json:"count"`
	}
	h.db.Model(&models.Task{}).
		Select("priority, COUNT(*) as count").
		Where("project_id = ?", projectID).
		Group("priority").
		Scan(&priorityCounts)
	stats["tasks_by_priority"] = priorityCounts

	var overdueTasks int64
	h.db.Model(&models.Task{}).
		Where("project_id = ? AND due_date < ? AND status != ?", projectID, time.Now(), models.TaskStatusDone).
		Count(&overdueTasks)
	stats["overdue_tasks"] = overdueTasks

	var memberCount int64
	h.db.Table("project_members").Where("project_id = ?", projectID).Count(&memberCount)
	stats["member_count"] = memberCount

	var completionRate float64
	var completedTasks int64
	h.db.Model(&models.Task{}).Where("project_id = ? AND status = ?", projectID, models.TaskStatusDone).Count(&completedTasks)
	if taskCount > 0 {
		completionRate = float64(completedTasks) / float64(taskCount) * 100
	}
	stats["completion_rate"] = completionRate

	c.JSON(http.StatusOK, stats)
}

func (h *AnalyticsHandler) GetUserStats(c *gin.Context) {
	userID := c.GetUint("userID")

	stats := gin.H{}

	var assignedTasks int64
	h.db.Model(&models.Task{}).Where("assignee_id = ?", userID).Count(&assignedTasks)
	stats["assigned_tasks"] = assignedTasks

	var completedTasks int64
	h.db.Model(&models.Task{}).
		Where("assignee_id = ? AND status = ?", userID, models.TaskStatusDone).
		Count(&completedTasks)
	stats["completed_tasks"] = completedTasks

	var createdTasks int64
	h.db.Model(&models.Task{}).Where("created_by = ?", userID).Count(&createdTasks)
	stats["created_tasks"] = createdTasks

	var overdueTasks int64
	h.db.Model(&models.Task{}).
		Where("assignee_id = ? AND due_date < ? AND status != ?", userID, time.Now(), models.TaskStatusDone).
		Count(&overdueTasks)
	stats["overdue_tasks"] = overdueTasks

	var projectCount int64
	h.db.Table("project_members").Where("user_id = ?", userID).Count(&projectCount)
	stats["projects"] = projectCount

	var teamCount int64
	h.db.Table("team_members").Where("user_id = ?", userID).Count(&teamCount)
	stats["teams"] = teamCount

	var recentActivity []models.Activity
	h.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(10).
		Find(&recentActivity)
	stats["recent_activity"] = recentActivity

	c.JSON(http.StatusOK, stats)
}

func (h *AnalyticsHandler) GetBurndownChart(c *gin.Context) {
	sprintID := c.Param("sprintId")

	var sprint models.Sprint
	if err := h.db.First(&sprint, sprintID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sprint not found"})
		return
	}

	// Sprint dates are now non-nullable
	days := int(sprint.EndDate.Sub(sprint.StartDate).Hours() / 24)
	data := make([]gin.H, days+1)

	var totalPoints int
	h.db.Model(&models.Task{}).
		Where("sprint_id = ?", sprintID).
		Select("SUM(story_points)").
		Scan(&totalPoints)

	for i := 0; i <= days; i++ {
		date := sprint.StartDate.AddDate(0, 0, i)

		var completedPoints int
		h.db.Model(&models.Task{}).
			Where("sprint_id = ? AND status = ? AND completed_at <= ?", sprintID, models.TaskStatusDone, date).
			Select("SUM(story_points)").
			Scan(&completedPoints)

		data[i] = gin.H{
			"date":           date.Format("2006-01-02"),
			"remaining":      totalPoints - completedPoints,
			"ideal":          float64(totalPoints) * (1 - float64(i)/float64(days)),
			"completed":      completedPoints,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"sprint":       sprint,
		"total_points": totalPoints,
		"data":         data,
	})
}

func (h *AnalyticsHandler) GetVelocityChart(c *gin.Context) {
	projectID := c.Query("project_id")
	limit := c.DefaultQuery("limit", "6")
	limitInt, _ := strconv.Atoi(limit)

	var sprints []models.Sprint
	query := h.db.Where("status = ?", "completed")

	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}

	query.Order("end_date DESC").Limit(limitInt).Find(&sprints)

	velocityData := make([]gin.H, len(sprints))
	for i, sprint := range sprints {
		var completedPoints int
		var plannedPoints int

		h.db.Model(&models.Task{}).
			Where("sprint_id = ?", sprint.ID).
			Select("SUM(story_points)").
			Scan(&plannedPoints)

		h.db.Model(&models.Task{}).
			Where("sprint_id = ? AND status = ?", sprint.ID, models.TaskStatusDone).
			Select("SUM(story_points)").
			Scan(&completedPoints)

		velocityData[i] = gin.H{
			"sprint_name":      sprint.Name,
			"planned_points":   plannedPoints,
			"completed_points": completedPoints,
			"end_date":        sprint.EndDate,
		}
	}

	var avgVelocity float64
	if len(velocityData) > 0 {
		total := 0
		for _, v := range velocityData {
			total += v["completed_points"].(int)
		}
		avgVelocity = float64(total) / float64(len(velocityData))
	}

	c.JSON(http.StatusOK, gin.H{
		"data":            velocityData,
		"average_velocity": avgVelocity,
	})
}

func (h *AnalyticsHandler) GetTaskDistribution(c *gin.Context) {
	projectID := c.Query("project_id")

	distribution := gin.H{}

	var userDistribution []struct {
		UserID   uint   `json:"user_id"`
		Username string `json:"username"`
		Count    int64  `json:"count"`
	}

	query := h.db.Table("tasks").
		Select("tasks.assignee_id as user_id, users.username, COUNT(*) as count").
		Joins("LEFT JOIN users ON tasks.assignee_id = users.id").
		Where("tasks.assignee_id IS NOT NULL")

	if projectID != "" {
		query = query.Where("tasks.project_id = ?", projectID)
	}

	query.Group("tasks.assignee_id, users.username").Scan(&userDistribution)
	distribution["by_assignee"] = userDistribution

	var labelDistribution []struct {
		Label string `json:"label"`
		Count int64  `json:"count"`
	}

	labelQuery := h.db.Table("task_labels").
		Select("labels.name as label, COUNT(*) as count").
		Joins("LEFT JOIN labels ON task_labels.label_id = labels.id")

	if projectID != "" {
		labelQuery = labelQuery.Joins("LEFT JOIN tasks ON task_labels.task_id = tasks.id").
			Where("tasks.project_id = ?", projectID)
	}

	labelQuery.Group("labels.name").Scan(&labelDistribution)
	distribution["by_label"] = labelDistribution

	c.JSON(http.StatusOK, distribution)
}

func (h *AnalyticsHandler) GetProductivityMetrics(c *gin.Context) {
	userID := c.Query("user_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" {
		startDate = time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	metrics := gin.H{}

	query := h.db.Model(&models.Task{}).
		Where("completed_at BETWEEN ? AND ?", startDate, endDate)

	if userID != "" {
		query = query.Where("assignee_id = ?", userID)
	}

	var tasksCompleted int64
	query.Where("status = ?", models.TaskStatusDone).Count(&tasksCompleted)
	metrics["tasks_completed"] = tasksCompleted

	var avgCompletionTime float64
	h.db.Model(&models.Task{}).
		Select("AVG(JULIANDAY(completed_at) - JULIANDAY(created_at))").
		Where("status = ? AND completed_at BETWEEN ? AND ?", models.TaskStatusDone, startDate, endDate).
		Scan(&avgCompletionTime)
	metrics["avg_completion_days"] = avgCompletionTime

	var totalHoursLogged float64
	timeQuery := h.db.Model(&models.TimeEntry{}).
		Select("SUM(duration) / 60.0").
		Where("date BETWEEN ? AND ?", startDate, endDate)

	if userID != "" {
		timeQuery = timeQuery.Where("user_id = ?", userID)
	}

	timeQuery.Scan(&totalHoursLogged)
	metrics["total_hours_logged"] = totalHoursLogged

	var dailyMetrics []gin.H
	for d := startDate; d <= endDate; {
		date, _ := time.Parse("2006-01-02", d)
		nextDate := date.AddDate(0, 0, 1)

		var dayCompleted int64
		dayQuery := h.db.Model(&models.Task{}).
			Where("completed_at BETWEEN ? AND ? AND status = ?", date, nextDate, models.TaskStatusDone)

		if userID != "" {
			dayQuery = dayQuery.Where("assignee_id = ?", userID)
		}

		dayQuery.Count(&dayCompleted)

		dailyMetrics = append(dailyMetrics, gin.H{
			"date":            d,
			"tasks_completed": dayCompleted,
		})

		d = nextDate.Format("2006-01-02")
	}

	metrics["daily_breakdown"] = dailyMetrics

	c.JSON(http.StatusOK, metrics)
}