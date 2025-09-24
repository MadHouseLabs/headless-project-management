package mcp

func toolDefinitions() []Tool {
	return []Tool{
		// Project Management (5 tools)
		{
			Name:        "create_project",
			Description: "Create a new project",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name":        map[string]string{"type": "string"},
					"description": map[string]string{"type": "string"},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "get_project",
			Description: "Get project details by ID",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id": map[string]string{"type": "number"},
				},
				"required": []string{"project_id"},
			},
		},
		{
			Name:        "update_project",
			Description: "Update an existing project",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id":  map[string]string{"type": "number"},
					"name":        map[string]string{"type": "string"},
					"description": map[string]string{"type": "string"},
					"status":      map[string]string{"type": "string"},
				},
				"required": []string{"project_id"},
			},
		},
		{
			Name:        "delete_project",
			Description: "Delete a project",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id": map[string]string{"type": "number"},
				},
				"required": []string{"project_id"},
			},
		},
		{
			Name:        "list_projects",
			Description: "List all projects",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},

		// Task Management (5 tools)
		{
			Name:        "create_task",
			Description: "Create a new task",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id":  map[string]string{"type": "number"},
					"title":       map[string]string{"type": "string"},
					"description": map[string]string{"type": "string"},
					"status":      map[string]string{"type": "string"},
					"priority":    map[string]string{"type": "string"},
					"assignee_id": map[string]string{"type": "number"},
					"epic_id":     map[string]string{"type": "number"},
					"labels":      map[string]interface{}{"type": "array", "items": map[string]string{"type": "string"}},
				},
				"required": []string{"project_id", "title"},
			},
		},
		{
			Name:        "get_task",
			Description: "Get task details by ID",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]string{"type": "number"},
				},
				"required": []string{"task_id"},
			},
		},
		{
			Name:        "update_task",
			Description: "Update an existing task",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id":     map[string]string{"type": "number"},
					"title":       map[string]string{"type": "string"},
					"description": map[string]string{"type": "string"},
					"status":      map[string]string{"type": "string"},
					"priority":    map[string]string{"type": "string"},
					"assignee_id": map[string]string{"type": "number"},
					"epic_id":     map[string]string{"type": "number"},
					"labels":      map[string]interface{}{"type": "array", "items": map[string]string{"type": "string"}},
				},
				"required": []string{"task_id"},
			},
		},
		{
			Name:        "delete_task",
			Description: "Delete a task",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]string{"type": "number"},
				},
				"required": []string{"task_id"},
			},
		},
		{
			Name:        "list_tasks",
			Description: "List tasks with optional filters",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id":  map[string]string{"type": "number"},
					"assignee_id": map[string]string{"type": "number"},
					"status":      map[string]string{"type": "string"},
					"epic_id":     map[string]string{"type": "number"},
				},
			},
		},

		// Epic Management (5 tools)
		{
			Name:        "create_epic",
			Description: "Create a new epic",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id":  map[string]string{"type": "number"},
					"name":        map[string]string{"type": "string"},
					"description": map[string]string{"type": "string"},
				},
				"required": []string{"project_id", "name"},
			},
		},
		{
			Name:        "get_epic",
			Description: "Get epic details by ID",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"epic_id": map[string]string{"type": "number"},
				},
				"required": []string{"epic_id"},
			},
		},
		{
			Name:        "update_epic",
			Description: "Update an existing epic",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"epic_id":     map[string]string{"type": "number"},
					"name":        map[string]string{"type": "string"},
					"description": map[string]string{"type": "string"},
					"status":      map[string]string{"type": "string"},
				},
				"required": []string{"epic_id"},
			},
		},
		{
			Name:        "delete_epic",
			Description: "Delete an epic. Optionally cascade delete all tasks associated with the epic",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"epic_id":       map[string]string{"type": "number", "description": "ID of the epic to delete"},
					"cascade_tasks": map[string]string{"type": "boolean", "description": "If true, delete all tasks associated with this epic. If false, tasks will have their epic_id set to null"},
				},
				"required": []string{"epic_id"},
			},
		},
		{
			Name:        "list_epics",
			Description: "List epics for a project",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id": map[string]string{"type": "number"},
				},
			},
		},

		// Label Management (5 tools)
		{
			Name:        "create_label",
			Description: "Create a new label",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id": map[string]string{"type": "number"},
					"name":       map[string]string{"type": "string"},
					"color":      map[string]string{"type": "string"},
				},
				"required": []string{"project_id", "name"},
			},
		},
		{
			Name:        "assign_label",
			Description: "Assign a label to a task",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id":  map[string]string{"type": "number"},
					"label_id": map[string]string{"type": "number"},
				},
				"required": []string{"task_id", "label_id"},
			},
		},
		{
			Name:        "list_labels",
			Description: "List all labels for a project",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id": map[string]string{"type": "number"},
				},
				"required": []string{"project_id"},
			},
		},
		{
			Name:        "update_label",
			Description: "Update an existing label",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"label_id": map[string]string{"type": "number"},
					"name":     map[string]string{"type": "string"},
					"color":    map[string]string{"type": "string"},
				},
				"required": []string{"label_id"},
			},
		},
		{
			Name:        "delete_label",
			Description: "Delete a label",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"label_id": map[string]string{"type": "number"},
				},
				"required": []string{"label_id"},
			},
		},

		// Assignee Management (2 tools)
		{
			Name:        "assign_task",
			Description: "Assign a task to a user",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id":     map[string]string{"type": "number"},
					"assignee_id": map[string]string{"type": "number"},
				},
				"required": []string{"task_id", "assignee_id"},
			},
		},
		{
			Name:        "list_assignees",
			Description: "List all assignees in a project",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id": map[string]string{"type": "number"},
				},
				"required": []string{"project_id"},
			},
		},

		// User Management (5 tools)
		{
			Name:        "create_user",
			Description: "Create a new user",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"username":   map[string]string{"type": "string"},
					"email":      map[string]string{"type": "string"},
					"password":   map[string]string{"type": "string"},
					"first_name": map[string]string{"type": "string"},
					"last_name":  map[string]string{"type": "string"},
					"role":       map[string]interface{}{"type": "string", "enum": []string{"admin", "member"}},
				},
				"required": []string{"username", "email", "password"},
			},
		},
		{
			Name:        "list_users",
			Description: "List all users",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "get_user",
			Description: "Get a user by ID",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"user_id": map[string]string{"type": "number"},
				},
				"required": []string{"user_id"},
			},
		},
		{
			Name:        "update_user",
			Description: "Update an existing user",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"user_id":    map[string]string{"type": "number"},
					"username":   map[string]string{"type": "string"},
					"email":      map[string]string{"type": "string"},
					"first_name": map[string]string{"type": "string"},
					"last_name":  map[string]string{"type": "string"},
					"role":       map[string]interface{}{"type": "string", "enum": []string{"admin", "member"}},
					"is_active":  map[string]string{"type": "boolean"},
				},
				"required": []string{"user_id"},
			},
		},
		{
			Name:        "delete_user",
			Description: "Delete a user",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"user_id": map[string]string{"type": "number"},
				},
				"required": []string{"user_id"},
			},
		},

		// Comments/Notes (3 tools)
		{
			Name:        "add_comment",
			Description: "Add a comment/note to a task",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]string{"type": "number"},
					"content": map[string]string{"type": "string"},
					"author":  map[string]string{"type": "string"},
				},
				"required": []string{"task_id", "content", "author"},
			},
		},
		{
			Name:        "update_comment",
			Description: "Update an existing comment/note",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"comment_id": map[string]string{"type": "number"},
					"content":    map[string]string{"type": "string"},
				},
				"required": []string{"comment_id", "content"},
			},
		},
		{
			Name:        "list_comments",
			Description: "List all comments/notes for a task",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]string{"type": "number"},
				},
				"required": []string{"task_id"},
			},
		},

		// Task Dependencies (7 tools)
		{
			Name:        "add_task_dependency",
			Description: "Add a dependency between two tasks",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id":      map[string]string{"type": "number"},
					"depends_on_id": map[string]string{"type": "number"},
					"type":         map[string]string{"type": "string"},
				},
				"required": []string{"task_id", "depends_on_id"},
			},
		},
		{
			Name:        "remove_task_dependency",
			Description: "Remove a task dependency",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"dependency_id": map[string]string{"type": "number"},
				},
				"required": []string{"dependency_id"},
			},
		},
		{
			Name:        "list_task_dependencies",
			Description: "List dependencies for a task",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]string{"type": "number"},
				},
				"required": []string{"task_id"},
			},
		},
		{
			Name:        "get_task_dependency_chain",
			Description: "Get the full dependency chain for a task (all transitive dependencies)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]string{"type": "number"},
				},
				"required": []string{"task_id"},
			},
		},
		{
			Name:        "get_task_dependent_chain",
			Description: "Get all tasks that depend on this task",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]string{"type": "number"},
				},
				"required": []string{"task_id"},
			},
		},
		{
			Name:        "can_start_task",
			Description: "Check if a task can start based on its dependencies",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]string{"type": "number"},
				},
				"required": []string{"task_id"},
			},
		},
		{
			Name:        "get_project_dependency_graph",
			Description: "Get the full dependency graph for a project",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id": map[string]string{"type": "number"},
				},
				"required": []string{"project_id"},
			},
		},
	}
}