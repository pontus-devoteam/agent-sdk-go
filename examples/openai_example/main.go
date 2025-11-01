package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Muhammadhamd/agent-sdk-go/pkg/agent"
	"github.com/Muhammadhamd/agent-sdk-go/pkg/model/providers/openai"
	"github.com/Muhammadhamd/agent-sdk-go/pkg/runner"
	"github.com/Muhammadhamd/agent-sdk-go/pkg/tool"
)

func main() {
	// Get API key from environment variable
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}

	// Create a provider for OpenAI
	provider := openai.NewProvider(apiKey)

	// Configure the provider
	provider.SetDefaultModel("gpt-3.5-turbo")

	// Customize rate limits if needed (adjust these based on your OpenAI tier)
	provider.WithRateLimit(50, 100000) // 50 requests per minute, 100,000 tokens per minute

	// Configure retry settings
	provider.WithRetryConfig(3, 2*time.Second)

	fmt.Println("Provider configured with:")
	fmt.Println("- Model:", "gpt-3.5-turbo")
	fmt.Println("- Rate limit:", "50 requests/min, 100,000 tokens/min")
	fmt.Println("- Max retries:", 3)

	// Create a simple tool
	getCurrentTimeTool := tool.NewFunctionTool(
		"get_current_time",
		"Get the current time in a specified format",
		func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			format := time.RFC3339

			if formatParam, ok := params["format"].(string); ok && formatParam != "" {
				switch formatParam {
				case "rfc3339":
					format = time.RFC3339
				case "kitchen":
					format = time.Kitchen
				case "date":
					format = "2006-01-02"
				case "datetime":
					format = "2006-01-02 15:04:05"
				case "unix":
					return time.Now().Unix(), nil
				}
			}

			return time.Now().Format(format), nil
		},
	).WithSchema(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"format": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"rfc3339", "kitchen", "date", "datetime", "unix"},
				"description": "The format to return the time in. Options: rfc3339, kitchen, date, datetime, unix",
			},
		},
		"required": []string{},
	})

	// Create an agent
	assistant := agent.NewAgent("OpenAI Assistant")
	assistant.SetModelProvider(provider)
	assistant.WithModel("gpt-3.5-turbo")
	assistant.SetSystemInstructions("You are a helpful assistant that can provide information and answer questions.")
	assistant.WithTools(getCurrentTimeTool)

	// Create a runner
	r := runner.NewRunner()
	r.WithDefaultProvider(provider)

	// Run the agent
	fmt.Println("\nSending a basic question to the agent...")
	result, err := r.RunSync(assistant, &runner.RunOptions{
		Input:    "What's the current time?",
		MaxTurns: 10,
	})
	if err != nil {
		log.Fatalf("Error running agent: %v", err)
	}

	// Print the result
	fmt.Println("\nAgent response:")
	fmt.Println(result.FinalOutput)

	// If there are any responses, display token usage from the last response
	if len(result.RawResponses) > 0 {
		lastResponse := result.RawResponses[len(result.RawResponses)-1]
		if lastResponse.Usage != nil {
			fmt.Printf("\nToken usage: %d total tokens\n", lastResponse.Usage.TotalTokens)
		}
	}

	// Run another example with a more complex question
	fmt.Println("\nSending a complex question to the agent...")
	result, err = r.RunSync(assistant, &runner.RunOptions{
		Input:    "Can you tell me the current time in both RFC3339 format and as a Unix timestamp?",
		MaxTurns: 10,
	})
	if err != nil {
		log.Fatalf("Error running agent: %v", err)
	}

	// Print the result
	fmt.Println("\nAgent response:")
	fmt.Println(result.FinalOutput)

	// If there are any responses, display token usage from the last response
	if len(result.RawResponses) > 0 {
		lastResponse := result.RawResponses[len(result.RawResponses)-1]
		if lastResponse.Usage != nil {
			fmt.Printf("\nToken usage: %d total tokens\n", lastResponse.Usage.TotalTokens)
		}
	}

	// Test streaming
	fmt.Println("\nTesting streaming response...")
	streamResult, err := r.RunStreaming(context.Background(), assistant, &runner.RunOptions{
		Input: "Count from 1 to 5 slowly, with a brief pause between each number.",
	})
	if err != nil {
		log.Fatalf("Error running streaming: %v", err)
	}

	fmt.Println("\nStreaming response:")
	for event := range streamResult.Stream {
		switch event.Type {
		case "content":
			fmt.Print(event.Content)
		case "tool_call":
			fmt.Printf("\n[Calling tool: %s]\n", event.ToolCall.Name)
		case "error":
			fmt.Printf("\nError: %v\n", event.Error)
		case "done":
			fmt.Println("\n[Done]")
		}
	}
}
