package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/headless-pm/headless-project-management/internal/database"
	"github.com/headless-pm/headless-project-management/internal/models"
	"github.com/headless-pm/headless-project-management/pkg/embeddings"
)

type VectorService struct {
	db       *database.Database
	provider embeddings.EmbeddingProvider
	chunker  *embeddings.TextChunker
}

func NewVectorService(db *database.Database, provider embeddings.EmbeddingProvider) *VectorService {
	return &VectorService{
		db:       db,
		provider: provider,
		chunker:  embeddings.NewTextChunker(512, 50),
	}
}

// IndexProject generates and stores embeddings for a project
func (s *VectorService) IndexProject(projectID uint) error {
	var project models.Project
	if err := s.db.Preload("Tasks").First(&project, projectID).Error; err != nil {
		return err
	}

	// Combine project text
	text := fmt.Sprintf("%s %s", project.Name, project.Description)

	// Add task titles
	for _, task := range project.Tasks {
		text += " " + task.Title
	}

	// Generate embedding
	embedding, err := s.provider.GenerateEmbedding(text)
	if err != nil {
		return err
	}

	// Store embedding
	return s.db.StoreEmbedding("project", projectID, embedding)
}

// IndexTask generates and stores embeddings for a task
func (s *VectorService) IndexTask(taskID uint) error {
	var task models.Task
	if err := s.db.Preload("Comments").First(&task, taskID).Error; err != nil {
		return err
	}

	// Combine task text
	text := fmt.Sprintf("%s %s", task.Title, task.Description)

	// Add comments
	for _, comment := range task.Comments {
		text += " " + comment.Content
	}

	// Generate embedding
	embedding, err := s.provider.GenerateEmbedding(text)
	if err != nil {
		return err
	}

	// Store embedding
	return s.db.StoreEmbedding("task", taskID, embedding)
}

// IndexDocument indexes a document for semantic search
func (s *VectorService) IndexDocument(doc *models.DocumentEmbedding) error {
	// Chunk the document if it's large
	chunks := s.chunker.ChunkText(doc.Content)

	for i, chunk := range chunks {
		// Generate embedding for chunk
		embedding, err := s.provider.GenerateEmbedding(chunk)
		if err != nil {
			return err
		}

		// Create document chunk
		chunkDoc := &models.DocumentEmbedding{
			ProjectID:  doc.ProjectID,
			Title:      doc.Title,
			Content:    chunk,
			ChunkIndex: i,
			ChunkSize:  len(chunk),
			Metadata:   doc.Metadata,
		}

		// Store document
		if err := s.db.Create(chunkDoc).Error; err != nil {
			return err
		}

		// Store embedding
		if err := s.db.StoreEmbedding("document", chunkDoc.ID, embedding); err != nil {
			return err
		}
	}

	return nil
}

// SemanticSearch performs semantic search across entities
func (s *VectorService) SemanticSearch(query string, entityType string, limit int) ([]models.SemanticSearchResult, error) {
	// Generate query embedding
	queryEmbedding, err := s.provider.GenerateEmbedding(query)
	if err != nil {
		return nil, err
	}

	// Search similar vectors
	matches, err := s.db.SearchSimilar(entityType, queryEmbedding, limit)
	if err != nil {
		return nil, err
	}

	// Convert to search results
	results := make([]models.SemanticSearchResult, 0, len(matches))

	for _, match := range matches {
		entityID := match["entity_id"].(uint)
		similarity := match["similarity"].(float64)

		result := models.SemanticSearchResult{
			EntityType: entityType,
			EntityID:   entityID,
			Score:      similarity,
		}

		// Fetch entity details
		switch entityType {
		case "project":
			var project models.Project
			if err := s.db.First(&project, entityID).Error; err == nil {
				result.Title = project.Name
				result.Content = project.Description
			}
		case "task":
			var task models.Task
			if err := s.db.First(&task, entityID).Error; err == nil {
				result.Title = task.Title
				result.Content = task.Description
				result.Metadata = map[string]interface{}{
					"project_id": task.ProjectID,
					"status":     task.Status,
					"priority":   task.Priority,
				}
			}
		case "document":
			var doc models.DocumentEmbedding
			if err := s.db.First(&doc, entityID).Error; err == nil {
				result.Title = doc.Title
				result.Content = doc.Content

				var metadata map[string]interface{}
				json.Unmarshal([]byte(doc.Metadata), &metadata)
				result.Metadata = metadata
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// HybridSearch combines keyword and semantic search
func (s *VectorService) HybridSearch(query string, entityType string, limit int) ([]models.SemanticSearchResult, error) {
	// Generate query embedding
	queryEmbedding, err := s.provider.GenerateEmbedding(query)
	if err != nil {
		return nil, err
	}

	// Perform hybrid search
	matches, err := s.db.HybridSearch(entityType, query, queryEmbedding, limit)
	if err != nil {
		return nil, err
	}

	// Convert to search results
	results := make([]models.SemanticSearchResult, 0, len(matches))

	for _, match := range matches {
		entityID := match["entity_id"].(uint)
		score := match["score"].(float64)

		result := models.SemanticSearchResult{
			EntityType: entityType,
			EntityID:   entityID,
			Score:      score,
		}

		// Fetch entity details (same as SemanticSearch)
		switch entityType {
		case "task":
			var task models.Task
			if err := s.db.First(&task, entityID).Error; err == nil {
				result.Title = task.Title
				result.Content = task.Description
			}
		case "project":
			var project models.Project
			if err := s.db.First(&project, entityID).Error; err == nil {
				result.Title = project.Name
				result.Content = project.Description
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// FindSimilarTasks finds tasks similar to a given task
func (s *VectorService) FindSimilarTasks(taskID uint, limit int) ([]models.Task, error) {
	// Get the task's text
	var task models.Task
	if err := s.db.First(&task, taskID).Error; err != nil {
		return nil, err
	}

	// Generate embedding for the task
	text := fmt.Sprintf("%s %s", task.Title, task.Description)
	embedding, err := s.provider.GenerateEmbedding(text)
	if err != nil {
		return nil, err
	}

	// Search for similar tasks
	matches, err := s.db.SearchSimilar("task", embedding, limit+1)
	if err != nil {
		return nil, err
	}

	// Fetch task details (excluding the source task)
	var similarTasks []models.Task
	for _, match := range matches {
		matchTaskID := match["entity_id"].(uint)
		if matchTaskID == taskID {
			continue
		}

		var similarTask models.Task
		if err := s.db.Preload("Project").First(&similarTask, matchTaskID).Error; err == nil {
			similarTasks = append(similarTasks, similarTask)
		}

		if len(similarTasks) >= limit {
			break
		}
	}

	return similarTasks, nil
}

// RecommendTasks recommends tasks based on user's work history
func (s *VectorService) RecommendTasks(userID uint, limit int) ([]models.Task, error) {
	// Get user's recent completed tasks
	var recentTasks []models.Task
	s.db.Where("assignee_id = ? AND status = ?", userID, models.TaskStatusDone).
		Order("updated_at DESC").
		Limit(5).
		Find(&recentTasks)

	if len(recentTasks) == 0 {
		// Return popular uncompleted tasks
		var tasks []models.Task
		s.db.Where("status != ?", models.TaskStatusDone).
			Order("priority DESC, created_at DESC").
			Limit(limit).
			Find(&tasks)
		return tasks, nil
	}

	// Build profile from recent tasks
	var profileText []string
	for _, task := range recentTasks {
		profileText = append(profileText, task.Title+" "+task.Description)
	}

	// Generate profile embedding
	profileEmbedding, err := s.provider.GenerateEmbedding(strings.Join(profileText, " "))
	if err != nil {
		return nil, err
	}

	// Find similar uncompleted tasks
	matches, err := s.db.SearchSimilar("task", profileEmbedding, limit*2)
	if err != nil {
		return nil, err
	}

	// Filter for uncompleted tasks
	var recommendations []models.Task
	for _, match := range matches {
		taskID := match["entity_id"].(uint)

		var task models.Task
		if err := s.db.Where("id = ? AND status != ? AND (assignee_id IS NULL OR assignee_id = ?)",
			taskID, models.TaskStatusDone, userID).
			Preload("Project").
			First(&task).Error; err == nil {
			recommendations = append(recommendations, task)
		}

		if len(recommendations) >= limit {
			break
		}
	}

	return recommendations, nil
}

// ClusterTasks groups similar tasks together
func (s *VectorService) ClusterTasks(projectID uint, numClusters int) (map[int][]models.Task, error) {
	// This would implement a clustering algorithm like K-means
	// For now, returning a simple implementation

	var tasks []models.Task
	if err := s.db.Where("project_id = ?", projectID).Find(&tasks).Error; err != nil {
		return nil, err
	}

	// Simple grouping by status for now
	clusters := make(map[int][]models.Task)
	for i, task := range tasks {
		clusterID := i % numClusters
		clusters[clusterID] = append(clusters[clusterID], task)
	}

	return clusters, nil
}