<template>
  <div class="home">
    <!-- 视频列表 -->
    <div class="video-list">
      <div
        v-for="video in videos"
        :key="video.id"
        class="video-card"
        @click="playVideo(video)"
      >
        <div class="video-cover">
          <video
            v-if="playingId === video.id"
            :src="getFullUrl(video.play_url)"
            controls
            autoplay
            class="video-player"
          />
          <div v-else class="cover-placeholder">
            <el-icon :size="48" color="rgba(255,255,255,0.3)"><VideoPlay /></el-icon>
          </div>
        </div>
        <div class="video-info">
          <h3>{{ video.title }}</h3>
          <p>@{{ video.username }}</p>
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
          </div>
        </div>
      </div>
    </div>

    <el-empty v-if="videos.length === 0" description="视频加载中..." />

    <!-- 评论抽屉 -->
    <el-drawer v-model="commentDrawer" title="评论" size="400px" :with-header="false" class="comment-drawer-wrapper">
      <div class="comment-drawer-content">
        <h3>评论</h3>
        <div class="comment-list">
          <div v-for="c in comments" :key="c.id" class="comment-item">
            <strong>@{{ c.username }}</strong>
            <p>{{ c.content }}</p>
            <span class="time">{{ formatTime(c.created_at) }}</span>
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
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { VideoPlay, Star, ChatDotRound } from '@element-plus/icons-vue'
import { useAuthStore } from '../stores/auth'
import { listLatest } from '../api/feed'
import { like, unlike, isLiked } from '../api/like'
import { listComments, publishComment } from '../api/comment'
import { recordPlay } from '../api/video'
import type { FeedVideoItem, CommentItem } from '../api/types'

const router = useRouter()
const auth = useAuthStore()

const videos = ref<(FeedVideoItem & { isLiked?: boolean })[]>([])
const playingId = ref<number | null>(null)
const commentDrawer = ref(false)
const comments = ref<CommentItem[]>([])
const currentVideoId = ref(0)
const newComment = ref('')

function getFullUrl(path: string) {
  if (path.startsWith('http')) return path
  return `http://localhost:8080${path}`
}

onMounted(async () => {
  await loadFeed()
})

async function loadFeed() {
  try {
    const data = await listLatest()
    videos.value = data
  } catch {
    ElMessage.error('加载失败')
  }
}

function playVideo(video: FeedVideoItem) {
  if (playingId.value === video.id) {
    playingId.value = null
  } else {
    playingId.value = video.id
    // 记录播放
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

async function showComments(video: FeedVideoItem) {
  currentVideoId.value = video.id
  commentDrawer.value = true
  try {
    comments.value = await listComments(video.id)
  } catch {
    ElMessage.error('加载评论失败')
  }
}

async function submitComment() {
  if (!newComment.value.trim()) return
  try {
    const comment = await publishComment(currentVideoId.value, newComment.value)
    comments.value.push(comment)
    newComment.value = ''
    ElMessage.success('评论成功')
  } catch {
    ElMessage.error('评论失败')
  }
}

function formatTime(time: string) {
  return new Date(time).toLocaleString('zh-CN')
}
</script>

<style>
/* 全局样式 - 覆盖 Element Plus Drawer 默认样式 */
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

.cover-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
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
  margin-bottom: 12px;
}

.actions {
  display: flex;
  gap: 8px;
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

.play-count-btn {
  cursor: default;
}

.play-count-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.2);
}

/* 评论抽屉内容 */
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
  padding: 12px 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
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