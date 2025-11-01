package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Muhammadhamd/agent-sdk-go/pkg/agent"
	"github.com/Muhammadhamd/agent-sdk-go/pkg/model/providers/openai"
	"github.com/Muhammadhamd/agent-sdk-go/pkg/runner"
	"github.com/Muhammadhamd/agent-sdk-go/pkg/tool"
)

// Sample research topic
const researchTopic = "Quantum Computing Advancements in 2023"

func main() {
	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Create an OpenAI provider
	provider := openai.NewProvider(apiKey)

	// Set GPT model as the default model
	provider.SetDefaultModel("gpt-3.5-turbo")

	// Configure retry settings
	provider.WithRetryConfig(3, 2*time.Second)

	fmt.Println("Provider configured with:")
	fmt.Println("- Model:", "gpt-3.5-turbo")
	fmt.Println("- Max retries:", 3)

	// Create tools
	getCurrentTime := tool.NewFunctionTool(
		"get_current_time",
		"Get the current time",
		func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"current_time": time.Now().Format(time.RFC3339),
				"timezone":     time.Now().Location().String(),
			}, nil
		},
	)

	searchTool := tool.NewFunctionTool(
		"search_information",
		"Search for information on a specific topic",
		func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			topic, ok := params["topic"].(string)
			if !ok {
				return nil, fmt.Errorf("topic parameter is required")
			}

			// Simulate searching for information
			time.Sleep(1 * time.Second)

			// Return simulated search results
			return map[string]interface{}{
				"search_results": fmt.Sprintf("Simulated search results for: %s", topic),
				"sources": []string{
					"https://example.com/research/quantum-computing",
					"https://example.org/papers/2023-quantum-advancements",
				},
				"timestamp": time.Now().Format(time.RFC3339),
			}, nil
		},
	).WithSchema(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"topic": map[string]interface{}{
				"type":        "string",
				"description": "The topic to search for information on",
			},
		},
		"required": []string{"topic"},
	})

	// Create specialized agents
	researchAgent := createResearchAgent(provider, searchTool, getCurrentTime)
	summaryAgent := createSummaryAgent(provider, getCurrentTime)
	factCheckAgent := createFactCheckAgent(provider, searchTool, getCurrentTime)

	// Create the orchestrator agent
	orchestratorAgent := agent.NewAgent("Orchestrator")
	orchestratorAgent.SetModelProvider(provider)
	orchestratorAgent.WithModel("gpt-3.5-turbo")
	orchestratorAgent.WithTools(getCurrentTime)

	// Add system instructions and configure as task delegator
	orchestratorAgent.SetSystemInstructions(`You are an orchestrator agent that coordinates a research workflow using bidirectional flow.
Your job is to manage the workflow by delegating tasks to specialized agents and processing their returns.

WORKFLOW PROCESS:
1. Delegate to the ResearchAgent to gather information on the topic
2. When the ResearchAgent returns, delegate to the FactCheckAgent to verify the information
3. When the FactCheckAgent returns, delegate to the SummaryAgent to create a final summary
4. After the SummaryAgent returns, present the final summary to the user as your final output

IMPORTANT:
- Always provide task IDs when delegating and track which tasks have been completed
- When an agent returns to you, check the task ID to determine the next steps in the workflow
- After all steps are complete, your final response should be a comprehensive presentation of the findings
- ALWAYS end with a clear, complete summary of the research results when the workflow is finished`)

	// Configure as task delegator with explicit delegator name
	fmt.Println("Configuring orchestrator as task delegator...")
	orchestratorAgent.AsTaskDelegator()

	// Set up bidirectional handoffs
	orchestratorAgent.WithBidirectionalHandoffs(
		researchAgent,
		summaryAgent,
		factCheckAgent,
	)

	// Add extra bidirectional configuration for child agents to ensure they know their delegator
	for _, childAgent := range []*agent.Agent{researchAgent, summaryAgent, factCheckAgent} {
		// Add return_to_delegator as a special handoff
		returnAgent := agent.NewAgent("return_to_delegator", "Special agent that returns to the orchestrator")
		childAgent.WithHandoffs(returnAgent)

		// Explicitly set orchestrator as a handoff for the child
		childAgent.WithHandoffs(orchestratorAgent)
	}

	// Create runner
	r := runner.NewRunner()
	r.WithDefaultProvider(provider)

	// Run the workflow
	fmt.Println("\nStarting the research workflow with bidirectional flow...")

	// Enable all debugging
	if err := os.Setenv("DEBUG", "1"); err != nil {
		log.Printf("Warning: Failed to set DEBUG environment variable: %v", err)
	}
	if err := os.Setenv("OPENAI_DEBUG", "1"); err != nil {
		log.Printf("Warning: Failed to set OPENAI_DEBUG environment variable: %v", err)
	}

	// Print debug info about the agents
	fmt.Printf("DEBUG: Orchestrator agent has %d handoffs configured\n", len(orchestratorAgent.Handoffs))
	for i, h := range orchestratorAgent.Handoffs {
		fmt.Printf("DEBUG: Handoff #%d: %s\n", i+1, h.Name)
	}

	// Run the workflow with a simpler approach that just tracks handoffs
	result, err := r.RunSync(orchestratorAgent, &runner.RunOptions{
		Input:    fmt.Sprintf("I need comprehensive research on %s. Please coordinate the research process.", researchTopic),
		MaxTurns: 5, // Use fewer turns for debugging
	})

	if err != nil {
		log.Fatalf("Error running agent: %v", err)
	}

	// Print a summary of what happened
	fmt.Println("\nWorkflow complete! Summary:")

	// Handle nil FinalOutput by providing a fallback message
	if result.FinalOutput == nil {
		fmt.Println("- Final output: (No final output generated)")
	} else {
		fmt.Printf("- Final output: %v\n", result.FinalOutput)
	}

	fmt.Printf("- Last agent: %s\n", result.LastAgent.Name)
	fmt.Printf("- Items generated: %d\n", len(result.NewItems))

	// Print details of any handoff items
	fmt.Println("\nHandoffs:")
	handoffCount := 0
	for _, item := range result.NewItems {
		if item.GetType() == "handoff" {
			handoffCount++
			handoffItem, ok := item.(interface{ GetAgentName() string })
			if ok {
				fmt.Printf("- Handoff #%d: %s\n", handoffCount, handoffItem.GetAgentName())
			} else {
				fmt.Printf("- Handoff #%d: (agent name not available)\n", handoffCount)
			}
		}
	}

	if handoffCount == 0 {
		fmt.Println("- No handoffs occurred")
	}
}

// Create the research agent
func createResearchAgent(provider *openai.Provider, searchTool, timeTool tool.Tool) *agent.Agent {
	researchAgent := agent.NewAgent("ResearchAgent")
	researchAgent.SetModelProvider(provider)
	researchAgent.WithModel("gpt-3.5-turbo")
	researchAgent.WithTools(searchTool, timeTool)

	// Add system instructions and configure as task executor
	researchAgent.SetSystemInstructions(`You are a research agent that specializes in gathering information on specific topics.

Your job is to:
1. Research the given topic thoroughly using the search_information tool
2. Organize the gathered information in a structured format
3. Include sources and references
4. Return your findings to the agent that delegated the task to you

When completing your research:
- Use multiple search queries to gather comprehensive information
- Organize information by subtopics
- Include both general overview and specific technical details
- Always cite sources for all information

When you complete your task, return to the delegating agent with your research findings.`)
	researchAgent.AsTaskExecutor()

	return researchAgent
}

// Create the summary agent
func createSummaryAgent(provider *openai.Provider, timeTool tool.Tool) *agent.Agent {
	summaryAgent := agent.NewAgent("SummaryAgent")
	summaryAgent.SetModelProvider(provider)
	summaryAgent.WithModel("gpt-3.5-turbo")
	summaryAgent.WithTools(timeTool)

	// Add system instructions and configure as task executor
	summaryAgent.SetSystemInstructions(`You are a summary agent that specializes in creating concise yet comprehensive summaries.

Your job is to:
1. Take research information and fact-check results specifically about the exact topic you were asked to summarize
2. Synthesize the information into a coherent summary strictly related to the given topic only
3. Organize the summary in a clear, reader-friendly format
4. Return your summary to the agent that delegated the task to you

When creating summaries:
- IMPORTANT: Stay STRICTLY focused on the original topic you were asked to summarize - never switch to a different topic
- Start with a high-level overview of the topic (and only that specific topic)
- Include key facts and important details about the exact topic
- Organize with headings and bullet points for readability
- Maintain accuracy while simplifying complex concepts
- Include a "Key Takeaways" section at the end

When you complete your task, return to the delegating agent with your summary that is strictly on the requested topic.`)
	summaryAgent.AsTaskExecutor()

	return summaryAgent
}

// Create the fact check agent
func createFactCheckAgent(provider *openai.Provider, searchTool, timeTool tool.Tool) *agent.Agent {
	factCheckAgent := agent.NewAgent("FactCheckAgent")
	factCheckAgent.SetModelProvider(provider)
	factCheckAgent.WithModel("gpt-3.5-turbo")
	factCheckAgent.WithTools(searchTool, timeTool)

	// Add system instructions and configure as task executor
	factCheckAgent.SetSystemInstructions(`You are a fact-checking agent that verifies information for accuracy.

Your job is to:
1. Review research information provided to you
2. Verify key claims and statistics using the search_information tool
3. Identify any inaccuracies or questionable information
4. Provide corrections and additional context where needed
5. Return your verification results to the agent that delegated the task to you

When fact-checking:
- Focus on verifying specific factual claims
- Check dates, statistics, and quoted statements for accuracy
- Look for consensus across multiple sources
- Rate claims on a scale from "Verified" to "Incorrect"
- Provide explanation for any corrections or clarifications

When you complete your task, return to the delegating agent with your fact-check results.`)
	factCheckAgent.AsTaskExecutor()

	return factCheckAgent
}
