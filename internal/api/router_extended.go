package api

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/headless-pm/headless-project-management/internal/database"
	"github.com/headless-pm/headless-project-management/internal/models"
	"github.com/headless-pm/headless-project-management/internal/service"
	"github.com/headless-pm/headless-project-management/pkg/auth"
)

func SetupExtendedRouter(router *gin.Engine, db *database.Database, vectorService *service.VectorService) {
	jwtManager := auth.NewJWTManager("your-secret-key-change-this", 24*time.Hour)

	authHandler := NewAuthHandler(db, jwtManager)
	epicHandler := NewEpicHandler(db)
	extendedHandler := NewExtendedHandler(db)

	api := router.Group("/api")
	{
		// Authentication endpoints
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)

			protected := auth.Group("")
			{
				protected.GET("/profile", authHandler.GetProfile)
				protected.PUT("/profile", authHandler.UpdateProfile)
				protected.POST("/change-password", authHandler.ChangePassword)
			}
		}

		// Epic endpoints
		epics := api.Group("/epics")
		{
			epics.POST("", epicHandler.CreateEpic)
			epics.GET("", epicHandler.ListEpics)
			epics.GET("/:id", epicHandler.GetEpic)
			epics.PUT("/:id", epicHandler.UpdateEpic)
			epics.DELETE("/:id", epicHandler.DeleteEpic)
			epics.POST("/:id/tasks", epicHandler.AssignTaskToEpic)
			epics.DELETE("/tasks/:taskId", epicHandler.RemoveTaskFromEpic)
			epics.GET("/:id/progress", epicHandler.GetEpicProgress)
			epics.GET("/project/:projectId", epicHandler.GetProjectEpics)
		}

		// Task dependency endpoints
		taskExtras := api.Group("/tasks")
		{
			taskExtras.POST("/:id/dependencies", extendedHandler.AddTaskDependency)
			taskExtras.GET("/:id/dependencies", extendedHandler.GetTaskDependencies)
		}

		// Label endpoints
		labels := api.Group("/labels")
		{
			labels.POST("", func(c *gin.Context) {
				var label models.Label
				if err := c.ShouldBindJSON(&label); err != nil {
					c.JSON(400, gin.H{"error": err.Error()})
					return
				}
				if err := db.CreateLabel(&label); err != nil {
					c.JSON(500, gin.H{"error": "Failed to create label"})
					return
				}
				c.JSON(201, label)
			})
			labels.GET("", func(c *gin.Context) {
				projectIDStr := c.Query("project_id")
				if projectIDStr == "" {
					c.JSON(400, gin.H{"error": "project_id is required"})
					return
				}
				projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
				if err != nil {
					c.JSON(400, gin.H{"error": "Invalid project_id"})
					return
				}
				labels, err := db.GetLabelsByProject(uint(projectID))
				if err != nil {
					c.JSON(500, gin.H{"error": "Failed to list labels"})
					return
				}
				c.JSON(200, labels)
			})
			labels.GET("/:id", func(c *gin.Context) {
				id := c.Param("id")
				var label models.Label
				if err := db.GetLabelByID(id, &label); err != nil {
					c.JSON(404, gin.H{"error": "Label not found"})
					return
				}
				c.JSON(200, label)
			})
			labels.PUT("/:id", func(c *gin.Context) {
				id := c.Param("id")
				var label models.Label
				if err := db.GetLabelByID(id, &label); err != nil {
					c.JSON(404, gin.H{"error": "Label not found"})
					return
				}
				if err := c.ShouldBindJSON(&label); err != nil {
					c.JSON(400, gin.H{"error": err.Error()})
					return
				}
				if err := db.UpdateLabel(&label); err != nil {
					c.JSON(500, gin.H{"error": "Failed to update label"})
					return
				}
				c.JSON(200, label)
			})
			labels.DELETE("/:id", func(c *gin.Context) {
				id := c.Param("id")
				if err := db.DeleteLabel(id); err != nil {
					c.JSON(500, gin.H{"error": "Failed to delete label"})
					return
				}
				c.JSON(204, nil)
			})
			labels.POST("/assign", func(c *gin.Context) {
				var req struct {
					TaskID  uint `json:"task_id" binding:"required"`
					LabelID uint `json:"label_id" binding:"required"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(400, gin.H{"error": err.Error()})
					return
				}
				if err := db.AssignLabelToTask(req.TaskID, req.LabelID); err != nil {
					c.JSON(500, gin.H{"error": "Failed to assign label"})
					return
				}
				c.JSON(200, gin.H{"message": "Label assigned successfully"})
			})
		}

		// Comment endpoints
		comments := api.Group("/comments")
		{
			comments.POST("", func(c *gin.Context) {
				var comment models.Comment
				if err := c.ShouldBindJSON(&comment); err != nil {
					c.JSON(400, gin.H{"error": err.Error()})
					return
				}
				if err := db.AddComment(&comment); err != nil {
					c.JSON(500, gin.H{"error": "Failed to add comment"})
					return
				}
				c.JSON(201, comment)
			})
			comments.GET("/task/:taskId", func(c *gin.Context) {
				taskIDStr := c.Param("taskId")
				taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
				if err != nil {
					c.JSON(400, gin.H{"error": "Invalid task ID"})
					return
				}
				var comments []models.Comment
				if err := db.Where("task_id = ?", taskID).Order("created_at DESC").Find(&comments).Error; err != nil {
					c.JSON(500, gin.H{"error": "Failed to get comments"})
					return
				}
				c.JSON(200, comments)
			})
		}

		// Attachment endpoints
		attachments := api.Group("/attachments")
		{
			attachments.POST("", func(c *gin.Context) {
				var attachment models.Attachment
				if err := c.ShouldBindJSON(&attachment); err != nil {
					c.JSON(400, gin.H{"error": err.Error()})
					return
				}
				if err := db.AddAttachment(&attachment); err != nil {
					c.JSON(500, gin.H{"error": "Failed to add attachment"})
					return
				}
				c.JSON(201, attachment)
			})
			attachments.GET("/task/:taskId", func(c *gin.Context) {
				taskIDStr := c.Param("taskId")
				taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
				if err != nil {
					c.JSON(400, gin.H{"error": "Invalid task ID"})
					return
				}
				var attachments []models.Attachment
				if err := db.Where("task_id = ?", taskID).Find(&attachments).Error; err != nil {
					c.JSON(500, gin.H{"error": "Failed to get attachments"})
					return
				}
				c.JSON(200, attachments)
			})
		}
	}
}