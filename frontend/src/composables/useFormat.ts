/**
 * 通用格式化工具函数
 * 在 HomeView、HotView、ProfileView、VideoDetailView 等多个组件中复用
 * 消除各组件中重复定义的 formatDate / formatTime / formatNumber 函数
 */

/**
 * 格式化日期为相对时间（如"刚刚"、"3分钟前"、"2天前"）
 * 超过7天则显示具体日期
 */
export function formatDate(dateStr: string): string {
  if (!dateStr) return ''

  const date = new Date(dateStr)
  if (isNaN(date.getTime())) return ''

  const now = new Date()
  const diff = now.getTime() - date.getTime()

  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  if (hours < 24) return `${hours}小时前`
  if (days < 7) return `${days}天前`

  return date.toLocaleDateString('zh-CN')
}

/**
 * 格式化时间为完整日期时间（如 2026/7/2 14:30:00）
 */
export function formatTime(timeStr: string): string {
  if (!timeStr) return ''
  return new Date(timeStr).toLocaleString('zh-CN')
}

/**
 * 格式化数字（过万显示"x.x万"）
 */
export function formatNumber(num: number): string {
  if (num >= 10000) {
    return (num / 10000).toFixed(1) + '万'
  }
  return num.toString()
}

/**
 * 格式化发布日期
 */
export function formatPublishDate(dateStr: string): string {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleDateString('zh-CN')
}
