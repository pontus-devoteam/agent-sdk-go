package result

import (
	"github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/model"
)

// StreamEvent represents an event in a streaming response
type StreamEvent struct {
	// Type is the type of the event
	Type string

	// Content is the content of the event
	Content string

	// Item is an item generated during the run
	Item RunItem

	// Agent is the current agent
	Agent *agent.Agent

	// Turn is the current turn
	Turn int

	// Done indicates whether the stream is done
	Done bool

	// Error is an error that occurred
	Error error
}

// StreamedRunResult contains the result of a streamed agent run
type StreamedRunResult struct {
	// RunResult is the base result
	*RunResult

	// Stream is the channel for streaming events
	Stream <-chan model.StreamEvent

	// IsComplete indicates whether the run is complete
	IsComplete bool

	// CurrentAgent is the current agent
	CurrentAgent *agent.Agent

	// CurrentInput is the current input
	CurrentInput any

	// ContinueLoop is continue stream request
	ContinueLoop bool

	// CurrentTurn is the current turn
	CurrentTurn int

	// Task management
	ActiveTasks       map[string]*TaskContext
	DelegationHistory map[string][]string // Agent name -> list of delegatees
}

// ContentEvent creates a content event
func ContentEvent(content string) StreamEvent {
	return StreamEvent{
		Type:    "content",
		Content: content,
	}
}

// ItemEvent creates an item event
func ItemEvent(item RunItem) StreamEvent {
	return StreamEvent{
		Type: "item",
		Item: item,
	}
}

// AgentEvent creates an agent event
func AgentEvent(agent *agent.Agent) StreamEvent {
	return StreamEvent{
		Type:  "agent",
		Agent: agent,
	}
}

// TurnEvent creates a turn event
func TurnEvent(turn int) StreamEvent {
	return StreamEvent{
		Type: "turn",
		Turn: turn,
	}
}

// DoneEvent creates a done event
func DoneEvent() StreamEvent {
	return StreamEvent{
		Type: "done",
		Done: true,
	}
}

// ErrorEvent creates an error event
func ErrorEvent(err error) StreamEvent {
	return StreamEvent{
		Type:  "error",
		Error: err,
	}
}
