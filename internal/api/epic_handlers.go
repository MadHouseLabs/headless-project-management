package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/headless-pm/headless-project-management/internal/database"
	"github.com/headless-pm/headless-project-management/internal/models"
)

type EpicHandler struct {
	db *database.Database
}

func NewEpicHandler(db *database.Database) *EpicHandler {
	return &EpicHandler{db: db}
}

func (h *EpicHandler) CreateEpic(c *gin.Context) {
	var epic models.Epic
	if err := c.ShouldBindJSON(&epic); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.CreateEpic(&epic); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, epic)
}

func (h *EpicHandler) GetEpic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid epic ID"})
		return
	}

	epic, err := h.db.GetEpic(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Epic not found"})
		return
	}

	// Calculate progress
	if len(epic.Tasks) > 0 {
		progress, _ := h.db.CalculateEpicProgress(epic.ID)
		epic.Progress = progress
	}

	c.JSON(http.StatusOK, epic)
}

func (h *EpicHandler) ListEpics(c *gin.Context) {
	var projectID *uint
	var status *models.EpicStatus

	if pid := c.Query("project_id"); pid != "" {
		id, err := strconv.ParseUint(pid, 10, 32)
		if err == nil {
			uid := uint(id)
			projectID = &uid
		}
	}

	if s := c.Query("status"); s != "" {
		epicStatus := models.EpicStatus(s)
		status = &epicStatus
	}

	epics, err := h.db.ListEpics(projectID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate progress for each epic
	for i := range epics {
		if len(epics[i].Tasks) > 0 {
			progress, _ := h.db.CalculateEpicProgress(epics[i].ID)
			epics[i].Progress = progress
		}
	}

	c.JSON(http.StatusOK, epics)
}

func (h *EpicHandler) UpdateEpic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid epic ID"})
		return
	}

	var epic models.Epic
	if err := c.ShouldBindJSON(&epic); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	epic.ID = uint(id)
	if err := h.db.UpdateEpic(&epic); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, epic)
}

func (h *EpicHandler) DeleteEpic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid epic ID"})
		return
	}

	if err := h.db.DeleteEpic(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *EpicHandler) GetProjectEpics(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	epics, err := h.db.GetEpicsByProject(uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate progress for each epic
	for i := range epics {
		if len(epics[i].Tasks) > 0 {
			progress, _ := h.db.CalculateEpicProgress(epics[i].ID)
			epics[i].Progress = progress
		}
	}

	c.JSON(http.StatusOK, epics)
}

func (h *EpicHandler) AssignTaskToEpic(c *gin.Context) {
	epicID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid epic ID"})
		return
	}

	var req struct {
		TaskID uint `json:"task_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.AssignTaskToEpic(req.TaskID, uint(epicID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Recalculate epic progress
	progress, _ := h.db.CalculateEpicProgress(uint(epicID))

	c.JSON(http.StatusOK, gin.H{
		"message":  "Task assigned to epic",
		"progress": progress,
	})
}

func (h *EpicHandler) RemoveTaskFromEpic(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	if err := h.db.RemoveTaskFromEpic(uint(taskID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task removed from epic"})
}

func (h *EpicHandler) GetEpicProgress(c *gin.Context) {
	epicID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid epic ID"})
		return
	}

	progress, err := h.db.CalculateEpicProgress(uint(epicID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"progress": progress})
}