<p align="center">
  <img src="./agent-sdk-go-header.gif" alt="Agent SDK Go">
</p>

<div align="center">
  <p><strong>Build, deploy, and scale AI agents with ease</strong></p>
  
  <a href="https://go-agent.org"><img src="https://img.shields.io/badge/website-go--agent.org-blue?style=for-the-badge" alt="Website" /></a>
  <a href="https://go-agent.org/#waitlist"><img src="https://img.shields.io/badge/Cloud_Waitlist-Sign_Up-4285F4?style=for-the-badge" alt="Cloud Waitlist" /></a>
  
</div>

<p align="center">
  Agent SDK Go is an open-source framework for building powerful AI agents with Go that supports multiple LLM providers, function calling, agent handoffs, and more.
</p>

<p align="center">
    <a href="https://github.com/Muhammadhamd/agent-sdk-go/actions/workflows/code-quality.yml"><img src="https://github.com/Muhammadhamd/agent-sdk-go/actions/workflows/code-quality.yml/badge.svg" alt="Code Quality"></a>
    <a href="https://goreportcard.com/report/github.com/Muhammadhamd/agent-sdk-go/"><img src="https://goreportcard.com/badge/github.com/Muhammadhamd/agent-sdk-go/" alt="Go Report Card"></a>
    <a href="https://github.com/Muhammadhamd/agent-sdk-go/blob/main/go.mod"><img src="https://img.shields.io/github/go-mod/go-version/Muhammadhamd/agent-sdk-go" alt="Go Version"></a>
    <a href="https://pkg.go.dev/github.com/Muhammadhamd/agent-sdk-go/"><img src="https://pkg.go.dev/badge/github.com/Muhammadhamd/agent-sdk-go/.svg" alt="PkgGoDev"></a><br>
    <a href="https://github.com/Muhammadhamd/agent-sdk-go/actions/workflows/codeql-analysis.yml"><img src="https://github.com/Muhammadhamd/agent-sdk-go/actions/workflows/codeql-analysis.yml/badge.svg" alt="CodeQL"></a>
    <a href="https://github.com/Muhammadhamd/agent-sdk-go/blob/main/LICENSE"><img src="https://img.shields.io/github/license/Muhammadhamd/agent-sdk-go" alt="License"></a>
    <a href="https://github.com/Muhammadhamd/agent-sdk-go/stargazers"><img src="https://img.shields.io/github/stars/Muhammadhamd/agent-sdk-go" alt="Stars"></a>
    <a href="https://github.com/Muhammadhamd/agent-sdk-go/graphs/contributors"><img src="https://img.shields.io/github/contributors/Muhammadhamd/agent-sdk-go" alt="Contributors"></a>
    <a href="https://github.com/Muhammadhamd/agent-sdk-go/commits/main"><img src="https://img.shields.io/github/last-commit/Muhammadhamd/agent-sdk-go" alt="Last Commit"></a>
</p>

<p align="center">
  <a href="https://go-agent.org/#waitlist">☁️ Cloud Waitlist</a> •
  <a href="https://github.com/Muhammadhamd/agent-sdk-go/blob/main/LICENSE">📜 License</a>
</p>

<p align="center">
  Inspired by <a href="https://platform.openai.com/docs/assistants/overview">OpenAI's Assistants API</a> and <a href="https://github.com/openai/openai-agents-python">OpenAI's Python Agent SDK</a>.
</p>

---

## 📋 Table of Contents

- [Overview](#-overview)
- [Features](#-features)
- [Installation](#-installation)
- [Quick Start](#-quick-start)
- [Provider Setup](#-provider-setup)
- [Key Components](#-key-components)
  - [Agent](#agent)
  - [Runner](#runner)
  - [Tools](#tools)
  - [Model Providers](#model-providers)
- [Advanced Features](#-advanced-features)
  - [Multi-Agent Workflows](#multi-agent-workflows)
  - [Tracing](#tracing)
  - [Structured Output](#structured-output)
  - [Streaming](#streaming)
  - [OpenAI Tool Definitions](#openai-tool-definitions)
  - [Workflow State Management](#workflow-state-management)
  - [Bidirectional Agent Flow](#bidirectional-agent-flow)
- [Examples](#-examples)
- [Cloud Support](#-cloud-support)
- [Development](#-development)
- [Contributing](#-contributing)
- [License](#-license)
- [Acknowledgements](#-acknowledgements)

---

## 🔍 Overview

Agent SDK Go provides a comprehensive framework for building AI agents in Go. It allows you to create agents that can use tools, perform handoffs to other specialized agents, and produce structured output - all while supporting multiple LLM providers.

**Visit [go-agent.org](https://go-agent.org) for comprehensive documentation, examples, and cloud service waitlist.**

## 🌟 Features

- ✅ **Multiple LLM Provider Support** - Support for OpenAI, Anthropic Claude, and LM Studio
- ✅ **Tool Integration** - Call Go functions directly from your LLM
- ✅ **Agent Handoffs** - Create complex multi-agent workflows with specialized agents
- ✅ **Structured Output** - Parse responses into Go structs
- ✅ **Streaming** - Get real-time streaming responses
- ✅ **Tracing & Monitoring** - Debug your agent flows
- ✅ **OpenAI Compatibility** - Compatible with OpenAI tool definitions and API
- ✅ **Workflow State Management** - Persist and manage state between agent executions

## 📦 Installation

There are several ways to add this module to your project:

### Option 1: Using `go get` (Recommended)

```bash
go get github.com/Muhammadhamd/agent-sdk-go/
```

### Option 2: Add to your imports and use `go mod tidy`

1. Add imports to your Go files:
   ```go
   import (
       "github.com/Muhammadhamd/agent-sdk-go/pkg/agent"
       "github.com/Muhammadhamd/agent-sdk-go/pkg/model/providers/lmstudio"
       "github.com/Muhammadhamd/agent-sdk-go/pkg/runner"
       "github.com/Muhammadhamd/agent-sdk-go/pkg/tool"
       // Import other packages as needed
   )
   ```

2. Run `go mod tidy` to automatically fetch dependencies:
   ```bash
   go mod tidy
   ```

### Option 3: Manually edit your `go.mod` file

Add the following line to your `go.mod` file:
```
require github.com/Muhammadhamd/agent-sdk-go/ latest
```

Then run:
```bash
go mod tidy
```

### New Project Setup

If you're starting a new project:

1. Create and navigate to your project directory:
   ```bash
   mkdir my-agent-project
   cd my-agent-project
   ```

2. Initialize a new Go module:
   ```bash
   go mod init github.com/yourusername/my-agent-project
   ```

3. Install the Agent SDK:
   ```bash
   go get github.com/Muhammadhamd/agent-sdk-go/
   ```

### Troubleshooting

- If you encounter version conflicts, you can specify a version:
  ```bash
  go get github.com/Muhammadhamd/agent-sdk-go/@v0.1.0  # Replace with desired version
  ```

- For private repositories or local development, consider using Go workspaces or replace directives in your go.mod file.

> **Note:** Requires Go 1.23 or later.

## 🚀 Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Muhammadhamd/agent-sdk-go/pkg/agent"
    "github.com/Muhammadhamd/agent-sdk-go/pkg/model/providers/openai"  // or providers/lmstudio or providers/anthropic
    "github.com/Muhammadhamd/agent-sdk-go/pkg/runner"
    "github.com/Muhammadhamd/agent-sdk-go/pkg/tool"
)

func main() {
    // Create a provider (OpenAI example)
    provider := openai.NewProvider("your-openai-api-key")
    provider.SetDefaultModel("gpt-3.5-turbo")

    // Or use Anthropic Claude (example)
    // provider := anthropic.NewProvider("your-anthropic-api-key")
    // provider.SetDefaultModel("claude-3-haiku-20240307")

    // Or use LM Studio (local model example)
    // provider := lmstudio.NewProvider()
    // provider.SetBaseURL("http://127.0.0.1:1234/v1")
    // provider.SetDefaultModel("gemma-3-4b-it")

    // Create a function tool
    getWeather := tool.NewFunctionTool(
        "get_weather",
        "Get the weather for a city",
        func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
            city := params["city"].(string)
            return fmt.Sprintf("The weather in %s is sunny.", city), nil
        },
    ).WithSchema(map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "city": map[string]interface{}{
                "type": "string",
                "description": "The city to get weather for",
            },
        },
        "required": []string{"city"},
    })

    // Create an agent
    assistant := agent.NewAgent("Assistant")
    assistant.SetModelProvider(provider)
    assistant.WithModel("gpt-3.5-turbo")  // or "gemma-3-4b-it" for LM Studio or "claude-3-haiku-20240307" for Anthropic
    assistant.SetSystemInstructions("You are a helpful assistant.")
    assistant.WithTools(getWeather)

    // Create a runner
    r := runner.NewRunner()
    r.WithDefaultProvider(provider)

    // Run the agent
    result, err := r.RunSync(assistant, &runner.RunOptions{
        Input: "What's the weather in Tokyo?",
    })
    if err != nil {
        log.Fatalf("Error running agent: %v", err)
    }

    // Print the result
    fmt.Println(result.FinalOutput)
}
```

## 🖥️ Provider Setup

### OpenAI Setup

To use the OpenAI provider:

1. **Get an API Key**
   - Sign up at [OpenAI](https://platform.openai.com/)
   - Create an API key in your account settings

2. **Configure the Provider**
   ```go
   provider := openai.NewProvider()
   provider.SetAPIKey("your-openai-api-key")
   provider.SetDefaultModel("gpt-3.5-turbo")  // or any other OpenAI model
   ```

### Anthropic Setup

<details>
<summary>Click to expand setup instructions</summary>

To use the Anthropic provider:

1. **Get an API Key**
   - Sign up at [Anthropic Console](https://console.anthropic.com/)
   - Create an API key in your account settings

2. **Configure the Provider**
   ```go
   provider := anthropic.NewProvider("your-anthropic-api-key")
   provider.SetDefaultModel("claude-3-haiku-20240307")  // or claude-3-sonnet/opus
   
   // Optional rate limiting configuration
   provider.WithRateLimit(40, 80000) // 40 requests/min, 80,000 tokens/min
   
   // Optional retry configuration
   provider.WithRetryConfig(3, 2*time.Second) // 3 retries with exponential backoff
   ```

</details>

### LM Studio Setup

<details>
<summary>Click to expand setup instructions</summary>

To use the LM Studio provider:

1. **Install LM Studio**
   - Download from [lmstudio.ai](https://lmstudio.ai/)
   - Install and run the application

2. **Load a Model**
   - Download a model in LM Studio (Like Gemma-3-4B-It, Llama3, or other compatible models)
   - Load the model

3. **Start the Server**
   - Go to the "Local Server" tab
   - Click "Start Server"
   - Note the server URL (default: http://127.0.0.1:1234)

4. **Configure the Provider**
   ```go
   provider := lmstudio.NewProvider()
   provider.SetBaseURL("http://127.0.0.1:1234/v1")
   provider.SetDefaultModel("gemma-3-4b-it") // Replace with your model
   ```

</details>

## 🧩 Key Components

### Agent

The Agent is the core component that encapsulates the LLM with instructions, tools, and other configuration.

```go
// Create a new agent
agent := agent.NewAgent("Assistant")
agent.SetSystemInstructions("You are a helpful assistant.")
agent.WithModel("gemma-3-4b-it")
agent.WithTools(tool1, tool2) // Add multiple tools at once
```

### Runner

The Runner executes agents, handling the agent loop, tool calls, and handoffs.

```go
// Create a runner
runner := runner.NewRunner()
runner.WithDefaultProvider(provider)

// Run the agent
result, err := runner.RunSync(agent, &runner.RunOptions{
    Input: "Hello, world!",
    MaxTurns: 10, // Optional: limit the number of turns
})
```

### Tools

Tools allow agents to perform actions using your Go functions.

```go
// Create a function tool
tool := tool.NewFunctionTool(
    "get_weather",
    "Get the weather for a city",
    func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
        city := params["city"].(string)
        return fmt.Sprintf("The weather in %s is sunny.", city), nil
    },
).WithSchema(map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "city": map[string]interface{}{
            "type": "string",
            "description": "The city to get weather for",
        },
    },
    "required": []string{"city"},
})
```

### Model Providers

Model providers allow you to use different LLM providers.

```go
// Create a provider for OpenAI
openaiProvider := openai.NewProvider("your-openai-api-key")
openaiProvider.SetDefaultModel("gpt-4")

// Create a provider for Anthropic Claude
anthropicProvider := anthropic.NewProvider("your-anthropic-api-key")
anthropicProvider.SetDefaultModel("claude-3-haiku-20240307")

// Create a provider for LM Studio
lmStudioProvider := lmstudio.NewProvider()
lmStudioProvider.SetBaseURL("http://127.0.0.1:1234/v1")
lmStudioProvider.SetDefaultModel("gemma-3-4b-it")

// Set a provider as the default provider
runner := runner.NewRunner()
runner.WithDefaultProvider(openaiProvider) // or anthropicProvider or lmStudioProvider
```

## 🔧 Advanced Features

### Multi-Agent Workflows

<details>
<summary>Create specialized agents that collaborate on complex tasks</summary>

```go
// Create specialized agents
mathAgent := agent.NewAgent("Math Agent")
mathAgent.SetModelProvider(provider)
mathAgent.WithModel("gemma-3-4b-it")
mathAgent.SetSystemInstructions("You are a specialized math agent.")
mathAgent.WithTools(calculatorTool)

weatherAgent := agent.NewAgent("Weather Agent")
weatherAgent.SetModelProvider(provider)
weatherAgent.WithModel("gemma-3-4b-it")
weatherAgent.SetSystemInstructions("You provide weather information.")
weatherAgent.WithTools(weatherTool)

// Create a frontend agent that coordinates tasks
frontendAgent := agent.NewAgent("Frontend Agent")
frontendAgent.SetModelProvider(provider)
frontendAgent.WithModel("gemma-3-4b-it")
frontendAgent.SetSystemInstructions(`You coordinate requests by delegating to specialized agents.
For math calculations, delegate to the Math Agent.
For weather information, delegate to the Weather Agent.`)
frontendAgent.WithHandoffs(mathAgent, weatherAgent)

// Run the frontend agent
result, err := runner.RunSync(frontendAgent, &runner.RunOptions{
    Input: "What is 42 divided by 6 and what's the weather in Paris?",
    MaxTurns: 20,
})
```

See the complete example in [examples/multi_agent_example](./examples/multi_agent_example).

</details>

### Bidirectional Agent Flow

<details>
<summary>Create agents that can hand off tasks and receive results back</summary>

Bidirectional agent flow allows agents to delegate tasks to other agents and receive results back once the tasks are complete. This enables more complex workflows with proper task context management.

```go
// Create specialized agents
orchestratorAgent := agent.NewAgent("Orchestrator")
orchestratorAgent.SetModelProvider(provider)
orchestratorAgent.WithModel("gpt-4")
orchestratorAgent.SetSystemInstructions("You coordinate tasks and analyze results.")

workerAgent := agent.NewAgent("Worker")
workerAgent.SetModelProvider(provider)
workerAgent.WithModel("gpt-3.5-turbo")
workerAgent.SetSystemInstructions("You process data and return results.")
workerAgent.WithTools(processingTool)

// Set up bidirectional handoffs
orchestratorAgent.WithHandoffs(workerAgent)
workerAgent.WithHandoffs(orchestratorAgent)  // Allow worker to return to orchestrator

// Run the orchestrator agent
result, err := runner.RunSync(orchestratorAgent, &runner.RunOptions{
    Input: "Analyze this data: [complex data]",
    MaxTurns: 10,
})
```

Key components of bidirectional flow:
- `TaskID`: Unique identifier for tracking tasks across agents
- `ReturnToAgent`: Specifies which agent to return to after task completion
- `IsTaskComplete`: Flag indicating whether the task is complete

See the complete example in [examples/bidirectional_flow_example](./examples/bidirectional_flow_example).

</details>

### Tracing

<details>
<summary>Debug your agent workflows with tracing</summary>

```go
// Run with tracing enabled
result, err := runner.RunSync(agent, &runner.RunOptions{
    Input: "Hello, world!",
    RunConfig: &runner.RunConfig{
        TracingDisabled: false,
        TracingConfig: &runner.TracingConfig{
            WorkflowName: "my_workflow",
        },
    },
})
```

</details>

### Structured Output

<details>
<summary>Parse responses into Go structs</summary>

```go
// Define an output type
type WeatherReport struct {
    City        string  `json:"city"`
    Temperature float64 `json:"temperature"`
    Condition   string  `json:"condition"`
}

// Create an agent with structured output
agent := agent.NewAgent("Weather Agent")
agent.SetSystemInstructions("You provide weather reports")
agent.SetOutputType(reflect.TypeOf(WeatherReport{}))
```

</details>

### Streaming

<details>
<summary>Get real-time streaming responses</summary>

```go
// Run the agent with streaming
streamedResult, err := runner.RunStreaming(context.Background(), agent, &runner.RunOptions{
    Input: "Hello, world!",
})
if err != nil {
    log.Fatalf("Error running agent: %v", err)
}

// Process streaming events
for event := range streamedResult.Stream {
    switch event.Type {
    case model.StreamEventTypeContent:
        fmt.Print(event.Content)
    case model.StreamEventTypeToolCall:
        fmt.Printf("\nCalling tool: %s\n", event.ToolCall.Name)
    case model.StreamEventTypeDone:
        fmt.Println("\nDone!")
    }
}
```

</details>

### OpenAI Tool Definitions

<details>
<summary>Work with OpenAI-compatible tool definitions</summary>

```go
// Auto-generate OpenAI-compatible tool definitions from Go functions
getCurrentTimeTool := tool.NewFunctionTool(
    "get_current_time",
    "Get the current time in a specified format",
    func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
        return time.Now().Format(time.RFC3339), nil
    },
)

// Convert it to OpenAI format (handled automatically when added to an agent)
openAITool := tool.ToOpenAITool(getCurrentTimeTool)

// Add an OpenAI-compatible tool definition directly to an agent
agent := agent.NewAgent("My Agent")
agent.AddToolFromDefinition(openAITool)

// Add multiple tool definitions at once
toolDefinitions := []map[string]interface{}{
    tool.ToOpenAITool(tool1),
    tool.ToOpenAITool(tool2),
}

agent.AddToolsFromDefinitions(toolDefinitions)
```

</details>

### Workflow State Management

<details>
<summary>Manage state between agent executions</summary>

```go
// Create a state store
stateStore := mocks.NewInMemoryStateStore()

// Create workflow configuration
workflowConfig := &runner.WorkflowConfig{
    RetryConfig: &runner.RetryConfig{
        MaxRetries:         2,
        RetryDelay:        time.Second,
        RetryBackoffFactor: 2.0,
    },
    StateManagement: &runner.StateManagementConfig{
        PersistState:        true,
        StateStore:          stateStore,
        CheckpointFrequency: time.Second * 5,
    },
    ValidationConfig: &runner.ValidationConfig{
        PreHandoffValidation: []runner.ValidationRule{
            {
                Name:         "StateValidation",
                Validate:     func(data interface{}) (bool, error) {
                    state, ok := data.(*runner.WorkflowState)
                    return ok && state != nil, nil
                },
                ErrorMessage: "Invalid workflow state",
                Severity:     runner.ValidationWarning,
            },
        },
    },
}

// Create workflow runner
workflowRunner := runner.NewWorkflowRunner(baseRunner, workflowConfig)

// Initialize workflow state
state := &runner.WorkflowState{
    CurrentPhase:    "",
    CompletedPhases: make([]string, 0),
    Artifacts:       make(map[string]interface{}),
    LastCheckpoint:  time.Now(),
    Metadata:        make(map[string]interface{}),
}

// Run workflow with state management
result, err := workflowRunner.RunWorkflow(context.Background(), agent, &runner.RunOptions{
    MaxTurns:       10,
    RunConfig:      runConfig,
    WorkflowConfig: workflowConfig,
    Input:         state,
})
```

See the complete example in [examples/workflow_example](./examples/workflow_example).
</details>

## 📚 Examples

The repository includes several examples to help you get started:

| Example | Description |
|---------|-------------|
| [Multi-Agent Example](./examples/multi_agent_example) | Demonstrates how to create a system of specialized agents that can collaborate on complex tasks using a local LLM via LM Studio |
| [OpenAI Example](./examples/openai_example) | Shows how to use the OpenAI provider with function calling capabilities |
| [OpenAI Multi-Agent Example](./examples/openai_multi_agent_example) | Illustrates multi-agent functionality using OpenAI models, with proper tool calling and streaming support |
| [Anthropic Example](./examples/anthropic_example) | Demonstrates how to use the Anthropic Claude API with tool calling capabilities |
| [Anthropic Handoff Example](./examples/anthropic_handoff_example) | Shows how to implement agent handoffs with Anthropic Claude models |
| [Bidirectional Flow Example](./examples/bidirectional_flow_example) | Demonstrates bidirectional agent communication with task delegation and return handoffs |
| [TypeScript Code Review Example](./examples/typescript_code_review_example) | Shows a practical application with specialized code review agents that collaborate using bidirectional handoffs |
| [Workflow Example](./examples/workflow_example) | Demonstrates advanced workflow management with state persistence between agent executions |

### Running Examples with a Local LLM

1. Make sure LM Studio is running with a server at `http://127.0.0.1:1234/v1`
2. Navigate to the example directory
   ```bash
   cd examples/multi_agent_example # or any other example using LM Studio
   ```
3. Run the example
   ```bash
   go run .
   ```

### Running Examples with OpenAI

1. Set your OpenAI API key as an environment variable
   ```bash
   export OPENAI_API_KEY=your-api-key
   ```
2. Navigate to the example directory
   ```bash
   cd examples/openai_example # or openai_multi_agent_example
   ```
3. Run the example
   ```bash
   go run .
   ```

### Running Examples with Anthropic

1. Set your Anthropic API key as an environment variable
   ```bash
   export ANTHROPIC_API_KEY=your-anthropic-api-key
   ```
2. Navigate to the example directory
   ```bash
   cd examples/anthropic_example # or anthropic_handoff_example
   ```
3. Run the example
   ```bash
   go run .
   ```

### Debugging

You can enable debug output for various components by setting the appropriate environment variable:

For general debugging (runner and core components):
```bash
DEBUG=1 go run examples/bidirectional_flow_example/main.go
```

For provider-specific debugging:
```bash
# OpenAI provider debugging
OPENAI_DEBUG=1 go run examples/openai_multi_agent_example/main.go

# Anthropic provider debugging
ANTHROPIC_DEBUG=1 go run examples/anthropic_example/main.go

# LM Studio provider debugging
LMSTUDIO_DEBUG=1 go run examples/multi_agent_example/main.go
```

You can also combine multiple debug flags:
```bash
DEBUG=1 OPENAI_DEBUG=1 go run examples/typescript_code_review_example/main.go
```

## 🛠️ Development

<details>
<summary>Development setup and workflows</summary>

### Requirements

- Go 1.23 or later

### Setup

1. Clone the repository
2. Run the setup script to install required tools:

```bash
./scripts/ci_setup.sh
```

### Development Workflow

The project includes several scripts to help with development:

- `./scripts/lint.sh`: Runs formatting and linting checks
- `./scripts/security_check.sh`: Runs security checks with gosec
- `./scripts/check_all.sh`: Runs all checks including tests
- `./scripts/version.sh`: Helps with versioning (run with `bump` argument to bump version)

### Running Tests

Tests are located in the `test` directory and can be run with:

```bash
cd test && make test
```

Or use the check_all script to run all checks including tests:

```bash
./scripts/check_all.sh
```

### CI/CD

The project uses GitHub Actions for CI/CD. The workflow is defined in `.github/workflows/ci.yml`.

</details>

## 👥 Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for details.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](https://github.com/Muhammadhamd/agent-sdk-go/blob/main/LICENSE) file for details.

## 🙏 Acknowledgements

This project is inspired by [OpenAI's Assistants API](https://platform.openai.com/docs/assistants/overview) and [OpenAI's Python Agent SDK](https://github.com/openai/openai-agents-py), with the goal of providing similar capabilities in Go while being compatible with local LLMs.

## ☁️ Cloud Support

For production deployments, we're developing a fully managed cloud service. Join our waitlist to be among the first to access:

- **Managed Agent Deployment** - Deploy agents without infrastructure hassle
- **Horizontal Scaling** - Handle any traffic volume
- **Observability & Monitoring** - Track performance and usage
- **Cost Optimization** - Pay only for what you use
- **Enterprise Security** - SOC2 compliance and data protection

**[Sign up for the Cloud Waitlist →](https://go-agent.org/#waitlist)**

## 👥 Community & Support

- **Website**: [go-agent.org](https://go-agent.org)
- **GitHub Issues**: [Report bugs or request features](https://github.com/Muhammadhamd/agent-sdk-go/issues)
- **Discussions**: [Join the conversation](https://github.com/Muhammadhamd/agent-sdk-go/discussions)
- **Waitlist**: [Join the cloud service waitlist](https://go-agent.org/#waitlist) 