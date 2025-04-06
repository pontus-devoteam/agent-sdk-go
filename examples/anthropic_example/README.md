# Anthropic (Claude) Example

This example demonstrates how to use the Anthropic (Claude) provider with the agent-sdk-go. It creates a simple calculator agent that can perform basic arithmetic operations using Claude's tool use capability.

## Features

- Uses Claude 3 Haiku as the LLM
- Implements a basic calculator tool with add, subtract, multiply, and divide operations
- Demonstrates both synchronous and streaming interactions
- Shows how to configure rate limiting and retry settings

## Prerequisites

- An Anthropic API key

## Running the Example

Set your Anthropic API key as an environment variable:

```bash
export ANTHROPIC_API_KEY=your_api_key_here
```

Or, the example includes a demo API key, but it's always better to use your own.

Then run the example:

```bash
go run examples/anthropic_example/main.go
```

## Example Interactions

The example will automatically run three scenarios:

1. A basic calculation: "What is 25 * 48?"
2. A more complex problem: "If I have 5 apples and each costs $2.50, and 3 oranges where each costs $1.75, how much would I pay in total?"
3. A streaming calculation with step-by-step explanation: "(15 * 8) + (24 / 3) - 7"

## Customization

You can modify the example to use different Claude models (like claude-3-opus or claude-3-sonnet) by changing the `SetDefaultModel` and `WithModel` calls. You can also add additional tools or modify the system instructions to change the assistant's behavior. 