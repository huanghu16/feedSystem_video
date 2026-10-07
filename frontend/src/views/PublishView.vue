<template>
  <div class="publish-page">
    <div class="header-info">
      <h2>📤 发布视频</h2>
      <p class="subtitle">分享你的精彩瞬间</p>
    </div>

    <div class="publish-form">
      <!-- 文件上传区域 -->
      <div class="upload-section">
        <div v-if="!previewUrl" class="upload-area" @click="triggerFileInput">
          <el-icon :size="64" color="rgba(255,255,255,0.3)"><Upload /></el-icon>
          <p class="upload-text">点击选择视频文件</p>
          <p class="upload-hint">支持 MP4、AVI、MOV 等格式</p>
        </div>
        <div v-else class="preview-area">
          <video :src="previewUrl" controls class="video-preview" />
          <el-button type="danger" size="small" @click="removeVideo">
            <el-icon><Delete /></el-icon>
            移除视频
          </el-button>
        </div>
        <input
          ref="fileInput"
          type="file"
          accept="video/*"
          style="display: none"
          @change="handleFileChange"
        />

        <!-- 分片上传进度 -->
        <div v-if="submitting && uploadPercent > 0" class="upload-progress">
          <el-progress
            :percentage="uploadPercent"
            :stroke-width="12"
            :text-inside="true"
            color="#e94560"
          />
          <p class="progress-hint">
            {{ uploadPercent < 100 ? '正在上传分片...' : '正在合并文件...' }}
          </p>
        </div>
      </div>

      <!-- 封面上传区域 -->
      <div class="upload-section">
        <div class="form-label">
          <el-icon><Picture /></el-icon>
          视频封面（可选）
        </div>
        <div v-if="!coverPreviewUrl" class="upload-area cover-upload-area" @click="triggerCoverInput">
          <el-icon :size="48" color="rgba(255,255,255,0.3)"><Upload /></el-icon>
          <p class="upload-text">点击上传封面图片</p>
          <p class="upload-hint">支持 JPG、PNG 格式，不上传则自动生成</p>
        </div>
        <div v-else class="preview-area cover-preview-area">
          <img :src="coverPreviewUrl" alt="封面预览" class="cover-preview" />
          <el-button type="danger" size="small" @click="removeCover">
            <el-icon><Delete /></el-icon>
            移除封面
          </el-button>
        </div>
        <input
          ref="coverFileInput"
          type="file"
          accept="image/*"
          style="display: none"
          @change="handleCoverChange"
        />
      </div>

      <!-- 表单字段 -->
      <div class="form-fields">
        <div class="form-item">
          <label class="form-label">
            <el-icon><Document /></el-icon>
            标题
          </label>
          <el-input
            v-model="form.title"
            placeholder="请输入视频标题"
            show-word-limit
            size="large"
          />
        </div>

        <div class="form-item">
          <label class="form-label">
            <el-icon><EditPen /></el-icon>
            内容描述
          </label>
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="4"
            placeholder="介绍一下你的视频内容吧..."
            show-word-limit
          />
        </div>

        <div class="form-item">
          <label class="form-label">
            <el-icon><Calendar /></el-icon>
            发布日期
          </label>
          <el-date-picker
            v-model="form.publishDate"
            type="date"
            placeholder="选择发布日期"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            size="large"
            style="width: 100%"
          />
        </div>

        <div class="form-actions">
          <el-button
            type="primary"
            size="large"
            :loading="submitting"
            :disabled="!canSubmit"
            @click="handleSubmit"
          >
            <el-icon><Check /></el-icon>
            {{ submitting ? '发布中...' : '发布视频' }}
          </el-button>
          <el-button size="large" @click="handleReset">
            <el-icon><Refresh /></el-icon>
            重置
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Upload, Delete, Document, EditPen, Calendar, Check, Refresh, Picture } from '@element-plus/icons-vue'
import { publishVideo, uploadCover } from '../api/video'
import { uploadVideoChunked } from '../api/upload'

const router = useRouter()

// 表单数据
const form = ref({
  title: '',
  description: '',
  publishDate: new Date().toISOString().split('T')[0], // 默认今天
})

// 文件相关
const fileInput = ref<HTMLInputElement | null>(null)
const selectedFile = ref<File | null>(null)
const previewUrl = ref<string>('')
const submitting = ref(false)
const uploadPercent = ref(0)

// 封面相关
const coverFileInput = ref<HTMLInputElement | null>(null)
const selectedCoverFile = ref<File | null>(null)
const coverPreviewUrl = ref<string>('')

// 是否可以提交
const canSubmit = computed(() => {
  return selectedFile.value && form.value.title.trim()
})

// 触发文件选择
function triggerFileInput() {
  fileInput.value?.click()
}

// 处理文件选择
function handleFileChange(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  
  if (!file) return

  // 验证文件类型
  if (!file.type.startsWith('video/')) {
    ElMessage.error('请选择视频文件')
    return
  }

  // 验证文件大小（限制100MB）
  const maxSize = 100 * 1024 * 1024
  if (file.size > maxSize) {
    ElMessage.error('视频文件大小不能超过100MB')
    return
  }

  selectedFile.value = file
  previewUrl.value = URL.createObjectURL(file)
}

// 移除视频
function removeVideo() {
  selectedFile.value = null
  if (previewUrl.value) {
    URL.revokeObjectURL(previewUrl.value)
    previewUrl.value = ''
  }
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

// 触发封面选择
function triggerCoverInput() {
  coverFileInput.value?.click()
}

// 处理封面选择
function handleCoverChange(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]

  if (!file) return

  // 验证文件类型
  if (!file.type.startsWith('image/')) {
    ElMessage.error('请选择图片文件')
    return
  }

  // 验证文件大小（限制10MB）
  const maxSize = 10 * 1024 * 1024
  if (file.size > maxSize) {
    ElMessage.error('封面图片大小不能超过10MB')
    return
  }

  selectedCoverFile.value = file
  coverPreviewUrl.value = URL.createObjectURL(file)
}

// 移除封面
function removeCover() {
  selectedCoverFile.value = null
  if (coverPreviewUrl.value) {
    URL.revokeObjectURL(coverPreviewUrl.value)
    coverPreviewUrl.value = ''
  }
  if (coverFileInput.value) {
    coverFileInput.value.value = ''
  }
}

// 提交发布
async function handleSubmit() {
  if (!canSubmit.value) {
    ElMessage.warning('请填写完整信息')
    return
  }

  if (!selectedFile.value) {
    ElMessage.warning('请选择视频文件')
    return
  }

  submitting.value = true
  uploadPercent.value = 0
  try {
    console.log('开始分片上传视频...')
    const uploadRes = await uploadVideoChunked(selectedFile.value, (p) => {
      uploadPercent.value = p.percent
    })
    console.log('视频上传响应:', uploadRes)

    // 使用后端自动生成的封面（如果有）
    let coverUrl = uploadRes.cover_url || ''
    console.log('封面URL:', coverUrl)

    // 如果用户上传了自定义封面，上传自定义封面
    if (selectedCoverFile.value) {
      console.log('上传自定义封面...')
      const coverRes = await uploadCover(selectedCoverFile.value)
      coverUrl = coverRes.url
      console.log('自定义封面URL:', coverUrl)
    }

    console.log('发布视频信息...')
    await publishVideo(form.value.title, uploadRes.play_url, coverUrl, form.value.description, form.value.publishDate)
    
    ElMessage.success('发布成功！')
    
    handleReset()
    
    setTimeout(() => {
      router.push('/')
    }, 1000)
  } catch (error: any) {
    console.error('发布失败:', error)
    ElMessage.error(error?.message || '发布失败，请重试')
  } finally {
    submitting.value = false
  }
}

// 重置表单
function handleReset() {
  form.value = {
    title: '',
    description: '',
    publishDate: new Date().toISOString().split('T')[0],
  }
  removeVideo()
  removeCover()
  uploadPercent.value = 0
}
</script>

<style scoped>
.publish-page {
  color: #fff;
  padding: 24px;
  max-width: 900px;
  margin: 0 auto;
}

.header-info {
  margin-bottom: 32px;
  text-align: center;
}

.header-info h2 {
  font-size: 28px;
  margin: 0 0 8px 0;
  background: linear-gradient(135deg, #e94560, #ff6b8a);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.subtitle {
  color: rgba(255, 255, 255, 0.5);
  font-size: 14px;
  margin: 0;
}

.publish-form {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.upload-section {
  width: 100%;
}

.upload-area {
  border: 2px dashed rgba(255, 255, 255, 0.2);
  border-radius: 12px;
  padding: 60px 20px;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s ease;
  background: rgba(255, 255, 255, 0.03);
}

.upload-area:hover {
  border-color: #e94560;
  background: rgba(233, 69, 96, 0.05);
}

.upload-text {
  font-size: 16px;
  color: rgba(255, 255, 255, 0.7);
  margin: 16px 0 8px;
}

.upload-hint {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.4);
}

.upload-progress {
  margin-top: 12px;
}

.progress-hint {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.5);
  text-align: center;
  margin: 8px 0 0;
}

:deep(.el-progress-bar__outer) {
  background-color: rgba(255, 255, 255, 0.1) !important;
}

.preview-area {
  position: relative;
  border-radius: 12px;
  overflow: hidden;
  background: #000;
}

.video-preview {
  width: 100%;
  max-height: 400px;
  display: block;
}

.preview-area .el-button {
  position: absolute;
  top: 12px;
  right: 12px;
}

.cover-upload-area {
  padding: 40px 20px;
}

.cover-preview-area {
  max-height: 300px;
}

.cover-preview {
  width: 100%;
  max-height: 300px;
  object-fit: cover;
  display: block;
}

.form-fields {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.8);
  font-weight: 500;
}

.form-label .el-icon {
  color: #e94560;
}

.form-actions {
  display: flex;
  gap: 12px;
  margin-top: 8px;
}

.form-actions .el-button {
  flex: 1;
}

/* Element Plus 组件样式覆盖 */
:deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.06) !important;
  border: 1px solid rgba(255, 255, 255, 0.15) !important;
  box-shadow: none !important;
}

:deep(.el-input__inner) {
  color: #fff !important;
}

:deep(.el-input__inner::placeholder) {
  color: rgba(255, 255, 255, 0.3) !important;
}

:deep(.el-textarea__inner) {
  background: rgba(255, 255, 255, 0.06) !important;
  border: 1px solid rgba(255, 255, 255, 0.15) !important;
  color: #fff !important;
}

:deep(.el-textarea__inner::placeholder) {
  color: rgba(255, 255, 255, 0.3) !important;
}

:deep(.el-date-editor.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.06) !important;
  border: 1px solid rgba(255, 255, 255, 0.15) !important;
}

:deep(.el-input__suffix) {
  color: rgba(255, 255, 255, 0.5) !important;
}

:deep(.el-counters) {
  color: rgba(255, 255, 255, 0.4) !important;
}

:deep(.el-button--primary) {
  background: linear-gradient(135deg, #e94560, #ff6b8a) !important;
  border: none !important;
}

:deep(.el-button--primary:hover) {
  background: linear-gradient(135deg, #d63851, #ff5a7a) !important;
}

:deep(.el-button--primary:disabled) {
  opacity: 0.5 !important;
}
</style>