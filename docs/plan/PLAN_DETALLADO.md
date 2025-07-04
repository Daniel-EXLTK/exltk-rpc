# 📋 Plan de Implementación Detallado

**Fecha:** 2025-01-27  
**Versión:** 1.0  
**Autor:** Daniel-EXLTK  
**Proyecto:** EXLTK-RPC Multi-Agente en Go  
**Duración:** 4 semanas  

## 🎯 Resumen Ejecutivo

### **Objetivo**
Implementar arquitectura LLM-Orchestrator en Go donde Gemini orquesta servicios JSON-RPC puros.

### **Componentes**
1. **r0d0-service** (Go): Encuestas de proyectos
2. **proposal-service** (Go): Generación de propuestas  
3. **LLM Orchestrator** (Go): Cerebro central
4. **API Gateway** (Go): Interfaz HTTP

### **Beneficios**
- **Rendimiento**: 10-50x más rápido que Python
- **Concurrencia**: Goroutines nativas
- **Simplicidad**: Sin dependencias complejas
- **Deployment**: Binarios únicos

## 🗓️ Cronograma

| **Semana** | **Fase** | **Entregables** | **% Progreso** |
|------------|----------|-----------------|----------------|
| 1 | Infraestructura + R0D0 | Estructura + R0D0 funcional | 25% |
| 2 | Proposal + Gemini | Proposal Service + Cliente Gemini | 50% |
| 3 | Orchestrator + Gateway | Orchestrator + API Gateway | 75% |
| 4 | Docker + Testing | Sistema completo | 100% |

## 📊 FASE 1: Infraestructura y R0D0 Service
**Duración:** Semana 1

### **Día 1: Estructura del Proyecto**

#### **Objetivos**
- Crear estructura Go estándar
- Configurar go.mod
- Establecer convenciones

#### **Entregables**
```
exltk-rpc/
├── cmd/
│   ├── r0d0-service/
│   ├── proposal-service/
│   └── gateway/
├── internal/
│   ├── services/
│   └── orchestrator/
├── pkg/
│   └── gemini/
├── go.mod
└── README.md
```

#### **Criterios de Aceptación**
- [ ] Estructura de carpetas Go estándar
- [ ] go.mod configurado
- [ ] Compilación exitosa: `go build ./...`

### **Día 2: Tipos R0D0**

#### **Objetivos**
- Definir tipos Go para R0D0
- Estructuras requests/responses
- Tipos ProjectSlot

#### **Archivos**
```go
// internal/services/r0d0/types.go
type SurveyStartRequest struct {
    UserID  string `json:"user_id"`
    Message string `json:"message"`
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

#### **Criterios de Aceptación**
- [ ] Tipos completos definidos
- [ ] JSON tags correctos
- [ ] Documentación godoc

### **Día 3: Lógica R0D0**

#### **Objetivos**
- Implementar lógica de negocio
- Manejo de sesiones thread-safe
- Métodos de encuesta

#### **Funcionalidades**
```go
func (s *R0D0Service) Describe(req interface{}, resp *Service) error
func (s *R0D0Service) SurveyStart(req *SurveyStartRequest, resp *SurveyStartResponse) error
func (s *R0D0Service) SurveyContinue(req *SurveyContinueRequest, resp *SurveyContinueResponse) error
func (s *R0D0Service) SurveyComplete(req *SurveyCompleteRequest, resp *SurveyCompleteResponse) error
```

#### **Criterios de Aceptación**
- [ ] Describe() retorna capabilities
- [ ] SurveyStart() crea sesión
- [ ] SurveyContinue() maneja progreso
- [ ] SurveyComplete() genera ProjectSlot
- [ ] Thread-safe con mutex

### **Día 4: Servidor HTTP R0D0**

#### **Objetivos**
- Servidor HTTP JSON-RPC
- Manejo de CORS
- Puerto 8501

#### **Implementación**
```go
func main() {
    service := r0d0.NewR0D0Service()
    server := rpc.NewServer()
    server.Register(service)
    
    http.HandleFunc("/jsonrpc", handleJSONRPC)
    log.Fatal(http.ListenAndServe(":8501", nil))
}
```

#### **Criterios de Aceptación**
- [ ] Servidor en puerto 8501
- [ ] Endpoint /jsonrpc funcional
- [ ] CORS configurado
- [ ] Respuestas JSON-RPC válidas

### **Día 5: Tests R0D0**

#### **Objetivos**
- Tests unitarios completos
- Tests integración HTTP
- Documentación

#### **Tests**
```go
func TestR0D0ServiceDescribe(t *testing.T)
func TestSurveyFullFlow(t *testing.T)
func TestConcurrentUsers(t *testing.T)
```

#### **Criterios de Aceptación**
- [ ] Cobertura >90%
- [ ] Tests HTTP exitosos
- [ ] Documentación completa

## 📊 FASE 2: Proposal Service + Cliente Gemini
**Duración:** Semana 2

### **Día 6: Tipos Proposal**

#### **Tipos**
```go
type GenerateRequest struct {
    ProjectSlot ProjectSlot `json:"project_slot"`
    Template    string      `json:"template"`
}

type Proposal struct {
    ID          string      `json:"id"`
    Content     string      `json:"content"`
    ProjectSlot ProjectSlot `json:"project_slot"`
    Template    string      `json:"template"`
    CreatedAt   time.Time   `json:"created_at"`
}
```

#### **Criterios de Aceptación**
- [ ] Tipos proposal completos
- [ ] Templates definidos
- [ ] Documentación godoc

### **Día 7: Lógica Proposal**

#### **Funcionalidades**
```go
func (s *ProposalService) Generate(req *GenerateRequest, resp *GenerateResponse) error
func (s *ProposalService) Export(req *ExportRequest, resp *ExportResponse) error
func (s *ProposalService) List(req *ListRequest, resp *ListResponse) error
```

#### **Criterios de Aceptación**
- [ ] Generate() crea propuestas
- [ ] Sistema de templates
- [ ] Export() simula exportación
- [ ] Thread-safe storage

### **Día 8: Servidor Proposal**

#### **Objetivos**
- Servidor HTTP JSON-RPC
- Puerto 8503
- Integración completa

#### **Criterios de Aceptación**
- [ ] Servidor en puerto 8503
- [ ] Métodos JSON-RPC funcionales
- [ ] CORS configurado

### **Día 9: Cliente Gemini**

#### **Implementación**
```go
type Client struct {
    apiKey string
    client *http.Client
}

func (c *Client) Generate(prompt string) (string, error) {
    // HTTP POST a Gemini API
}
```

#### **Criterios de Aceptación**
- [ ] Cliente HTTP Gemini
- [ ] Autenticación API key
- [ ] Manejo de errores
- [ ] Tests con mock

### **Día 10: Tests Proposal**

#### **Objetivos**
- Tests proposal service
- Tests cliente Gemini
- Integración end-to-end

#### **Criterios de Aceptación**
- [ ] Cobertura >90%
- [ ] Tests integración
- [ ] Mock Gemini API

## 📊 FASE 3: LLM Orchestrator + Gateway
**Duración:** Semana 3

### **Día 11: Tipos Orchestrator**

#### **Tipos**
```go
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
```

### **Día 12: Introspección**

#### **Funcionalidades**
```go
func (o *LLMOrchestrator) IntrospectAgents() (map[string]AgentCapabilities, error)
```

#### **Criterios de Aceptación**
- [ ] Introspección paralela goroutines
- [ ] Manejo errores servicios
- [ ] Cache capabilities
- [ ] Timeouts configurables

### **Día 13: Planificación LLM**

#### **Funcionalidades**
```go
func (o *LLMOrchestrator) planWorkflow(userMessage string, capabilities map[string]AgentCapabilities) (*WorkflowPlan, error)
```

#### **Criterios de Aceptación**
- [ ] Prompt engineering optimizado
- [ ] Parsing JSON robusto
- [ ] Validación planes
- [ ] Manejo errores LLM

### **Día 14: Ejecución Workflows**

#### **Funcionalidades**
```go
func (o *LLMOrchestrator) executeWorkflow(plan *WorkflowPlan, userID string) ([]StepResult, error)
```

#### **Criterios de Aceptación**
- [ ] Ejecución secuencial
- [ ] Propagación contexto
- [ ] Manejo variables ($var)
- [ ] Error handling

### **Día 15: API Gateway**

#### **Endpoints**
```go
POST /chat          // Procesar mensaje
GET  /capabilities  // Listar capacidades
GET  /health        // Health check
```

#### **Criterios de Aceptación**
- [ ] Servidor puerto 8500
- [ ] /chat procesa mensajes
- [ ] /capabilities introspección
- [ ] CORS configurado

## 📊 FASE 4: Docker + Testing + Docs
**Duración:** Semana 4

### **Día 16-17: Dockerización**

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
CMD ["./main"]
```

#### **Criterios de Aceptación**
- [ ] Multi-stage build
- [ ] Imágenes <50MB
- [ ] docker-compose funcional
- [ ] Health checks

### **Día 18: Testing Integración**

#### **Tests**
```go
func TestFullWorkflowUserToProposal(t *testing.T)
func TestConcurrentUsers(t *testing.T)
func TestServiceFailureRecovery(t *testing.T)
```

#### **Criterios de Aceptación**
- [ ] Tests end-to-end
- [ ] Performance tests
- [ ] Load testing
- [ ] CI/CD pipeline

### **Día 19-20: Documentación**

#### **Documentos**
```
docs/
├── README.md              
├── INSTALLATION.md        
├── API.md                 
├── ARCHITECTURE.md        
└── TROUBLESHOOTING.md     
```

#### **Criterios de Aceptación**
- [ ] README quick start
- [ ] Documentación API
- [ ] Guías instalación
- [ ] Troubleshooting

## 🎯 Criterios de Éxito

### **Funcionales**
- [ ] Usuario hace preguntas naturales
- [ ] LLM orquesta automáticamente
- [ ] Generación propuestas end-to-end
- [ ] Introspección automática
- [ ] Workflows dinámicos

### **No Funcionales**
- [ ] Respuesta <2 segundos
- [ ] 100+ usuarios concurrentes
- [ ] Uptime >99%
- [ ] Cobertura tests >90%

### **Técnicos**
- [ ] Código Go idiomático
- [ ] Docker optimizado
- [ ] CI/CD funcional
- [ ] Monitoring operativo

## 🚨 Riesgos

| **Riesgo** | **Probabilidad** | **Mitigación** |
|------------|------------------|----------------|
| Problemas Gemini API | Media | Mock client + fallback |
| Complejidad JSON-RPC | Baja | POC temprano |
| Performance | Media | Profiling + benchmarks |
| Integración | Media | Tests continuos |

## 📈 Métricas

### **Desarrollo**
- **LOC**: ~2000 líneas
- **Test Coverage**: >90%
- **Build Time**: <30s
- **Image Size**: <50MB

### **Performance**
- **Latency**: <500ms p95
- **Throughput**: >1000 req/min
- **Memory**: <100MB por servicio
- **CPU**: <50% load normal

## 🎉 Entregables

- [ ] Código completo en Git
- [ ] Docker images funcionales
- [ ] Documentación completa
- [ ] Tests automatizados
- [ ] CI/CD pipeline
- [ ] Monitoring dashboard

---

¿Apruebas este plan antes de comenzar la implementación? 