package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Muhammadhamd/agent-sdk-go/pkg/agent"
	"github.com/Muhammadhamd/agent-sdk-go/pkg/model/providers/anthropic"
	"github.com/Muhammadhamd/agent-sdk-go/pkg/runner"
	"github.com/Muhammadhamd/agent-sdk-go/pkg/tool"
)

func main() {
	// Get API key from environment or use provided key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is required")
	}

	// Create an Anthropic provider
	provider := anthropic.NewProvider(apiKey)

	// Set Claude 3 Haiku as the default model
	provider.SetDefaultModel("claude-3-haiku-20240307")

	// Configure rate limits if needed (adjust these based on your expected usage)
	provider.WithRateLimit(40, 80000) // 40 requests per minute, 80,000 tokens per minute

	// Configure retry settings
	provider.WithRetryConfig(3, 2*time.Second)

	fmt.Println("Provider configured with:")
	fmt.Println("- Model:", "claude-3-haiku-20240307")
	fmt.Println("- Rate limit:", "40 requests/min, 80,000 tokens/min")
	fmt.Println("- Max retries:", 3)

	// Create a simple calculator tool
	calculatorTool := tool.NewFunctionTool(
		"calculator",
		"Calculate the result of a mathematical expression",
		func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			// Extract parameters
			operation, ok := params["operation"].(string)
			if !ok {
				return nil, fmt.Errorf("operation must be a string")
			}

			// Extract the first number
			n1, ok := params["num1"].(float64)
			if !ok {
				return nil, fmt.Errorf("num1 must be a number")
			}
			num1 := n1

			// Extract the second number
			n2, ok := params["num2"].(float64)
			if !ok {
				return nil, fmt.Errorf("num2 must be a number")
			}
			num2 := n2

			// Perform the operation
			var result float64
			switch operation {
			case "add":
				result = num1 + num2
			case "subtract":
				result = num1 - num2
			case "multiply":
				result = num1 * num2
			case "divide":
				if num2 == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				result = num1 / num2
			default:
				return nil, fmt.Errorf("unsupported operation: %s", operation)
			}

			return map[string]interface{}{
				"result": result,
			}, nil
		},
	).WithSchema(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"operation": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"add", "subtract", "multiply", "divide"},
				"description": "The operation to perform",
			},
			"num1": map[string]interface{}{
				"type":        "number",
				"description": "The first number",
			},
			"num2": map[string]interface{}{
				"type":        "number",
				"description": "The second number",
			},
		},
		"required": []string{"operation", "num1", "num2"},
	})

	// Create a simple agent that can use the calculator
	calcAssistant := agent.NewAgent("Anthropic Calculator Assistant")
	calcAssistant.SetModelProvider(provider)
	calcAssistant.WithModel("claude-3-haiku-20240307")
	calcAssistant.SetSystemInstructions("You are an AI assistant that can perform calculations. You MUST use the calculator tool when asked to perform any mathematical operation. Never try to do the calculation yourself - always use the tool.")
	calcAssistant.WithTools(calculatorTool)

	// Create a runner
	r := runner.NewRunner()
	r.WithDefaultProvider(provider)

	// Run the agent with a calculation request
	fmt.Println("\nSending a calculation request to the agent...")
	result, err := r.RunSync(calcAssistant, &runner.RunOptions{
		Input:    "What is 42 * 23?",
		MaxTurns: 5,
	})
	if err != nil {
		log.Fatalf("Error running agent: %v", err)
	}

	// Print the result
	fmt.Println("\nAgent response:")
	fmt.Println(result.FinalOutput)

	// Run with another calculation
	fmt.Println("\nSending another calculation request...")
	result, err = r.RunSync(calcAssistant, &runner.RunOptions{
		Input:    "If I have 5 items that cost $12.50 each, and 3 items that cost $8.75 each, what's the total?",
		MaxTurns: 5,
	})
	if err != nil {
		log.Fatalf("Error running agent: %v", err)
	}

	// Print the result
	fmt.Println("\nAgent response:")
	fmt.Println(result.FinalOutput)
}
