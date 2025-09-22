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
			Description: "Delete an epic",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"epic_id": map[string]string{"type": "number"},
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

		// Label Management (3 tools)
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
	}
}