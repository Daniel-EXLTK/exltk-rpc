// Package main provides the HTTP server for the LLM Orchestrator
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Daniel-EXLTK/exltk-rpc/internal/orchestrator"
	"github.com/joho/godotenv"
)

// JSONRPCRequest represents a JSON-RPC request
type JSONRPCRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      string                 `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
}

// JSONRPCResponse represents a JSON-RPC response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

// OrchestratorServer represents the HTTP server for the orchestrator
type OrchestratorServer struct {
	orchestrator *orchestrator.Orchestrator
	config       *orchestrator.OrchestratorConfig
}

// NewOrchestratorServer creates a new server instance
func NewOrchestratorServer(config *orchestrator.OrchestratorConfig) (*OrchestratorServer, error) {
	orch, err := orchestrator.NewOrchestrator(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create orchestrator: %w", err)
	}

	return &OrchestratorServer{
		orchestrator: orch,
		config:       config,
	}, nil
}

// handleJSONRPC handles JSON-RPC requests
func (s *OrchestratorServer) handleJSONRPC(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")

	// Handle preflight requests
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse JSON-RPC request
	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendJSONRPCError(w, req.ID, -32700, "Parse error", err.Error())
		return
	}

	// Validate JSON-RPC version
	if req.JSONRPC != "2.0" {
		s.sendJSONRPCError(w, req.ID, -32600, "Invalid Request", "JSON-RPC version must be 2.0")
		return
	}

	// Route to appropriate method
	switch req.Method {
	case "health":
		s.handleHealth(w, req)
	case "metrics":
		s.handleMetrics(w, req)
	case "orchestrate":
		s.handleOrchestrate(w, req)
	case "refresh_capabilities":
		s.handleRefreshCapabilities(w, req)
	case "get_capabilities":
		s.handleGetCapabilities(w, req)
	default:
		s.sendJSONRPCError(w, req.ID, -32601, "Method not found", fmt.Sprintf("Method '%s' not found", req.Method))
	}
}

// handleHealth handles the health check method
func (s *OrchestratorServer) handleHealth(w http.ResponseWriter, req JSONRPCRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	health := s.orchestrator.HealthCheck(ctx)
	s.sendJSONRPCResponse(w, req.ID, health)
}

// handleMetrics handles the metrics method
func (s *OrchestratorServer) handleMetrics(w http.ResponseWriter, req JSONRPCRequest) {
	metrics := s.orchestrator.GetMetrics()
	s.sendJSONRPCResponse(w, req.ID, metrics)
}

// handleOrchestrate handles the orchestrate method
func (s *OrchestratorServer) handleOrchestrate(w http.ResponseWriter, req JSONRPCRequest) {
	// Extract parameters
	userID, ok := req.Params["user_id"].(string)
	if !ok || userID == "" {
		s.sendJSONRPCError(w, req.ID, -32602, "Invalid params", "user_id is required")
		return
	}

	message, ok := req.Params["message"].(string)
	if !ok || message == "" {
		s.sendJSONRPCError(w, req.ID, -32602, "Invalid params", "message is required")
		return
	}

	// Extract optional context
	var userContext map[string]interface{}
	if contextParam, exists := req.Params["context"]; exists {
		if contextMap, ok := contextParam.(map[string]interface{}); ok {
			userContext = contextMap
		}
	}

	// Create user request
	userRequest := orchestrator.CreateUserRequest(userID, message, userContext)

	// Process request with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	response, err := s.orchestrator.ProcessRequest(ctx, userRequest)
	if err != nil {
		s.sendJSONRPCError(w, req.ID, -32603, "Internal error", err.Error())
		return
	}

	s.sendJSONRPCResponse(w, req.ID, response)
}

// handleRefreshCapabilities handles the refresh_capabilities method
func (s *OrchestratorServer) handleRefreshCapabilities(w http.ResponseWriter, req JSONRPCRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := s.orchestrator.RefreshCapabilities(ctx)
	if err != nil {
		s.sendJSONRPCError(w, req.ID, -32603, "Internal error", err.Error())
		return
	}

	s.sendJSONRPCResponse(w, req.ID, result)
}

// handleGetCapabilities handles the get_capabilities method
func (s *OrchestratorServer) handleGetCapabilities(w http.ResponseWriter, req JSONRPCRequest) {
	capabilities := s.orchestrator.GetServiceCapabilities()
	s.sendJSONRPCResponse(w, req.ID, capabilities)
}

// sendJSONRPCResponse sends a JSON-RPC response
func (s *OrchestratorServer) sendJSONRPCResponse(w http.ResponseWriter, id string, result interface{}) {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// sendJSONRPCError sends a JSON-RPC error response
func (s *OrchestratorServer) sendJSONRPCError(w http.ResponseWriter, id string, code int, message, data string) {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// sendError sends a simple HTTP error
func (s *OrchestratorServer) sendError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	w.Write([]byte(message))
}

// setupRoutes sets up the HTTP routes
func (s *OrchestratorServer) setupRoutes() {
	http.HandleFunc("/", s.handleJSONRPC)
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Load configuration from environment
	config := orchestrator.LoadConfigFromEnv()

	// Validate required configuration
	if config.LLMAPIKey == "" {
		log.Fatal("GOOGLE_API_KEY environment variable is required")
	}

	if len(config.ServiceURLs) == 0 {
		log.Fatal("SERVICE_URLS environment variable is required")
	}

	// Create server
	server, err := NewOrchestratorServer(config)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Setup routes
	server.setupRoutes()

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8502"
	}

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting orchestrator server on port %s", port)
		log.Printf("Service URLs: %v", config.ServiceURLs)
		log.Printf("LLM Provider: %s, Model: %s", config.LLMProvider, config.LLMModel)

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
