// Package gemini provides a client for interacting with Google's Gemini API.
package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Common errors
var (
	ErrAPIKeyMissing = errors.New("API key is required")
	ErrRateLimited   = errors.New("rate limited")
	ErrQuotaExceeded = errors.New("quota exceeded")
)

// GeminiRequest represents a request to the Gemini API
type GeminiRequest struct {
	Contents []Content `json:"contents"`
}

// Content represents a content item in the request
type Content struct {
	Parts []Part `json:"parts"`
}

// Part represents a part of content
type Part struct {
	Text string `json:"text"`
}

// GeminiResponse represents a response from the Gemini API
type GeminiResponse struct {
	Candidates     []Candidate     `json:"candidates"`
	PromptFeedback *PromptFeedback `json:"promptFeedback,omitempty"`
}

// Candidate represents a response candidate
type Candidate struct {
	Content       Content        `json:"content"`
	FinishReason  string         `json:"finishReason"`
	Index         int            `json:"index"`
	SafetyRatings []SafetyRating `json:"safetyRatings,omitempty"`
}

// SafetyRating represents safety rating information
type SafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

// PromptFeedback represents feedback about the prompt
type PromptFeedback struct {
	SafetyRatings []SafetyRating `json:"safetyRatings"`
}

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
	var lastError error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		response, err := c.makeRequest(ctx, prompt)
		if err == nil {
			return response, nil
		}

		lastError = err

		// Don't retry on certain errors
		if errors.Is(err, ErrRateLimited) || errors.Is(err, ErrQuotaExceeded) {
			break
		}

		// If this was the last attempt, don't retry
		if attempt == c.maxRetries {
			break
		}

		// Wait before retry with exponential backoff
		backoff := time.Duration(1<<attempt) * time.Second
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(backoff):
			// Continue to next attempt
		}
	}

	return "", fmt.Errorf("failed after %d attempts, last error: %w", c.maxRetries+1, lastError)
}

// makeRequest makes a single request to the Gemini API
func (c *Client) makeRequest(ctx context.Context, prompt string) (string, error) {
	// Prepare request
	request := GeminiRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: prompt},
				},
			},
		},
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s?key=%s", c.baseURL, c.apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle HTTP status codes
	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusTooManyRequests:
			return "", ErrRateLimited
		case http.StatusForbidden:
			return "", ErrQuotaExceeded
		default:
			return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
		}
	}

	// Parse response
	var geminiResponse GeminiResponse
	if err := json.Unmarshal(body, &geminiResponse); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract text from response
	if len(geminiResponse.Candidates) == 0 {
		return "", fmt.Errorf("no candidates in response")
	}

	candidate := geminiResponse.Candidates[0]
	if len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("no content parts in candidate")
	}

	return candidate.Content.Parts[0].Text, nil
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
