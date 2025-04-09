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

func main() {
	// Enable verbose logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Set API key
	apiKey := "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"

	// Check if env var is set and use that instead if available
	if envKey := os.Getenv("OPENAI_API_KEY"); envKey != "" {
		apiKey = envKey
	}

	// Create a provider for OpenAI
	provider := openai.NewProvider(apiKey)

	// Configure the provider with GPT-4o-mini
	provider.SetDefaultModel("gpt-4o-mini")

	// Customize rate limits (adjust based on your OpenAI tier)
	provider.WithRateLimit(60, 150000) // 60 requests per minute, 150,000 tokens per minute

	// Configure retry settings
	provider.WithRetryConfig(3, 2*time.Second)

	fmt.Println("Provider configured with:")
	fmt.Println("- Model:", "gpt-4o-mini")
	fmt.Println("- Rate limit:", "60 requests/min, 150,000 tokens/min")
	fmt.Println("- Max retries:", 3)

	// Create a simple agent
	assistant := agent.NewAgent("Assistant")
	assistant.SetModelProvider(provider)
	assistant.WithModel("gpt-4o-mini")
	assistant.SetSystemInstructions(`You are a helpful assistant that can provide information and answer questions.
You can use tools to get information that you might not know, like the current time.`)

	// Time Tool
	timeTool := tool.NewFunctionTool(
		"get_current_time",
		"Get the current time in a specified format. This tool will return the current system time, not the time in a specific location.",
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

	// Add tools to the agent
	assistant.WithTools(timeTool)

	// Create a runner
	r := runner.NewRunner()
	r.WithDefaultProvider(provider)

	// Run the agent
	fmt.Println("\nRunning the agent...")
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

	// Display token usage if available
	if len(result.RawResponses) > 0 {
		lastResponse := result.RawResponses[len(result.RawResponses)-1]
		if lastResponse.Usage != nil {
			fmt.Printf("\nToken usage: %d total tokens\n", lastResponse.Usage.TotalTokens)
		}
	}

	// Test streaming
	fmt.Println("\nTesting streaming response...")
	streamResult, err := r.RunStreaming(context.Background(), assistant, &runner.RunOptions{
		Input: "Tell me the current time in both RFC3339 format and kitchen format.",
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
