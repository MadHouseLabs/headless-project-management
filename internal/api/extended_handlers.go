package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/headless-pm/headless-project-management/internal/database"
	"github.com/headless-pm/headless-project-management/internal/models"
)

type ExtendedHandler struct {
	db *database.Database
}

func NewExtendedHandler(db *database.Database) *ExtendedHandler {
	return &ExtendedHandler{db: db}
}

func (h *ExtendedHandler) CreateMilestone(c *gin.Context) {
	var milestone models.Milestone
	if err := c.ShouldBindJSON(&milestone); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Create(&milestone).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create milestone"})
		return
	}

	c.JSON(http.StatusCreated, milestone)
}

func (h *ExtendedHandler) ListMilestones(c *gin.Context) {
	projectID := c.Query("project_id")

	var milestones []models.Milestone
	query := h.db.DB

	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}

	if err := query.Preload("Tasks").Find(&milestones).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list milestones"})
		return
	}

	c.JSON(http.StatusOK, milestones)
}

func (h *ExtendedHandler) CreateSprint(c *gin.Context) {
	var sprint models.Sprint
	if err := c.ShouldBindJSON(&sprint); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Create(&sprint).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create sprint"})
		return
	}

	c.JSON(http.StatusCreated, sprint)
}

func (h *ExtendedHandler) ListSprints(c *gin.Context) {
	projectID := c.Query("project_id")
	status := c.Query("status")

	var sprints []models.Sprint
	query := h.db.DB

	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Preload("Tasks").Find(&sprints).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list sprints"})
		return
	}

	c.JSON(http.StatusOK, sprints)
}

func (h *ExtendedHandler) CreateWorkflow(c *gin.Context) {
	var workflow models.Workflow
	if err := c.ShouldBindJSON(&workflow); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Create(&workflow).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create workflow"})
		return
	}

	c.JSON(http.StatusCreated, workflow)
}

func (h *ExtendedHandler) GetWorkflow(c *gin.Context) {
	projectID := c.Param("projectId")

	var workflow models.Workflow
	if err := h.db.Where("project_id = ? AND is_default = ?", projectID, true).
		Preload("States").First(&workflow).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	c.JSON(http.StatusOK, workflow)
}

func (h *ExtendedHandler) LogTimeEntry(c *gin.Context) {
	userID := c.GetUint("userID")

	var entry models.TimeEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry.UserID = userID
	entry.Date = time.Now()

	if err := h.db.Create(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log time entry"})
		return
	}

	var task models.Task
	if err := h.db.First(&task, entry.TaskID).Error; err == nil {
		var totalHours float64
		h.db.Model(&models.TimeEntry{}).Where("task_id = ?", entry.TaskID).
			Select("SUM(duration)").Scan(&totalHours)

		actualHours := totalHours / 60
		h.db.Model(&task).Update("actual_hours", actualHours)
	}

	c.JSON(http.StatusCreated, entry)
}

func (h *ExtendedHandler) GetTimeEntries(c *gin.Context) {
	taskID := c.Query("task_id")
	userIDStr := c.Query("user_id")

	var entries []models.TimeEntry
	query := h.db.DB

	if taskID != "" {
		query = query.Where("task_id = ?", taskID)
	}
	if userIDStr != "" {
		query = query.Where("user_id = ?", userIDStr)
	}

	if err := query.Preload("Task").Preload("User").Find(&entries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get time entries"})
		return
	}

	c.JSON(http.StatusOK, entries)
}

func (h *ExtendedHandler) CreateCustomField(c *gin.Context) {
	var field models.CustomField
	if err := c.ShouldBindJSON(&field); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Create(&field).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create custom field"})
		return
	}

	c.JSON(http.StatusCreated, field)
}

func (h *ExtendedHandler) GetCustomFields(c *gin.Context) {
	projectID := c.Param("projectId")

	var fields []models.CustomField
	if err := h.db.Where("project_id = ?", projectID).Order("order").Find(&fields).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get custom fields"})
		return
	}

	c.JSON(http.StatusOK, fields)
}

func (h *ExtendedHandler) SetFieldValue(c *gin.Context) {
	var value models.FieldValue
	if err := c.ShouldBindJSON(&value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.FieldValue
	if err := h.db.Where("field_id = ? AND task_id = ?", value.FieldID, value.TaskID).First(&existing).Error; err == nil {
		h.db.Model(&existing).Update("value", value.Value)
		c.JSON(http.StatusOK, existing)
	} else {
		if err := h.db.Create(&value).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set field value"})
			return
		}
		c.JSON(http.StatusCreated, value)
	}
}

func (h *ExtendedHandler) AddTaskDependency(c *gin.Context) {
	taskID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req struct {
		DependsOnTaskID uint   `json:"depends_on_task_id" binding:"required"`
		Type           string `json:"type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dependency := models.TaskDependency{
		TaskID:          uint(taskID),
		DependsOnTaskID: req.DependsOnTaskID,
		Type:           req.Type,
	}

	if dependency.Type == "" {
		dependency.Type = "finish_to_start"
	}

	if err := h.db.Create(&dependency).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add dependency"})
		return
	}

	c.JSON(http.StatusCreated, dependency)
}

func (h *ExtendedHandler) GetTaskDependencies(c *gin.Context) {
	taskID := c.Param("id")

	var dependencies []models.TaskDependency
	if err := h.db.Where("task_id = ?", taskID).
		Preload("DependsOnTask").Find(&dependencies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dependencies"})
		return
	}

	c.JSON(http.StatusOK, dependencies)
}

func (h *ExtendedHandler) AddTaskWatcher(c *gin.Context) {
	userID := c.GetUint("userID")
	taskID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var task models.Task
	if err := h.db.First(&task, taskID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := h.db.Model(&task).Association("Watchers").Append(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add watcher"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Watcher added successfully"})
}

func (h *ExtendedHandler) RemoveTaskWatcher(c *gin.Context) {
	userID := c.GetUint("userID")
	taskID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var task models.Task
	if err := h.db.First(&task, taskID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := h.db.Model(&task).Association("Watchers").Delete(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove watcher"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Watcher removed successfully"})
}