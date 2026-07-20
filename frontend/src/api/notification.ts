import { getJson } from './client'

export interface NotificationItem {
  id: number
  actor_name: string
  type: string
  content: string
  is_read: boolean
  created_at: string
}

export interface ListResponse {
  list: NotificationItem[]
  total: number
}

export interface UnreadCountResponse {
  count: number
}

/** 获取未读通知数 */
export function unreadCount() {
  return getJson<UnreadCountResponse>('/notification/unreadCount')
}

/** 获取通知列表（读取后自动标记已读） */
export function listNotifications() {
  return getJson<ListResponse>('/notification/list')
}
