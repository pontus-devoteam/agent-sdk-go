package lmstudio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"bufio"
	"strings"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/model"
)

// LMStudioModel implements Model for LM Studio
type LMStudioModel struct {
	// Configuration
	ModelName string
	Provider  *LMStudioProvider

	// Internal state
	mu sync.RWMutex
}

// ChatMessage represents a message in a chat
type ChatMessage struct {
	Role     string                   `json:"role"`
	Content  string                   `json:"content,omitempty"`
	Name     string                   `json:"name,omitempty"`
	ToolCalls []ChatMessageToolCall   `json:"tool_calls,omitempty"`
}

// ChatMessageToolCall represents a tool call in a chat message
type ChatMessageToolCall struct {
	ID       string                   `json:"id"`
	Type     string                   `json:"type"`
	Function ChatMessageToolCallFunction `json:"function"`
}

// ChatMessageToolCallFunction represents a function in a tool call
type ChatMessageToolCallFunction struct {
	Name      string                  `json:"name"`
	Arguments string                  `json:"arguments"`
}

// ChatTool represents a tool in a chat
type ChatTool struct {
	Type     string                   `json:"type"`
	Function ChatToolFunction         `json:"function"`
}

// ChatToolFunction represents a function in a tool
type ChatToolFunction struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ChatCompletionRequest represents a request to the chat completions API
type ChatCompletionRequest struct {
	Model             string                `json:"model"`
	Messages          []ChatMessage         `json:"messages"`
	Tools             []ChatTool            `json:"tools,omitempty"`
	ToolChoice        interface{}           `json:"tool_choice,omitempty"`
	Temperature       float64               `json:"temperature,omitempty"`
	TopP              float64               `json:"top_p,omitempty"`
	FrequencyPenalty  float64               `json:"frequency_penalty,omitempty"`
	PresencePenalty   float64               `json:"presence_penalty,omitempty"`
	MaxTokens         int                   `json:"max_tokens,omitempty"`
	Stream            bool                  `json:"stream,omitempty"`
}

// ChatCompletionResponse represents a response from the chat completions API
type ChatCompletionResponse struct {
	ID      string                    `json:"id"`
	Object  string                    `json:"object"`
	Created int64                     `json:"created"`
	Model   string                    `json:"model"`
	Choices []ChatCompletionChoice    `json:"choices"`
	Usage   ChatCompletionUsage       `json:"usage"`
}

// ChatCompletionChoice represents a choice in a chat completion response
type ChatCompletionChoice struct {
	Index        int                  `json:"index"`
	Message      ChatMessage          `json:"message"`
	FinishReason string               `json:"finish_reason"`
}

// ChatCompletionUsage represents usage information in a chat completion response
type ChatCompletionUsage struct {
	PromptTokens     int              `json:"prompt_tokens"`
	CompletionTokens int              `json:"completion_tokens"`
	TotalTokens      int              `json:"total_tokens"`
}

// GetResponse gets a single response from the model
func (m *LMStudioModel) GetResponse(ctx context.Context, request *model.ModelRequest) (*model.ModelResponse, error) {
	// Construct the request
	chatRequest, err := m.constructRequest(request)
	if err != nil {
		return nil, fmt.Errorf("failed to construct request: %w", err)
	}

	// Marshal the request to JSON
	requestBody, err := json.Marshal(chatRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create the HTTP request
	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/chat/completions", m.Provider.BaseURL),
		bytes.NewReader(requestBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpRequest.Header.Set("Content-Type", "application/json")
	if m.Provider.APIKey != "" {
		httpRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.Provider.APIKey))
	}

	// Send the request
	httpResponse, err := m.Provider.HTTPClient.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResponse.Body.Close()

	// Check for errors
	if httpResponse.StatusCode != http.StatusOK {
		return nil, m.handleError(httpResponse)
	}

	// Read the response
	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Unmarshal the response
	var chatResponse ChatCompletionResponse
	if err := json.Unmarshal(responseBody, &chatResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Parse the response
	return m.parseResponse(&chatResponse)
}

// StreamResponse streams a response from the model
func (m *LMStudioModel) StreamResponse(ctx context.Context, request *model.ModelRequest) (<-chan model.StreamEvent, error) {
	// Create a channel for stream events
	eventChan := make(chan model.StreamEvent)

	// Construct the request
	chatRequest, err := m.constructRequest(request)
	if err != nil {
		return nil, fmt.Errorf("failed to construct request: %w", err)
	}

	// Set streaming to true
	chatRequest.Stream = true

	// Marshal the request to JSON
	requestBody, err := json.Marshal(chatRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create the HTTP request
	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/chat/completions", m.Provider.BaseURL),
		bytes.NewReader(requestBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpRequest.Header.Set("Content-Type", "application/json")
	if m.Provider.APIKey != "" {
		httpRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.Provider.APIKey))
	}

	// Send the request
	httpResponse, err := m.Provider.HTTPClient.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Check for errors
	if httpResponse.StatusCode != http.StatusOK {
		httpResponse.Body.Close()
		return nil, m.handleError(httpResponse)
	}

	// Start a goroutine to process the stream
	go func() {
		defer httpResponse.Body.Close()
		defer close(eventChan)

		// Create a scanner to read the response line by line
		scanner := bufio.NewScanner(httpResponse.Body)
		
		// Variables to accumulate the response
		var content string
		var toolCalls []model.ToolCall
		
		// Process each line
		for scanner.Scan() {
			line := scanner.Text()
			
			// Skip empty lines
			if line == "" {
				continue
			}
			
			// Skip lines that don't start with "data: "
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			
			// Extract the data
			data := strings.TrimPrefix(line, "data: ")
			
			// Check if this is the end of the stream
			if data == "[DONE]" {
				break
			}
			
			// Parse the data as JSON
			var chunk struct {
				Choices []struct {
					Delta struct {
						Content   string `json:"content"`
						ToolCalls []struct {
							ID       string `json:"id"`
							Index    int    `json:"index"`
							Type     string `json:"type"`
							Function struct {
								Name      string `json:"name"`
								Arguments string `json:"arguments"`
							} `json:"function"`
						} `json:"tool_calls"`
					} `json:"delta"`
					FinishReason string `json:"finish_reason"`
				} `json:"choices"`
			}
			
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				eventChan <- model.StreamEvent{
					Error: fmt.Errorf("failed to parse chunk: %w", err),
				}
				return
			}
			
			// Process the chunk
			if len(chunk.Choices) > 0 {
				choice := chunk.Choices[0]
				
				// Process content
				if choice.Delta.Content != "" {
					content += choice.Delta.Content
					eventChan <- model.StreamEvent{
						Type:    model.StreamEventTypeContent,
						Content: choice.Delta.Content,
					}
				}
				
				// Process tool calls
				if len(choice.Delta.ToolCalls) > 0 {
					for _, tc := range choice.Delta.ToolCalls {
						// Ensure we have enough tool calls
						for len(toolCalls) <= tc.Index {
							toolCalls = append(toolCalls, model.ToolCall{
								ID:         tc.ID,
								Name:       "",
								Parameters: make(map[string]interface{}),
							})
						}
						
						// Update the tool call
						if tc.Function.Name != "" {
							toolCalls[tc.Index].Name = tc.Function.Name
						}
						
						if tc.Function.Arguments != "" {
							// Try to parse the arguments
							var args map[string]interface{}
							if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err == nil {
								// Merge with existing parameters
								for k, v := range args {
									toolCalls[tc.Index].Parameters[k] = v
								}
							} else {
								// If we can't parse the arguments as JSON, use them as a string
								toolCalls[tc.Index].Parameters["raw_arguments"] = tc.Function.Arguments
							}
							
							// Send a tool call event
							eventChan <- model.StreamEvent{
								Type:     model.StreamEventTypeToolCall,
								ToolCall: &toolCalls[tc.Index],
							}
						}
					}
				}
				
				// Check if we're done
				if choice.FinishReason != "" {
					eventChan <- model.StreamEvent{
						Type: model.StreamEventTypeDone,
						Response: &model.ModelResponse{
							Content:   content,
							ToolCalls: toolCalls,
						},
					}
					break
				}
			}
		}
		
		// Check for scanner errors
		if err := scanner.Err(); err != nil {
			eventChan <- model.StreamEvent{
				Error: fmt.Errorf("error reading stream: %w", err),
			}
		}
	}()

	return eventChan, nil
}

// constructRequest constructs a chat completion request from a model request
func (m *LMStudioModel) constructRequest(request *model.ModelRequest) (*ChatCompletionRequest, error) {
	// Create the chat request
	chatRequest := &ChatCompletionRequest{
		Model:    m.ModelName,
		Messages: make([]ChatMessage, 0),
		Stream:   false,
	}

	// Add system message if provided
	if request.SystemInstructions != "" {
		chatRequest.Messages = append(chatRequest.Messages, ChatMessage{
			Role:    "system",
			Content: request.SystemInstructions,
		})
	}

	// Add input messages
	if input, ok := request.Input.(string); ok {
		// If input is a string, add it as a user message
		chatRequest.Messages = append(chatRequest.Messages, ChatMessage{
			Role:    "user",
			Content: input,
		})
	} else if inputList, ok := request.Input.([]interface{}); ok {
		// If input is a list, add each item as a message
		for _, item := range inputList {
			if message, ok := item.(map[string]interface{}); ok {
				// Handle different message types
				if message["type"] == "message" {
					chatMessage := ChatMessage{
						Role:    message["role"].(string),
						Content: message["content"].(string),
					}
					
					// Add name if provided
					if name, ok := message["name"].(string); ok && name != "" {
						chatMessage.Name = name
					}
					
					chatRequest.Messages = append(chatRequest.Messages, chatMessage)
				} else if message["type"] == "tool_result" {
					// Handle tool results
					toolResult, ok := message["tool_result"].(map[string]interface{})
					if !ok || toolResult == nil {
						// Skip invalid tool results
						continue
					}
					
					toolCall, ok := message["tool_call"].(map[string]interface{})
					if !ok || toolCall == nil {
						// Skip invalid tool calls
						continue
					}
					
					// Create a tool result message
					content := fmt.Sprintf("Tool '%s' returned: %v", 
						toolCall["name"].(string), 
						toolResult["content"])
					
					chatRequest.Messages = append(chatRequest.Messages, ChatMessage{
						Role:    "tool",
						Content: content,
						Name:    toolCall["name"].(string),
					})
				}
			}
		}
	}

	// Add tools if provided
	if len(request.Tools) > 0 || len(request.Handoffs) > 0 {
		// Calculate the expected capacity for tools
		capacity := len(request.Tools)
		if len(request.Handoffs) > 0 {
			capacity += len(request.Handoffs)
		}
		
		chatRequest.Tools = make([]ChatTool, 0, capacity)
		
		// First add regular tools
		for _, tool := range request.Tools {
			// Convert the tool to a chat tool
			var name string
			var description string
			var parameters map[string]interface{}
			
			// Check if the tool is already in OpenAI format, a map[string]interface{}, or implements the Tool interface
			if openAITool, ok := tool.(map[string]interface{}); ok {
				// Check if this is an OpenAI-compatible tool definition
				if openAITool["type"] == "function" && openAITool["function"] != nil {
					// Extract from OpenAI format
					function := openAITool["function"].(map[string]interface{})
					name = function["name"].(string)
					description = function["description"].(string)
					parameters = function["parameters"].(map[string]interface{})
				} else if openAITool["name"] != nil {
					// Legacy direct format
					name = openAITool["name"].(string)
					description = openAITool["description"].(string)
					parameters = openAITool["parameters"].(map[string]interface{})
				} else {
					// Skip invalid tool format
					continue
				}
			} else if tool != nil {
				// Tool implements the interface, call methods
				toolInterface := tool.(interface{
					GetName() string
					GetDescription() string
					GetParametersSchema() map[string]interface{}
				})
				name = toolInterface.GetName()
				description = toolInterface.GetDescription()
				parameters = toolInterface.GetParametersSchema()
			} else {
				// Skip nil tools
				continue
			}
			
			chatTool := ChatTool{
				Type: "function",
				Function: ChatToolFunction{
					Name:        name,
					Description: description,
					Parameters:  parameters,
				},
			}
			
			chatRequest.Tools = append(chatRequest.Tools, chatTool)
		}
		
		// Then add handoff tools 
		for _, handoff := range request.Handoffs {
			// Convert the handoff to a chat tool
			if handoffTool, ok := handoff.(map[string]interface{}); ok {
				// It's already in the right format
				if handoffTool["type"] == "function" && handoffTool["function"] != nil {
					function := handoffTool["function"].(map[string]interface{})
					
					chatTool := ChatTool{
						Type: "function",
						Function: ChatToolFunction{
							Name:        function["name"].(string),
							Description: function["description"].(string),
							Parameters:  function["parameters"].(map[string]interface{}),
						},
					}
					
					chatRequest.Tools = append(chatRequest.Tools, chatTool)
					fmt.Printf("Added handoff tool to request: %s\n", function["name"].(string))
				}
			}
		}
	}

	// Apply model settings if provided
	if request.Settings != nil {
		if request.Settings.Temperature != nil {
			chatRequest.Temperature = *request.Settings.Temperature
		}
		if request.Settings.TopP != nil {
			chatRequest.TopP = *request.Settings.TopP
		}
		if request.Settings.FrequencyPenalty != nil {
			chatRequest.FrequencyPenalty = *request.Settings.FrequencyPenalty
		}
		if request.Settings.PresencePenalty != nil {
			chatRequest.PresencePenalty = *request.Settings.PresencePenalty
		}
		if request.Settings.MaxTokens != nil {
			chatRequest.MaxTokens = *request.Settings.MaxTokens
		}
		if request.Settings.ToolChoice != nil {
			// Handle tool_choice parameter
			if *request.Settings.ToolChoice == "auto" || *request.Settings.ToolChoice == "none" {
				chatRequest.ToolChoice = *request.Settings.ToolChoice
			} else {
				// Assume it's a specific tool name
				chatRequest.ToolChoice = map[string]interface{}{
					"type": "function",
					"function": map[string]interface{}{
						"name": *request.Settings.ToolChoice,
					},
				}
			}
		}
		// Note: parallel_tool_calls is not directly supported in the OpenAI API request
		// It's a client-side setting that affects how tool calls are processed
	}

	return chatRequest, nil
}

// parseResponse parses a chat completion response into a model response
func (m *LMStudioModel) parseResponse(chatResponse *ChatCompletionResponse) (*model.ModelResponse, error) {
	// Check if we have any choices
	if len(chatResponse.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	// Get the first choice
	choice := chatResponse.Choices[0]

	// Create the model response
	response := &model.ModelResponse{
		Content:     choice.Message.Content,
		ToolCalls:   make([]model.ToolCall, 0),
		HandoffCall: nil,
		Usage: &model.Usage{
			PromptTokens:     chatResponse.Usage.PromptTokens,
			CompletionTokens: chatResponse.Usage.CompletionTokens,
			TotalTokens:      chatResponse.Usage.TotalTokens,
		},
	}

	// Parse tool calls if any
	if len(choice.Message.ToolCalls) > 0 {
		for _, toolCall := range choice.Message.ToolCalls {
			// Parse the arguments
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				// If we can't parse the arguments as JSON, use them as a string
				args = map[string]interface{}{
					"raw_arguments": toolCall.Function.Arguments,
				}
			}

			// Check if this is a handoff call
			// The model may try to handoff by calling a tool that begins with "handoff_to_" or similar patterns
			// or by using a tool name that matches an agent name
			if strings.HasPrefix(strings.ToLower(toolCall.Function.Name), "handoff_to_") {
				// Extract the agent name from the tool name
				agentName := strings.TrimPrefix(toolCall.Function.Name, "handoff_to_")
				response.HandoffCall = &model.HandoffCall{
					AgentName: agentName,
					Input:     args["input"].(string),
				}
				continue
			} else if strings.HasPrefix(strings.ToLower(toolCall.Function.Name), "handoff") {
				// Extract the agent name from the arguments
				if agentName, ok := args["agent"].(string); ok {
					input := ""
					if inputVal, ok := args["input"].(string); ok {
						input = inputVal
					} else {
						// Generate an input from the remaining arguments
						inputMap := make(map[string]interface{})
						for k, v := range args {
							if k != "agent" {
								inputMap[k] = v
							}
						}
						inputBytes, _ := json.Marshal(inputMap)
						input = string(inputBytes)
					}
					
					response.HandoffCall = &model.HandoffCall{
						AgentName: agentName,
						Input:     input,
					}
					continue
				}
			} else if strings.Contains(strings.ToLower(toolCall.Function.Name), "agent") {
				// It might be trying to call an agent directly
				possibleAgentName := strings.Replace(strings.ToLower(toolCall.Function.Name), "_agent", " agent", -1)
				possibleAgentName = strings.Title(possibleAgentName)
				
				// Only use this heuristic if the name ends with "Agent"
				if strings.HasSuffix(possibleAgentName, "Agent") {
					// Generate an input from all the arguments
					inputBytes, _ := json.Marshal(args)
					
					response.HandoffCall = &model.HandoffCall{
						AgentName: possibleAgentName,
						Input:     string(inputBytes),
					}
					continue
				}
			}

			// Add the tool call
			response.ToolCalls = append(response.ToolCalls, model.ToolCall{
				ID:         toolCall.ID,
				Name:       toolCall.Function.Name,
				Parameters: args,
			})
		}
	}

	return response, nil
}

// handleError handles an error response from the API
func (m *LMStudioModel) handleError(response *http.Response) error {
	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read error response: %w", err)
	}

	// Try to parse the error
	var errorResponse struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &errorResponse); err == nil && errorResponse.Error.Message != "" {
		return fmt.Errorf("API error (%s): %s", errorResponse.Error.Type, errorResponse.Error.Message)
	}

	// Fallback to status code
	return fmt.Errorf("API error: %s", response.Status)
}