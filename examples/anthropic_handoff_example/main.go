package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/model/providers/anthropic"
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
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is required")
	}

	// Create an Anthropic provider
	provider := anthropic.NewProvider(apiKey)

	// Set Claude model as the default model
	provider.SetDefaultModel("claude-3-haiku-20240307")

	// Configure rate limits
	provider.WithRateLimit(40, 80000) // 40 requests per minute, 80,000 tokens per minute

	// Configure retry settings
	provider.WithRetryConfig(3, 2*time.Second)

	fmt.Println("Provider configured with:")
	fmt.Println("- Model:", "claude-3-haiku-20240307")
	fmt.Println("- Rate limit:", "40 requests/min, 80,000 tokens/min")
	fmt.Println("- Max retries:", 3)
	fmt.Println("- Working directory:", "./code-review")

	// Create a simple code analysis tool
	getCodeInfo := tool.NewFunctionTool(
		"get_code_info",
		"Get information about the code being analyzed",
		func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"code":          sampleCode,
				"language":      "Go",
				"file_name":     "process.go",
				"function_name": "processItems",
				"timestamp":     time.Now().Format(time.RFC3339),
			}, nil
		},
	)

	// Create the specialized agents
	analyzerAgent := createAnalyzerAgent(provider, getCodeInfo)
	optimizerAgent := createOptimizerAgent(provider, getCodeInfo)

	// Create the orchestrator agent
	orchestratorAgent := agent.NewAgent("Orchestrator")
	orchestratorAgent.SetModelProvider(provider)
	orchestratorAgent.WithModel("claude-3-haiku-20240307")
	orchestratorAgent.WithTools(getCodeInfo)
	orchestratorAgent.SetSystemInstructions(`You are an orchestrator agent that coordinates a code review workflow.
Your job is to manage the workflow by delegating tasks to specialized agents.

WORKFLOW PROCESS:
1. First, hand off to the Analyzer to analyze the code and identify issues
2. Then, based on the analysis, hand off to the Optimizer to improve the code
3. Finally, present the results to the user

When receiving a request to review code:
1. First use the get_code_info tool to retrieve information about the code
2. Then hand off to the Analyzer agent to analyze the code
3. After receiving the analysis, hand off to the Optimizer to implement improvements
4. Summarize the process and present the final optimized code to the user

Always use the handoff_to_[Agent] tools to delegate tasks to the specialized agents.`)

	// Set up handoffs for the orchestrator
	orchestratorAgent.WithHandoffs(
		analyzerAgent,
		optimizerAgent,
	)

	// Create runner
	r := runner.NewRunner()
	r.WithDefaultProvider(provider)

	// Enable debug output if needed
	if err := os.Setenv("ANTHROPIC_DEBUG", "1"); err != nil {
		log.Printf("Warning: Failed to set ANTHROPIC_DEBUG environment variable: %v", err)
	}
	if err := os.Setenv("DEBUG", "1"); err != nil {
		log.Printf("Warning: Failed to set DEBUG environment variable: %v", err)
	}

	// Run the workflow
	fmt.Println("\nStarting the code review workflow with Anthropic Claude...")
	result, err := r.RunSync(orchestratorAgent, &runner.RunOptions{
		Input:    "Please review this code for me:\n\n" + sampleCode,
		MaxTurns: 10,
	})

	if err != nil {
		log.Fatalf("Error running agent: %v", err)
	}

	fmt.Println("\nWorkflow complete! Final result:")
	fmt.Println(result.FinalOutput)

	// Print detailed items if desired
	fmt.Println("\nItems generated:", len(result.NewItems))
	fmt.Println("\nDetailed items:")
	for i, item := range result.NewItems {
		fmt.Printf("Item %d: Type=%s\n", i, item.GetType())
	}
}

// Create the analyzer agent
func createAnalyzerAgent(provider *anthropic.Provider, getCodeInfo tool.Tool) *agent.Agent {
	analyzerAgent := agent.NewAgent("Analyzer")
	analyzerAgent.SetModelProvider(provider)
	analyzerAgent.WithModel("claude-3-haiku-20240307")
	analyzerAgent.WithTools(getCodeInfo)
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

When you receive code, first get the code information with get_code_info to see the code in context.
Then provide a comprehensive analysis with clear, actionable feedback.`)

	return analyzerAgent
}

// Create the optimizer agent
func createOptimizerAgent(provider *anthropic.Provider, getCodeInfo tool.Tool) *agent.Agent {
	optimizerAgent := agent.NewAgent("Optimizer")
	optimizerAgent.SetModelProvider(provider)
	optimizerAgent.WithModel("claude-3-haiku-20240307")
	optimizerAgent.WithTools(getCodeInfo)
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

When you receive a request, first get information about the code with get_code_info,
then provide the optimized code along with explanations of your changes.`)

	return optimizerAgent
}
