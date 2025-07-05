// Package orchestrator provides the central LLM orchestrator for coordinating JSON-RPC services
package orchestrator

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Orchestrator is the main orchestrator that coordinates all operations
type Orchestrator struct {
	config       *OrchestratorConfig
	introspector *Introspector
	planner      *Planner
	executor     *Executor
}

// NewOrchestrator creates a new orchestrator instance
func NewOrchestrator(config *OrchestratorConfig) (*Orchestrator, error) {
	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create components
	introspector := NewIntrospector(config)

	planner, err := NewPlanner(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create planner: %w", err)
	}

	executor := NewExecutor(config)

	orchestrator := &Orchestrator{
		config:       config,
		introspector: introspector,
		planner:      planner,
		executor:     executor,
	}

	if config.EnableLogging {
		log.Printf("Orchestrator initialized with %d service URLs", len(config.ServiceURLs))
	}

	return orchestrator, nil
}

// ProcessRequest processes a user's natural language request
func (o *Orchestrator) ProcessRequest(ctx context.Context, request *UserRequest) (*OrchestratorResponse, error) {
	startTime := time.Now()

	if o.config.EnableLogging {
		log.Printf("Processing request: %s from user: %s", request.ID, request.UserID)
	}

	// Step 1: Introspect services to discover capabilities
	introspectionResult, err := o.introspector.IntrospectServices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to introspect services: %w", err)
	}

	if len(introspectionResult.Services) == 0 {
		return o.buildErrorResponse(request, "No services available", ErrorServiceUnavailable), nil
	}

	// Step 2: Generate workflow plan using LLM
	plan, err := o.planner.PlanWorkflow(ctx, request, introspectionResult.Services)
	if err != nil {
		return nil, fmt.Errorf("failed to plan workflow: %w", err)
	}

	// Step 3: Enhance plan with user context
	o.planner.EnhancePlanWithContext(plan, request.Context)

	// Step 4: Execute the workflow
	response, err := o.executor.ExecuteWorkflow(ctx, plan)
	if err != nil {
		return nil, fmt.Errorf("failed to execute workflow: %w", err)
	}

	// Update response with request details
	response.RequestID = request.ID
	response.Duration = time.Since(startTime)

	if o.config.EnableLogging {
		log.Printf("Request %s processed successfully in %v", request.ID, response.Duration)
	}

	return response, nil
}

// GetServiceCapabilities returns the current service capabilities
func (o *Orchestrator) GetServiceCapabilities() []ServiceCapability {
	return o.introspector.GetAllCapabilities()
}

// RefreshCapabilities forces a refresh of service capabilities
func (o *Orchestrator) RefreshCapabilities(ctx context.Context) (*IntrospectionResult, error) {
	o.introspector.ClearCache()
	return o.introspector.IntrospectServices(ctx)
}

// GetWorkflowStatus returns the status of a workflow (if implemented with persistence)
func (o *Orchestrator) GetWorkflowStatus(workflowID string) (*WorkflowPlan, error) {
	// This would typically query a database or cache
	// For now, return an error indicating not implemented
	return nil, fmt.Errorf("workflow status tracking not implemented")
}

// buildErrorResponse builds an error response
func (o *Orchestrator) buildErrorResponse(request *UserRequest, message, errorCode string) *OrchestratorResponse {
	return &OrchestratorResponse{
		RequestID:  request.ID,
		WorkflowID: "",
		Response:   message,
		Result:     nil,
		Steps:      []WorkflowStep{},
		Context:    request.Context,
		Success:    false,
		Error:      errorCode,
		Duration:   0,
		Timestamp:  time.Now(),
	}
}

// validateConfig validates the orchestrator configuration
func validateConfig(config *OrchestratorConfig) error {
	if config == nil {
		return fmt.Errorf("configuration is required")
	}

	if config.LLMAPIKey == "" {
		return fmt.Errorf("LLM API key is required")
	}

	if config.LLMProvider == "" {
		config.LLMProvider = "gemini"
	}

	if config.LLMModel == "" {
		config.LLMModel = "gemini-2.0-flash-exp"
	}

	if config.IntrospectionTTL == 0 {
		config.IntrospectionTTL = 5 * time.Minute
	}

	if config.RequestTimeout == 0 {
		config.RequestTimeout = 30 * time.Second
	}

	if config.MaxConcurrency == 0 {
		config.MaxConcurrency = 10
	}

	return nil
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv() *OrchestratorConfig {
	config := &OrchestratorConfig{
		LLMProvider:      getEnvOrDefault("LLM_PROVIDER", "gemini"),
		LLMAPIKey:        getEnvOrDefault("GOOGLE_API_KEY", ""),
		LLMModel:         getEnvOrDefault("LLM_MODEL", "gemini-2.0-flash-exp"),
		IntrospectionTTL: 5 * time.Minute,
		MaxConcurrency:   10,
		RequestTimeout:   30 * time.Second,
		EnableLogging:    getEnvOrDefault("ENABLE_LOGGING", "true") == "true",
	}

	// Load service URLs
	serviceURLs := getEnvOrDefault("SERVICE_URLS", "http://localhost:8501")
	config.ServiceURLs = strings.Split(serviceURLs, ",")

	// Clean up URLs
	for i, url := range config.ServiceURLs {
		config.ServiceURLs[i] = strings.TrimSpace(url)
	}

	return config
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// CreateUserRequest creates a new user request
func CreateUserRequest(userID, message string, context map[string]interface{}) *UserRequest {
	if context == nil {
		context = make(map[string]interface{})
	}

	return &UserRequest{
		ID:        uuid.New().String(),
		UserID:    userID,
		Message:   message,
		Context:   context,
		Timestamp: time.Now(),
	}
}

// HealthCheck performs a health check on the orchestrator
func (o *Orchestrator) HealthCheck(ctx context.Context) map[string]interface{} {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
	}

	// Check LLM connectivity
	if o.config.EnableLogging {
		log.Printf("Performing health check...")
	}

	// Check service introspection
	introspectionResult, err := o.introspector.IntrospectServices(ctx)
	if err != nil {
		health["status"] = "unhealthy"
		health["error"] = err.Error()
		return health
	}

	health["services_available"] = introspectionResult.TotalFound
	health["services_failed"] = len(introspectionResult.FailedURLs)
	health["last_introspection"] = introspectionResult.Timestamp

	return health
}

// GetMetrics returns orchestrator metrics
func (o *Orchestrator) GetMetrics() map[string]interface{} {
	capabilities := o.introspector.GetAllCapabilities()

	metrics := map[string]interface{}{
		"total_services":     len(capabilities),
		"available_services": 0,
		"total_methods":      0,
	}

	for _, capability := range capabilities {
		if capability.IsAvailable {
			metrics["available_services"] = metrics["available_services"].(int) + 1
		}
		metrics["total_methods"] = metrics["total_methods"].(int) + len(capability.Methods)
	}

	return metrics
}
