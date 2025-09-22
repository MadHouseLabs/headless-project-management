#!/bin/bash

# Base URL
BASE_URL="http://localhost:8080/api"
AUTH_HEADER="Authorization: Bearer test-token"

# Delete all existing tasks
echo "Deleting existing tasks..."
for i in {1..20}; do
  curl -X DELETE "$BASE_URL/tasks/$i" -H "$AUTH_HEADER" 2>/dev/null
done

# Create labels
echo "Creating labels..."
curl -X POST "$BASE_URL/labels" -H "$AUTH_HEADER" -H "Content-Type: application/json" \
  -d '{"project_id": 1, "name": "Frontend", "color": "#3b82f6"}' 2>/dev/null

curl -X POST "$BASE_URL/labels" -H "$AUTH_HEADER" -H "Content-Type: application/json" \
  -d '{"project_id": 1, "name": "Backend", "color": "#10b981"}' 2>/dev/null

curl -X POST "$BASE_URL/labels" -H "$AUTH_HEADER" -H "Content-Type: application/json" \
  -d '{"project_id": 1, "name": "Bug", "color": "#ef4444"}' 2>/dev/null

curl -X POST "$BASE_URL/labels" -H "$AUTH_HEADER" -H "Content-Type: application/json" \
  -d '{"project_id": 1, "name": "Feature", "color": "#8b5cf6"}' 2>/dev/null

curl -X POST "$BASE_URL/labels" -H "$AUTH_HEADER" -H "Content-Type: application/json" \
  -d '{"project_id": 1, "name": "Documentation", "color": "#f59e0b"}' 2>/dev/null

# Create tasks with full metadata
echo "Creating new tasks with metadata..."

# Todo tasks
curl -X POST "$BASE_URL/tasks" -H "$AUTH_HEADER" -H "Content-Type: application/json" \
  -d '{
    "project_id": 1,
    "title": "Implement user authentication",
    "description": "Add JWT-based authentication system",
    "status": "todo",
    "priority": "high",
    "assignee": "John Doe",
    "estimated_hours": 8,
    "due_date": "'$(date -v +7d '+%Y-%m-%dT%H:%M:%SZ' 2>/dev/null || date -d '+7 days' '+%Y-%m-%dT%H:%M:%SZ')'"
  }' 2>/dev/null

curl -X POST "$BASE_URL/tasks" -H "$AUTH_HEADER" -H "Content-Type: application/json" \
  -d '{
    "project_id": 1,
    "title": "Design dashboard UI",
    "description": "Create mockups for admin dashboard",
    "status": "todo",
    "priority": "medium",
    "assignee": "Jane Smith",
    "estimated_hours": 6,
    "due_date": "'$(date -v +5d '+%Y-%m-%dT%H:%M:%SZ')'"
  }' 2>/dev/null

curl -X POST "$BASE_URL/tasks" -H "$AUTH_HEADER" -H "Content-Type: application/json" \
  -d '{
    "project_id": 1,
    "title": "Fix navigation menu bug",
    "description": "Menu doesn't close on mobile devices",
    "status": "todo",
    "priority": "urgent",
    "assignee": "Bob Wilson",
    "estimated_hours": 2,
    "due_date": "'$(date -v +1d '+%Y-%m-%dT%H:%M:%SZ')'"
  }' 2>/dev/null

# In Progress tasks
curl -X POST "$BASE_URL/tasks" -H "$AUTH_HEADER" -H "Content-Type: application/json" \
  -d '{
    "project_id": 1,
    "title": "Refactor database queries",
    "description": "Optimize slow queries in reports module",
    "status": "in_progress",
    "priority": "high",
    "assignee": "Alice Brown",
    "estimated_hours": 12,
    "due_date": "'$(date -v +3d '+%Y-%m-%dT%H:%M:%SZ')'"
  }' 2>/dev/null

curl -X POST "$BASE_URL/tasks" -H "$AUTH_HEADER" -H "Content-Type: application/json" \
  -d '{
    "project_id": 1,
    "title": "Update API documentation",
    "description": "Document new endpoints and update examples",
    "status": "in_progress",
    "priority": "low",
    "assignee": "John Doe",
    "estimated_hours": 4,
    "due_date": "'$(date -v +10d '+%Y-%m-%dT%H:%M:%SZ')'"
  }' 2>/dev/null

curl -X POST "$BASE_URL/tasks" -H "$AUTH_HEADER" -H "Content-Type: application/json" \
  -d '{
    "project_id": 1,
    "title": "Implement search functionality",
    "description": "Add full-text search for products",
    "status": "in_progress",
    "priority": "medium",
    "assignee": "Jane Smith",
    "estimated_hours": 10,
    "due_date": "'$(date -v +4d '+%Y-%m-%dT%H:%M:%SZ')'"
  }' 2>/dev/null

# Review tasks
curl -X POST "$BASE_URL/tasks" -H "$AUTH_HEADER" -H "Content-Type: application/json" \
  -d '{
    "project_id": 1,
    "title": "Payment integration",
    "description": "Integrate Stripe payment gateway",
    "status": "review",
    "priority": "high",
    "assignee": "Bob Wilson",
    "estimated_hours": 16,
    "due_date": "'$(date -v +2d '+%Y-%m-%dT%H:%M:%SZ')'"
  }' 2>/dev/null

curl -X POST "$BASE_URL/tasks" -H "$AUTH_HEADER" -H "Content-Type: application/json" \
  -d '{
    "project_id": 1,
    "title": "Add email notifications",
    "description": "Send email alerts for important events",
    "status": "review",
    "priority": "medium",
    "assignee": "Alice Brown",
    "estimated_hours": 6,
    "due_date": "'$(date -v +6d '+%Y-%m-%dT%H:%M:%SZ')'"
  }' 2>/dev/null

# Done tasks
curl -X POST "$BASE_URL/tasks" -H "$AUTH_HEADER" -H "Content-Type: application/json" \
  -d '{
    "project_id": 1,
    "title": "Setup CI/CD pipeline",
    "description": "Configure GitHub Actions for deployment",
    "status": "done",
    "priority": "high",
    "assignee": "John Doe",
    "estimated_hours": 8,
    "actual_hours": 7
  }' 2>/dev/null

curl -X POST "$BASE_URL/tasks" -H "$AUTH_HEADER" -H "Content-Type: application/json" \
  -d '{
    "project_id": 1,
    "title": "Create user registration form",
    "description": "Build registration form with validation",
    "status": "done",
    "priority": "medium",
    "assignee": "Jane Smith",
    "estimated_hours": 5,
    "actual_hours": 6
  }' 2>/dev/null

echo "Tasks created successfully!"