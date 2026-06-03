const BASE = '/api/v1';

function token(): string | null {
  return localStorage.getItem('token');
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {
    ...(options.headers as Record<string, string> ?? {}),
  };
  const tok = token();
  if (tok) {
    headers['Authorization'] = `Bearer ${tok}`;
  }
  // 仅在请求带有 JSON body 时设置 Content-Type（GET 或 FormData 不设置）
  if (options.body && !(options.body instanceof FormData)) {
    headers['Content-Type'] = 'application/json';
  }

  const res = await fetch(`${BASE}${path}`, { ...options, headers });
  const data = await res.json();
  if (!res.ok) {
    throw new ApiError(res.status, data?.error?.code ?? 'UNKNOWN', data?.error?.message ?? 'Request failed');
  }
  return data as T;
}

export class ApiError extends Error {
  public status: number
  public code: string
  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
  }
}

// 认证接口
export type User = {
  id: string;
  email: string;
  name: string;
  avatar_url: string;
  provider: string;
  provider_id: string;
  created_at: string;
  updated_at: string;
};

export const auth = {
  me: () => request<User>('/auth/me'),
  refresh: () => request<{ token: string }>('/auth/refresh', { method: 'POST' }),
  guest: () => request<{ token: string; user: User }>('/auth/guest', { method: 'POST' }),
  sendCode: (email: string) =>
    request<{ ok: boolean }>('/auth/email/send-code', {
      method: 'POST',
      body: JSON.stringify({ email }),
    }),
  verifyCode: (email: string, code: string) =>
    request<{ token: string; user: User }>('/auth/email/verify', {
      method: 'POST',
      body: JSON.stringify({ email, code }),
    }),
};

// 转换任务模型
export type Conversion = {
  id: string;
  user_id: string;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  original_url: string;
  svg_url: string | null;
  thumbnail_url: string | null;
  file_size_in: number;
  file_size_out: number;
  path_count: number;
  color_count: number;
  format_in: string;
  error_message: string;
  created_at: string;
  completed_at: string | null;
};

export type ConversionListResponse = {
  data: Conversion[];
};

export type ConversionSingleResponse = {
  data: Conversion;
};

export const conversions = {
  upload: (file: File) => {
    const form = new FormData();
    form.append('file', file);
    return request<ConversionSingleResponse>('/conversions', { method: 'POST', body: form });
  },
  list: (limit = 20, offset = 0) =>
    request<ConversionListResponse>(`/conversions?limit=${limit}&offset=${offset}`),
  get: (id: string) =>
    request<ConversionSingleResponse>(`/conversions/${id}`),
  downloadUrl: (id: string) => `${BASE}/conversions/${id}/download`,
};

// 保存的 SVG
export type SavedSvg = {
  id: string;
  user_id: string;
  name: string;
  svg_content?: string;
  created_at: string;
  updated_at: string;
};

export type SavedSvgListResponse = {
  data: SavedSvg[];
};

export type SavedSvgSingleResponse = {
  data: SavedSvg;
};

export const svgs = {
  save: (name: string, svgContent: string) =>
    request<SavedSvgSingleResponse>('/svgs', {
      method: 'POST',
      body: JSON.stringify({ name, svg_content: svgContent }),
    }),
  list: (limit = 20, offset = 0) =>
    request<SavedSvgListResponse>(`/svgs?limit=${limit}&offset=${offset}`),
  get: (id: string) =>
    request<SavedSvgSingleResponse>(`/svgs/${id}`),
  downloadUrl: (id: string) => `${BASE}/svgs/${id}/download`,
  delete: (id: string) =>
    request<{ ok: boolean }>(`/svgs/${id}`, { method: 'DELETE' }),
};

// 图标库
export type IconTag = {
  id: string;
  name: string;
  slug: string;
  type: 'usage' | 'style' | 'category';
  usage_count: number;
};

export type Icon = {
  id: string;
  user_id: string;
  name: string;
  svg_content?: string;
  is_public: boolean;
  download_count: number;
  created_at: string;
  updated_at: string;
  tags?: IconTag[];
  colors?: string[];
  theme?: string;
};

export type IconListResponse = { data: Icon[] };
export type IconSingleResponse = { data: Icon };
export type TagListResponse = { data: IconTag[] };

export type IconSearchParams = {
  q?: string;
  tags?: string;
  color?: string;
  theme?: string;
  sort?: 'popular' | 'newest';
  limit?: number;
  offset?: number;
};

export type CreateIconInput = {
  name: string;
  svg_content: string;
  is_public?: boolean;
  tags?: { name: string; type: string }[];
  theme?: string;
};

export type BatchIconInput = {
  icons: CreateIconInput[];
};

export const icons = {
  create: (input: CreateIconInput) =>
    request<IconSingleResponse>('/icons', {
      method: 'POST',
      body: JSON.stringify(input),
    }),
  batchCreate: (input: BatchIconInput) =>
    request<IconListResponse>('/icons/batch', {
      method: 'POST',
      body: JSON.stringify(input),
    }),
  list: (limit = 20, offset = 0) =>
    request<IconListResponse>(`/icons?limit=${limit}&offset=${offset}`),
  get: (id: string) =>
    request<IconSingleResponse>(`/icons/${id}`),
  search: (params: IconSearchParams) => {
    const qs = new URLSearchParams();
    if (params.q) qs.set('q', params.q);
    if (params.tags) qs.set('tags', params.tags);
    if (params.color) qs.set('color', params.color);
    if (params.theme) qs.set('theme', params.theme);
    if (params.sort) qs.set('sort', params.sort);
    if (params.limit) qs.set('limit', String(params.limit));
    if (params.offset) qs.set('offset', String(params.offset));
    return request<IconListResponse>(`/icons/search?${qs.toString()}`);
  },
  recommend: (id: string, limit = 10) =>
    request<IconListResponse>(`/icons/${id}/recommend?limit=${limit}`),
  delete: (id: string) =>
    request<{ ok: boolean }>(`/icons/${id}`, { method: 'DELETE' }),
};

export const tags = {
  list: (limit = 50, sort = 'popular') =>
    request<TagListResponse>(`/tags?limit=${limit}&sort=${sort}`),
};

// AI 图标生成
export type IconCandidate = {
  name: string;
  svg_content: string;
  tags: string[];
};

export type AiGenerateResponse = {
  candidates: IconCandidate[];
  remaining_quota: number;
};

export type AiQuotaResponse = {
  remaining: number;
  limit: number;
};

export const ai = {
  generate: (prompt: string, style: string) =>
    request<{ data: AiGenerateResponse }>('/ai/generate', {
      method: 'POST',
      body: JSON.stringify({ prompt, style }),
    }),
  quota: () =>
    request<{ data: AiQuotaResponse }>('/ai/quota'),
};
