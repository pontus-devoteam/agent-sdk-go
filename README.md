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
    <a href="https://github.com/pontus-devoteam/agent-sdk-go/actions/workflows/code-quality.yml"><img src="https://github.com/pontus-devoteam/agent-sdk-go/actions/workflows/code-quality.yml/badge.svg" alt="Code Quality"></a>
    <a href="https://goreportcard.com/report/github.com/pontus-devoteam/agent-sdk-go"><img src="https://goreportcard.com/badge/github.com/pontus-devoteam/agent-sdk-go" alt="Go Report Card"></a>
    <a href="https://github.com/pontus-devoteam/agent-sdk-go/blob/main/go.mod"><img src="https://img.shields.io/github/go-mod/go-version/pontus-devoteam/agent-sdk-go" alt="Go Version"></a>
    <a href="https://pkg.go.dev/github.com/pontus-devoteam/agent-sdk-go"><img src="https://pkg.go.dev/badge/github.com/pontus-devoteam/agent-sdk-go.svg" alt="PkgGoDev"></a><br>
    <a href="https://github.com/pontus-devoteam/agent-sdk-go/actions/workflows/codeql-analysis.yml"><img src="https://github.com/pontus-devoteam/agent-sdk-go/actions/workflows/codeql-analysis.yml/badge.svg" alt="CodeQL"></a>
    <a href="https://github.com/pontus-devoteam/agent-sdk-go/blob/main/LICENSE"><img src="https://img.shields.io/github/license/pontus-devoteam/agent-sdk-go" alt="License"></a>
    <a href="https://github.com/pontus-devoteam/agent-sdk-go/stargazers"><img src="https://img.shields.io/github/stars/pontus-devoteam/agent-sdk-go" alt="Stars"></a>
    <a href="https://github.com/pontus-devoteam/agent-sdk-go/graphs/contributors"><img src="https://img.shields.io/github/contributors/pontus-devoteam/agent-sdk-go" alt="Contributors"></a>
    <a href="https://github.com/pontus-devoteam/agent-sdk-go/commits/main"><img src="https://img.shields.io/github/last-commit/pontus-devoteam/agent-sdk-go" alt="Last Commit"></a>
</p>

<p align="center">
  <a href="https://go-agent.org/docs">📚 Documentation</a> •
  <a href="https://go-agent.org/#waitlist">☁️ Cloud Waitlist</a> •
  <a href="https://github.com/pontus-devoteam/agent-sdk-go/blob/main/LICENSE">📜 License</a>
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

- ✅ **Multiple LLM Provider Support** - Support for both OpenAI and LM Studio
- ✅ **Tool Integration** - Call Go functions directly from your LLM
- ✅ **Agent Handoffs** - Create complex multi-agent workflows with specialized agents
- ✅ **Structured Output** - Parse responses into Go structs
- ✅ **Streaming** - Get real-time streaming responses
- ✅ **Tracing & Monitoring** - Debug your agent flows
- ✅ **OpenAI Compatibility** - Compatible with OpenAI tool definitions and API

## 📦 Installation

There are several ways to add this module to your project:

### Option 1: Using `go get` (Recommended)

```bash
go get github.com/pontus-devoteam/agent-sdk-go
```

### Option 2: Add to your imports and use `go mod tidy`

1. Add imports to your Go files:
   ```go
   import (
       "github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
       "github.com/pontus-devoteam/agent-sdk-go/pkg/model/providers/lmstudio"
       "github.com/pontus-devoteam/agent-sdk-go/pkg/runner"
       "github.com/pontus-devoteam/agent-sdk-go/pkg/tool"
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
require github.com/pontus-devoteam/agent-sdk-go latest
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
   go get github.com/pontus-devoteam/agent-sdk-go
   ```

### Troubleshooting

- If you encounter version conflicts, you can specify a version:
  ```bash
  go get github.com/pontus-devoteam/agent-sdk-go@v0.1.0  # Replace with desired version
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

    "github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
    "github.com/pontus-devoteam/agent-sdk-go/pkg/model/providers/openai"  // or "github.com/pontus-devoteam/agent-sdk-go/pkg/model/providers/lmstudio"
    "github.com/pontus-devoteam/agent-sdk-go/pkg/runner"
    "github.com/pontus-devoteam/agent-sdk-go/pkg/tool"
)

func main() {
    // Create a provider (OpenAI example)
    provider := openai.NewProvider()
    provider.SetAPIKey("your-openai-api-key")
    provider.SetDefaultModel("gpt-3.5-turbo")

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
    assistant.WithModel("gpt-3.5-turbo")  // or "gemma-3-4b-it" for LM Studio
    assistant.SetSystemInstructions("You are a helpful assistant.")
    assistant.WithTools(getWeather)

    // Create a runner
    runner := runner.NewRunner()
    runner.WithDefaultProvider(provider)

    // Run the agent
    result, err := runner.RunSync(assistant, &runner.RunOptions{
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
// Create a provider for LM Studio
provider := lmstudio.NewProvider()
provider.SetBaseURL("http://127.0.0.1:1234/v1")
provider.SetDefaultModel("gemma-3-4b-it")

// Set as the default provider
runner := runner.NewRunner()
runner.WithDefaultProvider(provider)
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

## 📚 Examples

The repository includes several examples to help you get started:

| Example | Description |
|---------|-------------|
| [Multi-Agent Example](./examples/multi_agent_example) | Demonstrates how to create a system of specialized agents that can collaborate on complex tasks |

To run the multi-agent example:

1. Make sure LM Studio is running with a server at `http://127.0.0.1:1234/v1`
2. Navigate to the example directory
   ```bash
   cd examples/multi_agent_example
   ```
3. Run the example
   ```bash
   go run .
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

This project is licensed under the MIT License - see the [LICENSE](https://github.com/pontus-devoteam/agent-sdk-go/blob/main/LICENSE) file for details.

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
- **GitHub Issues**: [Report bugs or request features](https://github.com/pontus-devoteam/agent-sdk-go/issues)
- **Discussions**: [Join the conversation](https://github.com/pontus-devoteam/agent-sdk-go/discussions)
- **Waitlist**: [Join the cloud service waitlist](https://go-agent.org/#waitlist) 