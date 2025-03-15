package runner

import (
	"github.com/pontus-devoteam/agent-sdk-go/pkg/model"
)

// RunOptions configures a run
type RunOptions struct {
	// Input is the input to the run
	Input interface{}

	// Context is a user-provided context object
	Context interface{}

	// MaxTurns is the maximum number of turns
	MaxTurns int

	// Hooks are lifecycle hooks for the run
	Hooks RunHooks

	// RunConfig is global configuration
	RunConfig *RunConfig
}

// RunConfig configures global settings
type RunConfig struct {
	// Model is a model override (string or Model)
	Model interface{}

	// ModelProvider is the provider for resolving model names
	ModelProvider model.ModelProvider

	// ModelSettings are global model settings
	ModelSettings *model.ModelSettings

	// HandoffInputFilter is a global handoff input filter
	HandoffInputFilter HandoffInputFilter

	// InputGuardrails are global input guardrails
	InputGuardrails []InputGuardrail

	// OutputGuardrails are global output guardrails
	OutputGuardrails []OutputGuardrail

	// TracingDisabled indicates whether tracing is disabled
	TracingDisabled bool

	// TracingConfig is tracing configuration
	TracingConfig *TracingConfig
}

// HandoffInputFilter is a function that filters input during handoffs
type HandoffInputFilter func(input interface{}) (interface{}, error)

// InputGuardrail is an interface for input guardrails
type InputGuardrail interface {
	// Check checks the input
	Check(input interface{}) (bool, string, error)
}

// OutputGuardrail is an interface for output guardrails
type OutputGuardrail interface {
	// Check checks the output
	Check(output interface{}) (bool, string, error)
}

// TracingConfig configures tracing
type TracingConfig struct {
	// WorkflowName is the name of the workflow
	WorkflowName string

	// TraceID is a custom trace ID
	TraceID string

	// GroupID is a grouping identifier
	GroupID string

	// Metadata is additional metadata
	Metadata map[string]interface{}
} 