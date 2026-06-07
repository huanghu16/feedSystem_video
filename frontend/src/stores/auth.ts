import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

// 认证状态：管理 token 和登录状态
export const useAuthStore = defineStore('auth', () => {
    const token = ref(localStorage.getItem('access_token') || '')   // 登录令牌
    const refreshTokenValue = ref(localStorage.getItem('refresh_token') || '')  // 刷新令牌

    const isLoggedIn = computed(() => !!token.value)   // 登录状态

    // JWT payload 解析（纯前端 Base64 解码，不需要后端）
    const claims = computed(() => {
        if (!token.value) return null
        try {
            const payload = token.value.split('.')[1]   // JWT payload
            const decoded = atob(payload)  // Base64 解码
            return JSON.parse(decoded)  // JSON 解码
        } catch {
            return null
        }
    })

    function setTokens(access: string, refresh: string) {
        token.value = access  // 保存 token
        refreshTokenValue.value = refresh  // 保存刷新令牌
        localStorage.setItem('access_token', access)  // 保存 token
        localStorage.setItem('refresh_token', refresh)  // 保存刷新令牌
    }

    function clearTokens() {
        token.value = ''  // 清空 token
        refreshTokenValue.value = ''  // 清空刷新令牌
        localStorage.removeItem('access_token')  // 移除 token
        localStorage.removeItem('refresh_token')  // 移除刷新令牌
    }

    return { token, isLoggedIn, claims, setTokens, clearTokens }
})