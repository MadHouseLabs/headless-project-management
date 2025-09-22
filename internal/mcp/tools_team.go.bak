package mcp

import (
	"github.com/headless-pm/headless-project-management/internal/models"
)

// Team Management Tool Implementations

func (s *EnhancedMCPServer) createTeam(args []byte) (*ToolResponse, error) {
	var input TeamInput
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	team := &models.Team{
		Name:        input.Name,
		Description: input.Description,
	}

	if err := s.db.Create(team).Error; err != nil {
		return ErrorResponse(err), nil
	}

	return SuccessResponse(team), nil
}

func (s *EnhancedMCPServer) addTeamMember(args []byte) (*ToolResponse, error) {
	var input struct {
		TeamID uint   `json:"team_id"`
		UserID uint   `json:"user_id"`
		Role   string `json:"role"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	// Validate team exists
	var team models.Team
	if err := s.db.First(&team, input.TeamID).Error; err != nil {
		return ErrorResponse(ErrTeamNotFound), nil
	}

	// Validate user exists
	var user models.User
	if err := s.db.First(&user, input.UserID).Error; err != nil {
		return ErrorResponse(ErrUserNotFound), nil
	}

	// Check if member already exists
	var existing models.TeamMember
	if err := s.db.Where("team_id = ? AND user_id = ?", input.TeamID, input.UserID).
		First(&existing).Error; err == nil {
		// Update role if member exists
		existing.Role = models.TeamRole(input.Role)
		if err := s.db.Save(&existing).Error; err != nil {
			return ErrorResponse(err), nil
		}
		return SuccessResponse(existing), nil
	}

	member := &models.TeamMember{
		TeamID: input.TeamID,
		UserID: input.UserID,
		Role:   models.TeamRole(input.Role),
	}

	if member.Role == "" {
		member.Role = models.TeamRoleMember
	}

	if err := s.db.Create(member).Error; err != nil {
		return ErrorResponse(err), nil
	}

	// Send notification to the user
	s.createNotificationForUser(input.UserID, "team_invite",
		"Added to team", "You've been added to team: "+team.Name)

	return SuccessResponse(member), nil
}

func (s *EnhancedMCPServer) listTeams(args []byte) (*ToolResponse, error) {
	var teams []models.Team
	if err := s.db.
		Preload("Members.User").
		Order("created_at DESC").
		Find(&teams).Error; err != nil {
		return ErrorResponse(err), nil
	}

	// Format response with member counts
	type TeamResponse struct {
		models.Team
		MemberCount int `json:"member_count"`
	}

	response := make([]TeamResponse, len(teams))
	for i, team := range teams {
		response[i] = TeamResponse{
			Team:        team,
			MemberCount: len(team.Members),
		}
	}

	return SuccessResponse(response), nil
}

func (s *EnhancedMCPServer) getTeamVelocity(args []byte) (*ToolResponse, error) {
	var input struct {
		TeamID  uint `json:"team_id"`
		Sprints int  `json:"sprints"`
	}
	if err := UnmarshalArgs(args, &input); err != nil {
		return ErrorResponse(err), nil
	}

	if input.Sprints == 0 {
		input.Sprints = 5
	}

	// Get completed sprints for team's projects
	var sprintStats []struct {
		SprintID   uint    `json:"sprint_id"`
		SprintName string  `json:"sprint_name"`
		Completed  int     `json:"completed"`
		Total      int     `json:"total"`
		Points     float64 `json:"story_points"`
	}

	err := s.db.Table("sprints").
		Select(`
			sprints.id as sprint_id,
			sprints.name as sprint_name,
			COUNT(DISTINCT CASE WHEN tasks.status = ? THEN tasks.id END) as completed,
			COUNT(DISTINCT tasks.id) as total
		`, models.TaskStatusDone).
		Joins("JOIN projects ON projects.id = sprints.project_id").
		Joins("LEFT JOIN tasks ON tasks.sprint_id = sprints.id").
		Where("projects.team_id = ? AND sprints.status = ?", input.TeamID, models.SprintStatusCompleted).
		Group("sprints.id, sprints.name").
		Order("sprints.end_date DESC").
		Limit(input.Sprints).
		Scan(&sprintStats).Error

	if err != nil {
		return ErrorResponse(err), nil
	}

	// Calculate average velocity
	totalCompleted := 0
	for _, stat := range sprintStats {
		totalCompleted += stat.Completed
	}

	avgVelocity := 0.0
	if len(sprintStats) > 0 {
		avgVelocity = float64(totalCompleted) / float64(len(sprintStats))
	}

	// Get current sprint status
	var currentSprint struct {
		ID        uint   `json:"id"`
		Name      string `json:"name"`
		Completed int    `json:"completed"`
		Total     int    `json:"total"`
	}

	s.db.Table("sprints").
		Select(`
			sprints.id,
			sprints.name,
			COUNT(DISTINCT CASE WHEN tasks.status = ? THEN tasks.id END) as completed,
			COUNT(DISTINCT tasks.id) as total
		`, models.TaskStatusDone).
		Joins("JOIN projects ON projects.id = sprints.project_id").
		Joins("LEFT JOIN tasks ON tasks.sprint_id = sprints.id").
		Where("projects.team_id = ? AND sprints.status = ?", input.TeamID, models.SprintStatusActive).
		Group("sprints.id, sprints.name").
		First(&currentSprint)

	return SuccessResponse(map[string]interface{}{
		"team_id":          input.TeamID,
		"past_sprints":     sprintStats,
		"average_velocity": avgVelocity,
		"current_sprint":   currentSprint,
		"trend": calculateVelocityTrend(sprintStats),
	}), nil
}

// Helper to calculate velocity trend
func calculateVelocityTrend(stats []struct {
	SprintID   uint    `json:"sprint_id"`
	SprintName string  `json:"sprint_name"`
	Completed  int     `json:"completed"`
	Total      int     `json:"total"`
	Points     float64 `json:"story_points"`
}) string {
	if len(stats) < 2 {
		return "insufficient_data"
	}

	// Compare recent vs older sprints
	recentAvg := float64(stats[0].Completed)
	if len(stats) > 1 {
		recentAvg = (float64(stats[0].Completed) + float64(stats[1].Completed)) / 2
	}

	olderAvg := 0.0
	olderCount := 0
	for i := 2; i < len(stats); i++ {
		olderAvg += float64(stats[i].Completed)
		olderCount++
	}

	if olderCount > 0 {
		olderAvg /= float64(olderCount)
		if recentAvg > olderAvg*1.1 {
			return "improving"
		} else if recentAvg < olderAvg*0.9 {
			return "declining"
		}
	}

	return "stable"
}