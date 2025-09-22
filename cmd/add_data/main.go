package main

import (
	"log"
	"time"

	"headless-project-management/internal/db"
	"headless-project-management/internal/models"
)

func main() {
	// Initialize database
	database, err := db.InitDB("project.db")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Update existing tasks with more data
	var tasks []models.Task
	database.DB.Find(&tasks)

	assignees := []string{"John Doe", "Jane Smith", "Bob Wilson", "Alice Brown"}

	// Create some labels for project 1
	labels := []models.Label{
		{ProjectID: 1, Name: "Frontend", Color: "#3b82f6"},
		{ProjectID: 1, Name: "Backend", Color: "#10b981"},
		{ProjectID: 1, Name: "Bug", Color: "#ef4444"},
		{ProjectID: 1, Name: "Enhancement", Color: "#8b5cf6"},
	}

	for _, label := range labels {
		database.DB.FirstOrCreate(&label, models.Label{ProjectID: label.ProjectID, Name: label.Name})
	}

	// Get all labels
	var allLabels []models.Label
	database.DB.Where("project_id = ?", 1).Find(&allLabels)

	// Update tasks with assignees, labels, and hours
	for i, task := range tasks {
		if task.ProjectID == 1 {
			// Assign to different people
			task.Assignee = assignees[i%len(assignees)]

			// Add estimated hours
			hours := float64(2 + (i%8)*2)
			task.EstimatedHours = &hours

			// Add due dates for some tasks
			if i%3 == 0 {
				dueDate := time.Now().Add(time.Hour * 24 * time.Duration(i-5))
				task.DueDate = &dueDate
			}

			// Update task
			database.DB.Model(&task).Updates(map[string]interface{}{
				"assignee": task.Assignee,
				"estimated_hours": task.EstimatedHours,
				"due_date": task.DueDate,
			})

			// Add labels to tasks (1-2 labels per task)
			if i%2 == 0 && len(allLabels) > 0 {
				database.DB.Exec("INSERT OR IGNORE INTO task_labels (task_id, label_id) VALUES (?, ?)",
					task.ID, allLabels[i%len(allLabels)].ID)
			}
			if i%3 == 0 && len(allLabels) > 1 {
				database.DB.Exec("INSERT OR IGNORE INTO task_labels (task_id, label_id) VALUES (?, ?)",
					task.ID, allLabels[(i+1)%len(allLabels)].ID)
			}
		}
	}

	log.Println("Sample data added successfully!")
}