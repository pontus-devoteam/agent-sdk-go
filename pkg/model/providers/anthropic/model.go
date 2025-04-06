package anthropic

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math"
	mathrand "math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/model"
)

// Model implements the model.Model interface for Anthropic
type Model struct {
	// Configuration
	ModelName string
	Provider  *Provider
}

// AnthropicMessage represents a message in a conversation
type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnthropicTool represents a tool in Anthropic's API
type AnthropicTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// AnthropicToolUse represents a tool use in Anthropic's API
type AnthropicToolUse struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Input     map[string]interface{} `json:"input"`
	Type      string                 `json:"type"`
	Submitted bool                   `json:"submitted"`
}

// AnthropicMessageRequest represents a request to the messages API
type AnthropicMessageRequest struct {
	Model         string             `json:"model"`
	Messages      []AnthropicMessage `json:"messages"`
	System        string             `json:"system,omitempty"`
	MaxTokens     int                `json:"max_tokens,omitempty"`
	Temperature   float64            `json:"temperature,omitempty"`
	TopP          float64            `json:"top_p,omitempty"`
	TopK          int                `json:"top_k,omitempty"`
	Stream        bool               `json:"stream,omitempty"`
	Tools         []AnthropicTool    `json:"tools,omitempty"`
	ToolChoice    string             `json:"tool_choice,omitempty"`
	StopSequences []string           `json:"stop_sequences,omitempty"`
}

// AnthropicMessageResponse represents a response from the messages API
type AnthropicMessageResponse struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Role         string                 `json:"role"`
	Content      []AnthropicContent     `json:"content"`
	Model        string                 `json:"model"`
	StopReason   string                 `json:"stop_reason"`
	StopSequence string                 `json:"stop_sequence"`
	Usage        AnthropicUsage         `json:"usage"`
	ToolUse      []AnthropicToolUse     `json:"tool_use,omitempty"`
	Error        map[string]interface{} `json:"error,omitempty"`
}

// AnthropicStreamResponse represents a streaming response from the messages API
type AnthropicStreamResponse struct {
	Type         string                   `json:"type"`
	Message      AnthropicMessageResponse `json:"message,omitempty"`
	Index        int                      `json:"index,omitempty"`
	ContentBlock *AnthropicContent        `json:"content_block,omitempty"`
	Delta        *AnthropicDelta          `json:"delta,omitempty"`
	Error        map[string]interface{}   `json:"error,omitempty"`
}

// AnthropicDelta represents a delta in a streaming response
type AnthropicDelta struct {
	Type         string `json:"type"`
	Text         string `json:"text,omitempty"`
	StopReason   string `json:"stop_reason,omitempty"`
	StopSequence string `json:"stop_sequence,omitempty"`
}

// AnthropicContent represents content in a message
type AnthropicContent struct {
	Type  string                 `json:"type"`
	Text  string                 `json:"text,omitempty"`
	ID    string                 `json:"id,omitempty"`
	Name  string                 `json:"name,omitempty"`
	Input map[string]interface{} `json:"input,omitempty"`
}

// AnthropicUsage represents token usage in a response
type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ErrorResponse represents an error response from the API
type ErrorResponse struct {
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// GetResponse gets a single response from the model with retry logic
func (m *Model) GetResponse(ctx context.Context, request *model.Request) (*model.Response, error) {
	var response *model.Response
	var lastErr error

	// Try with exponential backoff
	for attempt := 0; attempt <= m.Provider.MaxRetries; attempt++ {
		// Wait for rate limit
		m.Provider.WaitForRateLimit()

		// If this is not the first attempt, wait with exponential backoff
		if attempt > 0 {
			backoffDuration := calculateBackoff(attempt, m.Provider.RetryAfter)
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during backoff: %w", ctx.Err())
			case <-time.After(backoffDuration):
				// Continue after backoff
			}
		}

		// Try to get a response
		response, lastErr = m.getResponseOnce(ctx, request)

		// If successful or not a rate limit error, return
		if lastErr == nil {
			return response, nil
		}

		// If it's not a rate limit error, don't retry
		if !isRateLimitError(lastErr) {
			return nil, lastErr
		}

		// If we've exceeded the maximum number of retries, return the last error
		if attempt == m.Provider.MaxRetries {
			return nil, fmt.Errorf("exceeded maximum number of retries (%d): %w", m.Provider.MaxRetries, lastErr)
		}
	}

	// This should never happen
	return nil, lastErr
}

// getResponseOnce attempts to get a response from the model once
func (m *Model) getResponseOnce(ctx context.Context, request *model.Request) (*model.Response, error) {
	// Construct the request
	anthropicRequest, err := m.constructRequest(request)
	if err != nil {
		return nil, fmt.Errorf("failed to construct request: %w", err)
	}

	// Marshal the request to JSON
	requestBody, err := json.Marshal(anthropicRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Print the request for debugging
	if os.Getenv("ANTHROPIC_DEBUG") == "1" {
		fmt.Println("DEBUG - Anthropic Request:", string(requestBody))
	}

	// Create the HTTP request
	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/messages", m.Provider.BaseURL),
		bytes.NewReader(requestBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("x-api-key", m.Provider.APIKey)
	httpRequest.Header.Set("anthropic-version", "2023-06-01")

	// Send the request
	httpResponse, err := m.Provider.HTTPClient.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if closeErr := httpResponse.Body.Close(); closeErr != nil {
			// If we already have an error, keep it as the primary error
			if err == nil {
				err = fmt.Errorf("error closing response body: %w", closeErr)
			}
		}
	}()

	// Check for errors
	if httpResponse.StatusCode != http.StatusOK {
		return nil, m.handleError(httpResponse)
	}

	// Read the response
	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Print the response for debugging
	if os.Getenv("ANTHROPIC_DEBUG") == "1" {
		fmt.Println("DEBUG - Anthropic Response:", string(responseBody))
	}

	// Unmarshal the response
	var anthropicResponse AnthropicMessageResponse
	if err := json.Unmarshal(responseBody, &anthropicResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Update token count for rate limiting
	if anthropicResponse.Usage.InputTokens > 0 && anthropicResponse.Usage.OutputTokens > 0 {
		m.Provider.UpdateTokenCount(anthropicResponse.Usage.InputTokens + anthropicResponse.Usage.OutputTokens)
	}

	// Parse the response
	return m.parseResponse(&anthropicResponse)
}

// StreamResponse streams a response from the model with retry logic
func (m *Model) StreamResponse(ctx context.Context, request *model.Request) (<-chan model.StreamEvent, error) {
	// Create a channel for stream events
	eventChan := make(chan model.StreamEvent)

	go func() {
		defer close(eventChan)

		var lastErr error

		// Try with exponential backoff
		for attempt := 0; attempt <= m.Provider.MaxRetries; attempt++ {
			// Wait for rate limit
			m.Provider.WaitForRateLimit()

			// If this is not the first attempt, wait with exponential backoff
			if attempt > 0 {
				backoffDuration := calculateBackoff(attempt, m.Provider.RetryAfter)
				select {
				case <-ctx.Done():
					eventChan <- model.StreamEvent{
						Type:  model.StreamEventTypeError,
						Error: fmt.Errorf("context cancelled during backoff: %w", ctx.Err()),
					}
					return
				case <-time.After(backoffDuration):
					// Continue after backoff
				}
			}

			// Try to stream a response
			err := m.streamResponseOnce(ctx, request, eventChan)

			// If successful or context cancelled, return
			if err == nil || ctx.Err() != nil {
				return
			}

			// Store the last error
			lastErr = err

			// If it's not a rate limit error, don't retry
			if !isRateLimitError(err) {
				eventChan <- model.StreamEvent{
					Type:  model.StreamEventTypeError,
					Error: err,
				}
				return
			}

			// If we've exceeded the maximum number of retries, return the last error
			if attempt == m.Provider.MaxRetries {
				eventChan <- model.StreamEvent{
					Type:  model.StreamEventTypeError,
					Error: fmt.Errorf("exceeded maximum number of retries (%d): %w", m.Provider.MaxRetries, lastErr),
				}
				return
			}

			// Inform the client that we're retrying
			eventChan <- model.StreamEvent{
				Type:    model.StreamEventTypeContent,
				Content: fmt.Sprintf("Rate limit exceeded, retrying in %v...", calculateBackoff(attempt+1, m.Provider.RetryAfter)),
			}
		}
	}()

	return eventChan, nil
}

// streamResponseOnce attempts to stream a response from the model once
func (m *Model) streamResponseOnce(ctx context.Context, request *model.Request, eventChan chan<- model.StreamEvent) error {
	// Construct the request
	anthropicRequest, err := m.constructRequest(request)
	if err != nil {
		return fmt.Errorf("failed to construct request: %w", err)
	}

	// Enable streaming
	anthropicRequest.Stream = true

	// Marshal the request to JSON
	requestBody, err := json.Marshal(anthropicRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create the HTTP request
	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/messages", m.Provider.BaseURL),
		bytes.NewReader(requestBody),
	)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("x-api-key", m.Provider.APIKey)
	httpRequest.Header.Set("anthropic-version", "2023-06-01")
	httpRequest.Header.Set("Accept", "text/event-stream")

	// Send the request
	httpResponse, err := m.Provider.HTTPClient.Do(httpRequest)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResponse.Body.Close()

	// Check for errors
	if httpResponse.StatusCode != http.StatusOK {
		return m.handleError(httpResponse)
	}

	// Read the response as a stream
	reader := bufio.NewReader(httpResponse.Body)
	tokenCount := 0
	var content strings.Builder
	var currentToolCall *model.ToolCall

	for {
		// Check if the context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Continue
		}

		// Read the next line
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// End of stream
				break
			}
			return fmt.Errorf("error reading stream: %w", err)
		}

		// Skip empty lines
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for data prefix
		if !strings.HasPrefix(line, "data:") {
			// Not a data line (could be a comment or other SSE field)
			continue
		}

		// Strip "data: " prefix
		data := strings.TrimPrefix(line, "data:")
		data = strings.TrimSpace(data)

		// Check for end marker
		if data == "[DONE]" {
			break
		}

		// Try to parse as JSON
		var streamResp AnthropicStreamResponse
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			// Skip lines that aren't valid JSON
			continue
		}

		// Handle different event types
		switch streamResp.Type {
		case "message_start":
			// Message start event
			continue

		case "content_block_start":
			// Content block start event
			if streamResp.ContentBlock != nil && streamResp.ContentBlock.Type == "text" {
				// Reset the content builder
				content.Reset()
			}
			continue

		case "content_block_delta":
			// Content block delta event
			if streamResp.Delta != nil && streamResp.Delta.Type == "text_delta" {
				// Append the text
				content.WriteString(streamResp.Delta.Text)
				tokenCount++

				// Send a content event
				eventChan <- model.StreamEvent{
					Type:    model.StreamEventTypeContent,
					Content: streamResp.Delta.Text,
				}
			}
			continue

		case "content_block_stop":
			// Content block stop event
			continue

		case "tool_use_start":
			// Tool start event
			if len(streamResp.Message.ToolUse) > 0 {
				toolUse := streamResp.Message.ToolUse[0]

				// Create a new tool call
				currentToolCall = &model.ToolCall{
					ID:         toolUse.ID,
					Name:       toolUse.Name,
					Parameters: toolUse.Input,
				}
			}
			continue

		case "tool_use_delta":
			// Tool delta event (we'll handle this if needed)
			continue

		case "tool_use_stop":
			// Tool stop event
			if currentToolCall != nil {
				// Send a tool call event
				eventChan <- model.StreamEvent{
					Type:     model.StreamEventTypeToolCall,
					ToolCall: currentToolCall,
				}

				// Reset the current tool call
				currentToolCall = nil
			}
			continue

		case "message_delta":
			// Message delta event
			if streamResp.Delta != nil && streamResp.Delta.StopReason != "" {
				// End of message
				break
			}
			continue

		case "message_stop":
			// Message stop event
			break

		case "error":
			// Error event
			errMsg := "Unknown error"
			if streamResp.Error != nil {
				if msg, ok := streamResp.Error["message"].(string); ok {
					errMsg = msg
				}
			}
			return fmt.Errorf("stream error: %s", errMsg)
		}
	}

	// Update token count for rate limiting
	if tokenCount > 0 {
		m.Provider.UpdateTokenCount(tokenCount)
	}

	// Final done event
	eventChan <- model.StreamEvent{
		Type: model.StreamEventTypeDone,
		Done: true,
	}

	return nil
}

// constructRequest constructs an Anthropic API request from a model.Request
func (m *Model) constructRequest(request *model.Request) (*AnthropicMessageRequest, error) {
	// Convert input to messages
	messages, err := m.createMessages(request.Input)
	if err != nil {
		return nil, fmt.Errorf("failed to create messages: %w", err)
	}

	// Create the Anthropic request
	anthropicRequest := &AnthropicMessageRequest{
		Model:     m.ModelName,
		Messages:  messages,
		MaxTokens: 4096, // Default max_tokens value since it's required by the API
	}

	// Set system instructions if provided
	if request.SystemInstructions != "" {
		anthropicRequest.System = request.SystemInstructions
	}

	// Set model settings if provided
	if request.Settings != nil {
		if request.Settings.Temperature != nil {
			anthropicRequest.Temperature = *request.Settings.Temperature
		}
		if request.Settings.TopP != nil {
			anthropicRequest.TopP = *request.Settings.TopP
		}
		if request.Settings.MaxTokens != nil {
			anthropicRequest.MaxTokens = *request.Settings.MaxTokens
		}
	}

	// Handle tools if provided
	if len(request.Tools) > 0 {
		tools, err := m.createTools(request.Tools)
		if err != nil {
			return nil, fmt.Errorf("failed to create tools: %w", err)
		}
		anthropicRequest.Tools = tools

		// Anthropic handles tool_choice differently from OpenAI
		// Only set it if it's explicitly in a format Anthropic understands
		if request.Settings != nil && request.Settings.ToolChoice != nil {
			toolChoice := *request.Settings.ToolChoice
			// Anthropic only supports "auto" and "none" as string values
			if toolChoice == "auto" || toolChoice == "none" {
				anthropicRequest.ToolChoice = toolChoice
			} else {
				// Don't set it if it's not a supported value
				// This avoids the "Input should be a valid dictionary" error
				if os.Getenv("ANTHROPIC_DEBUG") == "1" {
					fmt.Println("DEBUG - Ignoring unsupported tool_choice value:", toolChoice)
				}
			}
		}
	}

	return anthropicRequest, nil
}

// createMessages creates AnthropicMessages from a model.Request.Input
func (m *Model) createMessages(input interface{}) ([]AnthropicMessage, error) {
	// Debug the input
	if os.Getenv("ANTHROPIC_DEBUG") == "1" {
		fmt.Println("DEBUG - Creating Anthropic messages from input:", input)
	}

	var messages []AnthropicMessage

	// Convert input to messages based on type
	switch v := input.(type) {
	case string:
		// Single string input becomes a user message
		messages = []AnthropicMessage{
			{
				Role:    "user",
				Content: v,
			},
		}
	case []interface{}:
		// Array of messages
		for _, msgInterface := range v {
			msg, ok := msgInterface.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid message format: %v", msgInterface)
			}

			// Debug the message being processed
			if os.Getenv("ANTHROPIC_DEBUG") == "1" {
				fmt.Printf("DEBUG - Processing message: %+v\n", msg)
			}

			// First check if this is a tool result with special format
			if toolResultType, ok := msg["type"].(string); ok && toolResultType == "tool_result" {
				if os.Getenv("ANTHROPIC_DEBUG") == "1" {
					fmt.Println("DEBUG - Found tool result message")
				}

				// This is an OpenAI-style tool result, which needs special handling for Anthropic
				toolCall, _ := msg["tool_call"].(map[string]interface{})
				toolResult, _ := msg["tool_result"].(map[string]interface{})

				if toolCall != nil && toolResult != nil {
					// Format in Anthropic's expected format
					toolCallID, _ := toolCall["id"].(string)
					content, _ := toolResult["content"]

					if toolCallID != "" && content != nil {
						// Add a tool message in Anthropic's format
						messages = append(messages, AnthropicMessage{
							Role:    "user", // Anthropic expects tool results as user messages
							Content: fmt.Sprintf("Tool result for call %s: %v", toolCallID, content),
						})
						continue
					}
				}
			}

			// Next check for role and handle accordingly
			role, ok := msg["role"].(string)
			if !ok {
				return nil, fmt.Errorf("message missing role: %v", msg)
			}

			// Check if this is a tool message (Anthropic's format)
			if role == "tool" {
				if os.Getenv("ANTHROPIC_DEBUG") == "1" {
					fmt.Println("DEBUG - Found tool message with role=tool")
				}

				// Get tool call ID and content
				toolCallID, _ := msg["tool_call_id"].(string)
				content, _ := msg["content"].(string)

				if toolCallID != "" && content != "" {
					// Add a tool response as a user message for Anthropic
					messages = append(messages, AnthropicMessage{
						Role:    "user",
						Content: fmt.Sprintf("Tool response [%s]: %s", toolCallID, content),
					})
					continue
				}
			}

			// Standard message handling
			content, ok := msg["content"].(string)
			if !ok || strings.TrimSpace(content) == "" {
				// For Anthropic API, content is required and cannot be empty
				if role == "assistant" {
					// If an assistant message has tool calls but no content,
					// add a descriptive message about the tool usage
					if toolCallsInterface, exists := msg["tool_calls"]; exists {
						toolCallsArray, ok := toolCallsInterface.([]interface{})
						if ok && len(toolCallsArray) > 0 {
							var toolNames []string
							for _, tc := range toolCallsArray {
								if tcMap, ok := tc.(map[string]interface{}); ok {
									if fn, ok := tcMap["function"].(map[string]interface{}); ok {
										if name, ok := fn["name"].(string); ok {
											toolNames = append(toolNames, name)
										}
									}
								}
							}
							if len(toolNames) > 0 {
								content = fmt.Sprintf("I'll use the %s tool to help with this.", strings.Join(toolNames, ", "))
							} else {
								content = "I'll use a tool to help with this request."
							}
						} else {
							content = "I'll use a tool to help with this request."
						}
					} else {
						content = "I'm processing your request."
					}
				} else {
					// Fallback for other roles or cases
					content = "Processing your request..."
				}
			}

			// Map OpenAI roles to Anthropic roles
			anthropicRole := mapRole(role)

			messages = append(messages, AnthropicMessage{
				Role:    anthropicRole,
				Content: content,
			})
		}
	default:
		return nil, fmt.Errorf("unsupported input type: %T", input)
	}

	// Print the final messages
	if os.Getenv("ANTHROPIC_DEBUG") == "1" {
		fmt.Println("DEBUG - Final Anthropic messages:")
		for i, msg := range messages {
			fmt.Printf("DEBUG - Message %d: {Role: %s, Content: %s}\n", i, msg.Role, msg.Content)
		}
	}

	return messages, nil
}

// mapRole maps OpenAI message roles to Anthropic roles
func mapRole(role string) string {
	switch role {
	case "user":
		return "user"
	case "assistant":
		return "assistant"
	case "system":
		return "user" // Anthropic doesn't have a system role in messages
	case "tool":
		return "assistant" // Tools map to assistant in Anthropic
	default:
		return "user" // Default to user for unknown roles
	}
}

// createTools creates AnthropicTools from model.Request.Tools
func (m *Model) createTools(tools []interface{}) ([]AnthropicTool, error) {
	var anthropicTools []AnthropicTool

	if os.Getenv("ANTHROPIC_DEBUG") == "1" {
		fmt.Println("DEBUG - Creating tools from:", tools)
	}

	for _, tool := range tools {
		toolMap, ok := tool.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid tool format: %v", tool)
		}

		// Extract function details
		function, ok := toolMap["function"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("tool missing function: %v", toolMap)
		}

		name, ok := function["name"].(string)
		if !ok {
			return nil, fmt.Errorf("function missing name: %v", function)
		}

		description := ""
		if descVal, ok := function["description"].(string); ok {
			description = descVal
		}

		parameters := make(map[string]interface{})
		if paramsVal, ok := function["parameters"].(map[string]interface{}); ok {
			parameters = paramsVal
		}

		tool := AnthropicTool{
			Name:        name,
			Description: description,
			InputSchema: parameters,
		}

		if os.Getenv("ANTHROPIC_DEBUG") == "1" {
			fmt.Printf("DEBUG - Created Anthropic tool: %+v\n", tool)
		}
		anthropicTools = append(anthropicTools, tool)
	}

	return anthropicTools, nil
}

// parseResponse parses an Anthropic API response into a model.Response
func (m *Model) parseResponse(anthropicResponse *AnthropicMessageResponse) (*model.Response, error) {
	response := &model.Response{
		Usage: &model.Usage{
			PromptTokens:     anthropicResponse.Usage.InputTokens,
			CompletionTokens: anthropicResponse.Usage.OutputTokens,
			TotalTokens:      anthropicResponse.Usage.InputTokens + anthropicResponse.Usage.OutputTokens,
		},
	}

	// Extract content
	var textContent strings.Builder
	for _, content := range anthropicResponse.Content {
		if content.Type == "text" {
			textContent.WriteString(content.Text)
		} else if content.Type == "tool_use" {
			// Handle tool calls
			toolCall := model.ToolCall{
				ID:         content.ID,
				Name:       content.Name,
				Parameters: content.Input,
			}
			response.ToolCalls = append(response.ToolCalls, toolCall)
		}
	}
	response.Content = textContent.String()

	// Extract tool calls from top-level ToolUse field if present
	for _, tool := range anthropicResponse.ToolUse {
		// Skip duplicates - sometimes content can contain the same tool calls
		isDuplicate := false
		for _, existing := range response.ToolCalls {
			if existing.ID == tool.ID {
				isDuplicate = true
				break
			}
		}

		if !isDuplicate && tool.Type == "tool_use" && tool.Submitted {
			if os.Getenv("ANTHROPIC_DEBUG") == "1" {
				fmt.Printf("DEBUG - Found tool call: %+v\n", tool)
			}
			response.ToolCalls = append(response.ToolCalls, model.ToolCall{
				ID:         tool.ID,
				Name:       tool.Name,
				Parameters: tool.Input,
			})
		}
	}

	// If there are tool calls but no content, set a placeholder
	if response.Content == "" && len(response.ToolCalls) > 0 {
		response.Content = " " // Set a non-empty placeholder
	}

	return response, nil
}

// handleError handles an error response from the API
func (m *Model) handleError(response *http.Response) error {
	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("error reading error response: %w (status code: %d)", err, response.StatusCode)
	}

	// Try to parse the error response
	var errorResponse ErrorResponse
	if err := json.Unmarshal(body, &errorResponse); err != nil {
		// If we can't parse the error response, return a generic error
		return fmt.Errorf("API error: %s (%d)", http.StatusText(response.StatusCode), response.StatusCode)
	}

	// Return a formatted error
	return fmt.Errorf("API error: %s (%s)", errorResponse.Error.Message, errorResponse.Error.Type)
}

// isRateLimitError checks if an error is a rate limit error
func isRateLimitError(err error) bool {
	return strings.Contains(err.Error(), "rate limit") ||
		strings.Contains(err.Error(), "Too Many Requests") ||
		strings.Contains(err.Error(), "429")
}

// calculateBackoff calculates the backoff duration based on the attempt number
func calculateBackoff(attempt int, baseDelay time.Duration) time.Duration {
	// Use exponential backoff with jitter
	backoff := float64(baseDelay) * math.Pow(2, float64(attempt-1))

	// Add jitter (up to 20%)
	jitter := 0.2 * backoff
	if jitter > 0 {
		backoff += mathrand.Float64() * jitter
	}

	return time.Duration(backoff)
}

// generateID generates a random ID string
func generateID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("id_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("id_%x", b)
}
