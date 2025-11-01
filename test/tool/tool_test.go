package tool_test

import (
	"context"
	"testing"

	"github.com/Muhammadhamd/agent-sdk-go/pkg/tool"
)

// TestNewFunctionTool tests the creation of a new function tool
func TestNewFunctionTool(t *testing.T) {
	// Create a new function tool
	name := "test_tool"
	description := "Test tool for testing"

	testTool := tool.NewFunctionTool(
		name,
		description,
		func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return "test result", nil
		},
	)

	// Check if the tool was created correctly
	if testTool == nil {
		t.Fatalf("NewFunctionTool(%s, %s, func) returned nil", name, description)
	}

	// Check tool name
	if testTool.GetName() != name {
		t.Errorf("Tool name = %s, want %s", testTool.GetName(), name)
	}

	// Check tool description
	if testTool.GetDescription() != description {
		t.Errorf("Tool description = %s, want %s", testTool.GetDescription(), description)
	}
}

// TestFunctionToolExecution tests the execution of a function tool
func TestFunctionToolExecution(t *testing.T) {
	// Create a tool that adds two numbers
	addTool := tool.NewFunctionTool(
		"add",
		"Add two numbers",
		func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			a, ok := params["a"].(float64)
			if !ok {
				t.Fatalf("Parameter 'a' is not a float64")
			}

			b, ok := params["b"].(float64)
			if !ok {
				t.Fatalf("Parameter 'b' is not a float64")
			}

			return a + b, nil
		},
	)

	// Test executing the tool
	ctx := context.Background()
	params := map[string]interface{}{
		"a": float64(5),
		"b": float64(3),
	}

	result, err := addTool.Execute(ctx, params)

	// Check for errors
	if err != nil {
		t.Errorf("Tool execution returned error: %v", err)
	}

	// Check the result
	expectedResult := float64(8)
	if result != expectedResult {
		t.Errorf("Tool execution result = %v, want %v", result, expectedResult)
	}
}

// TestWithSchema tests adding a schema to a function tool
func TestWithSchema(t *testing.T) {
	// Create a tool
	testTool := tool.NewFunctionTool(
		"test_tool",
		"Test tool",
		func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return nil, nil
		},
	)

	// Add schema
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "The name parameter",
			},
			"age": map[string]interface{}{
				"type":        "integer",
				"description": "The age parameter",
			},
		},
		"required": []string{"name"},
	}

	toolWithSchema := testTool.WithSchema(schema)

	// Check if schema was added correctly
	if toolWithSchema == nil {
		t.Fatalf("WithSchema() returned nil")
	}

	// Get the schema
	actualSchema := toolWithSchema.GetParametersSchema()

	// Check if the schema has the expected properties
	props, ok := actualSchema["properties"].(map[string]interface{})
	if !ok {
		t.Fatalf("Schema does not have properties field")
	}

	nameProp, ok := props["name"].(map[string]interface{})
	if !ok {
		t.Fatalf("Properties does not have name field")
	}

	nameType, ok := nameProp["type"].(string)
	if !ok || nameType != "string" {
		t.Errorf("name.type = %v, want string", nameProp["type"])
	}

	ageProp, ok := props["age"].(map[string]interface{})
	if !ok {
		t.Fatalf("Properties does not have age field")
	}

	ageType, ok := ageProp["type"].(string)
	if !ok || ageType != "integer" {
		t.Errorf("age.type = %v, want integer", ageProp["type"])
	}
}

// TestToOpenAITool tests converting a tool to OpenAI format
func TestToOpenAITool(t *testing.T) {
	// Create a tool
	testTool := tool.NewFunctionTool(
		"test_tool",
		"Test tool",
		func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return nil, nil
		},
	).WithSchema(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "The search query",
			},
		},
		"required": []string{"query"},
	})

	// Convert to OpenAI format
	openAITool := tool.ToOpenAITool(testTool)

	// Check if conversion was successful
	if openAITool == nil {
		t.Fatalf("ToOpenAITool() returned nil")
	}

	// Check type
	toolType, ok := openAITool["type"].(string)
	if !ok || toolType != "function" {
		t.Errorf("Tool type = %v, want 'function'", toolType)
	}

	// Check function
	function, ok := openAITool["function"].(map[string]interface{})
	if !ok {
		t.Fatalf("Tool does not have a function field")
	}

	// Check function name
	name, ok := function["name"].(string)
	if !ok || name != "test_tool" {
		t.Errorf("Function name = %v, want 'test_tool'", name)
	}

	// Check function description
	description, ok := function["description"].(string)
	if !ok || description != "Test tool" {
		t.Errorf("Function description = %v, want 'Test tool'", description)
	}

	// Check parameters
	parameters, ok := function["parameters"].(map[string]interface{})
	if !ok {
		t.Fatalf("Function does not have parameters")
	}

	// Check if parameters contain the correct schema
	properties, ok := parameters["properties"].(map[string]interface{})
	if !ok {
		t.Fatalf("Parameters do not have properties")
	}

	query, ok := properties["query"].(map[string]interface{})
	if !ok {
		t.Fatalf("Properties do not have query")
	}

	queryType, ok := query["type"].(string)
	if !ok || queryType != "string" {
		t.Errorf("Query type = %v, want 'string'", queryType)
	}
}

// TestToOpenAITools tests converting multiple tools to OpenAI format
func TestToOpenAITools(t *testing.T) {
	// Create tools
	tool1 := tool.NewFunctionTool(
		"tool1",
		"Tool 1",
		func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return nil, nil
		},
	)

	tool2 := tool.NewFunctionTool(
		"tool2",
		"Tool 2",
		func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return nil, nil
		},
	)

	// Convert to OpenAI format
	openAITools := tool.ToOpenAITools([]tool.Tool{tool1, tool2})

	// Check if conversion was successful
	if openAITools == nil {
		t.Fatalf("ToOpenAITools() returned nil")
	}

	// Check number of tools
	if len(openAITools) != 2 {
		t.Errorf("Got %d tools, want 2", len(openAITools))
	}

	// Check first tool
	firstTool := openAITools[0]
	function, ok := firstTool["function"].(map[string]interface{})
	if !ok {
		t.Fatalf("First tool does not have a function field")
	}

	name, ok := function["name"].(string)
	if !ok || name != "tool1" {
		t.Errorf("First tool name = %v, want 'tool1'", name)
	}

	// Check second tool
	secondTool := openAITools[1]
	function, ok = secondTool["function"].(map[string]interface{})
	if !ok {
		t.Fatalf("Second tool does not have a function field")
	}

	name, ok = function["name"].(string)
	if !ok || name != "tool2" {
		t.Errorf("Second tool name = %v, want 'tool2'", name)
	}
}
