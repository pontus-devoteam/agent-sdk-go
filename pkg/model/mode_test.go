package model_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/model"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/model/providers/anthropic"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/model/providers/openai"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/runner"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/tool"
)

func setupMultiAgent(providerName string) (*runner.Runner, *agent.Agent, model.Provider) {
	var (
		modelName string
		provider  model.Provider
	)
	switch providerName {
	case "openai":
		p := openai.NewProvider(os.Getenv("OPENAI_API_KEY"))
		modelName = os.Getenv("OPENAI_MODEL")
		p.SetDefaultModel(modelName)
		provider = p
	case "anthropic":
		p := anthropic.NewProvider(os.Getenv("ANTHROPIC_API_KEY"))
		modelName = os.Getenv("ANTHROPIC_MODEL")
		p.SetDefaultModel(modelName)
		provider = p
	}

	getTodaysDate := tool.NewFunctionTool(
		"get_todays_date",
		"Get today's date",
		func(ctx context.Context, params map[string]any) (any, error) {
			return time.Now().Format("2006/01/02"), nil
		},
	)
	getTomorrowsDate := tool.NewFunctionTool(
		"get_tomorrows_date",
		"Get tomorrow's date",
		func(ctx context.Context, params map[string]any) (any, error) {
			return time.Now().AddDate(0, 0, 1).Format("2006/01/02"), nil
		},
	)

	weather := tool.NewFunctionTool(
		"get_weather",
		"Return weather information for the specified region.",
		func(ctx context.Context, params map[string]any) (any, error) {
			day, _ := params["day"].(string)
			location, _ := params["region"].(string)
			return fmt.Sprintf(`{"date":"%s","region":"%s","temperature":24,"weather":"fine"}`, day, location), nil
		},
	).WithSchema(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"day": map[string]any{
				"type":        "string",
				"description": "desired date",
			},
			"region": map[string]any{
				"type":        "string",
				"description": "specified region",
			},
		},
		"required": []string{"day", "region"},
	})
	weatherAgent := agent.NewAgent("WeatherAgent")
	weatherAgent.SetModelProvider(provider)
	weatherAgent.WithModel(modelName)
	weatherAgent.WithTools(weather)

	assistant := agent.NewAgent("Assistant")
	assistant.SetModelProvider(provider)
	assistant.WithModel(modelName)
	assistant.WithHandoffs(weatherAgent)
	assistant.WithTools(getTodaysDate, getTomorrowsDate)
	r := runner.NewRunner()
	r.WithDefaultProvider(provider)
	return r, assistant, provider
}

func TestRunStreaming(t *testing.T) {
	ctx := context.Background()
	r, assistant, _ := setupMultiAgent("azure")
	result, err := r.RunStreaming(ctx, assistant, &runner.RunOptions{
		Input: `Please tell me today's weather in Japan.(use get_todays_date from tools.)`,
		RunConfig: &runner.RunConfig{
			TracingDisabled: true,
		},
	})
	if err != nil {
		log.Fatalf("Error running agent: %v", err)
	}

	for st := range result.Stream {
		switch st.Type {
		case model.StreamEventTypeError:
			log.Fatalf("Error stream: %v", st.Error)
		}
		if result.IsComplete {
			break
		}
	}
	fmt.Println(result.FinalOutput)
}
