package runner

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
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

// Interaction represents a single interaction in a task's history
type Interaction struct {
	// Role is the role of the interaction (e.g., "user", "agent")
	Role string

	// Content is the content of the interaction
	Content interface{}

	// Timestamp is the time the interaction occurred
	Timestamp time.Time
}

// WorkingContext contains context information about the task
type WorkingContext struct {
	// Artifact is the primary artifact being worked on (e.g., code, text)
	Artifact interface{} `json:"artifact,omitempty"`

	// ArtifactType indicates the type of artifact (e.g., "code", "text")
	ArtifactType string `json:"artifact_type,omitempty"`

	// Metadata contains additional metadata about the task
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

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

	// RelatedTaskIDs contains IDs of related tasks (e.g., parent, children)
	RelatedTaskIDs []string

	// TaskDescription contains a human-readable description of the task
	TaskDescription string

	// WorkingContext contains context information about what is being worked on
	WorkingContext *WorkingContext

	// InteractionHistory contains the history of interactions for this task
	InteractionHistory []Interaction
}

// NewTaskContext creates a new task context
func NewTaskContext(taskID, parentName, childName string) *TaskContext {
	return &TaskContext{
		TaskID:             taskID,
		ParentAgentName:    parentName,
		ChildAgentName:     childName,
		Status:             TaskStatusPending,
		CreatedAt:          time.Now(),
		RelatedTaskIDs:     []string{},
		WorkingContext:     &WorkingContext{Metadata: make(map[string]interface{})},
		InteractionHistory: []Interaction{},
	}
}

// Complete marks the task as complete with a result
func (t *TaskContext) Complete(result interface{}) {
	t.Status = TaskStatusComplete
	t.Result = result
	now := time.Now()
	t.CompletedAt = &now
}

// Fail marks the task as failed with an error
func (t *TaskContext) Fail(err error) {
	t.Status = TaskStatusFailed
	t.Result = err
	now := time.Now()
	t.CompletedAt = &now
}

// IsPending checks if the task is still pending
func (t *TaskContext) IsPending() bool {
	return t.Status == TaskStatusPending
}

// IsComplete checks if the task is complete
func (t *TaskContext) IsComplete() bool {
	return t.Status == TaskStatusComplete
}

// IsFailed checks if the task has failed
func (t *TaskContext) IsFailed() bool {
	return t.Status == TaskStatusFailed
}

// IsFinished checks if the task is either complete or failed
func (t *TaskContext) IsFinished() bool {
	return t.IsComplete() || t.IsFailed()
}

// GetResult returns the task result
func (t *TaskContext) GetResult() interface{} {
	return t.Result
}

// GetDelegationChain returns the delegation chain in the format "parent -> child"
func (t *TaskContext) GetDelegationChain() string {
	return fmt.Sprintf("%s -> %s", t.ParentAgentName, t.ChildAgentName)
}

// SetDescription sets a human-readable description for the task
func (t *TaskContext) SetDescription(description string) {
	t.TaskDescription = description
}

// AddRelatedTask adds a related task ID
func (t *TaskContext) AddRelatedTask(taskID string) {
	t.RelatedTaskIDs = append(t.RelatedTaskIDs, taskID)
}

// SetArtifact sets the working artifact for the task
func (t *TaskContext) SetArtifact(artifact interface{}, artifactType string) {
	t.WorkingContext.Artifact = artifact
	t.WorkingContext.ArtifactType = artifactType
}

// GetArtifact returns the working artifact for the task
func (t *TaskContext) GetArtifact() interface{} {
	return t.WorkingContext.Artifact
}

// AddMetadata adds metadata to the task context
func (t *TaskContext) AddMetadata(key string, value interface{}) {
	t.WorkingContext.Metadata[key] = value
}

// GetMetadata retrieves metadata from the task context
func (t *TaskContext) GetMetadata(key string) interface{} {
	return t.WorkingContext.Metadata[key]
}

// AddInteraction adds an interaction to the task history
func (t *TaskContext) AddInteraction(role string, content interface{}) {
	t.InteractionHistory = append(t.InteractionHistory, Interaction{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
}

// GetInteractionHistory returns the interaction history for the task
func (t *TaskContext) GetInteractionHistory() []Interaction {
	return t.InteractionHistory
}

// GetLastInteraction returns the most recent interaction
func (t *TaskContext) GetLastInteraction() *Interaction {
	if len(t.InteractionHistory) == 0 {
		return nil
	}
	return &t.InteractionHistory[len(t.InteractionHistory)-1]
}

// ToJSON converts the task context to JSON
func (t *TaskContext) ToJSON() (string, error) {
	bytes, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// GenerateTaskID generates a unique task ID
func GenerateTaskID() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		// Fall back to a timestamp-based ID if crypto/rand fails
		return fmt.Sprintf("task-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("task-%x", b)
}
