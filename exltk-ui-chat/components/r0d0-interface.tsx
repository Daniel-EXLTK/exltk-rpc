"use client"

import { useState, useRef, useEffect } from "react"
import { Send, Circle, Sparkles, Brain, CheckCircle, AlertCircle } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"
import { useOrchestratorClient, OrchestratorResponse } from "@/lib/useOrchestratorClient"
import { ProgressIndicator } from "@/components/ui/progress-indicator"

type Message = {
  id: string
  content: string
  role: "user" | "assistant"
  step?: string
  mood?: "happy" | "celebrating" | "worried" | "sad" | "smiling" | "thinking"
  timestamp?: Date
}

type InterviewStep = "initial" | "project_details" | "requirements" | "proposal" | "decision" | "accepted" | "rejected"

// Tipos para el nuevo sistema de slots estructurados
type SlotKey = 
  | "proyecto.titulo"
  | "proyecto.objetivo" 
  | "proyecto.publico"
  | "problema"
  | "resultados_deseados"
  | "alcance"
  | "cronograma"
  | "recursos"
  | "indicadores_exito"

type ProjectSlots = {
  [K in SlotKey]?: string
}

export function R0D0Interface() {
  const [messages, setMessages] = useState<Message[]>([
    {
      id: "1",
      content:
        "¡Hola! Soy R0D0, tu asistente para propuestas de proyectos. Voy a entrevistarte sobre tu proyecto para crear la mejor propuesta posible. ¿Podrías contarme sobre qué tipo de proyecto tienes en mente?",
      role: "assistant",
      step: "initial",
      mood: "happy",
      timestamp: new Date(),
    },
  ])
  const [input, setInput] = useState("")
  const [currentStep, setCurrentStep] = useState<InterviewStep>("initial")
  const [projectData, setProjectData] = useState({
    type: "",
    requirements: "",
    timeline: "",
    budget: "",
  })
  
  // Nuevo estado para slots estructurados
  const [projectSlots, setProjectSlots] = useState<ProjectSlots>({})
  const [currentSlot, setCurrentSlot] = useState<number>(1)
  const [completedSlots, setCompletedSlots] = useState<SlotKey[]>([])
  const [isCollectingSlots, setIsCollectingSlots] = useState<boolean>(false)
  const [isInitialized, setIsInitialized] = useState<boolean>(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  // Nuevo: usar hook del orquestador
  const { sendMessage, isLoading, error } = useOrchestratorClient()

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  const handleSend = async () => {
    if (!input.trim()) return
    const userMessage: Message = {
      id: Date.now().toString(),
      content: input,
      role: "user",
      timestamp: new Date(),
    }
    setMessages((prev) => [...prev, userMessage])
    const userInput = input
    setInput("")
    try {
      // Usar orquestador vía JSON-RPC
      const orchestratorResp: OrchestratorResponse = await sendMessage("ui-user", userInput, {})
      const assistantMessage: Message = {
        id: Date.now().toString() + "-a",
        content: orchestratorResp.response || "(Sin respuesta)",
        role: "assistant",
        mood: orchestratorResp.success ? "happy" : "worried",
        timestamp: new Date(),
      }
      setMessages((prev) => [...prev, assistantMessage])
    } catch (err) {
      const assistantMessage: Message = {
        id: Date.now().toString() + "-e",
        content: error || "Error al contactar orquestador",
        role: "assistant",
        mood: "worried",
        timestamp: new Date(),
      }
      setMessages((prev) => [...prev, assistantMessage])
    }
  }

  const generateR0D0Response = (userInput: string, step: InterviewStep) => {
    switch (step) {
      case "initial":
        setProjectData((prev) => ({ ...prev, type: userInput }))
        return {
          message:
            "Excelente. Ahora necesito entender mejor los requerimientos específicos de tu proyecto. ¿Podrías describir las funcionalidades principales que necesitas?",
          nextStep: "project_details" as InterviewStep,
          mood: "smiling",
        }

      case "project_details":
        setProjectData((prev) => ({ ...prev, requirements: userInput }))
        return {
          message:
            "Perfecto. Para completar mi análisis, ¿cuál es tu timeline esperado y presupuesto aproximado para este proyecto?",
          nextStep: "requirements" as InterviewStep,
          mood: "thinking",
        }

      case "requirements":
        setProjectData((prev) => ({ ...prev, timeline: userInput }))
        return {
          message:
            "Basándome en toda la información que me has proporcionado, he generado una propuesta completa con la configuración de agentes especializados para tu proyecto. La propuesta incluye:\n\n• Análisis detallado de requerimientos\n• Arquitectura técnica recomendada\n• Timeline de desarrollo\n• Equipo de agentes especializados\n• Estimación de costos\n\n¿Aceptas esta propuesta para proceder con la configuración del equipo?",
          nextStep: "proposal" as InterviewStep,
          mood: "happy",
        }

      case "proposal":
        const isAccepted =
          userInput.toLowerCase().includes("sí") ||
          userInput.toLowerCase().includes("si") ||
          userInput.toLowerCase().includes("acepto") ||
          userInput.toLowerCase().includes("aceptar")

        if (isAccepted) {
          return {
            message:
              "¡Excelente! Tu propuesta ha sido aceptada. Estoy generando el archivo YAML con la configuración de los agentes especializados para tu proyecto. El Squad Builder Agent se encargará de configurar tu equipo de desarrollo.\n\nTu UserID, ProposalID y SessionID han sido almacenados en la base de datos. ¡Pronto recibirás más detalles sobre tu equipo de agentes!",
            nextStep: "accepted" as InterviewStep,
            mood: "celebrating",
          }
        } else {
          return {
            message:
              "Entiendo que la propuesta no cumple completamente con tus expectativas. Permíteme ajustar los parámetros. ¿Podrías indicarme qué aspectos te gustaría modificar?",
            nextStep: "rejected" as InterviewStep,
            mood: "worried",
          }
        }

      case "rejected":
        return {
          message:
            "Gracias por el feedback. He ajustado la propuesta según tus comentarios. ¿Te gustaría revisar la nueva propuesta?",
          nextStep: "proposal" as InterviewStep,
          mood: "smiling",
        }

      case "accepted":
        return {
          message: "Tu proyecto está en proceso. ¿Hay algo más en lo que pueda ayudarte?",
          nextStep: "accepted" as InterviewStep,
          mood: "happy",
        }

      default:
        return {
          message: "¿En qué más puedo ayudarte con tu proyecto?",
          nextStep: step,
          mood: "happy",
        }
    }
  }

  const getStepIndicator = () => {
    const steps = ["initial", "project_details", "requirements", "proposal", "decision"]
    const currentIndex = steps.indexOf(currentStep)
    return currentIndex >= 0 ? currentIndex + 1 : 1
  }

  const getStepIcon = () => {
    switch (currentStep) {
      case "initial":
        return <Brain className="h-4 w-4 text-teal-600" />
      case "project_details":
        return <Sparkles className="h-4 w-4 text-teal-600" />
      case "requirements":
        return <AlertCircle className="h-4 w-4 text-teal-600" />
      case "proposal":
        return <CheckCircle className="h-4 w-4 text-teal-600" />
      default:
        return <Circle className="h-4 w-4 text-teal-600" />
    }
  }

  const getStepDescription = () => {
    switch (currentStep) {
      case "initial":
        return "Conociendo tu proyecto"
      case "project_details":
        return "Analizando requerimientos"
      case "requirements":
        return "Definiendo alcance"
      case "proposal":
        return "Generando propuesta"
      case "accepted":
        return "¡Proyecto aprobado!"
      default:
        return "En progreso"
    }
  }

  return (
    <div className="flex flex-col h-screen max-w-5xl mx-auto bg-gradient-to-br from-slate-50 to-gray-100">
      {/* Header */}
      <div className="bg-white/80 backdrop-blur-sm border-b border-gray-200/50 px-3 py-3 shadow-sm">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-4">
              <img src="/images/exltk-logo.svg" alt="EXLTK" className="h-12 w-auto" style={{maxWidth: '180px'}} />
            </div>
          </div>
          <div className="flex items-center gap-3">
            <div className={cn(
              "flex items-center gap-2 px-3 py-1.5 rounded-full border",
              isLoading ? "bg-green-50 border-green-200" : "bg-red-50 border-red-200"
            )}>
              <Circle className={cn(
                "h-2.5 w-2.5 animate-pulse",
                isLoading ? "fill-green-500 text-green-500" : "fill-red-500 text-red-500"
              )} />
              <span className={cn(
                "text-sm font-medium",
                isLoading ? "text-green-700" : "text-red-700"
              )}>
                {isLoading ? "Agente conectado" : "Modo simulado"}
              </span>
            </div>
          </div>
        </div>

        {/* Enhanced Progress indicator */}
        {/* Indicador de progreso por slots o progreso tradicional */}
        {isCollectingSlots ? (
          <ProgressIndicator
            currentSlot={currentSlot}
            totalSlots={9}
            completedSlots={completedSlots}
            className="mt-4"
          />
        ) : (
          <div className="mt-4 bg-gray-50/50 rounded-lg p-4 border border-gray-200/50">
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-2">
                {getStepIcon()}
                <span className="text-sm font-medium text-gray-700">{getStepDescription()}</span>
              </div>
              <span className="text-xs text-gray-500 font-medium">Paso {getStepIndicator()} de 4</span>
            </div>
            <div className="flex-1 bg-gray-200 rounded-full h-2 overflow-hidden">
              <div
                className="bg-gradient-to-r from-teal-500 to-teal-600 h-2 rounded-full transition-all duration-500 ease-out shadow-sm"
                style={{ width: `${(getStepIndicator() / 4) * 100}%` }}
              />
            </div>
          </div>
        )}
      </div>

      {/* Chat Area */}
      <div className="flex-1 overflow-y-auto px-6 py-4 space-y-4">
        {messages.map((message, index) => (
          <div
            key={message.id}
            className={cn(
              "flex gap-4 max-w-[85%] animate-in slide-in-from-bottom-2 duration-300",
              message.role === "user" ? "ml-auto flex-row-reverse" : "",
            )}
            style={{ animationDelay: `${index * 100}ms` }}
          >
            <div className="relative shrink-0">
              <div
                className={cn(
                  "h-15 w-15 rounded-full flex items-center justify-center shadow-md transition-all duration-200 hover:scale-105",
                  message.role === "user"
                    ? "bg-gradient-to-br from-gray-400 to-gray-600"
                    : "bg-gradient-to-br from-teal-100 to-teal-200 border-2 border-teal-300/50",
                )}
              >
                {message.role === "user" ? (
                  <div className="h-8 w-8 rounded-full bg-white/20" />
                ) : (
                  <img
                    src="/images/r0d0-avatar.png"
                    alt="R0D0"
                    className="h-14 w-14 rounded-full transition-all duration-200"
                    style={{maxWidth: '56px', maxHeight: '56px'}}
                  />
                )}
              </div>
              {message.role === "assistant" && (
                <div className="absolute -bottom-1 -right-1 h-3 w-3 bg-teal-500 rounded-full border border-white shadow-sm" />
              )}
            </div>
            <div
              className={cn(
                "rounded-2xl px-5 py-4 text-sm leading-relaxed shadow-sm transition-all duration-200 hover:shadow-md",
                message.role === "user"
                  ? "bg-gradient-to-br from-teal-500 to-teal-600 text-white shadow-teal-500/20"
                  : "bg-white border border-gray-200/50 text-gray-800 shadow-gray-500/10",
              )}
            >
              {message.content.split("\n").map((line, lineIndex) => (
                <div key={lineIndex} className={cn(line.startsWith("•") ? "ml-2" : "")}>
                  {line}
                  {lineIndex < message.content.split("\n").length - 1 && <br />}
                </div>
              ))}
            </div>
          </div>
        ))}

        {isLoading && (
          <div className="flex gap-4 max-w-[85%] animate-in slide-in-from-bottom-2 duration-300">
            <div className="relative shrink-0">
              <div className="h-14 w-14 rounded-full bg-gradient-to-br from-teal-100 to-teal-200 border-2 border-teal-300/50 flex items-center justify-center shadow-md">
                <img src="/images/r0d0-avatar.png" alt="R0D0" className="h-12 w-12 rounded-full animate-pulse" />
              </div>
              <div className="absolute -bottom-1 -right-1 h-3 w-3 bg-teal-500 rounded-full border border-white shadow-sm animate-pulse" />
            </div>
            <div className="bg-white border border-gray-200/50 shadow-sm rounded-2xl px-5 py-4">
              <div className="flex items-center gap-3">
                <div className="flex gap-1">
                  <div className="w-2 h-2 bg-teal-500 rounded-full animate-bounce" />
                  <div className="w-2 h-2 bg-teal-500 rounded-full animate-bounce" style={{ animationDelay: "0.1s" }} />
                  <div className="w-2 h-2 bg-teal-500 rounded-full animate-bounce" style={{ animationDelay: "0.2s" }} />
                </div>
                <span className="text-sm text-gray-600 font-medium">R0D0 está analizando...</span>
                <Brain className="h-4 w-4 text-teal-500 animate-pulse" />
              </div>
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Enhanced Input Area */}
      <div className="bg-white/80 backdrop-blur-sm border-t border-gray-200/50 px-6 py-5 shadow-lg">
        <form
          onSubmit={(e) => {
            e.preventDefault()
            handleSend()
          }}
          className="flex gap-4"
        >
          <div className="flex-1 relative">
            <Input
              placeholder="Describe tu proyecto..."
              value={input}
              onChange={(e) => setInput(e.target.value)}
              className="rounded-full border-gray-300 px-6 h-12 text-base focus:border-teal-500 focus:ring-teal-500/20 focus:ring-4 transition-all duration-200 shadow-sm bg-white/50 backdrop-blur-sm"
              disabled={isLoading}
            />
            <div className="absolute right-4 top-1/2 transform -translate-y-1/2">
              <Sparkles className="h-4 w-4 text-gray-400" />
            </div>
          </div>
          <Button
            type="submit"
            size="icon"
            disabled={isLoading || !input.trim()}
            className="rounded-full bg-gradient-to-r from-teal-500 to-teal-600 hover:from-teal-600 hover:to-teal-700 h-12 w-12 shadow-lg hover:shadow-xl transition-all duration-200 hover:scale-105 disabled:opacity-50 disabled:hover:scale-100"
          >
            <Send className="h-5 w-5" />
            <span className="sr-only">Enviar mensaje</span>
          </Button>
        </form>
      </div>

      {/* Enhanced Footer */}
      <div className="bg-gradient-to-r from-gray-50 to-gray-100 px-6 py-4 text-center border-t border-gray-200/50">
        <div className="flex items-center justify-center gap-2">
          <div className="h-6 w-6 rounded bg-gradient-to-br from-teal-500 to-teal-600 flex items-center justify-center">
            <span className="text-white text-xs font-bold">R</span>
          </div>
          <p className="text-xs text-gray-500 font-medium">
            R0D0 • Powered by <span className="text-teal-600 font-semibold">ADK A2A Architecture</span>
          </p>
        </div>
      </div>
    </div>
  )
}
