<template>
  <div class="app">
    <!-- 侧边栏 -->
    <aside class="sidebar">
      <div class="logo">
        <h1>🎬 FeedVideo</h1>
      </div>

      <nav class="nav-menu">
        <router-link to="/" class="nav-item" :class="{ active: route.path === '/' }">
          <el-icon><HomeFilled /></el-icon>
          <span>首页</span>
        </router-link>

        <router-link to="/hot" class="nav-item" :class="{ active: route.path === '/hot' }">
          <el-icon><TrendCharts /></el-icon>
          <span>热榜</span>
        </router-link>

        <router-link v-if="auth.isLoggedIn" to="/video/publish" class="nav-item">
          <el-icon><Plus /></el-icon>
          <span>发布</span>
        </router-link>

        <router-link to="/messages" class="nav-item" :class="{ active: route.path === '/messages' }">
          <el-icon><ChatDotRound /></el-icon>
          <span>消息</span>
        </router-link>

        <router-link v-if="auth.isLoggedIn" to="/settings" class="nav-item" :class="{ active: route.path === '/settings' }">
          <el-icon><Setting /></el-icon>
          <span>设置</span>
        </router-link>
      </nav>

      <div class="user-status">
        <div v-if="auth.isLoggedIn" class="user-info">
          <div class="avatar">{{ auth.claims?.username?.[0]?.toUpperCase() || 'U' }}</div>
          <span class="username">{{ auth.claims?.username || '用户' }}</span>
          <div class="status-dot online"></div>
        </div>
        <div v-else class="user-info">
          <div class="avatar">?</div>
          <span class="username">未登录</span>
          <div class="status-dot offline"></div>
        </div>
      </div>
    </aside>

    <!-- 主内容区 -->
    <main class="main">
      <!-- 顶部栏 -->
      <header class="topbar">
        <h2>{{ pageTitle }}</h2>
        <div class="search-box">
          <el-input
            v-model="searchQuery"
            placeholder="搜索视频..."
            prefix-icon="Search"
            clearable
            @keyup.enter="handleSearch"
          />
        </div>
        <div class="top-actions">
          <el-button v-if="!auth.isLoggedIn" class="glass-btn" @click="router.push('/account')">
            登录
          </el-button>
          <el-button v-else class="glass-btn" @click="showUpload = true">
            <el-icon><Plus /></el-icon>
            发布视频
          </el-button>
        </div>
      </header>

      <!-- 页面内容 -->
      <div class="content">
        <RouterView />
      </div>
    </main>
  </div>

  <!-- 全局上传对话框 -->
  <el-dialog v-model="showUpload" title="发布视频" width="500px" :close-on-click-modal="false">
    <el-form label-position="top">
      <el-form-item label="选择视频">
        <el-upload
          accept="video/*"
          :auto-upload="false"
          :on-change="handleFileChange"
          :limit="1"
        >
          <el-button type="primary">选择文件</el-button>
        </el-upload>
      </el-form-item>
      <el-form-item label="标题">
        <el-input v-model="uploadTitle" placeholder="给你的视频起个标题" maxlength="256" show-word-limit />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="showUpload = false">取消</el-button>
      <el-button type="primary" :loading="uploading" @click="handleUpload">发布</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  HomeFilled, TrendCharts, Plus, ChatDotRound,
  User, Setting
} from '@element-plus/icons-vue'
import { useAuthStore } from './stores/auth'
import { uploadVideo, publishVideo } from './api/video'

const route = useRoute()  // 当前路由
const router = useRouter() // 路由实例
const auth = useAuthStore() // 认证状态

const searchQuery = ref('') // 搜索关键词
const showUpload = ref(false) // 上传对话框显示状态
const uploadFile = ref<File | null>(null) // 上传文件
const uploadTitle = ref('') // 上传标题
const uploading = ref(false) // 

// 页面标题
const pageTitle = computed(() => {
  const titles: Record<string, string> = {
    '/': '推荐',
    '/hot': '热榜',
    '/messages': '私信',
    '/account': '我的',
    '/settings': '设置',
  }
  return titles[route.path] || 'FeedVideo'
})

// 搜索
function handleSearch() {
  if (!searchQuery.value.trim()) return
  // 简单实现：跳转到首页并带搜索参数
  router.push({ path: '/', query: { q: searchQuery.value } })
}

// 上传视频
function handleFileChange(file: any) {
  uploadFile.value = file.raw
}

// 发布视频
async function handleUpload() {
  if (!uploadFile.value || !uploadTitle.value.trim()) {
    ElMessage.warning('请选择视频并填写标题')
    return
  }

  uploading.value = true
  try {
    const uploadRes = await uploadVideo(uploadFile.value)
    await publishVideo(uploadTitle.value, uploadRes.play_url, '')
    ElMessage.success('发布成功')
    showUpload.value = false
    uploadTitle.value = ''
    uploadFile.value = null
    // 刷新页面
    router.go(0)
  } catch {
    ElMessage.error('发布失败')
  } finally {
    uploading.value = false
  }
}
</script>

<style>
/* 全局样式 */
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  background: #0f0f1a;
  color: #fff;
}
</style>

<style scoped>
.app {
  display: flex;
  min-height: 100vh;
}

/* ===== 侧边栏 ===== */
.sidebar {
  width: 240px;
  background: rgba(255, 255, 255, 0.03);
  border-right: 1px solid rgba(255, 255, 255, 0.06);
  display: flex;
  flex-direction: column;
  padding: 20px 0;
  position: fixed;
  height: 100vh;
}

.logo {
  padding: 0 24px 24px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.logo h1 {
  font-size: 22px;
  background: linear-gradient(135deg, #e94560, #ff6b8a);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.nav-menu {
  flex: 1;
  padding: 16px 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-radius: 10px;
  color: rgba(255, 255, 255, 0.6);
  text-decoration: none;
  font-size: 15px;
  transition: all 0.2s;
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.05);
  color: #fff;
}

.nav-item.active {
  background: rgba(233, 69, 96, 0.15);
  color: #e94560;
}

.nav-item .el-icon {
  font-size: 20px;
}

.user-status {
  padding: 16px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
}

.user-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: linear-gradient(135deg, #e94560, #ff6b8a);
  display: flex;
  justify-content: center;
  align-items: center;
  font-size: 14px;
  font-weight: bold;
  color: #fff;
}

.username {
  flex: 1;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.8);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.status-dot.online {
  background: #2ecc71;
}

.status-dot.offline {
  background: #e74c3c;
}

/* ===== 主内容区 ===== */
.main {
  flex: 1;
  margin-left: 240px;
  display: flex;
  flex-direction: column;
}

.topbar {
  height: 64px;
  background: rgba(255, 255, 255, 0.02);
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  display: flex;
  align-items: center;
  padding: 0 32px;
  gap: 24px;
  position: sticky;
  top: 0;
  z-index: 100;
}

.topbar h2 {
  font-size: 20px;
  font-weight: 600;
  min-width: 80px;
}

.search-box {
  flex: 1;
  max-width: 400px;
}

.search-box :deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.06) !important;
  border: 1px solid rgba(255, 255, 255, 0.1) !important;
  border-radius: 20px !important;
  box-shadow: none !important;
}

.search-box :deep(.el-input__inner) {
  color: #fff !important;
}

.search-box :deep(.el-input__inner::placeholder) {
  color: rgba(255, 255, 255, 0.3) !important;
}

.top-actions {
  display: flex;
  gap: 12px;
}

.glass-btn {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.2);
  color: #fff;
  backdrop-filter: blur(10px);
  transition: all 0.3s ease;
}

.glass-btn:hover {
  background: rgba(233, 69, 96, 0.2);
  border-color: #e94560;
  color: #e94560;
  backdrop-filter: blur(15px);
}

.content {
  flex: 1;
  padding: 24px 32px;
}

/* ===== 对话框样式覆盖 ===== */
:deep(.el-dialog) {
  background: #1a1a2e;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

:deep(.el-dialog__title) {
  color: #fff;
}

:deep(.el-form-item__label) {
  color: rgba(255, 255, 255, 0.7);
}

:deep(.el-upload) {
  color: rgba(255, 255, 255, 0.5);
}
</style>