<template>
  <div class="profile-page">
    <!-- 用户信息卡片 -->
    <div class="profile-header">
      <div class="avatar-section">
        <div v-if="profile?.avatar_url" class="avatar-large">
          <img :src="getFullAvatarUrl(profile.avatar_url)" alt="头像" />
        </div>
        <div v-else class="avatar-large placeholder">
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
          <div class="stat-item" :class="{ clickable: isOwnProfile }" @click="isOwnProfile && openFollowingDrawer()">
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
      <div class="section-header">
        <h2 class="section-title">作品</h2>
        <div class="section-actions">
          <el-button
            v-if="isOwnProfile && videos.length > 0"
            :class="batchMode ? 'glass-btn active' : 'glass-btn'"
            size="small"
            @click="toggleBatchMode"
          >
            <el-icon><Delete /></el-icon>
            {{ batchMode ? '取消批量' : '批量管理' }}
          </el-button>
          <el-button
            v-if="isOwnProfile && batchMode && selectedVideos.size > 0"
            :class="'glass-btn delete-selected-btn'"
            size="small"
            :loading="deleting"
            @click="handleBatchDelete"
          >
            <el-icon><Delete /></el-icon>
            删除选中({{ selectedVideos.size }})
          </el-button>
        </div>
      </div>

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
          :class="{ selected: selectedVideos.has(video.id) }"
          @click="handleCardClick(video)"
        >
          <div v-if="isOwnProfile && batchMode" class="select-checkbox" @click.stop="toggleSelect(video.id)">
            <el-icon v-if="selectedVideos.has(video.id)" class="checked"><CircleCheckFilled /></el-icon>
            <div v-else class="checkbox-unchecked"></div>
          </div>
          <div class="work-cover">
            <img v-if="video.cover_url" :src="getFullUrl(video.cover_url)" class="cover-image" />
            <video
              v-else-if="playingId === video.id"
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
          <div v-if="isOwnProfile && !batchMode" class="video-actions-overlay">
            <el-button
              type="danger"
              size="small"
              circle
              class="delete-btn"
              @click.stop="handleDeleteVideo(video)"
            >
              <el-icon><Delete /></el-icon>
            </el-button>
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

    <!-- 关注列表抽屉 -->
    <el-drawer v-model="followingDrawer" title="我的关注" size="400px" :with-header="false" class="following-drawer-wrapper">
      <div class="following-drawer-content">
        <h3>我的关注</h3>
        <div class="following-list">
          <div v-if="followingLoading" class="drawer-loading">
            <el-icon class="is-loading" :size="24"><Loading /></el-icon>
            <span>加载中...</span>
          </div>
          <div v-else-if="followingList.length === 0" class="drawer-empty">
            <el-empty description="暂未关注任何人" />
          </div>
          <div v-for="user in followingList" v-else :key="user.id" class="following-item">
            <div class="following-avatar">
              <img v-if="user.avatar_url" :src="getFullAvatarUrl(user.avatar_url)" alt="头像" />
              <div v-else class="avatar-placeholder">
                {{ user.username?.[0]?.toUpperCase() || 'U' }}
              </div>
            </div>
            <div class="following-detail">
              <strong class="following-name" @click="goToUserProfile(user.id)">{{ user.username }}</strong>
              <p v-if="user.bio" class="following-bio">{{ user.bio }}</p>
              <div class="following-stats">
                <span>作品 {{ user.video_count || 0 }}</span>
                <span>粉丝 {{ user.fans_count || 0 }}</span>
                <span>关注 {{ user.following_count || 0 }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { VideoPlay, Star, ChatDotRound, Loading, Delete, CircleCheckFilled } from '@element-plus/icons-vue'
import { useAuthStore } from '../stores/auth'
import { getProfile, type ProfileInfo } from '../api/account'
import { listByAuthor, deleteVideo, deleteVideosBatch } from '../api/video'
import { getVloggers } from '../api/social'
import type { VideoItem, VloggerItem } from '../api/types'
import { getFullUrl, getFullAvatarUrl } from '../composables/useImageUrl'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

// 获取目标用户 ID：优先从 query.uid（查看他人），否则用当前登录用户
function getTargetUserId(): number | undefined {
  const queryUid = route.query.uid
  if (queryUid) {
    const uid = Number(queryUid)
    if (!isNaN(uid)) return uid
  }
  return auth.claims?.account_id
}

// 判断是否查看自己的主页
const isOwnProfile = ref(true)

const profile = ref<ProfileInfo | null>(null)
const videos = ref<VideoItem[]>([])
const playingId = ref<number | null>(null)
const loading = ref(false)
const deleting = ref(false)

// 批量选择相关
const batchMode = ref(false)
const selectedVideos = ref<Set<number>>(new Set())

// 分页相关
const currentPage = ref(1)
const pageSize = ref(12) // 每页12个（4列 x 3行）
const total = ref(0)

onMounted(async () => {
  if (!auth.isLoggedIn) {
    ElMessage.warning('请先登录')
    router.push('/account')
    return
  }

  await loadProfile()
  await loadVideos()
})

// 监听路由 query 变化，支持查看不同用户
watch(() => route.query.uid, async () => {
  if (auth.isLoggedIn) {
    await loadProfile()
    await loadVideos()
  }
})

async function loadProfile() {
  try {
    const userId = getTargetUserId()
    if (!userId) {
      ElMessage.error('无法获取用户信息')
      return
    }
    isOwnProfile.value = userId === auth.claims?.account_id
    const data = await getProfile(userId)
    profile.value = data
    if (isOwnProfile.value && data.avatar_url) {
      auth.setAvatar(data.avatar_url)
    }
  } catch (error: any) {
    console.error('加载用户资料失败:', error)
    ElMessage.error(error?.payload?.message || '加载用户资料失败')
  }
}

async function loadVideos() {
  loading.value = true
  try {
    const userId = getTargetUserId()
    if (!userId) {
      ElMessage.error('无法获取用户信息')
      return
    }

    const resp = await listByAuthor(userId)

    total.value = resp.total
    videos.value = resp.list
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

// 切换批量模式
function toggleBatchMode() {
  batchMode.value = !batchMode.value
  if (!batchMode.value) {
    selectedVideos.value.clear()
  }
}

// 切换单个视频选择
function toggleSelect(videoId: number) {
  if (selectedVideos.value.has(videoId)) {
    selectedVideos.value.delete(videoId)
  } else {
    selectedVideos.value.add(videoId)
  }
}

// 卡片点击处理
function handleCardClick(video: VideoItem) {
  if (batchMode.value) {
    toggleSelect(video.id)
  } else {
    playVideo(video)
  }
}

// 删除单个视频
async function handleDeleteVideo(video: VideoItem) {
  try {
    await ElMessageBox.confirm(
      `确定要删除视频"${video.title}"吗？`,
      '删除确认',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )

    deleting.value = true
    await deleteVideo(video.id)

    // 从列表中移除
    videos.value = videos.value.filter(v => v.id !== video.id)
    total.value = videos.value.length

    ElMessage.success('删除成功')
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
      ElMessage.error(error?.payload?.message || '删除失败')
    }
  } finally {
    deleting.value = false
  }
}

// 批量删除
async function handleBatchDelete() {
  if (selectedVideos.value.size === 0) {
    ElMessage.warning('请选择要删除的视频')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedVideos.value.size} 个视频吗？`,
      '批量删除确认',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )

    deleting.value = true
    await deleteVideosBatch(Array.from(selectedVideos.value))

    // 从列表中移除
    videos.value = videos.value.filter(v => !selectedVideos.value.has(v.id))
    total.value = videos.value.length

    // 记录删除数量（在 clear 之前）
    const deletedCount = selectedVideos.value.size

    // 清空选择
    selectedVideos.value.clear()
    batchMode.value = false

    ElMessage.success(`成功删除 ${deletedCount} 个视频`)
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('批量删除失败:', error)
      ElMessage.error(error?.payload?.message || '批量删除失败')
    }
  } finally {
    deleting.value = false
  }
}

// ===== 关注列表抽屉 =====
const followingDrawer = ref(false)
const followingLoading = ref(false)
const followingList = ref<VloggerItem[]>([])

async function openFollowingDrawer() {
  followingDrawer.value = true
  followingLoading.value = true
  try {
    const resp = await getVloggers(1, 50)
    followingList.value = resp.list
  } catch (error: any) {
    ElMessage.error(error?.payload?.message || '加载关注列表失败')
  } finally {
    followingLoading.value = false
  }
}

// 点击用户名，跳转到该用户的详情页（只读查看）
function goToUserProfile(userId: number) {
  followingDrawer.value = false
  router.push({ path: '/profile', query: { uid: String(userId) } })
}
</script>

<style>
.following-drawer-wrapper .el-drawer__body {
  padding: 0 !important;
  margin: 0 !important;
}

.following-drawer-wrapper .el-drawer {
  padding: 0 !important;
  margin: 0 !important;
}
</style>

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
  overflow: hidden;
}

.avatar-large.placeholder {
  background: linear-gradient(135deg, #e94560, #ff6b8a);
}

.avatar-large img {
  width: 100%;
  height: 100%;
  object-fit: cover;
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

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.section-title {
  color: #fff;
  font-size: 20px;
  margin: 0;
}

.section-actions {
  display: flex;
  gap: 8px;
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

.glass-btn.active {
  background: rgba(233, 69, 96, 0.2);
  border-color: #e94560;
  color: #e94560;
  backdrop-filter: blur(15px);
}

.delete-selected-btn:hover {
  background: rgba(233, 69, 96, 0.2);
  border-color: #e94560;
  color: #e94560;
  backdrop-filter: blur(15px);
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
  position: relative;
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

.work-card.selected {
  border-color: #e94560;
  box-shadow: 0 0 0 2px rgba(233, 69, 96, 0.3);
}

.select-checkbox {
  position: absolute;
  top: 12px;
  left: 12px;
  z-index: 10;
  width: 28px;
  height: 28px;
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: pointer;
}

.select-checkbox .checked {
  font-size: 28px;
  color: #e94560;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.3));
}

.checkbox-unchecked {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.3);
  border: 2px solid rgba(255, 255, 255, 0.6);
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

.cover-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
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

.video-actions-overlay {
  position: absolute;
  top: 12px;
  right: 12px;
  opacity: 0;
  transition: opacity 0.2s;
}

.work-card:hover .video-actions-overlay {
  opacity: 1;
}

.delete-btn {
  background: rgba(233, 69, 96, 0.9);
  border-color: #e94560;
  backdrop-filter: blur(8px);
}

.delete-btn:hover {
  background: #e94560;
  transform: scale(1.1);
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

/* ===== 关注统计项可点击 ===== */
.stat-item.clickable {
  cursor: pointer;
  transition: transform 0.2s;
}

.stat-item.clickable:hover {
  transform: scale(1.1);
}

.stat-item.clickable:hover .stat-value {
  color: #e94560;
}

/* ===== 关注列表抽屉（与评论抽屉同风格） ===== */
.following-drawer-content {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: #1a1a2e;
  width: 100%;
}

.following-drawer-content h3 {
  color: #fff;
  margin: 0;
  padding: 20px 24px 16px 24px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  font-size: 20px;
}

.following-list {
  flex: 1;
  overflow-y: auto;
  padding: 0 24px;
}

.drawer-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 48px 0;
  color: rgba(255, 255, 255, 0.5);
}

.drawer-empty {
  padding: 32px 0;
}

.following-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.following-avatar {
  flex-shrink: 0;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  overflow: hidden;
}

.following-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-placeholder {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  background: linear-gradient(135deg, #e94560, #ff6b8a);
  display: flex;
  justify-content: center;
  align-items: center;
  font-size: 16px;
  font-weight: bold;
  color: #fff;
}

.following-detail {
  flex: 1;
  min-width: 0;
}

.following-detail strong {
  color: #e94560;
  font-size: 13px;
}

.following-name {
  cursor: pointer;
  transition: opacity 0.2s;
}

.following-name:hover {
  opacity: 0.7;
  text-decoration: underline;
}

.following-bio {
  color: rgba(255, 255, 255, 0.5);
  font-size: 12px;
  margin: 4px 0;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.following-stats {
  display: flex;
  gap: 16px;
  color: rgba(255, 255, 255, 0.5);
  font-size: 12px;
  margin-top: 4px;
}

.following-stats span {
  white-space: nowrap;
}
</style>