package database

import (
	"fmt"
	"testing"

	"github.com/headless-pm/headless-project-management/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*Database, func()) {
	// Create a temporary database for testing
	tmpDir := t.TempDir()
	db, err := NewDatabase(tmpDir)
	require.NoError(t, err)

	cleanup := func() {
		// Clean up is handled by t.TempDir()
	}

	return db, cleanup
}

func TestTaskDependencyManagement(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a test project
	project := &models.Project{
		Name:        "Test Project",
		Description: "Test project for dependency management",
	}
	err := db.CreateProject(project)
	require.NoError(t, err)

	// Create test tasks
	task1 := &models.Task{
		ProjectID:   project.ID,
		Title:       "Task 1",
		Description: "First task",
		Status:      models.TaskStatusTodo,
	}
	err = db.CreateTask(task1)
	require.NoError(t, err)

	task2 := &models.Task{
		ProjectID:   project.ID,
		Title:       "Task 2",
		Description: "Second task",
		Status:      models.TaskStatusTodo,
	}
	err = db.CreateTask(task2)
	require.NoError(t, err)

	task3 := &models.Task{
		ProjectID:   project.ID,
		Title:       "Task 3",
		Description: "Third task",
		Status:      models.TaskStatusTodo,
	}
	err = db.CreateTask(task3)
	require.NoError(t, err)

	t.Run("CreateDependency", func(t *testing.T) {
		// Create dependency: task2 depends on task1
		dep := &models.TaskDependency{
			TaskID:      task2.ID,
			DependsOnID: task1.ID,
			Type:        "finish_to_start",
		}
		err := db.CreateTaskDependency(dep)
		assert.NoError(t, err)
		assert.NotZero(t, dep.ID)
	})

	t.Run("GetDependencies", func(t *testing.T) {
		deps, err := db.GetTaskDependencies(task2.ID)
		assert.NoError(t, err)
		assert.Len(t, deps, 1)
		assert.Equal(t, task1.ID, deps[0].DependsOnID)
	})

	t.Run("CircularDependencyCheck", func(t *testing.T) {
		// Create chain: task3 depends on task2
		dep := &models.TaskDependency{
			TaskID:      task3.ID,
			DependsOnID: task2.ID,
			Type:        "finish_to_start",
		}
		err := db.CreateTaskDependency(dep)
		assert.NoError(t, err)

		// Try to create circular dependency: task1 depends on task3
		hasCircular, err := db.CheckCircularDependency(task1.ID, task3.ID)
		assert.NoError(t, err)
		assert.True(t, hasCircular, "Should detect circular dependency")
	})

	t.Run("CanStartTask", func(t *testing.T) {
		// Task 2 depends on task 1, which is not done
		canStart, err := db.CanStartTask(task2.ID)
		assert.NoError(t, err)
		assert.False(t, canStart, "Task 2 should not be able to start")

		// Mark task 1 as done
		task1.Status = models.TaskStatusDone
		err = db.UpdateTask(task1)
		assert.NoError(t, err)

		// Now task 2 should be able to start
		canStart, err = db.CanStartTask(task2.ID)
		assert.NoError(t, err)
		assert.True(t, canStart, "Task 2 should be able to start now")
	})

	t.Run("DependencyChain", func(t *testing.T) {
		// Get dependency chain for task3 (should include task2 and task1)
		chain, err := db.GetTaskDependencyChain(task3.ID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(chain), 2, "Chain should include at least 2 tasks")
	})

	t.Run("DependentChain", func(t *testing.T) {
		// Get tasks that depend on task1 (should include task2 and task3)
		chain, err := db.GetTaskDependentChain(task1.ID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(chain), 2, "Chain should include at least 2 tasks")
	})

	t.Run("RemoveDependency", func(t *testing.T) {
		deps, err := db.GetTaskDependencies(task2.ID)
		require.NoError(t, err)
		require.Greater(t, len(deps), 0)

		// Remove the dependency
		err = db.RemoveTaskDependency(deps[0].ID)
		assert.NoError(t, err)

		// Verify it's removed
		deps, err = db.GetTaskDependencies(task2.ID)
		assert.NoError(t, err)
		assert.Len(t, deps, 0)
	})
}

func TestProjectDependencyGraph(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a test project
	project := &models.Project{
		Name:        "Test Project",
		Description: "Test project for dependency graph",
	}
	err := db.CreateProject(project)
	require.NoError(t, err)

	// Create multiple tasks
	var tasks []*models.Task
	for i := 1; i <= 5; i++ {
		task := &models.Task{
			ProjectID: project.ID,
			Title:     fmt.Sprintf("Task %d", i),
			Status:    models.TaskStatusTodo,
		}
		err = db.CreateTask(task)
		require.NoError(t, err)
		tasks = append(tasks, task)
	}

	// Create dependencies to form a graph
	// Task 2 depends on Task 1
	// Task 3 depends on Task 1
	// Task 4 depends on Task 2 and Task 3
	// Task 5 depends on Task 4
	dependencies := []struct {
		taskIdx      int
		dependsOnIdx int
	}{
		{1, 0}, // Task 2 depends on Task 1
		{2, 0}, // Task 3 depends on Task 1
		{3, 1}, // Task 4 depends on Task 2
		{3, 2}, // Task 4 depends on Task 3
		{4, 3}, // Task 5 depends on Task 4
	}

	for _, dep := range dependencies {
		err := db.CreateTaskDependency(&models.TaskDependency{
			TaskID:      tasks[dep.taskIdx].ID,
			DependsOnID: tasks[dep.dependsOnIdx].ID,
			Type:        "finish_to_start",
		})
		require.NoError(t, err)
	}

	// Get the dependency graph
	graph, err := db.GetProjectDependencyGraph(project.ID)
	assert.NoError(t, err)
	assert.NotNil(t, graph)

	// Verify the graph structure
	nodes, ok := graph["nodes"].([]interface{})
	assert.True(t, ok, "Graph should have nodes")
	assert.Len(t, nodes, 5, "Graph should have 5 nodes")

	edges, ok := graph["edges"].([]interface{})
	assert.True(t, ok, "Graph should have edges")
	assert.Len(t, edges, 5, "Graph should have 5 edges")
}