package anthropic

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/model"
)

const (
	// DefaultBaseURL is the default base URL for the Anthropic API
	DefaultBaseURL = "https://api.anthropic.com/v1"

	// DefaultRPM is the default rate limit for requests per minute
	DefaultRPM = 100

	// DefaultTPM is the default rate limit for tokens per minute
	DefaultTPM = 100000

	// DefaultMaxRetries is the default number of retries for rate limited requests
	DefaultMaxRetries = 5

	// DefaultRetryAfter is the default time to wait before retrying a rate limited request
	DefaultRetryAfter = 1 * time.Second
)

// Provider implements model.Provider for Anthropic
type Provider struct {
	// Configuration
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client

	// Model configuration
	DefaultModel string

	// Message history configuration
	MaxHistoryMessages  int  // Maximum number of previous messages to include
	IncludeToolMessages bool // Whether to include tool result messages in the history limit

	// Rate limiting configuration
	RPM        int           // Requests per minute
	TPM        int           // Tokens per minute
	MaxRetries int           // Maximum number of retries
	RetryAfter time.Duration // Time to wait before retrying

	// Internal state
	mu            sync.RWMutex
	requestCount  int
	tokenCount    int
	lastResetTime time.Time
	rateLimiter   *time.Ticker
}

// NewAnthropicProvider creates a new Provider with default settings
func NewAnthropicProvider(apiKey string) *Provider {
	return &Provider{
		BaseURL: DefaultBaseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		RPM:           DefaultRPM,
		TPM:           DefaultTPM,
		MaxRetries:    DefaultMaxRetries,
		RetryAfter:    DefaultRetryAfter,
		lastResetTime: time.Now(),
		rateLimiter:   time.NewTicker(time.Minute / time.Duration(DefaultRPM)),
	}
}

// WithAPIKey sets the API key for the provider
func (p *Provider) WithAPIKey(apiKey string) *Provider {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.APIKey = apiKey
	return p
}

// WithHTTPClient sets the HTTP client for the provider
func (p *Provider) WithHTTPClient(client *http.Client) *Provider {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.HTTPClient = client
	return p
}

// WithDefaultModel sets the default model for the provider
func (p *Provider) WithDefaultModel(modelName string) *Provider {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.DefaultModel = modelName
	return p
}

// WithRateLimit sets the rate limit configuration for the provider
func (p *Provider) WithRateLimit(rpm, tpm int) *Provider {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.RPM = rpm
	p.TPM = tpm
	p.rateLimiter = time.NewTicker(time.Minute / time.Duration(rpm))
	return p
}

// WithRetryConfig sets the retry configuration for the provider
func (p *Provider) WithRetryConfig(maxRetries int, retryAfter time.Duration) *Provider {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.MaxRetries = maxRetries
	p.RetryAfter = retryAfter
	return p
}

// SetBaseURL sets the base URL for the provider
func (p *Provider) SetBaseURL(baseURL string) *Provider {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.BaseURL = baseURL
	return p
}

// SetDefaultModel sets the default model for the provider
func (p *Provider) SetDefaultModel(modelName string) *Provider {
	return p.WithDefaultModel(modelName)
}

// WithMaxHistoryMessages sets the maximum number of previous messages to include in each request
func (p *Provider) WithMaxHistoryMessages(maxMessages int) *Provider {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.MaxHistoryMessages = maxMessages
	return p
}

// WithToolMessagesInHistory configures whether tool messages count toward the history limit
func (p *Provider) WithToolMessagesInHistory(include bool) *Provider {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.IncludeToolMessages = include
	return p
}

// GetModel returns a model by name
func (p *Provider) GetModel(name string) (model.Model, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// If no name is provided, use the default model
	if name == "" {
		if p.DefaultModel == "" {
			return nil, fmt.Errorf("no model name provided and no default model set")
		}
		name = p.DefaultModel
	}

	// Check if API key is set
	if p.APIKey == "" {
		return nil, fmt.Errorf("no API key provided")
	}

	// Create a new model
	return &Model{
		ModelName:           name,
		Provider:            p,
		MaxHistoryMessages:  p.MaxHistoryMessages,
		IncludeToolMessages: p.IncludeToolMessages,
	}, nil
}

// WaitForRateLimit waits for the rate limiter to allow a new request
func (p *Provider) WaitForRateLimit() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if we need to reset the rate limiter
	now := time.Now()
	if now.Sub(p.lastResetTime) > time.Minute {
		p.resetRateLimiter()
	}

	// Wait for the rate limiter
	<-p.rateLimiter.C

	// Increment the request count
	p.requestCount++
}

// UpdateTokenCount updates the token count for rate limiting
func (p *Provider) UpdateTokenCount(tokens int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if we need to reset the rate limiter
	now := time.Now()
	if now.Sub(p.lastResetTime) > time.Minute {
		p.resetRateLimiter()
	}

	// Update the token count
	p.tokenCount += tokens
}

// resetRateLimiter resets the rate limiter
func (p *Provider) resetRateLimiter() {
	p.requestCount = 0
	p.tokenCount = 0
	p.lastResetTime = time.Now()
	p.rateLimiter.Reset(time.Minute / time.Duration(p.RPM))
}

// NewProvider creates a new provider with default settings
func NewProvider(apiKey string) *Provider {
	return NewAnthropicProvider(apiKey)
}
