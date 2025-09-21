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

type NotificationHandler struct {
	db *database.Database
}

func NewNotificationHandler(db *database.Database) *NotificationHandler {
	return &NotificationHandler{db: db}
}

func (h *NotificationHandler) CreateNotification(userID uint, notifType, title, message string, data interface{}) error {
	dataJSON, _ := json.Marshal(data)

	notification := &models.Notification{
		UserID:    userID,
		Type:      notifType,
		Title:     title,
		Message:   message,
		Data:      string(dataJSON),
		CreatedAt: time.Now(),
	}

	return h.db.Create(notification).Error
}

func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID := c.GetUint("userID")
	isRead := c.Query("is_read")
	limit := c.DefaultQuery("limit", "50")

	var notifications []models.Notification
	query := h.db.Where("user_id = ?", userID)

	if isRead != "" {
		query = query.Where("is_read = ?", isRead == "true")
	}

	limitInt, _ := strconv.Atoi(limit)
	if err := query.Order("created_at DESC").Limit(limitInt).Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notifications"})
		return
	}

	c.JSON(http.StatusOK, notifications)
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID := c.GetUint("userID")
	notificationID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	now := time.Now()
	if err := h.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := c.GetUint("userID")

	now := time.Now()
	if err := h.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notifications as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}

func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID := c.GetUint("userID")

	var count int64
	if err := h.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get unread count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}

func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	userID := c.GetUint("userID")
	notificationID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := h.db.Where("id = ? AND user_id = ?", notificationID, userID).
		Delete(&models.Notification{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notification"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

type ActivityHandler struct {
	db *database.Database
	nh *NotificationHandler
}

func NewActivityHandler(db *database.Database, nh *NotificationHandler) *ActivityHandler {
	return &ActivityHandler{db: db, nh: nh}
}

func (h *ActivityHandler) LogActivity(userID uint, action, entityType string, entityID uint, projectID, taskID *uint, details string, c *gin.Context) {
	activity := &models.Activity{
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		ProjectID:  projectID,
		TaskID:     taskID,
		Details:    details,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
		CreatedAt:  time.Now(),
	}

	h.db.Create(activity)

	h.notifyRelevantUsers(activity)
}

func (h *ActivityHandler) notifyRelevantUsers(activity *models.Activity) {
	switch activity.EntityType {
	case "task":
		if activity.TaskID != nil {
			var task models.Task
			if err := h.db.Preload("Watchers").First(&task, *activity.TaskID).Error; err == nil {
				for _, watcher := range task.Watchers {
					if watcher.ID != activity.UserID {
						h.nh.CreateNotification(
							watcher.ID,
							"task_update",
							"Task Updated",
							activity.Details,
							map[string]interface{}{
								"task_id":    activity.TaskID,
								"project_id": activity.ProjectID,
							},
						)
					}
				}
			}
		}
	case "comment":
		if activity.TaskID != nil {
			var task models.Task
			if err := h.db.First(&task, *activity.TaskID).Error; err == nil && task.AssigneeID != nil {
				if *task.AssigneeID != activity.UserID {
					h.nh.CreateNotification(
						*task.AssigneeID,
						"new_comment",
						"New Comment",
						activity.Details,
						map[string]interface{}{
							"task_id": activity.TaskID,
						},
					)
				}
			}
		}
	}
}

func (h *ActivityHandler) GetActivities(c *gin.Context) {
	projectIDStr := c.Query("project_id")
	taskIDStr := c.Query("task_id")
	userIDStr := c.Query("user_id")
	limit := c.DefaultQuery("limit", "50")

	var activities []models.Activity
	query := h.db.DB

	if projectIDStr != "" {
		query = query.Where("project_id = ?", projectIDStr)
	}
	if taskIDStr != "" {
		query = query.Where("task_id = ?", taskIDStr)
	}
	if userIDStr != "" {
		query = query.Where("user_id = ?", userIDStr)
	}

	limitInt, _ := strconv.Atoi(limit)
	if err := query.Order("created_at DESC").
		Limit(limitInt).
		Preload("User").
		Preload("Project").
		Preload("Task").
		Find(&activities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get activities"})
		return
	}

	c.JSON(http.StatusOK, activities)
}

func (h *ActivityHandler) GetProjectActivityFeed(c *gin.Context) {
	projectID := c.Param("id")
	limit := c.DefaultQuery("limit", "50")

	var activities []models.Activity
	limitInt, _ := strconv.Atoi(limit)

	if err := h.db.Where("project_id = ?", projectID).
		Order("created_at DESC").
		Limit(limitInt).
		Preload("User").
		Preload("Task").
		Find(&activities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get activity feed"})
		return
	}

	c.JSON(http.StatusOK, activities)
}

func (h *ActivityHandler) GetUserActivityFeed(c *gin.Context) {
	userID := c.GetUint("userID")

	var projectIDs []uint
	h.db.Model(&models.Project{}).
		Joins("JOIN project_members ON projects.id = project_members.project_id").
		Where("project_members.user_id = ?", userID).
		Pluck("projects.id", &projectIDs)

	var activities []models.Activity
	if err := h.db.Where("project_id IN ?", projectIDs).
		Order("created_at DESC").
		Limit(100).
		Preload("User").
		Preload("Project").
		Preload("Task").
		Find(&activities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get activity feed"})
		return
	}

	c.JSON(http.StatusOK, activities)
}