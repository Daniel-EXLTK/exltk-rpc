import { NextRequest, NextResponse } from 'next/server';

export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const { agentUrl, message } = body;
    
    if (!agentUrl || !message) {
      return NextResponse.json(
        { error: 'URL del agente y mensaje requeridos' },
        { status: 400 }
      );
    }
    
    // Normalizar URL para evitar doble slash
    const normalizedUrl = agentUrl.replace(/\/+$/, '').replace(/\/+/g, '/').replace(/:\//g, '://');
    const finalUrl = normalizedUrl.endsWith('/') ? normalizedUrl.slice(0, -1) : normalizedUrl;
    
    // Convertir mensaje A2A nativo a formato JSON-RPC esperado por el servidor
    const jsonRpcMessage = {
      jsonrpc: "2.0",
      method: "message/send",  // CORREGIDO: singular, no plural
      params: {
        message: message  // El mensaje va dentro del objeto params como "message"
      },
      id: message.messageId || Date.now()
    };
    
    console.log(`🔍 Proxy: Enviando mensaje A2A a ${finalUrl}`);
    console.log(`📤 Mensaje JSON-RPC:`, jsonRpcMessage);
    
    // Hacer la request desde el servidor Next.js (sin restricciones CORS)
    const response = await fetch(finalUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(jsonRpcMessage)
    });
    
    if (!response.ok) {
      console.error(`❌ Proxy: Error ${response.status} al enviar mensaje A2A`);
      return NextResponse.json(
        { error: `Error ${response.status}: ${response.statusText}` },
        { status: response.status }
      );
    }
    
    const responseData = await response.json();
    console.log(`✅ Proxy: Respuesta A2A recibida`, responseData);
    
    return NextResponse.json(responseData);
    
  } catch (error) {
    console.error('❌ Proxy: Error al enviar mensaje A2A:', error);
    return NextResponse.json(
      { error: 'Error interno del servidor' },
      { status: 500 }
    );
  }
} 