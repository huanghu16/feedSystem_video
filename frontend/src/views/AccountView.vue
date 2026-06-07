<template>
  <div class="account-page">
    <div class="card">
    <h1>🎬 Feed System Video</h1>
      <h1>{{ isRegister ? '注册' : '登录' }}</h1>

      <el-form @submit.prevent="handleSubmit" label-position="top">
        <el-form-item>
          <el-input
            v-model="username"
            placeholder="用户名"
            :minlength="2"
            :maxlength="32"
            size="large"
          />
        </el-form-item>

        <el-form-item>
          <el-input
            v-model="password"
            type="password"
            placeholder="密码"
            :minlength="6"
            show-password
            size="large"
          />
        </el-form-item>

        <el-button
          type="primary"
          size="large"
          :loading="loading"
          class="submit-btn"
          native-type="submit"
        >
          {{ isRegister ? '注册' : '登录' }}
        </el-button>
      </el-form>

      <p class="switch" @click="isRegister = !isRegister">
        {{ isRegister ? '已有账号？去登录' : '没有账号？去注册' }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../stores/auth'
import { login, register } from '../api/account'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const isRegister = ref(false)
const username = ref('')
const password = ref('')
const loading = ref(false)

async function handleSubmit() {
  loading.value = true

  try {
    if (isRegister.value) {
      await register(username.value, password.value)
      ElMessage.success('注册成功，请登录')
      isRegister.value = false
    } else {
      const res = await login(username.value, password.value)
      auth.setTokens(res.access_token, res.refresh_token)
      ElMessage.success('登录成功')
      const redirect = (route.query.redirect as string) || '/'
      router.push(redirect)
    }
  } catch (e: any) {
    ElMessage.error(e?.payload?.message || '操作失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.account-page {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: #1a1a2e;
}

.card {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  padding: 40px;
  width: 360px;
}

.card h1 {
  color: #fff;
  text-align: center;
  margin-bottom: 24px;
  font-size: 24px;
}

/* ===== 覆盖 Element Plus 默认样式，保持暗色主题 ===== */

/* 输入框 */
.card :deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.08) !important;
  border: 1px solid rgba(255, 255, 255, 0.2) !important;
  border-radius: 8px !important;
  box-shadow: none !important;
  padding: 4px 16px !important;
}

.card :deep(.el-input__wrapper:hover) {
  border-color: rgba(255, 255, 255, 0.3) !important;
}

.card :deep(.el-input__wrapper.is-focus) {
  border-color: #e94560 !important;
  box-shadow: 0 0 0 1px rgba(233, 69, 96, 0.3) !important;
}

.card :deep(.el-input__inner) {
  color: #fff !important;
  font-size: 14px;
}

.card :deep(.el-input__inner::placeholder) {
  color: rgba(255, 255, 255, 0.4) !important;
}

/* 密码显示/隐藏图标 */
.card :deep(.el-input__password) {
  color: rgba(255, 255, 255, 0.5) !important;
}

/* 表单项间距 */
.card :deep(.el-form-item) {
  margin-bottom: 16px;
}

.card :deep(.el-form-item__label) {
  display: none; /* 隐藏 label，用 placeholder 代替 */
}

/* 提交按钮 */
.submit-btn {
  width: 100%;
  height: 44px;
  border-radius: 8px !important;
  font-size: 16px !important;
  background: #e94560 !important;
  border-color: #e94560 !important;
}

.submit-btn:hover {
  background: #d63851 !important;
  border-color: #d63851 !important;
}

.submit-btn:active {
  background: #c23047 !important;
  border-color: #c23047 !important;
}

/* 切换链接 */
.switch {
  color: rgba(255, 255, 255, 0.6);
  text-align: center;
  margin-top: 16px;
  cursor: pointer;
  font-size: 14px;
}

.switch:hover {
  color: #e94560;
}
</style>