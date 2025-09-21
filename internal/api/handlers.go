package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/headless-pm/headless-project-management/internal/database"
	"github.com/headless-pm/headless-project-management/internal/models"
	"github.com/headless-pm/headless-project-management/internal/service"
	"github.com/headless-pm/headless-project-management/internal/storage"
)

type Handler struct {
	db      *database.Database
	storage *storage.FileStorage
}

func NewHandler(db *database.Database, storage *storage.FileStorage) *Handler {
	return &Handler{
		db:      db,
		storage: storage,
	}
}

func (h *Handler) CreateProject(c *gin.Context) {
	var project models.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.CreateProject(&project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	// Queue embedding generation
	if worker := service.GetEmbeddingWorker(); worker != nil {
		worker.QueueJob("project", project.ID)
	}

	c.JSON(http.StatusCreated, project)
}

func (h *Handler) GetProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	project, err := h.db.GetProject(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.JSON(http.StatusOK, project)
}

func (h *Handler) ListProjects(c *gin.Context) {
	statusStr := c.Query("status")
	var status *models.ProjectStatus
	if statusStr != "" {
		s := models.ProjectStatus(statusStr)
		status = &s
	}

	projects, err := h.db.ListProjects(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list projects"})
		return
	}

	c.JSON(http.StatusOK, projects)
}

func (h *Handler) UpdateProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var project models.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project.ID = uint(id)
	if err := h.db.UpdateProject(&project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}

	// Queue embedding regeneration
	if worker := service.GetEmbeddingWorker(); worker != nil {
		worker.QueueJob("project", project.ID)
	}

	c.JSON(http.StatusOK, project)
}

func (h *Handler) DeleteProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	if err := h.db.DeleteProject(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *Handler) CreateTask(c *gin.Context) {
	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.CreateTask(&task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	// Queue embedding generation
	if worker := service.GetEmbeddingWorker(); worker != nil {
		worker.QueueJob("task", task.ID)
	}

	c.JSON(http.StatusCreated, task)
}

func (h *Handler) GetTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	task, err := h.db.GetTask(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *Handler) ListTasks(c *gin.Context) {
	var projectID *uint
	if pidStr := c.Query("project_id"); pidStr != "" {
		pid, err := strconv.ParseUint(pidStr, 10, 32)
		if err == nil {
			pidUint := uint(pid)
			projectID = &pidUint
		}
	}

	var status *models.TaskStatus
	if statusStr := c.Query("status"); statusStr != "" {
		s := models.TaskStatus(statusStr)
		status = &s
	}

	tasks, err := h.db.ListTasks(projectID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tasks"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func (h *Handler) UpdateTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task.ID = uint(id)
	if err := h.db.UpdateTask(&task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	// Queue embedding regeneration
	if worker := service.GetEmbeddingWorker(); worker != nil {
		worker.QueueJob("task", task.ID)
	}

	c.JSON(http.StatusOK, task)
}

func (h *Handler) DeleteTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	if err := h.db.DeleteTask(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *Handler) AddComment(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var comment models.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment.TaskID = uint(taskID)
	if err := h.db.AddComment(&comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add comment"})
		return
	}

	// Queue embedding regeneration for task (includes comments)
	if worker := service.GetEmbeddingWorker(); worker != nil {
		worker.QueueJob("task", uint(taskID))
	}

	c.JSON(http.StatusCreated, comment)
}

func (h *Handler) UploadAttachment(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}

	task, err := h.db.GetTask(uint(taskID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	path, err := h.storage.SaveFile(file, task.ProjectID, uint(taskID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	attachment := models.Attachment{
		TaskID:   uint(taskID),
		Filename: file.Filename,
		Path:     path,
		Size:     file.Size,
		MimeType: file.Header.Get("Content-Type"),
	}

	if err := h.db.AddAttachment(&attachment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save attachment record"})
		return
	}

	c.JSON(http.StatusCreated, attachment)
}