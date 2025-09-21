package models

import (
	"time"
)

type ProjectStatus string

const (
	ProjectStatusActive   ProjectStatus = "active"
	ProjectStatusArchived ProjectStatus = "archived"
	ProjectStatusDraft    ProjectStatus = "draft"
)

type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusReview     TaskStatus = "review"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityUrgent TaskPriority = "urgent"
)

type SprintStatus string

const (
	SprintStatusPlanned   SprintStatus = "planned"
	SprintStatusActive    SprintStatus = "active"
	SprintStatusCompleted SprintStatus = "completed"
)

type MilestoneStatus string

const (
	MilestoneStatusPlanned   MilestoneStatus = "planned"
	MilestoneStatusActive    MilestoneStatus = "active"
	MilestoneStatusCompleted MilestoneStatus = "completed"
	MilestoneStatusCancelled MilestoneStatus = "cancelled"
)

type WorkflowStateType string

const (
	WorkflowStateTypeStart      WorkflowStateType = "start"
	WorkflowStateTypeInProgress WorkflowStateType = "in_progress"
	WorkflowStateTypeDone       WorkflowStateType = "done"
)

type Project struct {
	ID          uint          `json:"id" gorm:"primaryKey"`
	Name        string        `json:"name" gorm:"not null"`
	Description string        `json:"description"`
	Status      ProjectStatus `json:"status" gorm:"default:'active'"`
	OwnerID     uint          `json:"owner_id"`
	TeamID      *uint         `json:"team_id"`
	StartDate   *time.Time    `json:"start_date"`
	EndDate     *time.Time    `json:"end_date"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	DeletedAt   *time.Time    `json:"deleted_at" gorm:"index"`
	Tasks       []Task        `json:"tasks,omitempty" gorm:"foreignKey:ProjectID"`
	Owner       *User         `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
	Team        *Team         `json:"team,omitempty" gorm:"foreignKey:TeamID"`
	Members     []User        `json:"members,omitempty" gorm:"many2many:project_members;"`
}

type Task struct {
	ID              uint         `json:"id" gorm:"primaryKey"`
	ProjectID       uint         `json:"project_id" gorm:"not null"`
	ParentID        *uint        `json:"parent_id,omitempty"`
	MilestoneID     *uint        `json:"milestone_id,omitempty"`
	SprintID        *uint        `json:"sprint_id,omitempty"`
	Title           string       `json:"title" gorm:"not null"`
	Description     string       `json:"description"`
	Status          TaskStatus   `json:"status" gorm:"default:'todo'"`
	Priority        TaskPriority `json:"priority" gorm:"default:'medium'"`
	Assignee        string       `json:"assignee"`
	AssigneeID      *uint        `json:"assignee_id"`
	EstimatedHours  *float64     `json:"estimated_hours"`
	ActualHours     *float64     `json:"actual_hours"`
	StoryPoints     *int         `json:"story_points"`
	DueDate         *time.Time   `json:"due_date,omitempty"`
	StartDate       *time.Time   `json:"start_date,omitempty"`
	CompletedAt     *time.Time   `json:"completed_at,omitempty"`
	CreatedBy       uint         `json:"created_by"`
	UpdatedBy       *uint        `json:"updated_by"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
	DeletedAt       *time.Time   `json:"deleted_at" gorm:"index"`
	Project         *Project     `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Parent          *Task        `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Milestone       *Milestone   `json:"milestone,omitempty" gorm:"foreignKey:MilestoneID"`
	Sprint          *Sprint      `json:"sprint,omitempty" gorm:"foreignKey:SprintID"`
	AssigneeUser    *User        `json:"assignee_user,omitempty" gorm:"foreignKey:AssigneeID"`
	Creator         *User        `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
	Updater         *User        `json:"updater,omitempty" gorm:"foreignKey:UpdatedBy"`
	Subtasks        []Task       `json:"subtasks,omitempty" gorm:"foreignKey:ParentID"`
	Comments        []Comment    `json:"comments,omitempty" gorm:"foreignKey:TaskID"`
	Attachments     []Attachment `json:"attachments,omitempty" gorm:"foreignKey:TaskID"`
	Labels          []Label      `json:"labels,omitempty" gorm:"many2many:task_labels;"`
	Watchers        []User       `json:"watchers,omitempty" gorm:"many2many:task_watchers;"`
	Dependencies    []TaskDependency `json:"dependencies,omitempty" gorm:"foreignKey:TaskID"`
}

type Comment struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	TaskID    uint      `json:"task_id" gorm:"not null"`
	Content   string    `json:"content" gorm:"not null"`
	Author    string    `json:"author" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	Task      *Task     `json:"task,omitempty" gorm:"foreignKey:TaskID"`
}

type Attachment struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	TaskID    uint      `json:"task_id" gorm:"not null"`
	Filename  string    `json:"filename" gorm:"not null"`
	Path      string    `json:"path" gorm:"not null"`
	Size      int64     `json:"size"`
	MimeType  string    `json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`
	Task      *Task     `json:"task,omitempty" gorm:"foreignKey:TaskID"`
}

type Label struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	ProjectID uint   `json:"project_id"`
	Name      string `json:"name" gorm:"not null"`
	Color     string `json:"color"`
	Tasks     []Task `json:"tasks,omitempty" gorm:"many2many:task_labels;"`
}