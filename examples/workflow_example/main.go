package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/model/providers/lmstudio"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/runner"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/tool"
)

// WorkflowContext maintains state between agent handoffs
type WorkflowContext struct {
	CurrentPhase    string                 // Current phase of development
	CompletedPhases []string               // List of completed phases
	Artifacts       map[string]string      // Store artifacts (design docs, code, tests, etc.)
	Metadata        map[string]interface{} // Additional metadata
}

// InMemoryStateStore is a simple in-memory implementation of WorkflowStateStore
type InMemoryStateStore struct {
	states map[string]interface{}
}

func NewInMemoryStateStore() *InMemoryStateStore {
	return &InMemoryStateStore{
		states: make(map[string]interface{}),
	}
}

func (s *InMemoryStateStore) SaveState(workflowID string, state interface{}) error {
	s.states[workflowID] = state
	return nil
}

func (s *InMemoryStateStore) LoadState(workflowID string) (interface{}, error) {
	if state, ok := s.states[workflowID]; ok {
		return state, nil
	}
	return nil, fmt.Errorf("state not found for workflow: %s", workflowID)
}

func (s *InMemoryStateStore) ListCheckpoints(workflowID string) ([]string, error) {
	return []string{}, nil // Not implemented for this example
}

func (s *InMemoryStateStore) DeleteCheckpoint(workflowID string, checkpointID string) error {
	return nil // Not implemented for this example
}

// Create specialized agents
func createDesignAgent(provider *lmstudio.Provider) *agent.Agent {
	agent := agent.NewAgent("Design Agent")
	agent.SetModelProvider(provider)
	agent.WithModel("gemma-3-4b-it")
	agent.SetSystemInstructions("You are a software design specialist. Create system designs and architecture documents.")
	return agent
}

func createCodeAgent(provider *lmstudio.Provider) *agent.Agent {
	agent := agent.NewAgent("Code Agent")
	agent.SetModelProvider(provider)
	agent.WithModel("gemma-3-4b-it")
	agent.SetSystemInstructions("You are a software implementation specialist. Write clean, efficient code based on designs.")
	return agent
}

func createTestAgent(provider *lmstudio.Provider) *agent.Agent {
	agent := agent.NewAgent("Test Agent")
	agent.SetModelProvider(provider)
	agent.WithModel("gemma-3-4b-it")
	agent.SetSystemInstructions("You are a testing specialist. Write comprehensive tests for the code.")
	return agent
}

func createReviewAgent(provider *lmstudio.Provider) *agent.Agent {
	agent := agent.NewAgent("Review Agent")
	agent.SetModelProvider(provider)
	agent.WithModel("gemma-3-4b-it")
	agent.SetSystemInstructions("You are a code review specialist. Review code for quality, security, and best practices.")
	return agent
}

func createFixAgent(provider *lmstudio.Provider) *agent.Agent {
	agent := agent.NewAgent("Fix Agent")
	agent.SetModelProvider(provider)
	agent.WithModel("gemma-3-4b-it")
	agent.SetSystemInstructions("You are a bug fixing specialist. Fix issues identified during code review.")
	return agent
}

func main() {
	// Enable verbose logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Create a model provider
	provider := lmstudio.NewProvider()
	provider.SetBaseURL("http://127.0.0.1:1234/v1")
	provider.SetDefaultModel("gemma-3-4b-it")

	fmt.Println("Provider configured with:")
	fmt.Println("- Base URL:", "http://127.0.0.1:1234/v1")
	fmt.Println("- Model:", "gemma-3-4b-it")

	// Initialize workflow context
	workflowCtx := &WorkflowContext{
		CurrentPhase:    "init",
		CompletedPhases: make([]string, 0),
		Artifacts:       make(map[string]string),
		Metadata:        make(map[string]interface{}),
	}

	// Create specialized agents
	designAgent := createDesignAgent(provider)
	codeAgent := createCodeAgent(provider)
	testAgent := createTestAgent(provider)
	reviewAgent := createReviewAgent(provider)
	fixAgent := createFixAgent(provider)

	// Create the orchestrator agent
	orchestratorAgent := agent.NewAgent("Orchestrator")
	orchestratorAgent.SetModelProvider(provider)
	orchestratorAgent.WithModel("gemma-3-4b-it")

	// Set up orchestrator's system instructions
	orchestratorAgent.SetSystemInstructions(`You are a software engineering workflow orchestrator.
Your job is to coordinate complex software development tasks by delegating to specialized agents.

WORKFLOW MANAGEMENT:
1. Analyze the user's request and break it down into steps
2. For each step, determine the appropriate specialized agent:
   - Design Agent: For system design and architecture
   - Code Agent: For implementation
   - Test Agent: For writing tests
   - Review Agent: For code review
   - Fix Agent: For bug fixes

3. IMPORTANT: After each handoff, you must:
   - Analyze the agent's response
   - Determine if the result is satisfactory
   - Decide the next step in the workflow
   - Either handoff to another agent or return final results

4. Maintain context between handoffs:
   - Keep track of previous agent outputs
   - Use them to inform next agent's instructions
   - Ensure continuity in the development process

WORKFLOW PHASES:
1. DESIGN:
   - Analyze requirements
   - Create system design
   - Define interfaces and data structures

2. IMPLEMENTATION:
   - Write code based on design
   - Follow best practices
   - Implement error handling

3. TESTING:
   - Write unit tests
   - Write integration tests
   - Ensure code coverage

4. REVIEW:
   - Code review
   - Identify issues
   - Suggest improvements

5. FIX (if needed):
   - Address review comments
   - Fix bugs
   - Improve code quality

For each phase:
1. Use the update_workflow_context tool to track progress
2. Include relevant context when delegating to specialized agents
3. Verify the output before moving to the next phase

Remember:
- Always maintain workflow state
- Make informed decisions about next steps
- Ensure quality at each phase
- Keep the user informed of progress`)

	// Create workflow management tools
	updateContextTool := tool.NewFunctionTool(
		"update_workflow_context",
		"Update the workflow context with new information",
		func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			// Update phase if provided
			if phase, ok := params["phase"].(string); ok {
				if workflowCtx.CurrentPhase != phase {
					workflowCtx.CompletedPhases = append(workflowCtx.CompletedPhases, workflowCtx.CurrentPhase)
					workflowCtx.CurrentPhase = phase
				}
			}

			// Update artifacts if provided
			if artifacts, ok := params["artifacts"].(map[string]interface{}); ok {
				for k, v := range artifacts {
					if strVal, ok := v.(string); ok {
						workflowCtx.Artifacts[k] = strVal
					}
				}
			}

			// Update metadata if provided
			if metadata, ok := params["metadata"].(map[string]interface{}); ok {
				for k, v := range metadata {
					workflowCtx.Metadata[k] = v
				}
			}

			return map[string]interface{}{
				"current_phase":     workflowCtx.CurrentPhase,
				"completed_phases":  workflowCtx.CompletedPhases,
				"artifact_count":    len(workflowCtx.Artifacts),
				"metadata_count":    len(workflowCtx.Metadata),
				"workflow_progress": fmt.Sprintf("Phase %s in progress", workflowCtx.CurrentPhase),
			}, nil
		},
	).WithSchema(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"phase": map[string]interface{}{
				"type":        "string",
				"description": "The current phase of the workflow",
				"enum":        []string{"init", "design", "implementation", "testing", "review", "fix", "complete"},
			},
			"artifacts": map[string]interface{}{
				"type":        "object",
				"description": "Artifacts produced in this phase",
				"additionalProperties": map[string]interface{}{
					"type": "string",
				},
			},
			"metadata": map[string]interface{}{
				"type":        "object",
				"description": "Additional metadata for the workflow",
				"additionalProperties": map[string]interface{}{
					"type": "string",
				},
			},
		},
	})

	// Add tools to orchestrator
	orchestratorAgent.WithTools(updateContextTool)

	// Set up agent handoffs
	orchestratorAgent.WithHandoffs(designAgent, codeAgent, testAgent, reviewAgent, fixAgent)

	// Create a base runner
	baseRunner := runner.NewRunner()
	baseRunner.WithDefaultProvider(provider)

	// Create workflow configuration
	workflowConfig := &runner.WorkflowConfig{
		RetryConfig: &runner.RetryConfig{
			MaxRetries:         3,
			RetryDelay:         time.Second,
			RetryBackoffFactor: 2.0,
			RetryableErrors:    []string{"handoff failed", "tool call failed"},
			OnRetry: func(attempt int, err error) error {
				fmt.Printf("Retrying workflow (attempt %d) after error: %v\n", attempt, err)
				return nil
			},
		},
		StateManagement: &runner.StateManagementConfig{
			PersistState:        true,
			StateStore:          NewInMemoryStateStore(),
			CheckpointFrequency: 30 * time.Second,
			RestoreOnFailure:    true,
		},
		ValidationConfig: &runner.ValidationConfig{
			PreHandoffValidation: []runner.ValidationRule{
				{
					Name: "RequiredArtifacts",
					Validate: func(data interface{}) (bool, error) {
						if state, ok := data.(*runner.WorkflowState); ok {
							if len(state.Artifacts) == 0 {
								return false, fmt.Errorf("no artifacts available for handoff")
							}
							return true, nil
						}
						return false, fmt.Errorf("invalid state type")
					},
					ErrorMessage: "Missing required artifacts for handoff",
					Severity:     runner.ValidationWarning,
				},
			},
		},
	}

	// Create workflow runner
	workflowRunner := runner.NewWorkflowRunner(baseRunner, workflowConfig)

	// Initialize workflow state
	state := &runner.WorkflowState{
		CurrentPhase:    "init",
		CompletedPhases: make([]string, 0),
		Artifacts:       make(map[string]interface{}),
		LastCheckpoint:  time.Now(),
		Metadata:        make(map[string]interface{}),
	}

	// Save initial state
	err := NewInMemoryStateStore().SaveState("default", state)
	if err != nil {
		log.Fatalf("Failed to save initial state: %v", err)
	}

	// Run the workflow
	result, err := workflowRunner.RunWorkflow(context.Background(), orchestratorAgent, &runner.RunOptions{
		MaxTurns:       20,
		RunConfig:      &runner.RunConfig{},
		WorkflowConfig: workflowConfig,
		Input:          "Create a simple REST API endpoint for user registration with email and password",
	})

	if err != nil {
		log.Fatalf("Workflow failed: %v", err)
	}

	fmt.Println("\nWorkflow completed successfully!")
	fmt.Println("Final output:", result.FinalOutput)
}
