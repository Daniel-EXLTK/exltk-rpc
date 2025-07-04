# 🚀 EXLTK-RPC: LLM-Orchestrator Multi-Agente

**Arquitectura LLM-Orchestrator en Go donde Gemini orquesta dinámicamente servicios JSON-RPC puros**

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![JSON-RPC](https://img.shields.io/badge/Protocol-JSON--RPC-blue?style=flat)](https://www.jsonrpc.org/)
[![Gemini](https://img.shields.io/badge/LLM-Gemini-4285F4?style=flat&logo=google)](https://ai.google.dev/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)

## 🎯 Descripción

Sistema innovador que utiliza un **LLM central (Gemini)** como cerebro orquestador para coordinar dinámicamente múltiples servicios JSON-RPC. El sistema descubre automáticamente las capacidades de los servicios y ejecuta workflows complejos sin programación manual.

### **Características Principales**
- 🧠 **LLM Central**: Gemini como orquestador inteligente
- 🔄 **Introspección Automática**: Descubrimiento dinámico de servicios
- ⚡ **Alto Rendimiento**: Implementado en Go con goroutines
- 🎯 **Workflows Dinámicos**: Sin programación manual
- 🐳 **Containerizado**: Docker con multi-stage builds

## 🏗️ Arquitectura

```
Usuario → API Gateway (8500) → LLM Orchestrator → Servicios JSON-RPC
                                     ↓
                                  Gemini
                                     ↓
                            [r0d0-service:8501]
                            [proposal-service:8503]
```

### **Componentes**
1. **r0d0-service** - Encuestas de proyectos y generación de ProjectSlots
2. **proposal-service** - Generación de propuestas con templates
3. **LLM Orchestrator** - Cerebro central con introspección y planificación
4. **API Gateway** - Interfaz HTTP principal con endpoints REST

## 📚 Documentación

### **🗂️ Estructura**
```
docs/
├── README.md                               # Índice general
├── arquitectura/                           # Diseño y decisiones técnicas
│   ├── ARQUITECTURA_GO.md                 # Arquitectura técnica en Go
│   └── ARQUITECTURA_LLM_ORCHESTRATOR_GO.md # Diseño LLM-Orchestrator
├── plan/                                  # Planificación e implementación
│   ├── PLAN_DETALLADO.md                  # Plan condensado 4 semanas
│   ├── PLAN_IMPLEMENTACION_DETALLADO.md   # Plan extendido subtareas
│   └── PLAN_IMPLEMENTACION_GO.md          # Plan técnico completo
└── seguimiento/                           # Progreso y métricas
    └── SEGUIMIENTO_PROGRESO.md            # Seguimiento diario
```

### **📖 Guía de Lectura**
1. **Empezar aquí**: [docs/README.md](docs/README.md) - Índice completo
2. **Arquitectura**: [docs/arquitectura/ARQUITECTURA_LLM_ORCHESTRATOR_GO.md](docs/arquitectura/ARQUITECTURA_LLM_ORCHESTRATOR_GO.md)
3. **Plan**: [docs/plan/PLAN_DETALLADO.md](docs/plan/PLAN_DETALLADO.md)
4. **Progreso**: [docs/seguimiento/SEGUIMIENTO_PROGRESO.md](docs/seguimiento/SEGUIMIENTO_PROGRESO.md)

## 🚀 Quick Start

### **1. Revisar Documentación**
```bash
# Leer índice general
cat docs/README.md

# Estudiar arquitectura
cat docs/arquitectura/ARQUITECTURA_LLM_ORCHESTRATOR_GO.md

# Consultar plan
cat docs/plan/PLAN_DETALLADO.md
```

### **2. Próximos Pasos**
1. **Inmediato**: Iniciar Fase 1 - Día 1 (Estructura del Proyecto)
2. **Esta Semana**: Completar R0D0 Service
3. **Próxima Semana**: Proposal Service + Cliente Gemini

## 🎯 Objetivos del Proyecto

### **Funcionales** 
- ✅ Usuario hace preguntas en lenguaje natural
- ✅ LLM orquesta servicios automáticamente  
- ✅ Generación de propuestas end-to-end
- ✅ Introspección automática de servicios
- ✅ Workflows dinámicos sin programación

### **Performance**
- ⚡ Respuesta <2 segundos workflows simples
- 🔥 Soporte 100+ usuarios concurrentes
- 📊 Uptime >99% ambiente pruebas
- 🧪 Cobertura tests >90%

## 🔧 Tecnologías

| **Categoría** | **Tecnología** | **Propósito** |
|---------------|----------------|---------------|
| **Backend** | Go 1.21+ | Lenguaje principal |
| **Protocolo** | JSON-RPC | Comunicación inter-servicios |
| **Concurrencia** | goroutines | Paralelismo nativo |
| **AI/ML** | Gemini API | LLM central |
| **DevOps** | Docker | Containerización |

## 📊 Estado del Proyecto

### **Progreso General**
- 📋 **Planificación**: ✅ 100% Completado
- 🔧 **Desarrollo**: ⏳ 0% - Por comenzar  
- 🧪 **Testing**: ⏳ 0% - Pendiente
- 📚 **Documentación**: ✅ 100% Completado

### **Cronograma**
| **Semana** | **Fase** | **Entregables** |
|------------|----------|-----------------|
| 1 | Infraestructura + R0D0 | ⏳ Estructura + R0D0 funcional |
| 2 | Proposal + Gemini | ⏳ Proposal Service + Cliente Gemini |
| 3 | Orchestrator + Gateway | ⏳ Orchestrator + API Gateway |
| 4 | Docker + Testing | ⏳ Sistema completo |

## 📈 Métricas Objetivo

### **Desarrollo**
- **LOC Target**: ~2000 líneas
- **Test Coverage**: >90%
- **Build Time**: <30 segundos
- **Image Size**: <50MB por servicio

### **Performance**
- **API Latency**: <500ms p95
- **Throughput**: >1000 req/min
- **Memory**: <100MB por servicio
- **CPU**: <50% carga normal

## 🚨 Gestión de Riesgos

| **Riesgo** | **Probabilidad** | **Mitigación** |
|------------|------------------|----------------|
| Gemini API | Media | Mock client + fallback |
| JSON-RPC Go | Baja | POC temprano |
| Performance | Media | Profiling + benchmarks |
| Integración | Media | Tests continuos |

## 📞 Información del Proyecto

**Proyecto:** EXLTK-RPC  
**Autor:** Daniel-EXLTK  
**Fecha Inicio:** 2025-01-27  
**Duración:** 4 semanas  
**Estado:** Planificación completada ✅

---

**Próxima Revisión:** 2025-01-28  
**Objetivo:** Iniciar desarrollo - Fase 1 Día 1

## 🔗 Enlaces Rápidos

- 📚 [Documentación Completa](docs/README.md)
- 🏗️ [Arquitectura LLM-Orchestrator](docs/arquitectura/ARQUITECTURA_LLM_ORCHESTRATOR_GO.md)
- 📋 [Plan de Implementación](docs/plan/PLAN_DETALLADO.md)
- 📊 [Seguimiento de Progreso](docs/seguimiento/SEGUIMIENTO_PROGRESO.md)

---

> **Nota**: Este proyecto implementa una arquitectura innovadora donde un LLM central actúa como orquestador inteligente de servicios JSON-RPC, eliminando la necesidad de programación manual de workflows complejos. 