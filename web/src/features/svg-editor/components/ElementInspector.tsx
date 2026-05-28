interface ElementInspectorProps {
  element: SVGElement | null
}

export default function ElementInspector({ element }: ElementInspectorProps) {
  if (!element) {
    return (
      <div className="p-4 text-sm text-gray-400 text-center border border-dashed border-gray-200 rounded-lg">
        点击画布中的元素查看属性
      </div>
    )
  }

  const fill = element.getAttribute('fill') || 'none'
  const stroke = element.getAttribute('stroke') || 'none'
  const tagName = element.tagName.toLowerCase()

  return (
    <div className="space-y-2 p-3 bg-gray-50 rounded-lg border border-gray-100">
      <div className="flex items-center gap-2">
        <span className="text-xs font-mono bg-gray-200 px-1.5 py-0.5 rounded">{tagName}</span>
        <span className="text-xs text-gray-400">id={element.id || '—'}</span>
      </div>
      <div className="flex gap-3">
        <div className="flex items-center gap-1.5">
          <div className="w-4 h-4 rounded border border-gray-300" style={{ backgroundColor: fill !== 'none' ? fill : '#fff' }} />
          <span className="text-xs text-gray-500">fill: {fill}</span>
        </div>
        <div className="flex items-center gap-1.5">
          <div className="w-4 h-4 rounded border border-gray-300 ring-1 ring-inset ring-gray-300" style={{ backgroundColor: stroke !== 'none' ? stroke : 'transparent' }} />
          <span className="text-xs text-gray-500">stroke: {stroke}</span>
        </div>
      </div>
    </div>
  )
}
