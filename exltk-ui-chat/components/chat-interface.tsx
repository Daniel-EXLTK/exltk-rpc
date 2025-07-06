"use client"

import { useState, useRef, useEffect } from "react"
import { Send, Bot, User, Loader2, FileText, Menu, X, Sparkles, Clock } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"
import { useOrchestratorClient } from "@/lib/useOrchestratorClient"
import Image from "next/image"

type Message = {
  id: string
  content: string
  role: "user" | "assistant"
  timestamp: Date
  mood?: "happy" | "thinking" | "worried" | "sad" | "celebrating" | "smiling"
}

type Proposal = {
  id: string
  title: string
  summary: string
  created_at: Date
  markdown_content: string
  status: "draft" | "in_progress" | "completed"
}

export function ChatInterface() {
  const [messages, setMessages] = useState<Message[]>([
    {
      id: "1",
      content: "¡Hola! Soy R0D0, tu asistente para propuestas de proyectos. Voy a entrevistarte sobre tu proyecto para crear la mejor propuesta posible. ¿Podrías contarme sobre qué tipo de proyecto tienes en mente?",
      role: "assistant",
      timestamp: new Date(),
      mood: "happy"
    },
  ])
  const [input, setInput] = useState("")
  const [isLoading, setIsLoading] = useState(false)
  const [isSideboardOpen, setIsSideboardOpen] = useState(false)
  const [isDraftViewerOpen, setIsDraftViewerOpen] = useState(false)
  const [currentProposal, setCurrentProposal] = useState<Proposal | null>(null)
  const [proposalHistory, setProposalHistory] = useState<Proposal[]>([])
  const messagesEndRef = useRef<HTMLDivElement>(null)

  const { sendMessage, isLoading: orchestratorLoading, conversationContext } = useOrchestratorClient()

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  const getAvatarImage = (mood: string = "happy") => {
    return `/images/avatar-${mood}.png`
  }

  const getMoodFromResponse = (response: string): Message['mood'] => {
    if (response.includes("perfecto") || response.includes("excelente") || response.includes("fantástico")) {
      return "celebrating"
    } else if (response.includes("¿") || response.includes("cuéntame") || response.includes("háblame")) {
      return "thinking"
    } else if (response.includes("preocupa") || response.includes("problema") || response.includes("error")) {
      return "worried"
    } else if (response.includes("lamento") || response.includes("disculpa") || response.includes("sorry")) {
      return "sad"
    } else if (response.includes("bien") || response.includes("correcto") || response.includes("entiendo")) {
      return "smiling"
    }
    return "happy"
  }

  const handleSend = async () => {
    if (!input.trim() || isLoading) return

    const userMessage: Message = {
      id: Date.now().toString(),
      content: input,
      role: "user",
      timestamp: new Date(),
    }

    setMessages((prev) => [...prev, userMessage])
    setInput("")
    setIsLoading(true)

    try {
      const response = await sendMessage("ui-user", input)
      
      const assistantMessage: Message = {
        id: (Date.now() + 1).toString(),
        content: response.response,
        role: "assistant",
        timestamp: new Date(),
        mood: getMoodFromResponse(response.response)
      }

      setMessages((prev) => [...prev, assistantMessage])

      // Si la respuesta incluye una propuesta, simular generación de draft
      if (response.response.includes("propuesta") || response.response.includes("proyecto")) {
        generateDraftProposal(input, response.response)
      }

    } catch (error) {
      console.error("Error al enviar mensaje:", error)
      const errorMessage: Message = {
        id: (Date.now() + 1).toString(),
        content: "Lo siento, hubo un error al procesar tu mensaje. ¿Podrías intentar nuevamente?",
        role: "assistant",
        timestamp: new Date(),
        mood: "worried"
      }
      setMessages((prev) => [...prev, errorMessage])
    } finally {
      setIsLoading(false)
    }
  }

  const generateDraftProposal = (userInput: string, assistantResponse: string) => {
    const proposal: Proposal = {
      id: Date.now().toString(),
      title: `Propuesta de proyecto - ${new Date().toLocaleDateString()}`,
      summary: `Propuesta generada basada en: ${userInput.substring(0, 50)}...`,
      created_at: new Date(),
      status: "draft",
      markdown_content: `# Propuesta de Proyecto

## Contexto
${userInput}

## Análisis de R0D0
${assistantResponse}

## Próximos Pasos
- Definir alcance específico
- Establecer cronograma
- Definir recursos necesarios
- Validar requisitos técnicos

*Generado automáticamente por R0D0 - ${new Date().toLocaleString()}*`
    }

    setCurrentProposal(proposal)
    setProposalHistory(prev => [proposal, ...prev])
    setIsDraftViewerOpen(true)
  }

  return (
    <div className="flex h-screen bg-gray-50">
      {/* Sideboard - Historial de Propuestas */}
      <div className={cn(
        "w-80 bg-white border-r border-gray-200 shadow-sm transition-all duration-300",
        isSideboardOpen ? "translate-x-0" : "-translate-x-full md:translate-x-0"
      )}>
        <div className="p-4 border-b border-gray-200 bg-gray-50">
          <div className="flex items-center gap-2 mb-2">
            <Image
              src="/images/exltk-logo.png"
              alt="EXLTK Logo"
              width={24}
              height={24}
              className="rounded"
            />
            <h2 className="font-semibold text-gray-800">Historial de Propuestas</h2>
          </div>
          <p className="text-sm text-gray-600">
            Acá debe estar el historial de propuestas que ha interactuado el usuario
          </p>
        </div>
        
        <div className="flex-1 overflow-y-auto p-4">
          {proposalHistory.length === 0 ? (
            <div className="text-center text-gray-500 py-8">
              <FileText className="h-12 w-12 mx-auto mb-2 text-gray-300" />
              <p className="text-sm">No hay propuestas aún</p>
            </div>
          ) : (
            <div className="space-y-3">
              {proposalHistory.map((proposal) => (
                <div
                  key={proposal.id}
                  className="p-3 border border-gray-200 rounded-lg hover:bg-gray-50 cursor-pointer transition-colors"
                  onClick={() => {
                    setCurrentProposal(proposal)
                    setIsDraftViewerOpen(true)
                  }}
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <h3 className="font-medium text-sm text-gray-800 mb-1 line-clamp-1">
                        {proposal.title}
                      </h3>
                      <p className="text-xs text-gray-600 mb-2 line-clamp-2">
                        {proposal.summary}
                      </p>
                      <div className="flex items-center gap-2 text-xs text-gray-500">
                        <Clock className="h-3 w-3" />
                        {proposal.created_at.toLocaleDateString()}
                      </div>
                    </div>
                    <div className={cn(
                      "px-2 py-1 rounded-full text-xs font-medium",
                      proposal.status === "draft" ? "bg-yellow-100 text-yellow-800" :
                      proposal.status === "in_progress" ? "bg-blue-100 text-blue-800" :
                      "bg-green-100 text-green-800"
                    )}>
                      {proposal.status}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Chat Central */}
      <div className="flex-1 flex flex-col">
        {/* Header */}
        <div className="bg-white border-b border-gray-200 p-4 shadow-sm">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Button
                variant="ghost"
                size="icon"
                className="md:hidden"
                onClick={() => setIsSideboardOpen(!isSideboardOpen)}
              >
                <Menu className="h-5 w-5" />
              </Button>
              <div className="flex items-center gap-2">
                <div className="w-8 h-8 bg-green-500 rounded-full flex items-center justify-center">
                  <Sparkles className="h-4 w-4 text-white" />
                </div>
                <div>
                  <h1 className="font-semibold text-gray-800">EXLTK</h1>
                  <p className="text-xs text-gray-600">Agente conectado (ADK)</p>
                </div>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <div className="text-sm text-gray-600">
                Fase 1 de 4
              </div>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setIsDraftViewerOpen(!isDraftViewerOpen)}
                className="text-sm"
              >
                <FileText className="h-4 w-4 mr-1" />
                Visor de draft
              </Button>
            </div>
          </div>
        </div>

        {/* Messages */}
        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          {messages.map((message) => (
            <div
              key={message.id}
              className={cn(
                "flex items-start gap-3 max-w-[80%]",
                message.role === "user" ? "ml-auto flex-row-reverse" : ""
              )}
            >
              <div className="h-10 w-10 rounded-full flex items-center justify-center shrink-0 overflow-hidden">
                {message.role === "user" ? (
                  <div className="w-full h-full bg-gray-200 rounded-full flex items-center justify-center">
                    <User className="h-5 w-5 text-gray-600" />
                  </div>
                ) : (
                  <Image
                    src={getAvatarImage(message.mood)}
                    alt="R0D0 Avatar"
                    width={40}
                    height={40}
                    className="rounded-full"
                  />
                )}
              </div>
              <div className={cn(
                "rounded-lg p-3 max-w-full",
                message.role === "user" 
                  ? "bg-teal-500 text-white" 
                  : "bg-white border border-gray-200 shadow-sm"
              )}>
                <div className="text-sm whitespace-pre-wrap">
                  {message.content}
                </div>
                <div className="text-xs mt-1 opacity-70">
                  {message.timestamp.toLocaleTimeString()}
                </div>
              </div>
            </div>
          ))}
          {isLoading && (
            <div className="flex items-start gap-3 max-w-[80%]">
              <div className="h-10 w-10 rounded-full flex items-center justify-center shrink-0 overflow-hidden">
                <Image
                  src={getAvatarImage("thinking")}
                  alt="R0D0 Avatar"
                  width={40}
                  height={40}
                  className="rounded-full"
                />
              </div>
              <div className="bg-white border border-gray-200 shadow-sm rounded-lg p-3">
                <div className="flex items-center gap-2">
                  <Loader2 className="h-4 w-4 animate-spin text-teal-500" />
                  <span className="text-sm text-gray-600">R0D0 está pensando...</span>
                </div>
              </div>
            </div>
          )}
          <div ref={messagesEndRef} />
        </div>

        {/* Input */}
        <div className="border-t border-gray-200 p-4 bg-white">
          <form
            onSubmit={(e) => {
              e.preventDefault()
              handleSend()
            }}
            className="flex gap-2"
          >
            <Input
              placeholder="Describe tu proyecto..."
              value={input}
              onChange={(e) => setInput(e.target.value)}
              className="flex-1"
              disabled={isLoading}
            />
            <Button type="submit" disabled={isLoading} className="bg-teal-500 hover:bg-teal-600">
              <Send className="h-4 w-4" />
              <span className="sr-only">Enviar mensaje</span>
            </Button>
          </form>
        </div>
      </div>

      {/* Visor de Draft de Propuesta */}
      <div className={cn(
        "w-96 bg-white border-l border-gray-200 shadow-sm transition-all duration-300",
        isDraftViewerOpen ? "translate-x-0" : "translate-x-full"
      )}>
        <div className="p-4 border-b border-gray-200 bg-gray-50">
          <div className="flex items-center justify-between">
            <h2 className="font-semibold text-gray-800">Visor de draft de propuesta en .md</h2>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => setIsDraftViewerOpen(false)}
            >
              <X className="h-4 w-4" />
            </Button>
          </div>
        </div>
        
        <div className="flex-1 overflow-y-auto p-4">
          {currentProposal ? (
            <div className="space-y-4">
              <div className="bg-gray-50 p-3 rounded-lg">
                <h3 className="font-medium text-gray-800 mb-1">
                  {currentProposal.title}
                </h3>
                <p className="text-sm text-gray-600 mb-2">
                  {currentProposal.summary}
                </p>
                <div className="text-xs text-gray-500">
                  Creado: {currentProposal.created_at.toLocaleString()}
                </div>
              </div>
              
              <div className="prose prose-sm max-w-none">
                <pre className="whitespace-pre-wrap text-xs bg-gray-50 p-3 rounded border font-mono overflow-x-auto">
                  {currentProposal.markdown_content}
                </pre>
              </div>
            </div>
          ) : (
            <div className="text-center text-gray-500 py-8">
              <FileText className="h-12 w-12 mx-auto mb-2 text-gray-300" />
              <p className="text-sm">No hay propuesta seleccionada</p>
              <p className="text-xs text-gray-400 mt-1">
                Las propuestas aparecerán aquí durante la conversación
              </p>
            </div>
          )}
        </div>
      </div>

      {/* Overlay para mobile */}
      {isSideboardOpen && (
        <div
          className="fixed inset-0 bg-black bg-opacity-50 z-10 md:hidden"
          onClick={() => setIsSideboardOpen(false)}
        />
      )}
    </div>
  )
}
