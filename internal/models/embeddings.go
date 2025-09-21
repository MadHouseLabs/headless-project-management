package models

import (
	"time"
)

// Embedding stores vector embeddings for various entities
type Embedding struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	EntityType string    `json:"entity_type" gorm:"not null;index"` // project, task, comment, document
	EntityID   uint      `json:"entity_id" gorm:"not null;index"`
	Vector     []byte    `json:"-" gorm:"type:blob"` // Store as BLOB for sqlite-vec
	Dimension  int       `json:"dimension" gorm:"default:384"`
	Model      string    `json:"model" gorm:"default:'all-MiniLM-L6-v2'"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ProjectEmbedding extends Project with embedding support
type ProjectEmbedding struct {
	ProjectID   uint   `json:"project_id" gorm:"primaryKey"`
	Embedding   []byte `json:"-" gorm:"type:blob"`
	TextContent string `json:"text_content" gorm:"type:text"` // Concatenated searchable text
	UpdatedAt   time.Time `json:"updated_at"`
}

// TaskEmbedding extends Task with embedding support
type TaskEmbedding struct {
	TaskID      uint   `json:"task_id" gorm:"primaryKey"`
	Embedding   []byte `json:"-" gorm:"type:blob"`
	TextContent string `json:"text_content" gorm:"type:text"` // Title + Description + Comments
	UpdatedAt   time.Time `json:"updated_at"`
}

// DocumentEmbedding for knowledge base documents
type DocumentEmbedding struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ProjectID   uint      `json:"project_id" gorm:"index"`
	Title       string    `json:"title" gorm:"not null"`
	Content     string    `json:"content" gorm:"type:text"`
	Embedding   []byte    `json:"-" gorm:"type:blob"`
	ChunkIndex  int       `json:"chunk_index" gorm:"default:0"` // For large documents
	ChunkSize   int       `json:"chunk_size" gorm:"default:512"`
	Metadata    string    `json:"metadata"` // JSON metadata
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SemanticSearchResult represents a search result with similarity score
type SemanticSearchResult struct {
	EntityType string      `json:"entity_type"`
	EntityID   uint        `json:"entity_id"`
	Score      float64     `json:"score"`
	Title      string      `json:"title"`
	Content    string      `json:"content"`
	Metadata   interface{} `json:"metadata,omitempty"`
}