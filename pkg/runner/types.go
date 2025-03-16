package runner

import (
	"github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/model"
)

// Type aliases to help with import resolution
// This helps avoid circular dependencies and import issues

// AgentType represents a pointer to an Agent.
// This alias helps avoid circular dependencies.
type AgentType = *agent.Agent

// ModelRequestType represents a request to be sent to a model.
// This alias helps avoid circular dependencies.
type ModelRequestType = model.Request

// ModelSettingsType represents configuration settings for a model.
// This alias helps avoid circular dependencies.
type ModelSettingsType = model.Settings
