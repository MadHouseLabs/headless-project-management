#!/bin/bash

# Base URL
BASE_URL="http://localhost:8080/api"

echo "Creating labels..."

# Create labels - removing auth header since it was causing issues
curl -X POST "$BASE_URL/labels" -H "Content-Type: application/json" \
  -d '{"project_id": 1, "name": "Frontend", "color": "#3b82f6"}'

curl -X POST "$BASE_URL/labels" -H "Content-Type: application/json" \
  -d '{"project_id": 1, "name": "Backend", "color": "#10b981"}'

curl -X POST "$BASE_URL/labels" -H "Content-Type: application/json" \
  -d '{"project_id": 1, "name": "Bug", "color": "#ef4444"}'

curl -X POST "$BASE_URL/labels" -H "Content-Type: application/json" \
  -d '{"project_id": 1, "name": "Feature", "color": "#8b5cf6"}'

curl -X POST "$BASE_URL/labels" -H "Content-Type: application/json" \
  -d '{"project_id": 1, "name": "Documentation", "color": "#f59e0b"}'

curl -X POST "$BASE_URL/labels" -H "Content-Type: application/json" \
  -d '{"project_id": 1, "name": "Testing", "color": "#06b6d4"}'

echo ""
echo "Labels created!"

# Now add labels to existing tasks
echo "Adding labels to tasks..."

# Task 9 - Implement user authentication (Backend, Feature)
curl -X POST "$BASE_URL/tasks/9/labels/2" 2>/dev/null
curl -X POST "$BASE_URL/tasks/9/labels/4" 2>/dev/null

# Task 10 - Design dashboard UI (Frontend, Feature)
curl -X POST "$BASE_URL/tasks/10/labels/1" 2>/dev/null
curl -X POST "$BASE_URL/tasks/10/labels/4" 2>/dev/null

# Task 11 - Fix navigation menu bug (Frontend, Bug)
curl -X POST "$BASE_URL/tasks/11/labels/1" 2>/dev/null
curl -X POST "$BASE_URL/tasks/11/labels/3" 2>/dev/null

# Task 12 - Refactor database queries (Backend)
curl -X POST "$BASE_URL/tasks/12/labels/2" 2>/dev/null

# Task 13 - Update API documentation (Documentation)
curl -X POST "$BASE_URL/tasks/13/labels/5" 2>/dev/null

# Task 14 - Implement search functionality (Backend, Feature)
curl -X POST "$BASE_URL/tasks/14/labels/2" 2>/dev/null
curl -X POST "$BASE_URL/tasks/14/labels/4" 2>/dev/null

# Task 15 - Payment integration (Backend, Feature)
curl -X POST "$BASE_URL/tasks/15/labels/2" 2>/dev/null
curl -X POST "$BASE_URL/tasks/15/labels/4" 2>/dev/null

# Task 16 - Add email notifications (Backend, Feature)
curl -X POST "$BASE_URL/tasks/16/labels/2" 2>/dev/null
curl -X POST "$BASE_URL/tasks/16/labels/4" 2>/dev/null

# Task 17 - Setup CI/CD pipeline (Testing)
curl -X POST "$BASE_URL/tasks/17/labels/6" 2>/dev/null

# Task 18 - Create user registration form (Frontend, Feature)
curl -X POST "$BASE_URL/tasks/18/labels/1" 2>/dev/null
curl -X POST "$BASE_URL/tasks/18/labels/4" 2>/dev/null

echo "Labels added to tasks!"