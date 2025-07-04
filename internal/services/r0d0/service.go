// Package r0d0 provides the R0D0 service for project discovery through conversational interaction.
// This service helps users discover and define their project requirements through natural conversation.
package r0d0

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// R0D0Service represents the main service for project discovery.
type R0D0Service struct {
	sessions    map[string]*Session
	sessionsMux sync.RWMutex

	// Discovery areas configuration
	discoveryAreas []DiscoveryArea

	// Service metadata
	serviceName    string
	serviceVersion string
}

// NewR0D0Service creates a new instance of the R0D0 service.
func NewR0D0Service() *R0D0Service {
	service := &R0D0Service{
		sessions:       make(map[string]*Session),
		serviceName:    "r0d0-service",
		serviceVersion: "1.0.0",
	}

	// Initialize discovery areas
	service.initDiscoveryAreas()

	// Start cleanup goroutine for expired sessions
	go service.cleanupExpiredSessions()

	return service
}

// Describe returns the service capabilities and metadata for introspection.
func (s *R0D0Service) Describe(args interface{}, reply *Service) error {
	*reply = Service{
		Name:        s.serviceName,
		Description: "R0D0 Service for conversational project discovery and ProjectSlot generation",
		Version:     s.serviceVersion,
		Methods: map[string]Method{
			"Describe": {
				Name:        "Describe",
				Description: "Returns service capabilities and metadata",
				Parameters:  []Parameter{},
				Returns:     "Service metadata and available methods",
			},
			"DiscoveryStart": {
				Name:        "DiscoveryStart",
				Description: "Starts a new project discovery session",
				Parameters: []Parameter{
					{Name: "user_id", Type: "string", Description: "Unique user identifier", Required: true},
					{Name: "message", Type: "string", Description: "Initial project description", Required: true},
				},
				Returns: "Session information and initial discovery response",
			},
			"DiscoveryContinue": {
				Name:        "DiscoveryContinue",
				Description: "Continues an existing discovery session",
				Parameters: []Parameter{
					{Name: "session_id", Type: "string", Description: "Session identifier", Required: true},
					{Name: "message", Type: "string", Description: "User's response/message", Required: true},
				},
				Returns: "Updated session with discovery progress",
			},
			"DiscoveryComplete": {
				Name:        "DiscoveryComplete",
				Description: "Completes discovery and generates ProjectSlot",
				Parameters: []Parameter{
					{Name: "session_id", Type: "string", Description: "Session identifier", Required: true},
					{Name: "force", Type: "boolean", Description: "Force completion even if confidence is low", Required: false},
				},
				Returns: "Generated ProjectSlot with discovery results",
			},
		},
		Capabilities: []string{
			"conversational_discovery",
			"project_slot_generation",
			"session_management",
			"confidence_analysis",
			"multi_area_discovery",
		},
		DiscoveryAreas: s.discoveryAreas,
		Purpose:        "Discovers project requirements through natural conversation to generate comprehensive ProjectSlots",
	}

	return nil
}

// DiscoveryStart starts a new discovery session with the user.
func (s *R0D0Service) DiscoveryStart(req *DiscoveryStartRequest, reply *DiscoveryStartResponse) error {
	// Validate request
	if req.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if req.Message == "" {
		return fmt.Errorf("message is required")
	}

	// Check concurrent sessions limit
	s.sessionsMux.RLock()
	userSessions := s.countUserSessions(req.UserID)
	s.sessionsMux.RUnlock()

	if userSessions >= MaxConcurrentSessions {
		return fmt.Errorf("maximum concurrent sessions exceeded for user")
	}

	// Create new session
	sessionID := uuid.New().String()
	now := time.Now()

	session := &Session{
		ID:         sessionID,
		UserID:     req.UserID,
		Status:     SessionStatusActive,
		Progress:   0,
		Confidence: 0,
		Conversation: []Message{
			{
				ID:        uuid.New().String(),
				Role:      "user",
				Content:   req.Message,
				Timestamp: now,
				Metadata:  map[string]interface{}{"intent": "project_start"},
			},
		},
		Insights:       []string{},
		MissingAreas:   s.getAllAreaKeys(),
		DiscoveredInfo: make(map[string]string),
		ProjectSlot:    nil,
		CreatedAt:      now,
		UpdatedAt:      now,
		ExpiresAt:      now.Add(DefaultSessionTimeout),
	}

	// Analyze initial message
	insights := s.analyzeMessage(req.Message)
	session.Insights = insights

	// Update discovered info and missing areas
	s.updateDiscoveredInfo(session, req.Message)
	s.updateMissingAreas(session)

	// Calculate progress and confidence
	session.Progress = s.calculateProgress(session)
	session.Confidence = s.calculateConfidence(session)

	// Generate response
	response := s.generateResponse(session)
	nextPrompt := s.generateNextPrompt(session)

	// Add assistant response to conversation
	session.Conversation = append(session.Conversation, Message{
		ID:        uuid.New().String(),
		Role:      "assistant",
		Content:   response,
		Timestamp: time.Now(),
		Metadata:  map[string]interface{}{"response_type": "discovery_start"},
	})

	// Store session
	s.sessionsMux.Lock()
	s.sessions[sessionID] = session
	s.sessionsMux.Unlock()

	// Build response
	*reply = DiscoveryStartResponse{
		SessionID:    sessionID,
		Response:     response,
		NextPrompt:   nextPrompt,
		Progress:     session.Progress,
		Insights:     session.Insights,
		Instructions: "Continue the conversation to help me understand your project better. Share more details about what you want to build.",
	}

	log.Printf("Started discovery session %s for user %s", sessionID, req.UserID)
	return nil
}

// DiscoveryContinue continues an existing discovery session.
func (s *R0D0Service) DiscoveryContinue(req *DiscoveryContinueRequest, reply *DiscoveryContinueResponse) error {
	// Validate request
	if req.SessionID == "" {
		return fmt.Errorf("session_id is required")
	}
	if req.Message == "" {
		return fmt.Errorf("message is required")
	}

	// Get session
	s.sessionsMux.Lock()
	session, exists := s.sessions[req.SessionID]
	if !exists {
		s.sessionsMux.Unlock()
		return fmt.Errorf("session not found")
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		session.Status = SessionStatusExpired
		s.sessionsMux.Unlock()
		return fmt.Errorf("session expired")
	}

	// Check if session is active
	if session.Status != SessionStatusActive {
		s.sessionsMux.Unlock()
		return fmt.Errorf("session is not active")
	}

	// Check conversation length limit
	if len(session.Conversation) >= MaxConversationLength {
		s.sessionsMux.Unlock()
		return fmt.Errorf("maximum conversation length exceeded")
	}

	// Add user message to conversation
	now := time.Now()
	userMessage := Message{
		ID:        uuid.New().String(),
		Role:      "user",
		Content:   req.Message,
		Timestamp: now,
		Metadata:  map[string]interface{}{"intent": "discovery_continue"},
	}
	session.Conversation = append(session.Conversation, userMessage)

	// Analyze new message
	newInsights := s.analyzeMessage(req.Message)
	session.Insights = s.mergeInsights(session.Insights, newInsights)

	// Update discovered info
	s.updateDiscoveredInfo(session, req.Message)
	s.updateMissingAreas(session)

	// Calculate updated progress and confidence
	session.Progress = s.calculateProgress(session)
	session.Confidence = s.calculateConfidence(session)

	// Generate response
	response := s.generateResponse(session)
	nextPrompt := s.generateNextPrompt(session)

	// Add assistant response to conversation
	assistantMessage := Message{
		ID:        uuid.New().String(),
		Role:      "assistant",
		Content:   response,
		Timestamp: time.Now(),
		Metadata:  map[string]interface{}{"response_type": "discovery_continue"},
	}
	session.Conversation = append(session.Conversation, assistantMessage)

	// Update session metadata
	session.UpdatedAt = now
	session.ExpiresAt = now.Add(DefaultSessionTimeout)

	// Check if discovery is complete
	isComplete := session.Confidence >= MinConfidenceLevel && len(session.MissingAreas) <= 2

	s.sessionsMux.Unlock()

	// Build response
	*reply = DiscoveryContinueResponse{
		SessionID:    req.SessionID,
		Response:     response,
		NextPrompt:   nextPrompt,
		Progress:     session.Progress,
		Insights:     session.Insights,
		IsComplete:   isComplete,
		Confidence:   session.Confidence,
		MissingAreas: session.MissingAreas,
	}

	log.Printf("Continued discovery session %s, progress: %d%%, confidence: %d%%",
		req.SessionID, session.Progress, session.Confidence)
	return nil
}

// DiscoveryComplete completes the discovery session and generates a ProjectSlot.
func (s *R0D0Service) DiscoveryComplete(req *DiscoveryCompleteRequest, reply *DiscoveryCompleteResponse) error {
	// Validate request
	if req.SessionID == "" {
		return fmt.Errorf("session_id is required")
	}

	// Get session
	s.sessionsMux.Lock()
	session, exists := s.sessions[req.SessionID]
	if !exists {
		s.sessionsMux.Unlock()
		return fmt.Errorf("session not found")
	}

	// Check if session is active
	if session.Status != SessionStatusActive {
		s.sessionsMux.Unlock()
		return fmt.Errorf("session is not active")
	}

	// Check confidence level unless forced
	if !req.Force && session.Confidence < MinConfidenceLevel {
		s.sessionsMux.Unlock()
		return fmt.Errorf("confidence level too low (%d%% < %d%%), use force=true to override",
			session.Confidence, MinConfidenceLevel)
	}

	// Generate ProjectSlot
	projectSlot := s.generateProjectSlot(session)
	session.ProjectSlot = projectSlot
	session.Status = SessionStatusCompleted
	session.UpdatedAt = time.Now()

	s.sessionsMux.Unlock()

	// Build response
	message := fmt.Sprintf("Discovery completed successfully! Generated ProjectSlot with %d%% confidence.",
		session.Confidence)

	var gaps []string
	if len(session.MissingAreas) > 0 {
		gaps = session.MissingAreas
		message += fmt.Sprintf(" Note: %d areas may need refinement.", len(gaps))
	}

	*reply = DiscoveryCompleteResponse{
		SessionID:   req.SessionID,
		ProjectSlot: *projectSlot,
		Message:     message,
		Success:     true,
		Confidence:  session.Confidence,
		Gaps:        gaps,
	}

	log.Printf("Completed discovery session %s, confidence: %d%%", req.SessionID, session.Confidence)
	return nil
}

// Helper methods for discovery logic

// initDiscoveryAreas initializes the discovery areas configuration.
func (s *R0D0Service) initDiscoveryAreas() {
	s.discoveryAreas = []DiscoveryArea{
		{
			Key:         "objective",
			Name:        "Project Objective",
			Description: "The main goal and purpose of the project",
			Priority:    1,
			Status:      "pending",
			Prompts:     []string{"What is the main goal of your project?", "What problem are you trying to solve?"},
			Keywords:    []string{"goal", "purpose", "objective", "problem", "solve", "create"},
		},
		{
			Key:         "audience",
			Name:        "Target Audience",
			Description: "Who will use or benefit from this project",
			Priority:    1,
			Status:      "pending",
			Prompts:     []string{"Who is your target audience?", "Who will be using this?"},
			Keywords:    []string{"users", "audience", "customers", "people", "target", "who"},
		},
		{
			Key:         "budget",
			Name:        "Budget Range",
			Description: "Financial constraints and budget expectations",
			Priority:    2,
			Status:      "pending",
			Prompts:     []string{"What's your budget range?", "How much are you looking to invest?"},
			Keywords:    []string{"budget", "cost", "price", "money", "invest", "spend"},
		},
		{
			Key:         "timeline",
			Name:        "Project Timeline",
			Description: "Expected duration and key milestones",
			Priority:    2,
			Status:      "pending",
			Prompts:     []string{"What's your expected timeline?", "When do you need this completed?"},
			Keywords:    []string{"timeline", "deadline", "time", "when", "duration", "schedule"},
		},
		{
			Key:         "technology",
			Name:        "Technology Preferences",
			Description: "Preferred technologies, platforms, and tools",
			Priority:    2,
			Status:      "pending",
			Prompts:     []string{"Any technology preferences?", "What platforms should we target?"},
			Keywords:    []string{"technology", "platform", "framework", "tools", "mobile", "web"},
		},
		{
			Key:         "resources",
			Name:        "Available Resources",
			Description: "Team, skills, and resources available",
			Priority:    3,
			Status:      "pending",
			Prompts:     []string{"What resources do you have available?", "Tell me about your team"},
			Keywords:    []string{"team", "resources", "skills", "people", "developers", "designers"},
		},
		{
			Key:         "risks",
			Name:        "Potential Risks",
			Description: "Challenges and risks to consider",
			Priority:    3,
			Status:      "pending",
			Prompts:     []string{"What challenges do you foresee?", "Any concerns or risks?"},
			Keywords:    []string{"risks", "challenges", "concerns", "problems", "issues", "obstacles"},
		},
		{
			Key:         "success",
			Name:        "Success Metrics",
			Description: "How success will be measured",
			Priority:    3,
			Status:      "pending",
			Prompts:     []string{"How will you measure success?", "What are your success criteria?"},
			Keywords:    []string{"success", "metrics", "measure", "kpi", "goals", "targets"},
		},
		{
			Key:         "context",
			Name:        "Additional Context",
			Description: "Background information and special requirements",
			Priority:    3,
			Status:      "pending",
			Prompts:     []string{"Any additional context?", "Is there anything else I should know?"},
			Keywords:    []string{"context", "background", "additional", "special", "requirements"},
		},
	}
}

// cleanupExpiredSessions runs a cleanup goroutine to remove expired sessions.
func (s *R0D0Service) cleanupExpiredSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.sessionsMux.Lock()
		now := time.Now()
		for sessionID, session := range s.sessions {
			if now.After(session.ExpiresAt) {
				session.Status = SessionStatusExpired
				delete(s.sessions, sessionID)
				log.Printf("Cleaned up expired session %s", sessionID)
			}
		}
		s.sessionsMux.Unlock()
	}
}

// countUserSessions counts active sessions for a user.
func (s *R0D0Service) countUserSessions(userID string) int {
	count := 0
	for _, session := range s.sessions {
		if session.UserID == userID && session.Status == SessionStatusActive {
			count++
		}
	}
	return count
}

// getAllAreaKeys returns all discovery area keys.
func (s *R0D0Service) getAllAreaKeys() []string {
	keys := make([]string, len(s.discoveryAreas))
	for i, area := range s.discoveryAreas {
		keys[i] = area.Key
	}
	return keys
}

// analyzeMessage extracts insights from a user message.
func (s *R0D0Service) analyzeMessage(message string) []string {
	insights := []string{}
	messageLower := strings.ToLower(message)

	// Simple keyword-based analysis (in real implementation, this would use NLP)
	projectTypes := map[string][]string{
		"mobile app": {"mobile", "app", "android", "ios", "smartphone"},
		"web app":    {"web", "website", "browser", "online"},
		"desktop":    {"desktop", "software", "application", "program"},
		"api":        {"api", "service", "backend", "rest"},
		"database":   {"database", "data", "storage", "sql"},
	}

	for projectType, keywords := range projectTypes {
		for _, keyword := range keywords {
			if strings.Contains(messageLower, keyword) {
				insights = append(insights, fmt.Sprintf("Project type: %s", projectType))
				break
			}
		}
	}

	// Analyze other aspects
	if strings.Contains(messageLower, "task") || strings.Contains(messageLower, "todo") || strings.Contains(messageLower, "manage") {
		insights = append(insights, "Functionality: Task management")
	}

	if strings.Contains(messageLower, "business") || strings.Contains(messageLower, "company") || strings.Contains(messageLower, "professional") {
		insights = append(insights, "Context: Business/Professional")
	}

	if strings.Contains(messageLower, "personal") || strings.Contains(messageLower, "individual") || strings.Contains(messageLower, "myself") {
		insights = append(insights, "Context: Personal use")
	}

	return insights
}

// updateDiscoveredInfo updates the discovered information based on the message.
func (s *R0D0Service) updateDiscoveredInfo(session *Session, message string) {
	messageLower := strings.ToLower(message)

	// Simple discovery logic - in real implementation, this would be more sophisticated
	for _, area := range s.discoveryAreas {
		for _, keyword := range area.Keywords {
			if strings.Contains(messageLower, keyword) {
				session.DiscoveredInfo[area.Key] = message
				break
			}
		}
	}
}

// updateMissingAreas updates the list of missing areas based on discovered info.
func (s *R0D0Service) updateMissingAreas(session *Session) {
	missing := []string{}
	for _, area := range s.discoveryAreas {
		if _, exists := session.DiscoveredInfo[area.Key]; !exists {
			missing = append(missing, area.Key)
		}
	}
	session.MissingAreas = missing
}

// calculateProgress calculates the discovery progress percentage.
func (s *R0D0Service) calculateProgress(session *Session) int {
	totalAreas := len(s.discoveryAreas)
	discoveredAreas := len(session.DiscoveredInfo)

	if totalAreas == 0 {
		return 0
	}

	progress := (discoveredAreas * 100) / totalAreas
	if progress > 100 {
		progress = 100
	}

	return progress
}

// calculateConfidence calculates the confidence level in the discovery.
func (s *R0D0Service) calculateConfidence(session *Session) int {
	// Base confidence on progress and conversation depth
	baseConfidence := session.Progress

	// Bonus for conversation depth
	conversationBonus := len(session.Conversation) * 2
	if conversationBonus > 30 {
		conversationBonus = 30
	}

	// Bonus for high-priority areas discovered
	priorityBonus := 0
	for _, area := range s.discoveryAreas {
		if _, exists := session.DiscoveredInfo[area.Key]; exists && area.Priority == 1 {
			priorityBonus += 15
		}
	}

	confidence := baseConfidence + conversationBonus + priorityBonus
	if confidence > 100 {
		confidence = 100
	}

	return confidence
}

// generateResponse generates a conversational response based on the session state.
func (s *R0D0Service) generateResponse(session *Session) string {
	if len(session.Conversation) == 1 {
		// First response
		return "¡Excelente! Me encanta ayudarte a desarrollar tu proyecto. Por lo que me cuentas, suena muy interesante. Déjame entender mejor los detalles para poder ayudarte de la mejor manera."
	}

	// Generate response based on missing areas
	if len(session.MissingAreas) > 0 {
		return fmt.Sprintf("Perfecto, voy entendiendo mejor tu proyecto. Tengo una idea del %d%% de lo que necesitas. Me gustaría profundizar en algunos aspectos más.", session.Progress)
	}

	return "¡Genial! Creo que ya tengo una buena comprensión de tu proyecto. Está tomando forma muy bien."
}

// generateNextPrompt generates the next prompt based on missing areas.
func (s *R0D0Service) generateNextPrompt(session *Session) string {
	// Find highest priority missing area
	for _, area := range s.discoveryAreas {
		for _, missingKey := range session.MissingAreas {
			if area.Key == missingKey && len(area.Prompts) > 0 {
				return area.Prompts[0]
			}
		}
	}

	return "¿Hay algo más que te gustaría agregar sobre tu proyecto?"
}

// mergeInsights merges new insights with existing ones, avoiding duplicates.
func (s *R0D0Service) mergeInsights(existing, new []string) []string {
	result := make([]string, len(existing))
	copy(result, existing)

	for _, newInsight := range new {
		found := false
		for _, existingInsight := range existing {
			if existingInsight == newInsight {
				found = true
				break
			}
		}
		if !found {
			result = append(result, newInsight)
		}
	}

	return result
}

// generateProjectSlot generates a ProjectSlot from the discovery session.
func (s *R0D0Service) generateProjectSlot(session *Session) *ProjectSlot {
	now := time.Now()

	// Extract information from discovered data
	objective := s.extractInfo(session.DiscoveredInfo, "objective", "Objetivo específico por definir")
	audience := s.extractInfo(session.DiscoveredInfo, "audience", "Usuario general")
	budget := s.extractInfo(session.DiscoveredInfo, "budget", "Presupuesto por definir")
	timeline := s.extractInfo(session.DiscoveredInfo, "timeline", "Timeline por definir")
	technology := s.extractInfo(session.DiscoveredInfo, "technology", "Tecnologías por definir")
	resources := s.extractInfo(session.DiscoveredInfo, "resources", "Recursos por definir")
	risks := s.extractInfo(session.DiscoveredInfo, "risks", "Riesgos por evaluar")
	success := s.extractInfo(session.DiscoveredInfo, "success", "Métricas de éxito por definir")

	// Generate project name from insights
	projectName := s.generateProjectName(session.Insights)

	return &ProjectSlot{
		ID:             uuid.New().String(),
		Name:           projectName,
		Objective:      objective,
		Budget:         budget,
		Timeline:       timeline,
		Technologies:   technology,
		Audience:       audience,
		Risks:          risks,
		Resources:      resources,
		SuccessMetrics: success,
		CreatedAt:      now,
		UpdatedAt:      now,
		UserID:         session.UserID,
		Status:         "discovered",
	}
}

// extractInfo extracts information from discovered data with fallback.
func (s *R0D0Service) extractInfo(discoveredInfo map[string]string, key, fallback string) string {
	if value, exists := discoveredInfo[key]; exists && value != "" {
		return value
	}
	return fallback
}

// generateProjectName generates a project name from insights.
func (s *R0D0Service) generateProjectName(insights []string) string {
	if len(insights) == 0 {
		return "Proyecto Descubierto"
	}

	// Simple name generation based on first insight
	firstInsight := insights[0]
	if strings.Contains(strings.ToLower(firstInsight), "mobile") {
		return "Aplicación Móvil"
	}
	if strings.Contains(strings.ToLower(firstInsight), "web") {
		return "Aplicación Web"
	}
	if strings.Contains(strings.ToLower(firstInsight), "task") {
		return "Sistema de Gestión de Tareas"
	}

	return "Proyecto Personalizado"
}
