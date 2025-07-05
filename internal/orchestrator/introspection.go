// Package orchestrator provides service introspection capabilities
package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Introspector handles service discovery and capability caching
type Introspector struct {
	config     *OrchestratorConfig
	cache      map[string]*ServiceCapability
	cacheMux   sync.RWMutex
	httpClient *http.Client
}

// NewIntrospector creates a new introspector instance
func NewIntrospector(config *OrchestratorConfig) *Introspector {
	return &Introspector{
		config: config,
		cache:  make(map[string]*ServiceCapability),
		httpClient: &http.Client{
			Timeout: config.RequestTimeout,
		},
	}
}

// IntrospectServices performs parallel introspection of all configured services
func (i *Introspector) IntrospectServices(ctx context.Context) (*IntrospectionResult, error) {
	startTime := time.Now()

	if len(i.config.ServiceURLs) == 0 {
		return &IntrospectionResult{
			Services:   []ServiceCapability{},
			TotalFound: 0,
			FailedURLs: []string{},
			Duration:   time.Since(startTime),
			Timestamp:  startTime,
		}, nil
	}

	// Use a semaphore to limit concurrent requests
	semaphore := make(chan struct{}, i.config.MaxConcurrency)

	var wg sync.WaitGroup
	results := make(chan *ServiceCapability, len(i.config.ServiceURLs))
	failedURLs := make(chan string, len(i.config.ServiceURLs))

	// Launch goroutines for each service URL
	for _, serviceURL := range i.config.ServiceURLs {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-ctx.Done():
				return
			}

			capability, err := i.introspectService(ctx, url)
			if err != nil {
				if i.config.EnableLogging {
					log.Printf("Failed to introspect service %s: %v", url, err)
				}
				failedURLs <- url
				return
			}

			results <- capability
		}(serviceURL)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(results)
	close(failedURLs)

	// Collect results
	var services []ServiceCapability
	var failed []string

	for capability := range results {
		services = append(services, *capability)
	}

	for failedURL := range failedURLs {
		failed = append(failed, failedURL)
	}

	// Update cache
	i.updateCache(services)

	result := &IntrospectionResult{
		Services:   services,
		TotalFound: len(services),
		FailedURLs: failed,
		Duration:   time.Since(startTime),
		Timestamp:  startTime,
	}

	if i.config.EnableLogging {
		log.Printf("Introspection completed: %d services found, %d failed, duration: %v",
			result.TotalFound, len(result.FailedURLs), result.Duration)
	}

	return result, nil
}

// introspectService performs introspection of a single service
func (i *Introspector) introspectService(ctx context.Context, serviceURL string) (*ServiceCapability, error) {
	startTime := time.Now()

	// Check cache first
	if cached := i.getFromCache(serviceURL); cached != nil && !i.isCacheExpired(cached) {
		if i.config.EnableLogging {
			log.Printf("Using cached capability for %s", serviceURL)
		}
		return cached, nil
	}

	// Prepare JSON-RPC request for Describe method
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      uuid.New().String(),
		"method":  "R0D0Service.Describe",
		"params":  []interface{}{map[string]interface{}{}},
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", serviceURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := i.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("service returned status %d", resp.StatusCode)
	}

	// Parse response
	var jsonRPCResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&jsonRPCResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for JSON-RPC error
	if errorObj, exists := jsonRPCResponse["error"]; exists && errorObj != nil {
		return nil, fmt.Errorf("JSON-RPC error: %v", errorObj)
	}

	// Extract result
	result, exists := jsonRPCResponse["result"]
	if !exists {
		return nil, fmt.Errorf("no result in JSON-RPC response")
	}

	// Convert result to ServiceCapability
	capability, err := i.parseServiceCapability(serviceURL, result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse service capability: %w", err)
	}

	capability.ResponseTime = time.Since(startTime)
	capability.LastUpdated = time.Now()
	capability.TTL = i.config.IntrospectionTTL
	capability.IsAvailable = true

	return capability, nil
}

// parseServiceCapability converts JSON-RPC result to ServiceCapability
func (i *Introspector) parseServiceCapability(serviceURL string, result interface{}) (*ServiceCapability, error) {
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	var capability ServiceCapability
	if err := json.Unmarshal(resultBytes, &capability); err != nil {
		return nil, err
	}

	// Extra: mapear el campo 'service' o 'Name' si ServiceName está vacío
	var raw map[string]interface{}
	if err := json.Unmarshal(resultBytes, &raw); err == nil {
		if capability.ServiceName == "" {
			if v, ok := raw["service"].(string); ok && v != "" {
				capability.ServiceName = v
			} else if v, ok := raw["Name"].(string); ok && v != "" {
				capability.ServiceName = v
			}
		}
	}

	capability.ServiceURL = serviceURL
	if capability.ServiceName == "" {
		capability.ServiceName = "unknown"
	}

	return &capability, nil
}

// GetServiceCapability retrieves a service capability from cache
func (i *Introspector) GetServiceCapability(serviceURL string) *ServiceCapability {
	i.cacheMux.RLock()
	defer i.cacheMux.RUnlock()

	capability, exists := i.cache[serviceURL]
	if !exists {
		return nil
	}

	if i.isCacheExpired(capability) {
		return nil
	}

	return capability
}

// GetAllCapabilities returns all cached service capabilities
func (i *Introspector) GetAllCapabilities() []ServiceCapability {
	i.cacheMux.RLock()
	defer i.cacheMux.RUnlock()

	var capabilities []ServiceCapability
	for _, capability := range i.cache {
		if !i.isCacheExpired(capability) {
			capabilities = append(capabilities, *capability)
		}
	}

	if capabilities == nil {
		return []ServiceCapability{}
	}
	return capabilities
}

// ClearCache clears the capability cache
func (i *Introspector) ClearCache() {
	i.cacheMux.Lock()
	defer i.cacheMux.Unlock()
	i.cache = make(map[string]*ServiceCapability)
}

// getFromCache retrieves a capability from cache
func (i *Introspector) getFromCache(serviceURL string) *ServiceCapability {
	i.cacheMux.RLock()
	defer i.cacheMux.RUnlock()

	capability, exists := i.cache[serviceURL]
	if !exists {
		return nil
	}

	return capability
}

// updateCache updates the cache with new capabilities
func (i *Introspector) updateCache(capabilities []ServiceCapability) {
	i.cacheMux.Lock()
	defer i.cacheMux.Unlock()

	for _, capability := range capabilities {
		i.cache[capability.ServiceURL] = &capability
	}
}

// isCacheExpired checks if a cached capability has expired
func (i *Introspector) isCacheExpired(capability *ServiceCapability) bool {
	return time.Since(capability.LastUpdated) > capability.TTL
}

// GetServiceByName finds a service by name across all cached capabilities
func (i *Introspector) GetServiceByName(serviceName string) *ServiceCapability {
	i.cacheMux.RLock()
	defer i.cacheMux.RUnlock()

	for _, capability := range i.cache {
		if capability.ServiceName == serviceName && !i.isCacheExpired(capability) {
			return capability
		}
	}

	return nil
}

// GetMethodByName finds a method by name across all services
func (i *Introspector) GetMethodByName(methodName string) (*Method, *ServiceCapability) {
	i.cacheMux.RLock()
	defer i.cacheMux.RUnlock()

	for _, capability := range i.cache {
		if i.isCacheExpired(capability) {
			continue
		}

		if method, exists := capability.Methods[methodName]; exists {
			return &method, capability
		}
	}

	return nil, nil
}
