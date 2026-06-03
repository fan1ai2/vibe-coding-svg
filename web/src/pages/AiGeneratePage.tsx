import { useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { ai, icons, ApiError, type IconCandidate, type AiQuotaResponse } from '../api/client'
import PublishDialog from '../components/PublishDialog'
import LoadingSpinner from '../components/LoadingSpinner'

type Phase = 'input' | 'generating' | 'candidates'

export default function AiGeneratePage() {
  const { token } = useAuth()
  const navigate = useNavigate()

  const [phase, setPhase] = useState<Phase>('input')
  const [prompt, setPrompt] = useState('')
  const [style, setStyle] = useState('line')
  const [candidates, setCandidates] = useState<IconCandidate[]>([])
  const [quota, setQuota] = useState<AiQuotaResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [toast, setToast] = useState<string | null>(null)
  const [publishTarget, setPublishTarget] = useState<IconCandidate | null>(null)

  const showToast = useCallback((msg: string) => {
    setToast(msg)
    setTimeout(() => setToast(null), 3000)
  }, [])

  const handleGenerate = useCallback(async () => {
    if (!prompt.trim() || !token) return
    setPhase('generating')
    setError(null)

    try {
      const res = await ai.generate(prompt.trim(), style)
      const data = res.data
      setCandidates(data.candidates)
      setQuota({ remaining: data.remaining_quota, limit: 20 })
      setPhase('candidates')
    } catch (err) {
      setPhase('input')
      if (err instanceof ApiError) {
        if (err.code === 'QUOTA_EXCEEDED') {
          setError('今日配额已用完，请明天再来')
        } else {
          setError(err.message || '生成失败，请重试')
        }
      } else {
        setError('网络错误，请检查连接后重试')
      }
    }
  }, [prompt, style, token])

  // Load quota on first render
  const [quotaLoaded, setQuotaLoaded] = useState(false)
  if (token && !quotaLoaded) {
    setQuotaLoaded(true)
    ai.quota().then(res => setQuota(res.data)).catch(() => {})
  }

  const handleOpenInEditor = useCallback((candidate: IconCandidate) => {
    if (!candidate.svg_content) {
      setError('SVG 内容为空，无法打开编辑器')
      return
    }
    sessionStorage.setItem('editor:pendingSvg', candidate.svg_content)
    navigate('/workspace/editor')
  }, [navigate])

  const handlePublish = useCallback(async (name: string, tags: { name: string; type: string }[], theme: string, isPublic: boolean) => {
    if (!token) {
      setError('登录已过期，请重新登录')
      navigate('/')
      return
    }
    if (!publishTarget) return
    try {
      const res = await icons.create({ name, svg_content: publishTarget.svg_content, is_public: isPublic, tags, theme })
      setPublishTarget(null)
      showToast(`已发布到图标库`)
      setTimeout(() => navigate(`/icons/${res.data.id}`), 1000)
    } catch (err) {
      console.error('Publish failed:', err)
      showToast('发布失败，请重试')
    }
  }, [token, publishTarget, navigate, showToast])

  return (
    <div className="mx-auto max-w-3xl">
      {toast && (
        <div className="fixed top-20 right-4 z-50 px-4 py-2 bg-gray-800 text-white text-sm rounded-lg shadow-lg">
          {toast}
        </div>
      )}
      {error && (
        <div className="fixed top-20 right-4 z-50 px-4 py-2 bg-red-500 text-white text-sm rounded-lg shadow-lg">
          {error}
        </div>
      )}

      <h1 className="text-xl font-bold text-gray-900 mb-6">AI 生成图标</h1>

      {/* Input Phase */}
      {phase === 'input' && (
        <div className="rounded-2xl border border-gray-200 bg-white p-6">
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              描述你想要的图标
            </label>
            <textarea
              value={prompt}
              onChange={e => setPrompt(e.target.value)}
              placeholder="例如：一枚简洁的邮件图标，线条风格，适合用在导航栏"
              rows={3}
              maxLength={200}
              className="w-full px-4 py-3 rounded-xl border border-gray-200 text-sm focus:outline-none focus:ring-2 focus:ring-amber-500/20 resize-none"
            />
            <div className="text-right text-xs text-gray-400 mt-1">{prompt.length}/200</div>
          </div>

          <div className="flex items-center gap-4 mb-4">
            <div className="flex items-center gap-2">
              <span className="text-sm text-gray-600">风格:</span>
              {(['line', 'filled'] as const).map(s => (
                <button
                  key={s}
                  onClick={() => setStyle(s)}
                  className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
                    style === s
                      ? 'bg-amber-500 text-white'
                      : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                  }`}
                >
                  {s === 'line' ? '线条' : '填充'}
                </button>
              ))}
            </div>
          </div>

          {quota && (
            <div className="text-xs text-gray-400 mb-4">
              今日剩余配额: {quota.remaining}/{quota.limit}
            </div>
          )}

          <button
            onClick={handleGenerate}
            disabled={!prompt.trim()}
            className="w-full rounded-xl bg-amber-500 px-6 py-3 text-base font-bold text-white shadow-md shadow-amber-200 transition-all duration-300 hover:-translate-y-0.5 hover:bg-amber-600 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:translate-y-0"
          >
            生成图标
          </button>
        </div>
      )}

      {/* Generating Phase */}
      {phase === 'generating' && (
        <LoadingSpinner label="AI 正在为你设计图标..." />
      )}

      {/* Candidates Phase */}
      {phase === 'candidates' && (
        <>
          <div className="flex items-center justify-between mb-4">
            <p className="text-sm text-gray-500">
              生成了 {candidates.length} 个候选 · 剩余配额 {quota?.remaining ?? '?'}
            </p>
            <button
              onClick={() => setPhase('input')}
              className="text-sm text-amber-600 hover:text-amber-700 font-medium"
            >
              ← 重新生成
            </button>
          </div>

          {candidates.length === 0 ? (
            <div className="text-center py-20">
              <p className="text-gray-500 mb-4">所有候选均未通过质量检查</p>
              <button
                onClick={handleGenerate}
                className="rounded-xl bg-amber-500 px-6 py-2 text-sm font-bold text-white hover:bg-amber-600"
              >
                重试
              </button>
            </div>
          ) : (
            <div className="grid grid-cols-2 gap-4">
              {candidates.map((c, i) => (
                <div
                  key={i}
                  className="rounded-2xl border border-gray-200 bg-white p-4 hover:shadow-md transition-shadow"
                >
                  <div className="h-40 flex items-center justify-center bg-gray-50 rounded-xl mb-3 p-3">
                    <div
                      className="w-full h-full flex items-center justify-center"
                      dangerouslySetInnerHTML={{ __html: c.svg_content }}
                    />
                  </div>
                  <h3 className="text-sm font-semibold text-gray-900 mb-1">{c.name}</h3>
                  {c.tags.length > 0 && (
                    <div className="flex flex-wrap gap-1 mb-3">
                      {c.tags.map(t => (
                        <span key={t} className="px-1.5 py-0.5 rounded text-[10px] bg-amber-50 text-amber-600">
                          {t}
                        </span>
                      ))}
                    </div>
                  )}
                  <div className="flex gap-2">
                    <button
                      onClick={() => handleOpenInEditor(c)}
                      className="flex-1 rounded-lg border border-amber-200 px-2 py-1.5 text-xs font-medium text-amber-700 hover:bg-amber-50 transition-colors"
                    >
                      在编辑器中打开
                    </button>
                    <button
                      onClick={() => setPublishTarget(c)}
                      className="flex-1 rounded-lg bg-amber-500 px-2 py-1.5 text-xs font-medium text-white hover:bg-amber-600 transition-colors"
                    >
                      保存到图标库
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </>
      )}

      {publishTarget && (
        <PublishDialog
          onClose={() => setPublishTarget(null)}
          onPublish={handlePublish}
          defaultName={publishTarget.name}
          defaultTags={publishTarget.tags}
        />
      )}
    </div>
  )
}
