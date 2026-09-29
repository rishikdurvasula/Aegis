package workflow

import "time"

type WorkflowStatus string

const (
	WorkflowPending   WorkflowStatus = "PENDING"
	WorkflowRunning   WorkflowStatus = "RUNNING"
	WorkflowCompleted WorkflowStatus = "COMPLETED"
	WorkflowFailed    WorkflowStatus = "FAILED"
)

type TaskStatus string

const (
	TaskPending   TaskStatus = "PENDING"
	TaskReady     TaskStatus = "READY"
	TaskRunning   TaskStatus = "RUNNING"
	TaskCompleted TaskStatus = "COMPLETED"
	TaskFailed    TaskStatus = "FAILED"
)

type Task struct {
	ID         string
	WorkflowID string
	Name       string
	Type       string
	Payload    []byte
	Status     TaskStatus

	Dependencies []string

	CreatedAt   *time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
}

type Workflow struct {
	ID     string
	Name   string
	Status WorkflowStatus

	Tasks []Task

	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
}
