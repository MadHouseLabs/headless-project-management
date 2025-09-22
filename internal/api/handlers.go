package api

import (
	"fmt"
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
	// Support both old "id" param and new "project" param
	projectParam := c.Param("project")
	if projectParam == "" {
		projectParam = c.Param("id")
	}

	// Try to parse as ID first
	if projectID, err := strconv.ParseUint(projectParam, 10, 32); err == nil {
		project, err := h.db.GetProject(uint(projectID))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		c.JSON(http.StatusOK, project)
		return
	}

	// Otherwise, treat as name and look up the project
	var project models.Project
	err := h.db.DB.Where("name = ?", projectParam).Preload("Tasks").First(&project).Error
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
	projectID, err := h.getProjectIDFromParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var project models.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project.ID = projectID
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
	projectID, err := h.getProjectIDFromParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.DeleteProject(projectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *Handler) CreateTask(c *gin.Context) {
	var input struct {
		models.Task
		Labels []string `json:"labels"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task := input.Task
	if err := h.db.CreateTask(&task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	// Handle labels if provided
	if len(input.Labels) > 0 {
		if err := h.db.AssignLabelsToTask(task.ID, task.ProjectID, input.Labels); err != nil {
			// Log error but don't fail the request
			_ = err
		}
	}

	// Queue embedding generation
	if worker := service.GetEmbeddingWorker(); worker != nil {
		worker.QueueJob("task", task.ID)
	}

	// Reload task with labels
	taskWithLabels, _ := h.db.GetTask(task.ID)
	if taskWithLabels != nil {
		c.JSON(http.StatusCreated, taskWithLabels)
	} else {
		c.JSON(http.StatusCreated, task)
	}
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

// Project-scoped task handlers
func (h *Handler) ListProjectTasks(c *gin.Context) {
	projectID, err := h.getProjectIDFromParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var status *models.TaskStatus
	if statusStr := c.Query("status"); statusStr != "" {
		s := models.TaskStatus(statusStr)
		status = &s
	}

	tasks, err := h.db.ListTasks(&projectID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tasks"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func (h *Handler) CreateProjectTask(c *gin.Context) {
	projectID, err := h.getProjectIDFromParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var input struct {
		models.Task
		Labels []string `json:"labels"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task := input.Task
	task.ProjectID = projectID

	if err := h.db.CreateTask(&task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	// Handle labels if provided
	if len(input.Labels) > 0 {
		if err := h.db.AssignLabelsToTask(task.ID, task.ProjectID, input.Labels); err != nil {
			// Log error but don't fail the request
			_ = err
		}
	}

	// Queue embedding generation
	if worker := service.GetEmbeddingWorker(); worker != nil {
		worker.QueueJob("task", task.ID)
	}

	// Reload task with labels
	taskWithLabels, _ := h.db.GetTask(task.ID)
	if taskWithLabels != nil {
		c.JSON(http.StatusCreated, taskWithLabels)
	} else {
		c.JSON(http.StatusCreated, task)
	}
}

func (h *Handler) UpdateProjectTask(c *gin.Context) {
	projectID, err := h.getProjectIDFromParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskID, err := strconv.ParseUint(c.Param("task_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// Verify task belongs to project
	existingTask, err := h.db.GetTask(uint(taskID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if existingTask.ProjectID != projectID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Task does not belong to this project"})
		return
	}

	var input struct {
		models.Task
		Labels []string `json:"labels"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task := input.Task
	task.ID = uint(taskID)
	task.ProjectID = projectID

	if err := h.db.UpdateTask(&task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	// Handle labels if provided
	if input.Labels != nil {
		if err := h.db.AssignLabelsToTask(task.ID, projectID, input.Labels); err != nil {
			// Log error but don't fail the request
			_ = err
		}
	}

	// Queue embedding regeneration
	if worker := service.GetEmbeddingWorker(); worker != nil {
		worker.QueueJob("task", task.ID)
	}

	// Reload task with labels
	taskWithLabels, _ := h.db.GetTask(task.ID)
	if taskWithLabels != nil {
		c.JSON(http.StatusOK, taskWithLabels)
	} else {
		c.JSON(http.StatusOK, task)
	}
}

func (h *Handler) GetProjectTask(c *gin.Context) {
	projectID, err := h.getProjectIDFromParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskID, err := strconv.ParseUint(c.Param("task_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	task, err := h.db.GetTask(uint(taskID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if task.ProjectID != projectID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Task does not belong to this project"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *Handler) UpdateTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var input struct {
		models.Task
		Labels []string `json:"labels"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task := input.Task
	task.ID = uint(id)

	// Get existing task to preserve ProjectID
	existingTask, err := h.db.GetTask(task.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if err := h.db.UpdateTask(&task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	// Handle labels if provided
	if input.Labels != nil {
		if err := h.db.AssignLabelsToTask(task.ID, existingTask.ProjectID, input.Labels); err != nil {
			// Log error but don't fail the request
			_ = err
		}
	}

	// Queue embedding regeneration
	if worker := service.GetEmbeddingWorker(); worker != nil {
		worker.QueueJob("task", task.ID)
	}

	// Reload task with labels
	taskWithLabels, _ := h.db.GetTask(task.ID)
	if taskWithLabels != nil {
		c.JSON(http.StatusOK, taskWithLabels)
	} else {
		c.JSON(http.StatusOK, task)
	}
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

// Project-scoped label handlers
func (h *Handler) ListProjectLabels(c *gin.Context) {
	projectID, err := h.getProjectIDFromParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	labels, err := h.db.GetLabelsByProject(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list labels"})
		return
	}

	c.JSON(http.StatusOK, labels)
}

// Project users handler
func (h *Handler) ListProjectUsers(c *gin.Context) {
	projectID, err := h.getProjectIDFromParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get unique assignees from tasks in this project
	type User struct {
		Name string `json:"name"`
	}

	var users []User
	err = h.db.DB.Table("tasks").
		Select("DISTINCT assignee as name").
		Where("project_id = ? AND assignee IS NOT NULL AND assignee != ''", projectID).
		Scan(&users).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// Helper to get project ID from path parameter (supports both ID and name)
func (h *Handler) getProjectIDFromParam(c *gin.Context) (uint, error) {
	projectParam := c.Param("project")

	// Try to parse as ID first
	if projectID, err := strconv.ParseUint(projectParam, 10, 32); err == nil {
		return uint(projectID), nil
	}

	// Otherwise, treat as name and look up the project
	var project models.Project
	err := h.db.DB.Where("name = ?", projectParam).First(&project).Error
	if err != nil {
		return 0, fmt.Errorf("project not found: %s", projectParam)
	}

	return project.ID, nil
}

