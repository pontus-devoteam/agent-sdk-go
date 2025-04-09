# TypeScript Code Review Example

This example demonstrates a bidirectional agent flow for developing and reviewing TypeScript code. It showcases how multiple AI agents can collaborate to create high-quality code through an iterative review process.

## Workflow

The example implements the following workflow:

1. The Orchestrator agent receives a request to create a TypeScript function
2. The Orchestrator delegates the task to the Coder agent
3. The Coder agent implements the function
4. The Orchestrator passes the code to the Reviewer agent
5. The Reviewer agent evaluates the code and provides feedback
6. If changes are needed, the Orchestrator sends the code back to the Coder with feedback
7. The cycle repeats until the Reviewer approves the code
8. The Orchestrator presents the final, approved code to the user

## Bidirectional Flow

This example demonstrates bidirectional agent flow where:
- The Orchestrator can delegate to both the Coder and Reviewer
- Both agents can return results back to the Orchestrator
- The Orchestrator manages the workflow by tracking task completion and delegating follow-up tasks

## Running the Example

1. Set your OpenAI API key:

```bash
export OPENAI_API_KEY=your_api_key_here
```

2. Build and run the example:

```bash
cd examples/typescript_code_review_example
go build
./typescript_code_review_example
```

## What to Expect

When you run the example, you'll see:
- Debug information showing the configuration of the agents
- A series of handoffs between the agents as they collaborate
- Logs showing the code implementation and review process
- A final output with the approved TypeScript function

The example will produce log files in the current directory:
- `trace_Orchestrator.log` - Shows the Orchestrator agent's actions
- `trace_CoderAgent.log` - Shows the code implementation process
- `trace_ReviewerAgent.log` - Shows the code review process

## Key Features

- **Simulated TypeScript Validation**: The example includes a simulated TypeScript code validator that checks for basic issues
- **Iterative Review Process**: The code goes through multiple rounds of review and improvement
- **Task Tracking**: The Orchestrator keeps track of tasks and their completion status
- **Bidirectional Handoffs**: Agents can return to their delegator with results

## Customization

You can modify the `functionRequirement` constant in `main.go` to request different types of TypeScript functions. 