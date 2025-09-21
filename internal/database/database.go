package database

import (
	"fmt"
	"path/filepath"

	"github.com/headless-pm/headless-project-management/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	*gorm.DB
}

func NewDatabase(dataDir string) (*Database, error) {
	dbPath := filepath.Join(dataDir, "db", "projects.db")

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Team{},
		&models.TeamMember{},
		&models.Session{},
		&models.RefreshToken{},
		&models.Project{},
		&models.Task{},
		&models.Comment{},
		&models.Attachment{},
		&models.Label{},
		&models.Milestone{},
		&models.Sprint{},
		&models.Workflow{},
		&models.WorkflowState{},
		&models.CustomField{},
		&models.FieldValue{},
		&models.TaskDependency{},
		&models.TimeEntry{},
		&models.Activity{},
		&models.Notification{},
		&models.Webhook{},
		&models.Embedding{},
		&models.ProjectEmbedding{},
		&models.TaskEmbedding{},
		&models.DocumentEmbedding{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Initialize vector extension for SQLite
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get SQL DB: %w", err)
	}

	// Try to initialize vector extension (will fail gracefully if not available)
	_ = InitializeVectorExtension(sqlDB)

	return &Database{DB: db}, nil
}

func (db *Database) CreateProject(project *models.Project) error {
	return db.Create(project).Error
}

func (db *Database) GetProject(id uint) (*models.Project, error) {
	var project models.Project
	err := db.Preload("Tasks").First(&project, id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (db *Database) ListProjects(status *models.ProjectStatus) ([]models.Project, error) {
	var projects []models.Project
	query := db.DB
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	err := query.Find(&projects).Error
	return projects, err
}

func (db *Database) UpdateProject(project *models.Project) error {
	return db.Save(project).Error
}

func (db *Database) DeleteProject(id uint) error {
	return db.Delete(&models.Project{}, id).Error
}

func (db *Database) CreateTask(task *models.Task) error {
	return db.Create(task).Error
}

func (db *Database) GetTask(id uint) (*models.Task, error) {
	var task models.Task
	err := db.Preload("Project").
		Preload("Subtasks").
		Preload("Comments").
		Preload("Attachments").
		Preload("Labels").
		First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (db *Database) ListTasks(projectID *uint, status *models.TaskStatus) ([]models.Task, error) {
	var tasks []models.Task
	query := db.DB

	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	query = query.Where("parent_id IS NULL")

	err := query.Preload("Subtasks").Find(&tasks).Error
	return tasks, err
}

func (db *Database) UpdateTask(task *models.Task) error {
	return db.Save(task).Error
}

func (db *Database) DeleteTask(id uint) error {
	return db.Delete(&models.Task{}, id).Error
}

func (db *Database) AddComment(comment *models.Comment) error {
	return db.Create(comment).Error
}

func (db *Database) AddAttachment(attachment *models.Attachment) error {
	return db.Create(attachment).Error
}

func (db *Database) CreateLabel(label *models.Label) error {
	return db.Create(label).Error
}

func (db *Database) AssignLabelToTask(taskID uint, labelID uint) error {
	var task models.Task
	var label models.Label

	if err := db.First(&task, taskID).Error; err != nil {
		return err
	}
	if err := db.First(&label, labelID).Error; err != nil {
		return err
	}

	return db.Model(&task).Association("Labels").Append(&label)
}