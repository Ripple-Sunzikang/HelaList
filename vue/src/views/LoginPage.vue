<template>
  <div class="login-page-container">
    <!-- Background layers -->
    <div
      v-for="(image, index) in backgroundImages"
      :key="index"
      class="background-layer"
      :class="{ active: index === currentImageIndex }"
      :style="{ backgroundImage: `url(${image})` }"
    ></div>

    <!-- Glassmorphism overlay -->
    <div class="background-overlay"></div>

    <!-- Content wrapper -->
    <div class="content-wrapper">
      <el-card class="login-card" shadow="always">
        <template #header>
          <div class="card-header">
            <span>Welcome to HelaList</span>
          </div>
        </template>

        <el-form
          ref="loginFormRef"
          :model="loginForm"
          :rules="loginRules"
          label-position="top"
          size="large"
          @submit.prevent="handleLogin"
        >
          <el-form-item label="Username" prop="username">
            <el-input v-model="loginForm.username" placeholder="Please enter username" :prefix-icon="User" />
          </el-form-item>
          <el-form-item label="Password" prop="password">
            <el-input
              v-model="loginForm.password"
              type="password"
              placeholder="Please enter password"
              show-password
              :prefix-icon="Lock"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" native-type="submit" class="login-button">Login</el-button>
          </el-form-item>
        </el-form>

        <div class="footer-actions">
          <el-button type="info" link @click="handleRegister">No account yet? Register</el-button>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { useBackgroundSlider } from '../composables/useBackgroundSlider'
import { api } from '@/api'

// Background slider
const { currentImageIndex, backgroundImages } = useBackgroundSlider(6000)

const router = useRouter()
const loginFormRef = ref<FormInstance>()

const loginForm = reactive({
  username: '',
  password: '',
})

const loginRules = reactive<FormRules>({
  username: [{ required: true, message: 'Username is required', trigger: 'blur' }],
  password: [{ required: true, message: 'Password is required', trigger: 'blur' }],
})

const handleLogin = async () => {
  if (!loginFormRef.value) return
  await loginFormRef.value.validate((valid) => {
    if (valid) {
      ;(async () => {
        try {
          const data = await api.post('/api/user/login', {
            username: loginForm.username,
            password: loginForm.password,
          })
          // 后端返回 { token, user }
          if (data && data.token) {
            localStorage.setItem('token', data.token)
            ElMessage.success('Login successful! Redirecting...')
            router.push('/home')
          } else {
            ElMessage.error('登录失败：未返回 token')
          }
        } catch (err: any) {
          ElMessage.error(err.message || '登录失败')
        }
      })()
    } else {
      ElMessage.error('Please check the form for errors.')
    }
  })
}

const handleRegister = () => {
  router.push('/register')
}
</script>

<style scoped>
.login-page-container {
  height: 100vh;
  width: 100%;
  position: relative;
  display: flex;
  justify-content: center;
  align-items: center;
  overflow: hidden;
}

.background-layer {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
  opacity: 0;
  transition: opacity 1.5s ease-in-out, transform 8s ease-in-out;
  will-change: opacity, transform;
}

.background-layer.active {
  opacity: 1;
  transform: scale(1.1);
}

.background-overlay {
  position: absolute;
  inset: 0;
  background-color: rgba(0, 0, 0, 0.3);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  z-index: 1;
}

.content-wrapper {
  position: relative;
  z-index: 2;
}

.login-card {
  width: 400px;
  max-width: 90vw;
  background-color: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: #fff;
  --el-card-border-color: transparent;
}

.card-header {
  text-align: center;
  font-size: 24px;
  font-weight: 600;
  color: #fff;
}

/* Override Element Plus styles for a more integrated look */
.login-card :deep(.el-card__header) {
  border-bottom: 1px solid rgba(255, 255, 255, 0.2);
}

.login-card :deep(.el-form-item__label) {
  color: #eee;
  font-weight: 500;
}

.login-card :deep(.el-input__wrapper) {
  background-color: rgba(0, 0, 0, 0.2);
  box-shadow: none;
}

.login-card :deep(.el-input__inner) {
  color: #fff;
}

.login-button {
  width: 100%;
  font-weight: 600;
}

.footer-actions {
  margin-top: 10px;
  text-align: center;
}
</style>