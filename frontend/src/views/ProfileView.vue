<template>
  <div class="profile-page">
    <!-- 用户信息卡片 -->
    <div class="profile-header">
      <div class="avatar-section">
        <div class="avatar-large">
          {{ profile?.username?.[0]?.toUpperCase() || 'U' }}
        </div>
      </div>

      <div class="user-details">
        <h1 class="username">{{ profile?.username || '未登录' }}</h1>
        <p class="bio">{{ profile?.bio || '这个人很神秘，什么都没有写~' }}</p>

        <div class="stats-row">
          <div class="stat-item">
            <span class="stat-value">{{ profile?.video_count || 0 }}</span>
            <span class="stat-label">作品</span>
          </div>
          <div class="stat-item">
            <span class="stat-value">{{ profile?.fans_count || 0 }}</span>
            <span class="stat-label">粉丝</span>
          </div>
          <div class="stat-item">
            <span class="stat-value">{{ profile?.following_count || 0 }}</span>
            <span class="stat-label">关注</span>
          </div>
          <div class="stat-item">
            <span class="stat-value">{{ profile?.likes_count || 0 }}</span>
            <span class="stat-label">获赞</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 作品列表 -->
    <div class="works-section">
      <h2 class="section-title">作品</h2>

      <div v-if="loading" class="loading">
        <el-icon class="is-loading" :size="32"><Loading /></el-icon>
        <span>加载中...</span>
      </div>

      <div v-else-if="videos.length === 0" class="empty-state">
        <el-empty description="暂无作品" />
      </div>

      <div v-else class="works-grid">
        <div
          v-for="video in videos"
          :key="video.id"
          class="work-card"
          @click="playVideo(video)"
        >
          <div class="work-cover">
            <video
              v-if="playingId === video.id"
              :src="getFullUrl(video.play_url)"
              controls
              autoplay
              class="work-player"
            />
            <div v-else class="cover-placeholder">
              <el-icon :size="48" color="rgba(255,255,255,0.3)"><VideoPlay /></el-icon>
            </div>
          </div>
          <div class="work-info">
            <h3>{{ video.title }}</h3>
            <div class="work-stats">
              <span><el-icon><VideoPlay /></el-icon> {{ video.play_count }}</span>
              <span><el-icon><Star /></el-icon> {{ video.likes_count }}</span>
              <span><el-icon><ChatDotRound /></el-icon> {{ video.comments_count }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 分页 -->
      <div v-if="total > pageSize" class="pagination">
        <el-pagination
          v-model:current-page="currentPage"
          :page-size="pageSize"
          :total="total"
          layout="prev, pager, next"
          prev-text="上一页"
          next-text="下一页"
          @current-change="handlePageChange"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { VideoPlay, Star, ChatDotRound, Loading } from '@element-plus/icons-vue'
import { useAuthStore } from '../stores/auth'
import { getProfile, type ProfileInfo } from '../api/account'
import { listByAuthor } from '../api/video'
import type { VideoItem } from '../api/types'

const router = useRouter()
const auth = useAuthStore()

const profile = ref<ProfileInfo | null>(null)
const videos = ref<VideoItem[]>([])
const playingId = ref<number | null>(null)
const loading = ref(false)

// 分页相关
const currentPage = ref(1)
const pageSize = ref(12) // 每页12个（4列 x 3行）
const total = ref(0)

function getFullUrl(path: string) {
  if (path.startsWith('http')) return path
  return `http://localhost:8080${path}`
}

onMounted(async () => {
  if (!auth.isLoggedIn) {
    ElMessage.warning('请先登录')
    router.push('/account')
    return
  }

  await loadProfile()
  await loadVideos()
})

async function loadProfile() {
  try {
    const userId = auth.claims?.account_id
    if (!userId) {
      ElMessage.error('无法获取用户信息')
      return
    }
    const data = await getProfile(userId)
    profile.value = data
  } catch (error: any) {
    console.error('加载用户资料失败:', error)
    ElMessage.error(error?.payload?.message || '加载用户资料失败')
  }
}

async function loadVideos() {
  loading.value = true
  try {
    const userId = auth.claims?.account_id
    if (!userId) {
      ElMessage.error('无法获取用户信息')
      return
    }

    const allVideos = await listByAuthor(userId)
    console.log('=== 视频数据 ===')
    console.log('allVideos:', allVideos)
    console.log('allVideos.length:', allVideos.length)
    console.log('Array.isArray(allVideos):', Array.isArray(allVideos))

    total.value = allVideos.length

    // 直接赋值，不分页（先确保能显示）
    videos.value = allVideos

    console.log('videos.value:', videos.value)
    console.log('videos.value.length:', videos.value.length)
  } catch (error: any) {
    console.error('加载作品失败:', error)
    ElMessage.error(error?.payload?.message || '加载作品失败')
  } finally {
    loading.value = false
  }
}

function playVideo(video: VideoItem) {
  router.push(`/video/${video.id}`)
}

async function handlePageChange(page: number) {
  currentPage.value = page
  await loadVideos()
}
</script>

<style scoped>
.profile-page {
  max-width: 1200px;
  margin: 0 auto;
}

/* ===== 用户信息头部 ===== */
.profile-header {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  padding: 40px;
  display: flex;
  gap: 32px;
  align-items: center;
  margin-bottom: 32px;
}

.avatar-section {
  flex-shrink: 0;
}

.avatar-large {
  width: 120px;
  height: 120px;
  border-radius: 50%;
  background: linear-gradient(135deg, #e94560, #ff6b8a);
  display: flex;
  justify-content: center;
  align-items: center;
  font-size: 48px;
  font-weight: bold;
  color: #fff;
  box-shadow: 0 8px 24px rgba(233, 69, 96, 0.3);
}

.user-details {
  flex: 1;
}

.username {
  color: #fff;
  font-size: 28px;
  margin-bottom: 8px;
}

.bio {
  color: rgba(255, 255, 255, 0.6);
  font-size: 14px;
  margin-bottom: 24px;
}

.stats-row {
  display: flex;
  gap: 32px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.stat-value {
  color: #fff;
  font-size: 24px;
  font-weight: bold;
}

.stat-label {
  color: rgba(255, 255, 255, 0.5);
  font-size: 13px;
}

/* ===== 作品区域 ===== */
.works-section {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 16px;
  padding: 24px;
}

.section-title {
  color: #fff;
  font-size: 20px;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 48px 0;
  color: rgba(255, 255, 255, 0.6);
}

.empty-state {
  padding: 48px 0;
}

/* ===== 作品网格（4列布局）===== */
.works-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

@media (max-width: 1200px) {
  .works-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 900px) {
  .works-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 600px) {
  .works-grid {
    grid-template-columns: 1fr;
  }
}

.work-card {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  overflow: hidden;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
}

.work-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
}

.work-cover {
  width: 100%;
  height: 200px;
  background: #000;
  position: relative;
}

.work-player {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.cover-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
}

.work-info {
  padding: 12px;
}

.work-info h3 {
  color: #fff;
  font-size: 14px;
  margin-bottom: 8px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.work-stats {
  display: flex;
  gap: 12px;
  color: rgba(255, 255, 255, 0.5);
  font-size: 12px;
}

.work-stats span {
  display: flex;
  align-items: center;
  gap: 4px;
}

.work-stats .el-icon {
  font-size: 14px;
}

/* ===== 分页 ===== */
.pagination {
  display: flex;
  justify-content: center;
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
}

.pagination :deep(.el-pagination) {
  --el-pagination-bg-color: rgba(255, 255, 255, 0.05);
  --el-pagination-text-color: rgba(255, 255, 255, 0.8);
  --el-pagination-border-color: rgba(255, 255, 255, 0.1);
  --el-pagination-hover-color: #e94560;
}

.pagination :deep(.el-pager li) {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.8);
  min-width: 32px;
  height: 32px;
  line-height: 32px;
  border-radius: 6px;
  margin: 0 4px;
}

.pagination :deep(.el-pager li.is-active) {
  background: linear-gradient(135deg, #e94560, #ff6b8a);
  border-color: #e94560;
  color: #fff;
}

.pagination :deep(.btn-prev),
.pagination :deep(.btn-next) {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.8);
  border-radius: 6px;
}
</style>