/**
 * URL 拼接工具
 * 统一处理后端返回的相对路径（如 /static/xxx.jpg）拼接为完整 URL
 * 消除 App.vue、HomeView、ProfileView、SettingsView、VideoDetailView 中重复定义且硬编码 localhost:8080 的问题
 */

// 从环境变量读取后端 baseURL，默认为开发环境地址
const API_BASE = import.meta.env.VITE_API_BASE || 'http://localhost:8080'

/**
 * 将后端返回的相对路径拼接为完整 URL
 * 如果已经是完整 URL（以 http 开头）则直接返回
 */
export function getFullUrl(path: string): string {
  if (!path) return ''
  if (path.startsWith('http')) return path
  return `${API_BASE}${path}`
}

/**
 * 获取完整头像 URL（与 getFullUrl 相同，语义化别名）
 */
export function getFullAvatarUrl(path: string): string {
  return getFullUrl(path)
}
