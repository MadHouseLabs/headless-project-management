package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/headless-pm/headless-project-management/internal/api"
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

	router := gin.Default()
	router.Use(cors.Default())

	apiHandler := api.NewHandler(db, fileStorage)
	// Use enhanced MCP server with all features
	mcpServer := mcp.NewEnhancedMCPServer(db, embeddingProvider, embeddingWorker)

	api.SetupExtendedRouter(router, db, vectorService)

	apiGroup := router.Group("/api")
	{
		projects := apiGroup.Group("/projects")
		{
			projects.POST("", apiHandler.CreateProject)
			projects.GET("", apiHandler.ListProjects)
			projects.GET("/:id", apiHandler.GetProject)
			projects.PUT("/:id", apiHandler.UpdateProject)
			projects.DELETE("/:id", apiHandler.DeleteProject)
		}

		tasks := apiGroup.Group("/tasks")
		{
			tasks.POST("", apiHandler.CreateTask)
			tasks.GET("", apiHandler.ListTasks)
			tasks.GET("/:id", apiHandler.GetTask)
			tasks.PUT("/:id", apiHandler.UpdateTask)
			tasks.DELETE("/:id", apiHandler.DeleteTask)
			tasks.POST("/:id/comments", apiHandler.AddComment)
			tasks.POST("/:id/attachments", apiHandler.UploadAttachment)
		}

	}

	// Register MCP routes at /mcp
	mcpGroup := router.Group("/mcp")
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

	router.GET("/", func(c *gin.Context) {
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
	log.Printf("API endpoints: http://%s/api", addr)
	log.Printf("MCP endpoints: http://%s/mcp", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}