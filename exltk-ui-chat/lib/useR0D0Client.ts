// DEPRECATED: Este hook ha sido reemplazado por useOrchestratorClient. Eliminar este archivo si no es necesario.

import { useState, useCallback, useRef } from 'react'
import type {
  R0D0ClientResult,
  R0D0ClientConfig,
  JSONRPCRequest,
  JSONRPCResponse,
  DiscoveryStartRequest,
  DiscoveryStartResponse,
  DiscoveryContinueRequest,
  DiscoveryContinueResponse,
  DiscoveryCompleteRequest,
  DiscoveryCompleteResponse,
  ServiceDescription
} from './types/r0d0'

// Configuración del cliente R0D0
const DEFAULT_CONFIG: R0D0ClientConfig = {
  baseUrl: '/api/r0d0', // Usar endpoint proxy en lugar de llamada directa
  timeout: 30000 // 30 segundos
}

// Debug
console.log('🔧 R0D0 Client Config:', DEFAULT_CONFIG)

/**
 * Hook para interactuar con el servicio R0D0 via JSON-RPC directo
 * Mucho más simple que el cliente A2A - solo hace llamadas HTTP POST
 */
export function useR0D0Client(config: Partial<R0D0ClientConfig> = {}): R0D0ClientResult {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [sessionId, setSessionId] = useState<string | null>(null)
  
  // Configuración combinada
  const clientConfig = { ...DEFAULT_CONFIG, ...config }
  
  // Contador para IDs de requests JSON-RPC
  const requestIdRef = useRef(1)
  
  // Función auxiliar para hacer llamadas JSON-RPC via proxy
  const makeJSONRPCCall = useCallback(async <T>(
    method: string,
    params: any
  ): Promise<T> => {
    const requestId = requestIdRef.current++
    
    // Formato para el endpoint proxy - parámetros como array para Go net/rpc
    const proxyRequest = {
      method: method,
      params: params ? [params] : [null], // Go net/rpc espera array de parámetros
      id: requestId
    }
    
    console.log(`📤 JSON-RPC Call via Proxy: ${method}`, { requestId, params })
    
    try {
      const response = await fetch(clientConfig.baseUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        },
        body: JSON.stringify(proxyRequest),
        signal: AbortSignal.timeout(clientConfig.timeout)
      })
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }
      
      const jsonResponse: JSONRPCResponse<T> = await response.json()
      
      console.log(`📥 JSON-RPC Response via Proxy: ${method}`, {
        requestId,
        success: !!jsonResponse.result,
        error: jsonResponse.error
      })
      
      // Validar respuesta JSON-RPC
      if (jsonResponse.error) {
        throw new Error(`JSON-RPC Error ${jsonResponse.error.code}: ${jsonResponse.error.message}`)
      }
      
      if (jsonResponse.id !== requestId) {
        throw new Error(`Request ID mismatch: expected ${requestId}, got ${jsonResponse.id}`)
      }
      
      if (!jsonResponse.result) {
        throw new Error('No result in JSON-RPC response')
      }
      
      return jsonResponse.result
      
    } catch (err) {
      if (err instanceof Error) {
        if (err.name === 'AbortError') {
          throw new Error(`Timeout: ${method} took longer than ${clientConfig.timeout}ms`)
        }
        throw err
      }
      throw new Error(`Unknown error calling ${method}`)
    }
  }, [clientConfig])
  
  // Iniciar discovery
  const discoverStart = useCallback(async (message: string): Promise<DiscoveryStartResponse> => {
    setError(null)
    setIsLoading(true)
    
    try {
      const request: DiscoveryStartRequest = {
        user_id: 'ui_user_' + Date.now(), // Generar user_id único para UI
        message: message
      }
      
      const response = await makeJSONRPCCall<DiscoveryStartResponse>(
        'R0D0Service.DiscoveryStart',
        request
      )
      
      // Guardar sessionId para futuras llamadas
      setSessionId(response.session_id)
      
      console.log('✅ Discovery started:', {
        sessionId: response.session_id,
        progress: response.progress,
        insights: response.insights
      })
      
      return response
      
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Error al iniciar discovery'
      console.error('❌ Error in discoverStart:', errorMsg)
      setError(errorMsg)
      throw err
    } finally {
      setIsLoading(false)
    }
  }, [makeJSONRPCCall])
  
  // Continuar discovery
  const discoverContinue = useCallback(async (
    sessionId: string,
    message: string
  ): Promise<DiscoveryContinueResponse> => {
    setError(null)
    setIsLoading(true)
    
    try {
      const request: DiscoveryContinueRequest = {
        session_id: sessionId,
        message: message
      }
      
      const response = await makeJSONRPCCall<DiscoveryContinueResponse>(
        'R0D0Service.DiscoveryContinue',
        request
      )
      
      console.log('✅ Discovery continued:', {
        sessionId: response.session_id,
        progress: response.progress,
        confidence: response.confidence,
        isComplete: response.is_complete
      })
      
      return response
      
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Error al continuar discovery'
      console.error('❌ Error in discoverContinue:', errorMsg)
      setError(errorMsg)
      throw err
    } finally {
      setIsLoading(false)
    }
  }, [makeJSONRPCCall])
  
  // Completar discovery
  const discoverComplete = useCallback(async (sessionId: string): Promise<DiscoveryCompleteResponse> => {
    setError(null)
    setIsLoading(true)
    
    try {
      const request: DiscoveryCompleteRequest = {
        session_id: sessionId
      }
      
      const response = await makeJSONRPCCall<DiscoveryCompleteResponse>(
        'R0D0Service.DiscoveryComplete',
        request
      )
      
      console.log('✅ Discovery completed:', {
        sessionId: response.session_id,
        confidence: response.confidence,
        projectSlot: response.project_slot?.name
      })
      
      return response
      
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Error al completar discovery'
      console.error('❌ Error in discoverComplete:', errorMsg)
      setError(errorMsg)
      throw err
    } finally {
      setIsLoading(false)
    }
  }, [makeJSONRPCCall])
  
  // Obtener descripción del servicio
  const describe = useCallback(async (): Promise<ServiceDescription> => {
    setError(null)
    setIsLoading(true)
    
    try {
      const response = await makeJSONRPCCall<ServiceDescription>(
        'R0D0Service.Describe',
        null
      )
      
      console.log('✅ Service described:', {
        name: response.service,
        version: response.version,
        methods: Object.keys(response.methods).length,
        areas: response.discovery_areas.length
      })
      
      return response
      
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Error al obtener descripción'
      console.error('❌ Error in describe:', errorMsg)
      setError(errorMsg)
      throw err
    } finally {
      setIsLoading(false)
    }
  }, [makeJSONRPCCall])
  
  return {
    discoverStart,
    discoverContinue,
    discoverComplete,
    describe,
    isLoading,
    error,
    sessionId
  }
}

// Función auxiliar para validar si R0D0 está disponible via proxy
export const checkR0D0Health = async (baseUrl: string = DEFAULT_CONFIG.baseUrl): Promise<boolean> => {
  try {
    const response = await fetch(baseUrl, {
      method: 'GET',
      headers: {
        'Accept': 'application/json'
      },
      signal: AbortSignal.timeout(5000)
    })
    
    if (!response.ok) {
      return false
    }
    
    const health = await response.json()
    return health.healthy === true
    
  } catch (err) {
    console.error('❌ R0D0 health check failed:', err)
    return false
  }
} 