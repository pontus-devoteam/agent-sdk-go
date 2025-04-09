package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/model"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/model/providers/openai"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/result"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/runner"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/tool"
)

// Sample TypeScript function requirement
const functionRequirement = "Create a TypeScript function that filters an array of objects based on multiple criteria, with support for AND/OR logic and nested conditions."

func main() {
	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Create an OpenAI provider
	provider := openai.NewProvider(apiKey)

	// Set GPT model as the default model
	provider.SetDefaultModel("gpt-4") // Using GPT-4 for better code generation

	// Configure retry settings
	provider.WithRetryConfig(3, 2*time.Second)

	fmt.Println("Provider configured with:")
	fmt.Println("- Model:", "gpt-4")
	fmt.Println("- Max retries:", 3)

	// Create tools
	getCurrentTime := tool.NewFunctionTool(
		"get_current_time",
		"Get the current time",
		func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"current_time": time.Now().Format(time.RFC3339),
				"timezone":     time.Now().Location().String(),
			}, nil
		},
	)

	validateTSCode := tool.NewFunctionTool(
		"validate_ts_code",
		"Validates TypeScript code for syntax and style issues",
		func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			code, ok := params["code"].(string)
			if !ok {
				return nil, fmt.Errorf("code parameter is required")
			}

			// Simulate code validation
			time.Sleep(1 * time.Second)

			// For demonstration purposes, we'll simulate finding issues in code that contains certain keywords
			var issues []map[string]interface{}

			// Simulate finding issues (in a real implementation, you would use a linter)
			if len(code) < 50 {
				issues = append(issues, map[string]interface{}{
					"line":     1,
					"severity": "error",
					"message":  "Function implementation is too short and likely incomplete",
				})
			}

			if !strings.Contains(code, "interface") && strings.Contains(code, "object") {
				issues = append(issues, map[string]interface{}{
					"line":     1,
					"severity": "warning",
					"message":  "Consider using TypeScript interfaces for type definitions",
				})
			}

			if !strings.Contains(code, "test") && !strings.Contains(code, "expect") {
				issues = append(issues, map[string]interface{}{
					"line":     1,
					"severity": "info",
					"message":  "No tests found. Consider adding unit tests",
				})
			}

			return map[string]interface{}{
				"valid":     len(issues) == 0,
				"issues":    issues,
				"timestamp": time.Now().Format(time.RFC3339),
			}, nil
		},
	).WithSchema(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"code": map[string]interface{}{
				"type":        "string",
				"description": "The TypeScript code to validate",
			},
		},
		"required": []string{"code"},
	})

	// Create a tool to track code versions
	trackCodeVersion := tool.NewFunctionTool(
		"track_code_version",
		"Tracks code versions and updates throughout the review process",
		func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			code, ok := params["code"].(string)
			if !ok {
				return nil, fmt.Errorf("code parameter is required")
			}

			version, _ := params["version"].(float64)
			if version == 0 {
				version = 1.0
			}

			description, _ := params["description"].(string)
			if description == "" {
				description = "Initial version"
			}

			return map[string]interface{}{
				"version":     version,
				"timestamp":   time.Now().Format(time.RFC3339),
				"code_hash":   fmt.Sprintf("hash-%d", len(code)), // Simple mock hash
				"description": description,
			}, nil
		},
	).WithSchema(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"code": map[string]interface{}{
				"type":        "string",
				"description": "The code to track",
			},
			"version": map[string]interface{}{
				"type":        "number",
				"description": "Version number (incremental)",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Description of this version",
			},
		},
		"required": []string{"code"},
	})

	// Create specialized agents
	coderAgent := createCoderAgent(provider, validateTSCode, getCurrentTime, trackCodeVersion)
	reviewerAgent := createReviewerAgent(provider, validateTSCode, getCurrentTime, trackCodeVersion)

	// Create the orchestrator agent
	orchestratorAgent := agent.NewAgent("Orchestrator")
	orchestratorAgent.SetModelProvider(provider)
	orchestratorAgent.WithModel("gpt-4")
	orchestratorAgent.WithTools(getCurrentTime, trackCodeVersion)

	// Add system instructions and configure as task delegator
	orchestratorAgent.SetSystemInstructions(`You are an orchestrator agent that coordinates the development of TypeScript code.
Your job is to manage the workflow by delegating tasks to specialized agents and processing their returns.

WORKFLOW PROCESS:
1. Delegate to the CoderAgent to implement the requested TypeScript function
2. When the CoderAgent returns with code, delegate to the ReviewerAgent to review the code
3. If the ReviewerAgent suggests changes, delegate back to the CoderAgent with the feedback
4. Repeat steps 2-3 until the ReviewerAgent approves the code
5. Present the final, approved code to the user as your final output

IMPORTANT TASK MANAGEMENT:
- Always provide task IDs when delegating and track which tasks have been completed
- When an agent returns to you, check the task ID to determine the next steps in the workflow
- Maintain context across the development workflow by tracking code versions
- Use the track_code_version tool to keep track of code versions and changes
- After all steps are complete, your final response should include the complete, approved code
- Include any notable aspects of the development process in your final summary

TASK CONTEXT:
- Each task in the workflow builds on previous tasks
- The CoderAgent needs to see the reviewer's feedback for improvements
- The ReviewerAgent needs to see the previous versions of code to track improvements
- Always include relevant context from previous tasks when delegating new tasks`)

	// Configure as task delegator with explicit delegator name
	fmt.Println("Configuring orchestrator as task delegator...")
	orchestratorAgent.AsTaskDelegator()

	// Add handoffs
	orchestratorAgent.WithHandoffs(coderAgent, reviewerAgent)

	// Configure bidirectional handoffs manually instead of using WithBidirectionalHandoffs
	coderAgent.WithHandoffs(orchestratorAgent)
	reviewerAgent.WithHandoffs(orchestratorAgent)

	// Create hooks to track task context information
	taskTrackingHooks := &TaskTrackingHooks{
		WorkContext: make(map[string]interface{}),
	}

	// Create runner with workflow configuration
	r := runner.NewRunner()
	r.WithDefaultProvider(provider)

	// Configure workflow options
	runOpts := &runner.RunOptions{
		Input:    fmt.Sprintf("I need a TypeScript function with the following requirements: %s", functionRequirement),
		MaxTurns: 15, // Allow more turns for the code review cycle
		Hooks:    taskTrackingHooks,
		RunConfig: &runner.RunConfig{
			// Specific model settings if needed
			ModelSettings: &model.Settings{
				Temperature: getFloatPtr(0.7), // More creative for coding tasks
			},
		},
	}

	// Run the workflow
	fmt.Println("\nStarting the TypeScript function development workflow...")

	// Enable debugging
	os.Setenv("DEBUG", "1")
	os.Setenv("OPENAI_DEBUG", "1")

	// Print debug info about the agents
	fmt.Printf("DEBUG: Orchestrator agent has %d handoffs configured\n", len(orchestratorAgent.Handoffs))
	for i, h := range orchestratorAgent.Handoffs {
		fmt.Printf("DEBUG: Handoff #%d: %s\n", i+1, h.Name)
	}

	// Run the workflow
	result, err := r.RunSync(orchestratorAgent, runOpts)

	if err != nil {
		log.Fatalf("Error running agent: %v", err)
	}

	// Print a summary of what happened
	fmt.Println("\nWorkflow complete! Summary:")

	// Handle nil FinalOutput by providing a fallback message
	if result.FinalOutput == nil {
		fmt.Println("- Final output: (No final output generated)")
	} else {
		fmt.Printf("- Final output: %v\n", result.FinalOutput)
	}

	fmt.Printf("- Last agent: %s\n", result.LastAgent.Name)
	fmt.Printf("- Items generated: %d\n", len(result.NewItems))

	// Print details of any handoff items
	fmt.Println("\nHandoffs:")
	handoffCount := 0
	for _, item := range result.NewItems {
		if item.GetType() == "handoff" {
			handoffCount++
			handoffItem, ok := item.(interface{ GetAgentName() string })
			if ok {
				fmt.Printf("- Handoff #%d: %s\n", handoffCount, handoffItem.GetAgentName())
			} else {
				fmt.Printf("- Handoff #%d: (agent name not available)\n", handoffCount)
			}
		}
	}

	if handoffCount == 0 {
		fmt.Println("- No handoffs occurred")
	}

	// Print task context information
	fmt.Println("\nTask Context Summary:")
	if len(taskTrackingHooks.WorkContext) > 0 {
		for key, value := range taskTrackingHooks.WorkContext {
			fmt.Printf("- %s: %v\n", key, value)
		}
	} else {
		fmt.Println("- No task context information available")
	}
}

// Helper function for creating float pointers
func getFloatPtr(val float64) *float64 {
	return &val
}

// TaskTrackingHooks implements RunHooks to track task context information
type TaskTrackingHooks struct {
	WorkContext map[string]interface{}
}

// OnRunStart is called when the run starts
func (h *TaskTrackingHooks) OnRunStart(ctx context.Context, agent *agent.Agent, input interface{}) error {
	h.WorkContext["start_time"] = time.Now().Format(time.RFC3339)
	h.WorkContext["initial_agent"] = agent.Name
	return nil
}

// OnTurnStart is called when a turn starts
func (h *TaskTrackingHooks) OnTurnStart(ctx context.Context, agent *agent.Agent, turn int) error {
	return nil
}

// OnTurnEnd is called when a turn ends
func (h *TaskTrackingHooks) OnTurnEnd(ctx context.Context, agent *agent.Agent, turn int, result *runner.SingleTurnResult) error {
	// Track agent activity
	turnKey := fmt.Sprintf("turn_%d", turn)
	h.WorkContext[turnKey] = map[string]interface{}{
		"agent":    agent.Name,
		"has_tool": result.Response != nil,
		"time":     time.Now().Format(time.RFC3339),
	}
	return nil
}

// OnRunEnd is called when the run ends
func (h *TaskTrackingHooks) OnRunEnd(ctx context.Context, result *result.RunResult) error {
	h.WorkContext["end_time"] = time.Now().Format(time.RFC3339)
	h.WorkContext["last_agent"] = result.LastAgent.Name
	return nil
}

// OnBeforeHandoff is called before a handoff occurs
func (h *TaskTrackingHooks) OnBeforeHandoff(ctx context.Context, sourceAgent *agent.Agent, targetAgent *agent.Agent) error {
	handoffKey := fmt.Sprintf("handoff_%s_to_%s", sourceAgent.Name, targetAgent.Name)
	h.WorkContext[handoffKey] = time.Now().Format(time.RFC3339)
	return nil
}

// OnAfterHandoff is called after a handoff completes
func (h *TaskTrackingHooks) OnAfterHandoff(ctx context.Context, sourceAgent *agent.Agent, targetAgent *agent.Agent, result interface{}) error {
	handoffKey := fmt.Sprintf("handoff_%s_from_%s_complete", sourceAgent.Name, targetAgent.Name)
	h.WorkContext[handoffKey] = time.Now().Format(time.RFC3339)
	return nil
}

// Create the coder agent
func createCoderAgent(provider *openai.Provider, validateTool, timeTool, trackCodeTool tool.Tool) *agent.Agent {
	coderAgent := agent.NewAgent("CoderAgent")
	coderAgent.SetModelProvider(provider)
	coderAgent.WithModel("gpt-4")
	coderAgent.WithTools(validateTool, timeTool, trackCodeTool)

	// Add system instructions and configure as task executor
	coderAgent.SetSystemInstructions(`You are a TypeScript coding agent that specializes in writing high-quality TypeScript code.

Your job is to:
1. Implement TypeScript functions based on requirements provided to you
2. Write clean, efficient, and well-documented code
3. Add proper type definitions and interfaces
4. Include unit tests for your implementation
5. Address feedback from code reviews
6. Return your implemented code to the agent that delegated the task to you

When writing TypeScript code:
- Always use proper TypeScript features (interfaces, types, generics where appropriate)
- Include detailed JSDoc comments for functions and parameters
- Follow best practices for error handling
- Write modular and reusable code
- Include example usage in comments
- Write unit tests that cover major use cases

If you receive feedback from a reviewer:
- Carefully address each point of feedback
- Explain what changes you made in response to the feedback
- Use the validate_ts_code tool to check your implementation
- Use the track_code_version tool to record your code versions

TASK CONTEXT:
- Maintain awareness of the current task context
- Review any provided context information about previous work
- When you receive code to revise, carefully examine both the code and the feedback
- Make sure to track code versions using the track_code_version tool

When you complete your task, return to the Orchestrator by calling handoff to "Orchestrator" with your implemented code as input.`)
	coderAgent.AsTaskExecutor()

	return coderAgent
}

// Create the reviewer agent
func createReviewerAgent(provider *openai.Provider, validateTool, timeTool, trackCodeTool tool.Tool) *agent.Agent {
	reviewerAgent := agent.NewAgent("ReviewerAgent")
	reviewerAgent.SetModelProvider(provider)
	reviewerAgent.WithModel("gpt-4")
	reviewerAgent.WithTools(validateTool, timeTool, trackCodeTool)

	// Add system instructions and configure as task executor
	reviewerAgent.SetSystemInstructions(`You are a code review agent that specializes in reviewing TypeScript code.

Your job is to:
1. Review TypeScript code for quality, correctness, and adherence to best practices
2. Identify potential bugs, edge cases, and performance issues
3. Check type definitions and ensure proper TypeScript features are used
4. Evaluate code structure, readability, and maintainability
5. Provide constructive feedback and specific suggestions for improvement
6. Return your review results to the agent that delegated the task to you

When reviewing code:
- Use the validate_ts_code tool to check for syntax and style issues
- Check for proper error handling and edge cases
- Verify that type definitions are complete and accurate
- Look for opportunities to simplify or optimize the code
- Ensure unit tests cover the main functionality
- Provide specific, actionable feedback
- Use the track_code_version tool to record the version of code you're reviewing

TASK CONTEXT:
- Maintain awareness of the current task context
- Check for previous versions of the code to understand the progress
- Compare the current version against any previous feedback you provided
- When reviewing revised code, check if previous issues were addressed

Your review should include:
1. Overall assessment (Approved/Needs Changes)
2. Specific issues with line references
3. Suggestions for improvement
4. Positive aspects of the code

When you complete your task, return to the Orchestrator by calling handoff to "Orchestrator" with your review results as input.`)
	reviewerAgent.AsTaskExecutor()

	return reviewerAgent
}
