package r0d0

import (
	"testing"
)

func TestNewR0D0Service(t *testing.T) {
	service := NewR0D0Service()

	if service == nil {
		t.Fatal("NewR0D0Service returned nil")
	}

	if service.serviceName != "r0d0-service" {
		t.Errorf("Expected service name 'r0d0-service', got '%s'", service.serviceName)
	}

	if service.serviceVersion != "1.0.0" {
		t.Errorf("Expected service version '1.0.0', got '%s'", service.serviceVersion)
	}

	if service.sessions == nil {
		t.Error("Sessions map should be initialized")
	}

	if len(service.discoveryAreas) != DiscoveryAreasCount {
		t.Errorf("Expected %d discovery areas, got %d", DiscoveryAreasCount, len(service.discoveryAreas))
	}
}

func TestR0D0Service_Describe(t *testing.T) {
	service := NewR0D0Service()
	var reply Service

	err := service.Describe(nil, &reply)
	if err != nil {
		t.Fatalf("Describe returned error: %v", err)
	}

	if reply.Name != "r0d0-service" {
		t.Errorf("Expected service name 'r0d0-service', got '%s'", reply.Name)
	}

	if reply.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", reply.Version)
	}

	// Check that all expected methods are present
	expectedMethods := []string{"Describe", "DiscoveryStart", "DiscoveryContinue", "DiscoveryComplete"}
	for _, methodName := range expectedMethods {
		if _, exists := reply.Methods[methodName]; !exists {
			t.Errorf("Method '%s' not found in service description", methodName)
		}
	}

	// Check capabilities
	if len(reply.Capabilities) == 0 {
		t.Error("Service should have capabilities")
	}

	// Check discovery areas
	if len(reply.DiscoveryAreas) != DiscoveryAreasCount {
		t.Errorf("Expected %d discovery areas, got %d", DiscoveryAreasCount, len(reply.DiscoveryAreas))
	}
}

func TestR0D0Service_DiscoveryStart(t *testing.T) {
	service := NewR0D0Service()

	// Test valid request
	req := &DiscoveryStartRequest{
		UserID:  "user123",
		Message: "I want to build a mobile app for task management",
	}

	var reply DiscoveryStartResponse
	err := service.DiscoveryStart(req, &reply)
	if err != nil {
		t.Fatalf("DiscoveryStart returned error: %v", err)
	}

	if reply.SessionID == "" {
		t.Error("SessionID should not be empty")
	}

	if reply.Response == "" {
		t.Error("Response should not be empty")
	}

	if reply.NextPrompt == "" {
		t.Error("NextPrompt should not be empty")
	}

	if reply.Progress < 0 || reply.Progress > 100 {
		t.Errorf("Progress should be between 0 and 100, got %d", reply.Progress)
	}

	if len(reply.Insights) == 0 {
		t.Error("Should have some insights from the initial message")
	}

	// Verify session was created
	service.sessionsMux.RLock()
	session, exists := service.sessions[reply.SessionID]
	service.sessionsMux.RUnlock()

	if !exists {
		t.Error("Session was not created")
	}

	if session.UserID != req.UserID {
		t.Errorf("Expected user ID '%s', got '%s'", req.UserID, session.UserID)
	}

	if session.Status != SessionStatusActive {
		t.Errorf("Expected status '%s', got '%s'", SessionStatusActive, session.Status)
	}
}

func TestR0D0Service_DiscoveryStart_InvalidRequest(t *testing.T) {
	service := NewR0D0Service()

	// Test with empty UserID
	req := &DiscoveryStartRequest{
		UserID:  "",
		Message: "I want to build an app",
	}

	var reply DiscoveryStartResponse
	err := service.DiscoveryStart(req, &reply)
	if err == nil {
		t.Error("Expected error for empty UserID")
	}

	// Test with empty Message
	req = &DiscoveryStartRequest{
		UserID:  "user123",
		Message: "",
	}

	err = service.DiscoveryStart(req, &reply)
	if err == nil {
		t.Error("Expected error for empty Message")
	}
}

func TestR0D0Service_DiscoveryContinue(t *testing.T) {
	service := NewR0D0Service()

	// First start a session
	startReq := &DiscoveryStartRequest{
		UserID:  "user123",
		Message: "I want to build a web application",
	}

	var startReply DiscoveryStartResponse
	err := service.DiscoveryStart(startReq, &startReply)
	if err != nil {
		t.Fatalf("DiscoveryStart failed: %v", err)
	}

	// Continue the session
	continueReq := &DiscoveryContinueRequest{
		SessionID: startReply.SessionID,
		Message:   "It's for managing tasks for small businesses",
	}

	var continueReply DiscoveryContinueResponse
	err = service.DiscoveryContinue(continueReq, &continueReply)
	if err != nil {
		t.Fatalf("DiscoveryContinue returned error: %v", err)
	}

	if continueReply.SessionID != startReply.SessionID {
		t.Errorf("Expected session ID '%s', got '%s'", startReply.SessionID, continueReply.SessionID)
	}

	if continueReply.Response == "" {
		t.Error("Response should not be empty")
	}

	if continueReply.Progress < 0 || continueReply.Progress > 100 {
		t.Errorf("Progress should be between 0 and 100, got %d", continueReply.Progress)
	}

	if continueReply.Confidence < 0 || continueReply.Confidence > 100 {
		t.Errorf("Confidence should be between 0 and 100, got %d", continueReply.Confidence)
	}

	// Verify session was updated
	service.sessionsMux.RLock()
	session := service.sessions[startReply.SessionID]
	service.sessionsMux.RUnlock()

	if len(session.Conversation) < 3 { // Initial user + assistant + new user message
		t.Error("Conversation should have at least 3 messages")
	}
}
