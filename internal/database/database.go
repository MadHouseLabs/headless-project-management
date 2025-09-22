package database

import (
	"fmt"
	"os"
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

	// Ensure the db directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.AutoMigrate(
		// Core entities
		&models.User{},
		&models.Project{},
		&models.Epic{},
		&models.Task{},
		&models.TaskDependency{},
		&models.Label{},
		&models.Comment{},
		&models.Attachment{},

		// Auth entities
		&models.Session{},
		&models.RefreshToken{},
		&models.APIToken{},

		// Embedding entities (keep for semantic search)
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

func (db *Database) GetLabelByID(id string, label *models.Label) error {
	return db.First(label, id).Error
}

func (db *Database) UpdateLabel(label *models.Label) error {
	return db.Save(label).Error
}

func (db *Database) DeleteLabel(id string) error {
	return db.Delete(&models.Label{}, id).Error
}

func (db *Database) GetOrCreateLabel(projectID uint, name string, color string) (*models.Label, error) {
	var label models.Label
	err := db.Where("project_id = ? AND name = ?", projectID, name).First(&label).Error
	if err != nil {
		// Create label if it doesn't exist
		if color == "" {
			// Generate a dark color if not provided
			colors := []string{
				"#DC2626", "#059669", "#2563EB", "#7C3AED",
				"#EA580C", "#0891B2", "#4F46E5", "#BE123C",
				"#15803D", "#B91C1C", "#0E7490", "#6B21A8",
				"#C2410C", "#1E40AF", "#86198F", "#166534",
			}
			hash := 0
			for _, c := range name {
				hash = hash*31 + int(c)
			}
			color = colors[hash%len(colors)]
		}

		label = models.Label{
			ProjectID: projectID,
			Name:      name,
			Color:     color,
		}
		if err := db.Create(&label).Error; err != nil {
			return nil, err
		}
	}
	return &label, nil
}

func (db *Database) GetLabelsByProject(projectID uint) ([]models.Label, error) {
	var labels []models.Label
	err := db.Where("project_id = ?", projectID).Find(&labels).Error
	return labels, err
}

func (db *Database) AssignLabelsToTask(taskID uint, projectID uint, labelNames []string) error {
	// First, get the task
	var task models.Task
	if err := db.First(&task, taskID).Error; err != nil {
		return err
	}

	// Clear existing labels
	if err := db.Model(&task).Association("Labels").Clear(); err != nil {
		return err
	}

	// Assign new labels
	for _, labelName := range labelNames {
		if labelName == "" {
			continue
		}

		label, err := db.GetOrCreateLabel(projectID, labelName, "")
		if err != nil {
			continue
		}

		if err := db.Model(&task).Association("Labels").Append(label); err != nil {
			continue
		}
	}

	return nil
}

// Epic CRUD methods
func (db *Database) CreateEpic(epic *models.Epic) error {
	return db.Create(epic).Error
}

func (db *Database) GetEpic(id uint) (*models.Epic, error) {
	var epic models.Epic
	err := db.Preload("Project").Preload("Tasks").First(&epic, id).Error
	if err != nil {
		return nil, err
	}
	return &epic, nil
}

func (db *Database) ListEpics(projectID *uint, status *models.EpicStatus) ([]models.Epic, error) {
	var epics []models.Epic
	query := db.DB

	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	err := query.Preload("Tasks").Find(&epics).Error
	return epics, err
}

func (db *Database) UpdateEpic(epic *models.Epic) error {
	return db.Save(epic).Error
}

func (db *Database) DeleteEpic(id uint) error {
	return db.Delete(&models.Epic{}, id).Error
}

func (db *Database) GetEpicsByProject(projectID uint) ([]models.Epic, error) {
	var epics []models.Epic
	err := db.Where("project_id = ?", projectID).Preload("Tasks").Find(&epics).Error
	return epics, err
}

func (db *Database) AssignTaskToEpic(taskID uint, epicID uint) error {
	return db.Model(&models.Task{}).Where("id = ?", taskID).Update("epic_id", epicID).Error
}

func (db *Database) RemoveTaskFromEpic(taskID uint) error {
	return db.Model(&models.Task{}).Where("id = ?", taskID).Update("epic_id", nil).Error
}

func (db *Database) CalculateEpicProgress(epicID uint) (int, error) {
	var epic models.Epic
	if err := db.Preload("Tasks").First(&epic, epicID).Error; err != nil {
		return 0, err
	}

	if len(epic.Tasks) == 0 {
		return 0, nil
	}

	completedTasks := 0
	for _, task := range epic.Tasks {
		if task.Status == models.TaskStatusDone {
			completedTasks++
		}
	}

	progress := (completedTasks * 100) / len(epic.Tasks)

	// Update the epic progress
	if err := db.Model(&epic).Update("progress", progress).Error; err != nil {
		return 0, err
	}

	return progress, nil
}