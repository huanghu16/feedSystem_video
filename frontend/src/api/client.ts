// HTTP 请求封装
// 所有 API 调用都通过这里发出，统一处理 token 和错误

const API_BASE = import.meta.env.VITE_API_BASE || '/api'  // API 地址

// 统一响应格式（和后端 apierror.Response 对应）
interface ApiResponse<T = any> {
    code: number    // 错误码
    message: string // 错误信息
    data: T  // 数据
}

// 自定义错误类
class ApiError extends Error {
    status: number    // 状态码
    payload: any  // 错误数据

    constructor(status: number, payload: any) {
        super(payload?.message || '请求失败')
        this.status = status
        this.payload = payload
    }
}

// 是否正在刷新 token
let isRefreshing = false    // 是否正在刷新 token
let refreshPromise: Promise<string> | null = null  // 刷新 token 的 Promise

// 获取 token
function getToken(): string {
    return localStorage.getItem('access_token') || ''
}

// POST JSON 请求
export async function postJson<T = any>(path: string, body?: any): Promise<T> {
    const res = await fetch(`${API_BASE}${path}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            ...(getToken() ? { Authorization: `Bearer ${getToken()}` } : {}),
        },
        body: body ? JSON.stringify(body) : undefined,
    })

    // 检查响应类型是否为 JSON
    const contentType = res.headers.get('content-type')
    if (!contentType || !contentType.includes('application/json')) {
        const text = await res.text()
        console.error('非JSON响应:', text)
        throw new ApiError(res.status, { message: '服务器响应格式错误' })
    }

    const json: ApiResponse<T> = await res.json()

    if (res.status === 401 && path !== '/account/refresh') {
        // token 过期，尝试刷新
        const newToken = await refreshToken()
        if (newToken) {
            // 用新 token 重试原请求
            return postJson<T>(path, body)
        }
    }

    if (json.code !== 0) {
        throw new ApiError(res.status, json)
    }

    return json.data
}

// POST FormData 请求（文件上传用）
export async function postForm<T = any>(path: string, formData: FormData): Promise<T> {
    const res = await fetch(`${API_BASE}${path}`, {
        method: 'POST',
        headers: {
            ...(getToken() ? { Authorization: `Bearer ${getToken()}` } : {}),
        },
        body: formData,
    })

    const json: ApiResponse<T> = await res.json()

    if (json.code !== 0) {
        throw new ApiError(res.status, json)
    }

    return json.data
}

// 刷新 token
async function refreshToken(): Promise<string | null> {
    const refreshTokenValue = localStorage.getItem('refresh_token')
    if (!refreshTokenValue) return null

    if (isRefreshing && refreshPromise) {
        return refreshPromise // 防止并发刷新
    }

    isRefreshing = true
    refreshPromise = (async () => {
        try {
            const res = await fetch(`${API_BASE}/account/refresh`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ refresh_token: refreshTokenValue }),
            })
            const json: ApiResponse<{ access_token: string; expires_in: number }> = await res.json()
            if (json.code === 0 && json.data) {
                localStorage.setItem('access_token', json.data.access_token)
                return json.data.access_token
            }
            return null
        } catch {
            return null
        } finally {
            isRefreshing = false
            refreshPromise = null
        }
    })()

    return refreshPromise
}