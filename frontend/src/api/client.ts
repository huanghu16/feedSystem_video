import { useAuthStore } from '../stores/auth'

// 后端 API 基础地址，支持环境变量覆盖
const baseURL = import.meta.env.VITE_API_BASE || '/api'

// 默认请求超时时间（毫秒）
const DEFAULT_TIMEOUT = 15000

// 分片上传超时（单次只传 5MB）
const UPLOAD_TIMEOUT = 60000

// 统一响应格式
export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

// 自定义 API 错误
export class ApiError extends Error {
  status: number
  payload: any

  constructor(status: number, payload: any) {
    super(payload?.message || `HTTP ${status}`)
    this.status = status
    this.payload = payload
  }
}

// ===== Token 刷新机制（防止并发刷新）=====
let isRefreshing = false
let refreshPromise: Promise<string> | null = null

async function refreshToken(): Promise<string> {
  const auth = useAuthStore()
  if (!auth.refreshTokenValue) {
    throw new Error('No refresh token')
  }

  // 如果正在刷新，复用同一个 Promise
  if (isRefreshing && refreshPromise) {
    return refreshPromise
  }

  isRefreshing = true
  refreshPromise = (async () => {
    try {
      const res = await fetch(`${baseURL}/account/refreshToken`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${auth.refreshTokenValue}`,
        },
        signal: AbortSignal.timeout(DEFAULT_TIMEOUT),
      })
      if (!res.ok) throw new Error('Refresh failed')
      const data: ApiResponse<{ access_token: string; refresh_token: string }> = await res.json()
      if (data.code !== 0) throw new Error(data.message)
      auth.setTokens(data.data.access_token, data.data.refresh_token)
      return data.data.access_token
    } finally {
      isRefreshing = false
      refreshPromise = null
    }
  })()

  return refreshPromise
}

// ===== 核心 HTTP 方法 =====

/**
 * 发送 GET 请求（查询类接口使用）
 */
export async function getJson<T>(path: string, params?: Record<string, any>): Promise<T> {
  const url = params
    ? `${baseURL}${path}?${new URLSearchParams(
        Object.entries(params).reduce((acc, [k, v]) => {
          acc[k] = String(v)
          return acc
        }, {} as Record<string, string>)
      )}`
    : `${baseURL}${path}`

  return doFetch<T>(url, {
    method: 'GET',
    headers: buildHeaders(),
  })
}

/**
 * 发送 POST JSON 请求
 */
export async function postJson<T>(path: string, body?: any): Promise<T> {
  const url = `${baseURL}${path}`
  return doFetch<T>(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...buildAuthHeader(),
    },
    body: body ? JSON.stringify(body) : undefined,
  })
}

/**
 * 发送 POST FormData 请求（文件上传）
 */
export async function postForm<T>(path: string, formData: FormData): Promise<T> {
  const url = `${baseURL}${path}`
  return doFetch<T>(url, {
    method: 'POST',
    headers: buildAuthHeader(), // FormData 不设 Content-Type，浏览器自动添加 boundary
    body: formData,
  })
}

/**
 * 发送 POST FormData 请求（长超时，用于分片上传）
 * 单次只传 5MB，但弱网下 15 秒不够，需要单独放宽
 */
export async function postFormLong<T>(
  path: string,
  formData: FormData,
  timeoutMs = UPLOAD_TIMEOUT
): Promise<T> {
  const url = `${baseURL}${path}`
  return doFetch<T>(
    url,
    {
      method: 'POST',
      headers: buildAuthHeader(),
      body: formData,
    },
    timeoutMs
  )
}

// ===== 内部辅助函数 =====

function buildAuthHeader(): Record<string, string> {
  const auth = useAuthStore()
  const headers: Record<string, string> = {}
  if (auth.token) {
    headers['Authorization'] = `Bearer ${auth.token}`
  }
  return headers
}

function buildHeaders(): Record<string, string> {
  return {
    'Content-Type': 'application/json',
    ...buildAuthHeader(),
  }
}

/**
 * 执行 fetch 请求，统一处理超时、401 刷新、错误解析
 */
async function doFetch<T>(url: string, init: RequestInit, timeoutMs = DEFAULT_TIMEOUT): Promise<T> {
  // 添加超时控制
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), timeoutMs)

  try {
    let res = await fetch(url, { ...init, signal: controller.signal })

    // 401 自动刷新重试
    if (res.status === 401) {
      try {
        const newToken = await refreshToken()
        // 用新 token 重试原请求
        const headers = new Headers(init.headers)
        headers.set('Authorization', `Bearer ${newToken}`)
        res = await fetch(url, { ...init, headers, signal: controller.signal })
      } catch {
        // 刷新失败，清除 token
        const auth = useAuthStore()
        auth.clearTokens()
        throw new ApiError(401, { message: '登录已过期，请重新登录' })
      }
    }

    // 检查 Content-Type 是否为 JSON
    const contentType = res.headers.get('content-type') || ''
    if (!contentType.includes('application/json')) {
      throw new ApiError(res.status, { message: `服务器返回非 JSON 响应 (${res.status})` })
    }

    const data: ApiResponse<T> = await res.json()

    if (data.code !== 0) {
      throw new ApiError(res.status, data)
    }

    return data.data
  } catch (err: any) {
    // AbortError 即超时
    if (err.name === 'AbortError') {
      throw new ApiError(0, { message: '请求超时，请检查网络' })
    }
    // 已经是 ApiError 直接抛出
    if (err instanceof ApiError) throw err
    // 其他网络错误
    throw new ApiError(0, { message: err.message || '网络错误' })
  } finally {
    clearTimeout(timeoutId)
  }
}
