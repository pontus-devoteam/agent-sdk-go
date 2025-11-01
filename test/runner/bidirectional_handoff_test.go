package runner_test

import (
	"testing"

	"github.com/Muhammadhamd/agent-sdk-go/pkg/agent"
	"github.com/Muhammadhamd/agent-sdk-go/pkg/model"
	"github.com/stretchr/testify/assert"
)

// TestBidirectionalHandoffFields tests the HandoffCall with bidirectional fields
func TestBidirectionalHandoffFields(t *testing.T) {
	// Create a HandoffCall with bidirectional flow fields
	handoffCall := &model.HandoffCall{
		AgentName:      "WorkerAgent",
		Parameters:     map[string]any{"input": "Please process this data"},
		Type:           model.HandoffTypeDelegate,
		ReturnToAgent:  "OrchestratorAgent",
		TaskID:         "task-123",
		IsTaskComplete: false,
	}

	// Verify the fields are set correctly
	assert.Equal(t, "WorkerAgent", handoffCall.AgentName)
	assert.Equal(t, "Please process this data", handoffCall.Parameters["input"])
	assert.Equal(t, model.HandoffTypeDelegate, handoffCall.Type)
	assert.Equal(t, "OrchestratorAgent", handoffCall.ReturnToAgent)
	assert.Equal(t, "task-123", handoffCall.TaskID)
	assert.Equal(t, false, handoffCall.IsTaskComplete)

	// Test a return handoff
	returnHandoff := &model.HandoffCall{
		AgentName:      "return_to_delegator",
		Parameters:     map[string]any{"input": "Processed data result"},
		Type:           model.HandoffTypeReturn,
		TaskID:         "task-123",
		IsTaskComplete: true,
	}

	// Verify return handoff fields
	assert.Equal(t, "return_to_delegator", returnHandoff.AgentName)
	assert.Equal(t, "Processed data result", returnHandoff.Parameters["input"])
	assert.Equal(t, model.HandoffTypeReturn, returnHandoff.Type)
	assert.Equal(t, "task-123", returnHandoff.TaskID)
	assert.Equal(t, true, returnHandoff.IsTaskComplete)
}

// TestBidirectionalAgentSetup tests the setup of bidirectional agent relationships
func TestBidirectionalAgentSetup(t *testing.T) {
	// Create test agents
	orchestratorAgent := agent.NewAgent("Orchestrator", "I am the orchestrator agent.")
	workerAgent := agent.NewAgent("Worker", "I am a worker agent.")

	// Setup for bidirectional handoff
	orchestratorAgent.WithHandoffs(workerAgent)
	workerAgent.WithHandoffs(orchestratorAgent)

	// Verify the handoffs are set correctly
	assert.Len(t, orchestratorAgent.Handoffs, 1)
	assert.Equal(t, "Worker", orchestratorAgent.Handoffs[0].Name)

	assert.Len(t, workerAgent.Handoffs, 1)
	assert.Equal(t, "Orchestrator", workerAgent.Handoffs[0].Name)
}
