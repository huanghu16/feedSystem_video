import { postJson, postForm } from './client'
import type { TokenResponse, Account } from './types'

/**
 * 注册
 */
export function register(username: string, password: string) {
  return postJson<Account>('/account/register', { username, password })
}

/**
 * 登录
 */
export function login(username: string, password: string) {
  return postJson<TokenResponse>('/account/login', { username, password })
}

/**
 * 查询用户信息
 */
export function findByID(id: number) {
  return postJson<Account>('/account/findByID', { id })
}

/**
 * 上传/更新头像（合并原 uploadAvatar 和 updateAvatar，两者实现完全相同）
 */
export function uploadAvatar(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return postForm<{ avatar_url: string }>('/account/uploadAvatar', formData)
}

// 向后兼容别名
export const updateAvatar = uploadAvatar

/**
 * 获取用户资料（包含统计信息）
 */
export interface ProfileInfo {
  id: number
  username: string
  avatar_url: string
  bio: string
  fans_count: number
  following_count: number
  video_count: number
  likes_count: number
}

export function getProfile(userId: number) {
  return postJson<ProfileInfo>('/account/getProfile', { user_id: userId })
}

/**
 * 修改密码
 */
export function changePassword(oldPassword: string, newPassword: string) {
  return postJson<{ message: string }>('/account/changePassword', {
    old_password: oldPassword,
    new_password: newPassword
  })
}

/**
 * 更新简介
 */
export function updateBio(bio: string) {
  return postJson<{ message: string }>('/account/updateBio', { bio })
}
