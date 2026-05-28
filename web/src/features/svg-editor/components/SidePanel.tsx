import { ReactNode } from 'react'

interface SidePanelProps {
  children: ReactNode
}

export default function SidePanel({ children }: SidePanelProps) {
  return (
    <div className="w-72 flex-shrink-0 space-y-4 p-4 bg-gray-50/50 rounded-xl border border-gray-100 overflow-y-auto max-h-[calc(100vh-8rem)]">
      {children}
    </div>
  )
}
