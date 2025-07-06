// Package orchestrator provides workflow execution capabilities
package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Executor handles workflow execution
type Executor struct {
	config     *OrchestratorConfig
	httpClient *http.Client
}

// NewExecutor creates a new executor instance
func NewExecutor(config *OrchestratorConfig) *Executor {
	return &Executor{
		config: config,
		httpClient: &http.Client{
			Timeout: config.RequestTimeout,
		},
	}
}

// ExecuteWorkflow executes a complete workflow plan
func (e *Executor) ExecuteWorkflow(ctx context.Context, plan *WorkflowPlan) (*OrchestratorResponse, error) {
	startTime := time.Now()

	if e.config.EnableLogging {
		log.Printf("Executing workflow: %s", plan.ID)
	}

	// Update workflow status
	plan.Status = WorkflowStatusRunning

	// Execute steps sequentially
	for i := range plan.Steps {
		step := &plan.Steps[i]

		// Check if step should be executed based on condition
		if step.Condition != "" {
			if !e.evaluateCondition(step.Condition, plan.Context) {
				step.Status = StepStatusSkipped
				if e.config.EnableLogging {
					log.Printf("Step %s skipped due to condition: %s", step.ID, step.Condition)
				}
				continue
			}
		}

		// Execute step with retry logic
		err := e.executeStepWithRetry(ctx, step, plan.Context)
		if err != nil {
			step.Status = StepStatusFailed
			step.Error = err.Error()
			plan.Status = WorkflowStatusFailed

			if e.config.EnableLogging {
				log.Printf("Step %s failed: %v", step.ID, err)
			}

			// Return partial response with error
			return e.buildResponse(plan, startTime, false, err.Error()), nil
		}

		step.Status = StepStatusCompleted
		now := time.Now()
		step.CompletedAt = &now

		if e.config.EnableLogging {
			log.Printf("Step %s completed successfully", step.ID)
		}
	}

	// All steps completed successfully
	plan.Status = WorkflowStatusCompleted

	if e.config.EnableLogging {
		log.Printf("Workflow %s completed successfully in %v", plan.ID, time.Since(startTime))
	}

	return e.buildResponse(plan, startTime, true, ""), nil
}

// executeStepWithRetry executes a single step with retry logic
func (e *Executor) executeStepWithRetry(ctx context.Context, step *WorkflowStep, context map[string]interface{}) error {
	retryPolicy := step.RetryPolicy
	if retryPolicy == nil {
		retryPolicy = &RetryPolicy{
			MaxRetries: 3,
			Delay:      5 * time.Second,
			Backoff:    2.0,
		}
	}

	var lastError error
	delay := retryPolicy.Delay

	for attempt := 0; attempt <= retryPolicy.MaxRetries; attempt++ {
		// Execute step
		err := e.executeStep(ctx, step, context)
		if err == nil {
			return nil // Success
		}

		lastError = err

		// If this was the last attempt, don't retry
		if attempt == retryPolicy.MaxRetries {
			break
		}

		if e.config.EnableLogging {
			log.Printf("Step %s failed (attempt %d/%d): %v, retrying in %v",
				step.ID, attempt+1, retryPolicy.MaxRetries+1, err, delay)
		}

		// Wait before retry
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}

		// Calculate next delay with exponential backoff
		delay = time.Duration(float64(delay) * retryPolicy.Backoff)
	}

	return fmt.Errorf("step failed after %d attempts, last error: %w", retryPolicy.MaxRetries+1, lastError)
}

// executeStep executes a single workflow step
func (e *Executor) executeStep(ctx context.Context, step *WorkflowStep, context map[string]interface{}) error {
	startTime := time.Now()
	step.StartedAt = &startTime
	step.Status = StepStatusRunning

	// Store current step ID in context for output resolution
	context["current_step_id"] = step.ID

	// Prepare parameters with context values
	parameters := e.prepareParameters(step.Parameters, step.Inputs, context)

	// Find service URL
	serviceURL := e.findServiceURL(step.Service)
	if serviceURL == "" {
		return fmt.Errorf("service URL not found for service: %s", step.Service)
	}

	// Prepare JSON-RPC request
	var methodName = step.Method
	if (step.Service == "r0d0-service" || step.Service == "r0d0") && !strings.HasPrefix(step.Method, "R0D0Service.") {
		methodName = "R0D0Service." + step.Method
	}
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      uuid.New().String(),
		"method":  methodName,
		"params":  []interface{}{parameters},
	}
	if e.config.EnableLogging {
		log.Printf("[EXEC] Llamando %s a %s con params: %+v", methodName, serviceURL, parameters)
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", serviceURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := e.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("service returned status %d", resp.StatusCode)
	}

	// Parse response
	var jsonRPCResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&jsonRPCResponse); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for JSON-RPC error
	if errorObj, exists := jsonRPCResponse["error"]; exists && errorObj != nil {
		return fmt.Errorf("JSON-RPC error: %v", errorObj)
	}

	// Extract result
	result, exists := jsonRPCResponse["result"]
	if !exists {
		return fmt.Errorf("no result in JSON-RPC response")
	}

	// Store result in step
	step.Result = result

	// Update context with outputs
	e.updateContextWithOutputs(step.Outputs, result, context)

	return nil
}

// prepareParameters prepares step parameters by substituting context values
func (e *Executor) prepareParameters(parameters map[string]interface{}, inputs []string, context map[string]interface{}) map[string]interface{} {
	prepared := make(map[string]interface{})

	// Copy parameters
	for key, value := range parameters {
		prepared[key] = value
	}

	// Substitute context values for inputs and JSONPath references
	for paramName, paramValue := range prepared {
		if strValue, ok := paramValue.(string); ok {
			// Handle JSONPath references like {$.steps.step_1.outputs.session_id}
			if strings.HasPrefix(strValue, "{") && strings.HasSuffix(strValue, "}") {
				jsonPath := strings.TrimPrefix(strings.TrimSuffix(strValue, "}"), "{")
				if resolvedValue := e.resolveJSONPath(jsonPath, context); resolvedValue != nil {
					prepared[paramName] = resolvedValue
				}
			} else if strings.HasPrefix(strValue, "${") && strings.HasSuffix(strValue, "}") {
				// Handle step reference format like ${step_1.session_id} or simple ${session_id}
				ref := strings.TrimPrefix(strings.TrimSuffix(strValue, "}"), "${")
				if strings.Contains(ref, ".") {
					// Step reference like ${step_1.session_id}
					if resolvedValue := e.resolveStepReference(ref, context); resolvedValue != nil {
						prepared[paramName] = resolvedValue
					}
				} else {
					// Simple reference like ${session_id} - check context directly
					if contextValue, exists := context[ref]; exists {
						prepared[paramName] = contextValue
					}
				}
			} else if strings.HasPrefix(strValue, "$") && strings.Contains(strValue, ".") {
				// Handle step reference format like $step_1.session_id (without braces)
				stepRef := strings.TrimPrefix(strValue, "$")
				if resolvedValue := e.resolveStepReference(stepRef, context); resolvedValue != nil {
					prepared[paramName] = resolvedValue
				}
			} else if strings.Contains(strValue, "$") {
				// Handle simple variable references like $variable
				for _, input := range inputs {
					if contextValue, exists := context[input]; exists {
						if strings.Contains(strValue, "$"+input) {
							prepared[paramName] = strings.ReplaceAll(strValue, "$"+input, fmt.Sprintf("%v", contextValue))
						}
					}
				}
			}
		}
	}

	return prepared
}

// resolveJSONPath resolves a JSONPath-like reference to a value from context
func (e *Executor) resolveJSONPath(jsonPath string, context map[string]interface{}) interface{} {
	// Handle simple JSONPath patterns like $.steps.step_1.outputs.session_id
	parts := strings.Split(jsonPath, ".")

	// Start from context
	current := context

	for i, part := range parts {
		if part == "$" {
			continue // Skip root reference
		}

		if i == len(parts)-1 {
			// Last part - this is the value we want
			if value, exists := current[part]; exists {
				return value
			}
		} else {
			// Navigate deeper
			if next, exists := current[part]; exists {
				if nextMap, ok := next.(map[string]interface{}); ok {
					current = nextMap
				} else {
					return nil // Can't navigate further
				}
			} else {
				return nil // Path not found
			}
		}
	}

	return nil
}

// resolveStepReference resolves a step reference like step_1.session_id
func (e *Executor) resolveStepReference(stepRef string, context map[string]interface{}) interface{} {
	// Parse step reference like "step_1.session_id"
	parts := strings.Split(stepRef, ".")
	if len(parts) != 2 {
		return nil // Invalid format
	}

	stepID := parts[0]     // "step_1"
	outputName := parts[1] // "session_id"

	// Check if steps structure exists
	if steps, exists := context["steps"]; exists {
		if stepsMap, ok := steps.(map[string]interface{}); ok {
			// Check if step exists
			if step, exists := stepsMap[stepID]; exists {
				if stepMap, ok := step.(map[string]interface{}); ok {
					// Check if outputs exist
					if outputs, exists := stepMap["outputs"]; exists {
						if outputsMap, ok := outputs.(map[string]interface{}); ok {
							// Return the output value
							if value, exists := outputsMap[outputName]; exists {
								return value
							}
						}
					}
				}
			}
		}
	}

	// Fallback: check flat context (for backward compatibility)
	if value, exists := context[outputName]; exists {
		return value
	}

	return nil
}

// updateContextWithOutputs updates the context with step outputs
func (e *Executor) updateContextWithOutputs(outputs []string, result interface{}, context map[string]interface{}) {
	if len(outputs) == 0 {
		return
	}

	// Ensure steps structure exists in context
	if _, exists := context["steps"]; !exists {
		context["steps"] = make(map[string]interface{})
	}
	steps := context["steps"].(map[string]interface{})

	// Get the current step ID from context (assuming it's stored there)
	stepID := "step_1" // Default fallback
	if currentStepID, exists := context["current_step_id"]; exists {
		if id, ok := currentStepID.(string); ok {
			stepID = id
		}
	}

	// Ensure step structure exists
	if _, exists := steps[stepID]; !exists {
		steps[stepID] = make(map[string]interface{})
	}
	step := steps[stepID].(map[string]interface{})

	// Ensure outputs structure exists
	if _, exists := step["outputs"]; !exists {
		step["outputs"] = make(map[string]interface{})
	}
	stepOutputs := step["outputs"].(map[string]interface{})

	// Store outputs in the structured format
	if len(outputs) == 1 {
		// Single output - try to extract from result if it's a map
		outputName := outputs[0]
		if resultMap, ok := result.(map[string]interface{}); ok {
			// Try to extract the field with the same name as the output
			if value, exists := resultMap[outputName]; exists {
				stepOutputs[outputName] = value
			} else {
				// Fallback: store the entire result
				stepOutputs[outputName] = result
			}
		} else {
			// Not a map, store the entire result
			stepOutputs[outputName] = result
		}
	} else {
		// Multiple outputs - try to extract from result map
		if resultMap, ok := result.(map[string]interface{}); ok {
			for _, output := range outputs {
				if value, exists := resultMap[output]; exists {
					stepOutputs[output] = value
				}
			}
		} else {
			// Fallback: store result in first output
			stepOutputs[outputs[0]] = result
		}
	}

	// Also store in flat context for backward compatibility
	if len(outputs) == 1 {
		// Single output - try to extract from result if it's a map
		outputName := outputs[0]
		if resultMap, ok := result.(map[string]interface{}); ok {
			// Try to extract the field with the same name as the output
			if value, exists := resultMap[outputName]; exists {
				context[outputName] = value
			} else {
				// Fallback: store the entire result
				context[outputName] = result
			}
		} else {
			// Not a map, store the entire result
			context[outputName] = result
		}
	} else {
		if resultMap, ok := result.(map[string]interface{}); ok {
			for _, output := range outputs {
				if value, exists := resultMap[output]; exists {
					context[output] = value
				}
			}
		} else {
			context[outputs[0]] = result
		}
	}
}

// findServiceURL finds the URL for a service
func (e *Executor) findServiceURL(serviceName string) string {
	// This is a simplified implementation
	// In a real system, you might have a service registry or configuration
	for _, url := range e.config.ServiceURLs {
		if strings.Contains(url, serviceName) || strings.Contains(serviceName, "r0d0") {
			return url
		}
	}
	return ""
}

// evaluateCondition evaluates a step condition
func (e *Executor) evaluateCondition(condition string, context map[string]interface{}) bool {
	// This is a simplified condition evaluator
	// In a real system, you might use a proper expression evaluator

	// Check if condition references a context variable
	if strings.HasPrefix(condition, "$") {
		contextKey := strings.TrimPrefix(condition, "$")
		if value, exists := context[contextKey]; exists {
			// Convert to boolean
			switch v := value.(type) {
			case bool:
				return v
			case string:
				return v != "" && v != "false" && v != "0"
			case int, int32, int64:
				return v != 0
			case float32, float64:
				return v != 0
			default:
				return value != nil
			}
		}
		return false
	}

	// Simple string-based conditions
	switch strings.ToLower(condition) {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return true // Default to true for unknown conditions
	}
}

// buildResponse builds the final orchestrator response
func (e *Executor) buildResponse(plan *WorkflowPlan, startTime time.Time, success bool, errorMsg string) *OrchestratorResponse {
	// Buscar respuesta conversacional generada por R0D0 en el resultado del último paso
	var conversational string
	if len(plan.Steps) > 0 {
		lastStep := plan.Steps[len(plan.Steps)-1]
		if lastStep.Result != nil {
			if resultMap, ok := lastStep.Result.(map[string]interface{}); ok {
				// Buscar en varios campos típicos (nivel superior) - incluir mayúsculas
				for _, key := range []string{"Response", "Message", "Respuesta", "message", "response", "respuesta", "text", "output"} {
					if val, ok := resultMap[key].(string); ok && val != "" {
						conversational = val
						break
					}
				}
				// Si no se encontró, buscar en result anidado
				if conversational == "" {
					if nested, ok := resultMap["result"].(map[string]interface{}); ok {
						for _, key := range []string{"Response", "Message", "Respuesta", "message", "response", "respuesta", "text", "output"} {
							if val, ok := nested[key].(string); ok && val != "" {
								conversational = val
								break
							}
						}
					}
				}
			}
		}
	}

	response := &OrchestratorResponse{
		RequestID:  plan.Context["request_id"].(string),
		WorkflowID: plan.ID,
		Response:   conversational,
		Result:     e.extractFinalResult(plan),
		Steps:      plan.Steps,
		Context:    plan.Context,
		Success:    success,
		Error:      errorMsg,
		Duration:   time.Since(startTime),
		Timestamp:  time.Now(),
	}

	// Si no hay respuesta conversacional, usar el método actual
	if response.Response == "" {
		response.Response = e.generateHumanReadableResponse(plan, success)
	}

	return response
}

// generateHumanReadableResponse generates a human-readable response
func (e *Executor) generateHumanReadableResponse(plan *WorkflowPlan, success bool) string {
	if !success {
		return e.generateErrorResponse(plan)
	}

	// Try to generate conversational response
	if conversationalResponse := e.generateConversationalResponse(plan); conversationalResponse != "" {
		return conversationalResponse
	}

	// Fallback to improved static responses
	return e.generateFallbackResponse(plan)
}

// generateConversationalResponse generates a natural, context-aware response
func (e *Executor) generateConversationalResponse(plan *WorkflowPlan) string {
	// Extract project context from the results
	projectContext := e.extractProjectContext(plan)

	// Get the final result to understand what was accomplished
	finalResult := e.extractFinalResult(plan)

	// Count completed steps
	completedSteps := 0
	for _, step := range plan.Steps {
		if step.Status == StepStatusCompleted {
			completedSteps++
		}
	}

	// Generate response based on what was accomplished
	if projectSlot := e.extractProjectSlot(finalResult); projectSlot != nil {
		return e.generateProjectDiscoveryResponse(projectSlot)
	}

	// If we have context about what was discovered
	if projectContext != "" {
		return e.generateContextualResponse(projectContext, completedSteps)
	}

	return ""
}

// extractProjectContext extracts meaningful project context from the workflow
func (e *Executor) extractProjectContext(plan *WorkflowPlan) string {
	// Look for project-related information in the context and results
	var contextParts []string

	// Check context for project info
	if userMessage, exists := plan.Context["user_message"]; exists {
		if msg, ok := userMessage.(string); ok && msg != "" {
			contextParts = append(contextParts, fmt.Sprintf("solicitud: %s", msg))
		}
	}

	// Check step results for discovery information
	for _, step := range plan.Steps {
		if step.Status == StepStatusCompleted && step.Result != nil {
			if resultMap, ok := step.Result.(map[string]interface{}); ok {
				// Look for project slot or discovery information
				if projectSlot, exists := resultMap["project_slot"]; exists {
					if ps, ok := projectSlot.(map[string]interface{}); ok {
						if name, exists := ps["name"]; exists {
							contextParts = append(contextParts, fmt.Sprintf("proyecto: %v", name))
						}
					}
				}
			}
		}
	}

	return strings.Join(contextParts, ", ")
}

// extractProjectSlot tries to extract a ProjectSlot from the final result
func (e *Executor) extractProjectSlot(result interface{}) map[string]interface{} {
	if result == nil {
		return nil
	}

	if resultMap, ok := result.(map[string]interface{}); ok {
		// Look for project_slot in the result
		if projectSlot, exists := resultMap["project_slot"]; exists {
			if ps, ok := projectSlot.(map[string]interface{}); ok {
				return ps
			}
		}

		// Check if the result itself looks like a project slot
		if _, hasName := resultMap["name"]; hasName {
			if _, hasObjective := resultMap["objective"]; hasObjective {
				if _, hasTimeline := resultMap["timeline"]; hasTimeline {
					return resultMap // This looks like a project slot
				}
			}
		}
	}

	return nil
}

// generateProjectDiscoveryResponse generates a response when a project was discovered
func (e *Executor) generateProjectDiscoveryResponse(projectSlot map[string]interface{}) string {
	responses := []string{
		"¡Excelente! Ya tengo una imagen clara de tu %s. He recopilado toda la información necesaria para crear una propuesta personalizada que se ajuste perfectamente a tu visión.",
		"¡Perfecto! Me encanta cómo ha quedado definido tu %s. Con todos los detalles que me has compartido, puedo crear una propuesta muy específica para tu proyecto.",
		"¡Genial! Tu %s está muy bien estructurado. He capturado todos los aspectos importantes para desarrollar una propuesta completa y detallada.",
	}

	projectName := "proyecto"
	if name, exists := projectSlot["name"]; exists {
		if nameStr, ok := name.(string); ok && nameStr != "" {
			projectName = strings.ToLower(nameStr)
		}
	}

	// Select a random response
	rand.Seed(time.Now().UnixNano())
	selectedResponse := responses[rand.Intn(len(responses))]

	return fmt.Sprintf(selectedResponse, projectName)
}

// generateContextualResponse generates a response based on available context
func (e *Executor) generateContextualResponse(context string, completedSteps int) string {
	responses := []string{
		"¡Perfecto! He procesado tu solicitud y recopilado toda la información necesaria. %s",
		"¡Excelente! He completado el análisis de tu proyecto. %s",
		"¡Genial! He terminado de procesar todos los detalles. %s",
	}

	contextMsg := "Los resultados están listos para el siguiente paso."
	if context != "" {
		contextMsg = fmt.Sprintf("Con base en %s, tengo todo lo necesario para continuar.", context)
	}

	// Select a random response
	rand.Seed(time.Now().UnixNano())
	selectedResponse := responses[rand.Intn(len(responses))]

	return fmt.Sprintf(selectedResponse, contextMsg)
}

// generateErrorResponse generates a natural error response
func (e *Executor) generateErrorResponse(plan *WorkflowPlan) string {
	errorResponses := []string{
		"Lo siento, encontré un problema al procesar tu solicitud. ¿Podrías intentar de nuevo?",
		"Disculpa, hubo un inconveniente técnico. Por favor, inténtalo nuevamente.",
		"Lamento informarte que no pude completar la operación. Vamos a intentarlo otra vez.",
	}

	rand.Seed(time.Now().UnixNano())
	return errorResponses[rand.Intn(len(errorResponses))]
}

// generateFallbackResponse generates improved static responses as fallback
func (e *Executor) generateFallbackResponse(plan *WorkflowPlan) string {
	// Count completed steps
	completedSteps := 0
	for _, step := range plan.Steps {
		if step.Status == StepStatusCompleted {
			completedSteps++
		}
	}

	if completedSteps == 0 {
		return "He procesado tu solicitud, aunque no fue necesario ejecutar pasos adicionales."
	}

	// More natural fallback responses
	fallbackResponses := []string{
		"¡Perfecto! He completado todo lo necesario para tu solicitud. Los resultados están listos.",
		"¡Excelente! He terminado de procesar tu proyecto. Todo está listo para continuar.",
		"¡Genial! He finalizado el análisis de tu solicitud. La información está completa.",
	}

	rand.Seed(time.Now().UnixNano())
	return fallbackResponses[rand.Intn(len(fallbackResponses))]
}

// extractFinalResult extracts the final result from the workflow
func (e *Executor) extractFinalResult(plan *WorkflowPlan) interface{} {
	// Return the result of the last completed step
	for i := len(plan.Steps) - 1; i >= 0; i-- {
		step := plan.Steps[i]
		if step.Status == StepStatusCompleted && step.Result != nil {
			return step.Result
		}
	}

	// Fallback to context
	return plan.Context
}
