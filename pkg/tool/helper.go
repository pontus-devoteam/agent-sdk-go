package tool

// ToOpenAITool converts a Tool to the OpenAI tool format
func ToOpenAITool(tool Tool) map[string]interface{} {
	return map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        tool.GetName(),
			"description": tool.GetDescription(),
			"parameters":  tool.GetParametersSchema(),
		},
	}
}

// ToOpenAITools converts a slice of Tools to the OpenAI tool format
func ToOpenAITools(tools []Tool) []map[string]interface{} {
	result := make([]map[string]interface{}, len(tools))
	for i, tool := range tools {
		result[i] = ToOpenAITool(tool)
	}
	return result
}

// CreateToolFromDefinition creates a Tool from an OpenAI tool definition
func CreateToolFromDefinition(definition map[string]interface{}, executeFn func(map[string]interface{}) (interface{}, error)) Tool {
	// Extract function details
	functionDef := definition["function"].(map[string]interface{})
	name := functionDef["name"].(string)
	description := functionDef["description"].(string)
	parameters := functionDef["parameters"].(map[string]interface{})

	// Create a new function tool
	tool := NewFunctionTool(name, description, func(ctx interface{}, params map[string]interface{}) (interface{}, error) {
		return executeFn(params)
	})

	// Set the schema
	tool.WithSchema(parameters)

	return tool
}
