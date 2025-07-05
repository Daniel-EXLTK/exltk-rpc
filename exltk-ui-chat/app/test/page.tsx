"use client"

import { useState } from 'react'

export default function TestPage() {
  const [result, setResult] = useState<string>('')
  const [loading, setLoading] = useState(false)

  const testA2AConnection = async () => {
    setLoading(true)
    setResult('🔄 Testing A2A Connection...\n')
    
    try {
      // Test 1: Agent Discovery
      setResult(prev => prev + '📡 Step 1: Agent Discovery...\n')
      const agentResponse = await fetch('http://localhost:8501/.well-known/agent.json')
      const agentCard = await agentResponse.json()
      setResult(prev => prev + `✅ Agent discovered: ${agentCard.name}\n`)
      
      // Test 2: A2A Message
      setResult(prev => prev + '📤 Step 2: Sending A2A message...\n')
      const message = {
        method: 'message/send',
        params: {
          message: {
            messageId: `test_${Date.now()}`,
            role: 'user',
            parts: [{ text: '¡Hola R0D0! Test desde página de debug' }],
            metadata: {
              source: 'test_page',
              timestamp: new Date().toISOString()
            }
          }
        },
        jsonrpc: '2.0',
        id: 1
      }
      
      const response = await fetch('http://localhost:8501/', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Origin': 'http://localhost:3000'
        },
        body: JSON.stringify(message)
      })
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }
      
      const responseData = await response.json()
      setResult(prev => prev + '✅ A2A Response received!\n')
      setResult(prev => prev + `📋 Status: ${responseData.result?.status?.state}\n`)
      setResult(prev => prev + `💬 Response: ${responseData.result?.status?.message?.parts[0]?.text?.substring(0, 200)}...\n`)
      setResult(prev => prev + '\n🎉 A2A Connection Test SUCCESSFUL!')
      
    } catch (error) {
      setResult(prev => prev + `❌ Error: ${error}\n`)
    } finally {
      setLoading(false)
    }
  }

  const testMainUIConnection = async () => {
    setLoading(true)
    setResult('🔄 Testing Main UI A2A Hook...\n')
    
    try {
      // Simular exactamente lo que hace useA2AClient
      const agentUrl = 'http://localhost:8501'
      
      // Discovery
      setResult(prev => prev + '🔍 Discovering agent...\n')
      const cardResponse = await fetch(`${agentUrl}/.well-known/agent.json`, {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' }
      })
      
      if (!cardResponse.ok) {
        throw new Error(`Discovery failed: ${cardResponse.status}`)
      }
      
      const agentCard = await cardResponse.json()
      setResult(prev => prev + `✅ Agent Card: ${agentCard.name}\n`)
      
      // Message
      setResult(prev => prev + '📤 Sending message via A2A...\n')
      const messageId = `ui_msg_${Date.now()}_test`
      const taskId = `ui_task_${Date.now()}_test`
      
      const a2aMessage = {
        messageId: messageId,
        role: 'user',
        parts: [{ text: 'Test desde hook UI simulado' }],
        metadata: {
          source: 'ui',
          timestamp: new Date().toISOString(),
          taskId: taskId
        }
      }
      
      const endpoint = agentCard.http_endpoint || agentUrl
      const response = await fetch(`${endpoint}/`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          method: 'message/send',
          params: { message: a2aMessage },
          jsonrpc: '2.0',
          id: 1
        })
      })
      
      if (!response.ok) {
        throw new Error(`Message failed: ${response.status}`)
      }
      
      const responseData = await response.json()
      setResult(prev => prev + '✅ Message sent successfully!\n')
      setResult(prev => prev + `📋 Task State: ${responseData.result?.status?.state}\n`)
      setResult(prev => prev + '\n🎉 Main UI Hook Test SUCCESSFUL!')
      
    } catch (error) {
      setResult(prev => prev + `❌ Hook Test Error: ${error}\n`)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="container mx-auto p-8">
      <h1 className="text-3xl font-bold mb-6">🧪 A2A Connection Test</h1>
      
      <div className="space-y-4 mb-6">
        <button
          onClick={testA2AConnection}
          disabled={loading}
          className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded disabled:opacity-50"
        >
          {loading ? '⏳ Testing...' : '🔍 Test Basic A2A Connection'}
        </button>
        
        <button
          onClick={testMainUIConnection}
          disabled={loading}
          className="bg-green-500 hover:bg-green-700 text-white font-bold py-2 px-4 rounded disabled:opacity-50 ml-4"
        >
          {loading ? '⏳ Testing...' : '🎯 Test UI Hook Flow'}
        </button>
      </div>
      
      <div className="bg-gray-100 p-4 rounded-lg">
        <h2 className="text-xl font-semibold mb-2">📊 Test Results:</h2>
        <pre className="whitespace-pre-wrap text-sm font-mono">
          {result || 'Click a test button to start...'}
        </pre>
      </div>
      
      <div className="mt-6 p-4 bg-blue-50 rounded-lg">
        <h3 className="font-semibold mb-2">🔧 Debug Info:</h3>
        <ul className="text-sm space-y-1">
          <li>• UI URL: {typeof window !== 'undefined' ? window.location.origin : 'http://localhost:3000'}</li>
          <li>• Agent URL: http://localhost:8501</li>
          <li>• Current Time: {new Date().toISOString()}</li>
        </ul>
      </div>
    </div>
  )
} 