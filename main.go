package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/model"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/model/providers/lmstudio"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/runner"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/tool"
)

func main() {
	// Create a new LM Studio provider
	provider := lmstudio.NewProvider()

	// Set the base URL and default model
	provider.SetBaseURL("http://127.0.0.1:1234/v1")
	provider.SetDefaultModel("gemma-3-4b-it")

	// Create a new agent
	agent := agent.NewAgent("Time Assistant")

	// Set the model provider
	agent.SetModelProvider(provider)
	
	// Set the agent's model
	agent.WithModel("gemma-3-4b-it")

	// Set system instructions
	agent.SetSystemInstructions(`You are a helpful time assistant that can provide the current time in various formats.
When a user asks for the time, use the get_current_time tool to get accurate information.
After using tools, ALWAYS provide a complete response to the user's question in natural language.
Make your responses helpful and to the point.`)

	// Add a simple tool to get the current time
	agent.WithTools(tool.NewFunctionTool(
		"get_current_time",
		"Get the current time in a specified format",
		func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			// Default format is RFC3339
			format := time.RFC3339
			
			// Check if a format is specified
			if formatParam, ok := params["format"]; ok {
				if formatStr, ok := formatParam.(string); ok && formatStr != "" {
					switch formatStr {
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
			}
			
			return time.Now().Format(format), nil
		},
	).WithSchema(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"format": map[string]interface{}{
				"type": "string",
				"enum": []string{"rfc3339", "kitchen", "date", "datetime", "unix"},
				"description": "The format to return the time in. Options: rfc3339, kitchen, date, datetime, unix",
			},
		},
		"required": []string{},
	}))

	// Create a new runner
	runner := runner.NewRunner()
	runner.WithDefaultProvider(provider)

	// Run the agent with a basic query
	fmt.Println("Running agent with a basic query...")
	result, err := runner.Run(context.Background(), agent, &RunOptions{
		Input: "Hi Gemma! Could you please tell me what time it is right now?",
	})
	if err != nil {
		log.Fatalf("Error running agent: %v", err)
	}

	// Print the result
	fmt.Println("\nAgent response:")
	fmt.Println(result.FinalOutput)

	// Run the agent with a specific format request
	fmt.Println("\nRunning agent with a specific format request...")
	result, err = runner.Run(context.Background(), agent, &RunOptions{
		Input: "What time is it in kitchen format?",
	})
	if err != nil {
		log.Fatalf("Error running agent: %v", err)
	}

	// Print the result
	fmt.Println("\nAgent response:")
	fmt.Println(result.FinalOutput)

	// Run the agent streaming
	fmt.Println("\nRunning agent with streaming...")
	streamResult, err := runner.RunStreaming(context.Background(), agent, &RunOptions{
		Input: "What time is it right now?",
	})
	if err != nil {
		log.Fatalf("Error running agent with streaming: %v", err)
	}

	// Process the stream
	fmt.Println("\nStreaming response:")
	for event := range streamResult.Stream {
		switch event.Type {
		case model.StreamEventTypeContent:
			fmt.Print(event.Content)
		case model.StreamEventTypeToolCall:
			fmt.Printf("\n[Tool Call: %s]\n", event.ToolCall.Name)
		case model.StreamEventTypeDone:
			fmt.Println("\n[Done]")
		case model.StreamEventTypeError:
			fmt.Printf("\n[Error: %v]\n", event.Error)
			os.Exit(1)
		}
	}

	fmt.Println("\nFinal output:")
	fmt.Println(streamResult.RunResult.FinalOutput)
} 