import { NextRequest, NextResponse } from 'next/server'

// Configuración del servicio Orquestador
const ORCHESTRATOR_URL = process.env.ORCHESTRATOR_URL || 'http://localhost:8502'

// Función para hacer retry con backoff exponencial
async function fetchWithRetry(url: string, options: RequestInit, maxRetries = 3, baseDelay = 1000) {
  let lastError: Error | null = null
  
  for (let attempt = 0; attempt <= maxRetries; attempt++) {
    try {
      const response = await fetch(url, options)
      return response
    } catch (error) {
      lastError = error as Error
      
      // No reintentamos en el último intento
      if (attempt === maxRetries) {
        break
      }
      
      // Verificamos si es un error de conectividad que vale la pena reintentar
      const isRetryableError = 
        error instanceof TypeError && 
        (error.message.includes('fetch failed') || 
         error.message.includes('ECONNREFUSED') || 
         error.message.includes('other side closed'))
      
      if (!isRetryableError) {
        break
      }
      
      // Esperamos antes del siguiente intento (backoff exponencial)
      const delay = baseDelay * Math.pow(2, attempt)
      console.log(`⏳ Retry ${attempt + 1}/${maxRetries} en ${delay}ms debido a: ${error.message}`)
      await new Promise(resolve => setTimeout(resolve, delay))
    }
  }
  
  throw lastError
}

export async function POST(request: NextRequest) {
  try {
    const body = await request.json()
    const { method, params, id } = body
    
    if (!method) {
      return NextResponse.json(
        { error: 'Método JSON-RPC requerido' },
        { status: 400 }
      )
    }
    
    // Construir request JSON-RPC
    const jsonRpcRequest = {
      jsonrpc: "2.0",
      method: method,
      params: params || null,
      id: String(id || Date.now())
    }
    
    console.log(`🔍 Orchestrator Proxy: Enviando ${method} a ${ORCHESTRATOR_URL}`)
    console.log(`📤 JSON-RPC Request:`, jsonRpcRequest)
    
    // Hacer la request con retry logic
    const response = await fetchWithRetry(ORCHESTRATOR_URL, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json'
      },
      body: JSON.stringify(jsonRpcRequest)
    }, 3, 1000) // 3 reintentos, empezando con 1 segundo de delay
    
    if (!response.ok) {
      console.error(`❌ Orchestrator Proxy: Error ${response.status} al llamar ${method}`)
      return NextResponse.json(
        { error: `Error ${response.status}: ${response.statusText}` },
        { status: response.status }
      )
    }
    
    const responseData = await response.json()
    console.log(`✅ Orchestrator Proxy: Respuesta recibida para ${method}`, {
      success: !!responseData.result,
      error: responseData.error
    })
    
    return NextResponse.json(responseData)
    
  } catch (error) {
    console.error('❌ Orchestrator Proxy: Error al procesar request:', error)
    return NextResponse.json(
      { error: 'Error interno del servidor - servicio no disponible' },
      { status: 503 }
    )
  }
}

// Endpoint para health check directo con retry
export async function GET() {
  try {
    console.log(`🔍 Orchestrator Proxy: Health check a ${ORCHESTRATOR_URL}`)
    
    const response = await fetchWithRetry(ORCHESTRATOR_URL, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json'
      },
      body: JSON.stringify({
        jsonrpc: '2.0',
        id: 'health-check',
        method: 'health',
        params: {}
      })
    }, 2, 500) // 2 reintentos para health check, más rápido
    
    if (!response.ok) {
      return NextResponse.json(
        { healthy: false, error: `HTTP ${response.status}` },
        { status: 503 }
      )
    }
    
    const healthData = await response.json()
    console.log(`✅ Orchestrator Proxy: Health check exitoso`, healthData)
    
    return NextResponse.json({
      healthy: true,
      service: healthData.result || healthData
    })
    
  } catch (error) {
    console.error('❌ Orchestrator Proxy: Error en health check:', error)
    return NextResponse.json(
      { healthy: false, error: 'Service unavailable' },
      { status: 503 }
    )
  }
} 