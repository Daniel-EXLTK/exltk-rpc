import { useState, useCallback, useEffect } from 'react'

// Tipos A2A según documentación oficial y formato real de r0d0
interface AgentCard {
  name: string
  description: string
  capabilities: Array<{
    capability_id: string
    description: string
  }> | {} // Puede ser array o objeto vacío
  grpc_endpoint?: string
  http_endpoint?: string
  url?: string // r0d0 usa 'url' en lugar de 'http_endpoint'
  version: string
  skills?: Array<{
    id: string
    name: string
    description: string
    examples?: string[]
    tags?: string[]
  }> // r0d0 incluye skills
}

interface A2AMessage {
  messageId: string
  role: 'user' | 'agent'
  parts: Array<{ text: string }>
  metadata?: Record<string, any>
}

interface TaskState {
  task_id: string
  state: 'submitted' | 'working' | 'completed' | 'failed' | 'canceled'
  message?: A2AMessage
  artifacts?: any[]
}

interface A2AResponse {
  task_id: string
  status: TaskState
}

export interface UseA2AClientResult {
  sendMessage: (message: string) => Promise<string>
  isLoading: boolean
  error: string | null
  isConnected: boolean
  agentCard: AgentCard | null
  taskId: string | null
  connectToAgent: (agentUrl: string) => Promise<boolean>
}

// Variables de entorno A2A
const R0D0_AGENT_URL = process.env.NEXT_PUBLIC_A2A_AGENT_URL || 'http://localhost:8501'
const TARGET_AGENT_NAME = process.env.NEXT_PUBLIC_TARGET_AGENT || 'r0d0_agent'

// Debug
console.log('🔧 A2A Client Config:')
console.log('R0D0_AGENT_URL:', R0D0_AGENT_URL)
console.log('TARGET_AGENT_NAME:', TARGET_AGENT_NAME)

export function useA2AClient(): UseA2AClientResult {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [isConnected, setIsConnected] = useState(false)
  const [agentCard, setAgentCard] = useState<AgentCard | null>(null)
  const [taskId, setTaskId] = useState<string | null>(null)
  
  // CORREGIDO: SessionID persistente para mantener contexto de conversación
  const [sessionId, setSessionId] = useState<string | null>(null)

  // CORREGIDO: Generar SessionID único para toda la conversación (movido aquí)
  const generateSessionId = useCallback((): string => {
    const timestamp = Date.now()
    const random = Math.random().toString(36).substring(2, 8)
    return `ui_session_${timestamp}_${random}`
  }, [])

  // Función para descubrir Agent Card usando proxy (evita problemas CORS)
  const discoverAgent = useCallback(async (agentUrl: string): Promise<AgentCard | null> => {
    try {
      setError(null)
      console.log(`🔍 Descubriendo Agent Card: ${agentUrl}`)
      
      // Usar nuestro endpoint proxy en lugar de conectar directamente
      const response = await fetch(`/api/agent-card?url=${encodeURIComponent(agentUrl)}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        }
      })

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}))
        throw new Error(errorData.error || `HTTP ${response.status}: ${response.statusText}`)
      }

      const card = await response.json() as AgentCard
      console.log('✅ Agent Card descubierto:', card.name)
      return card
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Error al descubrir agente'
      console.error('❌ Error descubriendo Agent Card:', errorMsg)
      setError(errorMsg)
      return null
    }
  }, [])

  // Conectar al agente A2A - OPTIMIZADO
  const connectToAgent = useCallback(async (agentUrl: string): Promise<boolean> => {
    try {
      setError(null)
      setIsLoading(true)

      console.log('🔄 Conectando a:', agentUrl)
      const card = await discoverAgent(agentUrl)
      
      if (!card) {
        throw new Error('No se pudo obtener Agent Card')
      }

      setAgentCard(card)
      setIsConnected(true)
      
      // CORREGIDO: Generar sessionId persistente una sola vez al conectar
      const newSessionId = generateSessionId()
      setSessionId(newSessionId)
      
      console.log('✅ Conectado al agente:', card.name)
      return true

    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Error al conectar con agente'
      console.error('❌ Error en connectToAgent:', errorMsg)
      setError(errorMsg)
      setIsConnected(false)
      setAgentCard(null)
      return false
    } finally {
      setIsLoading(false)
    }
  }, [discoverAgent, generateSessionId])

  // Generar IDs únicos para A2A
  const generateMessageId = useCallback((): string => {
    const timestamp = Date.now()
    const random = Math.random().toString(36).substring(2, 8)
    return `ui_msg_${timestamp}_${random}`
  }, [])

  const generateTaskId = useCallback((): string => {
    const timestamp = Date.now()
    const random = Math.random().toString(36).substring(2, 8)
    return `ui_task_${timestamp}_${random}`
  }, [])

  // Enviar mensaje usando protocolo A2A oficial
  const sendMessage = useCallback(async (message: string): Promise<string> => {
    if (!isConnected || !agentCard || !sessionId) {
      throw new Error('No conectado al agente A2A o SessionID no inicializado. Llamar connectToAgent() primero.')
    }

    try {
      setError(null)
      setIsLoading(true)

      const messageId = generateMessageId()
      
      // CORREGIDO: Usar sessionId persistente, NO generar taskId nuevo cada vez
      // Solo generar taskId si no existe (primer mensaje)
      let currentTaskId = taskId
      if (!currentTaskId) {
        currentTaskId = generateTaskId()
        setTaskId(currentTaskId)
      }

      // OPTIMIZADO: Logging mínimo para evitar saturación
      console.log(`📤 Enviando mensaje A2A - TaskID: ${currentTaskId}`)

      // Crear mensaje A2A según documentación oficial
      const a2aMessage: A2AMessage = {
        messageId: messageId,
        role: 'user',
        parts: [{ text: message }],
        metadata: {
          source: 'ui',
          timestamp: new Date().toISOString(),
          sessionId: sessionId,  // CORREGIDO: SessionID persistente
          taskId: currentTaskId  // CORREGIDO: TaskID consistente
        }
      }

      // Determinar endpoint según Agent Card (r0d0 usa 'url', otros usan 'http_endpoint')  
      const endpoint = agentCard.http_endpoint || agentCard.url || R0D0_AGENT_URL

      // Formato A2A nativo según A2AStarletteApplication
      const requestBody = a2aMessage

      // Enviar mensaje usando endpoint proxy (evita problemas CORS)
      const response = await fetch('/api/agent-message', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          agentUrl: endpoint,
          message: requestBody
        })
      })

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}))
        throw new Error(errorData.error || `HTTP ${response.status}: ${response.statusText}`)
      }

      const responseData = await response.json()
      console.log('📥 Respuesta A2A recibida')

      // OPTIMIZADO: Verificación mínima sin logging excesivo

      // Procesar respuesta A2A nativa
      let responseText = ''
      
      // OPTIMIZADO: Extraer texto de artifacts.parts sin logging excesivo
      if (responseData.result && responseData.result.artifacts && Array.isArray(responseData.result.artifacts)) {
        for (const artifact of responseData.result.artifacts) {
          if (artifact.parts && Array.isArray(artifact.parts)) {
            for (const part of artifact.parts) {
              // Extraer texto del part
              if (part.text) {
                responseText += part.text
              } else if (part.root && part.root.text) {
                responseText += part.root.text
              }
            }
          }
        }
      }
      // Verificar si es una respuesta directa con parts
      else if (responseData.parts && Array.isArray(responseData.parts)) {
        responseText = responseData.parts
          .map((part: any) => part.text)
          .filter((text: string) => text)
          .join('')
      }
      // Verificar si es un Task con mensaje
      else if (responseData.message && responseData.message.parts) {
        responseText = responseData.message.parts
          .map((part: any) => part.text)
          .filter((text: string) => text)
          .join('')
      }
      // Si tiene artifacts directos (sin result wrapper)
      else if (responseData.artifacts && Array.isArray(responseData.artifacts)) {
        for (const artifact of responseData.artifacts) {
          if (artifact.parts) {
            responseText += artifact.parts
              .map((part: any) => part.text || part.root?.text)
              .filter((text: string) => text)
              .join('')
          }
        }
      }
      // Fallback: buscar cualquier campo de texto
      else if (responseData.text) {
        responseText = responseData.text
      }
      // Último fallback: convertir a string si no hay formato reconocible
      else {
        responseText = typeof responseData === 'string' ? responseData : JSON.stringify(responseData)
      }

      // OPTIMIZADO: Validación mínima sin logging excesivo
      if (!responseText.trim()) {
        console.warn('⚠️ Respuesta A2A vacía')
        responseText = "Lo siento, no pude procesar la respuesta del agente. Por favor, intenta de nuevo."
      }
      return responseText

    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Error al enviar mensaje A2A vía proxy'
      console.error('❌ Error en comunicación A2A vía proxy:', errorMsg)
      setError(errorMsg)
      throw new Error(errorMsg)
    } finally {
      setIsLoading(false)
    }
  }, [isConnected, agentCard, sessionId, taskId, generateMessageId, generateTaskId])

  // Auto-conectar al cargar - CORREGIDO: Sin dependencias problemáticas
  useEffect(() => {
    const initializeA2AConnection = async () => {
      console.log('🚀 Inicializando conexión A2A...')
      
      try {
        const success = await connectToAgent(R0D0_AGENT_URL)
        
        if (!success) {
          console.warn('⚠️ Conexión A2A falló, la UI funcionará en modo simulado')
        } else {
          console.log('✅ Conexión A2A exitosa!')
        }
      } catch (error) {
        console.error('❌ Error en inicialización A2A:', error)
      }
    }

    // Solo ejecutar si no está conectado
    if (!isConnected) {
      initializeA2AConnection()
    }
  }, []) // CORREGIDO: Array vacío para ejecutar solo una vez

  return {
    sendMessage,
    isLoading,
    error,
    isConnected,
    agentCard,
    taskId,
    connectToAgent
  }
} 