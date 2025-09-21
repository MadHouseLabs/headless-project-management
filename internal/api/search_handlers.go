package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/headless-pm/headless-project-management/internal/database"
	"github.com/headless-pm/headless-project-management/internal/models"
)

type SearchHandler struct {
	db *database.Database
}

func NewSearchHandler(db *database.Database) *SearchHandler {
	return &SearchHandler{db: db}
}

func (h *SearchHandler) GlobalSearch(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query required"})
		return
	}

	limit := c.DefaultQuery("limit", "20")
	limitInt, _ := strconv.Atoi(limit)

	results := gin.H{
		"projects":    h.searchProjects(query, limitInt),
		"tasks":       h.searchTasks(query, limitInt),
		"comments":    h.searchComments(query, limitInt),
		"users":       h.searchUsers(query, limitInt),
		"query":       query,
		"total_count": 0,
	}

	c.JSON(http.StatusOK, results)
}

func (h *SearchHandler) searchProjects(query string, limit int) []models.Project {
	var projects []models.Project
	h.db.Where("name LIKE ? OR description LIKE ?", "%"+query+"%", "%"+query+"%").
		Limit(limit).
		Find(&projects)
	return projects
}

func (h *SearchHandler) searchTasks(query string, limit int) []models.Task {
	var tasks []models.Task
	h.db.Where("title LIKE ? OR description LIKE ?", "%"+query+"%", "%"+query+"%").
		Preload("Project").
		Limit(limit).
		Find(&tasks)
	return tasks
}

func (h *SearchHandler) searchComments(query string, limit int) []models.Comment {
	var comments []models.Comment
	h.db.Where("content LIKE ?", "%"+query+"%").
		Preload("Task").
		Limit(limit).
		Find(&comments)
	return comments
}

func (h *SearchHandler) searchUsers(query string, limit int) []models.User {
	var users []models.User
	h.db.Where("username LIKE ? OR email LIKE ? OR first_name LIKE ? OR last_name LIKE ?",
		"%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%").
		Limit(limit).
		Find(&users)
	return users
}

func (h *SearchHandler) AdvancedTaskSearch(c *gin.Context) {
	var filters struct {
		Query       string   `json:"query"`
		ProjectIDs  []uint   `json:"project_ids"`
		Status      []string `json:"status"`
		Priority    []string `json:"priority"`
		AssigneeIDs []uint   `json:"assignee_ids"`
		Labels      []string `json:"labels"`
		HasDueDate  *bool    `json:"has_due_date"`
		IsOverdue   *bool    `json:"is_overdue"`
		CreatedBy   *uint    `json:"created_by"`
		SortBy      string   `json:"sort_by"`
		SortOrder   string   `json:"sort_order"`
		Limit       int      `json:"limit"`
		Offset      int      `json:"offset"`
	}

	if err := c.ShouldBindJSON(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if filters.Limit == 0 {
		filters.Limit = 50
	}

	query := h.db.Model(&models.Task{})

	if filters.Query != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+filters.Query+"%", "%"+filters.Query+"%")
	}

	if len(filters.ProjectIDs) > 0 {
		query = query.Where("project_id IN ?", filters.ProjectIDs)
	}

	if len(filters.Status) > 0 {
		query = query.Where("status IN ?", filters.Status)
	}

	if len(filters.Priority) > 0 {
		query = query.Where("priority IN ?", filters.Priority)
	}

	if len(filters.AssigneeIDs) > 0 {
		query = query.Where("assignee_id IN ?", filters.AssigneeIDs)
	}

	if filters.HasDueDate != nil {
		if *filters.HasDueDate {
			query = query.Where("due_date IS NOT NULL")
		} else {
			query = query.Where("due_date IS NULL")
		}
	}

	if filters.IsOverdue != nil && *filters.IsOverdue {
		query = query.Where("due_date < ? AND status != ?", time.Now(), models.TaskStatusDone)
	}

	if filters.CreatedBy != nil {
		query = query.Where("created_by = ?", *filters.CreatedBy)
	}

	if len(filters.Labels) > 0 {
		query = query.Joins("JOIN task_labels ON tasks.id = task_labels.task_id").
			Joins("JOIN labels ON task_labels.label_id = labels.id").
			Where("labels.name IN ?", filters.Labels)
	}

	sortBy := "created_at"
	if filters.SortBy != "" {
		sortBy = filters.SortBy
	}

	sortOrder := "DESC"
	if filters.SortOrder != "" {
		sortOrder = filters.SortOrder
	}

	query = query.Order(sortBy + " " + sortOrder)

	var tasks []models.Task
	var total int64

	query.Count(&total)

	if err := query.Offset(filters.Offset).
		Limit(filters.Limit).
		Preload("Project").
		Preload("AssigneeUser").
		Preload("Labels").
		Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"total": total,
		"limit": filters.Limit,
		"offset": filters.Offset,
	})
}

func (h *SearchHandler) SavedSearches(c *gin.Context) {
	userID := c.GetUint("userID")

	var searches []struct {
		ID      uint   `json:"id"`
		Name    string `json:"name"`
		Filters string `json:"filters"`
	}

	if err := h.db.Table("saved_searches").
		Where("user_id = ?", userID).
		Find(&searches).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get saved searches"})
		return
	}

	c.JSON(http.StatusOK, searches)
}

func (h *SearchHandler) SaveSearch(c *gin.Context) {
	userID := c.GetUint("userID")

	var req struct {
		Name    string          `json:"name" binding:"required"`
		Filters json.RawMessage `json:"filters" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	savedSearch := map[string]interface{}{
		"user_id":    userID,
		"name":       req.Name,
		"filters":    string(req.Filters),
		"created_at": time.Now(),
	}

	if err := h.db.Table("saved_searches").Create(savedSearch).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save search"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Search saved successfully"})
}

func (h *SearchHandler) DeleteSavedSearch(c *gin.Context) {
	userID := c.GetUint("userID")
	searchID := c.Param("id")

	if err := h.db.Table("saved_searches").
		Where("id = ? AND user_id = ?", searchID, userID).
		Delete(nil).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete saved search"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}