<template>
  <div class="video-detail">
    <el-button class="back-btn" text @click="router.back()">
      <el-icon><ArrowLeft /></el-icon> 返回
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
        <div v-for="c in comments" :key="c.id" class="comment">
          <strong>@{{ c.username }}</strong>
          <p>{{ c.content }}</p>
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
import { postJson } from '../api/client'
import { like, unlike, isLiked } from '../api/like'
import { listComments } from '../api/comment'
import type { VideoItem, CommentItem } from '../api/types'

const route = useRoute()  // 获取路由对象
const router = useRouter() // 获取路由实例
const auth = useAuthStore() // 获取认证状态

const video = ref<VideoItem | null>(null) // 视频详情
const liked = ref(false) // 是否已点赞
const comments = ref<CommentItem[]>([]) // 评论列表

// 获取完整的图片路径
function getFullUrl(path: string) {
  if (path.startsWith('http')) return path
  return `http://localhost:8080${path}`
}

// 加载视频详情、点赞状态、评论
onMounted(async () => {
  const videoId = Number(route.params.id)

  // 并行(Promise)加载视频详情、点赞状态、评论
  const promises: Promise<void>[] = [
    postJson<VideoItem>('/video/getDetail', { id: videoId }).then(data => {
      video.value = data
    }).catch(() => {
      ElMessage.error('视频不存在')
    }),
  ]

  // 如果已登录，查询是否已赞
  if (auth.isLoggedIn) {
    promises.push(
      isLiked(videoId).then(data => {
        liked.value = data.is_liked
      }).catch(() => {})
    )
  }

  await Promise.all(promises)

  // 加载评论
  if (video.value) {
    try {
      comments.value = await listComments(video.value.id)
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
  color: #fff;
  margin-bottom: 16px;
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
}

.comments h3 {
  color: #fff;
  margin-bottom: 16px;
}

.comment {
  padding: 12px 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.comment strong {
  color: #e94560;
  font-size: 13px;
}

.comment p {
  color: #fff;
  font-size: 14px;
  margin-top: 4px;
}

.loading {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 50vh;
  color: rgba(255, 255, 255, 0.5);
}
</style>