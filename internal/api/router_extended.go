package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/headless-pm/headless-project-management/internal/database"
	"github.com/headless-pm/headless-project-management/internal/middleware"
	"github.com/headless-pm/headless-project-management/internal/models"
	"github.com/headless-pm/headless-project-management/internal/service"
	"github.com/headless-pm/headless-project-management/pkg/auth"
)

func SetupExtendedRouter(router *gin.Engine, db *database.Database, vectorService *service.VectorService) {
	jwtManager := auth.NewJWTManager("your-secret-key-change-this", 24*time.Hour)

	authHandler := NewAuthHandler(db, jwtManager)
	teamHandler := NewTeamHandler(db)
	extendedHandler := NewExtendedHandler(db)
	notificationHandler := NewNotificationHandler(db)
	activityHandler := NewActivityHandler(db, notificationHandler)
	webhookHandler := NewWebhookHandler(db)
	searchHandler := NewSearchHandler(db)
	analyticsHandler := NewAnalyticsHandler(db)
	vectorHandler := NewVectorHandler(db, vectorService)

	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)

			protected := auth.Group("")
			protected.Use(middleware.JWTAuth(jwtManager))
			{
				protected.GET("/profile", authHandler.GetProfile)
				protected.PUT("/profile", authHandler.UpdateProfile)
				protected.POST("/change-password", authHandler.ChangePassword)
			}
		}

		teams := api.Group("/teams")
		teams.Use(middleware.JWTAuth(jwtManager))
		{
			teams.POST("", teamHandler.CreateTeam)
			teams.GET("", teamHandler.ListTeams)
			teams.GET("/:id", teamHandler.GetTeam)
			teams.PUT("/:id", teamHandler.UpdateTeam)
			teams.DELETE("/:id", teamHandler.DeleteTeam)
			teams.POST("/:id/members", teamHandler.AddMember)
			teams.DELETE("/:id/members/:userId", teamHandler.RemoveMember)
		}

		milestones := api.Group("/milestones")
		milestones.Use(middleware.JWTAuth(jwtManager))
		{
			milestones.POST("", extendedHandler.CreateMilestone)
			milestones.GET("", extendedHandler.ListMilestones)
		}

		sprints := api.Group("/sprints")
		sprints.Use(middleware.JWTAuth(jwtManager))
		{
			sprints.POST("", extendedHandler.CreateSprint)
			sprints.GET("", extendedHandler.ListSprints)
		}

		workflows := api.Group("/workflows")
		workflows.Use(middleware.JWTAuth(jwtManager))
		{
			workflows.POST("", extendedHandler.CreateWorkflow)
			workflows.GET("/project/:projectId", extendedHandler.GetWorkflow)
		}

		customFields := api.Group("/custom-fields")
		customFields.Use(middleware.JWTAuth(jwtManager))
		{
			customFields.POST("", extendedHandler.CreateCustomField)
			customFields.GET("/project/:projectId", extendedHandler.GetCustomFields)
			customFields.POST("/value", extendedHandler.SetFieldValue)
		}

		timeEntries := api.Group("/time-entries")
		timeEntries.Use(middleware.JWTAuth(jwtManager))
		{
			timeEntries.POST("", extendedHandler.LogTimeEntry)
			timeEntries.GET("", extendedHandler.GetTimeEntries)
		}

		taskExtras := api.Group("/tasks")
		taskExtras.Use(middleware.JWTAuth(jwtManager))
		{
			taskExtras.POST("/:id/dependencies", extendedHandler.AddTaskDependency)
			taskExtras.GET("/:id/dependencies", extendedHandler.GetTaskDependencies)
			taskExtras.POST("/:id/watchers", extendedHandler.AddTaskWatcher)
			taskExtras.DELETE("/:id/watchers", extendedHandler.RemoveTaskWatcher)
		}

		notifications := api.Group("/notifications")
		notifications.Use(middleware.JWTAuth(jwtManager))
		{
			notifications.GET("", notificationHandler.GetNotifications)
			notifications.PUT("/:id/read", notificationHandler.MarkAsRead)
			notifications.PUT("/read-all", notificationHandler.MarkAllAsRead)
			notifications.GET("/unread-count", notificationHandler.GetUnreadCount)
			notifications.DELETE("/:id", notificationHandler.DeleteNotification)
		}

		activities := api.Group("/activities")
		activities.Use(middleware.JWTAuth(jwtManager))
		{
			activities.GET("", activityHandler.GetActivities)
			activities.GET("/project/:id", activityHandler.GetProjectActivityFeed)
			activities.GET("/feed", activityHandler.GetUserActivityFeed)
		}

		webhooks := api.Group("/webhooks")
		webhooks.Use(middleware.JWTAuth(jwtManager))
		{
			webhooks.POST("", webhookHandler.CreateWebhook)
			webhooks.GET("", webhookHandler.ListWebhooks)
			webhooks.GET("/:id", webhookHandler.GetWebhook)
			webhooks.PUT("/:id", webhookHandler.UpdateWebhook)
			webhooks.DELETE("/:id", webhookHandler.DeleteWebhook)
			webhooks.POST("/:id/test", webhookHandler.TestWebhook)
		}

		search := api.Group("/search")
		search.Use(middleware.OptionalAuth(jwtManager))
		{
			search.GET("", searchHandler.GlobalSearch)
			search.POST("/tasks", searchHandler.AdvancedTaskSearch)

			savedSearches := search.Group("/saved")
			savedSearches.Use(middleware.JWTAuth(jwtManager))
			{
				savedSearches.GET("", searchHandler.SavedSearches)
				savedSearches.POST("", searchHandler.SaveSearch)
				savedSearches.DELETE("/:id", searchHandler.DeleteSavedSearch)
			}
		}

		analytics := api.Group("/analytics")
		analytics.Use(middleware.JWTAuth(jwtManager))
		{
			analytics.GET("/project/:id/stats", analyticsHandler.GetProjectStats)
			analytics.GET("/user/stats", analyticsHandler.GetUserStats)
			analytics.GET("/sprint/:sprintId/burndown", analyticsHandler.GetBurndownChart)
			analytics.GET("/velocity", analyticsHandler.GetVelocityChart)
			analytics.GET("/task-distribution", analyticsHandler.GetTaskDistribution)
			analytics.GET("/productivity", analyticsHandler.GetProductivityMetrics)
		}

		labels := api.Group("/labels")
		labels.Use(middleware.JWTAuth(jwtManager))
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
				var labels []models.Label
				if err := db.Find(&labels).Error; err != nil {
					c.JSON(500, gin.H{"error": "Failed to list labels"})
					return
				}
				c.JSON(200, labels)
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

		// Vector/AI endpoints
		vectors := api.Group("/vectors")
		vectors.Use(middleware.JWTAuth(jwtManager))
		{
			vectors.GET("/search", vectorHandler.SemanticSearch)
			vectors.POST("/search/hybrid", vectorHandler.HybridSearch)
			vectors.GET("/similar/tasks/:id", vectorHandler.FindSimilarTasks)
			vectors.GET("/recommend/tasks", vectorHandler.RecommendTasks)
			vectors.GET("/cluster/project/:projectId", vectorHandler.ClusterTasks)
		}
	}
}