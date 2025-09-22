package api

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/headless-pm/headless-project-management/internal/database"
	"github.com/headless-pm/headless-project-management/internal/models"
)

type WebHandler struct {
	db        *database.Database
	templates *template.Template
}

type TaskPageData struct {
	Tasks         []models.Task
	TasksByStatus map[string][]models.Task
	Projects      []models.Project
	Filters       TaskFilters
}

type TaskFilters struct {
	ProjectID uint
	Status    string
	Priority  string
}

func NewWebHandler(db *database.Database) *WebHandler {
	templates := template.Must(template.ParseGlob("templates/*.html"))
	return &WebHandler{
		db:        db,
		templates: templates,
	}
}

func (h *WebHandler) ProjectsPage(c *gin.Context) {
	// Get all projects with their tasks
	var projects []models.Project
	if err := h.db.DB.Preload("Tasks").Find(&projects).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Failed to load projects",
		})
		return
	}

	// Calculate statistics
	type ProjectStats struct {
		models.Project
		TaskCount   int
		ActiveCount int
		DoneCount   int
		Progress    int
	}

	projectStats := make([]ProjectStats, len(projects))
	totalTasks := 0
	totalActive := 0
	totalDone := 0

	for i, project := range projects {
		stats := ProjectStats{
			Project:   project,
			TaskCount: len(project.Tasks),
		}

		for _, task := range project.Tasks {
			totalTasks++
			if task.Status == "in_progress" {
				stats.ActiveCount++
				totalActive++
			} else if task.Status == "done" {
				stats.DoneCount++
				totalDone++
			}
		}

		if stats.TaskCount > 0 {
			stats.Progress = (stats.DoneCount * 100) / stats.TaskCount
		}

		projectStats[i] = stats
	}

	// Render projects template
	c.HTML(http.StatusOK, "projects.html", gin.H{
		"Projects":    projectStats,
		"TotalProjects": len(projects),
		"TotalTasks":    totalTasks,
		"TotalActive":   totalActive,
		"TotalDone":     totalDone,
	})
}

func (h *WebHandler) ProjectBoardPage(c *gin.Context) {
	// Get project ID from URL
	projectID := c.Param("projectId")

	// Get project details
	var project models.Project
	if err := h.db.DB.First(&project, projectID).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"Error": "Project not found",
		})
		return
	}

	// Get all labels for this project
	var labels []models.Label
	h.db.DB.Where("project_id = ?", project.ID).Find(&labels)

	// Get unique assignees for this project
	var assignees []string
	h.db.DB.Model(&models.Task{}).
		Where("project_id = ? AND assignee IS NOT NULL AND assignee != ''", project.ID).
		Distinct("assignee").
		Pluck("assignee", &assignees)

	// Get selected filters from query
	selectedLabelIDs := c.QueryArray("labels")
	selectedAssignees := c.QueryArray("assignees")

	// Get tasks for this project with filters
	filters := TaskFilters{
		Status:   c.Query("status"),
		Priority: c.Query("priority"),
		ProjectID: project.ID,
	}

	// Build query for tasks with labels preloaded
	query := h.db.DB.Preload("Labels").Where("project_id = ?", project.ID)

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Priority != "" {
		query = query.Where("priority = ?", filters.Priority)
	}

	// Filter by labels if selected
	if len(selectedLabelIDs) > 0 {
		query = query.Joins("JOIN task_labels ON task_labels.task_id = tasks.id").
			Where("task_labels.label_id IN ?", selectedLabelIDs).
			Group("tasks.id")
	}

	// Filter by assignees if selected
	if len(selectedAssignees) > 0 {
		query = query.Where("assignee IN ?", selectedAssignees)
	}

	// Get tasks
	var tasks []models.Task
	if err := query.Order("created_at DESC").Find(&tasks).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Failed to load tasks",
		})
		return
	}

	// Group tasks by status for Kanban view
	tasksByStatus := map[string][]models.Task{
		"todo":        []models.Task{},
		"in_progress": []models.Task{},
		"review":      []models.Task{},
		"done":        []models.Task{},
	}

	for _, task := range tasks {
		statusStr := string(task.Status)
		if _, ok := tasksByStatus[statusStr]; ok {
			tasksByStatus[statusStr] = append(tasksByStatus[statusStr], task)
		}
	}

	// Render template
	c.HTML(http.StatusOK, "tasks.html", gin.H{
		"Tasks":         tasks,
		"TasksByStatus": tasksByStatus,
		"Project":       project,
		"Labels":        labels,
		"Assignees":     assignees,
		"SelectedLabels": selectedLabelIDs,
		"SelectedAssignees": selectedAssignees,
		"Filters":       filters,
	})
}

func (h *WebHandler) TaskDetailPage(c *gin.Context) {
	// Get project ID and task ID from URL
	projectID := c.Param("projectId")
	taskID := c.Param("taskId")

	// Fetch task with all related data
	var task models.Task
	if err := h.db.DB.Preload("Project").
		Preload("Comments").
		Preload("Attachments").
		Preload("Labels").
		Preload("Subtasks").
		First(&task, taskID).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"Error": "Task not found",
		})
		return
	}

	// Render task detail template
	c.HTML(http.StatusOK, "task_detail.html", gin.H{
		"Task":      task,
		"ProjectID": projectID,
	})
}

func (h *WebHandler) TasksPage(c *gin.Context) {
	// Parse filters from query params
	filters := TaskFilters{
		Status:   c.Query("status"),
		Priority: c.Query("priority"),
	}

	// Parse project ID if provided
	if projectIDStr := c.Query("project_id"); projectIDStr != "" {
		if id, err := strconv.ParseUint(projectIDStr, 10, 32); err == nil {
			filters.ProjectID = uint(id)
		}
	}

	// Build query for tasks
	query := h.db.DB.Preload("Project")

	if filters.ProjectID > 0 {
		query = query.Where("project_id = ?", filters.ProjectID)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Priority != "" {
		query = query.Where("priority = ?", filters.Priority)
	}

	// Get tasks
	var tasks []models.Task
	if err := query.Order("created_at DESC").Find(&tasks).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Failed to load tasks",
		})
		return
	}

	// Get all projects for filter dropdown
	var projects []models.Project
	if err := h.db.DB.Find(&projects).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Failed to load projects",
		})
		return
	}

	// Group tasks by status for Kanban view
	tasksByStatus := map[string][]models.Task{
		"todo":        []models.Task{},
		"in_progress": []models.Task{},
		"review":      []models.Task{},
		"done":        []models.Task{},
	}

	for _, task := range tasks {
		statusStr := string(task.Status)
		if _, ok := tasksByStatus[statusStr]; ok {
			tasksByStatus[statusStr] = append(tasksByStatus[statusStr], task)
		}
	}

	// Render template
	c.HTML(http.StatusOK, "tasks.html", TaskPageData{
		Tasks:         tasks,
		TasksByStatus: tasksByStatus,
		Projects:      projects,
		Filters:       filters,
	})
}

func (h *WebHandler) EpicsPage(c *gin.Context) {
	// Get project ID from URL
	projectID := c.Param("projectId")

	// Get project details
	var project models.Project
	if err := h.db.DB.First(&project, projectID).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"Error": "Project not found",
		})
		return
	}

	// Get epics for this project
	epics, err := h.db.GetEpicsByProject(project.ID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Failed to load epics",
		})
		return
	}

	// Calculate progress for each epic
	for i := range epics {
		if len(epics[i].Tasks) > 0 {
			progress, _ := h.db.CalculateEpicProgress(epics[i].ID)
			epics[i].Progress = progress
		}
	}

	// Render epics template
	c.HTML(http.StatusOK, "epics.html", gin.H{
		"Project": project,
		"Epics":   epics,
	})
}

func (h *WebHandler) EpicDetailPage(c *gin.Context) {
	// Get project ID and epic ID from URL
	projectID := c.Param("projectId")
	epicID := c.Param("epicId")

	// Get project details
	var project models.Project
	if err := h.db.DB.First(&project, projectID).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"Error": "Project not found",
		})
		return
	}

	// Get epic details with tasks
	epic, err := h.db.GetEpic(parseUint(epicID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"Error": "Epic not found",
		})
		return
	}

	// Calculate progress
	if len(epic.Tasks) > 0 {
		progress, _ := h.db.CalculateEpicProgress(epic.ID)
		epic.Progress = progress
	}

	// Calculate task statistics
	completedTasks := 0
	for _, task := range epic.Tasks {
		if task.Status == models.TaskStatusDone {
			completedTasks++
		}
	}

	// Render epic detail template
	c.HTML(http.StatusOK, "epic_detail.html", gin.H{
		"Project":        project,
		"Epic":           epic,
		"TotalTasks":     len(epic.Tasks),
		"CompletedTasks": completedTasks,
	})
}

func parseUint(s string) uint {
	val, _ := strconv.ParseUint(s, 10, 32)
	return uint(val)
}