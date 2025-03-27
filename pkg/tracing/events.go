package tracing

import (
	"context"
	"time"
)

// AgentStart records an agent start event
func AgentStart(ctx context.Context, agentName string, input interface{}) {
	RecordEventContext(ctx, Event{
		Type:      EventTypeAgentStart,
		AgentName: agentName,
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"input": input,
		},
	})
}

// AgentEnd records an agent end event
func AgentEnd(ctx context.Context, agentName string, output interface{}) {
	RecordEventContext(ctx, Event{
		Type:      EventTypeAgentEnd,
		AgentName: agentName,
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"output": output,
		},
	})
}

// ToolCall records a tool call event
func ToolCall(ctx context.Context, agentName string, toolName string, parameters interface{}) {
	RecordEventContext(ctx, Event{
		Type:      EventTypeToolCall,
		AgentName: agentName,
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"tool_name":  toolName,
			"parameters": parameters,
		},
	})
}

// ToolResult records a tool result event
func ToolResult(ctx context.Context, agentName string, toolName string, result interface{}, err error) {
	details := map[string]interface{}{
		"tool_name": toolName,
		"result":    result,
	}

	event := Event{
		Type:      EventTypeToolResult,
		AgentName: agentName,
		Timestamp: time.Now(),
		Details:   details,
	}

	if err != nil {
		event.Error = err
	}

	RecordEventContext(ctx, event)
}

// ModelRequest records a model request event
func ModelRequest(ctx context.Context, agentName string, model string, prompt interface{}, tools []interface{}) {
	RecordEventContext(ctx, Event{
		Type:      EventTypeModelRequest,
		AgentName: agentName,
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"model":  model,
			"prompt": prompt,
			"tools":  tools,
		},
	})
}

// ModelResponse records a model response event
func ModelResponse(ctx context.Context, agentName string, model string, response interface{}, err error) {
	details := map[string]interface{}{
		"model":    model,
		"response": response,
	}

	event := Event{
		Type:      EventTypeModelResponse,
		AgentName: agentName,
		Timestamp: time.Now(),
		Details:   details,
	}

	if err != nil {
		event.Error = err
	}

	RecordEventContext(ctx, event)
}

// Handoff records a handoff event
func Handoff(ctx context.Context, fromAgent string, toAgent string, input interface{}) {
	RecordEventContext(ctx, Event{
		Type:      EventTypeHandoff,
		AgentName: fromAgent,
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"to_agent": toAgent,
			"input":    input,
		},
	})
}

// HandoffComplete records when a handoff operation completes and control returns to the originating agent
func HandoffComplete(ctx context.Context, fromAgent string, toAgent string, result interface{}) {
	RecordEventContext(ctx, Event{
		Type:      EventTypeHandoffComplete,
		AgentName: fromAgent,
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"to_agent": toAgent,
			"result":   result,
		},
	})
}

// AgentMessage records an agent message event
func AgentMessage(ctx context.Context, agentName string, role string, content interface{}) {
	RecordEventContext(ctx, Event{
		Type:      EventTypeAgentMessage,
		AgentName: agentName,
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"role":    role,
			"content": content,
		},
	})
}

// Error records an error event
func Error(ctx context.Context, agentName string, message string, err error) {
	details := map[string]interface{}{
		"message": message,
	}

	event := Event{
		Type:      EventTypeError,
		AgentName: agentName,
		Timestamp: time.Now(),
		Details:   details,
		Error:     err,
	}

	RecordEventContext(ctx, event)
}
