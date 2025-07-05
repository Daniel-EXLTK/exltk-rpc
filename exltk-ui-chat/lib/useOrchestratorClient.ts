import { useState, useCallback, useRef } from 'react'

export interface OrchestratorResponse {
  request_id: string
  workflow_id: string
  response: string
  result: any
  steps: any[]
  context: any
  success: boolean
  error: string | null
  duration: number
  timestamp: string
}

export function useOrchestratorClient(baseUrl: string = '/api/r0d0') {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const requestIdRef = useRef(1)

  const sendMessage = useCallback(
    async (userId: string, message: string, context?: any): Promise<OrchestratorResponse> => {
      setIsLoading(true)
      setError(null)
      const requestId = requestIdRef.current++
      try {
        const jsonRpcRequest = {
          jsonrpc: '2.0',
          method: 'orchestrate',
          params: {
            user_id: userId,
            message,
            context: context || {}
          },
          id: requestId
        }
        const response = await fetch(baseUrl, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json'
          },
          body: JSON.stringify(jsonRpcRequest)
        })
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`)
        }
        const json = await response.json()
        if (json.error) {
          throw new Error(json.error.message || 'Error en orquestador')
        }
        return json.result as OrchestratorResponse
      } catch (err) {
        const msg = err instanceof Error ? err.message : 'Error desconocido'
        setError(msg)
        throw err
      } finally {
        setIsLoading(false)
      }
    },
    [baseUrl]
  )

  return { sendMessage, isLoading, error }
} 