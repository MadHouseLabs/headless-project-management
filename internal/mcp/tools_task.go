package mcp

import (
	"time"

	"github.com/headless-pm/headless-project-management/internal/models"
)

// Task Management Tool Implementations

func (s *EnhancedMCPServer) createTask(args []byte) (*ToolResponse, error) {
	var input TaskInput
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	task := &models.Task{
		ProjectID:   input.ProjectID,
		Title:       input.Title,
		Description: input.Description,
		Status:      models.TaskStatusTodo,
	}

	// Set optional pointer fields
	if input.AssigneeID > 0 {
		task.AssigneeID = &input.AssigneeID
	}
	if input.MilestoneID > 0 {
		task.MilestoneID = &input.MilestoneID
	}
	if input.SprintID > 0 {
		task.SprintID = &input.SprintID
	}

	// Set priority
	if input.Priority != "" {
		task.Priority = models.TaskPriority(input.Priority)
	} else {
		task.Priority = models.TaskPriorityMedium
	}

	// Parse due date if provided
	if input.DueDate != "" {
		dueDate, err := time.Parse("2006-01-02T15:04:05Z", input.DueDate)
		if err != nil {
			// Try simpler date format
			dueDate, err = time.Parse("2006-01-02", input.DueDate)
			if err != nil {
				return ErrorResponse(err), nil
			}
		}
		task.DueDate = &dueDate
	}

	if err := s.db.CreateTask(task); err != nil {
		return ErrorResponse(err), nil
	}

	// Queue embedding generation
	if s.embedWorker != nil {
		s.embedWorker.QueueJob("task", task.ID)
	}

	// Handle labels
	if err := s.assignLabelsToTask(task.ID, input.ProjectID, input.Labels); err != nil {
		// Log error but don't fail the task creation
		_ = err
	}

	return SuccessResponse(task), nil
}

func (s *EnhancedMCPServer) updateTask(args []byte) (*ToolResponse, error) {
	var input struct {
		ID          uint   `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
		Priority    string `json:"priority"`
		AssigneeID  uint   `json:"assignee_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	task, err := s.db.GetTask(input.ID)
	if err != nil {
		return ErrorResponse(err), nil
	}

	// Update only provided fields
	if input.Title != "" {
		task.Title = input.Title
	}
	if input.Description != "" {
		task.Description = input.Description
	}
	if input.Status != "" {
		task.Status = models.TaskStatus(input.Status)
	}
	if input.Priority != "" {
		task.Priority = models.TaskPriority(input.Priority)
	}
	if input.AssigneeID > 0 {
		task.AssigneeID = &input.AssigneeID
	}

	if err := s.db.UpdateTask(task); err != nil {
		return ErrorResponse(err), nil
	}

	// Queue embedding regeneration
	if s.embedWorker != nil {
		s.embedWorker.QueueJob("task", task.ID)
	}

	// Track activity if status changed
	if input.Status != "" {
		s.trackActivity("task_status_changed", task.ID, input.AssigneeID)
	}

	return SuccessResponse(task), nil
}

func (s *EnhancedMCPServer) listTasks(args []byte) (*ToolResponse, error) {
	var input struct {
		ProjectID   uint   `json:"project_id"`
		Status      string `json:"status"`
		AssigneeID  uint   `json:"assignee_id"`
		MilestoneID uint   `json:"milestone_id"`
		SprintID    uint   `json:"sprint_id"`
	}
	_ = UnmarshalArgs(args, &input)

	query := s.db.DB
	if input.ProjectID > 0 {
		query = query.Where("project_id = ?", input.ProjectID)
	}
	if input.Status != "" {
		query = query.Where("status = ?", input.Status)
	}
	if input.AssigneeID > 0 {
		query = query.Where("assignee_id = ?", input.AssigneeID)
	}
	if input.MilestoneID > 0 {
		query = query.Where("milestone_id = ?", input.MilestoneID)
	}
	if input.SprintID > 0 {
		query = query.Where("sprint_id = ?", input.SprintID)
	}

	var tasks []models.Task
	if err := query.
		Preload("Labels").
		Preload("Comments").
		Order("priority DESC, created_at DESC").
		Find(&tasks).Error; err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(tasks), nil
}

func (s *EnhancedMCPServer) addTaskDependency(args []byte) (*ToolResponse, error) {
	var input struct {
		TaskID      uint   `json:"task_id"`
		DependsOnID uint   `json:"depends_on_id"`
		Type        string `json:"type"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	// Validate that both tasks exist
	var count int64
	s.db.Model(&models.Task{}).Where("id IN ?", []uint{input.TaskID, input.DependsOnID}).Count(&count)
	if count != 2 {
		return ErrorResponse(ErrTaskNotFound), nil
	}

	dependency := &models.TaskDependency{
		TaskID:          input.TaskID,
		DependsOnTaskID: input.DependsOnID,
		Type:            input.Type,
	}

	if dependency.Type == "" {
		dependency.Type = "blocks"
	}

	if err := s.db.Create(dependency).Error; err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(dependency), nil
}

func (s *EnhancedMCPServer) addComment(args []byte) (*ToolResponse, error) {
	var input struct {
		TaskID  uint   `json:"task_id"`
		Content string `json:"content"`
		Author  string `json:"author"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	// Validate task exists
	var task models.Task
	if err := s.db.First(&task, input.TaskID).Error; err != nil {
		return ErrorResponse(ErrTaskNotFound), nil
	}

	comment := &models.Comment{
		TaskID:  input.TaskID,
		Content: input.Content,
		Author:  input.Author,
	}

	if err := s.db.AddComment(comment); err != nil {
		return ErrorResponse(err), nil
	}

	// Queue task embedding regeneration (includes comments)
	if s.embedWorker != nil {
		s.embedWorker.QueueJob("task", input.TaskID)
	}

	// Track activity
	s.trackActivity("comment_added", input.TaskID, 0)

	return SuccessResponse(comment), nil
}

func (s *EnhancedMCPServer) searchTasks(args []byte) (*ToolResponse, error) {
	var input struct {
		Query     string `json:"query"`
		ProjectID uint   `json:"project_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	if input.Query == "" {
		return ErrorResponse(ErrInvalidInput), nil
	}

	searchPattern := "%" + input.Query + "%"
	query := s.db.Where("title LIKE ? OR description LIKE ?", searchPattern, searchPattern)

	if input.ProjectID > 0 {
		query = query.Where("project_id = ?", input.ProjectID)
	}

	var tasks []models.Task
	if err := query.
		Preload("Project").
		Order("priority DESC, updated_at DESC").
		Limit(50).
		Find(&tasks).Error; err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(tasks), nil
}

// Helper function to assign labels to a task
func (s *EnhancedMCPServer) assignLabelsToTask(taskID, projectID uint, labelNames []string) error {
	for _, labelName := range labelNames {
		if labelName == "" {
			continue
		}

		var label models.Label
		err := s.db.Where("project_id = ? AND name = ?", projectID, labelName).First(&label).Error
		if err != nil {
			// Create label if it doesn't exist
			label = models.Label{
				ProjectID: projectID,
				Name:      labelName,
				Color:     generateLabelColor(labelName),
			}
			if err := s.db.Create(&label).Error; err != nil {
				continue
			}
		}

		// Assign label to task
		s.db.AssignLabelToTask(taskID, label.ID)
	}
	return nil
}

// Helper to generate a consistent dark color for a label based on its name
func generateLabelColor(name string) string {
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

	return colors[hash%len(colors)]
}