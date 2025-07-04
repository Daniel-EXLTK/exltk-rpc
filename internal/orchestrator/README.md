# LLM Orchestrator

Este paquete contiene la implementación del orquestador central que utiliza un LLM (Gemini) para coordinar dinámicamente los servicios JSON-RPC.

## Funcionalidades

### Introspección
- **IntrospectAgents()** - Descubre capacidades de servicios automáticamente
- **Paralelismo** - Introspección concurrente usando goroutines
- **Cache** - Almacenamiento temporal de capabilities con TTL

### Planificación
- **planWorkflow()** - Genera plan de ejecución usando Gemini
- **Prompt Engineering** - Optimizado para comprensión de capacidades
- **Parsing JSON** - Extracción robusta de planes estructurados

### Ejecución
- **executeWorkflow()** - Ejecuta pasos del workflow secuencialmente
- **Contexto** - Propagación de variables entre pasos
- **Error Handling** - Manejo de errores y rollback

## Arquitectura

```
internal/orchestrator/
├── types.go             # Tipos WorkflowPlan, WorkflowStep, etc.
├── introspection.go     # Descubrimiento de servicios
├── planning.go          # Planificación con LLM
├── execution.go         # Ejecución de workflows
├── orchestrator.go      # Orquestador principal
└── orchestrator_test.go # Tests de integración
```

## Flujo de Operación

1. **Usuario** envía mensaje natural al Gateway
2. **Gateway** llama al Orchestrator
3. **Orchestrator** hace introspección de servicios
4. **Gemini** genera plan de workflow
5. **Orchestrator** ejecuta pasos secuencialmente
6. **Respuesta** estructurada al usuario

## Configuración

- **Gemini API Key**: Variable de entorno `GEMINI_API_KEY`
- **Service URLs**: Configuración de endpoints de servicios
- **Timeouts**: Configurables por operación
- **Logging**: Estructurado para debugging 