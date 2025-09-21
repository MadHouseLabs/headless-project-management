// Package mcp provides Model Context Protocol server implementation
// for the headless project management system.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/headless-pm/headless-project-management/internal/database"
	"github.com/headless-pm/headless-project-management/internal/models"
	"github.com/headless-pm/headless-project-management/internal/service"
	"github.com/headless-pm/headless-project-management/pkg/embeddings"
)

// EnhancedMCPServer provides comprehensive MCP functionality
// with support for AI features, analytics, and team collaboration.
type EnhancedMCPServer struct {
	db            *database.Database
	vectorService *service.VectorService
	embedWorker   *service.EmbeddingWorker
}

// NewEnhancedMCPServer creates a new enhanced MCP server instance
func NewEnhancedMCPServer(
	db *database.Database,
	provider embeddings.EmbeddingProvider,
	worker *service.EmbeddingWorker,
) *EnhancedMCPServer {
	var vectorSvc *service.VectorService
	if provider != nil {
		vectorSvc = service.NewVectorService(db, provider)
	}

	return &EnhancedMCPServer{
		db:            db,
		vectorService: vectorSvc,
		embedWorker:   worker,
	}
}

// ListTools returns all available MCP tools
func (s *EnhancedMCPServer) ListTools() []Tool {
	return toolDefinitions()
}

// ExecuteTool executes a tool by name with the provided arguments
func (s *EnhancedMCPServer) ExecuteTool(ctx context.Context, call ToolCall) (*ToolResponse, error) {
	// Validate server state
	if s.db == nil {
		return ErrorResponse(ErrDatabaseNotConfigured), nil
	}

	// Route to appropriate handler based on tool name
	handlers := s.getToolHandlers()

	handler, exists := handlers[call.Name]
	if !exists {
		return ErrorResponse(fmt.Errorf("unknown tool: %s", call.Name)), nil
	}

	return handler(call.Arguments)
}

// getToolHandlers returns a map of tool names to their handler functions
func (s *EnhancedMCPServer) getToolHandlers() map[string]func([]byte) (*ToolResponse, error) {
	return map[string]func([]byte) (*ToolResponse, error){
		// Project Management
		"create_project":     s.createProject,
		"list_projects":      s.listProjects,
		"update_project":     s.updateProject,
		"get_project_stats":  s.getProjectStats,

		// Task Management
		"create_task":         s.createTask,
		"update_task":         s.updateTask,
		"list_tasks":          s.listTasks,
		"add_task_dependency": s.addTaskDependency,
		"search_tasks":        s.searchTasks,

		// Team Management
		"create_team":       s.createTeam,
		"add_team_member":   s.addTeamMember,
		"list_teams":        s.listTeams,
		"get_team_velocity": s.getTeamVelocity,

		// Milestone Management
		"create_milestone": s.createMilestone,
		"update_milestone": s.updateMilestone,
		"list_milestones":  s.listMilestones,

		// Sprint Management
		"create_sprint":   s.createSprint,
		"start_sprint":    s.startSprint,
		"complete_sprint": s.completeSprint,
		"list_sprints":    s.listSprints,

		// Workflow Management
		"create_workflow":     s.createWorkflow,
		"add_workflow_state":  s.addWorkflowState,

		// Comments and Collaboration
		"add_comment":    s.addComment,
		"list_comments":  s.listComments,

		// Time Tracking
		"log_time":          s.logTime,
		"get_time_entries":  s.getTimeEntries,

		// Notifications
		"create_notification":     s.createNotification,
		"list_notifications":      s.listNotifications,
		"mark_notification_read":  s.markNotificationRead,

		// Webhooks
		"create_webhook": s.createWebhook,
		"list_webhooks":  s.listWebhooks,

		// AI and Search
		"semantic_search":              s.semanticSearch,
		"hybrid_search":                s.hybridSearch,
		"find_similar_tasks":           s.findSimilarTasks,
		"recommend_tasks":              s.recommendTasks,
		"cluster_project_tasks":        s.clusterProjectTasks,
		"search_projects":              s.searchProjects,
		"search_documents":             s.searchDocuments,
		"intelligent_task_assignment":  s.intelligentTaskAssignment,

		// Analytics
		"get_burndown_chart": s.getBurndownChart,
		"get_activity_feed":  s.getActivityFeed,

		// Labels
		"create_label": s.createLabel,
		"assign_label": s.assignLabel,

		// Custom Fields
		"create_custom_field":     s.createCustomField,
		"set_custom_field_value":  s.setCustomFieldValue,
	}
}

// Milestone Management implementations
func (s *EnhancedMCPServer) createMilestone(args []byte) (*ToolResponse, error) {
	var input struct {
		ProjectID   uint   `json:"project_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		DueDate     string `json:"due_date"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	milestone := &models.Milestone{
		ProjectID:   input.ProjectID,
		Name:        input.Name,
		Description: input.Description,
		Status:      models.MilestoneStatusPlanned,
	}

	if input.DueDate != "" {
		dueDate, err := parseDate(input.DueDate)
		if err != nil {
			return ErrorResponse(err), nil
		}
		milestone.DueDate = dueDate
	}

	if err := s.db.Create(milestone).Error; err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(milestone), nil
}

func (s *EnhancedMCPServer) updateMilestone(args []byte) (*ToolResponse, error) {
	var input struct {
		ID          uint   `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	var milestone models.Milestone
	if err := s.db.First(&milestone, input.ID).Error; err != nil {
		return ErrorResponse(ErrTaskNotFound), nil
	}

	if input.Name != "" {
		milestone.Name = input.Name
	}
	if input.Description != "" {
		milestone.Description = input.Description
	}
	if input.Status != "" {
		milestone.Status = models.MilestoneStatus(input.Status)
	}

	if err := s.db.Save(&milestone).Error; err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(milestone), nil
}

func (s *EnhancedMCPServer) listMilestones(args []byte) (*ToolResponse, error) {
	var input struct {
		ProjectID uint `json:"project_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	var milestones []models.Milestone
	query := s.db.DB
	if input.ProjectID > 0 {
		query = query.Where("project_id = ?", input.ProjectID)
	}

	if err := query.Preload("Tasks").Find(&milestones).Error; err != nil {
		return ErrorResponse(err), nil
	}

	// Add completion metrics
	type MilestoneWithMetrics struct {
		models.Milestone
		TaskCount      int     `json:"task_count"`
		CompletedTasks int     `json:"completed_tasks"`
		Progress       float64 `json:"progress"`
	}

	result := make([]MilestoneWithMetrics, len(milestones))
	for i, m := range milestones {
		completed := 0
		for _, task := range m.Tasks {
			if task.Status == models.TaskStatusDone {
				completed++
			}
		}

		result[i] = MilestoneWithMetrics{
			Milestone:      m,
			TaskCount:      len(m.Tasks),
			CompletedTasks: completed,
			Progress:       calculateCompletionPercentage(completed, len(m.Tasks)),
		}
	}

	return SuccessResponse(result), nil
}

// Sprint Management implementations
func (s *EnhancedMCPServer) createSprint(args []byte) (*ToolResponse, error) {
	var input struct {
		ProjectID uint   `json:"project_id"`
		Name      string `json:"name"`
		Goal      string `json:"goal"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	startDate, err := parseDate(input.StartDate)
	if err != nil {
		return ErrorResponse(err), nil
	}

	endDate, err := parseDate(input.EndDate)
	if err != nil {
		return ErrorResponse(err), nil
	}

	if endDate.Before(startDate) {
		return ErrorResponse(fmt.Errorf("end date must be after start date")), nil
	}

	sprint := &models.Sprint{
		ProjectID: input.ProjectID,
		Name:      input.Name,
		Goal:      input.Goal,
		StartDate: startDate,
		EndDate:   endDate,
		Status:    models.SprintStatusPlanned,
	}

	if err := s.db.Create(sprint).Error; err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(sprint), nil
}

func (s *EnhancedMCPServer) startSprint(args []byte) (*ToolResponse, error) {
	var input struct {
		ID uint `json:"id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	var sprint models.Sprint
	if err := s.db.First(&sprint, input.ID).Error; err != nil {
		return ErrorResponse(ErrSprintNotFound), nil
	}

	// Check if another sprint is active in the project
	var activeCount int64
	s.db.Model(&models.Sprint{}).
		Where("project_id = ? AND status = ? AND id != ?",
			sprint.ProjectID, models.SprintStatusActive, sprint.ID).
		Count(&activeCount)

	if activeCount > 0 {
		return ErrorResponse(ErrSprintAlreadyActive), nil
	}

	sprint.Status = models.SprintStatusActive
	if err := s.db.Save(&sprint).Error; err != nil {
		return ErrorResponse(err), nil
	}

	// Track activity
	s.trackActivity("sprint_started", sprint.ID, 0)

	return SuccessResponse(sprint), nil
}

func (s *EnhancedMCPServer) completeSprint(args []byte) (*ToolResponse, error) {
	var input struct {
		ID uint `json:"id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	var sprint models.Sprint
	if err := s.db.First(&sprint, input.ID).Error; err != nil {
		return ErrorResponse(ErrSprintNotFound), nil
	}

	sprint.Status = models.SprintStatusCompleted

	// Move incomplete tasks to backlog or next sprint
	var incompleteTasks []models.Task
	s.db.Where("sprint_id = ? AND status != ?", sprint.ID, models.TaskStatusDone).
		Find(&incompleteTasks)

	for _, task := range incompleteTasks {
		task.SprintID = nil // Move to backlog
		s.db.Save(&task)
	}

	if err := s.db.Save(&sprint).Error; err != nil {
		return ErrorResponse(err), nil
	}

	// Track activity
	s.trackActivity("sprint_completed", sprint.ID, 0)

	return SuccessResponse(map[string]interface{}{
		"sprint":             sprint,
		"incomplete_tasks":   len(incompleteTasks),
		"moved_to_backlog":   len(incompleteTasks),
	}), nil
}

func (s *EnhancedMCPServer) listSprints(args []byte) (*ToolResponse, error) {
	var input struct {
		ProjectID uint   `json:"project_id"`
		Status    string `json:"status"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	query := s.db.Where("project_id = ?", input.ProjectID)
	if input.Status != "" {
		query = query.Where("status = ?", input.Status)
	}

	var sprints []models.Sprint
	if err := query.
		Preload("Tasks").
		Order("start_date DESC").
		Find(&sprints).Error; err != nil {
		return ErrorResponse(err), nil
	}

	// Add metrics to each sprint
	type SprintWithMetrics struct {
		models.Sprint
		TaskCount      int     `json:"task_count"`
		CompletedTasks int     `json:"completed_tasks"`
		Progress       float64 `json:"progress"`
		DaysRemaining  int     `json:"days_remaining"`
	}

	result := make([]SprintWithMetrics, len(sprints))
	for i, sprint := range sprints {
		completed := 0
		for _, task := range sprint.Tasks {
			if task.Status == models.TaskStatusDone {
				completed++
			}
		}

		result[i] = SprintWithMetrics{
			Sprint:         sprint,
			TaskCount:      len(sprint.Tasks),
			CompletedTasks: completed,
			Progress:       calculateCompletionPercentage(completed, len(sprint.Tasks)),
			DaysRemaining:  calculateDaysRemaining(sprint.EndDate),
		}
	}

	return SuccessResponse(result), nil
}

// Workflow Management implementations
func (s *EnhancedMCPServer) createWorkflow(args []byte) (*ToolResponse, error) {
	var input struct {
		ProjectID   uint   `json:"project_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	workflow := &models.Workflow{
		ProjectID:   input.ProjectID,
		Name:        input.Name,
		Description: input.Description,
		IsActive:    true,
	}

	if err := s.db.Create(workflow).Error; err != nil {
		return ErrorResponse(err), nil
	}

	// Create default workflow states
	defaultStates := []models.WorkflowState{
		{WorkflowID: workflow.ID, Name: "To Do", Type: models.WorkflowStateTypeStart, Order: 1},
		{WorkflowID: workflow.ID, Name: "In Progress", Type: models.WorkflowStateTypeInProgress, Order: 2},
		{WorkflowID: workflow.ID, Name: "Review", Type: models.WorkflowStateTypeInProgress, Order: 3},
		{WorkflowID: workflow.ID, Name: "Done", Type: models.WorkflowStateTypeDone, Order: 4},
	}

	for _, state := range defaultStates {
		s.db.Create(&state)
	}

	return SuccessResponse(workflow), nil
}

func (s *EnhancedMCPServer) addWorkflowState(args []byte) (*ToolResponse, error) {
	var input struct {
		WorkflowID uint   `json:"workflow_id"`
		Name       string `json:"name"`
		Type       string `json:"type"`
		Order      int    `json:"order"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	state := &models.WorkflowState{
		WorkflowID: input.WorkflowID,
		Name:       input.Name,
		Type:       models.WorkflowStateType(input.Type),
		Order:      input.Order,
	}

	if err := s.db.Create(state).Error; err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(state), nil
}

// Additional tool implementations...
func (s *EnhancedMCPServer) listComments(args []byte) (*ToolResponse, error) {
	var input struct {
		TaskID uint `json:"task_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	var comments []models.Comment
	if err := s.db.
		Where("task_id = ?", input.TaskID).
		Order("created_at DESC").
		Find(&comments).Error; err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(comments), nil
}

func (s *EnhancedMCPServer) logTime(args []byte) (*ToolResponse, error) {
	var input struct {
		TaskID      uint    `json:"task_id"`
		UserID      uint    `json:"user_id"`
		Hours       float64 `json:"hours"`
		Description string  `json:"description"`
		Date        string  `json:"date"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	entry := &models.TimeEntry{
		TaskID:      input.TaskID,
		UserID:      input.UserID,
		Hours:       input.Hours,
		Description: input.Description,
	}

	if input.Date != "" {
		date, err := parseDate(input.Date)
		if err != nil {
			return ErrorResponse(err), nil
		}
		entry.Date = date
	} else {
		entry.Date = time.Now()
	}

	if err := s.db.Create(entry).Error; err != nil {
		return ErrorResponse(err), nil
	}

	// Track activity
	s.trackActivity("time_logged", input.TaskID, input.UserID)

	return SuccessResponse(entry), nil
}

func (s *EnhancedMCPServer) getTimeEntries(args []byte) (*ToolResponse, error) {
	var input struct {
		TaskID    uint   `json:"task_id"`
		UserID    uint   `json:"user_id"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}
	_ = UnmarshalArgs(args, &input)

	query := s.db.DB
	if input.TaskID > 0 {
		query = query.Where("task_id = ?", input.TaskID)
	}
	if input.UserID > 0 {
		query = query.Where("user_id = ?", input.UserID)
	}
	if input.StartDate != "" {
		startDate, _ := parseDate(input.StartDate)
		query = query.Where("date >= ?", startDate)
	}
	if input.EndDate != "" {
		endDate, _ := parseDate(input.EndDate)
		query = query.Where("date <= ?", endDate)
	}

	var entries []models.TimeEntry
	if err := query.
		Order("date DESC").
		Find(&entries).Error; err != nil {
		return ErrorResponse(err), nil
	}

	// Calculate totals
	totalHours := 0.0
	for _, entry := range entries {
		totalHours += entry.Hours
	}

	return SuccessResponse(map[string]interface{}{
		"entries":     entries,
		"total_hours": totalHours,
		"count":       len(entries),
	}), nil
}

func (s *EnhancedMCPServer) createNotification(args []byte) (*ToolResponse, error) {
	var input struct {
		UserID  uint   `json:"user_id"`
		Type    string `json:"type"`
		Title   string `json:"title"`
		Message string `json:"message"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	notification := &models.Notification{
		UserID:  input.UserID,
		Type:    input.Type,
		Title:   input.Title,
		Message: input.Message,
	}

	if err := s.db.Create(notification).Error; err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(notification), nil
}

func (s *EnhancedMCPServer) listNotifications(args []byte) (*ToolResponse, error) {
	var input struct {
		UserID uint  `json:"user_id"`
		Read   *bool `json:"read"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	query := s.db.Where("user_id = ?", input.UserID)
	if input.Read != nil {
		query = query.Where("read = ?", *input.Read)
	}

	var notifications []models.Notification
	if err := query.
		Order("created_at DESC").
		Limit(100).
		Find(&notifications).Error; err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(notifications), nil
}

func (s *EnhancedMCPServer) markNotificationRead(args []byte) (*ToolResponse, error) {
	var input struct {
		ID uint `json:"id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	if err := s.db.Model(&models.Notification{}).
		Where("id = ?", input.ID).
		Update("read", true).Error; err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(map[string]interface{}{
		"success": true,
		"id":      input.ID,
	}), nil
}

func (s *EnhancedMCPServer) createWebhook(args []byte) (*ToolResponse, error) {
	var input struct {
		ProjectID uint     `json:"project_id"`
		URL       string   `json:"url"`
		Events    []string `json:"events"`
		Active    bool     `json:"active"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	eventsJSON, _ := json.Marshal(input.Events)

	webhook := &models.Webhook{
		ProjectID: input.ProjectID,
		URL:       input.URL,
		Events:    string(eventsJSON),
		Secret:    generateWebhookSecret(),
	}

	if err := s.db.Create(webhook).Error; err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(webhook), nil
}

func (s *EnhancedMCPServer) listWebhooks(args []byte) (*ToolResponse, error) {
	var input struct {
		ProjectID uint `json:"project_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	var webhooks []models.Webhook
	if err := s.db.
		Where("project_id = ?", input.ProjectID).
		Order("created_at DESC").
		Find(&webhooks).Error; err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(webhooks), nil
}

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
		label.Color = generateLabelColor(label.Name)
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

func (s *EnhancedMCPServer) createCustomField(args []byte) (*ToolResponse, error) {
	var input struct {
		ProjectID uint     `json:"project_id"`
		Name      string   `json:"name"`
		FieldType string   `json:"field_type"`
		Options   []string `json:"options"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	optionsJSON, _ := json.Marshal(input.Options)

	field := &models.CustomField{
		ProjectID: input.ProjectID,
		Name:      input.Name,
		FieldType: input.FieldType,
		Options:   string(optionsJSON),
	}

	if err := s.db.Create(field).Error; err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(field), nil
}

func (s *EnhancedMCPServer) setCustomFieldValue(args []byte) (*ToolResponse, error) {
	var input struct {
		TaskID  uint   `json:"task_id"`
		FieldID uint   `json:"field_id"`
		Value   string `json:"value"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	// Check if value already exists
	var existing models.FieldValue
	err := s.db.Where("task_id = ? AND field_id = ?", input.TaskID, input.FieldID).
		First(&existing).Error

	if err == nil {
		// Update existing value
		existing.Value = input.Value
		if err := s.db.Save(&existing).Error; err != nil {
			return ErrorResponse(err), nil
		}
		return SuccessResponse(existing), nil
	}

	// Create new value
	value := &models.FieldValue{
		TaskID:  input.TaskID,
		FieldID: input.FieldID,
		Value:   input.Value,
	}

	if err := s.db.Create(value).Error; err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(value), nil
}