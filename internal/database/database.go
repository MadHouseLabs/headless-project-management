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
	embeddingCallback func(entityType string, entityID uint)
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

	// Migrate all models except TaskDependency first
	if err := db.AutoMigrate(
		// Core entities
		&models.User{},
		&models.Project{},
		&models.Epic{},
		&models.Task{},
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

	// Handle TaskDependency migration separately due to potential schema issues
	// Check if table exists and has the correct structure
	var tableExists int
	db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='task_dependencies'").Scan(&tableExists)

	if tableExists == 0 {
		// Create the table if it doesn't exist
		if err := db.AutoMigrate(&models.TaskDependency{}); err != nil {
			return nil, fmt.Errorf("failed to create task_dependencies table: %w", err)
		}
	}

	// Fix column name for task_dependencies if needed
	// Check if the old column name exists and rename it
	var columnExists int
	db.Raw("SELECT COUNT(*) FROM pragma_table_info('task_dependencies') WHERE name = 'depends_on_task_id'").Scan(&columnExists)
	if columnExists > 0 {
		// Rename the column from depends_on_task_id to depends_on_id
		if err := db.Exec("ALTER TABLE task_dependencies RENAME COLUMN depends_on_task_id TO depends_on_id").Error; err != nil {
			// SQLite doesn't support RENAME COLUMN in older versions, need to recreate the table
			db.Exec("BEGIN TRANSACTION")
			db.Exec(`CREATE TABLE task_dependencies_new (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				task_id INTEGER NOT NULL,
				depends_on_id INTEGER NOT NULL,
				type TEXT DEFAULT 'finish_to_start',
				CONSTRAINT fk_tasks_dependencies FOREIGN KEY (task_id) REFERENCES tasks(id),
				CONSTRAINT fk_task_dependencies_depends_on FOREIGN KEY (depends_on_id) REFERENCES tasks(id)
			)`)
			db.Exec("INSERT INTO task_dependencies_new (id, task_id, depends_on_id, type) SELECT id, task_id, depends_on_task_id, type FROM task_dependencies")
			db.Exec("DROP TABLE task_dependencies")
			db.Exec("ALTER TABLE task_dependencies_new RENAME TO task_dependencies")
			db.Exec("COMMIT")
		}
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

func (db *Database) SetEmbeddingCallback(callback func(entityType string, entityID uint)) {
	db.embeddingCallback = callback
}

func (db *Database) CreateProject(project *models.Project) error {
	if err := db.Create(project).Error; err != nil {
		return err
	}

	// Queue embedding generation for the new project
	if db.embeddingCallback != nil {
		db.embeddingCallback("project", project.ID)
	}

	return nil
}

func (db *Database) GetProject(id uint) (*models.Project, error) {
	var project models.Project
	err := db.Preload("Tasks").Preload("Epics").First(&project, id).Error
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
	err := query.Preload("Tasks").Preload("Epics").Find(&projects).Error
	return projects, err
}

func (db *Database) UpdateProject(project *models.Project) error {
	if err := db.Save(project).Error; err != nil {
		return err
	}

	// Queue embedding generation for the updated project
	if db.embeddingCallback != nil {
		db.embeddingCallback("project", project.ID)
	}

	return nil
}

func (db *Database) DeleteProject(id uint) error {
	// Start a transaction to ensure all deletions happen atomically
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete all task dependencies for tasks in this project
	if err := tx.Exec(`
		DELETE FROM task_dependencies
		WHERE task_id IN (SELECT id FROM tasks WHERE project_id = ?)
		OR depends_on_id IN (SELECT id FROM tasks WHERE project_id = ?)
	`, id, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete all comments for tasks in this project
	if err := tx.Where("task_id IN (SELECT id FROM tasks WHERE project_id = ?)", id).
		Delete(&models.Comment{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete all attachments for tasks in this project
	if err := tx.Where("task_id IN (SELECT id FROM tasks WHERE project_id = ?)", id).
		Delete(&models.Attachment{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete all tasks (including subtasks) for this project
	if err := tx.Where("project_id = ?", id).Delete(&models.Task{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete all epics for this project
	if err := tx.Where("project_id = ?", id).Delete(&models.Epic{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete all labels for this project
	if err := tx.Where("project_id = ?", id).Delete(&models.Label{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Finally, delete the project itself
	if err := tx.Delete(&models.Project{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	return tx.Commit().Error
}

func (db *Database) CreateTask(task *models.Task) error {
	if err := db.Create(task).Error; err != nil {
		return err
	}

	// Queue embedding generation for the new task
	if db.embeddingCallback != nil {
		db.embeddingCallback("task", task.ID)
	}

	return nil
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
	if err := db.Save(task).Error; err != nil {
		return err
	}

	// Queue embedding generation for the updated task
	if db.embeddingCallback != nil {
		db.embeddingCallback("task", task.ID)
	}

	return nil
}

func (db *Database) DeleteTask(id uint) error {
	// Start a transaction to ensure all deletions happen atomically
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete all task dependencies where this task is involved
	if err := tx.Where("task_id = ? OR depends_on_id = ?", id, id).
		Delete(&models.TaskDependency{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete all comments for this task
	if err := tx.Where("task_id = ?", id).Delete(&models.Comment{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete all attachments for this task
	if err := tx.Where("task_id = ?", id).Delete(&models.Attachment{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete all subtasks recursively
	var subtasks []models.Task
	if err := tx.Where("parent_id = ?", id).Find(&subtasks).Error; err != nil {
		tx.Rollback()
		return err
	}

	for _, subtask := range subtasks {
		// Recursively delete each subtask
		if err := db.DeleteTask(subtask.ID); err != nil {
			tx.Rollback()
			return err
		}
	}

	// Finally, delete the task itself
	if err := tx.Delete(&models.Task{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	return tx.Commit().Error
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
	// Start a transaction to ensure all deletions happen atomically
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Remove epic association from all tasks (set epic_id to null)
	if err := tx.Model(&models.Task{}).Where("epic_id = ?", id).
		Update("epic_id", nil).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete the epic
	if err := tx.Delete(&models.Epic{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	return tx.Commit().Error
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

// Task Dependency Management Methods

func (db *Database) CreateTaskDependency(dependency *models.TaskDependency) error {
	// Check if dependency already exists
	var count int64
	if err := db.Model(&models.TaskDependency{}).
		Where("task_id = ? AND depends_on_id = ?", dependency.TaskID, dependency.DependsOnID).
		Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("dependency already exists")
	}

	// Check for self-dependency
	if dependency.TaskID == dependency.DependsOnID {
		return fmt.Errorf("task cannot depend on itself")
	}

	// Check for circular dependency
	if hasCycle, err := db.CheckCircularDependency(dependency.TaskID, dependency.DependsOnID); err != nil {
		return err
	} else if hasCycle {
		return fmt.Errorf("circular dependency detected")
	}

	// Create the dependency
	return db.Create(dependency).Error
}

func (db *Database) GetTaskDependencies(taskID uint) ([]models.TaskDependency, error) {
	var dependencies []models.TaskDependency
	err := db.Where("task_id = ?", taskID).
		Preload("DependsOn").
		Find(&dependencies).Error
	return dependencies, err
}

func (db *Database) GetTaskDependents(taskID uint) ([]models.TaskDependency, error) {
	var dependents []models.TaskDependency
	err := db.Where("depends_on_id = ?", taskID).
		Preload("Task").
		Find(&dependents).Error
	return dependents, err
}

func (db *Database) GetAllTaskDependencies(taskID uint) ([]models.TaskDependency, error) {
	var dependencies []models.TaskDependency
	err := db.Where("task_id = ? OR depends_on_id = ?", taskID, taskID).
		Preload("Task").
		Preload("DependsOn").
		Find(&dependencies).Error
	return dependencies, err
}

func (db *Database) DeleteTaskDependency(id uint) error {
	return db.Delete(&models.TaskDependency{}, id).Error
}

func (db *Database) DeleteTaskDependencyByTaskIDs(taskID, dependsOnID uint) error {
	return db.Where("task_id = ? AND depends_on_id = ?", taskID, dependsOnID).
		Delete(&models.TaskDependency{}).Error
}

// CheckCircularDependency checks if adding a dependency from taskID to dependsOnID would create a cycle
func (db *Database) CheckCircularDependency(taskID, dependsOnID uint) (bool, error) {
	// Use recursive CTE to check for cycles
	var count int64

	// This query checks if there's a path from dependsOnID to taskID
	// If there is, adding taskID -> dependsOnID would create a cycle
	query := `
		WITH RECURSIVE dependency_path AS (
			-- Base case: start from the potential dependent task
			SELECT task_id, depends_on_id FROM task_dependencies
			WHERE task_id = ?

			UNION

			-- Recursive case: follow the dependency chain
			SELECT td.task_id, td.depends_on_id
			FROM task_dependencies td
			INNER JOIN dependency_path dp ON td.task_id = dp.depends_on_id
		)
		SELECT COUNT(*) FROM dependency_path WHERE depends_on_id = ?
	`

	if err := db.Raw(query, dependsOnID, taskID).Scan(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetTaskDependencyChain returns all tasks that must be completed before the given task
func (db *Database) GetTaskDependencyChain(taskID uint) ([]models.Task, error) {
	var tasks []models.Task

	query := `
		WITH RECURSIVE dependency_chain AS (
			-- Base case: direct dependencies
			SELECT depends_on_id FROM task_dependencies
			WHERE task_id = ?

			UNION

			-- Recursive case: transitive dependencies
			SELECT td.depends_on_id
			FROM task_dependencies td
			INNER JOIN dependency_chain dc ON td.task_id = dc.depends_on_id
		)
		SELECT DISTINCT t.* FROM tasks t
		INNER JOIN dependency_chain dc ON t.id = dc.depends_on_id
		ORDER BY t.id
	`

	if err := db.Raw(query, taskID).Scan(&tasks).Error; err != nil {
		return nil, err
	}

	return tasks, nil
}

// GetTaskDependentChain returns all tasks that depend on the given task
func (db *Database) GetTaskDependentChain(taskID uint) ([]models.Task, error) {
	var tasks []models.Task

	query := `
		WITH RECURSIVE dependent_chain AS (
			-- Base case: direct dependents
			SELECT task_id FROM task_dependencies
			WHERE depends_on_id = ?

			UNION

			-- Recursive case: transitive dependents
			SELECT td.task_id
			FROM task_dependencies td
			INNER JOIN dependent_chain dc ON td.depends_on_id = dc.task_id
		)
		SELECT DISTINCT t.* FROM tasks t
		INNER JOIN dependent_chain dc ON t.id = dc.task_id
		ORDER BY t.id
	`

	if err := db.Raw(query, taskID).Scan(&tasks).Error; err != nil {
		return nil, err
	}

	return tasks, nil
}

// CanStartTask checks if a task can be started based on its dependencies
func (db *Database) CanStartTask(taskID uint) (bool, error) {
	// Get all dependencies
	dependencies, err := db.GetTaskDependencies(taskID)
	if err != nil {
		return false, err
	}

	// If no dependencies, task can start
	if len(dependencies) == 0 {
		return true, nil
	}

	// Check if all dependencies are completed
	for _, dep := range dependencies {
		if dep.Type == "finish_to_start" || dep.Type == "" {
			// Default dependency type: predecessor must be done
			if dep.DependsOn != nil && dep.DependsOn.Status != models.TaskStatusDone {
				return false, nil
			}
		} else if dep.Type == "start_to_start" {
			// Predecessor must be at least started
			if dep.DependsOn != nil && dep.DependsOn.Status == models.TaskStatusTodo {
				return false, nil
			}
		}
		// Add more dependency types as needed
	}

	return true, nil
}

// GetProjectDependencyGraph returns all tasks and their dependencies for visualization
func (db *Database) GetProjectDependencyGraph(projectID uint) (map[string]interface{}, error) {
	// Get all tasks in the project
	var tasks []models.Task
	if err := db.Where("project_id = ?", projectID).Find(&tasks).Error; err != nil {
		return nil, err
	}

	// Get all dependencies for tasks in this project
	var dependencies []models.TaskDependency
	query := db.Table("task_dependencies").
		Joins("JOIN tasks t1 ON task_dependencies.task_id = t1.id").
		Joins("JOIN tasks t2 ON task_dependencies.depends_on_id = t2.id").
		Where("t1.project_id = ? AND t2.project_id = ?", projectID, projectID).
		Preload("Task").
		Preload("DependsOn")

	if err := query.Find(&dependencies).Error; err != nil {
		return nil, err
	}

	// Build the graph structure
	graph := map[string]interface{}{
		"nodes": tasks,
		"edges": dependencies,
		"stats": map[string]int{
			"total_tasks":        len(tasks),
			"total_dependencies": len(dependencies),
		},
	}

	return graph, nil
}

// RemoveTaskDependency removes a specific task dependency by ID
func (db *Database) RemoveTaskDependency(dependencyID uint) error {
	return db.Delete(&models.TaskDependency{}, dependencyID).Error
}