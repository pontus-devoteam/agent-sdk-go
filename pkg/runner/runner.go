package runner

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/model"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/result"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/tool"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/tracing"
)

const (
	// DefaultMaxTurns is the default maximum number of turns
	DefaultMaxTurns = 10
)

// Runner executes agents
type Runner struct {
	// Default configuration
	defaultMaxTurns int
	defaultProvider model.Provider

	// Internal state
	mu sync.RWMutex
}

// NewRunner creates a new runner with default configuration
func NewRunner() *Runner {
	return &Runner{
		defaultMaxTurns: DefaultMaxTurns,
	}
}

// WithDefaultMaxTurns sets the default maximum turns
func (r *Runner) WithDefaultMaxTurns(maxTurns int) *Runner {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defaultMaxTurns = maxTurns
	return r
}

// WithDefaultProvider sets the default model provider
func (r *Runner) WithDefaultProvider(provider model.Provider) *Runner {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defaultProvider = provider
	return r
}

// Run executes an agent with the given input and options
func (r *Runner) Run(ctx context.Context, agent AgentType, opts *RunOptions) (*result.RunResult, error) {
	// Apply default options if not provided
	if opts == nil {
		opts = &RunOptions{}
	}

	// Apply default max turns if not provided
	if opts.MaxTurns <= 0 {
		r.mu.RLock()
		opts.MaxTurns = r.defaultMaxTurns
		r.mu.RUnlock()
	}

	// Apply default run config if not provided
	if opts.RunConfig == nil {
		opts.RunConfig = &RunConfig{}
	}

	// Apply default model provider if not provided
	if opts.RunConfig.ModelProvider == nil {
		r.mu.RLock()
		opts.RunConfig.ModelProvider = r.defaultProvider
		r.mu.RUnlock()
	}

	// Check if we have a model provider
	if opts.RunConfig.ModelProvider == nil {
		return nil, errors.New("no model provider available")
	}

	// Run the agent loop
	return r.runAgentLoop(ctx, agent, opts.Input, opts)
}

// RunSync is a synchronous version of Run
func (r *Runner) RunSync(agent AgentType, opts *RunOptions) (*result.RunResult, error) {
	ctx := context.Background()
	return r.Run(ctx, agent, opts)
}

// RunStreaming executes an agent with streaming responses
func (r *Runner) RunStreaming(ctx context.Context, agent AgentType, opts *RunOptions) (*result.StreamedRunResult, error) {
	// Initialize the streaming run with default options
	var err error
	opts, eventCh, err := r.initializeStreamingRun(ctx, agent, opts)
	if err != nil {
		return nil, err
	}

	// Create a streamed run result
	streamedResult := &result.StreamedRunResult{
		RunResult: &result.RunResult{
			Input:       opts.Input,
			NewItems:    make([]result.RunItem, 0),
			LastAgent:   agent,
			FinalOutput: nil,
		},
		Stream:       eventCh,
		IsComplete:   false,
		CurrentAgent: agent,
	}

	// Start a goroutine to run the agent loop
	go func() {
		defer close(eventCh)

		// Call run start hooks
		if err := r.callRunStartHooks(ctx, agent, opts.Input, opts, eventCh); err != nil {
			return
		}

		// Resolve the model
		modelInstance, err := r.resolveModel(agent, opts.RunConfig)
		if err != nil {
			eventCh <- model.StreamEvent{
				Type:  model.StreamEventTypeError,
				Error: fmt.Errorf("failed to resolve model: %w", err),
			}
			return
		}

		// Variables to track consecutive tool calls
		consecutiveToolCalls := 0

		// Run the agent loop
		currentAgent := agent
		currentInput := opts.Input
		for turn := 1; turn <= opts.MaxTurns; turn++ {
			// Update the current turn and agent
			streamedResult.CurrentTurn = turn
			streamedResult.CurrentAgent = currentAgent

			// Call turn start hooks
			if opts.Hooks != nil {
				if err := opts.Hooks.OnTurnStart(ctx, currentAgent, turn); err != nil {
					eventCh <- model.StreamEvent{
						Type:  model.StreamEventTypeError,
						Error: fmt.Errorf("turn start hook error: %w", err),
					}
					return
				}
			}

			// Prepare model settings
			modelSettings := r.prepareModelSettings(currentAgent, opts.RunConfig, consecutiveToolCalls)

			// Prepare model request
			request := &ModelRequestType{
				SystemInstructions: currentAgent.Instructions,
				Input:              currentInput,
				Tools:              r.prepareTools(currentAgent.Tools),
				OutputSchema:       r.prepareOutputSchema(currentAgent.OutputType),
				Handoffs:           r.prepareHandoffs(currentAgent.Handoffs),
				Settings:           modelSettings,
			}

			// Call agent hooks if provided
			if currentAgent.Hooks != nil {
				if err := currentAgent.Hooks.OnBeforeModelCall(ctx, currentAgent, request); err != nil {
					eventCh <- model.StreamEvent{
						Type:  model.StreamEventTypeError,
						Error: fmt.Errorf("before model call hook error: %w", err),
					}
					return
				}
			}

			// Record model request event
			tracing.ModelRequest(ctx, currentAgent.Name, fmt.Sprintf("%v", agent.Model), request.Input, request.Tools)

			// Stream the model response
			modelStream, err := modelInstance.StreamResponse(ctx, request)
			if err != nil {
				eventCh <- model.StreamEvent{
					Type:  model.StreamEventTypeError,
					Error: fmt.Errorf("model call error: %w", err),
				}
				return
			}

			// Process the model stream
			err = r.processModelStream(
				ctx,
				modelStream,
				currentAgent,
				opts,
				streamedResult,
				turn,
				eventCh,
				&consecutiveToolCalls,
			)

			// If the error is nil, we may need to update currentAgent and currentInput
			// Typically this happens after a handoff
			if err == nil && streamedResult.CurrentAgent != currentAgent {
				currentAgent = streamedResult.CurrentAgent
				// If there was a handoff, find the corresponding item to get its input
				if handoffItem := findHandoffItem(streamedResult.RunResult.NewItems); handoffItem != nil {
					currentInput = handoffItem.Input
				}
				continue
			}

			// If there was an error or we're done, exit the loop
			if err != nil || streamedResult.IsComplete {
				return
			}
		}

		// If we get here, we've exceeded the maximum number of turns
		eventCh <- model.StreamEvent{
			Type:  model.StreamEventTypeError,
			Error: fmt.Errorf("exceeded maximum number of turns (%d)", opts.MaxTurns),
		}
	}()

	return streamedResult, nil
}

// findHandoffItem finds the most recent handoff item in the list of run items
func findHandoffItem(items []result.RunItem) *result.HandoffItem {
	for i := len(items) - 1; i >= 0; i-- {
		if handoffItem, ok := items[i].(*result.HandoffItem); ok {
			return handoffItem
		}
	}
	return nil
}

// setupTracing sets up tracing for an agent if not disabled in the options
func (r *Runner) setupTracing(ctx context.Context, agent AgentType, input interface{}, opts *RunOptions) (context.Context, func(), error) {
	// Skip if tracing is disabled
	if opts.RunConfig != nil && opts.RunConfig.TracingDisabled {
		// Return no-op cleanup function
		return ctx, func() {}, nil
	}

	// Create tracer
	tracer, err := tracing.TraceForAgent(agent.Name)
	if err != nil {
		// Log error but continue without tracing
		fmt.Fprintf(os.Stderr, "Failed to create tracer: %v\n", err)
		return ctx, func() {}, nil
	}

	// Add tracer to context
	tracingCtx := tracing.WithTracer(ctx, tracer)

	// Record agent start event
	tracing.AgentStart(tracingCtx, agent.Name, input)

	// Create cleanup function for deferred execution
	cleanup := func() {
		// Record agent end event
		tracing.AgentEnd(tracingCtx, agent.Name, nil) // FinalOutput will be added later

		if err := tracer.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "Error flushing tracer: %v\n", err)
		}
		if err := tracer.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing tracer: %v\n", err)
		}
	}

	return tracingCtx, cleanup, nil
}

// prepareModelSettings creates model settings for a request based on agent and run configuration
func (r *Runner) prepareModelSettings(agent AgentType, runConfig *RunConfig, consecutiveToolCalls int) *ModelSettingsType {
	// Clone the model settings to avoid modifying the original
	var modelSettings *ModelSettingsType

	// Try to use agent settings first
	if agent.ModelSettings != nil {
		// Create a copy of the model settings
		tempSettings := *agent.ModelSettings
		modelSettings = &tempSettings
	} else if runConfig != nil && runConfig.ModelSettings != nil {
		// If no agent settings, use run config settings
		tempSettings := *runConfig.ModelSettings
		modelSettings = &tempSettings
	} else {
		// Create new settings if none exist
		modelSettings = &ModelSettingsType{}
	}

	// Adjust tool_choice if we've had many consecutive calls to the same tool
	if consecutiveToolCalls >= 3 {
		// Suggest the model to provide a text response after multiple tool calls
		// This is a suggestion, not a command - the model can still use tools if needed
		autoChoice := "auto"
		modelSettings.ToolChoice = &autoChoice
	}

	return modelSettings
}

// runAgentLoop runs the agent loop
func (r *Runner) runAgentLoop(ctx context.Context, agent AgentType, input interface{}, opts *RunOptions) (*result.RunResult, error) {
	// Initialize result
	runResult := &result.RunResult{
		Input:        input,
		NewItems:     make([]result.RunItem, 0),
		LastAgent:    agent,
		FinalOutput:  nil,
		RawResponses: make([]model.Response, 0), // Initialize the raw responses slice
	}

	// Set up tracing if not disabled
	var tracingCleanup func()
	ctx, tracingCleanup, _ = r.setupTracing(ctx, agent, input, opts)
	defer func() {
		// Update final output in tracing before cleanup
		if tracingCleanup != nil {
			tracing.AgentEnd(ctx, agent.Name, runResult.FinalOutput)
			tracingCleanup()
		}
	}()

	// Call hooks if provided
	if err := r.callStartHooks(ctx, agent, input, opts); err != nil {
		return nil, err
	}

	// Variables to track consecutive tool calls
	consecutiveToolCalls := 0

	// Run the agent loop
	currentAgent := agent
	currentInput := input
	for turn := 1; turn <= opts.MaxTurns; turn++ {
		// Call turn start hooks
		if err := r.callTurnStartHooks(ctx, currentAgent, turn, opts); err != nil {
			return nil, err
		}

		// Prepare and execute model request
		response, err := r.executeModelRequest(ctx, currentAgent, currentInput, consecutiveToolCalls, opts, turn)
		if err != nil {
			return nil, err
		}

		// Store the raw response in the result
		runResult.RawResponses = append(runResult.RawResponses, *response)

		// Process the response
		// Check if we have a final output (structured output)
		if currentAgent.OutputType != nil {
			// TODO: Implement structured output parsing
			runResult.FinalOutput = response.Content

			// Call hooks if provided
			if err := r.callTurnEndHooks(ctx, currentAgent, turn, response, runResult.FinalOutput, opts); err != nil {
				return nil, err
			}

			break
		}

		// Check if we have a handoff
		if response.HandoffCall != nil {
			nextAgent, nextInput, err := r.processHandoff(ctx, currentAgent, currentInput, response.HandoffCall, runResult, opts)
			if err != nil {
				return nil, err
			}

			if nextAgent != nil {
				// Reset consecutive tool calls counter on handoff
				consecutiveToolCalls = 0
				currentAgent = nextAgent
				currentInput = nextInput
				continue
			}
		}

		// Check if we have tool calls
		if len(response.ToolCalls) > 0 {
			// Process tool calls and update input
			nextInput, continueLoop, toolCallCount := r.processToolCalls(ctx, currentAgent, response, currentInput, consecutiveToolCalls, runResult, turn, opts)
			if continueLoop {
				currentInput = nextInput
				consecutiveToolCalls = toolCallCount
				continue
			}
		} else if response.Content != "" {
			// If we get here with content, we have a final output
			runResult.FinalOutput = response.Content

			// Call hooks if provided
			if err := r.callTurnEndHooks(ctx, currentAgent, turn, response, runResult.FinalOutput, opts); err != nil {
				return nil, err
			}

			break
		}

		// If we reached max turns without a final output, use the last response content
		if turn == opts.MaxTurns && runResult.FinalOutput == nil {
			runResult.FinalOutput = response.Content
		}
	}

	// Call end hooks
	if err := r.callEndHooks(ctx, agent, runResult, opts); err != nil {
		return nil, err
	}

	return runResult, nil
}

// callStartHooks calls the hooks at the start of a run
func (r *Runner) callStartHooks(ctx context.Context, agent AgentType, input interface{}, opts *RunOptions) error {
	// Call hooks if provided
	if opts.Hooks != nil {
		if err := opts.Hooks.OnRunStart(ctx, agent, input); err != nil {
			return fmt.Errorf("run start hook error: %w", err)
		}
	}

	// Call agent hooks if provided
	if agent.Hooks != nil {
		if err := agent.Hooks.OnAgentStart(ctx, agent, input); err != nil {
			return fmt.Errorf("agent start hook error: %w", err)
		}
	}

	return nil
}

// callTurnStartHooks calls the hooks at the start of a turn
func (r *Runner) callTurnStartHooks(ctx context.Context, agent AgentType, turn int, opts *RunOptions) error {
	// Call hooks if provided
	if opts.Hooks != nil {
		if err := opts.Hooks.OnTurnStart(ctx, agent, turn); err != nil {
			return fmt.Errorf("turn start hook error: %w", err)
		}
	}

	return nil
}

// callTurnEndHooks calls the hooks at the end of a turn
func (r *Runner) callTurnEndHooks(ctx context.Context, agent AgentType, turn int, response *model.Response, output interface{}, opts *RunOptions) error {
	// Call hooks if provided
	if opts.Hooks != nil {
		turnResult := &SingleTurnResult{
			Agent:    agent,
			Response: response,
			Output:   output,
		}
		if err := opts.Hooks.OnTurnEnd(ctx, agent, turn, turnResult); err != nil {
			return fmt.Errorf("turn end hook error: %w", err)
		}
	}

	return nil
}

// callEndHooks calls the hooks at the end of a run
func (r *Runner) callEndHooks(ctx context.Context, agent AgentType, runResult *result.RunResult, opts *RunOptions) error {
	// Call agent hooks if provided
	if agent.Hooks != nil {
		if err := agent.Hooks.OnAgentEnd(ctx, agent, runResult.FinalOutput); err != nil {
			return fmt.Errorf("agent end hook error: %w", err)
		}
	}

	// Call hooks if provided
	if opts.Hooks != nil {
		if err := opts.Hooks.OnRunEnd(ctx, runResult); err != nil {
			return fmt.Errorf("run end hook error: %w", err)
		}
	}

	return nil
}

// executeModelRequest prepares and executes a model request
func (r *Runner) executeModelRequest(ctx context.Context, agent AgentType, input interface{}, consecutiveToolCalls int, opts *RunOptions, turn int) (*model.Response, error) {
	// Prepare model settings
	modelSettings := r.prepareModelSettings(agent, opts.RunConfig, consecutiveToolCalls)

	// Prepare model request
	request := &ModelRequestType{
		SystemInstructions: agent.Instructions,
		Input:              input,
		Tools:              r.prepareTools(agent.Tools),
		OutputSchema:       r.prepareOutputSchema(agent.OutputType),
		Handoffs:           r.prepareHandoffs(agent.Handoffs),
		Settings:           modelSettings,
	}

	// Call agent hooks if provided
	if agent.Hooks != nil {
		if err := agent.Hooks.OnBeforeModelCall(ctx, agent, request); err != nil {
			return nil, fmt.Errorf("before model call hook error: %w", err)
		}
	}

	// Record model request event
	tracing.ModelRequest(ctx, agent.Name, fmt.Sprintf("%v", agent.Model), request.Input, request.Tools)

	// Resolve model
	modelInstance, err := r.resolveModel(agent, opts.RunConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve model: %w", err)
	}

	// Call the model
	response, err := modelInstance.GetResponse(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("model call error: %w", err)
	}

	// Record model response event
	tracing.ModelResponse(ctx, agent.Name, fmt.Sprintf("%v", agent.Model), response, err)

	// Call agent hooks if provided
	if agent.Hooks != nil {
		if err := agent.Hooks.OnAfterModelCall(ctx, agent, response); err != nil {
			return nil, fmt.Errorf("after model call hook error: %w", err)
		}
	}

	return response, nil
}

// processHandoff handles a handoff to another agent
func (r *Runner) processHandoff(ctx context.Context, currentAgent AgentType, currentInput interface{}, handoffCall *model.HandoffCall, runResult *result.RunResult, opts *RunOptions) (AgentType, interface{}, error) {
	// Find the handoff agent
	var handoffAgent AgentType
	for _, h := range currentAgent.Handoffs {
		if h.Name == handoffCall.AgentName {
			handoffAgent = h
			break
		}
	}

	// If we didn't find the handoff agent, log it but continue with current agent
	if handoffAgent == nil {
		log.Printf("Error: Handoff to unknown agent: %s", handoffCall.AgentName)
		return nil, nil, nil
	}

	// Record handoff event
	tracing.Handoff(ctx, currentAgent.Name, handoffAgent.Name, handoffCall.Input)

	// Call agent hooks if provided
	if currentAgent.Hooks != nil {
		if err := currentAgent.Hooks.OnBeforeHandoff(ctx, currentAgent, handoffAgent); err != nil {
			return nil, nil, fmt.Errorf("before handoff hook error: %w", err)
		}
	}

	// Add the handoff to the result
	handoffItem := &result.HandoffItem{
		AgentName: handoffAgent.Name,
		Input:     handoffCall.Input,
	}
	runResult.NewItems = append(runResult.NewItems, handoffItem)

	// Create a new context with additional metadata for the sub-agent
	subAgentCtx := context.WithValue(ctx, "parent_agent", currentAgent.Name)

	// Create sub-agent options based on current options
	subAgentOpts := &RunOptions{
		Input:     handoffCall.Input,
		MaxTurns:  opts.MaxTurns,
		RunConfig: opts.RunConfig,
		Hooks:     opts.Hooks, // Preserve hooks for consistent behavior
	}

	// Run the sub-agent to completion and collect its result
	subAgentResult, err := r.Run(subAgentCtx, handoffAgent, subAgentOpts)
	if err != nil {
		return nil, nil, fmt.Errorf("sub-agent execution error: %w", err)
	}

	// Add sub-agent's result items to the parent result
	runResult.NewItems = append(runResult.NewItems, subAgentResult.NewItems...)

	// Record sub-agent completion event
	tracing.HandoffComplete(ctx, currentAgent.Name, handoffAgent.Name, subAgentResult.FinalOutput)

	// Call agent hooks if provided
	if currentAgent.Hooks != nil {
		if err := currentAgent.Hooks.OnAfterHandoff(ctx, currentAgent, handoffAgent, subAgentResult.FinalOutput); err != nil {
			return nil, nil, fmt.Errorf("after handoff hook error: %w", err)
		}
	}

	// Create a special continuation message to ensure the workflow continues
	// This is more effective than just returning the sub-agent's output
	continuationMessage := fmt.Sprintf("Agent %s has completed its task and returned the following result: %s\n\nPlease continue with the next step in your workflow based on this information.",
		handoffAgent.Name,
		subAgentResult.FinalOutput)

	// Return control to the original agent with clear instructions to continue
	return currentAgent, continuationMessage, nil
}

// processToolCalls processes tool calls and updates the input
func (r *Runner) processToolCalls(ctx context.Context, agent AgentType, response *model.Response, currentInput interface{}, currentConsecutiveCalls int, runResult *result.RunResult, turn int, opts *RunOptions) (interface{}, bool, int) {
	// Track consecutive tool calls to the same tool
	toolCallCount := currentConsecutiveCalls
	if len(response.ToolCalls) == 1 {
		toolCallCount++
	} else {
		// Multiple different tools called - reset counter
		toolCallCount = 0
	}

	// Execute the tool calls
	toolResults := make([]interface{}, 0, len(response.ToolCalls))
	for i, tc := range response.ToolCalls {
		// Execute the tool call with our helper function
		modelToolResult, toolCallItem, toolResultItem, err := r.executeToolCall(ctx, agent, tc, turn, i)

		// Add the items to the result
		runResult.NewItems = append(runResult.NewItems, toolCallItem)
		runResult.NewItems = append(runResult.NewItems, toolResultItem)

		// Add the tool result to the list for model input
		toolResults = append(toolResults, modelToolResult)

		// If we had a critical error that wasn't handled in executeToolCall, return it
		if err != nil && (toolCallItem == nil || toolResultItem == nil) {
			// This shouldn't happen, but if it does, just stop processing
			return currentInput, false, toolCallCount
		}
	}

	// Update the input with the tool results
	if len(toolResults) > 0 {
		nextInput := r.updateInputWithToolResults(currentInput, response, toolResults, toolCallCount)
		return nextInput, true, toolCallCount
	}

	// If we didn't have any tool results, don't continue the loop
	return currentInput, false, toolCallCount
}

// updateInputWithToolResults updates the input with the tool results
func (r *Runner) updateInputWithToolResults(currentInput interface{}, response *model.Response, toolResults []interface{}, consecutiveToolCalls int) interface{} {
	// Debug output
	fmt.Println("DEBUG - Updating input with tool results")
	fmt.Printf("DEBUG - Current input type: %T\n", currentInput)
	fmt.Printf("DEBUG - Tool results: %+v\n", toolResults)

	// If the input is a string, convert it to a list
	if _, ok := currentInput.(string); ok {
		currentInput = []interface{}{
			map[string]interface{}{
				"type":    "message",
				"role":    "user",
				"content": currentInput,
			},
		}
	}

	// If the input is not a list, create a new list
	inputList, ok := currentInput.([]interface{})
	if !ok {
		inputList = []interface{}{}
	}

	// Ensure the content field is never null
	content := response.Content
	if content == "" {
		content = " " // Use a space character if content is empty
	}

	// If we have tool calls, we need to do special handling for OpenAI
	if len(response.ToolCalls) > 0 {
		// Create properly formatted tool calls for OpenAI
		toolCalls := make([]map[string]interface{}, len(response.ToolCalls))
		for i, tc := range response.ToolCalls {
			// Ensure we have a valid ID
			toolCallID := tc.ID
			if toolCallID == "" {
				randomBytes := make([]byte, 8)
				if _, err := rand.Read(randomBytes); err == nil {
					toolCallID = fmt.Sprintf("call_%x", randomBytes)
				} else {
					toolCallID = fmt.Sprintf("call_%d", i)
				}
			}

			toolCalls[i] = map[string]interface{}{
				"id":   toolCallID,
				"type": "function",
				"function": map[string]interface{}{
					"name": tc.Name,
					"arguments": func() string {
						args, err := json.Marshal(tc.Parameters)
						if err != nil {
							return "{}"
						}
						return string(args)
					}(),
				},
			}
		}

		// Add the assistant message with tool_calls
		assistantMessage := map[string]interface{}{
			"type":       "message",
			"role":       "assistant",
			"content":    content,
			"tool_calls": toolCalls,
		}
		inputList = append(inputList, assistantMessage)

		// Now add each tool result after the assistant message
		for _, result := range toolResults {
			if toolResult, ok := result.(map[string]interface{}); ok {
				// Debug the tool result before adding
				fmt.Printf("DEBUG - Adding tool result: %+v\n", toolResult)
				inputList = append(inputList, toolResult)
			}
		}
	} else if response.Content != "" {
		// Regular text response without tool calls
		inputList = append(inputList, map[string]interface{}{
			"type":    "message",
			"role":    "assistant",
			"content": response.Content,
		})
	}

	// If we've had several consecutive calls to the same tool, add a prompt
	if consecutiveToolCalls >= 3 {
		inputList = append(inputList, map[string]interface{}{
			"type":    "message",
			"role":    "user",
			"content": "Now that you have the information from the tool(s), please provide a complete response to my original question.",
		})
	}

	// Debug the final input list
	fmt.Println("DEBUG - Final input list:")
	for i, item := range inputList {
		fmt.Printf("DEBUG - Item %d: %+v\n", i, item)
	}

	return inputList
}

// executeToolCall executes a tool call and returns the result
func (r *Runner) executeToolCall(ctx context.Context, agent AgentType, tc model.ToolCall, turn int, idx int) (interface{}, *result.ToolCallItem, *result.ToolResultItem, error) {
	// Find the tool
	var toolToCall tool.Tool
	for _, t := range agent.Tools {
		if t.GetName() == tc.Name {
			toolToCall = t
			break
		}
	}

	// If we didn't find the tool, return an error result
	if toolToCall == nil {
		err := fmt.Errorf("tool not found: %s", tc.Name)
		return createToolResultForError(tc, err, turn, idx),
			&result.ToolCallItem{
				Name:       tc.Name,
				Parameters: tc.Parameters,
			},
			&result.ToolResultItem{
				Name:   tc.Name,
				Result: fmt.Sprintf("Error: %v", err),
			},
			err
	}

	// Record tool call event
	tracing.ToolCall(ctx, agent.Name, tc.Name, tc.Parameters)

	// Call agent hooks if provided
	if agent.Hooks != nil {
		if err := agent.Hooks.OnBeforeToolCall(ctx, agent, toolToCall, tc.Parameters); err != nil {
			return createToolResultForError(tc, err, turn, idx),
				&result.ToolCallItem{
					Name:       tc.Name,
					Parameters: tc.Parameters,
				},
				&result.ToolResultItem{
					Name:   tc.Name,
					Result: fmt.Sprintf("Error: %v", err),
				},
				fmt.Errorf("before tool call hook error: %w", err)
		}
	}

	// Execute the tool
	toolResult, err := toolToCall.Execute(ctx, tc.Parameters)

	// Record tool result event
	tracing.ToolResult(ctx, agent.Name, tc.Name, toolResult, err)

	// Call agent hooks if provided
	if agent.Hooks != nil {
		if hookErr := agent.Hooks.OnAfterToolCall(ctx, agent, toolToCall, toolResult, err); hookErr != nil {
			return createToolResultForError(tc, hookErr, turn, idx),
				&result.ToolCallItem{
					Name:       tc.Name,
					Parameters: tc.Parameters,
				},
				&result.ToolResultItem{
					Name:   tc.Name,
					Result: fmt.Sprintf("Error: %v", hookErr),
				},
				fmt.Errorf("after tool call hook error: %w", hookErr)
		}
	}

	// Handle tool execution error
	if err != nil {
		toolResult = fmt.Sprintf("Error: %v", err)
	}

	// Create the tool call item
	toolCallItem := &result.ToolCallItem{
		Name:       tc.Name,
		Parameters: tc.Parameters,
	}

	// Create the tool result item
	toolResultItem := &result.ToolResultItem{
		Name:   tc.Name,
		Result: toolResult,
	}

	// Use the actual ID from the tool call if available, otherwise generate one
	// For OpenAI, the tool call ID format is important and should be consistent
	toolCallID := tc.ID
	if toolCallID == "" {
		// Generate a tool call ID in the same format as OpenAI's: "call_<random_string>"
		randomBytes := make([]byte, 8)
		if _, err := rand.Read(randomBytes); err != nil {
			toolCallID = fmt.Sprintf("call_%d_%d", turn, idx)
		} else {
			toolCallID = fmt.Sprintf("call_%x", randomBytes)
		}
	}

	// Detect if the provider is Anthropic based on model provider name
	isAnthropic := false
	if agent.Model != nil {
		providerName := reflect.TypeOf(agent.Model).String()
		if strings.Contains(strings.ToLower(providerName), "anthropic") {
			isAnthropic = true
		}
	}

	var modelToolResult interface{}

	if isAnthropic {
		// Anthropic uses a different format for tool results
		modelToolResult = map[string]interface{}{
			"role":         "tool",
			"tool_call_id": toolCallID,
			"content":      fmt.Sprintf("%v", toolResult),
		}
	} else {
		// Standard OpenAI format
		modelToolResult = map[string]interface{}{
			"type": "tool_result",
			"tool_call": map[string]interface{}{
				"name":       tc.Name,
				"id":         toolCallID,
				"parameters": tc.Parameters,
			},
			"tool_result": map[string]interface{}{
				"content": toolResult,
			},
		}
	}

	return modelToolResult, toolCallItem, toolResultItem, nil
}

// createToolResultForError creates a structured tool result for an error
func createToolResultForError(tc model.ToolCall, err error, turn int, idx int) interface{} {
	// Generate a tool call ID
	toolCallID := tc.ID
	if toolCallID == "" {
		randomBytes := make([]byte, 8)
		if _, err := rand.Read(randomBytes); err != nil {
			toolCallID = fmt.Sprintf("call_%d_%d", turn, idx)
		} else {
			toolCallID = fmt.Sprintf("call_%x", randomBytes)
		}
	}

	return map[string]interface{}{
		"type": "tool_result",
		"tool_call": map[string]interface{}{
			"name":       tc.Name,
			"id":         toolCallID,
			"parameters": tc.Parameters,
		},
		"tool_result": map[string]interface{}{
			"content": fmt.Sprintf("Error: %v", err),
		},
	}
}

// resolveModel resolves the model for the agent
func (r *Runner) resolveModel(agent AgentType, runConfig *RunConfig) (model.Model, error) {
	// If runConfig.Model is set, it overrides agent.Model
	modelToUse := agent.Model
	if runConfig.Model != nil {
		modelToUse = runConfig.Model
	}

	// If model is a string, use the provider to resolve it
	if modelName, ok := modelToUse.(string); ok {
		return runConfig.ModelProvider.GetModel(modelName)
	}

	// If model is a Model instance, use it directly
	if model, ok := modelToUse.(model.Model); ok {
		return model, nil
	}

	return nil, fmt.Errorf("invalid model type: %T", modelToUse)
}

// prepareTools prepares tools for the model request
func (r *Runner) prepareTools(tools []tool.Tool) []interface{} {
	// If no tools, return nil
	if len(tools) == 0 {
		return nil
	}

	// Convert tools to the OpenAI format using the utility function
	openAITools := tool.ToOpenAITools(tools)

	// Convert to []interface{} for compatibility with the ModelRequest
	result := make([]interface{}, len(openAITools))
	for i, t := range openAITools {
		result[i] = t
	}

	return result
}

// prepareOutputSchema prepares the output schema for the model request
func (r *Runner) prepareOutputSchema(outputType reflect.Type) interface{} {
	// If no output type, return nil
	if outputType == nil {
		return nil
	}

	// Create a schema for the output type
	schema := map[string]interface{}{
		"type":       "object",
		"properties": make(map[string]interface{}),
		"required":   []string{},
	}

	// Process each field in the struct
	for i := 0; i < outputType.NumField(); i++ {
		field := outputType.Field(i)

		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		// Get the JSON tag
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Parse the JSON tag
		jsonName := strings.Split(jsonTag, ",")[0]

		// Add to required fields if not omitempty
		if !strings.Contains(jsonTag, "omitempty") {
			schema["required"] = append(schema["required"].([]string), jsonName)
		}

		// Get the field schema
		fieldSchema := r.getFieldSchema(field.Type)

		// Add to properties
		schema["properties"].(map[string]interface{})[jsonName] = fieldSchema
	}

	return schema
}

// getFieldSchema returns the JSON schema for a field type
func (r *Runner) getFieldSchema(fieldType reflect.Type) map[string]interface{} {
	schema := make(map[string]interface{})

	// Handle pointers
	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}

	// Handle different types
	switch fieldType.Kind() {
	case reflect.Bool:
		schema["type"] = "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema["type"] = "integer"
	case reflect.Float32, reflect.Float64:
		schema["type"] = "number"
	case reflect.String:
		schema["type"] = "string"
	case reflect.Slice, reflect.Array:
		schema["type"] = "array"
		schema["items"] = r.getFieldSchema(fieldType.Elem())
	case reflect.Map:
		schema["type"] = "object"
		if fieldType.Key().Kind() == reflect.String {
			schema["additionalProperties"] = r.getFieldSchema(fieldType.Elem())
		}
	case reflect.Struct:
		schema["type"] = "object"
		schema["properties"] = make(map[string]interface{})
		schema["required"] = []string{}

		for i := 0; i < fieldType.NumField(); i++ {
			field := fieldType.Field(i)

			// Skip unexported fields
			if field.PkgPath != "" {
				continue
			}

			// Get the JSON tag
			jsonTag := field.Tag.Get("json")
			if jsonTag == "" || jsonTag == "-" {
				continue
			}

			// Parse the JSON tag
			jsonName := strings.Split(jsonTag, ",")[0]

			// Add to required fields if not omitempty
			if !strings.Contains(jsonTag, "omitempty") {
				schema["required"] = append(schema["required"].([]string), jsonName)
			}

			// Get the field schema
			fieldSchema := r.getFieldSchema(field.Type)

			// Add to properties
			schema["properties"].(map[string]interface{})[jsonName] = fieldSchema
		}
	default:
		schema["type"] = "string"
	}

	return schema
}

// prepareHandoffs prepares handoffs for the model request
func (r *Runner) prepareHandoffs(handoffs []AgentType) []interface{} {
	// If no handoffs, return nil
	if len(handoffs) == 0 {
		return nil
	}

	// Convert handoffs to the format expected by the model
	// Format them as tools so the model can call them directly
	result := make([]interface{}, len(handoffs))
	for i, h := range handoffs {
		handoffToolName := fmt.Sprintf("handoff_to_%s", h.Name)

		result[i] = map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        handoffToolName,
				"description": fmt.Sprintf("Handoff the conversation to the %s. Use this when a query requires expertise from %s.", h.Name, h.Name),
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"input": map[string]interface{}{
							"type":        "string",
							"description": "The specific request to send to the agent. Be clear about what you're asking the agent to do.",
						},
					},
					"required": []string{"input"},
				},
			},
		}

		// Debug log
		fmt.Printf("Added handoff tool for agent: %s with name: %s\n", h.Name, handoffToolName)
	}

	return result
}

// initializeStreamingRun initializes the streaming run with default options and creates the result channel
func (r *Runner) initializeStreamingRun(ctx context.Context, agent AgentType, opts *RunOptions) (*RunOptions, chan model.StreamEvent, error) {
	// Apply default options if not provided
	if opts == nil {
		opts = &RunOptions{}
	}

	// Apply default max turns if not provided
	if opts.MaxTurns <= 0 {
		r.mu.RLock()
		opts.MaxTurns = r.defaultMaxTurns
		r.mu.RUnlock()
	}

	// Apply default run config if not provided
	if opts.RunConfig == nil {
		opts.RunConfig = &RunConfig{}
	}

	// Apply default model provider if not provided
	if opts.RunConfig.ModelProvider == nil {
		r.mu.RLock()
		opts.RunConfig.ModelProvider = r.defaultProvider
		r.mu.RUnlock()
	}

	// Check if we have a model provider
	if opts.RunConfig.ModelProvider == nil {
		return nil, nil, errors.New("no model provider available")
	}

	// Create a channel for stream events
	eventCh := make(chan model.StreamEvent)

	return opts, eventCh, nil
}

// callRunStartHooks calls the start hooks for the run and agent
func (r *Runner) callRunStartHooks(ctx context.Context, agent AgentType, input interface{}, opts *RunOptions, eventCh chan model.StreamEvent) error {
	// Call hooks if provided
	if opts.Hooks != nil {
		if err := opts.Hooks.OnRunStart(ctx, agent, input); err != nil {
			eventCh <- model.StreamEvent{
				Type:  model.StreamEventTypeError,
				Error: fmt.Errorf("run start hook error: %w", err),
			}
			return err
		}
	}

	// Call agent hooks if provided
	if agent.Hooks != nil {
		if err := agent.Hooks.OnAgentStart(ctx, agent, input); err != nil {
			eventCh <- model.StreamEvent{
				Type:  model.StreamEventTypeError,
				Error: fmt.Errorf("agent start hook error: %w", err),
			}
			return err
		}
	}

	return nil
}

// processModelStream processes the model's stream events
func (r *Runner) processModelStream(
	ctx context.Context,
	modelStream <-chan model.StreamEvent,
	currentAgent AgentType,
	opts *RunOptions,
	streamedResult *result.StreamedRunResult,
	turn int,
	eventCh chan model.StreamEvent,
	consecutiveToolCalls *int,
) error {
	var content string
	var toolCalls []model.ToolCall
	var handoffCall *model.HandoffCall

	for event := range modelStream {
		// Check for errors
		if event.Error != nil {
			eventCh <- model.StreamEvent{
				Type:  model.StreamEventTypeError,
				Error: fmt.Errorf("model stream error: %w", event.Error),
			}
			return event.Error
		}

		// Process the event based on its type
		switch event.Type {
		case model.StreamEventTypeContent:
			// Append to content
			content += event.Content
			// Forward the event
			eventCh <- event

		case model.StreamEventTypeToolCall:
			// Add to tool calls
			if event.ToolCall != nil {
				toolCalls = append(toolCalls, *event.ToolCall)
			}
			// Forward the event
			eventCh <- event

		case model.StreamEventTypeHandoff:
			// Set handoff call
			handoffCall = event.HandoffCall
			// Forward the event
			eventCh <- event

		case model.StreamEventTypeDone:
			// Create the final response
			response := &model.Response{
				Content:     content,
				ToolCalls:   toolCalls,
				HandoffCall: handoffCall,
			}

			// Call agent hooks if provided
			if currentAgent.Hooks != nil {
				if err := currentAgent.Hooks.OnAfterModelCall(ctx, currentAgent, response); err != nil {
					eventCh <- model.StreamEvent{
						Type:  model.StreamEventTypeError,
						Error: fmt.Errorf("after model call hook error: %w", err),
					}
					return err
				}
			}

			// Handle structured output if applicable
			if currentAgent.OutputType != nil {
				return r.handleFinalOutput(ctx, currentAgent, response, opts, streamedResult, turn, eventCh)
			}

			// Handle handoff if applicable
			if handoffCall != nil {
				// Process handoff and prepare for next turn
				nextAgent, _, err := r.handleHandoff(
					ctx,
					currentAgent,
					handoffCall,
					response,
					opts,
					streamedResult,
					turn,
					eventCh,
				)
				if err != nil {
					return err
				}

				// Update the current agent for the next turn
				streamedResult.CurrentAgent = nextAgent
				*consecutiveToolCalls = 0
				return nil // Exit the event loop to start the next turn
			}

			// Handle regular text response
			if response.Content != "" {
				return r.handleTextResponse(ctx, currentAgent, response, opts, streamedResult, turn, eventCh)
			}

			// If we reached max turns without a final output, use the last response content
			if turn == opts.MaxTurns && streamedResult.RunResult.FinalOutput == nil {
				streamedResult.RunResult.FinalOutput = response.Content
				streamedResult.IsComplete = true
				return nil
			}
		}
	}

	return nil
}

// handleFinalOutput processes final output (structured or text)
func (r *Runner) handleFinalOutput(
	ctx context.Context,
	currentAgent AgentType,
	response *model.Response,
	opts *RunOptions,
	streamedResult *result.StreamedRunResult,
	turn int,
	eventCh chan model.StreamEvent,
) error {
	// If the agent has an output type defined, treat this as structured output
	// TODO: Implement structured output parsing
	streamedResult.RunResult.FinalOutput = response.Content

	// Call hooks if provided
	if opts.Hooks != nil {
		turnResult := &SingleTurnResult{
			Agent:    currentAgent,
			Response: response,
			Output:   streamedResult.RunResult.FinalOutput,
		}
		if err := opts.Hooks.OnTurnEnd(ctx, currentAgent, turn, turnResult); err != nil {
			eventCh <- model.StreamEvent{
				Type:  model.StreamEventTypeError,
				Error: fmt.Errorf("turn end hook error: %w", err),
			}
			return err
		}
	}

	// Send done event
	eventCh <- model.StreamEvent{
		Type:     model.StreamEventTypeDone,
		Response: response,
	}

	// Mark as complete
	streamedResult.IsComplete = true
	streamedResult.RunResult.LastAgent = currentAgent

	// Call agent hooks if provided
	if currentAgent.Hooks != nil {
		if err := currentAgent.Hooks.OnAgentEnd(ctx, currentAgent, streamedResult.RunResult.FinalOutput); err != nil {
			eventCh <- model.StreamEvent{
				Type:  model.StreamEventTypeError,
				Error: fmt.Errorf("agent end hook error: %w", err),
			}
			return err
		}
	}

	// Call hooks if provided
	if opts.Hooks != nil {
		if err := opts.Hooks.OnRunEnd(ctx, streamedResult.RunResult); err != nil {
			eventCh <- model.StreamEvent{
				Type:  model.StreamEventTypeError,
				Error: fmt.Errorf("run end hook error: %w", err),
			}
			return err
		}
	}

	return nil
}

// handleTextResponse processes a text response
func (r *Runner) handleTextResponse(
	ctx context.Context,
	currentAgent AgentType,
	response *model.Response,
	opts *RunOptions,
	streamedResult *result.StreamedRunResult,
	turn int,
	eventCh chan model.StreamEvent,
) error {
	// Use the response content as the final output
	streamedResult.RunResult.FinalOutput = response.Content

	// Call hooks if provided
	if opts.Hooks != nil {
		turnResult := &SingleTurnResult{
			Agent:    currentAgent,
			Response: response,
			Output:   streamedResult.RunResult.FinalOutput,
		}
		if err := opts.Hooks.OnTurnEnd(ctx, currentAgent, turn, turnResult); err != nil {
			eventCh <- model.StreamEvent{
				Type:  model.StreamEventTypeError,
				Error: fmt.Errorf("turn end hook error: %w", err),
			}
			return err
		}
	}

	// Mark as complete
	streamedResult.IsComplete = true
	streamedResult.RunResult.LastAgent = currentAgent

	// Call agent hooks if provided
	if currentAgent.Hooks != nil {
		if err := currentAgent.Hooks.OnAgentEnd(ctx, currentAgent, streamedResult.RunResult.FinalOutput); err != nil {
			eventCh <- model.StreamEvent{
				Type:  model.StreamEventTypeError,
				Error: fmt.Errorf("agent end hook error: %w", err),
			}
			return err
		}
	}

	// Call hooks if provided
	if opts.Hooks != nil {
		if err := opts.Hooks.OnRunEnd(ctx, streamedResult.RunResult); err != nil {
			eventCh <- model.StreamEvent{
				Type:  model.StreamEventTypeError,
				Error: fmt.Errorf("run end hook error: %w", err),
			}
			return err
		}
	}

	return nil
}

// handleHandoff processes a handoff to another agent in streaming mode
func (r *Runner) handleHandoff(
	ctx context.Context,
	currentAgent AgentType,
	handoffCall *model.HandoffCall,
	response *model.Response,
	opts *RunOptions,
	streamedResult *result.StreamedRunResult,
	turn int,
	eventCh chan model.StreamEvent,
) (AgentType, interface{}, error) {
	// Find the handoff agent
	var handoffAgent AgentType
	for _, h := range currentAgent.Handoffs {
		if h.Name == handoffCall.AgentName {
			handoffAgent = h
			break
		}
	}

	// If we found the handoff agent, update the current agent and input
	if handoffAgent != nil {
		// Record handoff event
		tracing.Handoff(ctx, currentAgent.Name, handoffAgent.Name, handoffCall.Input)

		// Add a handoff item to the run result
		handoffItem := &result.HandoffItem{
			AgentName: handoffAgent.Name,
			Input:     handoffCall.Input,
		}
		streamedResult.RunResult.NewItems = append(streamedResult.RunResult.NewItems, handoffItem)

		// Call agent hooks if provided
		if handoffAgent.Hooks != nil {
			if err := handoffAgent.Hooks.OnAgentStart(ctx, handoffAgent, handoffCall.Input); err != nil {
				eventCh <- model.StreamEvent{
					Type:  model.StreamEventTypeError,
					Error: fmt.Errorf("agent start hook error: %w", err),
				}
				return nil, nil, err
			}
		}

		// Call hooks if provided
		if opts.Hooks != nil {
			turnResult := &SingleTurnResult{
				Agent:    currentAgent,
				Response: response,
				Output:   nil, // No output for handoff
			}
			if err := opts.Hooks.OnTurnEnd(ctx, currentAgent, turn, turnResult); err != nil {
				eventCh <- model.StreamEvent{
					Type:  model.StreamEventTypeError,
					Error: fmt.Errorf("turn end hook error: %w", err),
				}
				return nil, nil, err
			}
		}

		// Create a continuation prompt to ensure the workflow continues after handoff returns
		// This prompt is stored in the context but not sent directly to the model yet
		continuationPrompt := fmt.Sprintf("Agent %s has been called. When it returns, please continue with the next step in your workflow based on its response.", handoffAgent.Name)

		// Set the prompt in the context for use when the handoff returns
		ctx = context.WithValue(ctx, "continuation_prompt", continuationPrompt)

		// Update the streamed result with the handoff
		eventCh <- model.StreamEvent{
			Type:    model.StreamEventTypeHandoff,
			Content: fmt.Sprintf("Handing off to %s...", handoffAgent.Name),
			HandoffCall: &model.HandoffCall{
				AgentName: handoffAgent.Name,
				Input:     handoffCall.Input,
			},
		}

		// Return the next agent and input
		return handoffAgent, handoffCall.Input, nil
	}

	// If we didn't find the handoff agent, log it but continue with current agent
	log.Printf("Error: Handoff to unknown agent: %s", handoffCall.AgentName)
	eventCh <- model.StreamEvent{
		Type:    model.StreamEventTypeError,
		Content: fmt.Sprintf("Error: Handoff to unknown agent: %s", handoffCall.AgentName),
		Error:   fmt.Errorf("handoff to unknown agent: %s", handoffCall.AgentName),
	}

	// Continue with the current agent and input
	return currentAgent, streamedResult.RunResult.Input, nil
}
