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

func main() {
	// Get API key from environment
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is required")
	}

	// Create an Anthropic provider
	provider := anthropic.NewProvider(apiKey)

	// Set Claude 3 Haiku as the default model
	provider.SetDefaultModel("claude-3-haiku-20240307")

	// Configure rate limits
	provider.WithRateLimit(40, 80000) // 40 requests per minute, 80,000 tokens per minute

	// Configure retry settings
	provider.WithRetryConfig(3, 2*time.Second)

	fmt.Println("Provider configured with:")
	fmt.Println("- Model:", "claude-3-haiku-20240307")
	fmt.Println("- Rate limit:", "40 requests/min, 80,000 tokens/min")
	fmt.Println("- Max retries:", 3)

	// Create direct function tools for delegating to specialized agents
	weatherTool := createWeatherTool()
	translationTool := createTranslationTool()

	// Create the orchestrator agent
	orchestratorAgent := agent.NewAgent("Orchestrator")
	orchestratorAgent.SetModelProvider(provider)
	orchestratorAgent.WithModel("claude-3-haiku-20240307")
	orchestratorAgent.WithTools(weatherTool, translationTool)

	orchestratorAgent.SetSystemInstructions(`You are an orchestrator assistant that helps users by routing requests to specialized tools.

When given a question, determine which tool would best answer it:
1. For weather questions, use the "weather" tool
2. For language translation requests, use the "translation" tool

Always use the appropriate tool rather than trying to answer directly.
Be concise and focused in your tool usage, providing only the necessary information.`)

	// Create a runner for the workflow
	r := runner.NewRunner()
	r.WithDefaultProvider(provider)

	fmt.Println("\nStarting orchestrator workflow...")

	// Run the orchestrator with a weather question
	result, err := r.RunSync(orchestratorAgent, &runner.RunOptions{
		Input:    "What's the weather like in Paris today?",
		MaxTurns: 5,
	})
	if err != nil {
		log.Fatalf("Error running orchestrator: %v", err)
	}

	// Print the result
	fmt.Println("\nFinal response:")
	fmt.Println(result.FinalOutput)
}

// Create a mock weather tool
func createWeatherTool() tool.Tool {
	return tool.NewFunctionTool(
		"weather",
		"Get current weather conditions for a location",
		func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			location, ok := params["location"].(string)
			if !ok {
				return nil, fmt.Errorf("location parameter must be a string")
			}

			// Mock weather response
			return fmt.Sprintf("Current weather in %s: 22°C, Partly Cloudy", location), nil
		},
	).WithSchema(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"location": map[string]interface{}{
				"type":        "string",
				"description": "The city or location to get weather for",
			},
		},
		"required": []string{"location"},
	})
}

// Create a mock translation tool
func createTranslationTool() tool.Tool {
	return tool.NewFunctionTool(
		"translation",
		"Translate text to another language",
		func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			text, ok := params["text"].(string)
			if !ok {
				return nil, fmt.Errorf("text parameter must be a string")
			}

			targetLang, ok := params["target_language"].(string)
			if !ok {
				return nil, fmt.Errorf("target_language parameter must be a string")
			}

			// Simple mock translations for common phrases
			if text == "Hello, how are you?" && targetLang == "French" {
				return "Bonjour, comment allez-vous?", nil
			} else if text == "Thank you" && targetLang == "Spanish" {
				return "Gracias", nil
			} else {
				return fmt.Sprintf("[Translation of '%s' to %s]", text, targetLang), nil
			}
		},
	).WithSchema(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"text": map[string]interface{}{
				"type":        "string",
				"description": "The text to translate",
			},
			"target_language": map[string]interface{}{
				"type":        "string",
				"description": "The language to translate to",
			},
		},
		"required": []string{"text", "target_language"},
	})
}
