package api

import (
	"net/http"
	"strconv"

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

func (h *ExtendedHandler) AddTaskDependency(c *gin.Context) {
	taskID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req struct {
		DependsOnID uint   `json:"depends_on_id" binding:"required"`
		Type       string `json:"type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dependency := models.TaskDependency{
		TaskID:      uint(taskID),
		DependsOnID: req.DependsOnID,
		Type:       req.Type,
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
		Preload("DependsOn").Find(&dependencies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dependencies"})
		return
	}

	c.JSON(http.StatusOK, dependencies)
}