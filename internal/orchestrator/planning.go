// Package orchestrator provides LLM-based workflow planning capabilities
package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Daniel-EXLTK/exltk-rpc/pkg/gemini"
	"github.com/google/uuid"
)

// Planner handles LLM-based workflow planning
type Planner struct {
	config *OrchestratorConfig
	client *gemini.Client
}

// NewPlanner creates a new planner instance
func NewPlanner(config *OrchestratorConfig) (*Planner, error) {
	// Determine API key to use
	apiKey := config.LLMAPIKey
	if apiKey == "" {
		return nil, fmt.Errorf("LLM API key is required")
	}

	// Create Gemini client
	client := gemini.NewClient(
		apiKey,
		gemini.WithTimeout(config.RequestTimeout),
		gemini.WithMaxRetries(3),
	)

	return &Planner{
		config: config,
		client: client,
	}, nil
}

// PlanWorkflow generates a workflow plan using the LLM
func (p *Planner) PlanWorkflow(ctx context.Context, request *UserRequest, capabilities []ServiceCapability) (*WorkflowPlan, error) {
	startTime := time.Now()

	if p.config.EnableLogging {
		log.Printf("Planning workflow for request: %s", request.ID)
	}

	// Build prompt for the LLM
	prompt := p.buildPlanningPrompt(request, capabilities)

	// Generate plan using LLM
	response, err := p.client.GenerateWithContext(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate workflow plan: %w", err)
	}

	// Parse LLM response into workflow plan
	plan, err := p.parseWorkflowPlan(response, request)
	if err != nil {
		return nil, fmt.Errorf("failed to parse workflow plan: %w", err)
	}
	if p.config.EnableLogging {
		log.Printf("[LLM PLAN] Plan generado por Gemini: %s", response)
	}

	// Validate and enhance the plan
	if err := p.validateAndEnhancePlan(plan, capabilities); err != nil {
		return nil, fmt.Errorf("failed to validate workflow plan: %w", err)
	}

	if p.config.EnableLogging {
		log.Printf("Workflow planning completed in %v", time.Since(startTime))
	}

	return plan, nil
}

// buildPlanningPrompt creates the prompt for the LLM
func (p *Planner) buildPlanningPrompt(request *UserRequest, capabilities []ServiceCapability) string {
	var prompt strings.Builder

	prompt.WriteString(`Eres un orquestador inteligente que coordina servicios JSON-RPC. Tu tarea es crear un plan de workflow basado en la solicitud del usuario y las capacidades de los servicios disponibles.

SOLICITUD DEL USUARIO:
`)
	prompt.WriteString(request.Message)
	prompt.WriteString(`

SERVICIOS DISPONIBLES:
`)

	for _, capability := range capabilities {
		prompt.WriteString(fmt.Sprintf("\n- %s (%s): %s\n",
			capability.ServiceName, capability.ServiceURL, capability.Description))

		prompt.WriteString("  Métodos disponibles:\n")
		for methodName, method := range capability.Methods {
			prompt.WriteString(fmt.Sprintf("    - %s: %s\n", methodName, method.Description))
			if len(method.Parameters) > 0 {
				prompt.WriteString("      Parámetros:\n")
				for _, param := range method.Parameters {
					required := "opcional"
					if param.Required {
						required = "requerido"
					}
					prompt.WriteString(fmt.Sprintf("        - %s (%s, %s): %s\n",
						param.Name, param.Type, required, param.Description))
				}
			}
		}
	}

	prompt.WriteString(`

INSTRUCCIONES:
1. Analiza la solicitud del usuario
2. Identifica qué servicios y métodos necesitas usar
3. Si el método DiscoveryStart inicia una sesión, agrega pasos subsiguientes usando DiscoveryContinue y DiscoveryComplete para completar la conversación y obtener todos los datos necesarios.
4. Crea un plan de workflow con pasos secuenciales
5. Define las entradas y salidas de cada paso
6. Para referenciar outputs de pasos anteriores, usa el formato: ${step_id.output_name}
   Por ejemplo: ${step_1.session_id} para usar el session_id del step_1
7. Considera el manejo de errores y reintentos

FORMATO DE RESPUESTA (JSON):
{
  "description": "Descripción del workflow",
  "steps": [
    {
      "id": "step_1",
      "name": "Nombre del paso",
      "service": "nombre_del_servicio",
      "method": "nombre_del_método",
      "parameters": {
        "param1": "valor1",
        "param2": "valor2"
      },
      "inputs": ["variable1", "variable2"],
      "outputs": ["resultado1", "resultado2"],
      "condition": "condición_opcional",
      "retry_policy": {
        "max_retries": 3,
        "delay": "5s",
        "backoff": 2.0
      }
    }
  ],
  "context": {
    "variable_inicial": "valor_inicial"
  }
}

EJEMPLO DE REFERENCIA DE OUTPUTS:
Si step_1 produce un output llamado "session_id", el siguiente paso puede usarlo así:
{
  "id": "step_2",
  "parameters": {
    "session_id": "${step_1.session_id}"
  },
  "inputs": ["session_id"]
}

Responde SOLO con el JSON del plan de workflow, sin texto adicional.`)

	return prompt.String()
}

// parseWorkflowPlan parses the LLM response into a WorkflowPlan
func (p *Planner) parseWorkflowPlan(llmResponse string, request *UserRequest) (*WorkflowPlan, error) {
	// Clean the response - remove markdown code blocks if present
	response := strings.TrimSpace(llmResponse)
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
	}
	if strings.HasPrefix(response, "```") {
		response = strings.TrimPrefix(response, "```")
	}
	if strings.HasSuffix(response, "```") {
		response = strings.TrimSuffix(response, "```")
	}
	response = strings.TrimSpace(response)

	// Parse JSON
	var planData map[string]interface{}
	if err := json.Unmarshal([]byte(response), &planData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Create workflow plan
	plan := &WorkflowPlan{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
		Status:    WorkflowStatusPending,
		Context:   make(map[string]interface{}),
	}

	// Extract description
	if desc, exists := planData["description"].(string); exists {
		plan.Description = desc
	}

	// Extract context
	if contextData, exists := planData["context"].(map[string]interface{}); exists {
		plan.Context = contextData
	}

	// Add request context
	plan.Context["user_id"] = request.UserID
	plan.Context["user_message"] = request.Message
	plan.Context["request_id"] = request.ID

	// Extract steps
	if stepsData, exists := planData["steps"].([]interface{}); exists {
		for i, stepData := range stepsData {
			step, err := p.parseWorkflowStep(stepData, i+1)
			if err != nil {
				return nil, fmt.Errorf("failed to parse step %d: %w", i+1, err)
			}
			plan.Steps = append(plan.Steps, *step)
		}
	}

	return plan, nil
}

// parseWorkflowStep parses a single workflow step
func (p *Planner) parseWorkflowStep(stepData interface{}, stepNumber int) (*WorkflowStep, error) {
	stepMap, ok := stepData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("step data is not a map")
	}

	step := &WorkflowStep{
		ID:     fmt.Sprintf("step_%d", stepNumber),
		Status: StepStatusPending,
	}

	// Extract basic fields
	if name, exists := stepMap["name"].(string); exists {
		step.Name = name
	} else {
		step.Name = fmt.Sprintf("Paso %d", stepNumber)
	}

	if service, exists := stepMap["service"].(string); exists {
		step.Service = service
	}

	if method, exists := stepMap["method"].(string); exists {
		step.Method = method
	}

	if condition, exists := stepMap["condition"].(string); exists {
		step.Condition = condition
	}

	// Extract parameters
	if params, exists := stepMap["parameters"].(map[string]interface{}); exists {
		step.Parameters = params
	} else {
		step.Parameters = make(map[string]interface{})
	}

	// Extract inputs
	if inputs, exists := stepMap["inputs"].([]interface{}); exists {
		for _, input := range inputs {
			if inputStr, ok := input.(string); ok {
				step.Inputs = append(step.Inputs, inputStr)
			}
		}
	}

	// Extract outputs
	if outputs, exists := stepMap["outputs"].([]interface{}); exists {
		for _, output := range outputs {
			if outputStr, ok := output.(string); ok {
				step.Outputs = append(step.Outputs, outputStr)
			}
		}
	}

	// Extract retry policy
	if retryData, exists := stepMap["retry_policy"].(map[string]interface{}); exists {
		retryPolicy := &RetryPolicy{
			MaxRetries: 3,
			Delay:      5 * time.Second,
			Backoff:    2.0,
		}

		if maxRetries, exists := retryData["max_retries"].(float64); exists {
			retryPolicy.MaxRetries = int(maxRetries)
		}

		if delay, exists := retryData["delay"].(string); exists {
			if parsed, err := time.ParseDuration(delay); err == nil {
				retryPolicy.Delay = parsed
			}
		}

		if backoff, exists := retryData["backoff"].(float64); exists {
			retryPolicy.Backoff = backoff
		}

		step.RetryPolicy = retryPolicy
	}

	return step, nil
}

// validateAndEnhancePlan validates and enhances the generated plan
func (p *Planner) validateAndEnhancePlan(plan *WorkflowPlan, capabilities []ServiceCapability) error {
	// Create capability lookup map
	capabilityMap := make(map[string]*ServiceCapability)
	for i := range capabilities {
		capabilityMap[capabilities[i].ServiceName] = &capabilities[i]
	}

	// Validate each step
	for i := range plan.Steps {
		step := &plan.Steps[i]

		// Validate service exists
		if step.Service == "" {
			return fmt.Errorf("step %s: service name is required", step.ID)
		}

		capability, exists := capabilityMap[step.Service]
		if !exists {
			return fmt.Errorf("step %s: service '%s' not found in capabilities", step.ID, step.Service)
		}

		// Validate method exists
		if step.Method == "" {
			return fmt.Errorf("step %s: method name is required", step.ID)
		}

		method, exists := capability.Methods[step.Method]
		if !exists {
			return fmt.Errorf("step %s: method '%s' not found in service '%s'", step.ID, step.Method, step.Service)
		}

		// Validate required parameters
		for _, param := range method.Parameters {
			if param.Required {
				if _, exists := step.Parameters[param.Name]; !exists {
					return fmt.Errorf("step %s: required parameter '%s' is missing", step.ID, param.Name)
				}
			}
		}

		// Set default retry policy if not specified
		if step.RetryPolicy == nil {
			step.RetryPolicy = &RetryPolicy{
				MaxRetries: 3,
				Delay:      5 * time.Second,
				Backoff:    2.0,
			}
		}
	}

	return nil
}

// EnhancePlanWithContext enhances the plan with additional context and optimizations
func (p *Planner) EnhancePlanWithContext(plan *WorkflowPlan, userContext map[string]interface{}) {
	// Merge user context into plan context
	for key, value := range userContext {
		plan.Context[key] = value
	}

	// Add execution metadata
	plan.Context["workflow_id"] = plan.ID
	plan.Context["created_at"] = plan.CreatedAt
	plan.Context["llm_provider"] = p.config.LLMProvider
	plan.Context["llm_model"] = p.config.LLMModel

	// Optimize step parameters with context values
	for i := range plan.Steps {
		step := &plan.Steps[i]

		// Replace parameter placeholders with context values
		for paramName, paramValue := range step.Parameters {
			if strValue, ok := paramValue.(string); ok {
				if strings.HasPrefix(strValue, "$") {
					contextKey := strings.TrimPrefix(strValue, "$")
					if contextValue, exists := plan.Context[contextKey]; exists {
						step.Parameters[paramName] = contextValue
					}
				}
			}
		}
	}
}
