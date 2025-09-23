package api

import (
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/headless-pm/headless-project-management/internal/database"
	"github.com/headless-pm/headless-project-management/internal/models"
)

type WebHandler struct {
	db *database.Database
}

// RenderMarkdown converts markdown text to HTML
func RenderMarkdown(md string) template.HTML {
	// Create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(md))

	// Create HTML renderer with options
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	htmlContent := markdown.Render(doc, renderer)
	return template.HTML(htmlContent)
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
	return &WebHandler{
		db: db,
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

// ProjectOverviewPage shows the project overview with description and statistics
func (h *WebHandler) ProjectOverviewPage(c *gin.Context) {
	// Get project ID from URL
	projectID := c.Param("projectId")
	id, err := strconv.ParseUint(projectID, 10, 32)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"Error": "Invalid project ID",
		})
		return
	}

	// Get project with tasks and epics
	project, err := h.db.GetProject(uint(id))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"Error": "Project not found",
		})
		return
	}

	// Calculate statistics
	type ProjectStats struct {
		TotalTasks               int
		CompletedTasks           int
		TodoTasks                int
		InProgressTasks          int
		ReviewTasks              int
		TaskCompletionPercentage int
		TodoTasksPercentage      int
		InProgressTasksPercentage int
		ReviewTasksPercentage    int
		TotalEpics               int
		ActiveEpics              int
		CompletedEpics           int
		PlannedEpics             int
		TotalLabels              int
		HighPriorityTasks        int
		OverdueTasks             int
		TotalDependencies        int
	}

	stats := ProjectStats{}

	// Task statistics
	for _, task := range project.Tasks {
		stats.TotalTasks++
		switch task.Status {
		case models.TaskStatusDone:
			stats.CompletedTasks++
		case models.TaskStatusTodo:
			stats.TodoTasks++
		case models.TaskStatusInProgress:
			stats.InProgressTasks++
		case models.TaskStatusReview:
			stats.ReviewTasks++
		}

		if task.Priority == models.TaskPriorityHigh || task.Priority == models.TaskPriorityUrgent {
			stats.HighPriorityTasks++
		}

		if task.DueDate != nil && task.DueDate.Before(time.Now()) && task.Status != models.TaskStatusDone {
			stats.OverdueTasks++
		}
	}

	if stats.TotalTasks > 0 {
		stats.TaskCompletionPercentage = (stats.CompletedTasks * 100) / stats.TotalTasks
		stats.TodoTasksPercentage = (stats.TodoTasks * 100) / stats.TotalTasks
		stats.InProgressTasksPercentage = (stats.InProgressTasks * 100) / stats.TotalTasks
		stats.ReviewTasksPercentage = (stats.ReviewTasks * 100) / stats.TotalTasks
	}

	// Epic statistics
	for _, epic := range project.Epics {
		stats.TotalEpics++
		switch epic.Status {
		case models.EpicStatusActive:
			stats.ActiveEpics++
		case models.EpicStatusCompleted:
			stats.CompletedEpics++
		case models.EpicStatusPlanned:
			stats.PlannedEpics++
		}
	}

	// Count labels
	var labelCount int64
	h.db.Model(&models.Label{}).Where("project_id = ?", project.ID).Count(&labelCount)
	stats.TotalLabels = int(labelCount)

	// Count dependencies
	var depCount int64
	h.db.Table("task_dependencies").
		Joins("JOIN tasks ON tasks.id = task_dependencies.task_id").
		Where("tasks.project_id = ?", project.ID).
		Count(&depCount)
	stats.TotalDependencies = int(depCount)

	// Get recent tasks (last 5 updated)
	var recentTasks []models.Task
	h.db.Where("project_id = ?", project.ID).
		Order("updated_at DESC").
		Limit(5).
		Find(&recentTasks)

	// Render markdown description
	var renderedDescription template.HTML
	if project.Description != "" {
		renderedDescription = RenderMarkdown(project.Description)
	}

	// Render template
	c.HTML(http.StatusOK, "project_overview.html", gin.H{
		"Project":     project,
		"Stats":       stats,
		"RecentTasks": recentTasks,
		"RenderedDescription": renderedDescription,
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

	// Get unique assignees for this project (users who have tasks assigned)
	var assigneeUsers []models.User
	h.db.DB.Model(&models.User{}).
		Select("users.*").
		Joins("JOIN tasks ON tasks.assignee_id = users.id").
		Where("tasks.project_id = ?", project.ID).
		Group("users.id").
		Find(&assigneeUsers)

	// Get selected filters from query
	selectedLabelIDs := c.QueryArray("labels")
	selectedAssigneeIDs := c.QueryArray("assignees")

	// Get tasks for this project with filters
	filters := TaskFilters{
		Status:   c.Query("status"),
		Priority: c.Query("priority"),
		ProjectID: project.ID,
	}

	// Build query for tasks with labels and assignee user preloaded
	query := h.db.DB.Preload("Labels").Preload("AssigneeUser").Where("project_id = ?", project.ID)

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Priority != "" {
		query = query.Where("priority = ?", filters.Priority)
	}

	// Filter by labels if selected - require ALL selected labels (AND logic)
	if len(selectedLabelIDs) > 0 {
		// For AND logic: task must have ALL selected labels
		// We need to join and check for each label
		for _, labelID := range selectedLabelIDs {
			query = query.Where("EXISTS (SELECT 1 FROM task_labels WHERE task_labels.task_id = tasks.id AND task_labels.label_id = ?)", labelID)
		}
	}

	// Filter by assignees if selected
	if len(selectedAssigneeIDs) > 0 {
		query = query.Where("assignee_id IN ?", selectedAssigneeIDs)
	}

	// Get tasks
	var tasks []models.Task
	if err := query.Order("created_at DESC").Find(&tasks).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Failed to load tasks",
		})
		return
	}

	// Filter out done tasks that are older than 2 days
	twoDaysAgo := time.Now().AddDate(0, 0, -2)
	var filteredTasks []models.Task
	var archivedCount int
	for _, task := range tasks {
		// Keep task if it's not done, or if it's done but completed within the last 2 days
		if task.Status != models.TaskStatusDone ||
		   (task.CompletedAt != nil && task.CompletedAt.After(twoDaysAgo)) {
			filteredTasks = append(filteredTasks, task)
		} else {
			// Count archived tasks (done and older than 2 days)
			archivedCount++
		}
	}
	tasks = filteredTasks

	// Get dependency counts for each task
	taskDependencyCounts := make(map[uint]map[string]int)
	for _, task := range tasks {
		// Count dependencies
		var dependsOnCount int64
		h.db.DB.Model(&models.TaskDependency{}).Where("task_id = ?", task.ID).Count(&dependsOnCount)

		// Count dependent tasks
		var blockingCount int64
		h.db.DB.Model(&models.TaskDependency{}).Where("depends_on_id = ?", task.ID).Count(&blockingCount)

		taskDependencyCounts[task.ID] = map[string]int{
			"dependsOn": int(dependsOnCount),
			"blocking":  int(blockingCount),
		}
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

	// Sort tasks within each status by blocking importance and dependencies
	for status := range tasksByStatus {
		tasks := tasksByStatus[status]
		sort.Slice(tasks, func(i, j int) bool {
			// Get dependency and blocking counts for both tasks
			depCountI := 0
			depCountJ := 0
			blockingCountI := 0
			blockingCountJ := 0

			if counts, exists := taskDependencyCounts[tasks[i].ID]; exists {
				depCountI = counts["dependsOn"]
				blockingCountI = counts["blocking"]
			}
			if counts, exists := taskDependencyCounts[tasks[j].ID]; exists {
				depCountJ = counts["dependsOn"]
				blockingCountJ = counts["blocking"]
			}

			// First priority: Tasks with no dependencies (these can be started immediately)
			if depCountI == 0 && depCountJ > 0 {
				return true
			}
			if depCountJ == 0 && depCountI > 0 {
				return false
			}

			// Second priority: Tasks that are blocking more tasks (descending - more blocking first)
			if blockingCountI != blockingCountJ {
				return blockingCountI > blockingCountJ
			}

			// Third priority: Among remaining tasks, those with fewer dependencies first
			if depCountI != depCountJ {
				return depCountI < depCountJ
			}

			// Fourth priority: Sort by priority (urgent first)
			priorityOrder := map[models.TaskPriority]int{
				models.TaskPriorityUrgent: 0,
				models.TaskPriorityHigh:   1,
				models.TaskPriorityMedium: 2,
				models.TaskPriorityLow:    3,
			}

			priI := priorityOrder[tasks[i].Priority]
			priJ := priorityOrder[tasks[j].Priority]

			if priI != priJ {
				return priI < priJ
			}

			// Finally, sort by creation date (newer first)
			return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
		})
		tasksByStatus[status] = tasks
	}

	// Render template
	c.HTML(http.StatusOK, "tasks.html", gin.H{
		"Tasks":         tasks,
		"TasksByStatus": tasksByStatus,
		"Project":       project,
		"Labels":        labels,
		"AssigneeUsers": assigneeUsers,
		"SelectedLabels": selectedLabelIDs,
		"SelectedAssigneeIDs": selectedAssigneeIDs,
		"Filters":       filters,
		"DependencyCounts": taskDependencyCounts,
		"ArchivedCount": archivedCount,
		"ProjectID":     projectID,
	})
}

func (h *WebHandler) ArchivedTasksPage(c *gin.Context) {
	projectIDStr := c.Param("projectId")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"Error": "Invalid project ID",
		})
		return
	}

	// Get project
	var project models.Project
	if err := h.db.DB.First(&project, projectID).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"Error": "Project not found",
		})
		return
	}

	// Get only done tasks that are older than 2 days
	twoDaysAgo := time.Now().AddDate(0, 0, -2)
	var archivedTasks []models.Task

	query := h.db.DB.Preload("Labels").Preload("AssigneeUser").
		Where("project_id = ? AND status = ? AND completed_at < ?",
			project.ID, models.TaskStatusDone, twoDaysAgo)

	if err := query.Order("completed_at DESC").Find(&archivedTasks).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Failed to load archived tasks",
		})
		return
	}

	c.HTML(http.StatusOK, "archived_tasks.html", gin.H{
		"Project": project,
		"Tasks":   archivedTasks,
		"ProjectID": projectID,
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
		Preload("AssigneeUser").
		Preload("Creator").
		Preload("Updater").
		First(&task, taskID).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"Error": "Task not found",
		})
		return
	}

	// Get task dependencies (tasks this task depends on)
	dependencies, _ := h.db.GetTaskDependencies(task.ID)
	var dependsOnTasks []models.Task
	for _, dep := range dependencies {
		var depTask models.Task
		if err := h.db.DB.First(&depTask, dep.DependsOnID).Error; err == nil {
			dependsOnTasks = append(dependsOnTasks, depTask)
		}
	}

	// Get dependent tasks (tasks that depend on this task)
	var dependentTasks []models.Task
	var dependents []models.TaskDependency
	if err := h.db.DB.Where("depends_on_id = ?", task.ID).Find(&dependents).Error; err == nil {
		for _, dep := range dependents {
			var depTask models.Task
			if err := h.db.DB.First(&depTask, dep.TaskID).Error; err == nil {
				dependentTasks = append(dependentTasks, depTask)
			}
		}
	}

	// Get task activities (audit trail)
	activities, _ := h.db.GetTaskActivities(task.ID)

	// Render markdown description
	var renderedDescription template.HTML
	if task.Description != "" {
		renderedDescription = RenderMarkdown(task.Description)
	}

	// Render task detail template
	c.HTML(http.StatusOK, "task_detail.html", gin.H{
		"Task":                task,
		"ProjectID":           projectID,
		"RenderedDescription": renderedDescription,  // Markdown-rendered description
		"DependsOnTasks":      dependsOnTasks,       // Tasks this task depends on
		"DependentTasks":      dependentTasks,       // Tasks that depend on this task
		"Activities":          activities,           // Activity timeline
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

	// Calculate progress and counts
	plannedCount := 0
	activeCount := 0
	completedCount := 0

	for i := range epics {
		if len(epics[i].Tasks) > 0 {
			progress, _ := h.db.CalculateEpicProgress(epics[i].ID)
			epics[i].Progress = progress
		}

		switch epics[i].Status {
		case models.EpicStatusPlanned:
			plannedCount++
		case models.EpicStatusActive:
			activeCount++
		case models.EpicStatusCompleted:
			completedCount++
		}
	}

	// Render epics template
	c.HTML(http.StatusOK, "epics.html", gin.H{
		"Project": project,
		"Epics":   epics,
		"PlannedCount": plannedCount,
		"ActiveCount": activeCount,
		"CompletedCount": completedCount,
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

	// Calculate task statistics and group by status for Kanban view
	completedTasks := 0
	tasksByStatus := map[string][]models.Task{
		"todo":        []models.Task{},
		"in_progress": []models.Task{},
		"review":      []models.Task{},
		"done":        []models.Task{},
	}

	for _, task := range epic.Tasks {
		if task.Status == models.TaskStatusDone {
			completedTasks++
		}
		statusStr := string(task.Status)
		if _, ok := tasksByStatus[statusStr]; ok {
			tasksByStatus[statusStr] = append(tasksByStatus[statusStr], task)
		}
	}

	// Sort tasks within each status by priority and creation date
	for status := range tasksByStatus {
		tasks := tasksByStatus[status]
		sort.Slice(tasks, func(i, j int) bool {
			// Sort by priority (urgent first)
			priorityOrder := map[models.TaskPriority]int{
				models.TaskPriorityUrgent: 0,
				models.TaskPriorityHigh:   1,
				models.TaskPriorityMedium: 2,
				models.TaskPriorityLow:    3,
			}

			priI := priorityOrder[tasks[i].Priority]
			priJ := priorityOrder[tasks[j].Priority]

			if priI != priJ {
				return priI < priJ
			}

			// If same priority, sort by creation date (newer first)
			return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
		})
		tasksByStatus[status] = tasks
	}

	// Render markdown description
	var renderedDescription template.HTML
	if epic.Description != "" {
		renderedDescription = RenderMarkdown(epic.Description)
	}

	// Render epic detail template
	c.HTML(http.StatusOK, "epic_detail.html", gin.H{
		"Project":        project,
		"Epic":           epic,
		"TotalTasks":     len(epic.Tasks),
		"CompletedTasks": completedTasks,
		"TasksByStatus":  tasksByStatus,
		"RenderedDescription": renderedDescription,
	})
}

func parseUint(s string) uint {
	val, _ := strconv.ParseUint(s, 10, 32)
	return uint(val)
}