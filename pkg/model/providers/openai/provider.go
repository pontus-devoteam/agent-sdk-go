package openai

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Muhammadhamd/agent-sdk-go/pkg/model"
)

const (
	// DefaultBaseURL is the default base URL for the OpenAI API
	DefaultBaseURL = "https://api.openai.com/v1"

	// DefaultRPM is the default rate limit for requests per minute
	DefaultRPM = 200

	// DefaultTPM is the default rate limit for tokens per minute
	DefaultTPM = 150000

	// DefaultMaxRetries is the default number of retries for rate limited requests
	DefaultMaxRetries = 5

	// DefaultRetryAfter is the default time to wait before retrying a rate limited request
	DefaultRetryAfter = 1 * time.Second
)

// Provider implements model.Provider for OpenAI
type Provider struct {
	// Configuration
	BaseURL      string
	APIKey       string
	Organization string
	HTTPClient   *http.Client

	// Model configuration
	DefaultModel string

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

// NewOpenAIProvider creates a new Provider with default settings
func NewOpenAIProvider(apiKey string) *Provider {
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

// WithOrganization sets the organization for the provider
func (p *Provider) WithOrganization(org string) *Provider {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Organization = org
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
		ModelName: name,
		Provider:  p,
	}, nil
}

// WaitForRateLimit waits for the rate limiter to allow a new request
func (p *Provider) WaitForRateLimit() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Reset counters if it's been more than a minute since the last reset
	if time.Since(p.lastResetTime) >= time.Minute {
		p.requestCount = 0
		p.tokenCount = 0
		p.lastResetTime = time.Now()
	}

	// Check if we've exceeded our rate limits
	if p.requestCount >= p.RPM || p.tokenCount >= p.TPM {
		// Calculate how long to wait based on which limit was exceeded
		var waitTime time.Duration
		if p.requestCount >= p.RPM {
			waitTime = time.Minute / time.Duration(p.RPM)
		}
		if p.tokenCount >= p.TPM {
			tokenWaitTime := time.Minute / time.Duration(p.TPM)
			if tokenWaitTime > waitTime {
				waitTime = tokenWaitTime
			}
		}
		time.Sleep(waitTime)
	}

	// Increment request count
	p.requestCount++
}

// UpdateTokenCount updates the token count for rate limiting
func (p *Provider) UpdateTokenCount(tokens int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.tokenCount += tokens
}

// ResetRateLimiter resets the rate limit counters
func (p *Provider) ResetRateLimiter() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.requestCount = 0
	p.tokenCount = 0
	p.lastResetTime = time.Now()
}

// NewProvider creates a new provider with default settings, requires an API key
func NewProvider(apiKey string) *Provider {
	return NewOpenAIProvider(apiKey)
}
