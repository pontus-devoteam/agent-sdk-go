package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/model/providers/lmstudio"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/runner"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/tool"
)

// TimeTool provides functionality to get the current time in a specific format.
type TimeTool struct {
	Format string
}

func main() {
	// Enable verbose logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Create a model provider
	provider := lmstudio.NewProvider()
	provider.SetBaseURL("http://127.0.0.1:1234/v1")
	provider.SetDefaultModel("gemma-3-4b-it")

	fmt.Println("Provider configured with:")
	fmt.Println("- Base URL:", "http://127.0.0.1:1234/v1")
	fmt.Println("- Model:", "gemma-3-4b-it")

	// Create the primary agent (Frontend)
	frontendAgent := agent.NewAgent("Frontend Agent")
	frontendAgent.SetModelProvider(provider)
	frontendAgent.WithModel("gemma-3-4b-it")
	frontendAgent.SetSystemInstructions(`You are a helpful frontend agent that coordinates requests.
Your job is to understand the user's request and delegate tasks to specialized agents or use tools directly when appropriate.

IMPORTANT: For specialized tasks, you MUST delegate to the appropriate agent using the handoff mechanism:
- For mathematical calculations: DELEGATE to "Math Agent" - do NOT try to perform calculations yourself
- For weather information: DELEGATE to "Weather Agent" - do NOT try to get weather data yourself

When a user asks about:
- Any math calculation (adding, subtracting, multiplying, dividing) → handoff to "Math Agent"
- Weather conditions in any location → handoff to "Weather Agent"

You can only use the get_current_time tool directly. For all other tools, you must handoff to a specialized agent.

When you handoff to another agent, your response will be used to direct that agent. Be specific about what you're asking the specialized agent to do.

Always provide a final response to the user after receiving information from tools or specialized agents.
IMPORTANT: Never end with a tool call. Always provide a final human-readable response.`)

	// Create the math agent
	mathAgent := agent.NewAgent("Math Agent")
	mathAgent.SetModelProvider(provider)
	mathAgent.WithModel("gemma-3-4b-it")
	mathAgent.SetSystemInstructions(`You are a specialized math agent.
You excel at solving mathematical problems and performing calculations.
Use the calculation tools available to you to solve problems accurately.

IMPORTANT WORKFLOW:
1. When you receive a request, identify the mathematical operation needed
2. Use the appropriate calculation tool (calculate, generate_random_number, etc.)
3. Provide a clear, complete answer explaining the calculation and result
4. Always format your response with the calculation process AND the final answer

For example, if asked to calculate 25 divided by 5, you should:
1. Use the calculate tool with operation "divide", a=25, b=5
2. Respond with: "The calculation of 25 divided by 5 equals 5."

Always provide educational, clear responses that explain both the process and the result.
IMPORTANT: Never end with a tool call. Always provide a final human-readable response.`)

	// Create the weather agent
	weatherAgent := agent.NewAgent("Weather Agent")
	weatherAgent.SetModelProvider(provider)
	weatherAgent.WithModel("gemma-3-4b-it")
	weatherAgent.SetSystemInstructions(`You are a specialized weather agent.
You provide weather information and forecasts based on data from your tools.
Always use the available weather tools to get up-to-date information.

IMPORTANT WORKFLOW:
1. When you receive a request for weather information, identify the location
2. Use the get_weather tool with the appropriate location parameter
3. Interpret the weather data and provide a complete, human-friendly response
4. Include temperature, conditions, humidity, and any relevant context

For example, if asked about Paris weather, you should:
1. Use the get_weather tool with location "Paris"
2. Interpret the data and respond with something like:
   "Currently in Paris, it's 18°C (64°F) and partly cloudy with 65% humidity. It's a pleasant day with mild temperatures."

Always provide complete, context-rich interpretations of the weather data.
IMPORTANT: Never end with a tool call. Always provide a final human-readable response.`)

	// Create simple tools
	// Random Number Generator Tool
	randomNumberTool := tool.NewFunctionTool(
		"generate_random_number",
		"Generate a random number between min and max (inclusive)",
		func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			min := 1
			max := 100

			if minParam, ok := params["min"].(float64); ok {
				min = int(minParam)
			}
			if maxParam, ok := params["max"].(float64); ok {
				max = int(maxParam)
			}

			// Seed the random number generator
			rand.New(rand.NewSource(time.Now().UnixNano()))

			return rand.Intn(max-min+1) + min, nil
		},
	).WithSchema(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"min": map[string]interface{}{
				"type":        "integer",
				"description": "The minimum value (inclusive)",
			},
			"max": map[string]interface{}{
				"type":        "integer",
				"description": "The maximum value (inclusive)",
			},
		},
		"required": []string{},
	})

	// Simple Calculator Tool
	calculatorTool := tool.NewFunctionTool(
		"calculate",
		"Perform a simple calculation",
		func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			operation, ok := params["operation"].(string)
			if !ok {
				return nil, fmt.Errorf("operation parameter is required")
			}

			a, aOk := params["a"].(float64)
			b, bOk := params["b"].(float64)

			if !aOk || !bOk {
				return nil, fmt.Errorf("both 'a' and 'b' parameters are required and must be numbers")
			}

			switch strings.ToLower(operation) {
			case "add":
				return a + b, nil
			case "subtract":
				return a - b, nil
			case "multiply":
				return a * b, nil
			case "divide":
				if b == 0 {
					return nil, fmt.Errorf("division by zero is not allowed")
				}
				return a / b, nil
			default:
				return nil, fmt.Errorf("unknown operation: %s", operation)
			}
		},
	).WithSchema(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"operation": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"add", "subtract", "multiply", "divide"},
				"description": "The mathematical operation to perform",
			},
			"a": map[string]interface{}{
				"type":        "number",
				"description": "The first operand",
			},
			"b": map[string]interface{}{
				"type":        "number",
				"description": "The second operand",
			},
		},
		"required": []string{"operation", "a", "b"},
	})

	// Weather Tool
	weatherTool := tool.NewFunctionTool(
		"get_weather",
		"Get the current weather for a location",
		func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			location, ok := params["location"].(string)
			if !ok || location == "" {
				return nil, fmt.Errorf("location parameter is required")
			}

			// This is a mock implementation
			weatherConditions := []string{
				"sunny", "partly cloudy", "cloudy", "rainy", "stormy", "snowy", "foggy", "windy",
			}
			temperatures := []int{-10, 0, 5, 10, 15, 20, 25, 30, 35}

			// Seed the random generator
			r := rand.New(rand.NewSource(time.Now().UnixNano()))

			// Generate mock weather data
			condition := weatherConditions[r.Intn(len(weatherConditions))]
			temperature := temperatures[r.Intn(len(temperatures))]
			humidity := r.Intn(100)

			return map[string]interface{}{
				"location":    location,
				"condition":   condition,
				"temperature": temperature,
				"humidity":    humidity,
				"unit":        "celsius",
				"timestamp":   time.Now().Format(time.RFC3339),
			}, nil
		},
	).WithSchema(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"location": map[string]interface{}{
				"type":        "string",
				"description": "The location to get weather for (city name)",
			},
		},
		"required": []string{"location"},
	})

	// Time Tool (shared by all agents)
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

	// Add tools to agents - using consistent approach with WithTools
	frontendAgent.WithTools(timeTool)

	mathAgent.WithTools(calculatorTool)
	mathAgent.WithTools(randomNumberTool)
	mathAgent.WithTools(timeTool)

	weatherAgent.WithTools(weatherTool)
	weatherAgent.WithTools(timeTool)

	// Set up agent handoffs
	frontendAgent.WithHandoffs(mathAgent, weatherAgent)

	// Create a runner
	r := runner.NewRunner()
	r.WithDefaultProvider(provider)

	// Run example with a math query
	fmt.Println("Running with a math query...")
	result, err := r.RunSync(frontendAgent, &runner.RunOptions{
		Input:    "What is 42 divided by 6?",
		MaxTurns: 20,
	})
	if err != nil {
		log.Fatalf("Error running agent: %v", err)
	}

	// Print the result
	fmt.Println("\nAgent response:")
	fmt.Println(result.FinalOutput)
	fmt.Println("\nItems generated:", len(result.NewItems))

	// Print detailed items for debugging
	fmt.Println("\nDetailed items:")
	for i, item := range result.NewItems {
		fmt.Printf("Item %d: Type=%s\n", i, item.GetType())
	}

	// Run example with a weather query
	fmt.Println("\nRunning with a weather query...")
	result, err = r.RunSync(frontendAgent, &runner.RunOptions{
		Input:    "What's the current weather in Paris?",
		MaxTurns: 20,
	})
	if err != nil {
		log.Fatalf("Error running agent: %v", err)
	}

	// Print the result
	fmt.Println("\nAgent response:")
	fmt.Println(result.FinalOutput)
	fmt.Println("\nItems generated:", len(result.NewItems))

	// Print detailed items for debugging
	fmt.Println("\nDetailed items:")
	for i, item := range result.NewItems {
		fmt.Printf("Item %d: Type=%s\n", i, item.GetType())
	}

	// Run example with a mixed query
	fmt.Println("\nRunning with a mixed query...")
	result, err = r.RunSync(frontendAgent, &runner.RunOptions{
		Input:    "What is 15 × 4 and what's the current time?",
		MaxTurns: 20,
	})
	if err != nil {
		log.Fatalf("Error running agent: %v", err)
	}

	// Print the result
	fmt.Println("\nAgent response:")
	fmt.Println(result.FinalOutput)
	fmt.Println("\nItems generated:", len(result.NewItems))

	// Print detailed items for debugging
	fmt.Println("\nDetailed items:")
	for i, item := range result.NewItems {
		fmt.Printf("Item %d: Type=%s\n", i, item.GetType())
	}
}
