#!/bin/bash

# Script de prueba para el Orquestador LLM
# Este script configura las variables de entorno y prueba los endpoints JSON-RPC

set -e

echo "🚀 Iniciando pruebas del Orquestador LLM..."

# Configurar variables de entorno
export GOOGLE_API_KEY="AIzaSyD2PFVVO91YR3o7iuV9zV8jjAMidNOcapA"
export SERVICE_URLS="http://localhost:8501"
export PORT="8502"
export LLM_PROVIDER="gemini"
export LLM_MODEL="gemini-2.0-flash-exp"
export ENABLE_LOGGING="true"

echo "📋 Configuración:"
echo "  - Puerto: $PORT"
echo "  - Servicios: $SERVICE_URLS"
echo "  - LLM: $LLM_PROVIDER/$LLM_MODEL"
echo "  - Logging: $ENABLE_LOGGING"

# Función para hacer requests JSON-RPC
make_request() {
    local method=$1
    local params=$2
    local description=$3
    
    echo ""
    echo "🔍 Probando: $description"
    echo "   Método: $method"
    
    response=$(curl -s -X POST http://localhost:$PORT/ \
        -H "Content-Type: application/json" \
        -d "{
            \"jsonrpc\": \"2.0\",
            \"id\": \"test-$(date +%s)\",
            \"method\": \"$method\",
            \"params\": $params
        }")
    
    echo "   Respuesta:"
    echo "$response" | jq '.' 2>/dev/null || echo "$response"
}

# Verificar si el servidor está corriendo
echo ""
echo "🔌 Verificando si el servidor está corriendo..."
if ! curl -s http://localhost:$PORT/ > /dev/null 2>&1; then
    echo "❌ El servidor no está corriendo en puerto $PORT"
    echo "   Ejecuta: ./orchestrator-server"
    exit 1
fi

echo "✅ Servidor detectado en puerto $PORT"

# Probar health check
make_request "health" "{}" "Health Check"

# Probar métricas
make_request "metrics" "{}" "Métricas"

# Probar obtener capacidades
make_request "get_capabilities" "{}" "Obtener Capacidades"

# Probar orquestar solicitud
make_request "orchestrate" '{
    "user_id": "test-user-123",
    "message": "Quiero crear una aplicación móvil para gestionar tareas de mi equipo",
    "context": {
        "project_type": "mobile_app",
        "priority": "high",
        "team_size": "5-10"
    }
}' "Orquestar Solicitud"

# Probar refrescar capacidades
make_request "refresh_capabilities" "{}" "Refrescar Capacidades"

echo ""
echo "✅ Todas las pruebas completadas!"
echo ""
echo "📊 Para monitorear logs en tiempo real:"
echo "   tail -f orchestrator.log"
echo ""
echo "🔧 Para detener el servidor:"
echo "   pkill -f orchestrator-server" 