package result

import (
	"time"
)

// TaskStatus represents the status of a delegated task
type TaskStatus string

const (
	// TaskStatusPending indicates the task is in progress
	TaskStatusPending TaskStatus = "pending"

	// TaskStatusComplete indicates the task is complete
	TaskStatusComplete TaskStatus = "complete"

	// TaskStatusFailed indicates the task failed
	TaskStatusFailed TaskStatus = "failed"
)

// TaskContext tracks information about a delegated task
type TaskContext struct {
	// TaskID is a unique identifier for the task
	TaskID string

	// ParentAgentName is the name of the agent that delegated the task
	ParentAgentName string

	// ChildAgentName is the name of the agent executing the task
	ChildAgentName string

	// Status indicates the current status of the task
	Status TaskStatus

	// CreatedAt is the time the task was created
	CreatedAt time.Time

	// CompletedAt is the time the task was completed
	CompletedAt *time.Time

	// Result contains the result of the task
	Result interface{}
}
