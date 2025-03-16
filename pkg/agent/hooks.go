package agent

import (
	"context"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/model"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/tool"
)

// AgentHooks defines lifecycle hooks for an agent
type AgentHooks interface {
	// OnAgentStart is called when the agent starts processing
	OnAgentStart(ctx context.Context, agent *Agent, input interface{}) error

	// OnBeforeModelCall is called before the model is called
	OnBeforeModelCall(ctx context.Context, agent *Agent, request *model.ModelRequest) error

	// OnAfterModelCall is called after the model is called
	OnAfterModelCall(ctx context.Context, agent *Agent, response *model.ModelResponse) error

	// OnBeforeToolCall is called before a tool is called
	OnBeforeToolCall(ctx context.Context, agent *Agent, tool tool.Tool, params map[string]interface{}) error

	// OnAfterToolCall is called after a tool is called
	OnAfterToolCall(ctx context.Context, agent *Agent, tool tool.Tool, result interface{}, err error) error

	// OnBeforeHandoff is called before a handoff to another agent
	OnBeforeHandoff(ctx context.Context, agent *Agent, handoffAgent *Agent) error

	// OnAfterHandoff is called after a handoff to another agent
	OnAfterHandoff(ctx context.Context, agent *Agent, handoffAgent *Agent, result interface{}) error

	// OnAgentEnd is called when the agent finishes processing
	OnAgentEnd(ctx context.Context, agent *Agent, result interface{}) error
}

// DefaultAgentHooks provides a default implementation of AgentHooks
type DefaultAgentHooks struct{}

// OnAgentStart is called when the agent starts processing
func (h *DefaultAgentHooks) OnAgentStart(ctx context.Context, agent *Agent, input interface{}) error {
	return nil
}

// OnBeforeModelCall is called before the model is called
func (h *DefaultAgentHooks) OnBeforeModelCall(ctx context.Context, agent *Agent, request *model.ModelRequest) error {
	return nil
}

// OnAfterModelCall is called after the model is called
func (h *DefaultAgentHooks) OnAfterModelCall(ctx context.Context, agent *Agent, response *model.ModelResponse) error {
	return nil
}

// OnBeforeToolCall is called before a tool is called
func (h *DefaultAgentHooks) OnBeforeToolCall(ctx context.Context, agent *Agent, tool tool.Tool, params map[string]interface{}) error {
	return nil
}

// OnAfterToolCall is called after a tool is called
func (h *DefaultAgentHooks) OnAfterToolCall(ctx context.Context, agent *Agent, tool tool.Tool, result interface{}, err error) error {
	return nil
}

// OnBeforeHandoff is called before a handoff to another agent
func (h *DefaultAgentHooks) OnBeforeHandoff(ctx context.Context, agent *Agent, handoffAgent *Agent) error {
	return nil
}

// OnAfterHandoff is called after a handoff to another agent
func (h *DefaultAgentHooks) OnAfterHandoff(ctx context.Context, agent *Agent, handoffAgent *Agent, result interface{}) error {
	return nil
}

// OnAgentEnd is called when the agent finishes processing
func (h *DefaultAgentHooks) OnAgentEnd(ctx context.Context, agent *Agent, result interface{}) error {
	return nil
}
