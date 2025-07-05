// Package orchestrator provides integration tests for the orchestrator
package orchestrator

import (
	"context"
	"testing"
	"time"
)

func TestOrchestrator_Integration(t *testing.T) {
	// Create test configuration
	config := &OrchestratorConfig{
		LLMProvider:      "gemini",
		LLMAPIKey:        "test-api-key",
		LLMModel:         "gemini-2.0-flash-exp",
		ServiceURLs:      []string{"http://localhost:8501"},
		IntrospectionTTL: 1 * time.Minute,
		MaxConcurrency:   5,
		RequestTimeout:   10 * time.Second,
		EnableLogging:    false, // Disable logging for tests
	}

	// Create orchestrator
	orchestrator, err := NewOrchestrator(config)
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}

	// Test health check
	t.Run("HealthCheck", func(t *testing.T) {
		health := orchestrator.HealthCheck(context.Background())

		if health["status"] == nil {
			t.Error("Health check should return status")
		}

		if health["timestamp"] == nil {
			t.Error("Health check should return timestamp")
		}
	})

	// Test metrics
	t.Run("GetMetrics", func(t *testing.T) {
		metrics := orchestrator.GetMetrics()

		if metrics["total_services"] == nil {
			t.Error("Metrics should return total_services")
		}

		if metrics["available_services"] == nil {
			t.Error("Metrics should return available_services")
		}

		if metrics["total_methods"] == nil {
			t.Error("Metrics should return total_methods")
		}
	})

	// Test service capabilities
	t.Run("GetServiceCapabilities", func(t *testing.T) {
		capabilities := orchestrator.GetServiceCapabilities()

		// Should return empty slice if no services are available
		if capabilities == nil {
			t.Error("GetServiceCapabilities should return empty slice, not nil")
		}

		// Should be a slice (even if empty)
		if capabilities == nil {
			t.Error("GetServiceCapabilities should return a slice")
		}
	})
}

func TestOrchestrator_Configuration(t *testing.T) {
	t.Run("ValidConfig", func(t *testing.T) {
		config := &OrchestratorConfig{
			LLMProvider: "gemini",
			LLMAPIKey:   "test-key",
			LLMModel:    "gemini-2.0-flash-exp",
		}

		orchestrator, err := NewOrchestrator(config)
		if err != nil {
			t.Fatalf("Valid config should not cause error: %v", err)
		}

		if orchestrator == nil {
			t.Error("Orchestrator should not be nil")
		}
	})

	t.Run("MissingAPIKey", func(t *testing.T) {
		config := &OrchestratorConfig{
			LLMProvider: "gemini",
			LLMAPIKey:   "", // Missing API key
			LLMModel:    "gemini-2.0-flash-exp",
		}

		_, err := NewOrchestrator(config)
		if err == nil {
			t.Error("Missing API key should cause error")
		}
	})

	t.Run("NilConfig", func(t *testing.T) {
		_, err := NewOrchestrator(nil)
		if err == nil {
			t.Error("Nil config should cause error")
		}
	})
}

func TestOrchestrator_UserRequest(t *testing.T) {
	t.Run("CreateUserRequest", func(t *testing.T) {
		userID := "test-user"
		message := "I want to start a project discovery session"
		context := map[string]interface{}{
			"session_type": "discovery",
		}

		request := CreateUserRequest(userID, message, context)

		if request.ID == "" {
			t.Error("Request ID should not be empty")
		}

		if request.UserID != userID {
			t.Errorf("Expected UserID %s, got %s", userID, request.UserID)
		}

		if request.Message != message {
			t.Errorf("Expected Message %s, got %s", message, request.Message)
		}

		if request.Context["session_type"] != "discovery" {
			t.Error("Context should contain session_type")
		}

		if request.Timestamp.IsZero() {
			t.Error("Timestamp should not be zero")
		}
	})

	t.Run("CreateUserRequestWithNilContext", func(t *testing.T) {
		request := CreateUserRequest("test-user", "test message", nil)

		if request.Context == nil {
			t.Error("Context should not be nil")
		}

		if len(request.Context) != 0 {
			t.Error("Context should be empty map")
		}
	})
}

func TestOrchestrator_LoadConfigFromEnv(t *testing.T) {
	// This test would require setting environment variables
	// For now, just test that the function doesn't panic
	t.Run("LoadConfigFromEnv", func(t *testing.T) {
		config := LoadConfigFromEnv()

		if config == nil {
			t.Error("Config should not be nil")
		}

		if config.LLMProvider == "" {
			t.Error("LLMProvider should have default value")
		}

		if config.IntrospectionTTL == 0 {
			t.Error("IntrospectionTTL should have default value")
		}

		if config.RequestTimeout == 0 {
			t.Error("RequestTimeout should have default value")
		}

		if config.MaxConcurrency == 0 {
			t.Error("MaxConcurrency should have default value")
		}
	})
}

func TestOrchestrator_ErrorHandling(t *testing.T) {
	config := &OrchestratorConfig{
		LLMProvider: "gemini",
		LLMAPIKey:   "test-key",
		LLMModel:    "gemini-2.0-flash-exp",
		ServiceURLs: []string{"http://invalid-url:9999"}, // Invalid URL
	}

	orchestrator, err := NewOrchestrator(config)
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}

	t.Run("ProcessRequestWithNoServices", func(t *testing.T) {
		request := CreateUserRequest("test-user", "test message", nil)

		response, err := orchestrator.ProcessRequest(context.Background(), request)
		if err != nil {
			t.Fatalf("ProcessRequest should not return error: %v", err)
		}

		if response.Success {
			t.Error("Response should not be successful with no services")
		}

		if response.Error == "" {
			t.Error("Response should have error code")
		}
	})
}

func TestOrchestrator_RefreshCapabilities(t *testing.T) {
	config := &OrchestratorConfig{
		LLMProvider: "gemini",
		LLMAPIKey:   "test-key",
		LLMModel:    "gemini-2.0-flash-exp",
		ServiceURLs: []string{"http://localhost:8501"},
	}

	orchestrator, err := NewOrchestrator(config)
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}

	t.Run("RefreshCapabilities", func(t *testing.T) {
		result, err := orchestrator.RefreshCapabilities(context.Background())
		if err != nil {
			t.Fatalf("RefreshCapabilities should not return error: %v", err)
		}

		if result == nil {
			t.Error("IntrospectionResult should not be nil")
		}

		if result.Timestamp.IsZero() {
			t.Error("Timestamp should not be zero")
		}
	})
}

func TestOrchestrator_WorkflowStatus(t *testing.T) {
	config := &OrchestratorConfig{
		LLMProvider: "gemini",
		LLMAPIKey:   "test-key",
		LLMModel:    "gemini-2.0-flash-exp",
	}

	orchestrator, err := NewOrchestrator(config)
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}

	t.Run("GetWorkflowStatus", func(t *testing.T) {
		_, err := orchestrator.GetWorkflowStatus("test-workflow-id")
		if err == nil {
			t.Error("GetWorkflowStatus should return error (not implemented)")
		}
	})
}

// Benchmark tests
func BenchmarkOrchestrator_HealthCheck(b *testing.B) {
	config := &OrchestratorConfig{
		LLMProvider:   "gemini",
		LLMAPIKey:     "test-key",
		LLMModel:      "gemini-2.0-flash-exp",
		EnableLogging: false,
	}

	orchestrator, err := NewOrchestrator(config)
	if err != nil {
		b.Fatalf("Failed to create orchestrator: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		orchestrator.HealthCheck(context.Background())
	}
}

func BenchmarkOrchestrator_GetMetrics(b *testing.B) {
	config := &OrchestratorConfig{
		LLMProvider:   "gemini",
		LLMAPIKey:     "test-key",
		LLMModel:      "gemini-2.0-flash-exp",
		EnableLogging: false,
	}

	orchestrator, err := NewOrchestrator(config)
	if err != nil {
		b.Fatalf("Failed to create orchestrator: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		orchestrator.GetMetrics()
	}
}
