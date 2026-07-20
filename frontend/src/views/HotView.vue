<template>
  <div class="hot">
    <div class="header-info">
      <h2>🔥 实时热榜</h2>
    </div>

    <div v-if="loading" class="loading">
      <el-icon class="is-loading"><Loading /></el-icon>
      <p>加载中...</p>
    </div>

    <div v-else-if="hotVideos.length === 0" class="empty">
      <el-icon><VideoCamera /></el-icon>
      <p>暂无热门视频</p>
    </div>

    <div v-else class="video-list">
      <div
        v-for="(video, index) in hotVideos"
        :key="video.id"
        class="video-card"
        @click="goToVideo(video.id)"
      >
        <div class="rank">{{ index + 1 }}</div>
        <div class="video-info">
          <h3 class="title">{{ video.title }}</h3>
          <p v-if="video.description" class="description">{{ video.description }}</p>
          <div class="meta">
            <span class="author">@{{ video.username }}
              <span class="publish-date">
                <el-icon><Calendar /></el-icon>
                {{ formatDate(video.created_at) }}
              </span>
            </span>
            <span class="stats">
              <el-icon><VideoPlay /></el-icon>
              {{ formatNumber(video.play_count) }}
              <el-icon style="margin-left: 12px;"><Star /></el-icon>
              {{ formatNumber(video.likes_count) }}
              <el-icon style="margin-left: 12px;"><ChatDotRound /></el-icon>
              {{ formatNumber(video.comments_count || 0) }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Loading, VideoCamera, VideoPlay, Star, Calendar, ChatDotRound } from '@element-plus/icons-vue'
import { listHotVideos } from '../api/video'
import type { VideoItem } from '../api/types'
import { formatDate, formatNumber } from '../composables/useFormat'

const router = useRouter()
const loading = ref(false)
const hotVideos = ref<VideoItem[]>([])

function formatPublishDate(dateStr: string): string {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN')
}

function goToVideo(videoId: number) {
  router.push(`/video/${videoId}`)
}

async function loadHotVideos() {
  loading.value = true
  try {
    const res = await listHotVideos(10)
    hotVideos.value = res
  } catch (error: any) {
    console.error('加载热榜失败:', error)
    const errorMsg = error?.message || error?.payload?.message || '加载热榜失败'
    ElMessage.error(errorMsg)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadHotVideos()
})
</script>

<style scoped>
.hot {
  color: #fff;
  padding: 24px;
  max-width: 900px;
  margin: 0 auto;
}

.header-info {
  margin-bottom: 24px;
  text-align: center;
}

.header-info h2 {
  font-size: 28px;
  margin: 0 0 8px 0;
  background: linear-gradient(135deg, #ff6b6b, #ffd93d);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.subtitle {
  color: rgba(255, 255, 255, 0.5);
  font-size: 14px;
  margin: 0;
}

.loading, .empty {
  text-align: center;
  padding: 80px 20px;
  color: rgba(255, 255, 255, 0.5);
}

.loading .el-icon, .empty .el-icon {
  font-size: 64px;
  margin-bottom: 16px;
}

.empty .el-icon {
  opacity: 0.3;
}

.video-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.video-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
}

.video-card:hover {
  background: rgba(255, 255, 255, 0.06);
  border-color: rgba(233, 69, 96, 0.3);
  transform: translateX(4px);
}

.rank {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  font-weight: bold;
  color: rgba(255, 255, 255, 0.3);
  flex-shrink: 0;
}

.video-card:nth-child(1) .rank {
  color: #ffd700;
  text-shadow: 0 0 10px rgba(255, 215, 0, 0.5);
}

.video-card:nth-child(2) .rank {
  color: #c0c0c0;
  text-shadow: 0 0 10px rgba(192, 192, 192, 0.5);
}

.video-card:nth-child(3) .rank {
  color: #cd7f32;
  text-shadow: 0 0 10px rgba(205, 127, 50, 0.5);
}

.video-info {
  flex: 1;
  min-width: 0;
}

.title {
  font-size: 16px;
  font-weight: 600;
  margin: 0 0 8px 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: #fff;
}

.meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.6);
  margin-bottom: 4px;
}

.author {
  color: rgba(255, 255, 255, 0.7);
  display: flex;
  align-items: center;
  gap: 8px;
}

.stats {
  display: flex;
  align-items: center;
  gap: 4px;
}

.stats .el-icon {
  font-size: 14px;
}

.description {
  color: rgba(255, 255, 255, 0.6);
  font-size: 13px;
  margin: 8px 0;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.publish-date {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.4);
}

.publish-date .el-icon {
  font-size: 13px;
}
</style>