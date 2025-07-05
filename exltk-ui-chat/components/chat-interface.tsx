"use client"

import { useState, useRef, useEffect } from "react"
import { Send, Bot, User, Loader2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"

type Message = {
  id: string
  content: string
  role: "user" | "assistant"
}

export function ChatInterface() {
  const [messages, setMessages] = useState<Message[]>([
    {
      id: "1",
      content: "Hola, soy R0D0. ¿En qué puedo ayudarte hoy?",
      role: "assistant",
    },
  ])
  const [input, setInput] = useState("")
  const [isLoading, setIsLoading] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  const handleSend = async () => {
    if (!input.trim()) return

    // Add user message
    const userMessage: Message = {
      id: Date.now().toString(),
      content: input,
      role: "user",
    }

    setMessages((prev) => [...prev, userMessage])
    setInput("")
    setIsLoading(true)

    // Simulate R0D0 response (replace with actual API call)
    setTimeout(() => {
      const botResponse: Message = {
        id: (Date.now() + 1).toString(),
        content: generateResponse(input),
        role: "assistant",
      }
      setMessages((prev) => [...prev, botResponse])
      setIsLoading(false)
    }, 1500)
  }

  const generateResponse = (userInput: string): string => {
    // Simple response generation logic (replace with actual AI integration)
    const input = userInput.toLowerCase()

    if (input.includes("hola") || input.includes("saludos")) {
      return "¡Hola! ¿En qué puedo ayudarte hoy?"
    } else if (input.includes("nombre")) {
      return "Mi nombre es R0D0, soy un asistente virtual."
    } else if (input.includes("gracias")) {
      return "¡De nada! Estoy aquí para ayudar."
    } else if (input.includes("ayuda")) {
      return "Puedo responder preguntas, proporcionar información o simplemente conversar. ¿Qué necesitas?"
    } else {
      return "Entiendo tu mensaje. ¿Hay algo específico en lo que pueda ayudarte?"
    }
  }

  return (
    <div className="flex flex-col h-[600px] border rounded-lg overflow-hidden bg-white shadow-sm">
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.map((message) => (
          <div
            key={message.id}
            className={cn(
              "flex items-start gap-3 max-w-[80%] rounded-lg p-3",
              message.role === "user" ? "ml-auto bg-zinc-100 text-zinc-900" : "bg-zinc-800 text-white",
            )}
          >
            <div className="h-8 w-8 rounded-full flex items-center justify-center shrink-0">
              {message.role === "user" ? <User className="h-5 w-5" /> : <Bot className="h-5 w-5" />}
            </div>
            <div className="text-sm">{message.content}</div>
          </div>
        ))}
        {isLoading && (
          <div className="flex items-start gap-3 max-w-[80%] rounded-lg p-3 bg-zinc-800 text-white">
            <div className="h-8 w-8 rounded-full flex items-center justify-center shrink-0">
              <Bot className="h-5 w-5" />
            </div>
            <div className="flex items-center gap-2">
              <Loader2 className="h-4 w-4 animate-spin" />
              <span className="text-sm">R0D0 está pensando...</span>
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>
      <div className="border-t p-4 bg-zinc-50">
        <form
          onSubmit={(e) => {
            e.preventDefault()
            handleSend()
          }}
          className="flex gap-2"
        >
          <Input
            placeholder="Escribe un mensaje..."
            value={input}
            onChange={(e) => setInput(e.target.value)}
            className="flex-1"
          />
          <Button type="submit" size="icon" disabled={isLoading}>
            <Send className="h-4 w-4" />
            <span className="sr-only">Enviar mensaje</span>
          </Button>
        </form>
      </div>
    </div>
  )
}
