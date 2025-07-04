# Internal Services

Este paquete contiene la implementación de los servicios internos del proyecto.

## Servicios

### R0D0 Service
- **Propósito**: Manejo de encuestas de proyectos
- **Funcionalidades**:
  - Describe() - Retorna capacidades del servicio
  - SurveyStart() - Inicia encuesta de proyecto
  - SurveyContinue() - Continúa progreso de encuesta
  - SurveyComplete() - Completa encuesta y genera ProjectSlot

### Proposal Service
- **Propósito**: Generación de propuestas de proyectos
- **Funcionalidades**:
  - Describe() - Retorna capacidades del servicio
  - Generate() - Genera propuesta desde ProjectSlot
  - Export() - Exporta propuesta en diferentes formatos
  - List() - Lista propuestas del usuario

## Arquitectura

```
internal/services/
├── r0d0/
│   ├── types.go          # Tipos y estructuras
│   ├── service.go        # Lógica de negocio
│   └── service_test.go   # Tests unitarios
└── proposal/
    ├── types.go          # Tipos y estructuras
    ├── service.go        # Lógica de negocio
    └── service_test.go   # Tests unitarios
```

## Implementación

Los servicios son implementados como JSON-RPC puros sin lógica de IA interna. La inteligencia está centralizada en el LLM Orchestrator. 