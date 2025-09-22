package mcp

import (
	"fmt"
	"time"

	"github.com/headless-pm/headless-project-management/internal/models"
)

// Project CRUD operations
func (s *EnhancedMCPServer) createProject(args []byte) (*ToolResponse, error) {
	var input ProjectInput
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	project := &models.Project{
		Name:        input.Name,
		Description: input.Description,
		OwnerID:     input.OwnerID,
		Status:      models.ProjectStatusActive,
	}

	if err := s.db.CreateProject(project); err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(project), nil
}

func (s *EnhancedMCPServer) getProject(args []byte) (*ToolResponse, error) {
	var input struct {
		ProjectID uint `json:"project_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	project, err := s.db.GetProject(input.ProjectID)
	if err != nil {
		return ErrorResponse(fmt.Errorf("project not found: %w", err)), nil
	}

	return SuccessResponse(project), nil
}

func (s *EnhancedMCPServer) updateProject(args []byte) (*ToolResponse, error) {
	var input struct {
		ID          uint   `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	project, err := s.db.GetProject(input.ID)
	if err != nil {
		return ErrorResponse(err), nil
	}

	if input.Name != "" {
		project.Name = input.Name
	}
	if input.Description != "" {
		project.Description = input.Description
	}
	if input.Status != "" {
		project.Status = models.ProjectStatus(input.Status)
	}

	if err := s.db.UpdateProject(project); err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(project), nil
}

func (s *EnhancedMCPServer) deleteProject(args []byte) (*ToolResponse, error) {
	var input struct {
		ProjectID uint `json:"project_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	if err := s.db.DeleteProject(input.ProjectID); err != nil {
		return ErrorResponse(fmt.Errorf("failed to delete project: %w", err)), nil
	}

	return SuccessResponse(map[string]string{"status": "deleted"}), nil
}

func (s *EnhancedMCPServer) listProjects(args []byte) (*ToolResponse, error) {
	var input struct {
		Status string `json:"status"`
	}
	_ = UnmarshalArgs(args, &input)

	var status *models.ProjectStatus
	if input.Status != "" {
		s := models.ProjectStatus(input.Status)
		status = &s
	}

	projects, err := s.db.ListProjects(status)
	if err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(projects), nil
}

// Task CRUD operations
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
		Priority:    models.TaskPriorityMedium,
	}

	if input.Priority != "" {
		task.Priority = models.TaskPriority(input.Priority)
	}
	if input.AssigneeID > 0 {
		task.AssigneeID = &input.AssigneeID
	}
	if input.EpicID > 0 {
		task.EpicID = &input.EpicID
	}
	if input.DueDate != "" {
		dueDate, err := time.Parse("2006-01-02", input.DueDate)
		if err == nil {
			task.DueDate = &dueDate
		}
	}

	if err := s.db.CreateTask(task); err != nil {
		return ErrorResponse(err), nil
	}

	// Handle labels
	if len(input.Labels) > 0 {
		s.db.AssignLabelsToTask(task.ID, input.ProjectID, input.Labels)
	}

	return SuccessResponse(task), nil
}

func (s *EnhancedMCPServer) getTask(args []byte) (*ToolResponse, error) {
	var input struct {
		TaskID uint `json:"task_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	task, err := s.db.GetTask(input.TaskID)
	if err != nil {
		return ErrorResponse(fmt.Errorf("task not found: %w", err)), nil
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

	return SuccessResponse(task), nil
}

func (s *EnhancedMCPServer) deleteTask(args []byte) (*ToolResponse, error) {
	var input struct {
		TaskID uint `json:"task_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	if err := s.db.DeleteTask(input.TaskID); err != nil {
		return ErrorResponse(fmt.Errorf("failed to delete task: %w", err)), nil
	}

	return SuccessResponse(map[string]string{"status": "deleted"}), nil
}

func (s *EnhancedMCPServer) listTasks(args []byte) (*ToolResponse, error) {
	var input struct {
		ProjectID  uint   `json:"project_id"`
		Status     string `json:"status"`
		AssigneeID uint   `json:"assignee_id"`
		EpicID     uint   `json:"epic_id"`
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
	if input.EpicID > 0 {
		query = query.Where("epic_id = ?", input.EpicID)
	}

	var tasks []models.Task
	if err := query.Preload("Labels").Find(&tasks).Error; err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(tasks), nil
}

// Epic CRUD operations
func (s *EnhancedMCPServer) createEpic(args []byte) (*ToolResponse, error) {
	var input struct {
		ProjectID   uint   `json:"project_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	epic := &models.Epic{
		ProjectID:   input.ProjectID,
		Name:        input.Name,
		Description: input.Description,
		Status:      models.EpicStatusPlanned,
	}

	if err := s.db.CreateEpic(epic); err != nil {
		return ErrorResponse(fmt.Errorf("failed to create epic: %w", err)), nil
	}

	return SuccessResponse(epic), nil
}

func (s *EnhancedMCPServer) getEpic(args []byte) (*ToolResponse, error) {
	var input struct {
		EpicID uint `json:"epic_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	epic, err := s.db.GetEpic(input.EpicID)
	if err != nil {
		return ErrorResponse(fmt.Errorf("epic not found: %w", err)), nil
	}

	return SuccessResponse(epic), nil
}

func (s *EnhancedMCPServer) updateEpic(args []byte) (*ToolResponse, error) {
	var input struct {
		EpicID      uint   `json:"epic_id"`
		Name        string `json:"name,omitempty"`
		Description string `json:"description,omitempty"`
		Status      string `json:"status,omitempty"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	epic, err := s.db.GetEpic(input.EpicID)
	if err != nil {
		return ErrorResponse(err), nil
	}

	if input.Name != "" {
		epic.Name = input.Name
	}
	if input.Description != "" {
		epic.Description = input.Description
	}
	if input.Status != "" {
		epic.Status = models.EpicStatus(input.Status)
	}

	if err := s.db.UpdateEpic(epic); err != nil {
		return ErrorResponse(fmt.Errorf("failed to update epic: %w", err)), nil
	}

	return SuccessResponse(epic), nil
}

func (s *EnhancedMCPServer) deleteEpic(args []byte) (*ToolResponse, error) {
	var input struct {
		EpicID uint `json:"epic_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	if err := s.db.DeleteEpic(input.EpicID); err != nil {
		return ErrorResponse(fmt.Errorf("failed to delete epic: %w", err)), nil
	}

	return SuccessResponse(map[string]string{"status": "deleted"}), nil
}

func (s *EnhancedMCPServer) listEpics(args []byte) (*ToolResponse, error) {
	var input struct {
		ProjectID uint `json:"project_id,omitempty"`
	}
	_ = UnmarshalArgs(args, &input)

	var projectID *uint
	if input.ProjectID > 0 {
		projectID = &input.ProjectID
	}

	epics, err := s.db.ListEpics(projectID, nil)
	if err != nil {
		return ErrorResponse(fmt.Errorf("failed to list epics: %w", err)), nil
	}

	return SuccessResponse(epics), nil
}

// Label operations
func (s *EnhancedMCPServer) createLabel(args []byte) (*ToolResponse, error) {
	var input struct {
		ProjectID uint   `json:"project_id"`
		Name      string `json:"name"`
		Color     string `json:"color"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	label := &models.Label{
		ProjectID: input.ProjectID,
		Name:      input.Name,
		Color:     input.Color,
	}

	if label.Color == "" {
		// Generate a color based on the name
		colors := []string{"#DC2626", "#059669", "#2563EB", "#7C3AED"}
		hash := 0
		for _, c := range label.Name {
			hash = hash*31 + int(c)
		}
		label.Color = colors[hash%len(colors)]
	}

	if err := s.db.CreateLabel(label); err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(label), nil
}

func (s *EnhancedMCPServer) assignLabel(args []byte) (*ToolResponse, error) {
	var input struct {
		TaskID  uint `json:"task_id"`
		LabelID uint `json:"label_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	if err := s.db.AssignLabelToTask(input.TaskID, input.LabelID); err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(map[string]interface{}{
		"success":  true,
		"task_id":  input.TaskID,
		"label_id": input.LabelID,
	}), nil
}

func (s *EnhancedMCPServer) listLabels(args []byte) (*ToolResponse, error) {
	var input struct {
		ProjectID uint `json:"project_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	labels, err := s.db.GetLabelsByProject(input.ProjectID)
	if err != nil {
		return ErrorResponse(fmt.Errorf("failed to list labels: %w", err)), nil
	}

	return SuccessResponse(labels), nil
}

// Assignee operations
func (s *EnhancedMCPServer) assignTask(args []byte) (*ToolResponse, error) {
	var input struct {
		TaskID     uint `json:"task_id"`
		AssigneeID uint `json:"assignee_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	task, err := s.db.GetTask(input.TaskID)
	if err != nil {
		return ErrorResponse(err), nil
	}

	task.AssigneeID = &input.AssigneeID
	if err := s.db.UpdateTask(task); err != nil {
		return ErrorResponse(fmt.Errorf("failed to assign task: %w", err)), nil
	}

	return SuccessResponse(task), nil
}

func (s *EnhancedMCPServer) listAssignees(args []byte) (*ToolResponse, error) {
	var input struct {
		ProjectID uint `json:"project_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	// Get all users who have tasks in this project
	var users []models.User
	if err := s.db.Distinct("users.*").
		Joins("JOIN tasks ON tasks.assignee_id = users.id").
		Where("tasks.project_id = ?", input.ProjectID).
		Find(&users).Error; err != nil {
		return ErrorResponse(fmt.Errorf("failed to list assignees: %w", err)), nil
	}

	return SuccessResponse(users), nil
}

// Comment operations
func (s *EnhancedMCPServer) addComment(args []byte) (*ToolResponse, error) {
	var input struct {
		TaskID  uint   `json:"task_id"`
		Content string `json:"content"`
		Author  string `json:"author"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	comment := &models.Comment{
		TaskID:  input.TaskID,
		Content: input.Content,
		Author:  input.Author,
	}

	if err := s.db.AddComment(comment); err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(comment), nil
}

func (s *EnhancedMCPServer) listComments(args []byte) (*ToolResponse, error) {
	var input struct {
		TaskID uint `json:"task_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	var comments []models.Comment
	if err := s.db.Where("task_id = ?", input.TaskID).
		Order("created_at DESC").
		Find(&comments).Error; err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(comments), nil
}

// Task dependency operations
func (s *EnhancedMCPServer) addTaskDependency(args []byte) (*ToolResponse, error) {
	var input struct {
		TaskID      uint   `json:"task_id"`
		DependsOnID uint   `json:"depends_on_id"`
		Type        string `json:"type"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	if input.Type == "" {
		input.Type = "finish_to_start"
	}

	dep := &models.TaskDependency{
		TaskID:      input.TaskID,
		DependsOnID: input.DependsOnID,
		Type:        input.Type,
	}

	if err := s.db.Create(dep).Error; err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(dep), nil
}

func (s *EnhancedMCPServer) removeTaskDependency(args []byte) (*ToolResponse, error) {
	var input struct {
		DependencyID uint `json:"dependency_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	if err := s.db.Delete(&models.TaskDependency{}, input.DependencyID).Error; err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(map[string]string{"status": "removed"}), nil
}

func (s *EnhancedMCPServer) listTaskDependencies(args []byte) (*ToolResponse, error) {
	var input struct {
		TaskID uint `json:"task_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	var deps []models.TaskDependency
	if err := s.db.Where("task_id = ? OR depends_on_id = ?", input.TaskID, input.TaskID).
		Preload("Task").Preload("DependsOn").
		Find(&deps).Error; err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(deps), nil
}