package models

import (
	"time"
)

type Milestone struct {
	ID          uint            `json:"id" gorm:"primaryKey"`
	ProjectID   uint            `json:"project_id" gorm:"not null"`
	Name        string          `json:"name" gorm:"not null"`
	Description string          `json:"description"`
	DueDate     time.Time       `json:"due_date"`
	Status      MilestoneStatus `json:"status" gorm:"default:'planned'"`
	Progress    int             `json:"progress" gorm:"default:0"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	Project     *Project        `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Tasks       []Task          `json:"tasks,omitempty" gorm:"foreignKey:MilestoneID"`
}

type Sprint struct {
	ID        uint         `json:"id" gorm:"primaryKey"`
	ProjectID uint         `json:"project_id" gorm:"not null"`
	Name      string       `json:"name" gorm:"not null"`
	Goal      string       `json:"goal"`
	StartDate time.Time    `json:"start_date"`
	EndDate   time.Time    `json:"end_date"`
	Status    SprintStatus `json:"status" gorm:"default:'planned'"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	Project   *Project     `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Tasks     []Task       `json:"tasks,omitempty" gorm:"foreignKey:SprintID"`
}

type Workflow struct {
	ID          uint             `json:"id" gorm:"primaryKey"`
	ProjectID   uint             `json:"project_id" gorm:"not null"`
	Name        string           `json:"name" gorm:"not null"`
	Description string           `json:"description"`
	IsDefault   bool             `json:"is_default" gorm:"default:false"`
	IsActive    bool             `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	Project     *Project         `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	States      []WorkflowState  `json:"states,omitempty" gorm:"foreignKey:WorkflowID"`
}

type WorkflowState struct {
	ID         uint              `json:"id" gorm:"primaryKey"`
	WorkflowID uint              `json:"workflow_id" gorm:"not null"`
	Name       string            `json:"name" gorm:"not null"`
	Type       WorkflowStateType `json:"type"`
	Order      int               `json:"order"`
	Color      string            `json:"color"`
	IsInitial  bool              `json:"is_initial" gorm:"default:false"`
	IsFinal    bool              `json:"is_final" gorm:"default:false"`
	Workflow   *Workflow         `json:"workflow,omitempty" gorm:"foreignKey:WorkflowID"`
}

type CustomField struct {
	ID          uint        `json:"id" gorm:"primaryKey"`
	ProjectID   uint        `json:"project_id" gorm:"not null"`
	Name        string      `json:"name" gorm:"not null"`
	FieldType   string      `json:"field_type"`
	Options     string      `json:"options"`
	IsRequired  bool        `json:"is_required" gorm:"default:false"`
	Order       int         `json:"order"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	Project     *Project    `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Values      []FieldValue `json:"values,omitempty" gorm:"foreignKey:FieldID"`
}

type FieldValue struct {
	ID        uint         `json:"id" gorm:"primaryKey"`
	FieldID   uint         `json:"field_id" gorm:"not null"`
	TaskID    uint         `json:"task_id" gorm:"not null"`
	Value     string       `json:"value"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	Field     *CustomField `json:"field,omitempty" gorm:"foreignKey:FieldID"`
	Task      *Task        `json:"task,omitempty" gorm:"foreignKey:TaskID"`
}

type TaskDependency struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	TaskID          uint      `json:"task_id" gorm:"not null"`
	DependsOnTaskID uint      `json:"depends_on_task_id" gorm:"not null"`
	Type            string    `json:"type" gorm:"default:'finish_to_start'"`
	CreatedAt       time.Time `json:"created_at"`
	Task            *Task     `json:"task,omitempty" gorm:"foreignKey:TaskID"`
	DependsOnTask   *Task     `json:"depends_on_task,omitempty" gorm:"foreignKey:DependsOnTaskID"`
}

type TimeEntry struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	TaskID      uint      `json:"task_id" gorm:"not null"`
	UserID      uint      `json:"user_id" gorm:"not null"`
	Hours       float64   `json:"hours"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Task        *Task     `json:"task,omitempty" gorm:"foreignKey:TaskID"`
	User        *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type Activity struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"not null"`
	ProjectID   *uint     `json:"project_id"`
	TaskID      *uint     `json:"task_id"`
	Action      string    `json:"action" gorm:"not null"`
	EntityType  string    `json:"entity_type" gorm:"not null"`
	EntityID    uint      `json:"entity_id" gorm:"not null"`
	Details     string    `json:"details"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
	CreatedAt   time.Time `json:"created_at"`
	User        *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Project     *Project  `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Task        *Task     `json:"task,omitempty" gorm:"foreignKey:TaskID"`
}

type Notification struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	Type      string    `json:"type" gorm:"not null"`
	Title     string    `json:"title" gorm:"not null"`
	Message   string    `json:"message"`
	Data      string    `json:"data"`
	IsRead    bool      `json:"is_read" gorm:"default:false"`
	ReadAt    *time.Time `json:"read_at"`
	CreatedAt time.Time `json:"created_at"`
	User      *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type Webhook struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ProjectID uint      `json:"project_id" gorm:"not null"`
	URL       string    `json:"url" gorm:"not null"`
	Secret    string    `json:"secret"`
	Events    string    `json:"events"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Project   *Project  `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
}