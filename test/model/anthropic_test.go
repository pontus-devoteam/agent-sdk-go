package model_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/model"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/model/providers/anthropic"
	"github.com/stretchr/testify/assert"
)

func TestAnthropicProvider(t *testing.T) {
	t.Run("NewProvider", func(t *testing.T) {
		provider := anthropic.NewProvider("test-key")
		assert.NotNil(t, provider)
		assert.Equal(t, "test-key", provider.APIKey)
		assert.Equal(t, anthropic.DefaultMaxRetries, provider.MaxRetries)
	})

	t.Run("WithAPIKey", func(t *testing.T) {
		provider := anthropic.NewProvider("initial-key")
		provider = provider.WithAPIKey("new-key")
		assert.Equal(t, "new-key", provider.APIKey)
	})

	t.Run("SetBaseURL", func(t *testing.T) {
		provider := anthropic.NewProvider("test-key")
		provider = provider.SetBaseURL("https://test.anthropic.com/v1")
		assert.Equal(t, "https://test.anthropic.com/v1", provider.BaseURL)
	})

	t.Run("WithDefaultModel", func(t *testing.T) {
		provider := anthropic.NewProvider("test-key")
		provider = provider.WithDefaultModel("claude-3-haiku")
		assert.Equal(t, "claude-3-haiku", provider.DefaultModel)
	})

	t.Run("GetModel", func(t *testing.T) {
		provider := anthropic.NewProvider("test-key")
		provider.WithDefaultModel("claude-3-haiku")

		anthropicModel, err := provider.GetModel("claude-3-opus")
		assert.NoError(t, err)
		assert.NotNil(t, anthropicModel)
		assert.Equal(t, "claude-3-opus", anthropicModel.(*anthropic.Model).ModelName)

		// Test with default model
		anthropicModel, err = provider.GetModel("")
		assert.NoError(t, err)
		assert.NotNil(t, anthropicModel)
		assert.Equal(t, "claude-3-haiku", anthropicModel.(*anthropic.Model).ModelName)
	})
}

func TestAnthropicModel(t *testing.T) {
	t.Run("GetResponse_Success", func(t *testing.T) {
		// Create a test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request
			assert.Equal(t, "/messages", r.URL.Path)
			assert.Equal(t, "test-key", r.Header.Get("X-Api-Key"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			// Return a mock response
			response := map[string]interface{}{
				"id":   "msg_test",
				"type": "message",
				"role": "assistant",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": "Test response",
					},
				},
				"model":         "claude-3-haiku",
				"stop_reason":   "end_turn",
				"stop_sequence": nil,
				"usage": map[string]interface{}{
					"input_tokens":  10,
					"output_tokens": 5,
				},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		// Create provider and model
		provider := anthropic.NewProvider("test-key")
		provider.SetBaseURL(server.URL)
		anthropicModel, err := provider.GetModel("claude-3-haiku")
		assert.NoError(t, err)

		// Test request
		request := &model.Request{
			Input:              "Test input",
			SystemInstructions: "Test system instructions",
		}

		response, err := anthropicModel.GetResponse(context.Background(), request)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "Test response", response.Content)
		assert.Equal(t, 15, response.Usage.TotalTokens) // sum of input_tokens and output_tokens
	})

	t.Run("GetResponse_WithHandoff", func(t *testing.T) {
		// Create a test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Return a mock response with a handoff
			response := map[string]interface{}{
				"id":   "msg_test",
				"type": "message",
				"role": "assistant",
				"content": []map[string]interface{}{
					{
						"type": "tool_use",
						"id":   "tool_use_123",
						"name": "handoff_to_agent_b",
						"input": map[string]interface{}{
							"input":            "Process this data",
							"return_to_agent":  "agent_a",
							"task_id":          "task_123",
							"is_task_complete": false,
						},
					},
				},
				"model":         "claude-3-haiku",
				"stop_reason":   "tool_use",
				"stop_sequence": nil,
				"usage": map[string]interface{}{
					"input_tokens":  10,
					"output_tokens": 5,
				},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		// Create provider and model
		provider := anthropic.NewProvider("test-key")
		provider.SetBaseURL(server.URL)
		anthropicModel, err := provider.GetModel("claude-3-haiku")
		assert.NoError(t, err)

		// Create a test handoff tool
		handoffTool := map[string]interface{}{
			"name":        "handoff_to_agent_b",
			"description": "Handoff to Agent B",
			"input_schema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"input": map[string]interface{}{
						"type": "string",
					},
					"return_to_agent": map[string]interface{}{
						"type": "string",
					},
					"task_id": map[string]interface{}{
						"type": "string",
					},
					"is_task_complete": map[string]interface{}{
						"type": "boolean",
					},
				},
			},
		}

		// Test request with handoff tool
		request := &model.Request{
			Input: "Test input requiring Agent B",
			Tools: []interface{}{handoffTool},
		}

		response, err := anthropicModel.GetResponse(context.Background(), request)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.HandoffCall)
		assert.Equal(t, "agent_b", response.HandoffCall.AgentName)
		assert.Equal(t, "Process this data", response.HandoffCall.Parameters["input"])
		assert.Equal(t, "agent_a", response.HandoffCall.ReturnToAgent)
		assert.Equal(t, "task_123", response.HandoffCall.TaskID)
		assert.Equal(t, false, response.HandoffCall.IsTaskComplete)
	})

	t.Run("GetResponse_Error", func(t *testing.T) {
		// Create a test server that returns an error
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"type":    "invalid_request_error",
					"message": "Test error message",
				},
			})
		}))
		defer server.Close()

		// Create provider and model
		provider := anthropic.NewProvider("test-key")
		provider.SetBaseURL(server.URL)
		anthropicModel, err := provider.GetModel("claude-3-haiku")
		assert.NoError(t, err)

		// Test request
		request := &model.Request{
			Input: "Test input",
		}

		_, err = anthropicModel.GetResponse(context.Background(), request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Test error message")
	})
}
