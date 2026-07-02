<template>
  <div class="video-detail">
    <el-button class="back-btn" @click="router.back()">
      <el-icon><ArrowLeft /></el-icon>
      返回
    </el-button>

    <div v-if="video" class="content">
      <video :src="getFullUrl(video.play_url)" controls class="main-video" />

      <div class="info">
        <h1>{{ video.title }}</h1>
        <p class="author">@{{ video.username }}</p>
        <div class="stats">
          <el-button
            :type="liked ? 'danger' : 'default'"
            @click="toggleLike"
          >
            <el-icon><Star /></el-icon>
            {{ video.likes_count }}
          </el-button>
        </div>
      </div>

      <div class="comments">
        <h3>评论</h3>

        <!-- 评论发表 -->
        <div v-if="auth.isLoggedIn" class="comment-input">
          <el-input
            v-model="newComment"
            placeholder="发表评论..."
            @keyup.enter="submitComment"
          />
          <el-button type="primary" @click="submitComment" :loading="commenting">
            发表
          </el-button>
        </div>

        <div v-for="c in comments" :key="c.id" class="comment">
          <div class="comment-avatar">
            <img v-if="c.avatar_url" :src="getFullAvatarUrl(c.avatar_url)" alt="头像" />
            <div v-else class="avatar-fallback">{{ c.username?.[0]?.toUpperCase() || 'U' }}</div>
          </div>
          <div class="comment-body">
            <strong>@{{ c.username }}</strong>
            <span class="time">{{ formatTime(c.created_at) }}</span>
            <p>{{ c.content }}</p>
          </div>
        </div>
        <el-empty v-if="comments.length === 0" description="暂无评论" />
      </div>
    </div>

    <div v-else class="loading">
      <p>加载中...</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Star } from '@element-plus/icons-vue'
import { useAuthStore } from '../stores/auth'
import { like, unlike, isLiked as checkIsLiked } from '../api/like'
import { getDetail } from '../api/video'
import { listComments, publishComment } from '../api/comment'
import type { VideoItem, CommentItem } from '../api/types'
import { getFullUrl, getFullAvatarUrl } from '../composables/useImageUrl'
import { formatTime } from '../composables/useFormat'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const video = ref<VideoItem | null>(null)
const liked = ref(false)
const comments = ref<CommentItem[]>([])
const newComment = ref('')
const commenting = ref(false)

// 加载视频详情、点赞状态、评论
onMounted(async () => {
  const videoId = Number(route.params.id)

  const promises: Promise<void>[] = [
    getDetail(videoId).then(data => {
      video.value = data
    }).catch(() => {
      ElMessage.error('视频不存在')
    }),
  ]

  if (auth.isLoggedIn) {
    promises.push(
      checkIsLiked(videoId).then(data => {
        liked.value = data.is_liked
      }).catch(() => {})
    )
  }

  await Promise.all(promises)

  // 加载评论
  if (video.value) {
    try {
      const resp = await listComments(video.value.id)
      comments.value = resp.list
    } catch {
      // 评论加载失败不影响页面
    }
  }
})

// 点赞、取消点赞
async function toggleLike() {
  if (!video.value) return
  if (!auth.isLoggedIn) {
    ElMessage.warning('请先登录')
    return
  }
  try {
    if (liked.value) {
      await unlike(video.value.id)
      liked.value = false
      video.value.likes_count--
    } else {
      await like(video.value.id)
      liked.value = true
      video.value.likes_count++
    }
  } catch {
    ElMessage.error('操作失败')
  }
}

// 发表评论
async function submitComment() {
  if (!video.value) return
  if (!newComment.value.trim()) return

  commenting.value = true
  try {
    const comment = await publishComment(video.value.id, newComment.value)
    comments.value.push(comment)
    newComment.value = ''
    ElMessage.success('评论成功')
  } catch {
    ElMessage.error('评论失败')
  } finally {
    commenting.value = false
  }
}
</script>

<style scoped>
.video-detail {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
  min-height: 100vh;
  background: #1a1a2e;
}

.back-btn {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.2);
  color: #fff;
  backdrop-filter: blur(10px);
  transition: all 0.3s ease;
  margin-bottom: 16px;
}

.back-btn:hover {
  background: rgba(233, 69, 96, 0.2);
  border-color: #e94560;
  color: #e94560;
  backdrop-filter: blur(15px);
}

.main-video {
  width: 100%;
  border-radius: 12px;
  background: #000;
}

.info {
  padding: 16px 0;
}

.info h1 {
  color: #fff;
  font-size: 18px;
  margin-bottom: 8px;
}

.author {
  color: rgba(255, 255, 255, 0.5);
  font-size: 14px;
  margin-bottom: 12px;
}

.stats :deep(.el-button) {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.2);
  color: #fff;
}

.stats :deep(.el-button--danger) {
  background: #e94560;
  border-color: #e94560;
}

.comments {
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  max-width: 100%;
  box-sizing: border-box;
}

.comments h3 {
  color: #fff;
  margin-bottom: 16px;
}

.comment-input {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
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

.comment {
  display: flex;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  max-width: 100%;
  box-sizing: border-box;
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

.comment strong {
  color: #e94560;
  font-size: 13px;
}

.comment .time {
  color: rgba(255, 255, 255, 0.3);
  font-size: 12px;
  margin-left: 8px;
}

.comment p {
  color: #fff;
  font-size: 14px;
  margin-top: 4px;
  word-wrap: break-word;
  overflow-wrap: break-word;
}

.loading {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 50vh;
  color: rgba(255, 255, 255, 0.5);
}
</style>
