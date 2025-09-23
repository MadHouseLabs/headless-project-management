package main

import (
	"time"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/headless-pm/headless-project-management/internal/api"
	"github.com/headless-pm/headless-project-management/internal/auth"
	"github.com/headless-pm/headless-project-management/internal/database"
	"github.com/headless-pm/headless-project-management/internal/mcp"
	"github.com/headless-pm/headless-project-management/internal/service"
	"github.com/headless-pm/headless-project-management/internal/storage"
	"github.com/headless-pm/headless-project-management/pkg/config"
	"github.com/headless-pm/headless-project-management/pkg/embeddings"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to optional config file (env vars take precedence)")
	flag.Parse()

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.NewDatabase(cfg.Database.DataDir)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	fileStorage, err := storage.NewFileStorage(cfg.Storage.UploadDir)
	if err != nil {
		log.Fatalf("Failed to initialize file storage: %v", err)
	}

	// Initialize embedding provider and worker
	var embeddingProvider embeddings.EmbeddingProvider
	switch cfg.Embedding.Provider {
	case "azure_openai":
		provider, err := embeddings.NewAzureOpenAIEmbeddingProvider(
			cfg.Embedding.AzureEndpoint,
			cfg.Embedding.AzureAPIKey,
			cfg.Embedding.DeploymentName,
		)
		if err != nil {
			log.Printf("Failed to initialize Azure OpenAI embeddings, using local: %v", err)
			embeddingProvider = embeddings.NewLocalEmbeddingProvider("all-MiniLM-L6-v2", "")
		} else {
			embeddingProvider = provider
		}
	case "openai":
		embeddingProvider = embeddings.NewOpenAIEmbeddingProvider(cfg.Embedding.AzureAPIKey)
	default:
		embeddingProvider = embeddings.NewLocalEmbeddingProvider("all-MiniLM-L6-v2", "")
	}

	// Initialize vector service and worker
	vectorService := service.NewVectorService(db, embeddingProvider)
	embeddingWorker := service.InitializeEmbeddingWorker(vectorService)

	// Set up embedding callback for database operations
	db.SetEmbeddingCallback(func(entityType string, entityID uint) {
		if embeddingWorker != nil {
			embeddingWorker.QueueJob(entityType, entityID)
		}
	})

	router := gin.Default()
	router.RedirectTrailingSlash = false
	router.Use(cors.Default())

	// Load HTML templates
	router.LoadHTMLGlob("templates/*")

	// Check for admin token
	adminToken := os.Getenv("ADMIN_API_TOKEN")
	if adminToken == "" {
		log.Println("WARNING: ADMIN_API_TOKEN not set. Generating a temporary admin token...")
		adminToken = "admin-" + fmt.Sprintf("%d", time.Now().Unix())
		log.Printf("Temporary admin token: %s", adminToken)
		log.Println("Please set ADMIN_API_TOKEN environment variable for production use.")
	}

	apiHandler := api.NewHandler(db, fileStorage)
	webHandler := api.NewWebHandler(db)
	tokenHandler := api.NewTokenHandler(db)
	// Use enhanced MCP server with all features
	mcpServer := mcp.NewEnhancedMCPServer(db, embeddingProvider, embeddingWorker)

	// Serve static files (CSS)
	router.Static("/static", "./web/dist")

	// Web routes
	router.GET("/", webHandler.ProjectsPage)
	router.GET("/projects/:projectId", webHandler.ProjectOverviewPage)
	router.GET("/projects/:projectId/tasks", webHandler.ProjectBoardPage)
	router.GET("/projects/:projectId/epics", webHandler.EpicsPage)
	router.GET("/projects/:projectId/epics/:epicId", webHandler.EpicDetailPage)
	router.GET("/projects/:projectId/tasks/:taskId", webHandler.TaskDetailPage)

	api.SetupExtendedRouter(router, db, vectorService)

	// Public auth endpoints (no authentication required)
	authGroup := router.Group("/auth")
	{
		// Validate endpoint can be used to check if a token is valid
		authGroup.GET("/validate", auth.AuthMiddleware(db), tokenHandler.ValidateAPIToken)
	}

	// Admin endpoints for token management (requires admin token)
	adminGroup := router.Group("/admin")
	adminGroup.Use(auth.AuthMiddleware(db), auth.AdminOnly())
	{
		tokens := adminGroup.Group("/tokens")
		{
			tokens.POST("", tokenHandler.CreateAPIToken)
			tokens.GET("", tokenHandler.ListAPITokens)
			tokens.GET("/:id", tokenHandler.GetAPIToken)
			tokens.DELETE("/:id", tokenHandler.RevokeAPIToken)
		}
	}

	// API endpoints (require authentication)
	apiGroup := router.Group("/api")
	apiGroup.Use(auth.AuthMiddleware(db))
	{
		// Root-level project endpoints
		projects := apiGroup.Group("/projects")
		{
			projects.GET("", apiHandler.ListProjects)
			projects.POST("", apiHandler.CreateProject)

			// Project-specific endpoints (by ID or name)
			projectScope := projects.Group("/:project")
			{
				projectScope.GET("", apiHandler.GetProject)
				projectScope.PUT("", apiHandler.UpdateProject)
				projectScope.DELETE("", apiHandler.DeleteProject)

				// Project tasks
				projectScope.GET("/tasks", apiHandler.ListProjectTasks)
				projectScope.POST("/tasks", apiHandler.CreateProjectTask)
				projectScope.GET("/tasks/:task_id", apiHandler.GetProjectTask)
				projectScope.PUT("/tasks/:task_id", apiHandler.UpdateProjectTask)

				// Project labels
				projectScope.GET("/labels", apiHandler.ListProjectLabels)

				// Project users
				projectScope.GET("/users", apiHandler.ListProjectUsers)
			}
		}

		// Legacy task endpoints (kept for backward compatibility)
		tasks := apiGroup.Group("/tasks")
		{
			tasks.POST("", apiHandler.CreateTask)
			tasks.GET("", apiHandler.ListTasks)
			tasks.GET("/:id", apiHandler.GetTask)
			tasks.PUT("/:id", apiHandler.UpdateTask)
			tasks.DELETE("/:id", apiHandler.DeleteTask)
			tasks.POST("/:id/comments", apiHandler.AddComment)
			tasks.POST("/:id/attachments", apiHandler.UploadAttachment)
			tasks.GET("/:id/dependencies", apiHandler.GetTaskDependencies)
			tasks.POST("/:id/dependencies", apiHandler.AddTaskDependency)
			tasks.DELETE("/:id/dependencies/:dep_id", apiHandler.RemoveTaskDependency)
		}
	}

	// Register MCP routes at /mcp (require authentication)
	mcpGroup := router.Group("/mcp")
	mcpGroup.Use(auth.AuthMiddleware(db))
	mcpServer.RegisterRoutes(mcpGroup)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"services": gin.H{
				"api": true,
				"mcp": cfg.MCP.Enabled,
			},
		})
	})

	router.GET("/info", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":    "Headless Project Management System",
			"version": "1.0.0",
			"endpoints": gin.H{
				"api":    "/api",
				"mcp":    "/mcp",
				"health": "/health",
			},
		})
	})

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting unified server on %s", addr)
	log.Printf("API endpoints: http://%s/api (requires authentication)", addr)
	log.Printf("MCP endpoints: http://%s/mcp (requires authentication)", addr)
	log.Printf("Admin endpoints: http://%s/admin (requires admin token)", addr)
	log.Printf("Web UI: http://%s (no authentication)", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}