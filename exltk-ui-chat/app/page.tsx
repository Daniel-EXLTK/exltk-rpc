import R0D0Interface from "@/components/r0d0-interface"

export default function Home() {
  return (
    <main className="min-h-screen bg-white">
      <div className="container mx-auto p-4">
        <h1 className="text-2xl font-bold text-gray-900 mb-4">
          R0D0 - Asistente de Propuestas
        </h1>
        <div className="bg-white rounded-lg shadow-sm border">
          <R0D0Interface />
        </div>
      </div>
    </main>
  )
}
