# Orquestador LLM - Servidor JSON-RPC

Este servidor expone el orquestador LLM como un endpoint JSON-RPC, permitiendo la integración con cualquier cliente que soporte el protocolo JSON-RPC 2.0.

## Configuración

### Variables de Entorno

```bash
# API Key de Google Gemini (requerida)
export GOOGLE_API_KEY="tu-api-key-de-gemini"

# URLs de los servicios JSON-RPC (requerida)
export SERVICE_URLS="http://localhost:8501,http://localhost:8503"

# Puerto del servidor (opcional, por defecto: 8502)
export PORT="8502"

# Configuración del LLM (opcional)
export LLM_PROVIDER="gemini"
export LLM_MODEL="gemini-2.0-flash-exp"

# Configuración de logging (opcional)
export ENABLE_LOGGING="true"
```

### Ejemplo de archivo .env

```bash
# .env
GOOGLE_API_KEY=AIzaSyD2PFVVO91YR3o7iuV9zV8jjAMidNOcapA
SERVICE_URLS=http://localhost:8501
PORT=8502
LLM_PROVIDER=gemini
LLM_MODEL=gemini-2.0-flash-exp
ENABLE_LOGGING=true
```

## Compilación y Ejecución

```bash
# Compilar
go build ./cmd/orchestrator-server/

# Ejecutar
./orchestrator-server
```

## Métodos JSON-RPC Disponibles

### 1. Health Check

Verifica el estado del orquestador y los servicios disponibles.

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": "health-1",
  "method": "health",
  "params": {}
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": "health-1",
  "result": {
    "status": "healthy",
    "timestamp": "2024-01-15T10:30:00Z",
    "version": "1.0.0",
    "services_available": 1,
    "services_failed": 0,
    "last_introspection": "2024-01-15T10:29:55Z"
  }
}
```

### 2. Métricas

Obtiene métricas del orquestador.

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": "metrics-1",
  "method": "metrics",
  "params": {}
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": "metrics-1",
  "result": {
    "total_services": 1,
    "available_services": 1,
    "total_methods": 4
  }
}
```

### 3. Orquestar Solicitud

Procesa una solicitud de usuario usando el LLM para generar y ejecutar un workflow.

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": "orchestrate-1",
  "method": "orchestrate",
  "params": {
    "user_id": "user123",
    "message": "Quiero crear una aplicación móvil para gestionar tareas",
    "context": {
      "session_type": "discovery",
      "priority": "high"
    }
  }
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": "orchestrate-1",
  "result": {
    "request_id": "req-123",
    "workflow_id": "wf-456",
    "response": "¡Perfecto! He completado tu solicitud ejecutando 2 pasos del workflow. El resultado está listo.",
    "result": {
      "session_id": "session-789",
      "response": "¡Excelente! Me encanta ayudarte a desarrollar tu proyecto...",
      "progress": 25,
      "insights": ["Project type: mobile app", "Functionality: Task management"]
    },
    "steps": [
      {
        "id": "step_1",
        "name": "Iniciar sesión de descubrimiento",
        "service": "r0d0-service",
        "method": "DiscoveryStart",
        "status": "completed",
        "result": { ... }
      }
    ],
    "context": { ... },
    "success": true,
    "error": "",
    "duration": "2.5s",
    "timestamp": "2024-01-15T10:30:05Z"
  }
}
```

### 4. Refrescar Capacidades

Fuerza una actualización de las capacidades de los servicios.

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": "refresh-1",
  "method": "refresh_capabilities",
  "params": {}
}
```

### 5. Obtener Capacidades

Obtiene las capacidades actuales de los servicios (desde cache).

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": "capabilities-1",
  "method": "get_capabilities",
  "params": {}
}
```

## Ejemplos de Uso con curl

### Health Check
```bash
curl -X POST http://localhost:8502/ \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": "health-1",
    "method": "health",
    "params": {}
  }'
```

### Orquestar Solicitud
```bash
curl -X POST http://localhost:8502/ \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": "orchestrate-1",
    "method": "orchestrate",
    "params": {
      "user_id": "test-user",
      "message": "Necesito ayuda para crear una aplicación web de comercio electrónico",
      "context": {
        "project_type": "ecommerce",
        "budget": "medium"
      }
    }
  }'
```

### Métricas
```bash
curl -X POST http://localhost:8502/ \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": "metrics-1",
    "method": "metrics",
    "params": {}
  }'
```

## Códigos de Error JSON-RPC

- `-32700`: Parse error (JSON malformado)
- `-32600`: Invalid Request (versión JSON-RPC incorrecta)
- `-32601`: Method not found (método no existe)
- `-32602`: Invalid params (parámetros inválidos)
- `-32603`: Internal error (error interno del servidor)

## Integración con la UI

Para integrar con la UI de Next.js, puedes crear un cliente JSON-RPC:

```typescript
// lib/orchestrator-client.ts
export class OrchestratorClient {
  private baseURL: string;

  constructor(baseURL: string = 'http://localhost:8502') {
    this.baseURL = baseURL;
  }

  async call(method: string, params: any = {}) {
    const response = await fetch(this.baseURL, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        jsonrpc: '2.0',
        id: crypto.randomUUID(),
        method,
        params,
      }),
    });

    const result = await response.json();
    
    if (result.error) {
      throw new Error(result.error.message);
    }
    
    return result.result;
  }

  async orchestrate(userId: string, message: string, context?: any) {
    return this.call('orchestrate', { user_id: userId, message, context });
  }

  async health() {
    return this.call('health');
  }

  async metrics() {
    return this.call('metrics');
  }
}
```

## Logs y Debugging

El servidor registra información detallada cuando `ENABLE_LOGGING=true`:

```
2024/01/15 10:30:00 Starting orchestrator server on port 8502
2024/01/15 10:30:00 Service URLs: [http://localhost:8501]
2024/01/15 10:30:00 LLM Provider: gemini, Model: gemini-2.0-flash-exp
2024/01/15 10:30:01 Introspection completed: 1 services found, 0 failed, duration: 500ms
2024/01/15 10:30:05 Processing request: req-123 from user: user123
2024/01/15 10:30:05 Planning workflow for request: req-123
2024/01/15 10:30:07 Workflow planning completed in 2.1s
2024/01/15 10:30:07 Executing workflow: wf-456
2024/01/15 10:30:08 Step step_1 completed successfully
2024/01/15 10:30:08 Workflow wf-456 completed successfully in 1.2s
2024/01/15 10:30:08 Request req-123 processed successfully in 3.3s
```

## Arquitectura

```
┌─────────────────┐    JSON-RPC    ┌──────────────────┐
│   Cliente UI    │ ──────────────► │  Orquestador     │
│   (Next.js)     │                │  Server          │
└─────────────────┘                └──────────────────┘
                                           │
                                           ▼
                                   ┌──────────────────┐
                                   │   Orquestador    │
                                   │   (Core)         │
                                   └──────────────────┘
                                           │
                    ┌──────────────────────┼──────────────────────┐
                    ▼                      ▼                      ▼
            ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
            │Introspection │    │  Planning    │    │  Execution   │
            │   (Cache)    │    │  (Gemini)    │    │  (JSON-RPC)  │
            └──────────────┘    └──────────────┘    └──────────────┘
                    │                      │                      │
                    ▼                      ▼                      ▼
            ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
            │   Services   │    │   Gemini     │    │   Services   │
            │  (R0D0, etc) │    │   2.0 Flash  │    │  (R0D0, etc) │
            └──────────────┘    └──────────────┘    └──────────────┘
``` 