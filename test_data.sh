#!/bin/bash

# Base URL and auth token
BASE_URL="http://localhost:8080/api"
ADMIN_TOKEN="test-admin-token-123"

# Create a project
echo "Creating project..."
PROJECT_RESPONSE=$(curl -s -X POST "$BASE_URL/projects" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "E-Commerce Platform",
    "description": "Next-gen e-commerce platform with microservices architecture",
    "status": "active"
  }')

PROJECT_ID=$(echo $PROJECT_RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
echo "Created project with ID: $PROJECT_ID"

# Create labels
echo "Creating labels..."
LABEL1_RESPONSE=$(curl -s -X POST "$BASE_URL/labels" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"project_id\": $PROJECT_ID,
    \"name\": \"Backend\",
    \"color\": \"#06c\"
  }")

LABEL2_RESPONSE=$(curl -s -X POST "$BASE_URL/labels" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"project_id\": $PROJECT_ID,
    \"name\": \"Frontend\",
    \"color\": \"#16a34a\"
  }")

LABEL3_RESPONSE=$(curl -s -X POST "$BASE_URL/labels" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"project_id\": $PROJECT_ID,
    \"name\": \"Bug\",
    \"color\": \"#dc2626\"
  }")

LABEL4_RESPONSE=$(curl -s -X POST "$BASE_URL/labels" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"project_id\": $PROJECT_ID,
    \"name\": \"Feature\",
    \"color\": \"#ca8a04\"
  }")

LABEL1_ID=$(echo $LABEL1_RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
LABEL2_ID=$(echo $LABEL2_RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
LABEL3_ID=$(echo $LABEL3_RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
LABEL4_ID=$(echo $LABEL4_RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*')

echo "Created labels: Backend($LABEL1_ID), Frontend($LABEL2_ID), Bug($LABEL3_ID), Feature($LABEL4_ID)"

# Create tasks with different assignees and statuses
echo "Creating tasks..."

# Task 1 - John's task
TASK1_RESPONSE=$(curl -s -X POST "$BASE_URL/tasks" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"project_id\": $PROJECT_ID,
    \"title\": \"Setup API Gateway\",
    \"description\": \"Configure Kong or AWS API Gateway for microservices\",
    \"status\": \"in_progress\",
    \"priority\": \"high\",
    \"assignee\": \"John Smith\"
  }")
TASK1_ID=$(echo $TASK1_RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*')

# Task 2 - Sarah's task
TASK2_RESPONSE=$(curl -s -X POST "$BASE_URL/tasks" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"project_id\": $PROJECT_ID,
    \"title\": \"Implement React Dashboard\",
    \"description\": \"Create admin dashboard using React and Material-UI\",
    \"status\": \"todo\",
    \"priority\": \"medium\",
    \"assignee\": \"Sarah Johnson\"
  }")
TASK2_ID=$(echo $TASK2_RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*')

# Task 3 - Mike's task
TASK3_RESPONSE=$(curl -s -X POST "$BASE_URL/tasks" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"project_id\": $PROJECT_ID,
    \"title\": \"Fix Payment Processing Bug\",
    \"description\": \"Stripe webhook not processing correctly\",
    \"status\": \"review\",
    \"priority\": \"urgent\",
    \"assignee\": \"Mike Chen\"
  }")
TASK3_ID=$(echo $TASK3_RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*')

# Task 4 - John's second task
TASK4_RESPONSE=$(curl -s -X POST "$BASE_URL/tasks" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"project_id\": $PROJECT_ID,
    \"title\": \"Database Schema Design\",
    \"description\": \"Design PostgreSQL schema for product catalog\",
    \"status\": \"done\",
    \"priority\": \"high\",
    \"assignee\": \"John Smith\"
  }")
TASK4_ID=$(echo $TASK4_RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*')

# Task 5 - Sarah's second task
TASK5_RESPONSE=$(curl -s -X POST "$BASE_URL/tasks" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"project_id\": $PROJECT_ID,
    \"title\": \"Mobile App UI Design\",
    \"description\": \"Design mobile app screens in Figma\",
    \"status\": \"in_progress\",
    \"priority\": \"medium\",
    \"assignee\": \"Sarah Johnson\"
  }")
TASK5_ID=$(echo $TASK5_RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*')

# Task 6 - Unassigned task
TASK6_RESPONSE=$(curl -s -X POST "$BASE_URL/tasks" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"project_id\": $PROJECT_ID,
    \"title\": \"Performance Testing\",
    \"description\": \"Load testing with K6 or JMeter\",
    \"status\": \"todo\",
    \"priority\": \"low\"
  }")
TASK6_ID=$(echo $TASK6_RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*')

echo "Created 6 tasks"

# Assign labels to tasks
echo "Assigning labels to tasks..."

# Task 1 - Backend, Feature
curl -s -X POST "$BASE_URL/labels/assign" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"task_id\": $TASK1_ID, \"label_id\": $LABEL1_ID}"

curl -s -X POST "$BASE_URL/labels/assign" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"task_id\": $TASK1_ID, \"label_id\": $LABEL4_ID}"

# Task 2 - Frontend, Feature
curl -s -X POST "$BASE_URL/labels/assign" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"task_id\": $TASK2_ID, \"label_id\": $LABEL2_ID}"

curl -s -X POST "$BASE_URL/labels/assign" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"task_id\": $TASK2_ID, \"label_id\": $LABEL4_ID}"

# Task 3 - Backend, Bug
curl -s -X POST "$BASE_URL/labels/assign" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"task_id\": $TASK3_ID, \"label_id\": $LABEL1_ID}"

curl -s -X POST "$BASE_URL/labels/assign" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"task_id\": $TASK3_ID, \"label_id\": $LABEL3_ID}"

# Task 4 - Backend
curl -s -X POST "$BASE_URL/labels/assign" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"task_id\": $TASK4_ID, \"label_id\": $LABEL1_ID}"

# Task 5 - Frontend
curl -s -X POST "$BASE_URL/labels/assign" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"task_id\": $TASK5_ID, \"label_id\": $LABEL2_ID}"

echo "Labels assigned to tasks"

# Create an epic
echo "Creating epic..."
EPIC_RESPONSE=$(curl -s -X POST "$BASE_URL/epics" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"project_id\": $PROJECT_ID,
    \"name\": \"User Authentication System\",
    \"description\": \"Complete authentication flow with OAuth2 support\",
    \"status\": \"active\"
  }")

EPIC_ID=$(echo $EPIC_RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
echo "Created epic with ID: $EPIC_ID"

# Assign some tasks to the epic
curl -s -X POST "$BASE_URL/epics/$EPIC_ID/tasks" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"task_id\": $TASK1_ID}"

curl -s -X POST "$BASE_URL/epics/$EPIC_ID/tasks" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"task_id\": $TASK4_ID}"

echo "Tasks assigned to epic"

echo ""
echo "Sample data created successfully!"
echo "Visit http://localhost:8080/projects/$PROJECT_ID/tasks to see the tasks with filters"
echo ""
echo "You can filter by:"
echo "  - Assignees: John Smith, Sarah Johnson, Mike Chen"
echo "  - Labels: Backend, Frontend, Bug, Feature"