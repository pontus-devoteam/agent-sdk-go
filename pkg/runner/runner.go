package runner

import (
	"context"
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
	defaultProvider model.ModelProvider

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
func (r *Runner) WithDefaultProvider(provider model.ModelProvider) *Runner {
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

// RunStreaming executes an agent with streaming results
func (r *Runner) RunStreaming(ctx context.Context, agent AgentType, opts *RunOptions) (*result.StreamedRunResult, error) {
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

	// Create a channel for stream events
	eventCh := make(chan model.StreamEvent)

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

		// Call hooks if provided
		if opts.Hooks != nil {
			if err := opts.Hooks.OnRunStart(ctx, agent, opts.Input); err != nil {
				eventCh <- model.StreamEvent{
					Type:  model.StreamEventTypeError,
					Error: fmt.Errorf("run start hook error: %w", err),
				}
				return
			}
		}

		// Call agent hooks if provided
		if agent.Hooks != nil {
			if err := agent.Hooks.OnAgentStart(ctx, agent, opts.Input); err != nil {
				eventCh <- model.StreamEvent{
					Type:  model.StreamEventTypeError,
					Error: fmt.Errorf("agent start hook error: %w", err),
				}
				return
			}
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
			// Update the current turn
			streamedResult.CurrentTurn = turn
			streamedResult.CurrentAgent = currentAgent

			// Call hooks if provided
			if opts.Hooks != nil {
				if err := opts.Hooks.OnTurnStart(ctx, currentAgent, turn); err != nil {
					eventCh <- model.StreamEvent{
						Type:  model.StreamEventTypeError,
						Error: fmt.Errorf("turn start hook error: %w", err),
					}
					return
				}
			}

			// Clone the model settings to avoid modifying the original
			var modelSettings *ModelSettingsType
			if currentAgent.ModelSettings != nil {
				// Create a copy of the model settings
				tempSettings := *currentAgent.ModelSettings
				modelSettings = &tempSettings
			} else if opts.RunConfig.ModelSettings != nil {
				// Create a copy of the run config settings
				tempSettings := *opts.RunConfig.ModelSettings
				modelSettings = &tempSettings
			} else {
				// Create new settings
				modelSettings = &ModelSettingsType{}
			}

			// Adjust tool_choice if we've had many consecutive calls to the same tool
			if consecutiveToolCalls >= 3 {
				// Suggest the model to provide a text response after multiple tool calls
				// This is a suggestion, not a command - the model can still use tools if needed
				autoChoice := "auto"
				modelSettings.ToolChoice = &autoChoice
			}

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

			// Process the stream
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
					return
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
					response := &model.ModelResponse{
						Content:     content,
						ToolCalls:   toolCalls,
						HandoffCall: handoffCall,
						Usage:       nil, // Usage not available in streaming
					}
					
					// Call agent hooks if provided
					if currentAgent.Hooks != nil {
						if err := currentAgent.Hooks.OnAfterModelCall(ctx, currentAgent, response); err != nil {
							eventCh <- model.StreamEvent{
								Type:  model.StreamEventTypeError,
								Error: fmt.Errorf("after model call hook error: %w", err),
							}
							return
						}
					}

					// Process the response
					// Check if we have a final output (structured output)
					if currentAgent.OutputType != nil {
						// TODO: Implement structured output parsing
						streamedResult.RunResult.FinalOutput = content
						
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
								return
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
						if agent.Hooks != nil {
							if err := agent.Hooks.OnAgentEnd(ctx, agent, streamedResult.RunResult.FinalOutput); err != nil {
								eventCh <- model.StreamEvent{
									Type:  model.StreamEventTypeError,
									Error: fmt.Errorf("agent end hook error: %w", err),
								}
								return
							}
						}
						
						// Call hooks if provided
						if opts.Hooks != nil {
							if err := opts.Hooks.OnRunEnd(ctx, streamedResult.RunResult); err != nil {
								eventCh <- model.StreamEvent{
									Type:  model.StreamEventTypeError,
									Error: fmt.Errorf("run end hook error: %w", err),
								}
								return
							}
						}
						
						return
					}

					// Check if we have a handoff
					if handoffCall != nil {
						// Reset consecutive tool calls counter on handoff
						consecutiveToolCalls = 0

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

							// Update the current agent and input
							currentAgent = handoffAgent
							currentInput = handoffCall.Input

							// Call agent hooks if provided
							if currentAgent.Hooks != nil {
								if err := currentAgent.Hooks.OnAgentStart(ctx, currentAgent, currentInput); err != nil {
									eventCh <- model.StreamEvent{
										Type:  model.StreamEventTypeError,
										Error: fmt.Errorf("agent start hook error: %w", err),
									}
									return
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
									return
								}
							}

							// Break out of the current event loop to start the next turn with the new agent
							break
						} else {
							// If we didn't find the handoff agent, log an error
							log.Printf("Error: Handoff to unknown agent: %s", handoffCall.AgentName)

							// Continue with the current agent
							continue
						}
					} else if response.Content != "" {
						// Reset consecutive tool calls counter if we have content
						consecutiveToolCalls = 0

						// If we get here with content, we have a final output
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
								return
							}
						}

						// Make sure we exit the loop by returning here
						streamedResult.IsComplete = true
						streamedResult.RunResult.LastAgent = currentAgent
						
						// Call agent hooks if provided
						if agent.Hooks != nil {
							if err := agent.Hooks.OnAgentEnd(ctx, agent, streamedResult.RunResult.FinalOutput); err != nil {
								eventCh <- model.StreamEvent{
									Type:  model.StreamEventTypeError,
									Error: fmt.Errorf("agent end hook error: %w", err),
								}
								return
							}
						}
						
						// Call hooks if provided
						if opts.Hooks != nil {
							if err := opts.Hooks.OnRunEnd(ctx, streamedResult.RunResult); err != nil {
								eventCh <- model.StreamEvent{
									Type:  model.StreamEventTypeError,
									Error: fmt.Errorf("run end hook error: %w", err),
								}
								return
							}
						}
						
						return
					}

					// If we reached max turns without a final output, use the last response content
					if turn == opts.MaxTurns && streamedResult.RunResult.FinalOutput == nil {
						streamedResult.RunResult.FinalOutput = response.Content
						streamedResult.IsComplete = true
						return
					}
				}
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

// runAgentLoop runs the agent loop
func (r *Runner) runAgentLoop(ctx context.Context, agent AgentType, input interface{}, opts *RunOptions) (*result.RunResult, error) {
	// Initialize result
	runResult := &result.RunResult{
		Input:      input,
		NewItems:   make([]result.RunItem, 0),
		LastAgent:  agent,
		FinalOutput: nil,
	}

	// Set up tracing if not disabled
	if opts.RunConfig == nil || !opts.RunConfig.TracingDisabled {
		tracer, err := tracing.TraceForAgent(agent.Name)
		if err != nil {
			// Log error but continue
			fmt.Fprintf(os.Stderr, "Failed to create tracer: %v\n", err)
		} else {
			// Add tracer to context
			ctx = tracing.WithTracer(ctx, tracer)
			
			// Record agent start event
			tracing.AgentStart(ctx, agent.Name, input)
			
			// Ensure tracer is closed when done
			defer func() {
				tracing.AgentEnd(ctx, agent.Name, runResult.FinalOutput)
				tracer.Flush()
				tracer.Close()
			}()
		}
	}

	// Call hooks if provided
	if opts.Hooks != nil {
		if err := opts.Hooks.OnRunStart(ctx, agent, input); err != nil {
			return nil, fmt.Errorf("run start hook error: %w", err)
		}
	}

	// Call agent hooks if provided
	if agent.Hooks != nil {
		if err := agent.Hooks.OnAgentStart(ctx, agent, input); err != nil {
			return nil, fmt.Errorf("agent start hook error: %w", err)
		}
	}

	// Resolve the model
	model, err := r.resolveModel(agent, opts.RunConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve model: %w", err)
	}

	// Variables to track consecutive tool calls
	consecutiveToolCalls := 0

	// Run the agent loop
	currentAgent := agent
	currentInput := input
	for turn := 1; turn <= opts.MaxTurns; turn++ {
		// Call hooks if provided
		if opts.Hooks != nil {
			if err := opts.Hooks.OnTurnStart(ctx, currentAgent, turn); err != nil {
				return nil, fmt.Errorf("turn start hook error: %w", err)
			}
		}

		// Clone the model settings to avoid modifying the original
		var modelSettings *ModelSettingsType
		if currentAgent.ModelSettings != nil {
			// Create a copy of the model settings
			tempSettings := *currentAgent.ModelSettings
			modelSettings = &tempSettings
		} else if opts.RunConfig.ModelSettings != nil {
			// Create a copy of the run config settings
			tempSettings := *opts.RunConfig.ModelSettings
			modelSettings = &tempSettings
		} else {
			// Create new settings
			modelSettings = &ModelSettingsType{}
		}

		// Adjust tool_choice if we've had many consecutive calls to the same tool
		if consecutiveToolCalls >= 3 {
			// Suggest the model to provide a text response after multiple tool calls
			// This is a suggestion, not a command - the model can still use tools if needed
			autoChoice := "auto"
			modelSettings.ToolChoice = &autoChoice
		}

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
				return nil, fmt.Errorf("before model call hook error: %w", err)
			}
		}

		// Record model request event
		tracing.ModelRequest(ctx, currentAgent.Name, fmt.Sprintf("%v", agent.Model), request.Input, request.Tools)

		// Call the model
		response, err := model.GetResponse(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("model call error: %w", err)
		}

		// Record model response event
		tracing.ModelResponse(ctx, currentAgent.Name, fmt.Sprintf("%v", agent.Model), response, err)

		// Call agent hooks if provided
		if currentAgent.Hooks != nil {
			if err := currentAgent.Hooks.OnAfterModelCall(ctx, currentAgent, response); err != nil {
				return nil, fmt.Errorf("after model call hook error: %w", err)
			}
		}

		// Process the response
		// Check if we have a final output (structured output)
		if currentAgent.OutputType != nil {
			// TODO: Implement structured output parsing
			runResult.FinalOutput = response.Content
			
			// Call hooks if provided
			if opts.Hooks != nil {
				turnResult := &SingleTurnResult{
					Agent:    currentAgent,
					Response: response,
					Output:   runResult.FinalOutput,
				}
				if err := opts.Hooks.OnTurnEnd(ctx, currentAgent, turn, turnResult); err != nil {
					return nil, fmt.Errorf("turn end hook error: %w", err)
				}
			}
			
			break
		}

		// Check if we have a handoff
		if response.HandoffCall != nil {
			// Reset consecutive tool calls counter on handoff
			consecutiveToolCalls = 0

			// Find the handoff agent
			var handoffAgent AgentType
			for _, h := range currentAgent.Handoffs {
				if h.Name == response.HandoffCall.AgentName {
					handoffAgent = h
					break
				}
			}

			// If we found the handoff agent, update the current agent and input
			if handoffAgent != nil {
				// Record handoff event
				tracing.Handoff(ctx, currentAgent.Name, handoffAgent.Name, response.HandoffCall.Input)

				// Call agent hooks if provided
				if currentAgent.Hooks != nil {
					if err := currentAgent.Hooks.OnBeforeHandoff(ctx, currentAgent, handoffAgent); err != nil {
						return nil, fmt.Errorf("before handoff hook error: %w", err)
					}
				}

				// Add the handoff to the result
				handoffItem := &result.HandoffItem{
					AgentName: handoffAgent.Name,
					Input:     response.HandoffCall.Input,
				}
				runResult.NewItems = append(runResult.NewItems, handoffItem)

				// Update the current agent and input
				currentAgent = handoffAgent
				currentInput = response.HandoffCall.Input

				// Call agent hooks if provided
				if currentAgent.Hooks != nil {
					if err := currentAgent.Hooks.OnAfterHandoff(ctx, currentAgent, handoffAgent, response.HandoffCall.Input); err != nil {
						return nil, fmt.Errorf("after handoff hook error: %w", err)
					}
				}

				// Continue to the next turn
				continue
			}
		}

		// Check if we have tool calls
		if len(response.ToolCalls) > 0 {
			// Track consecutive tool calls to the same tool
			if len(response.ToolCalls) == 1 {
				consecutiveToolCalls++
			} else {
				// Multiple different tools called - reset counter
				consecutiveToolCalls = 0
			}

			// Execute the tool calls
			toolResults := make([]interface{}, 0, len(response.ToolCalls))
			for _, tc := range response.ToolCalls {
				// Find the tool
				var toolToCall tool.Tool
				for _, t := range currentAgent.Tools {
					if t.GetName() == tc.Name {
						toolToCall = t
						break
					}
				}

				// If we found the tool, execute it
				if toolToCall != nil {
					// Record tool call event
					tracing.ToolCall(ctx, currentAgent.Name, tc.Name, tc.Parameters)

					// Call agent hooks if provided
					if currentAgent.Hooks != nil {
						if err := currentAgent.Hooks.OnBeforeToolCall(ctx, currentAgent, toolToCall, tc.Parameters); err != nil {
							return nil, fmt.Errorf("before tool call hook error: %w", err)
						}
					}

					// Execute the tool
					toolResult, err := toolToCall.Execute(ctx, tc.Parameters)
					
					// Record tool result event
					tracing.ToolResult(ctx, currentAgent.Name, tc.Name, toolResult, err)

					// Call agent hooks if provided
					if currentAgent.Hooks != nil {
						if err := currentAgent.Hooks.OnAfterToolCall(ctx, currentAgent, toolToCall, toolResult, err); err != nil {
							return nil, fmt.Errorf("after tool call hook error: %w", err)
						}
					}

					// Handle tool execution error
					if err != nil {
						toolResult = fmt.Sprintf("Error: %v", err)
					}

					// Add the tool call and result to the result
					toolCallItem := &result.ToolCallItem{
						Name:       tc.Name,
						Parameters: tc.Parameters,
					}
					runResult.NewItems = append(runResult.NewItems, toolCallItem)

					toolResultItem := &result.ToolResultItem{
						Name:   tc.Name,
						Result: toolResult,
					}
					runResult.NewItems = append(runResult.NewItems, toolResultItem)

					// Add the tool result to the list
					toolResults = append(toolResults, map[string]interface{}{
						"type": "tool_result",
						"tool_call": map[string]interface{}{
							"name": tc.Name,
							"id": fmt.Sprintf("call_%d_%d", turn, len(toolResults)),
							"parameters": tc.Parameters,
						},
						"tool_result": map[string]interface{}{
							"content": toolResult,
						},
					})
				}
			}

			// Update the input with the tool results
			if len(toolResults) > 0 {
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

				// Add the response as an assistant message
				if response.Content != "" {
					inputList = append(inputList, map[string]interface{}{
						"type":    "message",
						"role":    "assistant",
						"content": response.Content,
					})
				} else {
					// Add a message representing the tool calls
					var toolCallsDescription string
					if len(response.ToolCalls) == 1 {
						toolCallsDescription = fmt.Sprintf("You called the tool: %s", response.ToolCalls[0].Name)
					} else {
						toolNames := make([]string, len(response.ToolCalls))
						for i, tc := range response.ToolCalls {
							toolNames[i] = tc.Name
						}
						toolCallsDescription = fmt.Sprintf("You called these tools: %s", strings.Join(toolNames, ", "))
					}
					
					inputList = append(inputList, map[string]interface{}{
						"type":    "message",
						"role":    "assistant",
						"content": toolCallsDescription,
					})
				}

				// Add the tool results
				inputList = append(inputList, toolResults...)

				// If we've had several consecutive calls to the same tool, add a prompt
				if consecutiveToolCalls >= 3 {
					inputList = append(inputList, map[string]interface{}{
						"type":    "message",
						"role":    "user",
						"content": "Now that you have the information from the tool(s), please provide a complete response to my original question.",
					})
				}

				// Update the input
				currentInput = inputList

				// Continue to the next turn
				continue
			}
		} else if response.Content != "" {
			// Reset consecutive tool calls counter if we have content
			consecutiveToolCalls = 0
            
			// If we get here with content, we have a final output
			runResult.FinalOutput = response.Content

			// Call hooks if provided
			if opts.Hooks != nil {
				turnResult := &SingleTurnResult{
					Agent:    currentAgent,
					Response: response,
					Output:   runResult.FinalOutput,
				}
				if err := opts.Hooks.OnTurnEnd(ctx, currentAgent, turn, turnResult); err != nil {
					return nil, fmt.Errorf("turn end hook error: %w", err)
				}
			}

			break
		}

		// If we reached max turns without a final output, use the last response content
		if turn == opts.MaxTurns && runResult.FinalOutput == nil {
			runResult.FinalOutput = response.Content
		}
	}

	// Call agent hooks if provided
	if agent.Hooks != nil {
		if err := agent.Hooks.OnAgentEnd(ctx, agent, runResult.FinalOutput); err != nil {
			return nil, fmt.Errorf("agent end hook error: %w", err)
		}
	}

	// Call hooks if provided
	if opts.Hooks != nil {
		if err := opts.Hooks.OnRunEnd(ctx, runResult); err != nil {
			return nil, fmt.Errorf("run end hook error: %w", err)
		}
	}

	return runResult, nil
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
		"type": "object",
		"properties": make(map[string]interface{}),
		"required": []string{},
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
				"name": handoffToolName,
				"description": fmt.Sprintf("Handoff the conversation to the %s. Use this when a query requires expertise from %s.", h.Name, h.Name),
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"input": map[string]interface{}{
							"type": "string",
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

// mergeModelSettings merges agent model settings with run config model settings
func (r *Runner) mergeModelSettings(agentSettings, configSettings *model.ModelSettings) *model.ModelSettings {
	// If no agent settings, return config settings
	if agentSettings == nil {
		return configSettings
	}

	// If no config settings, return agent settings
	if configSettings == nil {
	return agentSettings
	}

	// Create a new settings object
	mergedSettings := &model.ModelSettings{}

	// Merge temperature
	if configSettings.Temperature != nil {
		mergedSettings.Temperature = configSettings.Temperature
	} else {
		mergedSettings.Temperature = agentSettings.Temperature
	}

	// Merge top_p
	if configSettings.TopP != nil {
		mergedSettings.TopP = configSettings.TopP
	} else {
		mergedSettings.TopP = agentSettings.TopP
	}

	// Merge frequency penalty
	if configSettings.FrequencyPenalty != nil {
		mergedSettings.FrequencyPenalty = configSettings.FrequencyPenalty
	} else {
		mergedSettings.FrequencyPenalty = agentSettings.FrequencyPenalty
	}

	// Merge presence penalty
	if configSettings.PresencePenalty != nil {
		mergedSettings.PresencePenalty = configSettings.PresencePenalty
	} else {
		mergedSettings.PresencePenalty = agentSettings.PresencePenalty
	}

	// Merge tool choice
	if configSettings.ToolChoice != nil {
		mergedSettings.ToolChoice = configSettings.ToolChoice
	} else {
		mergedSettings.ToolChoice = agentSettings.ToolChoice
	}

	// Merge parallel tool calls
	if configSettings.ParallelToolCalls != nil {
		mergedSettings.ParallelToolCalls = configSettings.ParallelToolCalls
	} else {
		mergedSettings.ParallelToolCalls = agentSettings.ParallelToolCalls
	}

	// Merge max tokens
	if configSettings.MaxTokens != nil {
		mergedSettings.MaxTokens = configSettings.MaxTokens
	} else {
		mergedSettings.MaxTokens = agentSettings.MaxTokens
	}

	return mergedSettings
} 