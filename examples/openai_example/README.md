# OpenAI Provider Example

This example demonstrates how to use the OpenAI provider with the Agent SDK Go. It shows how to configure rate limiting, retries, and handle API errors gracefully.

## Prerequisites

1. An OpenAI API key (get one from [OpenAI's platform](https://platform.openai.com/api-keys))
2. Go 1.23 or later

## Setup

1. Set your OpenAI API key as an environment variable:

```bash
export OPENAI_API_KEY="your-api-key-here"
```

2. Run the example:

```bash
go run main.go
```

## Features Demonstrated

### 1. Provider Configuration

The example shows how to create and configure an OpenAI provider with:

```go
// Create a provider
provider := openai.NewProvider(apiKey)

// Set the default model
provider.SetDefaultModel("gpt-3.5-turbo")

// Configure rate limits (requests per minute, tokens per minute)
provider.WithRateLimit(50, 100000)

// Configure retry settings (max retries, base delay)
provider.WithRetryConfig(3, 2*time.Second)
```

### 2. Rate Limit Handling

The OpenAI provider implements automatic rate limiting with:

- Token counting to stay within OpenAI's limits
- Request throttling based on requests per minute
- Automatic tracking of usage

### 3. Retry with Exponential Backoff

When rate limit errors occur, the provider automatically:

- Implements exponential backoff (delay increases with each retry)
- Adds jitter to prevent thundering herd problems
- Provides clear error messages about rate limiting

### 4. Streaming Support

The example demonstrates streaming responses, which is particularly useful for:

- Providing real-time responses to users
- Handling long-running generations
- Improving user experience with incremental output

## Understanding Rate Limits

OpenAI imposes several types of rate limits:

1. **RPM (Requests Per Minute)**: Limits how many API calls you can make in a minute
2. **TPM (Tokens Per Minute)**: Limits the total number of tokens processed in a minute

The rate limits vary by:
- Your OpenAI plan (Free, Pay-as-you-go, Enterprise)
- The model you're using (GPT-3.5 vs GPT-4)

This provider helps you stay within these limits by:
- Tracking request and token usage
- Implementing wait mechanisms when approaching limits
- Adding retry logic with exponential backoff

## Configuration Options

The OpenAI provider supports several configuration options:

| Method | Description |
|--------|-------------|
| `SetBaseURL` | Change the API endpoint (useful for proxies) |
| `SetDefaultModel` | Set the default model for requests |
| `WithAPIKey` | Set the API key |
| `WithOrganization` | Set the OpenAI organization ID |
| `WithRateLimit` | Configure RPM and TPM limits |
| `WithRetryConfig` | Configure retry attempts and backoff timing |
| `WithHTTPClient` | Use a custom HTTP client |

## Error Handling

The provider handles several types of errors:

1. **Rate limit errors**: Automatically retries with backoff
2. **Network errors**: Retries with appropriate backoff
3. **Authentication errors**: Returns clear error messages
4. **Context cancellation**: Respects context deadlines during retries

## Handling Different Models

The provider works with all OpenAI models including:
- GPT-3.5 Turbo
- GPT-4 and GPT-4 Turbo
- Future models (as they use the same API format)

Simply specify the model name:

```go
assistant.WithModel("gpt-4-turbo")
``` 