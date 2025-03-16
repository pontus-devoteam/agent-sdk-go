package lmstudio

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/model"
)

const (
	// DefaultBaseURL is the default base URL for the LM Studio API
	DefaultBaseURL = "http://localhost:1234/v1"
)

// Provider implements model.Provider for LM Studio
type Provider struct {
	// Configuration
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client

	// Model configuration
	DefaultModel string

	// Internal state
	mu sync.RWMutex
}

// NewLMStudioProvider creates a new Provider with default settings
func NewLMStudioProvider(baseURL string) *Provider {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	return &Provider{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
		},
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

	// Create a new model
	return &Model{
		ModelName: name,
		Provider:  p,
	}, nil
}

// NewProvider creates a new provider with default settings
func NewProvider() *Provider {
	return NewLMStudioProvider(DefaultBaseURL)
}
