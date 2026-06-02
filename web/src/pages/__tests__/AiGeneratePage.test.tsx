import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import AiGeneratePage from '../../pages/AiGeneratePage'

// Mock useAuth
const mockUseAuth = vi.fn()
vi.mock('../../context/AuthContext', () => ({
  useAuth: () => mockUseAuth(),
}))

// Mock API client
vi.mock('../../api/client', () => ({
  ai: {
    generate: vi.fn(),
    quota: vi.fn(),
  },
  icons: {
    create: vi.fn(),
  },
  ApiError: class extends Error {
    status: number
    code: string
    constructor(status: number, code: string, message: string) {
      super(message)
      this.status = status
      this.code = code
    }
  },
}))

import { ai, ApiError } from '../../api/client'

// Mock navigate
const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return { ...actual, useNavigate: () => mockNavigate }
})

function renderPage() {
  return render(
    <MemoryRouter>
      <AiGeneratePage />
    </MemoryRouter>
  )
}

describe('AiGeneratePage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockUseAuth.mockReturnValue({ token: 'fake-token', loading: false })
    vi.mocked(ai.quota).mockResolvedValue({
      data: { remaining: 18, limit: 20 },
    } as any)
  })

  it('renders the input phase with title', () => {
    renderPage()
    expect(screen.getByText('AI 生成图标')).toBeInTheDocument()
    expect(screen.getByPlaceholderText(/邮件图标/)).toBeInTheDocument()
  })

  it('disables generate button when prompt is empty', () => {
    renderPage()
    const btn = screen.getByRole('button', { name: '生成图标' })
    expect(btn).toBeDisabled()
  })

  it('enables generate button when prompt has text', async () => {
    renderPage()
    const textarea = screen.getByPlaceholderText(/邮件图标/)
    await userEvent.type(textarea, 'a cat icon')
    const btn = screen.getByRole('button', { name: '生成图标' })
    expect(btn).not.toBeDisabled()
  })

  it('shows character count', async () => {
    renderPage()
    const textarea = screen.getByPlaceholderText(/邮件图标/)
    await userEvent.type(textarea, 'abc')
    expect(screen.getByText('3/200')).toBeInTheDocument()
  })

  it('switches to generating phase on submit', async () => {
    // Make generate never resolve so we stay in generating phase
    vi.mocked(ai.generate).mockImplementation(
      () => new Promise(() => {})
    )
    renderPage()
    await userEvent.type(screen.getByPlaceholderText(/邮件图标/), 'a cat icon')
    await userEvent.click(screen.getByRole('button', { name: '生成图标' }))
    await waitFor(() => {
      expect(screen.getByText('AI 正在为你设计图标...')).toBeInTheDocument()
    })
  })

  it('shows error when quota exceeded', async () => {
    vi.mocked(ai.generate).mockRejectedValue(
      new ApiError(429, 'QUOTA_EXCEEDED', '今日配额已用完，请明天再来')
    )
    renderPage()
    await userEvent.type(screen.getByPlaceholderText(/邮件图标/), 'a cat icon')
    await userEvent.click(screen.getByRole('button', { name: '生成图标' }))
    await waitFor(() => {
      expect(screen.getByText('今日配额已用完，请明天再来')).toBeInTheDocument()
    })
  })

  it('shows error on network failure', async () => {
    vi.mocked(ai.generate).mockRejectedValue(new Error('Network error'))
    renderPage()
    await userEvent.type(screen.getByPlaceholderText(/邮件图标/), 'a cat icon')
    await userEvent.click(screen.getByRole('button', { name: '生成图标' }))
    await waitFor(() => {
      expect(screen.getByText('网络错误，请检查连接后重试')).toBeInTheDocument()
    })
  })

  it('displays candidates after successful generation', async () => {
    vi.mocked(ai.generate).mockResolvedValue({
      data: {
        candidates: [
          {
            name: 'Cat Icon',
            svg_content: '<svg viewBox="0 0 24 24"><path d="M12 2L2 22h20L12 2z"/></svg>',
            tags: ['animal', 'pet'],
          },
          {
            name: 'Dog Icon',
            svg_content: '<svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/></svg>',
            tags: ['animal'],
          },
        ],
        remaining_quota: 17,
      },
    } as any)
    renderPage()
    await userEvent.type(screen.getByPlaceholderText(/邮件图标/), 'a cat icon')
    await userEvent.click(screen.getByRole('button', { name: '生成图标' }))
    await waitFor(() => {
      expect(screen.getByText('Cat Icon')).toBeInTheDocument()
      expect(screen.getByText('Dog Icon')).toBeInTheDocument()
    })
  })

  it('shows quota remaining after generation', async () => {
    vi.mocked(ai.generate).mockResolvedValue({
      data: {
        candidates: [
          {
            name: 'Icon',
            svg_content: '<svg viewBox="0 0 24 24"><path d="M12 2L2 22h20L12 2z"/></svg>',
            tags: [],
          },
        ],
        remaining_quota: 17,
      },
    } as any)
    renderPage()
    await userEvent.type(screen.getByPlaceholderText(/邮件图标/), 'a cat icon')
    await userEvent.click(screen.getByRole('button', { name: '生成图标' }))
    await waitFor(() => {
      expect(screen.getByText(/剩余配额 17/)).toBeInTheDocument()
    })
  })

  it('style toggle defaults to line', () => {
    renderPage()
    const lineBtn = screen.getByRole('button', { name: '线条' })
    const filledBtn = screen.getByRole('button', { name: '填充' })
    expect(lineBtn.className).toContain('bg-amber-500')
    expect(filledBtn.className).not.toContain('bg-amber-500')
  })

  it('allows switching to filled style', async () => {
    renderPage()
    await userEvent.click(screen.getByRole('button', { name: '填充' }))
    const filledBtn = screen.getByRole('button', { name: '填充' })
    expect(filledBtn.className).toContain('bg-amber-500')
  })

  it('renders title with correct classes', () => {
    renderPage()
    const h1 = screen.getByRole('heading', { level: 1 })
    expect(h1.className).toContain('text-xl')
    expect(h1.className).toContain('font-bold')
  })
})
