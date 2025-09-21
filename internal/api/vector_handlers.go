package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/headless-pm/headless-project-management/internal/database"
	"github.com/headless-pm/headless-project-management/internal/models"
	"github.com/headless-pm/headless-project-management/internal/service"
)

type VectorHandler struct {
	db            *database.Database
	vectorService *service.VectorService
}

func NewVectorHandler(db *database.Database, vectorService *service.VectorService) *VectorHandler {
	return &VectorHandler{
		db:            db,
		vectorService: vectorService,
	}
}

// SemanticSearch performs AI-powered semantic search
func (h *VectorHandler) SemanticSearch(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	entityType := c.DefaultQuery("type", "all")
	limit := c.DefaultQuery("limit", "10")
	limitInt, _ := strconv.Atoi(limit)

	var results []models.SemanticSearchResult

	if entityType == "all" {
		// Search across all entity types
		taskResults, _ := h.vectorService.SemanticSearch(query, "task", limitInt/3)
		projectResults, _ := h.vectorService.SemanticSearch(query, "project", limitInt/3)
		docResults, _ := h.vectorService.SemanticSearch(query, "document", limitInt/3)

		results = append(results, taskResults...)
		results = append(results, projectResults...)
		results = append(results, docResults...)
	} else {
		searchResults, err := h.vectorService.SemanticSearch(query, entityType, limitInt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
			return
		}
		results = searchResults
	}

	c.JSON(http.StatusOK, gin.H{
		"query":   query,
		"results": results,
		"count":   len(results),
	})
}

// HybridSearch combines keyword and semantic search
func (h *VectorHandler) HybridSearch(c *gin.Context) {
	var req struct {
		Query      string `json:"query" binding:"required"`
		EntityType string `json:"entity_type"`
		Limit      int    `json:"limit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.EntityType == "" {
		req.EntityType = "task"
	}
	if req.Limit == 0 {
		req.Limit = 20
	}

	results, err := h.vectorService.HybridSearch(req.Query, req.EntityType, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Hybrid search failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"query":   req.Query,
		"results": results,
		"count":   len(results),
	})
}

// IndexProject generates embeddings for a project
func (h *VectorHandler) IndexProject(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	if err := h.vectorService.IndexProject(uint(projectID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to index project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project indexed successfully"})
}

// IndexTask generates embeddings for a task
func (h *VectorHandler) IndexTask(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	if err := h.vectorService.IndexTask(uint(taskID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to index task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task indexed successfully"})
}

// IndexDocument indexes a document for semantic search
func (h *VectorHandler) IndexDocument(c *gin.Context) {
	var doc models.DocumentEmbedding
	if err := c.ShouldBindJSON(&doc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.vectorService.IndexDocument(&doc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to index document"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Document indexed successfully"})
}

// FindSimilarTasks finds tasks similar to a given task
func (h *VectorHandler) FindSimilarTasks(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	limit := c.DefaultQuery("limit", "5")
	limitInt, _ := strconv.Atoi(limit)

	tasks, err := h.vectorService.FindSimilarTasks(uint(taskID), limitInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find similar tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id":       taskID,
		"similar_tasks": tasks,
		"count":         len(tasks),
	})
}

// RecommendTasks recommends tasks for a user
func (h *VectorHandler) RecommendTasks(c *gin.Context) {
	userID := c.GetUint("userID")
	limit := c.DefaultQuery("limit", "10")
	limitInt, _ := strconv.Atoi(limit)

	tasks, err := h.vectorService.RecommendTasks(userID, limitInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recommendations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recommendations": tasks,
		"count":          len(tasks),
	})
}

// ClusterTasks groups similar tasks
func (h *VectorHandler) ClusterTasks(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	numClusters := c.DefaultQuery("clusters", "5")
	numClustersInt, _ := strconv.Atoi(numClusters)

	clusters, err := h.vectorService.ClusterTasks(uint(projectID), numClustersInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cluster tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"project_id":   projectID,
		"clusters":     clusters,
		"num_clusters": numClustersInt,
	})
}

// BatchIndex indexes multiple entities
func (h *VectorHandler) BatchIndex(c *gin.Context) {
	var req struct {
		EntityType string `json:"entity_type" binding:"required"`
		EntityIDs  []uint `json:"entity_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	successful := 0
	failed := 0

	for _, entityID := range req.EntityIDs {
		var err error
		switch req.EntityType {
		case "project":
			err = h.vectorService.IndexProject(entityID)
		case "task":
			err = h.vectorService.IndexTask(entityID)
		default:
			failed++
			continue
		}

		if err != nil {
			failed++
		} else {
			successful++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"successful": successful,
		"failed":     failed,
		"total":      len(req.EntityIDs),
	})
}

// GetEmbedding returns the embedding vector for an entity
func (h *VectorHandler) GetEmbedding(c *gin.Context) {
	entityType := c.Param("type")
	entityID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity ID"})
		return
	}

	var embedding models.Embedding
	if err := h.db.Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		First(&embedding).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Embedding not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entity_type": entityType,
		"entity_id":   entityID,
		"dimension":   embedding.Dimension,
		"model":       embedding.Model,
		"created_at":  embedding.CreatedAt,
	})
}