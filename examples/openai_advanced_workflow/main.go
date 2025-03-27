package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/model/providers/openai"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/runner"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/tool"
)

// Sample code to analyze
const sampleCode = `
func processItems(items []string) []string {
	results := []string{}
	for i := 0; i < len(items); i++ {
		item := items[i]
		// Process the item
		processed := item + "_processed"
		results = append(results, processed)
	}
	return results
}
`

func main() {
	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Create OpenAI provider
	provider := openai.NewProvider(apiKey)

	// Choose a model based on what's available
	modelName := "gpt-4o-mini" // Can be changed to gpt-4, gpt-3.5-turbo, etc.
	provider.SetDefaultModel(modelName)

	// Configure rate limits
	provider.WithRateLimit(60, 150000) // 60 requests/min, 150K tokens/min
	provider.WithRetryConfig(3, time.Second*5)

	fmt.Println("Provider configured with:")
	fmt.Printf("- Model: %s\n", provider.DefaultModel)
	fmt.Printf("- Rate limit: %d requests/min, %d tokens/min\n", provider.RPM, provider.TPM)
	fmt.Printf("- Max retries: %d\n", provider.MaxRetries)
	fmt.Println()

	// Create shared tools for all agents
	getWorkflowStateInfo := tool.NewFunctionTool(
		"get_workflow_state",
		"Get information about the current state of the workflow",
		func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"workflow_id":       "code-review-1234",
				"current_phase":     currentPhase,
				"completed_phases":  completedPhases,
				"remaining_phases":  []string{"analyze", "optimize", "test", "document", "summarize"}[len(completedPhases):],
				"timestamp":         time.Now().Format(time.RFC3339),
				"code_under_review": sampleCode,
			}, nil
		},
	)

	updateWorkflowState := tool.NewFunctionTool(
		"update_workflow_phase",
		"Update the current phase of the workflow",
		func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			phase, ok := params["phase"].(string)
			if !ok {
				return nil, fmt.Errorf("phase parameter must be a string")
			}

			// Handle special cases
			if phase == "complete_current" {
				if currentPhase != "" && currentPhase != "complete" {
					completedPhases = append(completedPhases, currentPhase)
					return fmt.Sprintf("Completed phase: %s", currentPhase), nil
				}
				return "No current phase to complete", nil
			}

			// Set the current phase
			currentPhase = phase
			return fmt.Sprintf("Updated current phase to: %s", phase), nil
		},
	).WithSchema(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"phase": map[string]interface{}{
				"type":        "string",
				"description": "The phase to set (e.g., 'analyze', 'optimize', 'test', 'document', 'summarize', 'complete', 'complete_current')",
				"enum":        []string{"analyze", "optimize", "test", "document", "summarize", "complete", "complete_current"},
			},
		},
		"required": []string{"phase"},
	})

	// Create the agents
	orchestratorAgent := createOrchestratorAgent(provider, getWorkflowStateInfo, updateWorkflowState, modelName)
	analyzerAgent := createAnalyzerAgent(provider, getWorkflowStateInfo, modelName)
	optimizerAgent := createOptimizerAgent(provider, getWorkflowStateInfo, modelName)
	testerAgent := createTesterAgent(provider, getWorkflowStateInfo, modelName)
	documentorAgent := createDocumentorAgent(provider, getWorkflowStateInfo, modelName)
	summarizerAgent := createSummarizerAgent(provider, getWorkflowStateInfo, modelName)

	// Set up handoffs for the orchestrator
	orchestratorAgent.WithHandoffs(
		analyzerAgent,
		optimizerAgent,
		testerAgent,
		documentorAgent,
		summarizerAgent,
	)

	// Create runner
	r := runner.NewRunner()
	r.WithDefaultProvider(provider)

	// Run the workflow with the orchestrator agent
	fmt.Println("Starting the code review workflow...")
	result, err := r.RunSync(orchestratorAgent, &runner.RunOptions{
		Input:    "Please review and improve this code:\n\n" + sampleCode,
		MaxTurns: 25, // Set a higher limit for complex workflows
	})

	if err != nil {
		log.Fatalf("Error running agent: %v", err)
	}

	fmt.Println("\nWorkflow complete! Final report:")
	fmt.Println(result.FinalOutput)

	fmt.Println("\nWorkflow phases completed:", completedPhases)
}

// Track workflow state
var currentPhase string
var completedPhases []string

// Create the orchestrator agent
func createOrchestratorAgent(provider *openai.Provider, getWorkflowStateInfo, updateWorkflowState tool.Tool, modelName string) *agent.Agent {
	orchestratorAgent := agent.NewAgent("Orchestrator")
	orchestratorAgent.SetModelProvider(provider)
	orchestratorAgent.WithModel(modelName)
	orchestratorAgent.WithTools(getWorkflowStateInfo, updateWorkflowState)
	orchestratorAgent.SetSystemInstructions(`You are an orchestrator agent that coordinates a code review workflow. 
Your job is to manage the workflow by delegating tasks to specialized agents in the correct sequence.

WORKFLOW PROCESS:
1. First, hand off to the Analyzer to analyze the code and identify issues
2. Based on the Analyzer's response, hand off to the Optimizer to improve the code
3. Once optimized, hand off to the Tester to create tests for the code
4. Then hand off to the Documentor to add documentation
5. Finally, hand off to the Summarizer to create a summary report of all improvements

RULES:
1. Start by getting the workflow state with get_workflow_state
2. For each phase:
   a. Update the workflow phase with update_workflow_phase (use the phase name)
   b. Hand off to the appropriate agent
   c. When the agent responds, mark the phase complete with update_workflow_phase phase="complete_current"
   d. Decide the next phase based on the agent's response

3. You MUST follow this exact pattern for each agent handoff:
   - Call update_workflow_phase with the current phase name
   - Call the handoff_to_[Agent] tool
   - When the agent responds, call update_workflow_phase with phase="complete_current"
   - THEN decide the next agent based on the response

4. After the Summarizer completes, call update_workflow_phase phase="complete" to finalize the workflow

IMPORTANT: Always use the handoff_to_[Agent] tools to delegate work. Never try to solve problems yourself.`)

	return orchestratorAgent
}

// Create the analyzer agent
func createAnalyzerAgent(provider *openai.Provider, getWorkflowStateInfo tool.Tool, modelName string) *agent.Agent {
	analyzerAgent := agent.NewAgent("Analyzer")
	analyzerAgent.SetModelProvider(provider)
	analyzerAgent.WithModel(modelName)
	analyzerAgent.WithTools(getWorkflowStateInfo)
	analyzerAgent.SetSystemInstructions(`You are a code analyzer that specializes in identifying issues, bugs, and areas for improvement in code.

Your job is to:
1. Analyze the code thoroughly
2. Identify potential bugs, inefficiencies, and areas for improvement
3. Provide a detailed analysis with specific issues and recommendations

Be precise and technical in your analysis. Include details about:
- Time and space complexity issues
- Potential edge cases not handled
- Code style and best practices violations
- Performance optimizations
- Any other issues you identify

When you receive code, first get the workflow state with get_workflow_state to see the code in context.
Then provide a comprehensive analysis with clear, actionable feedback.`)

	return analyzerAgent
}

// Create the optimizer agent
func createOptimizerAgent(provider *openai.Provider, getWorkflowStateInfo tool.Tool, modelName string) *agent.Agent {
	optimizerAgent := agent.NewAgent("Optimizer")
	optimizerAgent.SetModelProvider(provider)
	optimizerAgent.WithModel(modelName)
	optimizerAgent.WithTools(getWorkflowStateInfo)
	optimizerAgent.SetSystemInstructions(`You are a code optimizer that specializes in improving code quality, performance, and readability.

Your job is to:
1. Take the code and the analysis from the previous agent
2. Implement the suggested improvements
3. Optimize for performance, readability, and maintainability
4. Provide an improved version of the code

When optimizing code:
- Focus on both algorithmic optimizations (time/space complexity) and code quality
- Use idiomatic patterns for the language
- Ensure proper error handling and edge case coverage
- Maintain or improve readability

When you receive a request, first get the workflow state with get_workflow_state to see the code and context.
Then provide the optimized code along with explanations of your changes.`)

	return optimizerAgent
}

// Create the tester agent
func createTesterAgent(provider *openai.Provider, getWorkflowStateInfo tool.Tool, modelName string) *agent.Agent {
	testerAgent := agent.NewAgent("Tester")
	testerAgent.SetModelProvider(provider)
	testerAgent.WithModel(modelName)
	testerAgent.WithTools(getWorkflowStateInfo)
	testerAgent.SetSystemInstructions(`You are a testing expert that specializes in creating comprehensive test suites for code.

Your job is to:
1. Examine the optimized code
2. Design test cases that cover functionality, edge cases, and error conditions
3. Write actual test code that could be run to verify the code works correctly
4. Ensure high code coverage with your tests

Focus on:
- Unit tests for individual functions
- Edge case testing
- Error condition handling
- Performance testing if applicable

When you receive a request, first get the workflow state with get_workflow_state to see the current code.
Then provide complete test code that would properly test the functionality.`)

	return testerAgent
}

// Create the documentor agent
func createDocumentorAgent(provider *openai.Provider, getWorkflowStateInfo tool.Tool, modelName string) *agent.Agent {
	documentorAgent := agent.NewAgent("Documentor")
	documentorAgent.SetModelProvider(provider)
	documentorAgent.WithModel(modelName)
	documentorAgent.WithTools(getWorkflowStateInfo)
	documentorAgent.SetSystemInstructions(`You are a documentation expert that specializes in creating clear, comprehensive documentation for code.

Your job is to:
1. Examine the optimized code and tests
2. Create proper documentation including:
   - Function/method documentation
   - Parameter and return value descriptions
   - Usage examples
   - Any important notes about edge cases or limitations

Good documentation should:
- Be clear and concise
- Explain the "why" not just the "what"
- Include examples for complex functionality
- Follow documentation standards for the language

When you receive a request, first get the workflow state with get_workflow_state to see the current code.
Then provide complete documentation for the code including properly formatted comments and any external documentation.`)

	return documentorAgent
}

// Create the summarizer agent
func createSummarizerAgent(provider *openai.Provider, getWorkflowStateInfo tool.Tool, modelName string) *agent.Agent {
	summarizerAgent := agent.NewAgent("Summarizer")
	summarizerAgent.SetModelProvider(provider)
	summarizerAgent.WithModel(modelName)
	summarizerAgent.WithTools(getWorkflowStateInfo)
	summarizerAgent.SetSystemInstructions(`You are a summarizer that creates comprehensive reports of code improvement processes.

Your job is to:
1. Review all the work done by previous agents (analysis, optimization, testing, documentation)
2. Create a clear, well-organized summary of:
   - Initial issues identified
   - Changes made to the code
   - Testing approach and coverage
   - Documentation provided
   - Overall improvements in code quality, performance, and maintainability

Your summary should:
- Be professional and well-structured
- Highlight the most important improvements
- Quantify improvements where possible (e.g., "reduced time complexity from O(n²) to O(n)")
- Include both technical details and business value

When you receive a request, first get the workflow state with get_workflow_state to see all the context.
Then provide a comprehensive but concise summary report of the entire code improvement process.`)

	return summarizerAgent
}
