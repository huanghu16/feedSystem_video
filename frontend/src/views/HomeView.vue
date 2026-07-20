<template>
  <div class="home">
    <!-- 搜索结果提示 -->
    <div v-if="searchKeyword" class="search-info">
      <span>搜索 "{{ searchKeyword }}" 的结果</span>
      <el-button size="small" @click="clearSearch">清除搜索</el-button>
    </div>

    <!-- 视频列表 -->
    <div class="video-list">
      <div
        v-for="video in videos"
        :key="video.id"
        class="video-card"
        @click="playVideo(video)"
      >
        <div class="video-cover">
          <!-- 播放中：显示 video 播放器 -->
          <video
            v-if="playingId === video.id"
            :src="getFullUrl(video.play_url)"
            controls
            autoplay
            class="video-player"
          />
          <!-- 未播放：显示封面图 + 播放图标叠加 -->
          <template v-else>
            <img v-if="video.cover_url" :src="getFullUrl(video.cover_url)" class="cover-image" />
            <div v-else class="cover-placeholder"></div>
            <div class="play-overlay">
              <el-icon :size="48" color="#fff"><VideoPlay /></el-icon>
            </div>
          </template>
        </div>
        <div class="video-info">
          <h3>{{ video.title }}</h3>
          <p v-if="video.description" class="description">{{ video.description }}</p>
          <p class="author">@{{ video.username }} <span class="publish-date"><el-icon><Calendar /></el-icon> {{ formatDate(video.created_at) }}</span></p>
          <div class="actions">
            <el-button
              :type="video.isLiked ? 'danger' : 'default'"
              size="small"
              @click.stop="toggleLike(video)"
            >
              <el-icon><Star /></el-icon>
              {{ video.likes_count }}
            </el-button>
            <el-button size="small" @click.stop="showComments(video)">
              <el-icon><ChatDotRound /></el-icon>
              {{ video.comments_count || 0 }}
            </el-button>
            <el-button size="small" class="play-count-btn">
              <el-icon><VideoPlay /></el-icon>
              {{ video.play_count || 0 }}
            </el-button>
            <el-button
              v-if="auth.isLoggedIn && auth.user?.id !== video.author_id"
              :type="video.isFollowing ? 'success' : 'default'"
              size="small"
              @click.stop="toggleFollow(video)"
              class="follow-btn"
            >
              <el-icon><UserFilled /></el-icon>
              {{ video.isFollowing ? '已关注' : '关注' }}
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <el-empty v-if="videos.length === 0 && !loading" :description="searchKeyword ? '未找到相关视频' : '视频加载中...'" />

    <!-- 加载状态 -->
    <div v-if="loading" class="loading">
      <el-icon class="is-loading" :size="32"><Loading /></el-icon>
      <span>加载中...</span>
    </div>

    <!-- 评论抽屉 -->
    <el-drawer v-model="commentDrawer" title="评论" size="400px" :with-header="false" class="comment-drawer-wrapper">
      <div class="comment-drawer-content">
        <h3>评论</h3>
        <div class="comment-list">
          <div v-for="c in comments" :key="c.id" class="comment-item">
            <div class="comment-avatar">
              <img v-if="c.avatar_url" :src="getFullAvatarUrl(c.avatar_url)" alt="头像" />
              <div v-else class="avatar-fallback">{{ c.username?.[0]?.toUpperCase() || 'U' }}</div>
            </div>
            <div class="comment-body">
              <strong>@{{ c.username }}</strong>
              <p>{{ c.content }}</p>
              <span class="time">{{ formatTime(c.created_at) }}</span>
            </div>
          </div>
          <el-empty v-if="comments.length === 0" description="暂无评论" />
        </div>
        <div v-if="auth.isLoggedIn" class="comment-input">
          <el-input
            v-model="newComment"
            placeholder="写评论..."
            @keyup.enter="submitComment"
          >
            <template #append>
              <el-button type="primary" @click="submitComment">发送</el-button>
            </template>
          </el-input>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { VideoPlay, Star, ChatDotRound, Loading, UserFilled, Calendar } from '@element-plus/icons-vue'
import { useAuthStore } from '../stores/auth'
import { listLatest } from '../api/feed'
import { like, unlike, isLiked as checkIsLiked } from '../api/like'
import { listComments, publishComment } from '../api/comment'
import { recordPlay, searchVideos } from '../api/video'
import { follow as followUser, unfollow as unfollowUser, isFollowing as checkIsFollowing } from '../api/social'
import type { FeedVideoItem, CommentItem } from '../api/types'
import { getFullUrl, getFullAvatarUrl } from '../composables/useImageUrl'
import { formatDate, formatTime } from '../composables/useFormat'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const videos = ref<(FeedVideoItem & { isLiked?: boolean; isFollowing?: boolean })[]>([])
const playingId = ref<number | null>(null)
const commentDrawer = ref(false)
const comments = ref<CommentItem[]>([])
const currentVideoId = ref(0)
const newComment = ref('')

const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const loading = ref(false)
const searchKeyword = ref('')

onMounted(async () => {
  const q = route.query.q as string
  if (q) {
    searchKeyword.value = q
    await loadSearchResults(1)
  } else {
    await loadFeed()
  }
})

watch(() => route.query.q, async (newQ) => {
  const q = newQ as string
  if (q) {
    searchKeyword.value = q
    await loadSearchResults(1)
  } else {
    searchKeyword.value = ''
    await loadFeed()
  }
})

async function loadFeed() {
  try {
    const data = await listLatest()
    videos.value = data.list.map(v => ({ ...v, isLiked: false, isFollowing: false }))
    total.value = data.total

    if (auth.isLoggedIn) {
      await loadVideoInteractions()
    }
  } catch {
    ElMessage.error('加载失败')
  }
}

async function loadSearchResults(page: number) {
  if (!searchKeyword.value.trim()) return

  loading.value = true
  try {
    const result = await searchVideos(searchKeyword.value, page, pageSize.value)
    videos.value = result.list.map(item => ({ ...item, isLiked: false, isFollowing: false }))
    total.value = result.total
    currentPage.value = result.page

    if (auth.isLoggedIn) {
      await loadVideoInteractions()
    }
  } catch {
    ElMessage.error('搜索失败')
  } finally {
    loading.value = false
  }
}

// 加载视频的点赞状态和关注状态（登录后调用）
async function loadVideoInteractions() {
  await Promise.all(
    videos.value.map(async (video) => {
      // 跳过自己的视频
      if (auth.user?.id === video.author_id) return

      try {
        const result = await checkIsFollowing(video.author_id)
        video.isFollowing = result.is_following
      } catch {}

      try {
        const liked = await checkIsLiked(video.id)
        video.isLiked = liked.is_liked
      } catch {}
    })
  )
}

function playVideo(video: FeedVideoItem) {
  if (playingId.value === video.id) {
    playingId.value = null
  } else {
    playingId.value = video.id
    recordPlay(video.id).catch(err => {
      console.error('记录播放失败:', err)
    })
  }
}

async function toggleLike(video: FeedVideoItem & { isLiked?: boolean }) {
  if (!auth.isLoggedIn) {
    ElMessage.warning('请先登录')
    router.push('/account')
    return
  }

  try {
    if (video.isLiked) {
      await unlike(video.id)
      video.isLiked = false
      video.likes_count--
      ElMessage.success('取消点赞')
    } else {
      await like(video.id)
      video.isLiked = true
      video.likes_count++
      ElMessage.success('点赞成功')
    }
  } catch (e: any) {
    ElMessage.error(e?.payload?.message || '操作失败')
  }
}

async function toggleFollow(video: FeedVideoItem & { isFollowing?: boolean }) {
  if (!auth.isLoggedIn) {
    ElMessage.warning('请先登录')
    router.push('/account')
    return
  }

  try {
    if (video.isFollowing) {
      await unfollowUser(video.author_id)
      video.isFollowing = false
      ElMessage.success('取消关注成功')
    } else {
      await followUser(video.author_id)
      video.isFollowing = true
      ElMessage.success('关注成功')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '操作失败')
  }
}

async function showComments(video: FeedVideoItem) {
  currentVideoId.value = video.id
  commentDrawer.value = true
  try {
    const resp = await listComments(video.id)
    comments.value = resp.list
  } catch {
    ElMessage.error('加载评论失败')
  }
}

async function submitComment() {
  if (!newComment.value.trim()) return
  try {
    const comment = await publishComment(currentVideoId.value, newComment.value)
    comments.value.push(comment)
    // 同步更新视频卡片上的评论计数
    const video = videos.value.find(v => v.id === currentVideoId.value)
    if (video) {
      video.comments_count = (video.comments_count || 0) + 1
    }
    newComment.value = ''
    ElMessage.success('评论成功')
  } catch {
    ElMessage.error('评论失败')
  }
}

function clearSearch() {
  searchKeyword.value = ''
  router.push('/')
}
</script>

<style>
.comment-drawer-wrapper .el-drawer__body {
  padding: 0 !important;
  margin: 0 !important;
}

.comment-drawer-wrapper .el-drawer {
  padding: 0 !important;
  margin: 0 !important;
}
</style>

<style scoped>
.home {
  max-width: 800px;
  margin: 0 auto;
}

.search-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: rgba(233, 69, 96, 0.1);
  border: 1px solid rgba(233, 69, 96, 0.3);
  border-radius: 8px;
  margin-bottom: 16px;
  color: #fff;
}

.search-info span {
  font-size: 14px;
}

.search-info :deep(.el-button) {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.2);
  color: #fff;
}

.video-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.video-card {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  overflow: hidden;
  cursor: pointer;
  transition: transform 0.2s;
}

.video-card:hover {
  transform: translateY(-2px);
}

.video-cover {
  width: 100%;
  height: 400px;
  background: #000;
  position: relative;
}

.video-player {
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
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
}

/* 播放图标叠加层：覆盖在封面图上，鼠标悬停时高亮 */
.play-overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  background: rgba(0, 0, 0, 0.3);
  transition: background 0.3s ease;
  pointer-events: none; /* 不拦截点击事件，让父级 card 的 @click 生效 */
}

.video-card:hover .play-overlay {
  background: rgba(0, 0, 0, 0.15);
}

.video-card:hover .play-overlay .el-icon {
  transform: scale(1.15);
  transition: transform 0.3s ease;
}

.video-info {
  padding: 16px;
}

.video-info h3 {
  color: #fff;
  font-size: 16px;
  margin-bottom: 4px;
}

.video-info p {
  color: rgba(255, 255, 255, 0.5);
  font-size: 13px;
  margin-bottom: 8px;
}

.author {
  display: flex;
  align-items: center;
  gap: 16px;
}

.publish-date {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: rgba(255, 255, 255, 0.4);
  font-size: 12px;
}

.actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.actions :deep(.el-button) {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.2);
  color: #fff;
}

.actions :deep(.el-button--danger) {
  background: #e94560;
  border-color: #e94560;
}

.actions :deep(.el-button--success) {
  background: linear-gradient(135deg, #e94560, #ff6b8a);
  border-color: transparent;
  color: #fff;
}

.actions :deep(.el-button--success:hover) {
  background: linear-gradient(135deg, #d63851, #e94560);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(233, 69, 96, 0.3);
}

.follow-btn {
  min-width: 80px;
}

.play-count-btn {
  cursor: default;
}

.play-count-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.2);
}

.loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 32px 0;
  color: rgba(255, 255, 255, 0.6);
}

.comment-drawer-content {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: #1a1a2e;
  width: 100%;
}

.comment-drawer-content h3 {
  color: #fff;
  margin: 0;
  padding: 20px 24px 16px 24px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  font-size: 20px;
}

.comment-list {
  flex: 1;
  overflow-y: auto;
  padding: 0 24px;
}

.comment-item {
  display: flex;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.comment-avatar {
  flex-shrink: 0;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  overflow: hidden;
}

.comment-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-fallback {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  background: linear-gradient(135deg, #e94560, #ff6b8a);
  display: flex;
  justify-content: center;
  align-items: center;
  font-size: 14px;
  font-weight: bold;
  color: #fff;
}

.comment-body {
  flex: 1;
  min-width: 0;
}

.comment-item strong {
  color: #e94560;
  font-size: 13px;
}

.comment-item p {
  color: #fff;
  font-size: 14px;
  margin: 4px 0;
}

.comment-item .time {
  color: rgba(255, 255, 255, 0.3);
  font-size: 12px;
}

.comment-input {
  padding: 16px 24px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  display: flex;
  gap: 12px;
}

.comment-input :deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.2);
}

.comment-input :deep(.el-input__inner) {
  color: #fff;
}

.comment-input :deep(.el-button--primary) {
  background: linear-gradient(135deg, #e94560, #ff6b8a) !important;
  border: none !important;
  color: #fff !important;
  font-weight: 500;
  transition: all 0.3s ease;
}

.comment-input :deep(.el-button--primary:hover) {
  background: linear-gradient(135deg, #d63851, #ff5a7a) !important;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(233, 69, 96, 0.4);
}

.comment-input :deep(.el-button--primary:active) {
  transform: translateY(0);
}
</style>
