# Plan de Integración UI ↔ R0D0 (JSON-RPC Directo)

## Resumen
Conectar la UI existente (`exltk-ui-chat`) con el servicio R0D0 usando JSON-RPC directo, sin necesidad de Agent Card.

## Estado Actual
- ✅ Servicio R0D0 funcionando en puerto 8501 con JSON-RPC
- ✅ UI Next.js funcionando en puerto 3000 
- ✅ Tipos R0D0 implementados (Discovery conversacional)
- ✅ 13/13 tests pasando en R0D0

## Arquitectura de Integración

```
UI (Next.js) → JSON-RPC HTTP → R0D0 Service (Go)
     ↓                               ↓
  useR0D0Client                 JSON-RPC Server
     ↓                               ↓
  API Routes                     Discovery Methods
     ↓                               ↓
  React Components               Session Management
```

## Plan de Implementación (2 horas)

### Fase 1: Cliente JSON-RPC (45 min)
**Objetivo**: Crear hook `useR0D0Client` para llamadas directas

1. **Crear `hooks/useR0D0Client.ts`** (25 min)
   - Implementar cliente HTTP para JSON-RPC
   - Métodos: `discoverStart`, `discoverContinue`, `discoverComplete`, `describe`
   - Manejo de errores y tipos TypeScript
   - URL base configurada: `http://localhost:8501`

2. **Crear tipos TypeScript** (20 min)
   - Mapear tipos Go R0D0 a TypeScript
   - Interfaces para requests/responses
   - Enums para estados y áreas

### Fase 2: Adaptación API Routes (30 min)
**Objetivo**: Adaptar endpoints existentes para usar R0D0

1. **Actualizar `api/chat/route.ts`** (15 min)
   - Reemplazar lógica A2A con llamadas R0D0
   - Mapear mensajes UI → JSON-RPC calls
   - Mantener formato response esperado

2. **Crear `api/r0d0/route.ts`** (15 min)
   - Endpoint dedicado para operaciones R0D0
   - Proxy directo a métodos JSON-RPC
   - Manejo de CORS si es necesario

### Fase 3: Integración UI (45 min)
**Objetivo**: Conectar componentes React con R0D0

1. **Actualizar `R0D0Interface.tsx`** (30 min)
   - Reemplazar `useA2AClient` con `useR0D0Client`
   - Mapear estados UI a flujo discovery
   - Integrar confidence y progress en tiempo real

2. **Actualizar gestión de estado** (15 min)
   - Sincronizar `InterviewStep` con discovery progress
   - Mapear `ProjectSlots` con output R0D0
   - Mantener compatibilidad con UI existente

## Mapeo de Datos

### Estados UI → R0D0
```typescript
// UI States
"initial" → DiscoveryStartRequest
"project_details" → DiscoveryContinueRequest  
"complete" → DiscoveryCompleteRequest

// Messages
UI: { content: string, sender: "user" | "assistant" }
R0D0: { Message: string, SessionID: string }
```

### Flujo de Llamadas
```
1. Usuario escribe mensaje → UI
2. UI → useR0D0Client.discoverStart/Continue()
3. Cliente → HTTP POST a localhost:8501 (JSON-RPC)
4. R0D0 → Procesa y responde
5. Cliente → Parsea response
6. UI → Actualiza estado y muestra respuesta
```

### Estructura JSON-RPC
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "R0D0.DiscoveryStart",
  "params": {
    "InitialMessage": "Quiero crear una app móvil"
  },
  "id": "1"
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "SessionID": "sess_123",
    "Response": "¡Perfecto! Cuéntame más sobre tu app móvil...",
    "CurrentArea": "objetivo",
    "Progress": 10,
    "Confidence": 0.3
  },
  "id": "1"
}
```

## Archivos a Modificar

### Nuevos archivos:
- `hooks/useR0D0Client.ts` - Cliente JSON-RPC
- `types/r0d0.ts` - Tipos TypeScript
- `api/r0d0/route.ts` - API endpoint

### Archivos existentes:
- `components/R0D0Interface.tsx` - Integración UI
- `api/chat/route.ts` - Adaptación endpoint
- `hooks/useA2AClient.ts` - Deprecar o mantener como fallback

## Validación

### Tests de Integración:
1. **Conectividad**: UI → R0D0 Service
2. **Flujo Discovery**: Start → Continue → Complete  
3. **Manejo de errores**: Timeouts, conexión perdida
4. **Estados UI**: Sincronización con progress R0D0

### Criterios de Éxito:
- ✅ Usuario puede iniciar discovery desde UI
- ✅ Conversación fluida entre UI y R0D0
- ✅ Progress y confidence se muestran en tiempo real
- ✅ ProjectSlot se genera correctamente
- ✅ Sesiones se manejan apropiadamente

## Próximos Pasos

1. **Inmediato**: Crear `useR0D0Client` hook
2. **Seguimiento**: Adaptar API routes
3. **Final**: Integrar con componentes React
4. **Validación**: Tests end-to-end

---

**Tiempo Total Estimado**: 2 horas  
**Prioridad**: Alta  
**Dependencias**: Servicio R0D0 funcionando (✅ Completado) 