package runner

import (
	"github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/model"
)

// Type aliases to help with import resolution
// This helps avoid circular dependencies and import issues
type AgentType = *agent.Agent
type ModelRequestType = model.ModelRequest 
type ModelSettingsType = model.ModelSettings 