package main

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
	Hooks interface{}

	// RunConfig is global configuration
	RunConfig *RunConfig
}

// RunConfig configures global settings
type RunConfig struct {
	// Model is a model override (string or Model)
	Model interface{}

	// ModelProvider is the provider for resolving model names
	ModelProvider model.Provider

	// ModelSettings are global model settings
	ModelSettings *model.Settings

	// HandoffInputFilter is a global handoff input filter
	HandoffInputFilter interface{}

	// InputGuardrails are global input guardrails
	InputGuardrails []interface{}

	// OutputGuardrails are global output guardrails
	OutputGuardrails []interface{}

	// TracingDisabled indicates whether tracing is disabled
	TracingDisabled bool

	// TracingConfig is tracing configuration
	TracingConfig *TracingConfig
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
