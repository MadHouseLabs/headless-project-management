package mcp

import (
	"fmt"

	"github.com/headless-pm/headless-project-management/internal/models"
)

// Extended Search Tool Implementations

func (s *EnhancedMCPServer) hybridSearch(args []byte) (*ToolResponse, error) {
	var input struct {
		Query      string `json:"query"`
		EntityType string `json:"entity_type"`
		Limit      int    `json:"limit"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	if input.Query == "" {
		return ErrorResponse(fmt.Errorf("query is required")), nil
	}

	if input.EntityType == "" {
		input.EntityType = "task"
	}
	if input.Limit == 0 {
		input.Limit = 20
	}

	results, err := s.vectorService.HybridSearch(input.Query, input.EntityType, input.Limit)
	if err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(map[string]interface{}{
		"query":   input.Query,
		"results": results,
		"count":   len(results),
	}), nil
}

// Note: findSimilarTasks and recommendTasks are defined in tools_ai.go

func (s *EnhancedMCPServer) clusterProjectTasks(args []byte) (*ToolResponse, error) {
	var input struct {
		ProjectID   uint `json:"project_id"`
		NumClusters int  `json:"num_clusters"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	if input.ProjectID == 0 {
		return ErrorResponse(fmt.Errorf("project_id is required")), nil
	}

	if input.NumClusters == 0 {
		input.NumClusters = 5
	}

	clusters, err := s.vectorService.ClusterTasks(input.ProjectID, input.NumClusters)
	if err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(map[string]interface{}{
		"project_id":   input.ProjectID,
		"clusters":     clusters,
		"num_clusters": input.NumClusters,
	}), nil
}

func (s *EnhancedMCPServer) searchProjects(args []byte) (*ToolResponse, error) {
	var input struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	if input.Query == "" {
		return ErrorResponse(fmt.Errorf("query is required")), nil
	}

	if input.Limit == 0 {
		input.Limit = 10
	}

	// Use semantic search for projects
	results, err := s.vectorService.SemanticSearch(input.Query, "project", input.Limit)
	if err != nil {
		return ErrorResponse(err), nil
	}

	// Also do keyword search on project names and descriptions
	searchPattern := "%" + input.Query + "%"
	var projects []models.Project
	if err := s.db.Where("name LIKE ? OR description LIKE ?", searchPattern, searchPattern).
		Limit(input.Limit).
		Find(&projects).Error; err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(map[string]interface{}{
		"query":            input.Query,
		"semantic_results": results,
		"keyword_results":  projects,
	}), nil
}

func (s *EnhancedMCPServer) searchDocuments(args []byte) (*ToolResponse, error) {
	var input struct {
		Query      string `json:"query"`
		ProjectID  uint   `json:"project_id"`
		DocumentID uint   `json:"document_id"`
		Limit      int    `json:"limit"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	if input.Query == "" {
		return ErrorResponse(fmt.Errorf("query is required")), nil
	}

	if input.Limit == 0 {
		input.Limit = 10
	}

	// Build query for document search
	query := s.db.DB.Model(&models.DocumentEmbedding{})

	if input.ProjectID > 0 {
		query = query.Where("project_id = ?", input.ProjectID)
	}
	if input.DocumentID > 0 {
		query = query.Where("document_id = ?", input.DocumentID)
	}

	// Use semantic search for documents
	results, err := s.vectorService.SemanticSearch(input.Query, "document", input.Limit)
	if err != nil {
		return ErrorResponse(err), nil
	}

	// Also search in document metadata
	searchPattern := "%" + input.Query + "%"
	var documents []models.DocumentEmbedding
	if err := query.Where("content LIKE ? OR metadata LIKE ?", searchPattern, searchPattern).
		Limit(input.Limit).
		Find(&documents).Error; err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(map[string]interface{}{
		"query":            input.Query,
		"semantic_results": results,
		"keyword_results":  documents,
	}), nil
}

func (s *EnhancedMCPServer) intelligentTaskAssignment(args []byte) (*ToolResponse, error) {
	var input struct {
		TaskID uint `json:"task_id"`
		TeamID uint `json:"team_id"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	// Get task details
	var task models.Task
	if err := s.db.First(&task, input.TaskID).Error; err != nil {
		return ErrorResponse(err), nil
	}

	// Find similar completed tasks
	similarTasks, err := s.vectorService.FindSimilarTasks(input.TaskID, 10)
	if err != nil {
		return ErrorResponse(err), nil
	}

	// Analyze who completed similar tasks successfully
	assigneeStats := make(map[uint]int)
	for _, similarTask := range similarTasks {
		if similarTask.AssigneeID != nil && similarTask.Status == models.TaskStatusDone {
			assigneeStats[*similarTask.AssigneeID]++
		}
	}

	// Get team members if team is specified
	var teamMembers []models.TeamMember
	if input.TeamID > 0 {
		if err := s.db.Where("team_id = ?", input.TeamID).Find(&teamMembers).Error; err != nil {
			return ErrorResponse(err), nil
		}
	}

	// Calculate workload for potential assignees
	type AssigneeRecommendation struct {
		UserID          uint    `json:"user_id"`
		Username        string  `json:"username"`
		SimilarTasksDone int    `json:"similar_tasks_done"`
		CurrentWorkload int     `json:"current_workload"`
		Score           float64 `json:"score"`
	}

	var recommendations []AssigneeRecommendation

	// Process team members or top performers
	userIDs := make(map[uint]bool)
	if len(teamMembers) > 0 {
		for _, member := range teamMembers {
			userIDs[member.UserID] = true
		}
	} else {
		// Use top performers from similar tasks
		for userID := range assigneeStats {
			userIDs[userID] = true
		}
	}

	for userID := range userIDs {
		var user models.User
		if err := s.db.First(&user, userID).Error; err != nil {
			continue
		}

		// Count current active tasks
		var activeTaskCount int64
		s.db.Model(&models.Task{}).
			Where("assignee_id = ? AND status IN ?", userID,
				[]string{string(models.TaskStatusTodo), string(models.TaskStatusInProgress)}).
			Count(&activeTaskCount)

		// Calculate score (higher is better)
		similarDone := assigneeStats[userID]
		score := float64(similarDone*10) - float64(activeTaskCount*2)

		recommendations = append(recommendations, AssigneeRecommendation{
			UserID:          userID,
			Username:        user.Username,
			SimilarTasksDone: similarDone,
			CurrentWorkload: int(activeTaskCount),
			Score:           score,
		})
	}

	// Sort by score
	for i := 0; i < len(recommendations)-1; i++ {
		for j := i + 1; j < len(recommendations); j++ {
			if recommendations[j].Score > recommendations[i].Score {
				recommendations[i], recommendations[j] = recommendations[j], recommendations[i]
			}
		}
	}

	return SuccessResponse(map[string]interface{}{
		"task_id":          input.TaskID,
		"recommendations": recommendations,
		"analysis": map[string]interface{}{
			"similar_tasks_analyzed": len(similarTasks),
			"team_members_considered": len(teamMembers),
		},
	}), nil
}