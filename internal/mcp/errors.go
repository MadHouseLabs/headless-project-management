package mcp

import "errors"

// Common errors used across MCP tools
var (
	// Entity not found errors
	ErrProjectNotFound = errors.New("project not found")
	ErrTaskNotFound    = errors.New("task not found")
	ErrUserNotFound    = errors.New("user not found")
	ErrLabelNotFound   = errors.New("label not found")

	// Configuration errors
	ErrDatabaseNotConfigured = errors.New("database not configured")

	// Validation errors
	ErrInvalidInput    = errors.New("invalid input parameters")
	ErrMissingRequired = errors.New("missing required parameters")

	// Business logic errors
	ErrDuplicateEntry = errors.New("duplicate entry already exists")

	// System errors
	ErrDatabaseOperation = errors.New("database operation failed")
)