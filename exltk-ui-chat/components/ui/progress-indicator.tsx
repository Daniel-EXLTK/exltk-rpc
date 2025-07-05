"use client"

import { CheckCircle, Circle, Clock } from "lucide-react"
import { cn } from "@/lib/utils"

interface ProgressIndicatorProps {
  currentSlot: number
  totalSlots: number
  completedSlots: string[]
  className?: string
}

const SLOT_NAMES = [
  "Título del proyecto",
  "Objetivo principal", 
  "Público objetivo",
  "Problema a resolver",
  "Resultados deseados",
  "Alcance del proyecto",
  "Cronograma",
  "Recursos necesarios",
  "Indicadores de éxito"
]

export function ProgressIndicator({ 
  currentSlot, 
  totalSlots, 
  completedSlots, 
  className 
}: ProgressIndicatorProps) {
  const progressPercentage = (completedSlots.length / totalSlots) * 100

  return (
    <div className={cn("bg-white/80 backdrop-blur-sm rounded-lg p-4 border border-gray-200/50", className)}>
      {/* Header con progreso general */}
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <Clock className="h-4 w-4 text-teal-600" />
          <span className="text-sm font-medium text-gray-700">
            Recopilando información
          </span>
        </div>
        <span className="text-xs text-gray-500 font-medium">
          {completedSlots.length} de {totalSlots} completados
        </span>
      </div>

      {/* Barra de progreso */}
      <div className="mb-4">
        <div className="flex-1 bg-gray-200 rounded-full h-2 overflow-hidden">
          <div 
            className="bg-gradient-to-r from-teal-500 to-teal-600 h-2 rounded-full transition-all duration-500 ease-out"
            style={{ width: `${progressPercentage}%` }}
          />
        </div>
      </div>

      {/* Lista de slots */}
      <div className="space-y-2">
        {SLOT_NAMES.map((slotName, index) => {
          const slotKey = getSlotKey(index)
          const isCompleted = completedSlots.includes(slotKey)
          const isCurrent = index + 1 === currentSlot
          const isPending = index + 1 > currentSlot

          return (
            <div
              key={slotKey}
              className={cn(
                "flex items-center gap-3 p-2 rounded-md transition-all duration-200",
                {
                  "bg-teal-50 border border-teal-200": isCurrent,
                  "bg-green-50": isCompleted,
                  "opacity-50": isPending
                }
              )}
            >
              {/* Icono de estado */}
              <div className="shrink-0">
                {isCompleted ? (
                  <CheckCircle className="h-4 w-4 text-green-600" />
                ) : isCurrent ? (
                  <div className="h-4 w-4 rounded-full border-2 border-teal-500 bg-teal-100 animate-pulse" />
                ) : (
                  <Circle className="h-4 w-4 text-gray-400" />
                )}
              </div>

              {/* Nombre del slot */}
              <span
                className={cn(
                  "text-sm transition-colors",
                  {
                    "text-teal-700 font-medium": isCurrent,
                    "text-green-700": isCompleted,
                    "text-gray-500": isPending,
                    "text-gray-700": !isCurrent && !isCompleted && !isPending
                  }
                )}
              >
                {slotName}
              </span>

              {/* Número del slot */}
              <span
                className={cn(
                  "ml-auto text-xs px-2 py-1 rounded-full",
                  {
                    "bg-teal-100 text-teal-700": isCurrent,
                    "bg-green-100 text-green-700": isCompleted,
                    "bg-gray-100 text-gray-500": isPending
                  }
                )}
              >
                {index + 1}
              </span>
            </div>
          )
        })}
      </div>

      {/* Footer con tiempo estimado */}
      <div className="mt-4 pt-3 border-t border-gray-200/50">
        <div className="flex items-center justify-between text-xs text-gray-500">
          <span>Tiempo estimado restante</span>
          <span className="font-medium">
            {Math.max(0, (totalSlots - completedSlots.length) * 1)} min
          </span>
        </div>
      </div>
    </div>
  )
}

// Función helper para mapear índices a claves de slots
function getSlotKey(index: number): string {
  const slotKeys = [
    "proyecto.titulo",
    "proyecto.objetivo", 
    "proyecto.publico",
    "problema",
    "resultados_deseados",
    "alcance",
    "cronograma",
    "recursos",
    "indicadores_exito"
  ]
  return slotKeys[index] || `slot_${index}`
} 