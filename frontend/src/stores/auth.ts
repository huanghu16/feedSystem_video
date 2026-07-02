import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAuthStore = defineStore('auth', () => {
    const token = ref(localStorage.getItem('access_token') || '')
    const refreshTokenValue = ref(localStorage.getItem('refresh_token') || '')
    const avatarUrl = ref(localStorage.getItem('avatar_url') || '')

    const isLoggedIn = computed(() => !!token.value)

    const claims = computed(() => {
        if (!token.value) return null
        try {
            const payload = token.value.split('.')[1]
            const decoded = atob(payload)
            return JSON.parse(decoded)
        } catch {
            return null
        }
    })

    function setTokens(access: string, refresh: string) {
        token.value = access
        refreshTokenValue.value = refresh
        localStorage.setItem('access_token', access)
        localStorage.setItem('refresh_token', refresh)
    }

    function setAvatar(url: string) {
        avatarUrl.value = url
        localStorage.setItem('avatar_url', url)
    }

    function clearTokens() {
        token.value = ''
        refreshTokenValue.value = ''
        avatarUrl.value = ''
        localStorage.removeItem('access_token')
        localStorage.removeItem('refresh_token')
        localStorage.removeItem('avatar_url')
    }

    // 从 JWT claims 提取用户信息（计算属性）
    // 提供 user 别名，修复 HomeView 中 auth.user?.id 的引用问题
    const user = computed(() => {
        if (!claims.value) return null
        return {
            id: claims.value.account_id,
            username: claims.value.username,
        }
    })

    return { token, refreshTokenValue, isLoggedIn, claims, avatarUrl, user, setTokens, setAvatar, clearTokens }
})