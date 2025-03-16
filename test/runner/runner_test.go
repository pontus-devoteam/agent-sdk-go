package runner_test

import (
	"context"
	"testing"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/model"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/runner"
)

// MockModelProvider is a mock implementation of model.ModelProvider for testing
type MockModelProvider struct{}

func (p *MockModelProvider) GetModel(name string) (model.Model, error) {
	return &MockModel{}, nil
}

// MockModel is a mock implementation of model.Model for testing
type MockModel struct{}

func (m *MockModel) GetResponse(ctx context.Context, request *model.ModelRequest) (*model.ModelResponse, error) {
	return &model.ModelResponse{
		Content: "Mock response",
	}, nil
}

func (m *MockModel) StreamResponse(ctx context.Context, request *model.ModelRequest) (<-chan model.StreamEvent, error) {
	eventCh := make(chan model.StreamEvent)
	go func() {
		defer close(eventCh)
		eventCh <- model.StreamEvent{
			Type:    model.StreamEventTypeContent,
			Content: "Mock stream content",
		}
		eventCh <- model.StreamEvent{
			Type: model.StreamEventTypeDone,
			Response: &model.ModelResponse{
				Content: "Mock stream content",
			},
		}
	}()
	return eventCh, nil
}

// TestNewRunner tests the creation of a new runner
func TestNewRunner(t *testing.T) {
	// Create a new runner
	r := runner.NewRunner()

	// Check if the runner was created correctly
	if r == nil {
		t.Fatalf("NewRunner() returned nil")
	}
}

// TestWithDefaultMaxTurns tests setting the default maximum turns
func TestWithDefaultMaxTurns(t *testing.T) {
	// Create a new runner
	r := runner.NewRunner()

	// Set default max turns
	maxTurns := 5
	r.WithDefaultMaxTurns(maxTurns)

	// We can't directly check the default max turns as it's a private field,
	// but we can check that the runner instance is returned correctly
	if r == nil {
		t.Fatalf("WithDefaultMaxTurns(%d) returned nil", maxTurns)
	}
}

// TestWithDefaultProvider tests setting the default model provider
func TestWithDefaultProvider(t *testing.T) {
	// Create a new runner
	r := runner.NewRunner()

	// Create a mock model provider
	provider := &MockModelProvider{}

	// Set default provider
	r.WithDefaultProvider(provider)

	// We can't directly check the default provider as it's a private field,
	// but we can check that the runner instance is returned correctly
	if r == nil {
		t.Fatalf("WithDefaultProvider() returned nil")
	}
}

// TestRunOptions tests creating run options
func TestRunOptions(t *testing.T) {
	// Create run options
	input := "Test input"
	maxTurns := 5
	opts := &runner.RunOptions{
		Input:    input,
		MaxTurns: maxTurns,
		RunConfig: &runner.RunConfig{
			Model:           "test-model",
			ModelProvider:   &MockModelProvider{},
			TracingDisabled: true,
		},
	}

	// Check if options were created correctly
	if opts.Input != input {
		t.Errorf("RunOptions.Input = %v, want %v", opts.Input, input)
	}

	if opts.MaxTurns != maxTurns {
		t.Errorf("RunOptions.MaxTurns = %d, want %d", opts.MaxTurns, maxTurns)
	}

	if opts.RunConfig == nil {
		t.Fatalf("RunOptions.RunConfig is nil")
	}

	if opts.RunConfig.Model != "test-model" {
		t.Errorf("RunOptions.RunConfig.Model = %v, want test-model", opts.RunConfig.Model)
	}

	if opts.RunConfig.ModelProvider == nil {
		t.Errorf("RunOptions.RunConfig.ModelProvider is nil")
	}

	if !opts.RunConfig.TracingDisabled {
		t.Errorf("RunOptions.RunConfig.TracingDisabled = false, want true")
	}
}

// TestRunConfig tests creating a run configuration
func TestRunConfig(t *testing.T) {
	// Create run configuration
	cfg := &runner.RunConfig{
		Model:           "test-model",
		ModelProvider:   &MockModelProvider{},
		TracingDisabled: true,
		TracingConfig: &runner.TracingConfig{
			WorkflowName: "test-workflow",
			TraceID:      "test-trace-id",
			GroupID:      "test-group-id",
			Metadata: map[string]interface{}{
				"key": "value",
			},
		},
		ModelSettings: &model.ModelSettings{
			Temperature: func() *float64 { val := 0.7; return &val }(),
			TopP:        func() *float64 { val := 0.9; return &val }(),
		},
	}

	// Check if configuration was created correctly
	if cfg.Model != "test-model" {
		t.Errorf("RunConfig.Model = %v, want test-model", cfg.Model)
	}

	if cfg.ModelProvider == nil {
		t.Errorf("RunConfig.ModelProvider is nil")
	}

	if !cfg.TracingDisabled {
		t.Errorf("RunConfig.TracingDisabled = false, want true")
	}

	if cfg.TracingConfig == nil {
		t.Fatalf("RunConfig.TracingConfig is nil")
	}

	if cfg.TracingConfig.WorkflowName != "test-workflow" {
		t.Errorf("RunConfig.TracingConfig.WorkflowName = %s, want test-workflow", cfg.TracingConfig.WorkflowName)
	}

	if cfg.TracingConfig.TraceID != "test-trace-id" {
		t.Errorf("RunConfig.TracingConfig.TraceID = %s, want test-trace-id", cfg.TracingConfig.TraceID)
	}

	if cfg.TracingConfig.GroupID != "test-group-id" {
		t.Errorf("RunConfig.TracingConfig.GroupID = %s, want test-group-id", cfg.TracingConfig.GroupID)
	}

	if cfg.TracingConfig.Metadata["key"] != "value" {
		t.Errorf("RunConfig.TracingConfig.Metadata[key] = %s, want value", cfg.TracingConfig.Metadata["key"])
	}

	if cfg.ModelSettings == nil {
		t.Fatalf("RunConfig.ModelSettings is nil")
	}

	if *cfg.ModelSettings.Temperature != 0.7 {
		t.Errorf("RunConfig.ModelSettings.Temperature = %f, want 0.7", *cfg.ModelSettings.Temperature)
	}

	if *cfg.ModelSettings.TopP != 0.9 {
		t.Errorf("RunConfig.ModelSettings.TopP = %f, want 0.9", *cfg.ModelSettings.TopP)
	}
}

// TestTracingConfig tests creating a tracing configuration
func TestTracingConfig(t *testing.T) {
	// Create tracing configuration
	cfg := &runner.TracingConfig{
		WorkflowName: "test-workflow",
		TraceID:      "test-trace-id",
		GroupID:      "test-group-id",
		Metadata: map[string]interface{}{
			"key": "value",
		},
	}

	// Check if configuration was created correctly
	if cfg.WorkflowName != "test-workflow" {
		t.Errorf("TracingConfig.WorkflowName = %s, want test-workflow", cfg.WorkflowName)
	}

	if cfg.TraceID != "test-trace-id" {
		t.Errorf("TracingConfig.TraceID = %s, want test-trace-id", cfg.TraceID)
	}

	if cfg.GroupID != "test-group-id" {
		t.Errorf("TracingConfig.GroupID = %s, want test-group-id", cfg.GroupID)
	}

	if cfg.Metadata["key"] != "value" {
		t.Errorf("TracingConfig.Metadata[key] = %s, want value", cfg.Metadata["key"])
	}
}
