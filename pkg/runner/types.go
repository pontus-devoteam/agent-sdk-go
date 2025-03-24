package runner
import (
	"time"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/model"
)
// Type aliases to help with import resolution
type (
	AgentType = *agent.Agent
	ModelRequestType = model.Request
	ModelSettingsType = model.Settings
)
// WorkflowState represents the current state of a workflow
type WorkflowState struct {
	// CurrentPhase is the current phase of the workflow
	CurrentPhase string
	// CompletedPhases are the phases that have been completed
	CompletedPhases []string
	// Artifacts are data produced during workflow execution
	Artifacts map[string]interface{}
	// LastCheckpoint is when the state was last saved
	LastCheckpoint time.Time
	// Metadata is additional workflow metadata
	Metadata map[string]interface{}
} 