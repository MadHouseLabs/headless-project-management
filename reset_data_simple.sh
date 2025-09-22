#!/bin/bash

# Base URL
BASE_URL="http://localhost:8080/api"

# Delete all existing tasks
echo "Deleting existing tasks..."
for i in {1..20}; do
  curl -X DELETE "$BASE_URL/tasks/$i" 2>/dev/null
done

# Create tasks with full metadata
echo "Creating new tasks with metadata..."

# Todo tasks
curl -X POST "$BASE_URL/tasks" -H "Content-Type: application/json" \
  -d '{
    "project_id": 1,
    "title": "Implement user authentication",
    "description": "Add JWT-based authentication system",
    "status": "todo",
    "priority": "high",
    "assignee": "John Doe",
    "estimated_hours": 8,
    "due_date": "2025-09-29T10:00:00Z"
  }'

curl -X POST "$BASE_URL/tasks" -H "Content-Type: application/json" \
  -d '{
    "project_id": 1,
    "title": "Design dashboard UI",
    "description": "Create mockups for admin dashboard",
    "status": "todo",
    "priority": "medium",
    "assignee": "Jane Smith",
    "estimated_hours": 6,
    "due_date": "2025-09-27T10:00:00Z"
  }'

curl -X POST "$BASE_URL/tasks" -H "Content-Type: application/json" \
  -d '{
    "project_id": 1,
    "title": "Fix navigation menu bug",
    "description": "Menu does not close on mobile devices",
    "status": "todo",
    "priority": "urgent",
    "assignee": "Bob Wilson",
    "estimated_hours": 2,
    "due_date": "2025-09-23T10:00:00Z"
  }'

# In Progress tasks
curl -X POST "$BASE_URL/tasks" -H "Content-Type: application/json" \
  -d '{
    "project_id": 1,
    "title": "Refactor database queries",
    "description": "Optimize slow queries in reports module",
    "status": "in_progress",
    "priority": "high",
    "assignee": "Alice Brown",
    "estimated_hours": 12,
    "due_date": "2025-09-25T10:00:00Z"
  }'

curl -X POST "$BASE_URL/tasks" -H "Content-Type: application/json" \
  -d '{
    "project_id": 1,
    "title": "Update API documentation",
    "description": "Document new endpoints and update examples",
    "status": "in_progress",
    "priority": "low",
    "assignee": "John Doe",
    "estimated_hours": 4,
    "due_date": "2025-10-02T10:00:00Z"
  }'

curl -X POST "$BASE_URL/tasks" -H "Content-Type: application/json" \
  -d '{
    "project_id": 1,
    "title": "Implement search functionality",
    "description": "Add full-text search for products",
    "status": "in_progress",
    "priority": "medium",
    "assignee": "Jane Smith",
    "estimated_hours": 10,
    "due_date": "2025-09-26T10:00:00Z"
  }'

# Review tasks
curl -X POST "$BASE_URL/tasks" -H "Content-Type: application/json" \
  -d '{
    "project_id": 1,
    "title": "Payment integration",
    "description": "Integrate Stripe payment gateway",
    "status": "review",
    "priority": "high",
    "assignee": "Bob Wilson",
    "estimated_hours": 16,
    "due_date": "2025-09-24T10:00:00Z"
  }'

curl -X POST "$BASE_URL/tasks" -H "Content-Type: application/json" \
  -d '{
    "project_id": 1,
    "title": "Add email notifications",
    "description": "Send email alerts for important events",
    "status": "review",
    "priority": "medium",
    "assignee": "Alice Brown",
    "estimated_hours": 6,
    "due_date": "2025-09-28T10:00:00Z"
  }'

# Done tasks
curl -X POST "$BASE_URL/tasks" -H "Content-Type: application/json" \
  -d '{
    "project_id": 1,
    "title": "Setup CI/CD pipeline",
    "description": "Configure GitHub Actions for deployment",
    "status": "done",
    "priority": "high",
    "assignee": "John Doe",
    "estimated_hours": 8,
    "actual_hours": 7,
    "completed_at": "2025-09-20T14:00:00Z"
  }'

curl -X POST "$BASE_URL/tasks" -H "Content-Type: application/json" \
  -d '{
    "project_id": 1,
    "title": "Create user registration form",
    "description": "Build registration form with validation",
    "status": "done",
    "priority": "medium",
    "assignee": "Jane Smith",
    "estimated_hours": 5,
    "actual_hours": 6,
    "completed_at": "2025-09-19T16:00:00Z"
  }'

echo ""
echo "Tasks created successfully!"