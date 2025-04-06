# Anthropic Provider for Agent SDK

This package implements the Anthropic Claude API provider for the Agent SDK, supporting interaction with Claude models for agents, tool calls, and handoffs.

## Overview

The Anthropic provider connects to Claude models via the Anthropic API, allowing you to use models like Claude 3 Haiku, Claude 3 Sonnet, and Claude 3 Opus in your agents.

## Setup

To use the Anthropic provider, you'll need:

1. An Anthropic API key from [anthropic.com](https://anthropic.com)
2. The Agent SDK installed in your Go project

## Getting Started

Here's a basic example of setting up the Anthropic provider:

```go
import (
    "github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
    "github.com/pontus-devoteam/agent-sdk-go/pkg/model/providers/anthropic"
    "github.com/pontus-devoteam/agent-sdk-go/pkg/runner"
)

// Create a new Anthropic provider
provider := anthropic.NewProvider("your-anthropic-api-key")

// Set default model (optional)
provider.SetDefaultModel("claude-3-haiku-20240307")

// Configure rate limits (optional)
provider.WithRateLimit(40, 80000) // 40 requests/min, 80k tokens/min

// Configure retry settings (optional)
provider.WithRetryConfig(3, 2*time.Second) // 3 retries, 2s base delay

// Create an agent with the Anthropic provider
myAgent := agent.NewAgent("My Assistant")
myAgent.SetModelProvider(provider)
myAgent.WithModel("claude-3-haiku-20240307")
myAgent.SetSystemInstructions("You are a helpful assistant.")

// Run the agent
runner := runner.NewRunner()
result, err := runner.RunSync(myAgent, &runner.RunOptions{
    Input: "Hello, who are you?",
})
```

## Using Tools

The Anthropic provider supports tools for function calling with Claude models. Here's how to set up an agent with tools:

```go
// Create a calculator tool
calculatorTool := tool.NewFunctionTool(
    "calculator",
    "Calculate the result of a mathematical expression",
    func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
        // Tool implementation
        expression := params["expression"].(string)
        // Calculate result
        return "42", nil
    },
).WithSchema(map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "expression": map[string]interface{}{
            "type": "string",
            "description": "The math expression to calculate",
        },
    },
    "required": []string{"expression"},
})

// Add the tool to your agent
myAgent.WithTools(calculatorTool)
```

## Examples

Check out the example code in the `examples/anthropic_example` and `examples/anthropic_handoff_example` directories for complete working examples.

### Basic Example

```go
// Create an Anthropic provider
provider := anthropic.NewProvider(os.Getenv("ANTHROPIC_API_KEY"))
provider.SetDefaultModel("claude-3-haiku-20240307")

// Create an agent
agent := agent.NewAgent("Assistant")
agent.SetModelProvider(provider)
agent.WithModel("claude-3-haiku-20240307")
agent.SetSystemInstructions("You are a helpful assistant.")

// Run the agent
runner := runner.NewRunner()
result, err := runner.RunSync(agent, &runner.RunOptions{
    Input: "What is the capital of France?",
})

fmt.Println(result.FinalOutput)
```

### Tool Calling Example

```go
// Create an Anthropic provider
provider := anthropic.NewProvider(os.Getenv("ANTHROPIC_API_KEY"))
provider.SetDefaultModel("claude-3-haiku-20240307")

// Create a weather tool
weatherTool := tool.NewFunctionTool(
    "weather",
    "Get current weather conditions for a location",
    func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
        location := params["location"].(string)
        return fmt.Sprintf("Current weather in %s: 22°C, Partly Cloudy", location), nil
    },
).WithSchema(map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "location": map[string]interface{}{
            "type": "string",
            "description": "The city or location to get weather for",
        },
    },
    "required": []string{"location"},
})

// Create an agent with the tool
agent := agent.NewAgent("Weather Assistant")
agent.SetModelProvider(provider)
agent.WithModel("claude-3-haiku-20240307")
agent.SetSystemInstructions("You help users get weather information.")
agent.WithTools(weatherTool)

// Run the agent
runner := runner.NewRunner()
result, err := runner.RunSync(agent, &runner.RunOptions{
    Input: "What's the weather like in Paris today?",
})

fmt.Println(result.FinalOutput)
```

## Model Settings

You can configure various model settings:

```go
// Create an agent with specific settings
agent := agent.NewAgent("Assistant")
agent.SetModelProvider(provider)
agent.WithModel("claude-3-haiku-20240307")

// Apply settings
temperature := 0.7
maxTokens := 1024
agent.WithSettings(&model.Settings{
    Temperature: &temperature,
    MaxTokens: &maxTokens,
})
```

## Known Limitations

1. The Anthropic provider is designed to work with Claude 3 models. Older Claude versions may have limited functionality.

2. Some features specific to OpenAI like parallel tool calling are not currently supported by Anthropic's API.

## Troubleshooting

If you encounter issues with the Anthropic provider:

1. Verify your API key is valid and has sufficient quota
2. Check rate limits if you're making many requests
3. Ensure your tool schemas follow the correct format
4. Look for any error messages in the response

## Further Reading

- [Anthropic API Documentation](https://docs.anthropic.com/claude/reference/getting-started-with-the-api)
- [Claude 3 Model Information](https://www.anthropic.com/claude) 