// Package gemini provides a client for interacting with Google's Gemini API.
package gemini

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// Common errors
var (
	ErrAPIKeyMissing = errors.New("API key is required")
	ErrRateLimited   = errors.New("rate limited")
	ErrQuotaExceeded = errors.New("quota exceeded")
)

// Client represents a Gemini API client
type Client struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
	timeout    time.Duration
	maxRetries int
}

// NewClient creates a new Gemini client
func NewClient(apiKey string, options ...Option) *Client {
	if apiKey == "" {
		panic("API key is required")
	}

	c := &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-exp:generateContent",
		timeout:    30 * time.Second,
		maxRetries: 3,
	}

	for _, option := range options {
		option(c)
	}

	return c
}

// Generate generates content using the Gemini API
func (c *Client) Generate(prompt string) (string, error) {
	return c.GenerateWithContext(context.Background(), prompt)
}

// GenerateWithContext generates content with a context
func (c *Client) GenerateWithContext(ctx context.Context, prompt string) (string, error) {
	// TODO: Implement actual API call to Gemini
	// This is a placeholder implementation
	return "TODO: Implement Gemini API integration", nil
}

// Option represents a configuration option for the Client
type Option func(*Client)

// WithTimeout sets the timeout for requests
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
		c.httpClient.Timeout = timeout
	}
}

// WithMaxRetries sets the maximum number of retries
func WithMaxRetries(maxRetries int) Option {
	return func(c *Client) {
		c.maxRetries = maxRetries
	}
}

// WithBaseURL sets the base URL for the API
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}
