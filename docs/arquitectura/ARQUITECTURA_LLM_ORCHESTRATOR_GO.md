# 🧠 Arquitectura LLM-Orchestrator en Go

**Fecha:** 2025-01-27  
**Versión:** 2.0-Go  
**Autor:** Daniel-EXLTK  
**Proyecto:** EXLTK-RPC con LLM Central en Go  

## 🎯 Concepto con Go

### **Ventajas de Go para esta Arquitectura**
- **Concurrencia nativa**: Goroutines para servicios paralelos
- **JSON-RPC integrado**: `net/rpc/jsonrpc` nativo
- **Alto rendimiento**: Compilado, mucho más rápido que Python
- **Deployment simple**: Binarios únicos sin dependencias
- **Networking**: `net/http` robusto para servicios distribuidos
- **Channels**: Comunicación entre goroutines elegante

### **Arquitectura**
```
Usuario → API Gateway (Go) → LLM Orchestrator (Go) → Servicios JSON-RPC (Go)
    ↓           ↓                    ↓                        ↓
  HTTP       Goroutines         Gemini API              JSON-RPC calls
```

## 🔧 Implementación en Go

### **1. Estructura del Proyecto**

```
exltk-rpc/
├── cmd/
│   ├── r0d0-service/
│   │   └── main.go
│   ├── proposal-service/
│   │   └── main.go
│   ├── orchestrator/
│   │   └── main.go
│   └── gateway/
│       └── main.go
├── internal/
│   ├── services/
│   │   ├── r0d0/
│   │   │   ├── service.go
│   │   │   └── types.go
│   │   └── proposal/
│   │       ├── service.go
│   │       └── types.go
│   ├── orchestrator/
│   │   ├── llm.go
│   │   ├── introspection.go
│   │   └── execution.go
│   └── common/
│       ├── jsonrpc.go
│       └── types.go
├── pkg/
│   └── gemini/
│       └── client.go
├── go.mod
├── go.sum
└── docker-compose.yml
```

### **2. Servicio R0D0 (Puerto 8501)**

#### **internal/services/r0d0/types.go**
```go
package r0d0

import "time"

type Service struct {
    Name        string             `json:"service"`
    Description string             `json:"description"`
    Methods     map[string]Method  `json:"methods"`
}

type Method struct {
    Description string                 `json:"description"`
    Params      map[string]string      `json:"params"`
    Returns     map[string]string      `json:"returns"`
}

type SurveyStartRequest struct {
    UserID  string `json:"user_id"`
    Message string `json:"message"`
}

type SurveyStartResponse struct {
    SessionID string   `json:"session_id"`
    Question  string   `json:"question"`
    Progress  Progress `json:"progress"`
}

type SurveyContinueRequest struct {
    SessionID string `json:"session_id"`
    Answer    string `json:"answer"`
}

type SurveyContinueResponse struct {
    NextQuestion string   `json:"next_question,omitempty"`
    Progress     Progress `json:"progress"`
    IsComplete   bool     `json:"is_complete"`
}

type SurveyCompleteRequest struct {
    SessionID string `json:"session_id"`
}

type SurveyCompleteResponse struct {
    ProjectSlot ProjectSlot `json:"project_slot"`
}

type Progress struct {
    Current int `json:"current"`
    Total   int `json:"total"`
}

type ProjectSlot struct {
    Name           string    `json:"name"`
    Objective      string    `json:"objective"`
    Budget         string    `json:"budget"`
    Timeline       string    `json:"timeline"`
    Technologies   string    `json:"technologies"`
    Audience       string    `json:"audience"`
    Risks          string    `json:"risks"`
    Resources      string    `json:"resources"`
    SuccessMetrics string    `json:"success_metrics"`
    CreatedAt      time.Time `json:"created_at"`
}

type Session struct {
    UserID    string            `json:"user_id"`
    Questions []string          `json:"questions"`
    Answers   map[int]string    `json:"answers"`
    Current   int               `json:"current"`
    CreatedAt time.Time         `json:"created_at"`
}
```

#### **internal/services/r0d0/service.go**
```go
package r0d0

import (
    "fmt"
    "sync"
    "time"
    "crypto/rand"
    "encoding/hex"
)

type R0D0Service struct {
    sessions map[string]*Session
    mutex    sync.RWMutex
}

func NewR0D0Service() *R0D0Service {
    return &R0D0Service{
        sessions: make(map[string]*Session),
    }
}

func (s *R0D0Service) Describe(req interface{}, resp *Service) error {
    *resp = Service{
        Name:        "r0d0_service",
        Description: "Servicio de encuestas de proyectos",
        Methods: map[string]Method{
            "survey.start": {
                Description: "Iniciar encuesta de proyecto",
                Params: map[string]string{
                    "user_id": "string",
                    "message": "string",
                },
                Returns: map[string]string{
                    "session_id": "string",
                    "question":   "string",
                    "progress":   "object",
                },
            },
            "survey.continue": {
                Description: "Continuar encuesta con respuesta",
                Params: map[string]string{
                    "session_id": "string",
                    "answer":     "string",
                },
                Returns: map[string]string{
                    "next_question": "string",
                    "progress":      "object",
                    "is_complete":   "boolean",
                },
            },
            "survey.complete": {
                Description: "Completar encuesta y generar ProjectSlot",
                Params: map[string]string{
                    "session_id": "string",
                },
                Returns: map[string]string{
                    "project_slot": "object",
                },
            },
        },
    }
    return nil
}

func (s *R0D0Service) SurveyStart(req *SurveyStartRequest, resp *SurveyStartResponse) error {
    sessionID := generateSessionID()
    
    questions := []string{
        "¿Cuál es el nombre de tu proyecto?",
        "¿Cuál es el objetivo principal?",
        "¿Cuál es el presupuesto estimado?",
        "¿Cuál es el timeline esperado?",
        "¿Qué tecnologías usarás?",
        "¿Quién es tu audiencia?",
        "¿Cuáles son los riesgos?",
        "¿Qué recursos necesitas?",
        "¿Cómo medirás el éxito?",
    }
    
    session := &Session{
        UserID:    req.UserID,
        Questions: questions,
        Answers:   make(map[int]string),
        Current:   0,
        CreatedAt: time.Now(),
    }
    
    s.mutex.Lock()
    s.sessions[sessionID] = session
    s.mutex.Unlock()
    
    *resp = SurveyStartResponse{
        SessionID: sessionID,
        Question:  questions[0],
        Progress: Progress{
            Current: 1,
            Total:   len(questions),
        },
    }
    
    return nil
}

func (s *R0D0Service) SurveyContinue(req *SurveyContinueRequest, resp *SurveyContinueResponse) error {
    s.mutex.RLock()
    session, exists := s.sessions[req.SessionID]
    s.mutex.RUnlock()
    
    if !exists {
        return fmt.Errorf("session not found: %s", req.SessionID)
    }
    
    // Guardar respuesta
    session.Answers[session.Current] = req.Answer
    session.Current++
    
    // Verificar si terminamos
    if session.Current >= len(session.Questions) {
        *resp = SurveyContinueResponse{
            Progress: Progress{
                Current: len(session.Questions),
                Total:   len(session.Questions),
            },
            IsComplete: true,
        }
        return nil
    }
    
    // Siguiente pregunta
    *resp = SurveyContinueResponse{
        NextQuestion: session.Questions[session.Current],
        Progress: Progress{
            Current: session.Current + 1,
            Total:   len(session.Questions),
        },
        IsComplete: false,
    }
    
    return nil
}

func (s *R0D0Service) SurveyComplete(req *SurveyCompleteRequest, resp *SurveyCompleteResponse) error {
    s.mutex.RLock()
    session, exists := s.sessions[req.SessionID]
    s.mutex.RUnlock()
    
    if !exists {
        return fmt.Errorf("session not found: %s", req.SessionID)
    }
    
    projectSlot := ProjectSlot{
        Name:           session.Answers[0],
        Objective:      session.Answers[1],
        Budget:         session.Answers[2],
        Timeline:       session.Answers[3],
        Technologies:   session.Answers[4],
        Audience:       session.Answers[5],
        Risks:          session.Answers[6],
        Resources:      session.Answers[7],
        SuccessMetrics: session.Answers[8],
        CreatedAt:      time.Now(),
    }
    
    *resp = SurveyCompleteResponse{
        ProjectSlot: projectSlot,
    }
    
    return nil
}

func generateSessionID() string {
    bytes := make([]byte, 8)
    rand.Read(bytes)
    return fmt.Sprintf("survey_%d_%s", time.Now().Unix(), hex.EncodeToString(bytes))
}
```

#### **cmd/r0d0-service/main.go**
```go
package main

import (
    "log"
    "net"
    "net/http"
    "net/rpc"
    "net/rpc/jsonrpc"
    
    "github.com/exltk/exltk-rpc/internal/services/r0d0"
)

func main() {
    service := r0d0.NewR0D0Service()
    
    // Registrar servicio RPC
    server := rpc.NewServer()
    server.Register(service)
    
    // Crear handler HTTP para JSON-RPC
    http.HandleFunc("/jsonrpc", func(w http.ResponseWriter, r *http.Request) {
        // Configurar CORS
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        
        if r.Method == "OPTIONS" {
            return
        }
        
        // Servir JSON-RPC
        server.ServeCodec(jsonrpc.NewServerCodec(&HTTPConn{
            in:  r.Body,
            out: w,
        }))
    })
    
    log.Println("🤖 R0D0 Service iniciado en puerto 8501")
    log.Println("📍 Endpoint: http://localhost:8501/jsonrpc")
    log.Fatal(http.ListenAndServe(":8501", nil))
}

// HTTPConn adapta HTTP request/response para JSON-RPC
type HTTPConn struct {
    in  io.ReadCloser
    out io.Writer
}

func (c *HTTPConn) Read(p []byte) (n int, err error) {
    return c.in.Read(p)
}

func (c *HTTPConn) Write(p []byte) (n int, err error) {
    return c.out.Write(p)
}

func (c *HTTPConn) Close() error {
    return c.in.Close()
}
```

### **3. Servicio Proposal (Puerto 8503)**

#### **internal/services/proposal/types.go**
```go
package proposal

import "time"

type Service struct {
    Name        string             `json:"service"`
    Description string             `json:"description"`
    Methods     map[string]Method  `json:"methods"`
}

type Method struct {
    Description string                 `json:"description"`
    Params      map[string]string      `json:"params"`
    Returns     map[string]string      `json:"returns"`
}

type GenerateRequest struct {
    ProjectSlot ProjectSlot `json:"project_slot"`
    Template    string      `json:"template"`
}

type GenerateResponse struct {
    ProposalID string `json:"proposal_id"`
    Content    string `json:"content"`
}

type ExportRequest struct {
    ProposalID string `json:"proposal_id"`
    Format     string `json:"format"`
}

type ExportResponse struct {
    DownloadURL string `json:"download_url"`
}

type ProjectSlot struct {
    Name           string    `json:"name"`
    Objective      string    `json:"objective"`
    Budget         string    `json:"budget"`
    Timeline       string    `json:"timeline"`
    Technologies   string    `json:"technologies"`
    Audience       string    `json:"audience"`
    Risks          string    `json:"risks"`
    Resources      string    `json:"resources"`
    SuccessMetrics string    `json:"success_metrics"`
    CreatedAt      time.Time `json:"created_at"`
}

type Proposal struct {
    ID          string      `json:"id"`
    Content     string      `json:"content"`
    ProjectSlot ProjectSlot `json:"project_slot"`
    Template    string      `json:"template"`
    CreatedAt   time.Time   `json:"created_at"`
}
```

#### **internal/services/proposal/service.go**
```go
package proposal

import (
    "fmt"
    "sync"
    "time"
    "crypto/rand"
    "encoding/hex"
    "strings"
)

type ProposalService struct {
    proposals map[string]*Proposal
    templates map[string]string
    mutex     sync.RWMutex
}

func NewProposalService() *ProposalService {
    return &ProposalService{
        proposals: make(map[string]*Proposal),
        templates: map[string]string{
            "basic": `# {{.Name}}

## Objetivo
{{.Objective}}

## Presupuesto
{{.Budget}}

## Timeline
{{.Timeline}}

## Tecnologías
{{.Technologies}}

## Audiencia
{{.Audience}}

## Riesgos
{{.Risks}}

## Recursos
{{.Resources}}

## Métricas de Éxito
{{.SuccessMetrics}}`,
            "technical": `# Propuesta Técnica: {{.Name}}

## Resumen Ejecutivo
{{.Objective}}

## Enfoque Técnico
Tecnologías: {{.Technologies}}

## Timeline del Proyecto
{{.Timeline}}

## Presupuesto
{{.Budget}}

## Evaluación de Riesgos
{{.Risks}}

## Recursos Necesarios
{{.Resources}}

## Criterios de Éxito
{{.SuccessMetrics}}

## Audiencia Objetivo
{{.Audience}}`,
            "executive": `# Propuesta Ejecutiva: {{.Name}}

## Objetivo de Negocio
{{.Objective}}

## Inversión Requerida
{{.Budget}}

## Timeline y Milestones
{{.Timeline}}

## Mercado Objetivo
{{.Audience}}

## Stack Tecnológico
{{.Technologies}}

## Mitigación de Riesgos
{{.Risks}}

## Asignación de Recursos
{{.Resources}}

## Indicadores Clave de Rendimiento
{{.SuccessMetrics}}`,
        },
    }
}

func (s *ProposalService) Describe(req interface{}, resp *Service) error {
    *resp = Service{
        Name:        "proposal_service",
        Description: "Servicio de generación de propuestas",
        Methods: map[string]Method{
            "proposal.generate": {
                Description: "Generar propuesta desde ProjectSlot",
                Params: map[string]string{
                    "project_slot": "object",
                    "template":     "string",
                },
                Returns: map[string]string{
                    "proposal_id": "string",
                    "content":     "string",
                },
            },
            "proposal.export": {
                Description: "Exportar propuesta a formato",
                Params: map[string]string{
                    "proposal_id": "string",
                    "format":      "string",
                },
                Returns: map[string]string{
                    "download_url": "string",
                },
            },
        },
    }
    return nil
}

func (s *ProposalService) Generate(req *GenerateRequest, resp *GenerateResponse) error {
    templateName := req.Template
    if templateName == "" {
        templateName = "basic"
    }
    
    template, exists := s.templates[templateName]
    if !exists {
        return fmt.Errorf("template not found: %s", templateName)
    }
    
    // Reemplazar variables en el template
    content := template
    content = strings.ReplaceAll(content, "{{.Name}}", req.ProjectSlot.Name)
    content = strings.ReplaceAll(content, "{{.Objective}}", req.ProjectSlot.Objective)
    content = strings.ReplaceAll(content, "{{.Budget}}", req.ProjectSlot.Budget)
    content = strings.ReplaceAll(content, "{{.Timeline}}", req.ProjectSlot.Timeline)
    content = strings.ReplaceAll(content, "{{.Technologies}}", req.ProjectSlot.Technologies)
    content = strings.ReplaceAll(content, "{{.Audience}}", req.ProjectSlot.Audience)
    content = strings.ReplaceAll(content, "{{.Risks}}", req.ProjectSlot.Risks)
    content = strings.ReplaceAll(content, "{{.Resources}}", req.ProjectSlot.Resources)
    content = strings.ReplaceAll(content, "{{.SuccessMetrics}}", req.ProjectSlot.SuccessMetrics)
    
    proposalID := generateProposalID()
    
    proposal := &Proposal{
        ID:          proposalID,
        Content:     content,
        ProjectSlot: req.ProjectSlot,
        Template:    templateName,
        CreatedAt:   time.Now(),
    }
    
    s.mutex.Lock()
    s.proposals[proposalID] = proposal
    s.mutex.Unlock()
    
    *resp = GenerateResponse{
        ProposalID: proposalID,
        Content:    content,
    }
    
    return nil
}

func (s *ProposalService) Export(req *ExportRequest, resp *ExportResponse) error {
    s.mutex.RLock()
    proposal, exists := s.proposals[req.ProposalID]
    s.mutex.RUnlock()
    
    if !exists {
        return fmt.Errorf("proposal not found: %s", req.ProposalID)
    }
    
    // Simular exportación (en implementación real sería PDF/DOCX/HTML)
    downloadURL := fmt.Sprintf("http://localhost:8503/download/%s.%s", 
        req.ProposalID, req.Format)
    
    *resp = ExportResponse{
        DownloadURL: downloadURL,
    }
    
    return nil
}

func generateProposalID() string {
    bytes := make([]byte, 8)
    rand.Read(bytes)
    return fmt.Sprintf("prop_%d_%s", time.Now().Unix(), hex.EncodeToString(bytes))
}
```

### **4. LLM Orchestrator**

#### **pkg/gemini/client.go**
```go
package gemini

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
)

type Client struct {
    apiKey string
    client *http.Client
}

type GenerateRequest struct {
    Contents []Content `json:"contents"`
    GenerationConfig GenerationConfig `json:"generationConfig"`
}

type Content struct {
    Parts []Part `json:"parts"`
}

type Part struct {
    Text string `json:"text"`
}

type GenerationConfig struct {
    Temperature     float64 `json:"temperature"`
    MaxOutputTokens int     `json:"maxOutputTokens"`
}

type GenerateResponse struct {
    Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
    Content Content `json:"content"`
}

func NewClient() *Client {
    return &Client{
        apiKey: os.Getenv("GEMINI_API_KEY"),
        client: &http.Client{},
    }
}

func (c *Client) Generate(prompt string) (string, error) {
    url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-exp:generateContent?key=%s", c.apiKey)
    
    req := GenerateRequest{
        Contents: []Content{
            {
                Parts: []Part{
                    {Text: prompt},
                },
            },
        },
        GenerationConfig: GenerationConfig{
            Temperature:     0.7,
            MaxOutputTokens: 2048,
        },
    }
    
    jsonData, err := json.Marshal(req)
    if err != nil {
        return "", err
    }
    
    httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return "", err
    }
    
    httpReq.Header.Set("Content-Type", "application/json")
    
    resp, err := c.client.Do(httpReq)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }
    
    var genResp GenerateResponse
    if err := json.Unmarshal(body, &genResp); err != nil {
        return "", err
    }
    
    if len(genResp.Candidates) == 0 || len(genResp.Candidates[0].Content.Parts) == 0 {
        return "", fmt.Errorf("no response from Gemini")
    }
    
    return genResp.Candidates[0].Content.Parts[0].Text, nil
}
```

#### **internal/orchestrator/llm.go**
```go
package orchestrator

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "sync"
    
    "github.com/exltk/exltk-rpc/pkg/gemini"
)

type LLMOrchestrator struct {
    agents map[string]string
    client *gemini.Client
    mutex  sync.RWMutex
}

type AgentCapabilities struct {
    Service     string             `json:"service"`
    Description string             `json:"description"`
    Methods     map[string]Method  `json:"methods"`
}

type Method struct {
    Description string                 `json:"description"`
    Params      map[string]string      `json:"params"`
    Returns     map[string]string      `json:"returns"`
}

type WorkflowPlan struct {
    Analysis string         `json:"analysis"`
    Workflow []WorkflowStep `json:"workflow"`
}

type WorkflowStep struct {
    Step        int                    `json:"step"`
    Service     string                 `json:"service"`
    Method      string                 `json:"method"`
    Params      map[string]interface{} `json:"params"`
    Description string                 `json:"description"`
}

type ProcessResult struct {
    Response string                 `json:"response"`
    Plan     WorkflowPlan           `json:"plan"`
    Results  []StepResult           `json:"results"`
}

type StepResult struct {
    Step    int                    `json:"step"`
    Service string                 `json:"service"`
    Method  string                 `json:"method"`
    Result  map[string]interface{} `json:"result"`
}

func NewLLMOrchestrator() *LLMOrchestrator {
    return &LLMOrchestrator{
        agents: map[string]string{
            "r0d0":     "http://localhost:8501",
            "proposal": "http://localhost:8503",
        },
        client: gemini.NewClient(),
    }
}

func (o *LLMOrchestrator) IntrospectAgents() (map[string]AgentCapabilities, error) {
    capabilities := make(map[string]AgentCapabilities)
    
    // Usar goroutines para introspección paralela
    var wg sync.WaitGroup
    var mu sync.Mutex
    
    for name, url := range o.agents {
        wg.Add(1)
        go func(agentName, agentURL string) {
            defer wg.Done()
            
            cap, err := o.introspectAgent(agentURL)
            if err != nil {
                fmt.Printf("Error introspecting %s: %v\n", agentName, err)
                return
            }
            
            mu.Lock()
            capabilities[agentName] = cap
            mu.Unlock()
        }(name, url)
    }
    
    wg.Wait()
    return capabilities, nil
}

func (o *LLMOrchestrator) introspectAgent(agentURL string) (AgentCapabilities, error) {
    reqBody := map[string]interface{}{
        "jsonrpc": "2.0",
        "method":  "Describe",
        "params":  map[string]interface{}{},
        "id":      1,
    }
    
    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return AgentCapabilities{}, err
    }
    
    resp, err := http.Post(agentURL+"/jsonrpc", "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return AgentCapabilities{}, err
    }
    defer resp.Body.Close()
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return AgentCapabilities{}, err
    }
    
    var rpcResp struct {
        Result AgentCapabilities `json:"result"`
    }
    
    if err := json.Unmarshal(body, &rpcResp); err != nil {
        return AgentCapabilities{}, err
    }
    
    return rpcResp.Result, nil
}

func (o *LLMOrchestrator) ProcessMessage(userMessage, userID string) (*ProcessResult, error) {
    // 1. Introspección
    capabilities, err := o.IntrospectAgents()
    if err != nil {
        return nil, err
    }
    
    // 2. Planificación con LLM
    plan, err := o.planWorkflow(userMessage, capabilities)
    if err != nil {
        return nil, err
    }
    
    // 3. Ejecución
    results, err := o.executeWorkflow(plan, userID)
    if err != nil {
        return nil, err
    }
    
    // 4. Generar respuesta final
    finalResponse, err := o.generateResponse(userMessage, results)
    if err != nil {
        return nil, err
    }
    
    return &ProcessResult{
        Response: finalResponse,
        Plan:     *plan,
        Results:  results,
    }, nil
}

func (o *LLMOrchestrator) planWorkflow(userMessage string, capabilities map[string]AgentCapabilities) (*WorkflowPlan, error) {
    capJSON, _ := json.MarshalIndent(capabilities, "", "  ")
    
    prompt := fmt.Sprintf(`
USUARIO: "%s"

SERVICIOS DISPONIBLES:
%s

Analiza el mensaje y crea un plan de ejecución JSON:
{
    "analysis": "qué quiere el usuario",
    "workflow": [
        {
            "step": 1,
            "service": "nombre_servicio",
            "method": "método_a_llamar",
            "params": {},
            "description": "qué hace"
        }
    ]
}`, userMessage, string(capJSON))
    
    response, err := o.client.Generate(prompt)
    if err != nil {
        return nil, err
    }
    
    var plan WorkflowPlan
    if err := json.Unmarshal([]byte(response), &plan); err != nil {
        return nil, err
    }
    
    return &plan, nil
}

func (o *LLMOrchestrator) executeWorkflow(plan *WorkflowPlan, userID string) ([]StepResult, error) {
    results := make([]StepResult, 0)
    context := map[string]interface{}{
        "user_id": userID,
    }
    
    for _, step := range plan.Workflow {
        result, err := o.executeStep(step, context)
        if err != nil {
            return nil, err
        }
        
        results = append(results, result)
        
        // Actualizar contexto para siguientes pasos
        for key, value := range result.Result {
            context[key] = value
        }
    }
    
    return results, nil
}

func (o *LLMOrchestrator) executeStep(step WorkflowStep, context map[string]interface{}) (StepResult, error) {
    agentURL := o.agents[step.Service]
    
    // Reemplazar variables del contexto
    params := step.Params
    for key, value := range params {
        if str, ok := value.(string); ok && len(str) > 0 && str[0] == '$' {
            if contextValue, exists := context[str[1:]]; exists {
                params[key] = contextValue
            }
        }
    }
    
    // Ejecutar llamada JSON-RPC
    reqBody := map[string]interface{}{
        "jsonrpc": "2.0",
        "method":  step.Method,
        "params":  params,
        "id":      step.Step,
    }
    
    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return StepResult{}, err
    }
    
    resp, err := http.Post(agentURL+"/jsonrpc", "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return StepResult{}, err
    }
    defer resp.Body.Close()
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return StepResult{}, err
    }
    
    var rpcResp struct {
        Result map[string]interface{} `json:"result"`
    }
    
    if err := json.Unmarshal(body, &rpcResp); err != nil {
        return StepResult{}, err
    }
    
    return StepResult{
        Step:    step.Step,
        Service: step.Service,
        Method:  step.Method,
        Result:  rpcResp.Result,
    }, nil
}

func (o *LLMOrchestrator) generateResponse(userMessage string, results []StepResult) (string, error) {
    resultsJSON, _ := json.MarshalIndent(results, "", "  ")
    
    prompt := fmt.Sprintf(`
El usuario preguntó: "%s"

RESULTADOS EJECUTADOS:
%s

Genera una respuesta natural y útil en español explicando:
1. Qué se hizo
2. Los resultados obtenidos
3. Próximos pasos si los hay

Responde de manera conversacional y profesional.`, userMessage, string(resultsJSON))
    
    return o.client.Generate(prompt)
}
```

### **5. API Gateway**

#### **cmd/gateway/main.go**
```go
package main

import (
    "encoding/json"
    "log"
    "net/http"
    
    "github.com/exltk/exltk-rpc/internal/orchestrator"
)

type Gateway struct {
    orchestrator *orchestrator.LLMOrchestrator
}

type ChatRequest struct {
    Message string `json:"message"`
    UserID  string `json:"user_id"`
}

type ChatResponse struct {
    Response string                              `json:"response"`
    Plan     orchestrator.WorkflowPlan          `json:"plan"`
    Results  []orchestrator.StepResult          `json:"results"`
}

func NewGateway() *Gateway {
    return &Gateway{
        orchestrator: orchestrator.NewLLMOrchestrator(),
    }
}

func (g *Gateway) handleChat(w http.ResponseWriter, r *http.Request) {
    // Configurar CORS
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
    w.Header().Set("Content-Type", "application/json")
    
    if r.Method == "OPTIONS" {
        return
    }
    
    var req ChatRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    userID := req.UserID
    if userID == "" {
        userID = "anonymous"
    }
    
    // Procesar mensaje con LLM Orchestrator
    result, err := g.orchestrator.ProcessMessage(req.Message, userID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    response := ChatResponse{
        Response: result.Response,
        Plan:     result.Plan,
        Results:  result.Results,
    }
    
    json.NewEncoder(w).Encode(response)
}

func (g *Gateway) handleCapabilities(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Content-Type", "application/json")
    
    capabilities, err := g.orchestrator.IntrospectAgents()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(capabilities)
}

func main() {
    gateway := NewGateway()
    
    http.HandleFunc("/chat", gateway.handleChat)
    http.HandleFunc("/capabilities", gateway.handleCapabilities)
    
    log.Println("🚀 API Gateway iniciado en puerto 8500")
    log.Println("📍 Endpoints:")
    log.Println("  - POST /chat")
    log.Println("  - GET /capabilities")
    log.Fatal(http.ListenAndServe(":8500", nil))
}
```

### **6. Docker y Deployment**

#### **go.mod**
```go
module github.com/exltk/exltk-rpc

go 1.21

require (
    // No external dependencies needed!
    // Go standard library provides everything
)
```

#### **docker-compose.yml**
```yaml
version: '3.8'
services:
  r0d0-service:
    build:
      context: .
      dockerfile: Dockerfile.r0d0
    ports:
      - "8501:8501"
    environment:
      - PORT=8501
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8501/jsonrpc"]
      interval: 30s
      timeout: 10s
      retries: 3

  proposal-service:
    build:
      context: .
      dockerfile: Dockerfile.proposal
    ports:
      - "8503:8503"
    environment:
      - PORT=8503
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8503/jsonrpc"]
      interval: 30s
      timeout: 10s
      retries: 3

  gateway:
    build:
      context: .
      dockerfile: Dockerfile.gateway
    ports:
      - "8500:8500"
    environment:
      - GEMINI_API_KEY=${GEMINI_API_KEY}
    depends_on:
      - r0d0-service
      - proposal-service
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8500/capabilities"]
      interval: 30s
      timeout: 10s
      retries: 3
```

#### **Dockerfile.base**
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/${SERVICE_NAME}

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE ${PORT}
CMD ["./main"]
```

## 🚀 Ventajas de Go para esta Arquitectura

### **1. Rendimiento**
- **Compilado**: 10-50x más rápido que Python
- **Concurrencia**: Goroutines para servicios paralelos
- **Memoria**: Uso eficiente de memoria

### **2. Simplicidad**
- **JSON-RPC nativo**: Sin dependencias externas
- **Deployment**: Binario único
- **Cross-platform**: Linux, Windows, macOS

### **3. Escalabilidad**
- **Goroutines**: Miles de conexiones concurrentes
- **Channels**: Comunicación segura entre goroutines
- **Load balancing**: Fácil replicación horizontal

### **4. Mantenibilidad**
- **Tipado fuerte**: Errores en tiempo de compilación
- **Estándar**: `go fmt`, `go vet`, `go test`
- **Documentación**: `godoc` integrado

## 🎯 Plan de Implementación

### **Fase 1: Servicios Base (1 semana)**
- Implementar r0d0-service en Go
- Implementar proposal-service en Go
- JSON-RPC endpoints con método Describe

### **Fase 2: LLM Orchestrator (1 semana)**
- Cliente Gemini en Go
- Introspección paralela con goroutines
- Planificación y ejecución de workflows

### **Fase 3: API Gateway (1 semana)**
- Gateway HTTP con endpoints /chat y /capabilities
- CORS y manejo de errores
- Integración con LLM Orchestrator

### **Fase 4: Docker y Testing (1 semana)**
- Dockerfiles para cada servicio
- docker-compose.yml completo
- Tests unitarios y de integración

## 🎉 Beneficios Esperados

1. **Rendimiento**: 10-50x más rápido que Python
2. **Simplicidad**: Sin dependencias externas complejas
3. **Deployment**: Binarios únicos fáciles de distribuir
4. **Escalabilidad**: Concurrencia nativa con goroutines
5. **Mantenibilidad**: Tipado fuerte y herramientas estándar

¿Te parece que empecemos con la implementación en Go? 