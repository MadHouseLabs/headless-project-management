package mcp

import "errors"

// Common errors used across MCP tools
var (
	// Entity not found errors
	ErrProjectNotFound = errors.New("project not found")
	ErrTaskNotFound    = errors.New("task not found")
	ErrUserNotFound    = errors.New("user not found")
	ErrTeamNotFound    = errors.New("team not found")
	ErrSprintNotFound  = errors.New("sprint not found")
	ErrLabelNotFound   = errors.New("label not found")

	// Configuration errors
	ErrEmbeddingsNotConfigured = errors.New("embeddings not configured - semantic features unavailable")
	ErrDatabaseNotConfigured   = errors.New("database not configured")

	// Validation errors
	ErrInvalidInput      = errors.New("invalid input parameters")
	ErrMissingRequired   = errors.New("missing required parameters")
	ErrInvalidDateFormat = errors.New("invalid date format - use YYYY-MM-DD or ISO 8601")
	ErrInvalidStatus     = errors.New("invalid status value")
	ErrInvalidPriority   = errors.New("invalid priority value")

	// Business logic errors
	ErrDuplicateEntry      = errors.New("duplicate entry already exists")
	ErrCircularDependency  = errors.New("circular dependency detected")
	ErrInsufficientRights  = errors.New("insufficient permissions for this operation")
	ErrSprintAlreadyActive = errors.New("another sprint is already active")
	ErrCannotDeleteActive  = errors.New("cannot delete active entity")

	// System errors
	ErrDatabaseOperation = errors.New("database operation failed")
	ErrJSONMarshal       = errors.New("failed to marshal JSON")
	ErrJSONUnmarshal     = errors.New("failed to unmarshal JSON")
)

// ValidateRequired checks if required fields are present
func ValidateRequired(fields map[string]interface{}, required []string) error {
	for _, field := range required {
		if val, ok := fields[field]; !ok || val == nil || val == "" || val == 0 {
			return errors.New("missing required field: " + field)
		}
	}
	return nil
}

// ValidateEnum checks if a value is within allowed options
func ValidateEnum(value string, allowed []string) error {
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return errors.New("invalid value: " + value + " (allowed: " + joinStrings(allowed))
}

func joinStrings(strings []string) string {
	if len(strings) == 0 {
		return ""
	}
	result := strings[0]
	for i := 1; i < len(strings); i++ {
		result += ", " + strings[i]
	}
	return result
}