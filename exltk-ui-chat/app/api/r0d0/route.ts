import { NextRequest, NextResponse } from 'next/server'

// Configuración del servicio Orquestador
const ORCHESTRATOR_URL = process.env.ORCHESTRATOR_URL || 'http://localhost:8502'

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
    
    // Hacer la request desde el servidor Next.js (sin restricciones CORS)
    const response = await fetch(ORCHESTRATOR_URL, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json'
      },
      body: JSON.stringify(jsonRpcRequest)
    })
    
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
      { error: 'Error interno del servidor' },
      { status: 500 }
    )
  }
}

// Endpoint para health check directo
export async function GET() {
  try {
    console.log(`🔍 Orchestrator Proxy: Health check a ${ORCHESTRATOR_URL}/health`)
    
    const response = await fetch(`${ORCHESTRATOR_URL}`, {
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
    })
    
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