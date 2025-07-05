import { useState, useCallback } from 'react'

// Tipos oficiales del ADK API Server
interface ADKMessage {
  role: 'user' | 'model'
  parts: Array<{ text: string }>
}

interface ADKRunRequest {
  appName: string
  userId: string
  sessionId: string
  newMessage: ADKMessage
  streaming?: boolean
}

interface ADKRunResponse {
  content: {
    parts: Array<{ text: string }>
    role: string
  }
  usageMetadata?: {
    candidatesTokenCount: number
    promptTokenCount: number
    totalTokenCount: number
  }
  invocationId: string
  author: string
  actions: {
    stateDelta: object
    artifactDelta: object
    requestedAuthConfigs: object
  }
  id: string
  timestamp: number
}

interface ADKSession {
  id: string
  appName: string
  userId: string
  state: object
  events: Array<any>
  lastUpdateTime: number
}

interface ADKApp {
  name: string
}

export interface UseADKClientResult {
  sendMessage: (message: string) => Promise<string>
  isLoading: boolean
  error: string | null
  sessionId: string | null
  isConnected: boolean
  apps: string[]
  createSession: (appName: string, userId: string) => Promise<string>
  listApps: () => Promise<string[]>
}

const ADK_API_BASE = process.env.NEXT_PUBLIC_ADK_API_URL || 'http://localhost:8000/adk'
const DEFAULT_APP_NAME = process.env.NEXT_PUBLIC_APP_NAME || 'rodo_client'
const DEFAULT_USER_ID = process.env.NEXT_PUBLIC_USER_ID || 'frontend-user-local'

// Debug: verificar qué valores están tomando las variables
console.log('🔧 DEBUG - Variables de entorno:')
console.log('ADK_API_BASE:', ADK_API_BASE)
console.log('DEFAULT_APP_NAME:', DEFAULT_APP_NAME)
console.log('DEFAULT_USER_ID:', DEFAULT_USER_ID)
console.log('NEXT_PUBLIC_ADK_API_URL:', process.env.NEXT_PUBLIC_ADK_API_URL)

export function useADKClient(): UseADKClientResult {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [sessionId, setSessionId] = useState<string | null>(null)
  const [isConnected, setIsConnected] = useState(false)
  const [apps, setApps] = useState<string[]>([])

  const listApps = useCallback(async (): Promise<string[]> => {
    try {
      setError(null)
      const response = await fetch(`${ADK_API_BASE}/list-apps`)
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }
      const appList = await response.json() as string[]
      setApps(appList)
      setIsConnected(true)
      return appList
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Error al listar aplicaciones'
      setError(errorMsg)
      setIsConnected(false)
      return []
    }
  }, [])

  const createSession = useCallback(async (appName: string, userId: string): Promise<string> => {
    try {
      setError(null)
      setIsLoading(true)
      
      const response = await fetch(`${ADK_API_BASE}/apps/${appName}/users/${userId}/sessions`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({})
      })

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }

      const session = await response.json() as ADKSession
      setSessionId(session.id)
      setIsConnected(true)
      return session.id
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Error al crear sesión'
      setError(errorMsg)
      setIsConnected(false)
      throw new Error(errorMsg)
    } finally {
      setIsLoading(false)
    }
  }, [])

  const sendMessage = useCallback(async (message: string): Promise<string> => {
    if (!sessionId) {
      throw new Error('No hay sesión activa. Crear sesión primero.')
    }

    try {
      setError(null)
      setIsLoading(true)

      const requestBody: ADKRunRequest = {
        appName: DEFAULT_APP_NAME,
        userId: DEFAULT_USER_ID,
        sessionId: sessionId,
        newMessage: {
          role: 'user',
          parts: [{ text: message }]
        }
      }

      const response = await fetch(`${ADK_API_BASE}/run`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestBody)
      })

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }

      const responseData = await response.json() as ADKRunResponse[]
      
      if (!responseData || responseData.length === 0) {
        throw new Error('Respuesta vacía del servidor')
      }

      // Extraer el texto de la respuesta del agente
      const lastResponse = responseData[responseData.length - 1]
      const responseText = lastResponse.content.parts
        .map(part => part.text)
        .join('')

      return responseText
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Error al enviar mensaje'
      setError(errorMsg)
      throw new Error(errorMsg)
    } finally {
      setIsLoading(false)
    }
  }, [sessionId])

  return {
    sendMessage,
    isLoading,
    error,
    sessionId,
    isConnected,
    apps,
    createSession,
    listApps
  }
} 