<template>
  <div class="settings">
    <h1>设置</h1>
    <button @click="handleLogout">退出登录</button>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()

function handleLogout() {
  auth.clearTokens()
  router.push('/account')
}
</script>

<style scoped>
.settings {
  padding: 40px;
  color: #fff;
  background: #1a1a2e;
  min-height: 100vh;
}

.settings button {
  padding: 10px 24px;
  border: 1px solid #e94560;
  border-radius: 8px;
  background: transparent;
  color: #e94560;
  cursor: pointer;
  font-size: 14px;
}
</style>