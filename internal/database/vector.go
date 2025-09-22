package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/asg017/sqlite-vec-go-bindings/cgo"
)

// InitializeVectorExtension loads the sqlite-vec extension
func InitializeVectorExtension(db *sql.DB) error {
	// Load the vector extension
	_, err := db.Exec("SELECT load_extension('vec0', 'sqlite3_vec_init')")
	if err != nil {
		// Try alternative loading method
		_, err = db.Exec("SELECT vec_version()")
		if err != nil {
			return fmt.Errorf("failed to load sqlite-vec extension: %w", err)
		}
	}

	// Create vector tables for embeddings
	queries := []string{
		// Create virtual table for project embeddings
		`CREATE VIRTUAL TABLE IF NOT EXISTS project_vectors USING vec0(
			project_id INTEGER PRIMARY KEY,
			embedding FLOAT[1536]
		)`,

		// Create virtual table for task embeddings
		`CREATE VIRTUAL TABLE IF NOT EXISTS task_vectors USING vec0(
			task_id INTEGER PRIMARY KEY,
			embedding FLOAT[1536]
		)`,

		// Create virtual table for document embeddings
		`CREATE VIRTUAL TABLE IF NOT EXISTS document_vectors USING vec0(
			document_id INTEGER PRIMARY KEY,
			embedding FLOAT[1536]
		)`,

		// Create index for faster similarity searches
		`CREATE INDEX IF NOT EXISTS idx_embedding_entity ON embeddings(entity_type, entity_id)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to create vector table: %w", err)
		}
	}

	return nil
}

// StoreEmbedding stores a vector embedding in the database
func (db *Database) StoreEmbedding(entityType string, entityID uint, vector []float32) error {
	tx := db.Begin()
	defer tx.Rollback()

	// Store in the appropriate vector table
	var query string
	switch entityType {
	case "project":
		query = `INSERT OR REPLACE INTO project_vectors(project_id, embedding) VALUES (?, ?)`
	case "task":
		query = `INSERT OR REPLACE INTO task_vectors(task_id, embedding) VALUES (?, ?)`
	case "document":
		query = `INSERT OR REPLACE INTO document_vectors(document_id, embedding) VALUES (?, ?)`
	default:
		return fmt.Errorf("unsupported entity type: %s", entityType)
	}

	if err := tx.Exec(query, entityID, vector).Error; err != nil {
		return err
	}

	// Also store in the general embeddings table for tracking
	embedding := map[string]interface{}{
		"entity_type": entityType,
		"entity_id":   entityID,
		"dimension":   len(vector),
	}

	if err := tx.Table("embeddings").Create(embedding).Error; err != nil {
		return err
	}

	return tx.Commit().Error
}

// SearchSimilar performs vector similarity search
func (db *Database) SearchSimilar(entityType string, queryVector []float32, limit int) ([]map[string]interface{}, error) {
	var query string
	switch entityType {
	case "project":
		query = `
			SELECT
				project_id,
				vec_distance_cosine(embedding, ?) as distance
			FROM project_vectors
			WHERE embedding IS NOT NULL
			ORDER BY distance
			LIMIT ?
		`
	case "task":
		query = `
			SELECT
				task_id,
				vec_distance_cosine(embedding, ?) as distance
			FROM task_vectors
			WHERE embedding IS NOT NULL
			ORDER BY distance
			LIMIT ?
		`
	case "document":
		query = `
			SELECT
				document_id,
				vec_distance_cosine(embedding, ?) as distance
			FROM document_vectors
			WHERE embedding IS NOT NULL
			ORDER BY distance
			LIMIT ?
		`
	default:
		return nil, fmt.Errorf("unsupported entity type: %s", entityType)
	}

	rows, err := db.Raw(query, queryVector, limit).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var entityID uint
		var distance float64

		if err := rows.Scan(&entityID, &distance); err != nil {
			continue
		}

		// Convert cosine distance to similarity score (1 - distance)
		similarity := 1.0 - distance

		results = append(results, map[string]interface{}{
			"entity_id":  entityID,
			"similarity": similarity,
			"distance":   distance,
		})
	}

	return results, nil
}

// HybridSearch combines keyword search with vector similarity
func (db *Database) HybridSearch(entityType string, keywords string, queryVector []float32, limit int) ([]map[string]interface{}, error) {
	// First get keyword matches
	keywordMatches := make(map[uint]float64)

	switch entityType {
	case "task":
		var tasks []struct {
			ID uint
		}
		db.Table("tasks").
			Where("title LIKE ? OR description LIKE ?", "%"+keywords+"%", "%"+keywords+"%").
			Select("id").
			Find(&tasks)

		for _, task := range tasks {
			keywordMatches[task.ID] = 0.5 // Base keyword score
		}
	case "project":
		var projects []struct {
			ID uint
		}
		db.Table("projects").
			Where("name LIKE ? OR description LIKE ?", "%"+keywords+"%", "%"+keywords+"%").
			Select("id").
			Find(&projects)

		for _, project := range projects {
			keywordMatches[project.ID] = 0.5
		}
	}

	// Get vector similarity matches
	vectorMatches, err := db.SearchSimilar(entityType, queryVector, limit*2)
	if err != nil {
		return nil, err
	}

	// Combine scores (0.3 keyword weight, 0.7 vector weight)
	combinedScores := make(map[uint]float64)
	for _, match := range vectorMatches {
		id := match["entity_id"].(uint)
		vectorScore := match["similarity"].(float64)

		keywordScore := keywordMatches[id]
		combinedScore := (keywordScore * 0.3) + (vectorScore * 0.7)

		combinedScores[id] = combinedScore
	}

	// Sort by combined score and return top results
	var results []map[string]interface{}
	for id, score := range combinedScores {
		results = append(results, map[string]interface{}{
			"entity_id": id,
			"score":     score,
		})
	}

	// Sort and limit results
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// UpdateTaskEmbedding updates the embedding when task content changes
func (db *Database) UpdateTaskEmbedding(taskID uint, title, description string, comments []string) error {
	// Combine all text content
	textContent := title + " " + description
	for _, comment := range comments {
		textContent += " " + comment
	}

	taskEmbedding := map[string]interface{}{
		"task_id":      taskID,
		"text_content": textContent,
		"updated_at":   time.Now(),
	}

	return db.Table("task_embeddings").
		Where("task_id = ?", taskID).
		Updates(taskEmbedding).Error
}

// UpdateProjectEmbedding updates the embedding when project content changes
func (db *Database) UpdateProjectEmbedding(projectID uint, name, description string) error {
	textContent := name + " " + description

	projectEmbedding := map[string]interface{}{
		"project_id":   projectID,
		"text_content": textContent,
		"updated_at":   time.Now(),
	}

	return db.Table("project_embeddings").
		Where("project_id = ?", projectID).
		Updates(projectEmbedding).Error
}