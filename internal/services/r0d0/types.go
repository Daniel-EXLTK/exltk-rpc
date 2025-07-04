// Package r0d0 provides types and services for project discovery and ProjectSlot generation.
// This service handles conversational discovery sessions to understand what project the user wants to create.
package r0d0

import (
	"time"
)

// Service represents the capabilities and metadata of the R0D0 service.
// This is returned by the Describe() method for service introspection.
type Service struct {
	Name           string            `json:"service"`         // Service name: "r0d0-service"
	Description    string            `json:"description"`     // Human-readable description
	Version        string            `json:"version"`         // Service version
	Methods        map[string]Method `json:"methods"`         // Available JSON-RPC methods
	Capabilities   []string          `json:"capabilities"`    // List of service capabilities
	DiscoveryAreas []DiscoveryArea   `json:"discovery_areas"` // Areas this service can discover
	Purpose        string            `json:"purpose"`         // Main purpose of the service
}

// Method represents a JSON-RPC method with its metadata.
type Method struct {
	Name        string      `json:"name"`        // Method name
	Description string      `json:"description"` // Method description
	Parameters  []Parameter `json:"parameters"`  // Method parameters
	Returns     string      `json:"returns"`     // Return type description
}

// Parameter represents a method parameter specification.
type Parameter struct {
	Name        string `json:"name"`        // Parameter name
	Type        string `json:"type"`        // Parameter type
	Description string `json:"description"` // Parameter description
	Required    bool   `json:"required"`    // Whether parameter is required
}

// DiscoveryStartRequest represents the request to start a new project discovery session.
type DiscoveryStartRequest struct {
	UserID  string `json:"user_id"` // Unique user identifier
	Message string `json:"message"` // Initial user message describing their project idea
}

// DiscoveryStartResponse represents the response when starting a discovery session.
type DiscoveryStartResponse struct {
	SessionID    string   `json:"session_id"`   // Unique session identifier
	Response     string   `json:"response"`     // Conversational response to engage the user
	NextPrompt   string   `json:"next_prompt"`  // Suggested next topic to explore
	Progress     int      `json:"progress"`     // Discovery progress percentage (0-100)
	Insights     []string `json:"insights"`     // Key insights discovered so far
	Instructions string   `json:"instructions"` // Instructions for the user
}

// DiscoveryContinueRequest represents the request to continue an existing discovery session.
type DiscoveryContinueRequest struct {
	SessionID string `json:"session_id"` // Session identifier
	Message   string `json:"message"`    // User's message/response in the conversation
}

// DiscoveryContinueResponse represents the response when continuing a discovery session.
type DiscoveryContinueResponse struct {
	SessionID    string   `json:"session_id"`    // Session identifier
	Response     string   `json:"response"`      // Conversational response
	NextPrompt   string   `json:"next_prompt"`   // Suggested next topic to explore
	Progress     int      `json:"progress"`      // Discovery progress percentage (0-100)
	Insights     []string `json:"insights"`      // Key insights discovered so far
	IsComplete   bool     `json:"is_complete"`   // Whether discovery is complete enough
	Confidence   int      `json:"confidence"`    // Confidence level in understanding (0-100)
	MissingAreas []string `json:"missing_areas"` // Areas that still need clarification
}

// DiscoveryCompleteRequest represents the request to complete and generate ProjectSlot.
type DiscoveryCompleteRequest struct {
	SessionID string `json:"session_id"` // Session identifier
	Force     bool   `json:"force"`      // Force completion even if not fully discovered
}

// DiscoveryCompleteResponse represents the response when completing discovery.
type DiscoveryCompleteResponse struct {
	SessionID   string      `json:"session_id"`   // Session identifier
	ProjectSlot ProjectSlot `json:"project_slot"` // Generated ProjectSlot
	Message     string      `json:"message"`      // Completion message
	Success     bool        `json:"success"`      // Whether completion was successful
	Confidence  int         `json:"confidence"`   // Confidence in the generated ProjectSlot
	Gaps        []string    `json:"gaps"`         // Areas that may need refinement
}

// ProjectSlot represents a structured project definition generated from survey responses.
// This is the main output of the R0D0 service and input for other services.
type ProjectSlot struct {
	ID             string    `json:"id"`              // Unique ProjectSlot identifier
	Name           string    `json:"name"`            // Project name
	Objective      string    `json:"objective"`       // Project objective and goals
	Budget         string    `json:"budget"`          // Budget information and constraints
	Timeline       string    `json:"timeline"`        // Project timeline and milestones
	Technologies   string    `json:"technologies"`    // Required technologies and tools
	Audience       string    `json:"audience"`        // Target audience description
	Risks          string    `json:"risks"`           // Identified risks and mitigation strategies
	Resources      string    `json:"resources"`       // Required resources and team
	SuccessMetrics string    `json:"success_metrics"` // Success criteria and KPIs
	CreatedAt      time.Time `json:"created_at"`      // Creation timestamp
	UpdatedAt      time.Time `json:"updated_at"`      // Last update timestamp
	UserID         string    `json:"user_id"`         // Associated user identifier
	Status         string    `json:"status"`          // ProjectSlot status (draft, complete, approved)
}

// Session represents an active discovery session with a user.
// Sessions are used to maintain state across multiple conversational interactions.
type Session struct {
	ID             string            `json:"id"`              // Unique session identifier
	UserID         string            `json:"user_id"`         // Associated user identifier
	Status         SessionStatus     `json:"status"`          // Current session status
	Progress       int               `json:"progress"`        // Discovery progress percentage (0-100)
	Confidence     int               `json:"confidence"`      // Confidence in understanding (0-100)
	Conversation   []Message         `json:"conversation"`    // Full conversation history
	Insights       []string          `json:"insights"`        // Key insights discovered
	MissingAreas   []string          `json:"missing_areas"`   // Areas needing clarification
	DiscoveredInfo map[string]string `json:"discovered_info"` // Key-value pairs of discovered information
	ProjectSlot    *ProjectSlot      `json:"project_slot"`    // Generated ProjectSlot (when complete)
	CreatedAt      time.Time         `json:"created_at"`      // Session creation time
	UpdatedAt      time.Time         `json:"updated_at"`      // Last update time
	ExpiresAt      time.Time         `json:"expires_at"`      // Session expiration time
}

// Message represents a single message in the conversation history.
type Message struct {
	ID        string                 `json:"id"`        // Unique message identifier
	Role      string                 `json:"role"`      // Message role: "user" or "assistant"
	Content   string                 `json:"content"`   // Message content
	Timestamp time.Time              `json:"timestamp"` // Message timestamp
	Metadata  map[string]interface{} `json:"metadata"`  // Additional metadata
}

// SessionStatus represents the current status of a discovery session.
type SessionStatus string

const (
	// SessionStatusActive indicates the session is active and awaiting user input
	SessionStatusActive SessionStatus = "active"
	// SessionStatusCompleted indicates the session has been completed successfully
	SessionStatusCompleted SessionStatus = "completed"
	// SessionStatusExpired indicates the session has expired due to inactivity
	SessionStatusExpired SessionStatus = "expired"
	// SessionStatusAbandoned indicates the session was abandoned by the user
	SessionStatusAbandoned SessionStatus = "abandoned"
)

// DiscoveryArea represents an area of project understanding that needs to be explored.
type DiscoveryArea struct {
	Key         string   `json:"key"`         // Unique area key (e.g., "objective", "audience")
	Name        string   `json:"name"`        // Human-readable area name
	Description string   `json:"description"` // Description of what to discover
	Priority    int      `json:"priority"`    // Priority level (1=high, 3=low)
	Status      string   `json:"status"`      // Status: "pending", "partial", "complete"
	Prompts     []string `json:"prompts"`     // Suggested conversation prompts
	Keywords    []string `json:"keywords"`    // Keywords to look for in responses
}

// ErrorResponse represents an error response from the service.
type ErrorResponse struct {
	Code    string `json:"code"`    // Error code
	Message string `json:"message"` // Human-readable error message
	Details string `json:"details"` // Additional error details
}

// ServiceError represents service-specific errors.
type ServiceError struct {
	Code    string
	Message string
	Details string
}

// Error implements the error interface for ServiceError.
func (e ServiceError) Error() string {
	return e.Message
}

// Common error codes
const (
	ErrorSessionNotFound     = "SESSION_NOT_FOUND"
	ErrorSessionExpired      = "SESSION_EXPIRED"
	ErrorInvalidMessage      = "INVALID_MESSAGE"
	ErrorDiscoveryIncomplete = "DISCOVERY_INCOMPLETE"
	ErrorInternalError       = "INTERNAL_ERROR"
	ErrorInvalidRequest      = "INVALID_REQUEST"
	ErrorUserNotFound        = "USER_NOT_FOUND"
	ErrorLowConfidence       = "LOW_CONFIDENCE"
)

// Discovery configuration constants
const (
	// DefaultSessionTimeout is the default session timeout duration
	DefaultSessionTimeout = 30 * time.Minute
	// MaxConcurrentSessions is the maximum number of concurrent sessions per user
	MaxConcurrentSessions = 3
	// MinConfidenceLevel is the minimum confidence required for completion
	MinConfidenceLevel = 70
	// MaxConversationLength is the maximum number of messages in a session
	MaxConversationLength = 50
	// DiscoveryAreasCount is the number of key areas to discover
	DiscoveryAreasCount = 9
)
