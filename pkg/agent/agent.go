package agent

import (
	"reflect"
	"sync"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/model"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/tool"
)

// Agent represents an AI agent with specific configuration
type Agent struct {
	// Core properties
	Name         string
	Instructions string
	Description  string

	// Model configuration
	Model         interface{} // Can be a string (model name) or a Model instance
	ModelSettings *model.ModelSettings

	// Capabilities
	Tools    []tool.Tool
	Handoffs []*Agent

	// Output configuration
	OutputType reflect.Type

	// Lifecycle hooks
	Hooks AgentHooks

	// Internal state
	mu sync.RWMutex
}

// NewAgent creates a new agent with the given name and instructions
func NewAgent(name ...string) *Agent {
	agent := &Agent{
		Tools:    make([]tool.Tool, 0),
		Handoffs: make([]*Agent, 0),
	}
	
	// Set name and instructions if provided
	if len(name) > 0 {
		agent.Name = name[0]
	}
	if len(name) > 1 {
		agent.Instructions = name[1]
	}
	
	return agent
}

// WithModel sets the model for the agent
func (a *Agent) WithModel(model interface{}) *Agent {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Model = model
	return a
}

// WithModelSettings sets the model settings for the agent
func (a *Agent) WithModelSettings(settings *model.ModelSettings) *Agent {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ModelSettings = settings
	return a
}

// WithTools adds tools to the agent
func (a *Agent) WithTools(tools ...tool.Tool) *Agent {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Tools = append(a.Tools, tools...)
	return a
}

// WithHandoffs adds handoffs to the agent
func (a *Agent) WithHandoffs(handoffs ...*Agent) *Agent {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Handoffs = append(a.Handoffs, handoffs...)
	return a
}

// WithOutputType sets the output type for the agent
func (a *Agent) WithOutputType(outputType interface{}) *Agent {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	// Get the type of the output type
	t := reflect.TypeOf(outputType)
	
	// If it's a pointer, get the element type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	
	a.OutputType = t
	return a
}

// WithHooks sets the hooks for the agent
func (a *Agent) WithHooks(hooks AgentHooks) *Agent {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Hooks = hooks
	return a
}

// Clone creates a copy of the agent with optional overrides
func (a *Agent) Clone(overrides map[string]interface{}) *Agent {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	// Create a new agent with the same properties
	clone := &Agent{
		Name:         a.Name,
		Instructions: a.Instructions,
		Description:  a.Description,
		Model:        a.Model,
		ModelSettings: a.ModelSettings,
		Tools:        make([]tool.Tool, len(a.Tools)),
		Handoffs:     make([]*Agent, len(a.Handoffs)),
		OutputType:   a.OutputType,
		Hooks:        a.Hooks,
	}
	
	// Copy tools
	copy(clone.Tools, a.Tools)
	
	// Copy handoffs
	copy(clone.Handoffs, a.Handoffs)
	
	// Apply overrides
	for key, value := range overrides {
		switch key {
		case "Name":
			clone.Name = value.(string)
		case "Instructions":
			clone.Instructions = value.(string)
		case "Description":
			clone.Description = value.(string)
		case "Model":
			clone.Model = value
		case "ModelSettings":
			clone.ModelSettings = value.(*model.ModelSettings)
		case "OutputType":
			clone.WithOutputType(value)
		case "Hooks":
			clone.Hooks = value.(AgentHooks)
		}
	}
	
	return clone
}

// AddFunctionTool adds a function as a tool to the agent
func (a *Agent) AddFunctionTool(name, description string, fn interface{}) *Agent {
	functionTool := tool.NewFunctionTool(name, description, fn)
	return a.WithTools(functionTool)
}

// AsTool transforms this agent into a tool callable by other agents
func (a *Agent) AsTool(toolName, toolDescription string) tool.Tool {
	// TODO: Implement agent as tool
	panic("not implemented")
}

// SetModelProvider sets the model provider for the agent
func (a *Agent) SetModelProvider(provider model.ModelProvider) *Agent {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Model = provider
	return a
}

// SetSystemInstructions sets the system instructions for the agent
func (a *Agent) SetSystemInstructions(instructions string) *Agent {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Instructions = instructions
	return a
}

// AddToolFromDefinition adds a tool from an OpenAI-compatible tool definition
func (a *Agent) AddToolFromDefinition(definition map[string]interface{}, executeFn func(map[string]interface{}) (interface{}, error)) *Agent {
	// Create a tool from the definition
	newTool := tool.CreateToolFromDefinition(definition, executeFn)
	
	// Add the tool to the agent
	return a.WithTools(newTool)
}

// AddToolsFromDefinitions adds multiple tools from OpenAI-compatible tool definitions
func (a *Agent) AddToolsFromDefinitions(definitions []map[string]interface{}, executeFns map[string]func(map[string]interface{}) (interface{}, error)) *Agent {
	tools := make([]tool.Tool, 0, len(definitions))
	
	for _, definition := range definitions {
		// Extract the function name
		functionDef := definition["function"].(map[string]interface{})
		name := functionDef["name"].(string)
		
		// Find the execution function
		executeFn, ok := executeFns[name]
		if !ok {
			// Skip if no execution function
			continue
		}
		
		// Create the tool
		newTool := tool.CreateToolFromDefinition(definition, executeFn)
		tools = append(tools, newTool)
	}
	
	// Add the tools to the agent
	return a.WithTools(tools...)
} 