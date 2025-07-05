import { NextRequest, NextResponse } from 'next/server';

export async function GET(request: NextRequest) {
  try {
    const searchParams = request.nextUrl.searchParams;
    const agentUrl = searchParams.get('url');
    
    if (!agentUrl) {
      return NextResponse.json(
        { error: 'URL del agente requerida' },
        { status: 400 }
      );
    }
    
    console.log(`🔍 Proxy: Obteniendo Agent Card de ${agentUrl}/.well-known/agent.json`);
    
    // Hacer la request desde el servidor Next.js (sin restricciones CORS)
    const response = await fetch(`${agentUrl}/.well-known/agent.json`);
    
    if (!response.ok) {
      console.error(`❌ Proxy: Error ${response.status} al obtener Agent Card`);
      return NextResponse.json(
        { error: `Error ${response.status}: ${response.statusText}` },
        { status: response.status }
      );
    }
    
    const agentCard = await response.json();
    console.log(`✅ Proxy: Agent Card obtenido exitosamente`, agentCard);
    
    return NextResponse.json(agentCard);
    
  } catch (error) {
    console.error('❌ Proxy: Error al obtener Agent Card:', error);
    return NextResponse.json(
      { error: 'Error interno del servidor' },
      { status: 500 }
    );
  }
} 