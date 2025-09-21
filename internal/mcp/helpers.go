package mcp

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/headless-pm/headless-project-management/internal/models"
)

// Activity tracking helper
func (s *EnhancedMCPServer) trackActivity(action string, entityID uint, userID uint) {
	activity := &models.Activity{
		UserID:     userID,
		Action:     action,
		EntityType: inferEntityType(action),
		EntityID:   entityID,
	}

	// Best effort - don't fail operations if activity tracking fails
	_ = s.db.Create(activity)
}

// Notification helper
func (s *EnhancedMCPServer) createNotificationForUser(userID uint, notifType, title, message string) {
	notification := &models.Notification{
		UserID:  userID,
		Type:    notifType,
		Title:   title,
		Message: message,
	}

	// Best effort - don't fail operations if notification fails
	_ = s.db.Create(notification)
}

// Generate secure random token
func generateSecureToken(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based token
		return hex.EncodeToString([]byte(time.Now().String()))[:length*2]
	}
	return hex.EncodeToString(bytes)
}

// Generate webhook secret
func generateWebhookSecret() string {
	return generateSecureToken(16)
}

// Infer entity type from action
func inferEntityType(action string) string {
	switch {
	case contains(action, "task"):
		return "task"
	case contains(action, "project"):
		return "project"
	case contains(action, "sprint"):
		return "sprint"
	case contains(action, "milestone"):
		return "milestone"
	case contains(action, "team"):
		return "team"
	default:
		return "unknown"
	}
}

// Simple string contains helper
func contains(str, substr string) bool {
	return len(str) >= len(substr) && str[:len(substr)] == substr
}

// Parse date with multiple format support
func parseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, ErrInvalidDateFormat
}

// Calculate completion percentage
func calculateCompletionPercentage(completed, total int) float64 {
	if total == 0 {
		return 0
	}
	return (float64(completed) / float64(total)) * 100
}

// Format duration in human-readable format
func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24

	if days > 0 {
		if hours > 0 {
			return formatPlural(days, "day") + " " + formatPlural(hours, "hour")
		}
		return formatPlural(days, "day")
	}

	if hours > 0 {
		return formatPlural(hours, "hour")
	}

	minutes := int(d.Minutes())
	return formatPlural(minutes, "minute")
}

func formatPlural(count int, singular string) string {
	if count == 1 {
		return "1 " + singular
	}
	return string(rune(count)) + " " + singular + "s"
}

// Batch processing helper for large datasets
func processBatch(total int, batchSize int, processor func(offset, limit int) error) error {
	for offset := 0; offset < total; offset += batchSize {
		limit := batchSize
		if offset+limit > total {
			limit = total - offset
		}

		if err := processor(offset, limit); err != nil {
			return err
		}
	}
	return nil
}

// Safe string truncation
func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	if maxLength <= 3 {
		return s[:maxLength]
	}
	return s[:maxLength-3] + "..."
}

// Merge maps (second overwrites first)
func mergeMaps(first, second map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range first {
		result[k] = v
	}

	for k, v := range second {
		result[k] = v
	}

	return result
}