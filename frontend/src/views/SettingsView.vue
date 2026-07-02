<template>
  <div class="settings">
    <h1>账号设置</h1>

    <div class="settings-content">
      <!-- 左侧：头像 + 简介 -->
      <div class="left-panel">
        <div class="setting-section">
          <h2>头像设置</h2>
          <div class="avatar-preview">
            <div v-if="currentAvatar" class="avatar-image">
              <img :src="getFullAvatarUrl(currentAvatar)" alt="头像" />
            </div>
            <div v-else class="avatar-placeholder">
              {{ username?.[0]?.toUpperCase() || 'U' }}
            </div>
          </div>
          <div class="avatar-upload">
            <input
              ref="avatarInput"
              type="file"
              accept="image/*"
              @change="handleAvatarChange"
              style="display: none"
            />
            <el-button class="glass-btn" @click="triggerAvatarUpload">
              <el-icon><Upload /></el-icon>
              更换头像
            </el-button>
            <p class="hint">支持 JPG、PNG、GIF、WebP 格式，最大10MB</p>
          </div>
        </div>

        <div class="setting-section">
          <h2>个人简介</h2>
          <el-input
            v-model="bio"
            type="textarea"
            :rows="4"
            placeholder="介绍一下你自己吧~"
            maxlength="256"
            show-word-limit
          />
          <el-button class="glass-btn" @click="handleUpdateBio" :loading="updatingBio">
            保存简介
          </el-button>
        </div>
      </div>

      <!-- 右侧：密码修改 + 退出登录 -->
      <div class="right-panel">
        <div class="setting-section">
          <h2>修改密码</h2>
          <el-form :model="passwordForm" label-width="100px">
            <el-form-item label="原密码">
              <el-input
                v-model="passwordForm.oldPassword"
                type="password"
                placeholder="请输入原密码"
                show-password
              />
            </el-form-item>
            <el-form-item label="新密码">
              <el-input
                v-model="passwordForm.newPassword"
                type="password"
                placeholder="请输入新密码（6-64位）"
                show-password
              />
            </el-form-item>
            <el-form-item label="确认密码">
              <el-input
                v-model="passwordForm.confirmPassword"
                type="password"
                placeholder="请再次输入新密码"
                show-password
              />
            </el-form-item>
          </el-form>
          <el-button class="glass-btn" @click="handleChangePassword" :loading="changingPassword">
            修改密码
          </el-button>
        </div>

        <div class="setting-section danger-zone">
          <h2>退出操作</h2>
          <el-button type="danger" @click="handleLogout">
            <el-icon><SwitchButton /></el-icon>
            退出登录
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Upload, SwitchButton } from '@element-plus/icons-vue'
import { useAuthStore } from '../stores/auth'
import { getProfile, updateAvatar, changePassword, updateBio } from '../api/account'
import { getFullAvatarUrl } from '../composables/useImageUrl'

const router = useRouter()
const auth = useAuthStore()

const currentAvatar = ref<string>('')
const username = ref<string>('')
const bio = ref<string>('')
const updatingBio = ref(false)
const changingPassword = ref(false)

const passwordForm = ref({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

onMounted(async () => {
  if (!auth.isLoggedIn) {
    ElMessage.warning('请先登录')
    router.push('/account')
    return
  }

  await loadProfile()
})

async function loadProfile() {
  try {
    const userId = auth.claims?.account_id
    if (!userId) {
      ElMessage.error('无法获取用户信息')
      return
    }
    const data = await getProfile(userId)
    currentAvatar.value = data.avatar_url
    username.value = data.username
    bio.value = data.bio || ''
    if (data.avatar_url) {
      auth.setAvatar(data.avatar_url)
    }
  } catch (error: any) {
    console.error('加载用户资料失败:', error)
    ElMessage.error(error?.payload?.message || '加载用户资料失败')
  }
}

const avatarInput = ref<HTMLInputElement | null>(null)

function triggerAvatarUpload() {
  avatarInput.value?.click()
}

async function handleAvatarChange(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  if (file.size > 10 * 1024 * 1024) {
    ElMessage.error('图片大小不能超过10MB')
    return
  }

  const allowedTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp']
  if (!allowedTypes.includes(file.type)) {
    ElMessage.error('只支持 JPG、PNG、GIF、WebP 格式的图片')
    return
  }

  try {
    ElMessage.info('上传中...')
    const result = await updateAvatar(file)
    currentAvatar.value = result.avatar_url
    auth.setAvatar(result.avatar_url)
    ElMessage.success('头像更新成功')
  } catch (error: any) {
    console.error('上传头像失败:', error)
    ElMessage.error(error?.payload?.message || '上传头像失败')
  }
}

async function handleUpdateBio() {
  if (bio.value.length > 256) {
    ElMessage.error('简介不能超过256个字符')
    return
  }

  updatingBio.value = true
  try {
    await updateBio(bio.value)
    ElMessage.success('简介更新成功')
  } catch (error: any) {
    console.error('更新简介失败:', error)
    ElMessage.error(error?.payload?.message || '更新简介失败')
  } finally {
    updatingBio.value = false
  }
}

async function handleChangePassword() {
  if (!passwordForm.value.oldPassword) {
    ElMessage.error('请输入原密码')
    return
  }
  if (!passwordForm.value.newPassword || passwordForm.value.newPassword.length < 6) {
    ElMessage.error('新密码长度不能少于6位')
    return
  }
  if (passwordForm.value.newPassword !== passwordForm.value.confirmPassword) {
    ElMessage.error('两次输入的密码不一致')
    return
  }

  changingPassword.value = true
  try {
    await changePassword(passwordForm.value.oldPassword, passwordForm.value.newPassword)
    ElMessage.success('密码修改成功，请重新登录')
    passwordForm.value = {
      oldPassword: '',
      newPassword: '',
      confirmPassword: ''
    }
    setTimeout(() => {
      auth.clearTokens()
      router.push('/account')
    }, 1500)
  } catch (error: any) {
    console.error('修改密码失败:', error)
    ElMessage.error(error?.payload?.message || '修改密码失败')
  } finally {
    changingPassword.value = false
  }
}

function handleLogout() {
  auth.clearTokens()
  ElMessage.success('已退出登录')
  router.push('/account')
}
</script>

<style scoped>
.settings {
  padding: 40px;
  color: #fff;
  background: #1a1a2e;
  min-height: 100vh;
  max-width: 900px;
  margin: 0 auto;
}

.settings h1 {
  font-size: 28px;
  margin-bottom: 32px;
  color: #fff;
  border-bottom: 2px solid #e94560;
  padding-bottom: 16px;
}

/* 左右两栏布局 */
.settings-content {
  display: flex;
  gap: 24px;
  align-items: flex-start;
}

.left-panel,
.right-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

/* 响应式：小屏幕切换为上下布局 */
@media (max-width: 768px) {
  .settings-content {
    flex-direction: column;
  }
}

.setting-section {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 24px;
}

.setting-section h2 {
  font-size: 20px;
  margin-bottom: 20px;
  color: #fff;
}

.avatar-preview {
  display: flex;
  justify-content: center;
  margin-bottom: 20px;
}

.avatar-image {
  width: 120px;
  height: 120px;
  border-radius: 50%;
  overflow: hidden;
  border: 3px solid #e94560;
}

.avatar-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-placeholder {
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

.avatar-upload {
  text-align: center;
}

.avatar-upload .hint {
  margin-top: 12px;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.5);
}

.setting-section :deep(.el-textarea__inner) {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: #fff;
}

.setting-section :deep(.el-input__count) {
  background: transparent;
  color: rgba(255, 255, 255, 0.5);
}

.setting-section :deep(.el-form-item__label) {
  color: rgba(255, 255, 255, 0.8);
}

.setting-section :deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.05);
  box-shadow: 0 0 0 1px rgba(255, 255, 255, 0.2);
}

.setting-section :deep(.el-input__inner) {
  color: #fff;
}

.setting-section .el-button {
  margin-top: 16px;
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

.danger-zone {
  border-color: rgba(233, 69, 96, 0.3);
}

.danger-zone h2 {
  color: #e94560;
}

.danger-zone .el-button {
  width: 100%;
}

/* ========== 新增：让左侧“头像设置”与右侧“修改密码”高度一致 ========== */
.left-panel .setting-section:first-child,
.right-panel .setting-section:first-child {
  min-height: 360px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}
</style>