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
  // Only set Content-Type for requests with a JSON body (not for GET or FormData)
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
  constructor(
    public status: number,
    public code: string,
    message: string,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

// Auth
// 用户模型
export type User = {
  id: string;
  name: string;
  email: string;
  avatar_url: string;
  provider: string;
  created_at: string;
};

// 认证接口：获取当前用户完整信息、刷新 token
export const auth = {
  me: () => request<User>('/auth/me'),
  refresh: () => request<{ token: string }>('/auth/refresh', { method: 'POST' }),
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
