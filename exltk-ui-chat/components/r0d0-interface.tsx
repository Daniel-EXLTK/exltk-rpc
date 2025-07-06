"use client"

import { useState, useRef, useEffect } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { useOrchestratorClient } from "@/lib/useOrchestratorClient"

type Message = {
  id: string
  content: string
  sender: 'user' | 'r0d0'
  mood?: string
  timestamp?: Date
  isTyping?: boolean
  context?: string
}

// Función para obtener emoji basado en el estado de ánimo
const getMoodEmoji = (mood: string) => {
  // Lista de emojis conocidos
  const knownEmojis = ['😊', '🚀', '🤔', '💭', '🎉', '😰', '🤖']
  
  // Si ya es un emoji conocido, devolverlo directamente
  if (knownEmojis.includes(mood)) {
    return mood
  }
  
  // Mapear valores antiguos a emojis
  const moodMap: { [key: string]: string } = {
    'happy': '😊',
    'excited': '🚀',
    'thoughtful': '🤔',
    'confused': '💭',
    'worried': '😰',
    'celebrating': '🎉'
  }
  
  return moodMap[mood] || '😊'
}

// Función para calcular tiempo de escritura realista
const calculateTypingTime = (message: string): number => {
  const wordsPerMinute = 40 // Velocidad de escritura promedio
  const words = message.split(' ').length
  const baseTime = (words / wordsPerMinute) * 60 * 1000 // En milisegundos
  return Math.max(1500, Math.min(baseTime, 4000)) // Entre 1.5-4 segundos
}

export default function R0D0Interface() {
  const [messages, setMessages] = useState<Message[]>([
    {
      id: '1',
      content: "¡Hola! Soy R0D0, tu asistente para propuestas de proyectos. Voy a entrevistarte sobre tu proyecto para crear la mejor propuesta posible. ¿Podrías contarme sobre qué tipo de proyecto tienes en mente?",
      sender: 'r0d0',
      mood: '😊',
      timestamp: new Date(),
    }
  ])
  
  const [input, setInput] = useState("")
  const [isTyping, setIsTyping] = useState(false)
  const [currentMood, setCurrentMood] = useState('😊')
  const messagesEndRef = useRef<HTMLDivElement>(null)
  
  const { sendMessage, isLoading, error, conversationContext, clearContext } = useOrchestratorClient()

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages, isTyping])

  const simulateTyping = (duration: number = 2000) => {
    setIsTyping(true)
    setTimeout(() => setIsTyping(false), duration)
  }

  const handleSend = async () => {
    if (!input.trim() || isLoading) return

    const userMessage: Message = {
      id: Date.now().toString(),
      content: input,
      sender: 'user',
      timestamp: new Date()
    }

    setMessages(prev => [...prev, userMessage])
    setInput('')
    setIsTyping(true)

    try {
      console.log('🔄 Enviando mensaje con contexto:', conversationContext)
      const response = await sendMessage('ui-user', input)
      
      console.log('📥 Respuesta recibida:', response)
      
      // Simular tiempo de escritura realista
      const typingDuration = Math.max(1500, Math.min(4000, response.response.length * 50))
      setTimeout(() => {
        const r0d0Message: Message = {
          id: Date.now().toString() + '-r0d0',
          content: response.response,
          sender: 'r0d0',
          timestamp: new Date(),
          mood: currentMood,
          context: response.context
        }
        
        setMessages(prev => [...prev, r0d0Message])
        setIsTyping(false)
        
        // Actualizar mood basado en el contexto
        if (response.context) {
          const moods = ['😊', '🚀', '🤔', '💭', '🎉']
          const randomMood = moods[Math.floor(Math.random() * moods.length)]
          setCurrentMood(randomMood)
        }
      }, typingDuration)
      
    } catch (error) {
      console.error('❌ Error:', error)
      setIsTyping(false)
      
      const errorMessage: Message = {
        id: Date.now().toString() + '-error',
        content: 'Lo siento, hubo un problema al procesar tu mensaje. ¿Podrías intentar de nuevo?',
        sender: 'r0d0',
        timestamp: new Date(),
        mood: '😰'
      }
      
      setMessages(prev => [...prev, errorMessage])
    }
  }

  const handleClearChat = () => {
    setMessages([
      {
        id: '1',
        content: "¡Hola! Soy R0D0, tu asistente para propuestas de proyectos. Voy a entrevistarte sobre tu proyecto para crear la mejor propuesta posible. ¿Podrías contarme sobre qué tipo de proyecto tienes en mente?",
        sender: 'r0d0',
        mood: '😊',
        timestamp: new Date(),
      }
    ])
    clearContext()
    setCurrentMood('😊')
  }

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  return (
    <div className="flex flex-col h-full max-w-4xl mx-auto">
      {/* Header */}
      <div className="p-4 border-b bg-gradient-to-r from-teal-50 to-blue-50">
        <div className="flex items-center gap-3">
          <div className="text-2xl">🤖</div>
          <div>
            <h1 className="text-xl font-bold text-teal-800">R0D0 Discovery</h1>
            <p className="text-sm text-gray-600">
              Conversación natural para conocer tu proyecto
            </p>
          </div>
        </div>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.map((message) => (
          <div
            key={message.id}
            className={`flex ${message.sender === "user" ? "justify-end" : "justify-start"}`}
          >
            <div
              className={`max-w-[80%] rounded-lg px-4 py-2 ${
                message.sender === "user"
                  ? "bg-teal-600 text-white"
                  : "bg-gray-100 text-gray-800"
              }`}
            >
              {message.sender === "r0d0" && (
                <div className="flex items-center gap-2 mb-1">
                  <span className="text-lg">{getMoodEmoji(message.mood || "happy")}</span>
                  <span className="text-xs font-medium text-gray-500">R0D0</span>
                </div>
              )}
              <p className="whitespace-pre-wrap">{message.content}</p>
              {message.timestamp && (
                <p className="text-xs opacity-70 mt-1">
                  {message.timestamp.toLocaleTimeString()}
                </p>
              )}
            </div>
          </div>
        ))}
        
        {/* Typing indicator */}
        {isTyping && (
          <div className="flex justify-start">
            <div className="max-w-[80%] rounded-lg px-4 py-2 bg-gray-100 text-gray-800">
              <div className="flex items-center gap-2 mb-1">
                <span className="text-lg">💭</span>
                <span className="text-xs font-medium text-gray-500">R0D0</span>
              </div>
              <div className="flex items-center gap-1">
                <span className="text-sm text-gray-600">
                  Escribiendo...
                </span>
                <div className="flex space-x-1 ml-2">
                  <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce"></div>
                  <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{animationDelay: '0.1s'}}></div>
                  <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{animationDelay: '0.2s'}}></div>
                </div>
              </div>
            </div>
          </div>
        )}
        
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div className="p-4 border-t bg-white">
        <div className="flex gap-2">
          <Input
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyPress}
            placeholder="Describe tu proyecto o responde a las preguntas..."
            disabled={isLoading || isTyping}
            className="flex-1"
          />
          <Button
            onClick={handleSend}
            disabled={!input.trim() || isLoading || isTyping}
            className="bg-teal-600 hover:bg-teal-700"
          >
            {isLoading || isTyping ? "..." : "Enviar"}
          </Button>
        </div>
        
        {/* Estado */}
        <div className="mt-2 text-xs text-gray-500 text-center">
          {isTyping ? "R0D0 está escribiendo..." : "Listo para conversar"}
        </div>
      </div>
    </div>
  )
}
