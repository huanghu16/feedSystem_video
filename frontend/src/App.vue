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


         <router-link v-if="auth.isLoggedIn" to="/profile" class="nav-item" :class="{ active: route.path === '/profile' }">
           <el-icon><User /></el-icon>
           <span>我的</span>
         </router-link>

        <router-link v-if="auth.isLoggedIn" to="/settings" class="nav-item" :class="{ active: route.path === '/settings' }">
          <el-icon><Setting /></el-icon>
          <span>设置</span>
        </router-link>
      </nav>

      <div class="user-status">
        <div v-if="auth.isLoggedIn" class="user-info">
          <div v-if="auth.avatarUrl" class="avatar-img">
            <img :src="getFullAvatarUrl(auth.avatarUrl)" alt="头像" />
          </div>
          <div v-else class="avatar">{{ auth.claims?.username?.[0]?.toUpperCase() || 'U' }}</div>
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
        <div class="search-container">
          <div class="search-box">
            <el-input
              v-model="searchQuery"
              placeholder="搜索视频..."
              prefix-icon="Search"
              clearable
              @keyup.enter="handleSearch"
            />
          </div>
          <el-button class="search-btn" @click="handleSearch" :disabled="!searchQuery.trim()">
            <el-icon><Search /></el-icon>
            搜索
          </el-button>
        </div>
        <div class="top-actions">
          <el-button v-if="!auth.isLoggedIn" class="glass-btn" @click="router.push('/account')">
            登录
          </el-button>
          <el-button v-else-if="route.path !== '/video/publish'" class="glass-btn" @click="router.push('/video/publish')">
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
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  HomeFilled, TrendCharts, Plus,
  User, Setting, Search
} from '@element-plus/icons-vue'
import { useAuthStore } from './stores/auth'
import { getFullAvatarUrl } from './composables/useImageUrl'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const searchQuery = ref('')

const pageTitle = computed(() => {
  const titles: Record<string, string> = {
    '/': '推荐',
    '/hot': '热榜',
    '/video/publish': '发布',
    '/profile': '我的',
    '/settings': '设置',

  }
  return titles[route.path] || 'FeedVideo'
})

function handleSearch() {
  if (!searchQuery.value.trim()) return
  router.push({ path: '/', query: { q: searchQuery.value } })
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
  padding: 0 14px 14px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  display: flex;
  justify-content: center;
  align-items: center;
}

.logo h1 {
  font-size: 22px;
  background: linear-gradient(135deg, #e94560, #ff6b8a);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  margin: 0;
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

.avatar-img {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  overflow: hidden;
  border: 2px solid #e94560;
}

.avatar-img img {
  width: 100%;
  height: 100%;
  object-fit: cover;
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

.search-container {
  flex: 1;
  max-width: 500px;
  display: flex;
  gap: 12px;
}

.search-box {
  flex: 1;
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

.search-btn {
  background: linear-gradient(135deg, #e94560, #ff6b8a) !important;
  border: none !important;
  color: #fff !important;
  font-weight: 500;
  transition: all 0.3s ease;
  white-space: nowrap;
}

.search-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, #d63851, #ff5a7a) !important;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(233, 69, 96, 0.4);
}

.search-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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
</style>
