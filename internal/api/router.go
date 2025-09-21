package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(handler *Handler) *gin.Engine {
	router := gin.Default()

	router.Use(cors.Default())

	api := router.Group("/api")
	{
		projects := api.Group("/projects")
		{
			projects.POST("", handler.CreateProject)
			projects.GET("", handler.ListProjects)
			projects.GET("/:id", handler.GetProject)
			projects.PUT("/:id", handler.UpdateProject)
			projects.DELETE("/:id", handler.DeleteProject)
		}

		tasks := api.Group("/tasks")
		{
			tasks.POST("", handler.CreateTask)
			tasks.GET("", handler.ListTasks)
			tasks.GET("/:id", handler.GetTask)
			tasks.PUT("/:id", handler.UpdateTask)
			tasks.DELETE("/:id", handler.DeleteTask)
			tasks.POST("/:id/comments", handler.AddComment)
			tasks.POST("/:id/attachments", handler.UploadAttachment)
		}
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	return router
}