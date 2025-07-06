import { useState } from 'react'

export interface OrchestratorResponse {
  response: string
  success: boolean
  result?: any
  workflow_id?: string
  request_id?: string
  steps?: any[]
  context?: any
  error?: string
  duration?: string
  timestamp?: string
}

interface OrchestratorContext {
  session_id?: string
  user_id?: string
  conversation_state?: string
  [key: string]: any
}

export function useOrchestratorClient() {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [conversationContext, setConversationContext] = useState<OrchestratorContext>({})

  const sendMessage = async (
    userId: string,
    message: string,
    additionalContext: Record<string, any> = {}
  ): Promise<OrchestratorResponse> => {
    setIsLoading(true)
    setError(null)

    try {
      console.log('🔍 Enviando mensaje al orquestador')
      console.log('📤 Mensaje:', message)
      console.log('📤 Contexto:', { ...conversationContext, ...additionalContext })

      const response = await fetch('/api/r0d0', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          jsonrpc: '2.0',
          method: 'orchestrate',
          params: {
            user_id: userId,
            message: message,
            context: {
              ...conversationContext,
              ...additionalContext,
              user_id: userId,
            }
          },
          id: Date.now().toString()
        })
      })

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}))
        throw new Error(errorData.error || `HTTP ${response.status}: ${response.statusText}`)
      }

      const data = await response.json()
      console.log('📥 Respuesta del orquestador:', data)

      if (data.error) {
        throw new Error(data.error.message || 'Error del orquestador')
      }

      const orchestratorResp: OrchestratorResponse = data.result

      // Actualizar contexto conversacional con información de la respuesta
      if (orchestratorResp.context) {
        console.log('🔄 Actualizando contexto conversacional:', orchestratorResp.context)
        setConversationContext(prev => ({
          ...prev,
          ...orchestratorResp.context,
          user_id: userId,
        }))
      }

      // Si hay un session_id en los resultados de los pasos, guardarlo
      if (orchestratorResp.steps) {
        for (const step of orchestratorResp.steps) {
          if (step.result && step.result.session_id) {
            console.log('💾 Guardando session_id:', step.result.session_id)
            setConversationContext(prev => ({
              ...prev,
              session_id: step.result.session_id,
              previous_session_id: step.result.session_id,
            }))
            break
          }
        }
      }

      return orchestratorResp

    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Error al comunicar con orquestador'
      console.error('❌ Error en useOrchestratorClient:', errorMsg)
      setError(errorMsg)
      throw new Error(errorMsg)
    } finally {
      setIsLoading(false)
    }
  }

  const clearContext = () => {
    setConversationContext({})
  }

  return {
    sendMessage,
    isLoading,
    error,
    conversationContext,
    clearContext
  }
} 