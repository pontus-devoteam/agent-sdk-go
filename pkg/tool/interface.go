package tool

import (
	"context"
)

// Tool represents a capability that can be used by an agent
type Tool interface {
	// GetName returns the name of the tool
	GetName() string

	// GetDescription returns the description of the tool
	GetDescription() string

	// GetParametersSchema returns the JSON schema for the tool parameters
	GetParametersSchema() map[string]interface{}

	// Execute executes the tool with the given parameters
	Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
}
