package mcp

// toolDefinitions returns all available MCP tool definitions
func toolDefinitions() []Tool {
	return []Tool{
		// Project Management Tools
		{
			Name:        "create_project",
			Description: "Create a new project",
			InputSchema: projectCreateSchema(),
		},
		{
			Name:        "list_projects",
			Description: "List all projects with filters",
			InputSchema: projectListSchema(),
		},
		{
			Name:        "update_project",
			Description: "Update project details",
			InputSchema: projectUpdateSchema(),
		},

		// Task Management Tools
		{
			Name:        "create_task",
			Description: "Create a new task",
			InputSchema: taskCreateSchema(),
		},
		{
			Name:        "update_task",
			Description: "Update task details",
			InputSchema: taskUpdateSchema(),
		},
		{
			Name:        "list_tasks",
			Description: "List tasks with filters",
			InputSchema: taskListSchema(),
		},
		{
			Name:        "add_task_dependency",
			Description: "Add dependency between tasks",
			InputSchema: taskDependencySchema(),
		},

		// Team Management Tools
		{
			Name:        "create_team",
			Description: "Create a new team",
			InputSchema: teamCreateSchema(),
		},
		{
			Name:        "add_team_member",
			Description: "Add member to team",
			InputSchema: teamMemberSchema(),
		},
		{
			Name:        "list_teams",
			Description: "List all teams",
			InputSchema: emptySchema(),
		},

		// Milestone Management Tools
		{
			Name:        "create_milestone",
			Description: "Create a milestone",
			InputSchema: milestoneCreateSchema(),
		},
		{
			Name:        "update_milestone",
			Description: "Update milestone",
			InputSchema: milestoneUpdateSchema(),
		},
		{
			Name:        "list_milestones",
			Description: "List project milestones",
			InputSchema: milestoneListSchema(),
		},

		// Sprint Management Tools
		{
			Name:        "create_sprint",
			Description: "Create a sprint",
			InputSchema: sprintCreateSchema(),
		},
		{
			Name:        "start_sprint",
			Description: "Start a sprint",
			InputSchema: idOnlySchema("Sprint ID"),
		},
		{
			Name:        "complete_sprint",
			Description: "Complete a sprint",
			InputSchema: idOnlySchema("Sprint ID"),
		},
		{
			Name:        "list_sprints",
			Description: "List project sprints",
			InputSchema: sprintListSchema(),
		},

		// Search and AI Tools
		{
			Name:        "search_tasks",
			Description: "Search tasks by keyword",
			InputSchema: searchSchema(),
		},
		{
			Name:        "semantic_search",
			Description: "Semantic search using AI across all entities",
			InputSchema: semanticSearchSchema(),
		},
		{
			Name:        "hybrid_search",
			Description: "Combined keyword and semantic search for better results",
			InputSchema: hybridSearchSchema(),
		},
		{
			Name:        "find_similar_tasks",
			Description: "Find tasks similar to a given task using AI",
			InputSchema: similarTasksSchema(),
		},
		{
			Name:        "recommend_tasks",
			Description: "Get AI-powered task recommendations for a user",
			InputSchema: recommendTasksSchema(),
		},
		{
			Name:        "cluster_project_tasks",
			Description: "Group similar tasks in a project using AI clustering",
			InputSchema: clusterTasksSchema(),
		},
		{
			Name:        "search_projects",
			Description: "Search projects using semantic and keyword search",
			InputSchema: searchProjectsSchema(),
		},
		{
			Name:        "search_documents",
			Description: "Search documents and attachments using AI",
			InputSchema: searchDocumentsSchema(),
		},
		{
			Name:        "intelligent_task_assignment",
			Description: "Get AI recommendations for task assignment based on skills and workload",
			InputSchema: intelligentAssignmentSchema(),
		},

		// Analytics Tools
		{
			Name:        "get_project_stats",
			Description: "Get project statistics",
			InputSchema: projectStatsSchema(),
		},
		{
			Name:        "get_team_velocity",
			Description: "Get team velocity metrics",
			InputSchema: teamVelocitySchema(),
		},
		{
			Name:        "get_burndown_chart",
			Description: "Get sprint burndown data",
			InputSchema: burndownSchema(),
		},

		// Collaboration Tools
		{
			Name:        "add_comment",
			Description: "Add comment to task",
			InputSchema: commentSchema(),
		},
		{
			Name:        "log_time",
			Description: "Log time entry",
			InputSchema: timeEntrySchema(),
		},
		{
			Name:        "create_notification",
			Description: "Create notification",
			InputSchema: notificationSchema(),
		},
		{
			Name:        "create_webhook",
			Description: "Create webhook",
			InputSchema: webhookSchema(),
		},
	}
}

// Schema helper functions
func emptySchema() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

func idOnlySchema(description string) map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "integer",
				"description": description,
			},
		},
		"required": []string{"id"},
	}
}

func projectCreateSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name":        stringProp("Project name"),
			"description": stringProp("Project description"),
			"owner_id":    integerProp("Owner user ID"),
			"team_id":     integerProp("Team ID"),
		},
		"required": []string{"name"},
	}
}

func projectListSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"status":  enumProp("Filter by status", []string{"active", "archived", "draft"}),
			"team_id": integerProp("Filter by team"),
		},
	}
}

func projectUpdateSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id":          integerProp("Project ID"),
			"name":        stringProp("Project name"),
			"description": stringProp("Project description"),
			"status":      enumProp("Project status", []string{"active", "archived", "draft"}),
		},
		"required": []string{"id"},
	}
}

func taskCreateSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_id":   integerProp("Project ID"),
			"title":        stringProp("Task title"),
			"description":  stringProp("Task description"),
			"priority":     enumProp("Task priority", []string{"low", "medium", "high", "urgent"}),
			"assignee_id":  integerProp("Assignee user ID"),
			"milestone_id": integerProp("Milestone ID"),
			"sprint_id":    integerProp("Sprint ID"),
			"due_date":     dateProp("Due date"),
			"labels":       arrayProp("Labels", "string"),
		},
		"required": []string{"project_id", "title"},
	}
}

func taskUpdateSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id":          integerProp("Task ID"),
			"title":       stringProp("Task title"),
			"description": stringProp("Task description"),
			"status":      enumProp("Task status", []string{"todo", "in_progress", "review", "done", "cancelled"}),
			"priority":    enumProp("Task priority", []string{"low", "medium", "high", "urgent"}),
			"assignee_id": integerProp("Assignee user ID"),
		},
		"required": []string{"id"},
	}
}

func taskListSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_id":   integerProp("Filter by project"),
			"status":       stringProp("Filter by status"),
			"assignee_id":  integerProp("Filter by assignee"),
			"milestone_id": integerProp("Filter by milestone"),
			"sprint_id":    integerProp("Filter by sprint"),
		},
	}
}

func taskDependencySchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"task_id":       integerProp("Task ID"),
			"depends_on_id": integerProp("Depends on task ID"),
			"type":          enumProp("Dependency type", []string{"blocks", "relates_to", "duplicates"}),
		},
		"required": []string{"task_id", "depends_on_id"},
	}
}

func teamCreateSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name":        stringProp("Team name"),
			"description": stringProp("Team description"),
		},
		"required": []string{"name"},
	}
}

func teamMemberSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"team_id": integerProp("Team ID"),
			"user_id": integerProp("User ID"),
			"role":    enumProp("Member role", []string{"owner", "admin", "member", "viewer"}),
		},
		"required": []string{"team_id", "user_id"},
	}
}

func milestoneCreateSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_id":  integerProp("Project ID"),
			"name":        stringProp("Milestone name"),
			"description": stringProp("Milestone description"),
			"due_date":    dateProp("Due date"),
		},
		"required": []string{"project_id", "name"},
	}
}

func milestoneUpdateSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id":          integerProp("Milestone ID"),
			"name":        stringProp("Milestone name"),
			"description": stringProp("Milestone description"),
			"status":      enumProp("Milestone status", []string{"planned", "active", "completed", "cancelled"}),
		},
		"required": []string{"id"},
	}
}

func milestoneListSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_id": integerProp("Project ID"),
		},
		"required": []string{"project_id"},
	}
}

func sprintCreateSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_id": integerProp("Project ID"),
			"name":       stringProp("Sprint name"),
			"goal":       stringProp("Sprint goal"),
			"start_date": dateProp("Start date"),
			"end_date":   dateProp("End date"),
		},
		"required": []string{"project_id", "name", "start_date", "end_date"},
	}
}

func sprintListSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_id": integerProp("Project ID"),
			"status":     enumProp("Sprint status", []string{"planned", "active", "completed"}),
		},
		"required": []string{"project_id"},
	}
}

func searchSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query":      stringProp("Search query"),
			"project_id": integerProp("Filter by project"),
		},
		"required": []string{"query"},
	}
}

func semanticSearchSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query":       stringProp("Search query"),
			"entity_type": enumProp("Entity type to search", []string{"all", "project", "task", "document"}),
			"limit":       integerPropWithDefault("Result limit", 10),
		},
		"required": []string{"query"},
	}
}

func hybridSearchSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query":       stringProp("Search query"),
			"entity_type": enumProp("Entity type", []string{"task", "project", "document"}),
			"limit":       integerPropWithDefault("Result limit", 20),
		},
		"required": []string{"query"},
	}
}

func clusterTasksSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_id":   integerProp("Project ID"),
			"num_clusters": integerPropWithDefault("Number of clusters", 5),
		},
		"required": []string{"project_id"},
	}
}

func searchProjectsSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": stringProp("Search query"),
			"limit": integerPropWithDefault("Result limit", 10),
		},
		"required": []string{"query"},
	}
}

func searchDocumentsSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query":       stringProp("Search query"),
			"project_id":  integerProp("Project ID"),
			"document_id": integerProp("Document ID"),
			"limit":       integerPropWithDefault("Result limit", 10),
		},
		"required": []string{"query"},
	}
}

func intelligentAssignmentSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"task_id": integerProp("Task ID"),
			"team_id": integerProp("Team ID"),
		},
		"required": []string{"task_id"},
	}
}

func similarTasksSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"task_id": integerProp("Task ID"),
			"limit":   integerPropWithDefault("Result limit", 5),
		},
		"required": []string{"task_id"},
	}
}

func recommendTasksSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"user_id": integerProp("User ID"),
			"limit":   integerPropWithDefault("Result limit", 10),
		},
		"required": []string{"user_id"},
	}
}

func projectStatsSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_id": integerProp("Project ID"),
		},
		"required": []string{"project_id"},
	}
}

func teamVelocitySchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"team_id": integerProp("Team ID"),
			"sprints": integerPropWithDefault("Number of past sprints", 5),
		},
		"required": []string{"team_id"},
	}
}

func burndownSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"sprint_id": integerProp("Sprint ID"),
		},
		"required": []string{"sprint_id"},
	}
}

func commentSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"task_id": integerProp("Task ID"),
			"content": stringProp("Comment content"),
			"author":  stringProp("Comment author"),
		},
		"required": []string{"task_id", "content"},
	}
}

func timeEntrySchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"task_id":     integerProp("Task ID"),
			"user_id":     integerProp("User ID"),
			"hours":       numberProp("Hours worked"),
			"description": stringProp("Description"),
			"date":        dateProp("Date"),
		},
		"required": []string{"task_id", "hours"},
	}
}

func notificationSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"user_id": integerProp("User ID"),
			"type":    stringProp("Notification type"),
			"title":   stringProp("Title"),
			"message": stringProp("Message"),
		},
		"required": []string{"user_id", "type", "title", "message"},
	}
}

func webhookSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_id": integerProp("Project ID"),
			"url":        uriProp("Webhook URL"),
			"events":     arrayProp("Events to trigger", "string"),
			"active":     booleanProp("Is active"),
		},
		"required": []string{"project_id", "url", "events"},
	}
}

// Property helpers
func stringProp(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "string",
		"description": description,
	}
}

func integerProp(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "integer",
		"description": description,
	}
}

func integerPropWithDefault(description string, defaultVal int) map[string]interface{} {
	return map[string]interface{}{
		"type":        "integer",
		"description": description,
		"default":     defaultVal,
	}
}

func numberProp(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "number",
		"description": description,
	}
}

func booleanProp(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "boolean",
		"description": description,
	}
}

func dateProp(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "string",
		"format":      "date",
		"description": description,
	}
}

func uriProp(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "string",
		"format":      "uri",
		"description": description,
	}
}

func enumProp(description string, values []string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "string",
		"description": description,
		"enum":        values,
	}
}

func arrayProp(description string, itemType string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "array",
		"description": description,
		"items": map[string]interface{}{
			"type": itemType,
		},
	}
}