package r0d0

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDiscoveryStartRequest_JSON(t *testing.T) {
	req := DiscoveryStartRequest{
		UserID:  "user123",
		Message: "I want to build a mobile app for task management",
	}

	// Test marshaling
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal DiscoveryStartRequest: %v", err)
	}

	// Test unmarshaling
	var unmarshaled DiscoveryStartRequest
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal DiscoveryStartRequest: %v", err)
	}

	// Verify data integrity
	if unmarshaled.UserID != req.UserID {
		t.Errorf("Expected UserID %s, got %s", req.UserID, unmarshaled.UserID)
	}
	if unmarshaled.Message != req.Message {
		t.Errorf("Expected Message %s, got %s", req.Message, unmarshaled.Message)
	}
}

func TestProjectSlot_JSON(t *testing.T) {
	now := time.Now()
	slot := ProjectSlot{
		ID:             "project123",
		Name:           "Task Management App",
		Objective:      "Create a mobile app to help users manage their daily tasks",
		Budget:         "$50,000 - $100,000",
		Timeline:       "6 months",
		Technologies:   "React Native, Node.js, MongoDB",
		Audience:       "Busy professionals and students",
		Risks:          "Competition from existing apps, user adoption challenges",
		Resources:      "2 developers, 1 designer, 1 PM",
		SuccessMetrics: "10,000+ downloads in first 3 months",
		CreatedAt:      now,
		UpdatedAt:      now,
		UserID:         "user123",
		Status:         "draft",
	}

	// Test marshaling
	data, err := json.Marshal(slot)
	if err != nil {
		t.Fatalf("Failed to marshal ProjectSlot: %v", err)
	}

	// Test unmarshaling
	var unmarshaled ProjectSlot
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal ProjectSlot: %v", err)
	}

	// Verify critical fields
	if unmarshaled.ID != slot.ID {
		t.Errorf("Expected ID %s, got %s", slot.ID, unmarshaled.ID)
	}
	if unmarshaled.Name != slot.Name {
		t.Errorf("Expected Name %s, got %s", slot.Name, unmarshaled.Name)
	}
	if unmarshaled.UserID != slot.UserID {
		t.Errorf("Expected UserID %s, got %s", slot.UserID, unmarshaled.UserID)
	}
}

func TestSession_JSON(t *testing.T) {
	now := time.Now()
	session := Session{
		ID:         "session123",
		UserID:     "user123",
		Status:     SessionStatusActive,
		Progress:   45,
		Confidence: 70,
		Conversation: []Message{
			{
				ID:        "msg1",
				Role:      "user",
				Content:   "I want to build a mobile app",
				Timestamp: now,
				Metadata:  map[string]interface{}{"intent": "project_start"},
			},
			{
				ID:        "msg2",
				Role:      "assistant",
				Content:   "That's exciting! What kind of mobile app are you thinking about?",
				Timestamp: now.Add(time.Second),
				Metadata:  map[string]interface{}{"response_type": "clarification"},
			},
		},
		Insights:       []string{"User wants mobile app", "Needs clarification on type"},
		MissingAreas:   []string{"target_audience", "budget_range"},
		DiscoveredInfo: map[string]string{"project_type": "mobile_app"},
		CreatedAt:      now,
		UpdatedAt:      now,
		ExpiresAt:      now.Add(DefaultSessionTimeout),
	}

	// Test marshaling
	data, err := json.Marshal(session)
	if err != nil {
		t.Fatalf("Failed to marshal Session: %v", err)
	}

	// Test unmarshaling
	var unmarshaled Session
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal Session: %v", err)
	}

	// Verify critical fields
	if unmarshaled.ID != session.ID {
		t.Errorf("Expected ID %s, got %s", session.ID, unmarshaled.ID)
	}
	if unmarshaled.Status != session.Status {
		t.Errorf("Expected Status %s, got %s", session.Status, unmarshaled.Status)
	}
	if len(unmarshaled.Conversation) != len(session.Conversation) {
		t.Errorf("Expected %d messages, got %d", len(session.Conversation), len(unmarshaled.Conversation))
	}
	if unmarshaled.Progress != session.Progress {
		t.Errorf("Expected Progress %d, got %d", session.Progress, unmarshaled.Progress)
	}
}

func TestServiceError(t *testing.T) {
	err := ServiceError{
		Code:    ErrorSessionNotFound,
		Message: "Session not found",
		Details: "The specified session ID does not exist",
	}

	// Test Error() method
	if err.Error() != "Session not found" {
		t.Errorf("Expected error message 'Session not found', got '%s'", err.Error())
	}

	// Test JSON marshaling
	data, jsonErr := json.Marshal(err)
	if jsonErr != nil {
		t.Fatalf("Failed to marshal ServiceError: %v", jsonErr)
	}

	var unmarshaled ServiceError
	jsonErr = json.Unmarshal(data, &unmarshaled)
	if jsonErr != nil {
		t.Fatalf("Failed to unmarshal ServiceError: %v", jsonErr)
	}

	if unmarshaled.Code != err.Code {
		t.Errorf("Expected Code %s, got %s", err.Code, unmarshaled.Code)
	}
}

func TestDiscoveryArea(t *testing.T) {
	area := DiscoveryArea{
		Key:         "objective",
		Name:        "Project Objective",
		Description: "Understanding the main goal and purpose of the project",
		Priority:    1,
		Status:      "pending",
		Prompts:     []string{"What is the main goal?", "What problem are you solving?"},
		Keywords:    []string{"goal", "purpose", "objective", "problem"},
	}

	// Test JSON marshaling
	data, err := json.Marshal(area)
	if err != nil {
		t.Fatalf("Failed to marshal DiscoveryArea: %v", err)
	}

	var unmarshaled DiscoveryArea
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal DiscoveryArea: %v", err)
	}

	if unmarshaled.Key != area.Key {
		t.Errorf("Expected Key %s, got %s", area.Key, unmarshaled.Key)
	}
	if len(unmarshaled.Prompts) != len(area.Prompts) {
		t.Errorf("Expected %d prompts, got %d", len(area.Prompts), len(unmarshaled.Prompts))
	}
}

func TestSessionStatus_Constants(t *testing.T) {
	// Test that constants are defined correctly
	if SessionStatusActive != "active" {
		t.Errorf("Expected SessionStatusActive to be 'active', got '%s'", SessionStatusActive)
	}
	if SessionStatusCompleted != "completed" {
		t.Errorf("Expected SessionStatusCompleted to be 'completed', got '%s'", SessionStatusCompleted)
	}
	if SessionStatusExpired != "expired" {
		t.Errorf("Expected SessionStatusExpired to be 'expired', got '%s'", SessionStatusExpired)
	}
	if SessionStatusAbandoned != "abandoned" {
		t.Errorf("Expected SessionStatusAbandoned to be 'abandoned', got '%s'", SessionStatusAbandoned)
	}
}

func TestConstants(t *testing.T) {
	// Test configuration constants
	if DefaultSessionTimeout != 30*time.Minute {
		t.Errorf("Expected DefaultSessionTimeout to be 30 minutes, got %v", DefaultSessionTimeout)
	}
	if MaxConcurrentSessions != 3 {
		t.Errorf("Expected MaxConcurrentSessions to be 3, got %d", MaxConcurrentSessions)
	}
	if MinConfidenceLevel != 70 {
		t.Errorf("Expected MinConfidenceLevel to be 70, got %d", MinConfidenceLevel)
	}
	if DiscoveryAreasCount != 9 {
		t.Errorf("Expected DiscoveryAreasCount to be 9, got %d", DiscoveryAreasCount)
	}
}

func TestDiscoveryContinueResponse_CompleteFlow(t *testing.T) {
	response := DiscoveryContinueResponse{
		SessionID:    "session123",
		Response:     "Great! I understand you want to build a task management app.",
		NextPrompt:   "Can you tell me more about your target audience?",
		Progress:     60,
		Insights:     []string{"Mobile app", "Task management", "For personal use"},
		IsComplete:   false,
		Confidence:   75,
		MissingAreas: []string{"budget", "timeline", "technical_requirements"},
	}

	// Test that confidence is above minimum
	if response.Confidence < MinConfidenceLevel {
		t.Logf("Confidence %d is below minimum %d - discovery should continue", response.Confidence, MinConfidenceLevel)
	}

	// Test JSON serialization
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	var unmarshaled DiscoveryContinueResponse
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if unmarshaled.SessionID != response.SessionID {
		t.Errorf("Expected SessionID %s, got %s", response.SessionID, unmarshaled.SessionID)
	}
}
