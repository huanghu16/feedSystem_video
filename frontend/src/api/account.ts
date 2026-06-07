import { postJson, postForm } from './client'
import type { TokenResponse, Account } from './types'

// 注册
export function register(username: string, password: string) {
    return postJson<Account>('/account/register', { username, password })
}

// 登录
export function login(username: string, password: string) {
    return postJson<TokenResponse>('/account/login', { username, password })
}

// 查询用户信息
export function findByID(id: number) {
    return postJson<Account>('/account/findByID', { id })
}

// 查询用户主页
export function getProfile(id: number) {
    return postJson<Account>('/account/getProfile', { id })
}

// 上传头像
export function uploadAvatar(file: File) {
    const formData = new FormData()
    formData.append('file', file)
    return postForm<{ avatar_url: string }>('/account/uploadAvatar', formData)
}