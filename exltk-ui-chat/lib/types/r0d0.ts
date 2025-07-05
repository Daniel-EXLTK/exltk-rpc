// DEPRECATED: Este archivo ha sido reemplazado por los tipos del orquestador. Eliminar si no es necesario.

// Tipos R0D0 eliminados tras migración a orquestador
// Si algún componente requiere compatibilidad, migrar a los tipos del orquestador en useOrchestratorClient

export interface DiscoveryStartRequest {
  user_id: string
  message: string
}

export interface DiscoveryStartResponse {
  session_id: string
  response: string
  next_prompt: string
  progress: number
  insights: string[]
  instructions: string
}

export interface DiscoveryContinueRequest {
  session_id: string
  message: string
}

export interface DiscoveryContinueResponse {
  session_id: string
  response: string
  next_prompt: string
  progress: number
  insights: string[]
  is_complete: boolean
  confidence: number
  missing_areas: string[]
}

export interface DiscoveryCompleteRequest {
  session_id: string
  force?: boolean
}

export interface DiscoveryCompleteResponse {
  session_id: string
  project_slot: ProjectSlot
  message: string
  success: boolean
  confidence: number
  gaps: string[]
}

export interface ProjectSlot {
  id: string
  name: string
  objective: string
  budget: string
  timeline: string
  technologies: string
  audience: string
  risks: string
  resources: string
  success_metrics: string
  created_at: string
  updated_at: string
  user_id: string
  status: string
}

export interface ServiceDescription {
  service: string
  description: string
  version: string
  purpose: string
  methods: { [key: string]: any }
  capabilities: string[]
  discovery_areas: Array<{
    key: string
    name: string
    description: string
    priority: number
    status: string
    prompts: string[]
    keywords: string[]
  }>
}

// Tipos para JSON-RPC
export interface JSONRPCRequest {
  jsonrpc: "2.0"
  method: string
  params: any
  id: string | number
}

export interface JSONRPCResponse<T = any> {
  jsonrpc: "2.0"
  result?: T
  error?: {
    code: number
    message: string
    data?: any
  }
  id: string | number
}

// Tipos para el cliente R0D0
export interface R0D0ClientConfig {
  baseUrl: string
  timeout: number
}

export interface R0D0ClientResult {
  discoverStart: (message: string) => Promise<DiscoveryStartResponse>
  discoverContinue: (sessionId: string, message: string) => Promise<DiscoveryContinueResponse>
  discoverComplete: (sessionId: string) => Promise<DiscoveryCompleteResponse>
  describe: () => Promise<ServiceDescription>
  isLoading: boolean
  error: string | null
  sessionId: string | null
}

// Enums para áreas de discovery
export enum DiscoveryArea {
  OBJETIVO = "objetivo",
  AUDIENCIA = "audiencia",
  PRESUPUESTO = "presupuesto",
  TIMELINE = "timeline",
  TECNOLOGIA = "tecnologia",
  RECURSOS = "recursos",
  RIESGOS = "riesgos",
  METRICAS = "metricas_exito",
  CONTEXTO = "contexto"
}

// Estados de discovery para mapeo con UI
export enum DiscoveryState {
  INITIAL = "initial",
  PROJECT_DETAILS = "project_details",
  REQUIREMENTS = "requirements",
  PROPOSAL = "proposal",
  COMPLETE = "complete"
}

// Mapeo de progress a estados UI
export const progressToState = (progress: number): DiscoveryState => {
  if (progress < 10) return DiscoveryState.INITIAL
  if (progress < 50) return DiscoveryState.PROJECT_DETAILS
  if (progress < 80) return DiscoveryState.REQUIREMENTS
  if (progress < 100) return DiscoveryState.PROPOSAL
  return DiscoveryState.COMPLETE
}

// Mapeo de confianza a mood UI (el servicio R0D0 devuelve 0-100, no 0-1)
export const confidenceToMood = (confidence: number): string => {
  if (confidence < 30) return "worried"
  if (confidence < 60) return "thinking"
  if (confidence < 80) return "smiling"
  return "happy"
} 