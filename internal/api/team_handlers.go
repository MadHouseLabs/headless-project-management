package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/headless-pm/headless-project-management/internal/database"
	"github.com/headless-pm/headless-project-management/internal/models"
)

type TeamHandler struct {
	db *database.Database
}

func NewTeamHandler(db *database.Database) *TeamHandler {
	return &TeamHandler{db: db}
}

func (h *TeamHandler) CreateTeam(c *gin.Context) {
	userID := c.GetUint("userID")

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	team := &models.Team{
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     userID,
	}

	if err := h.db.Create(team).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create team"})
		return
	}

	h.db.Model(team).Association("Members").Append(&models.User{ID: userID})

	c.JSON(http.StatusCreated, team)
}

func (h *TeamHandler) GetTeam(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	var team models.Team
	if err := h.db.Preload("Owner").Preload("Members").Preload("Projects").First(&team, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	c.JSON(http.StatusOK, team)
}

func (h *TeamHandler) ListTeams(c *gin.Context) {
	userID := c.GetUint("userID")

	var teams []models.Team
	if err := h.db.Joins("JOIN team_members ON teams.id = team_members.team_id").
		Where("team_members.user_id = ?", userID).
		Preload("Owner").
		Find(&teams).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list teams"})
		return
	}

	c.JSON(http.StatusOK, teams)
}

func (h *TeamHandler) UpdateTeam(c *gin.Context) {
	userID := c.GetUint("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	var team models.Team
	if err := h.db.First(&team, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	if team.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team owner can update team"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
		"updated_at":  time.Now(),
	}

	if err := h.db.Model(&team).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update team"})
		return
	}

	c.JSON(http.StatusOK, team)
}

func (h *TeamHandler) DeleteTeam(c *gin.Context) {
	userID := c.GetUint("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	var team models.Team
	if err := h.db.First(&team, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	if team.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team owner can delete team"})
		return
	}

	if err := h.db.Delete(&team).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete team"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *TeamHandler) AddMember(c *gin.Context) {
	userID := c.GetUint("userID")
	teamID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	var team models.Team
	if err := h.db.First(&team, teamID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	var member models.TeamMember
	if err := h.db.Where("team_id = ? AND user_id = ?", teamID, userID).First(&member).Error; err == nil {
		if member.Role != models.TeamRoleOwner && member.Role != models.TeamRoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			return
		}
	}

	var req struct {
		UserID uint   `json:"user_id" binding:"required"`
		Role   string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	teamMember := models.TeamMember{
		TeamID:   uint(teamID),
		UserID:   req.UserID,
		Role:     models.TeamRoleMember,
		JoinedAt: time.Now(),
	}

	if req.Role != "" {
		teamMember.Role = models.TeamRole(req.Role)
	}

	if err := h.db.Create(&teamMember).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Member added successfully"})
}

func (h *TeamHandler) RemoveMember(c *gin.Context) {
	userID := c.GetUint("userID")
	teamID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	memberID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var team models.Team
	if err := h.db.First(&team, teamID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	if team.OwnerID != userID {
		var member models.TeamMember
		if err := h.db.Where("team_id = ? AND user_id = ?", teamID, userID).First(&member).Error; err != nil || member.Role != models.TeamRoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			return
		}
	}

	if uint(memberID) == team.OwnerID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot remove team owner"})
		return
	}

	if err := h.db.Where("team_id = ? AND user_id = ?", teamID, memberID).Delete(&models.TeamMember{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed successfully"})
}