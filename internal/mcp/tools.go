package mcp

import (
	"fmt"
	"time"

	"github.com/headless-pm/headless-project-management/internal/models"
	"github.com/headless-pm/headless-project-management/pkg/auth"
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
		ID          uint   `json:"project_id"`
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
		ID          uint   `json:"task_id"`
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
		EpicID       uint `json:"epic_id"`
		CascadeTasks bool `json:"cascade_tasks,omitempty"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	// Use the new DeleteEpicWithOptions function
	if err := s.db.DeleteEpicWithOptions(input.EpicID, input.CascadeTasks); err != nil {
		return ErrorResponse(fmt.Errorf("failed to delete epic: %w", err)), nil
	}

	result := map[string]interface{}{
		"status": "deleted",
		"epic_id": input.EpicID,
		"cascade_tasks": input.CascadeTasks,
	}

	return SuccessResponse(result), nil
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

func (s *EnhancedMCPServer) updateLabel(args []byte) (*ToolResponse, error) {
	var input struct {
		LabelID uint   `json:"label_id"`
		Name    string `json:"name"`
		Color   string `json:"color"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	// Get the existing label
	var label models.Label
	if err := s.db.GetLabelByID(fmt.Sprintf("%d", input.LabelID), &label); err != nil {
		return ErrorResponse(fmt.Errorf("label not found: %w", err)), nil
	}

	// Update fields if provided
	if input.Name != "" {
		label.Name = input.Name
	}
	if input.Color != "" {
		label.Color = input.Color
	}

	if err := s.db.UpdateLabel(&label); err != nil {
		return ErrorResponse(fmt.Errorf("failed to update label: %w", err)), nil
	}

	return SuccessResponse(label), nil
}

func (s *EnhancedMCPServer) deleteLabel(args []byte) (*ToolResponse, error) {
	var input struct {
		LabelID uint `json:"label_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	if err := s.db.DeleteLabel(fmt.Sprintf("%d", input.LabelID)); err != nil {
		return ErrorResponse(fmt.Errorf("failed to delete label: %w", err)), nil
	}

	return SuccessResponse(map[string]interface{}{
		"message": fmt.Sprintf("Label %d deleted successfully", input.LabelID),
	}), nil
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

// User management operations
func (s *EnhancedMCPServer) createUser(args []byte) (*ToolResponse, error) {
	var input struct {
		Username  string `json:"username"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Role      string `json:"role"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(input.Password)
	if err != nil {
		return ErrorResponse(fmt.Errorf("failed to hash password: %w", err)), nil
	}

	// Default role if not specified
	role := models.UserRoleMember
	if input.Role != "" {
		switch input.Role {
		case "admin":
			role = models.UserRoleAdmin
		case "member":
			role = models.UserRoleMember
		default:
			role = models.UserRoleMember
		}
	}

	user := &models.User{
		Username:  input.Username,
		Email:     input.Email,
		Password:  hashedPassword,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Role:      role,
		IsActive:  true,
	}

	if err := s.db.CreateUser(user); err != nil {
		return ErrorResponse(fmt.Errorf("failed to create user: %w", err)), nil
	}

	// Don't return password in response
	user.Password = ""
	return SuccessResponse(user), nil
}

func (s *EnhancedMCPServer) listUsers(args []byte) (*ToolResponse, error) {
	users, err := s.db.ListUsers()
	if err != nil {
		return ErrorResponse(fmt.Errorf("failed to list users: %w", err)), nil
	}

	// Remove passwords from response
	for i := range users {
		users[i].Password = ""
	}

	return SuccessResponse(users), nil
}

func (s *EnhancedMCPServer) getUserByID(args []byte) (*ToolResponse, error) {
	var input struct {
		UserID uint `json:"user_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	user, err := s.db.GetUserByID(input.UserID)
	if err != nil {
		return ErrorResponse(fmt.Errorf("user not found: %w", err)), nil
	}

	// Don't return password in response
	user.Password = ""
	return SuccessResponse(user), nil
}

func (s *EnhancedMCPServer) updateUser(args []byte) (*ToolResponse, error) {
	var input struct {
		UserID    uint   `json:"user_id"`
		Username  string `json:"username"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Role      string `json:"role"`
		IsActive  *bool  `json:"is_active"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	user, err := s.db.GetUserByID(input.UserID)
	if err != nil {
		return ErrorResponse(fmt.Errorf("user not found: %w", err)), nil
	}

	// Update fields if provided
	if input.Username != "" {
		user.Username = input.Username
	}
	if input.Email != "" {
		user.Email = input.Email
	}
	if input.FirstName != "" {
		user.FirstName = input.FirstName
	}
	if input.LastName != "" {
		user.LastName = input.LastName
	}
	if input.Role != "" {
		switch input.Role {
		case "admin":
			user.Role = models.UserRoleAdmin
		case "member":
			user.Role = models.UserRoleMember
		}
	}
	if input.IsActive != nil {
		user.IsActive = *input.IsActive
	}

	if err := s.db.UpdateUser(user); err != nil {
		return ErrorResponse(fmt.Errorf("failed to update user: %w", err)), nil
	}

	// Don't return password in response
	user.Password = ""
	return SuccessResponse(user), nil
}

func (s *EnhancedMCPServer) deleteUser(args []byte) (*ToolResponse, error) {
	var input struct {
		UserID uint `json:"user_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	if err := s.db.DeleteUser(input.UserID); err != nil {
		return ErrorResponse(fmt.Errorf("failed to delete user: %w", err)), nil
	}

	return SuccessResponse(map[string]interface{}{
		"message": fmt.Sprintf("User %d deleted successfully", input.UserID),
	}), nil
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

func (s *EnhancedMCPServer) updateComment(args []byte) (*ToolResponse, error) {
	var input struct {
		CommentID uint   `json:"comment_id"`
		Content   string `json:"content"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	// Get the existing comment first to verify it exists
	comment, err := s.db.GetComment(input.CommentID)
	if err != nil {
		return ErrorResponse(fmt.Errorf("comment not found: %w", err)), nil
	}

	// Update the comment content
	if err := s.db.UpdateComment(input.CommentID, input.Content); err != nil {
		return ErrorResponse(err), nil
	}

	// Return the updated comment
	comment.Content = input.Content
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

func (s *EnhancedMCPServer) deleteComment(args []byte) (*ToolResponse, error) {
	var input struct {
		CommentID uint `json:"comment_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	// Get the comment first to verify it exists
	comment, err := s.db.GetComment(input.CommentID)
	if err != nil {
		return ErrorResponse(fmt.Errorf("comment not found: %w", err)), nil
	}

	// Delete the comment
	if err := s.db.DeleteComment(input.CommentID); err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(map[string]interface{}{
		"deleted":    true,
		"comment_id": input.CommentID,
		"task_id":    comment.TaskID,
	}), nil
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

	// Check for circular dependencies before creating
	hasCircular, err := s.db.CheckCircularDependency(input.TaskID, input.DependsOnID)
	if err != nil {
		return ErrorResponse(fmt.Errorf("failed to check circular dependency: %w", err)), nil
	}
	if hasCircular {
		return ErrorResponse(fmt.Errorf("cannot create dependency: would create a circular dependency chain")), nil
	}

	dep := &models.TaskDependency{
		TaskID:      input.TaskID,
		DependsOnID: input.DependsOnID,
		Type:        input.Type,
	}

	if err := s.db.CreateTaskDependency(dep); err != nil {
		return ErrorResponse(fmt.Errorf("failed to create dependency: %w", err)), nil
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

	deps, err := s.db.GetTaskDependencies(input.TaskID)
	if err != nil {
		return ErrorResponse(fmt.Errorf("failed to get dependencies: %w", err)), nil
	}

	return SuccessResponse(deps), nil
}

// Get the full dependency chain for a task (all transitive dependencies)
func (s *EnhancedMCPServer) getTaskDependencyChain(args []byte) (*ToolResponse, error) {
	var input struct {
		TaskID uint `json:"task_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	chain, err := s.db.GetTaskDependencyChain(input.TaskID)
	if err != nil {
		return ErrorResponse(fmt.Errorf("failed to get dependency chain: %w", err)), nil
	}

	return SuccessResponse(chain), nil
}

// Get tasks that depend on this task
func (s *EnhancedMCPServer) getTaskDependentChain(args []byte) (*ToolResponse, error) {
	var input struct {
		TaskID uint `json:"task_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	chain, err := s.db.GetTaskDependentChain(input.TaskID)
	if err != nil {
		return ErrorResponse(fmt.Errorf("failed to get dependent chain: %w", err)), nil
	}

	return SuccessResponse(chain), nil
}

// Check if a task can start based on its dependencies
func (s *EnhancedMCPServer) canStartTask(args []byte) (*ToolResponse, error) {
	var input struct {
		TaskID uint `json:"task_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	canStart, err := s.db.CanStartTask(input.TaskID)
	if err != nil {
		return ErrorResponse(fmt.Errorf("failed to check task readiness: %w", err)), nil
	}

	// Get blocking tasks if can't start
	var blockingTasks []models.Task
	if !canStart {
		deps, _ := s.db.GetTaskDependencies(input.TaskID)
		for _, dep := range deps {
			task, _ := s.db.GetTask(dep.DependsOnID)
			if task != nil && task.Status != models.TaskStatusDone {
				blockingTasks = append(blockingTasks, *task)
			}
		}
	}

	return SuccessResponse(map[string]interface{}{
		"can_start":      canStart,
		"blocking_tasks": blockingTasks,
	}), nil
}

// Get the full dependency graph for a project
func (s *EnhancedMCPServer) getProjectDependencyGraph(args []byte) (*ToolResponse, error) {
	var input struct {
		ProjectID uint `json:"project_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	graph, err := s.db.GetProjectDependencyGraph(input.ProjectID)
	if err != nil {
		return ErrorResponse(fmt.Errorf("failed to get dependency graph: %w", err)), nil
	}

	return SuccessResponse(graph), nil
}