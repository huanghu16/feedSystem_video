<template>
  <div class="home">
    <!-- 顶部导航 -->
    <div class="header">
      <h2>推荐</h2>
      <el-button v-if="auth.isLoggedIn" type="primary" size="small" @click="showUpload = true">
        + 发布
      </el-button>
      <el-button v-else size="small" @click="router.push('/account')">登录</el-button>
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
              评论
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- 评论抽屉 -->
    <el-drawer v-model="commentDrawer" title="评论" size="400px" :with-header="false">
      <div class="comment-drawer">
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
              <el-button @click="submitComment">发送</el-button>
            </template>
          </el-input>
        </div>
      </div>
    </el-drawer>

    <!-- 上传对话框 -->
    <el-dialog v-model="showUpload" title="发布视频" width="500px">
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
          <el-input v-model="uploadTitle" placeholder="给你的视频起个标题" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showUpload = false">取消</el-button>
        <el-button type="primary" :loading="uploading" @click="handleUpload">发布</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { VideoPlay, Star, ChatDotRound } from '@element-plus/icons-vue'
import { useAuthStore } from '../stores/auth'
import { listLatest } from '../api/feed'
import { uploadVideo, publishVideo } from '../api/video'
import { like, unlike, isLiked } from '../api/like'
import { listComments, publishComment } from '../api/comment'
import type { FeedVideoItem, CommentItem } from '../api/types'

const router = useRouter()   // 用于页面跳转
const auth = useAuthStore()  // 用于获取当前登录用户信息


const videos = ref<(FeedVideoItem & { isLiked?: boolean })[]>([])  // 视频列表
const playingId = ref<number | null>(null) // 当前正在播放的视频id，null表示没有视频在播放
const commentDrawer = ref(false) // 评论抽屉
const comments = ref<CommentItem[]>([]) // 评论列表
const currentVideoId = ref(0) // 当前视频的id
const newComment = ref('') // 新增的评论内容

const showUpload = ref(false) // 上传对话框
const uploadFile = ref<File | null>(null) // 上传的视频文件
const uploadTitle = ref('') // 新增的视频标题
const uploading = ref(false) // 上传中状态

// 获取视频的完整url，如果已经是完整url则直接返回，否则拼接服务器地址
function getFullUrl(path: string) {
  if (path.startsWith('http')) return path
  return `http://localhost:8080${path}`
}

// 页面加载时获取视频列表
onMounted(async () => {
  await loadFeed()
})

// 加载视频列表并检查每个视频是否已点赞
async function loadFeed() {
  try {
    const data = await listLatest()
    videos.value = data
  } catch {
    ElMessage.error('加载失败')
  }
}

// 播放视频
function playVideo(video: FeedVideoItem) {
  if (playingId.value === video.id) {
    playingId.value = null
  } else {
    playingId.value = video.id
  }
}

// 点赞或取消点赞视频
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

// 显示评论抽屉并加载评论列表
async function showComments(video: FeedVideoItem) {
  currentVideoId.value = video.id
  commentDrawer.value = true
  try {
    comments.value = await listComments(video.id)
  } catch {
    ElMessage.error('加载评论失败')
  }
}

// 发表评论
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

// 处理文件选择
function handleFileChange(file: any) {
  uploadFile.value = file.raw
}

// 上传视频并发布
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
    await loadFeed()
  } catch {
    ElMessage.error('发布失败')
  } finally {
    uploading.value = false
  }
}

// 格式化时间为中文格式
function formatTime(time: string) {
  return new Date(time).toLocaleString('zh-CN')
}
</script>

<style scoped>
.home {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
  min-height: 100vh;
  background: #1a1a2e;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.header h2 {
  color: #fff;
  font-size: 20px;
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

/* 评论抽屉 */
.comment-drawer {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #1a1a2e;
}

.comment-drawer h3 {
  color: #fff;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.comment-list {
  flex: 1;
  overflow-y: auto;
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
  padding-top: 12px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.comment-input :deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.2);
}

.comment-input :deep(.el-input__inner) {
  color: #fff;
}

/* 上传对话框 */
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