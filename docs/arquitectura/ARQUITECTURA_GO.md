# 🧠 Arquitectura LLM-Orchestrator en Go

**Fecha:** 2025-01-27  
**Versión:** 2.0-Go  
**Autor:** Daniel-EXLTK  
**Proyecto:** EXLTK-RPC con LLM Central en Go  

## 🎯 Ventajas de Go

### **Por qué Go es perfecto para esta arquitectura:**
- **Concurrencia nativa**: Goroutines para servicios paralelos
- **JSON-RPC integrado**: `net/rpc/jsonrpc` sin dependencias
- **Alto rendimiento**: 10-50x más rápido que Python
- **Deployment simple**: Binarios únicos compilados
- **Networking robusto**: `net/http` para servicios distribuidos

## 🏗️ Arquitectura

```
Usuario → API Gateway (Go) → LLM Orchestrator (Go) → Servicios JSON-RPC (Go)
  ↓           ↓                    ↓                        ↓
HTTP       :8500                :8500                 :8501, :8503
```

## 🔧 Implementación

### **1. Estructura del Proyecto**

```
exltk-rpc/
├── cmd/
│   ├── r0d0-service/main.go
│   ├── proposal-service/main.go
│   └── gateway/main.go
├── internal/
│   ├── services/
│   │   ├── r0d0/
│   │   │   ├── service.go
│   │   │   └── types.go
│   │   └── proposal/
│   │       ├── service.go
│   │       └── types.go
│   └── orchestrator/
│       ├── llm.go
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
    Description string            `json:"description"`
    Params      map[string]string `json:"params"`
    Returns     map[string]string `json:"returns"`
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

type Progress struct {
    Current int `json:"current"`
    Total   int `json:"total"`
}

type ProjectSlot struct {
    Name           string `json:"name"`
    Objective      string `json:"objective"`
    Budget         string `json:"budget"`
    Timeline       string `json:"timeline"`
    Technologies   string `json:"technologies"`
    Audience       string `json:"audience"`
    Risks          string `json:"risks"`
    Resources      string `json:"resources"`
    SuccessMetrics string `json:"success_metrics"`
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

type Session struct {
    UserID    string         `json:"user_id"`
    Questions []string       `json:"questions"`
    Answers   map[int]string `json:"answers"`
    Current   int            `json:"current"`
    CreatedAt time.Time      `json:"created_at"`
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
            "SurveyStart": {
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
            "SurveyContinue": {
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
            "SurveyComplete": {
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

func generateSessionID() string {
    bytes := make([]byte, 8)
    rand.Read(bytes)
    return fmt.Sprintf("survey_%d_%s", time.Now().Unix(), hex.EncodeToString(bytes))
}
```

### **3. LLM Orchestrator**

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
}

type Content struct {
    Parts []Part `json:"parts"`
}

type Part struct {
    Text string `json:"text"`
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
    }
    
    jsonData, _ := json.Marshal(req)
    
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
    
    body, _ := io.ReadAll(resp.Body)
    
    var genResp GenerateResponse
    json.Unmarshal(body, &genResp)
    
    if len(genResp.Candidates) == 0 {
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
    Response string       `json:"response"`
    Plan     WorkflowPlan `json:"plan"`
    Results  []StepResult `json:"results"`
}

type StepResult struct {
    Step   int                    `json:"step"`
    Result map[string]interface{} `json:"result"`
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

func (o *LLMOrchestrator) ProcessMessage(userMessage, userID string) (*ProcessResult, error) {
    // 1. Introspección
    capabilities, err := o.introspectAgents()
    if err != nil {
        return nil, err
    }
    
    // 2. Planificación
    plan, err := o.planWorkflow(userMessage, capabilities)
    if err != nil {
        return nil, err
    }
    
    // 3. Ejecución
    results, err := o.executeWorkflow(plan, userID)
    if err != nil {
        return nil, err
    }
    
    // 4. Respuesta final
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

func (o *LLMOrchestrator) introspectAgents() (map[string]interface{}, error) {
    capabilities := make(map[string]interface{})
    
    // Introspección paralela con goroutines
    var wg sync.WaitGroup
    var mu sync.Mutex
    
    for name, url := range o.agents {
        wg.Add(1)
        go func(agentName, agentURL string) {
            defer wg.Done()
            
            reqBody := map[string]interface{}{
                "jsonrpc": "2.0",
                "method":  "Describe",
                "params":  map[string]interface{}{},
                "id":      1,
            }
            
            jsonData, _ := json.Marshal(reqBody)
            resp, err := http.Post(agentURL+"/jsonrpc", "application/json", bytes.NewBuffer(jsonData))
            if err != nil {
                return
            }
            defer resp.Body.Close()
            
            body, _ := io.ReadAll(resp.Body)
            var rpcResp struct {
                Result interface{} `json:"result"`
            }
            json.Unmarshal(body, &rpcResp)
            
            mu.Lock()
            capabilities[agentName] = rpcResp.Result
            mu.Unlock()
        }(name, url)
    }
    
    wg.Wait()
    return capabilities, nil
}

func (o *LLMOrchestrator) planWorkflow(userMessage string, capabilities map[string]interface{}) (*WorkflowPlan, error) {
    capJSON, _ := json.MarshalIndent(capabilities, "", "  ")
    
    prompt := fmt.Sprintf(`
USUARIO: "%s"

SERVICIOS DISPONIBLES:
%s

Crea un plan JSON:
{
    "analysis": "qué quiere el usuario", 
    "workflow": [
        {
            "step": 1,
            "service": "r0d0",
            "method": "SurveyStart",
            "params": {"user_id": "user123", "message": "mensaje"},
            "description": "iniciar encuesta"
        }
    ]
}`, userMessage, string(capJSON))
    
    response, err := o.client.Generate(prompt)
    if err != nil {
        return nil, err
    }
    
    var plan WorkflowPlan
    json.Unmarshal([]byte(response), &plan)
    
    return &plan, nil
}

func (o *LLMOrchestrator) executeWorkflow(plan *WorkflowPlan, userID string) ([]StepResult, error) {
    results := make([]StepResult, 0)
    context := map[string]interface{}{
        "user_id": userID,
    }
    
    for _, step := range plan.Workflow {
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
        
        // Ejecutar JSON-RPC
        reqBody := map[string]interface{}{
            "jsonrpc": "2.0",
            "method":  step.Method,
            "params":  params,
            "id":      step.Step,
        }
        
        jsonData, _ := json.Marshal(reqBody)
        resp, err := http.Post(agentURL+"/jsonrpc", "application/json", bytes.NewBuffer(jsonData))
        if err != nil {
            return nil, err
        }
        defer resp.Body.Close()
        
        body, _ := io.ReadAll(resp.Body)
        var rpcResp struct {
            Result map[string]interface{} `json:"result"`
        }
        json.Unmarshal(body, &rpcResp)
        
        result := StepResult{
            Step:   step.Step,
            Result: rpcResp.Result,
        }
        results = append(results, result)
        
        // Actualizar contexto
        for key, value := range rpcResp.Result {
            context[key] = value
        }
    }
    
    return results, nil
}

func (o *LLMOrchestrator) generateResponse(userMessage string, results []StepResult) (string, error) {
    resultsJSON, _ := json.MarshalIndent(results, "", "  ")
    
    prompt := fmt.Sprintf(`
Usuario: "%s"

RESULTADOS:
%s

Genera una respuesta natural en español explicando qué se hizo y los resultados.`, userMessage, string(resultsJSON))
    
    return o.client.Generate(prompt)
}
```

### **4. API Gateway**

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

func NewGateway() *Gateway {
    return &Gateway{
        orchestrator: orchestrator.NewLLMOrchestrator(),
    }
}

func (g *Gateway) handleChat(w http.ResponseWriter, r *http.Request) {
    // CORS
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
    
    // Procesar con LLM Orchestrator
    result, err := g.orchestrator.ProcessMessage(req.Message, userID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(result)
}

func main() {
    gateway := NewGateway()
    
    http.HandleFunc("/chat", gateway.handleChat)
    
    log.Println("🚀 API Gateway iniciado en puerto 8500")
    log.Println("📍 Endpoint: POST /chat")
    log.Fatal(http.ListenAndServe(":8500", nil))
}
```

### **5. Servicio R0D0 Main**

#### **cmd/r0d0-service/main.go**
```go
package main

import (
    "io"
    "log"
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
    
    // Handler HTTP para JSON-RPC
    http.HandleFunc("/jsonrpc", func(w http.ResponseWriter, r *http.Request) {
        // CORS
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
    log.Fatal(http.ListenAndServe(":8501", nil))
}

// HTTPConn adapta HTTP para JSON-RPC
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

### **6. Docker**

#### **docker-compose.yml**
```yaml
version: '3.8'
services:
  r0d0-service:
    build:
      context: .
      dockerfile: Dockerfile
      args:
        SERVICE_NAME: r0d0-service
    ports:
      - "8501:8501"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8501/jsonrpc"]
      interval: 30s
      timeout: 10s
      retries: 3

  proposal-service:
    build:
      context: .
      dockerfile: Dockerfile
      args:
        SERVICE_NAME: proposal-service
    ports:
      - "8503:8503"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8503/jsonrpc"]
      interval: 30s
      timeout: 10s
      retries: 3

  gateway:
    build:
      context: .
      dockerfile: Dockerfile
      args:
        SERVICE_NAME: gateway
    ports:
      - "8500:8500"
    environment:
      - GEMINI_API_KEY=${GEMINI_API_KEY}
    depends_on:
      - r0d0-service
      - proposal-service
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8500/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

#### **Dockerfile**
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG SERVICE_NAME
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/${SERVICE_NAME}

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8500
CMD ["./main"]
```

## 🚀 Ventajas de Go

### **1. Rendimiento**
- **10-50x más rápido** que Python
- **Menor uso de memoria**
- **Compilado** a código nativo

### **2. Concurrencia**
- **Goroutines** para introspección paralela
- **Channels** para comunicación segura
- **Escalabilidad** con miles de conexiones

### **3. Simplicidad**
- **Sin dependencias externas**
- **Binario único** para deployment
- **Cross-platform** (Linux, Windows, macOS)

### **4. JSON-RPC Nativo**
- **net/rpc/jsonrpc** integrado
- **Serialización automática**
- **Tipos seguros**

## 🎯 Ejemplo de Uso

```
Usuario: "Necesito una propuesta para mi e-commerce"

1. Gateway recibe HTTP POST /chat
2. LLM Orchestrator:
   - Introspecciona servicios (paralelo con goroutines)
   - Planifica con Gemini
   - Ejecuta: r0d0.SurveyStart → SurveyContinue → SurveyComplete → proposal.Generate
3. Respuesta: "He creado tu propuesta de e-commerce..."
```

## 🚀 Plan de Implementación

### **Fase 1: Servicios Base (1 semana)**
- Implementar r0d0-service en Go
- Implementar proposal-service en Go
- JSON-RPC con método Describe

### **Fase 2: LLM Orchestrator (1 semana)**
- Cliente Gemini
- Introspección paralela
- Planificación y ejecución

### **Fase 3: Gateway (1 semana)**
- API Gateway HTTP
- Endpoints /chat
- CORS y manejo de errores

### **Fase 4: Docker (1 semana)**
- Dockerfiles
- docker-compose.yml
- Tests de integración

## 🎉 Beneficios

1. **Rendimiento**: 10-50x más rápido
2. **Simplicidad**: Sin dependencias externas
3. **Deployment**: Binarios únicos
4. **Escalabilidad**: Goroutines nativas
5. **Mantenibilidad**: Tipado fuerte

¿Empezamos con la implementación en Go? 