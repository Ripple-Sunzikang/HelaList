<template>
  <div class="register-page-container">
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
      <el-card class="register-card" shadow="always">
        <template #header>
          <div class="card-header">
            <span>创建账户</span>
          </div>
        </template>

        <el-form
          ref="registerFormRef"
          :model="registerForm"
          :rules="registerRules"
          label-position="top"
          size="large"
          @submit.prevent="handleRegister"
        >
          <el-form-item label="用户名" prop="username">
            <el-input v-model="registerForm.username" placeholder="请选择用户名" :prefix-icon="User" />
          </el-form-item>
          <el-form-item label="邮箱" prop="email">
            <el-input v-model="registerForm.email" placeholder="请输入邮箱地址" :prefix-icon="Message" />
          </el-form-item>
          <el-form-item label="密码" prop="password">
            <el-input
              v-model="registerForm.password"
              type="password"
              placeholder="请创建密码"
              show-password
              :prefix-icon="Lock"
            />
          </el-form-item>
          <el-form-item label="确认密码" prop="confirmPassword">
            <el-input
              v-model="registerForm.confirmPassword"
              type="password"
              placeholder="请确认密码"
              show-password
              :prefix-icon="Lock"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" native-type="submit" class="register-button">注册</el-button>
          </el-form-item>
        </el-form>

        <div class="footer-actions">
          <el-button type="info" link @click="handleBack">已有账号？立即登录</el-button>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { User, Lock, Message } from '@element-plus/icons-vue'
import { useBackgroundSlider } from '../composables/useBackgroundSlider'
import { api } from '@/api'

// Background slider
const { currentImageIndex, backgroundImages } = useBackgroundSlider(6000)

const router = useRouter()
const registerFormRef = ref<FormInstance>()

const registerForm = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: '',
})

// Custom validator for confirming password
const validatePass2 = (rule: any, value: any, callback: any) => {
  if (value === '') {
    callback(new Error('请确认密码'))
  } else if (value !== registerForm.password) {
    callback(new Error('两次输入的密码不一致！'))
  } else {
    callback()
  }
}

const registerRules = reactive<FormRules>({
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  email: [
    { required: true, message: '请输入邮箱地址', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: ['blur', 'change'] },
  ],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  confirmPassword: [{ required: true, validator: validatePass2, trigger: 'blur' }],
})

const handleRegister = async () => {
  if (!registerFormRef.value) return
  await registerFormRef.value.validate((valid) => {
    if (valid) {
      ;(async () => {
        try {
          await api.post('/api/user/create', {
            username: registerForm.username,
            password: registerForm.password,
            email: registerForm.email,
          })
          ElMessage.success('注册成功！正在跳转到登录页...')
          router.push('/login')
        } catch (err: any) {
          ElMessage.error(err.message || '注册失败')
        }
      })()
    } else {
      ElMessage.error('请检查表单中的错误。')
    }
  })
}

const handleBack = () => {
  router.push('/login')
}
</script>

<style scoped>
/* Styles are shared with LoginPage, can be extracted to a common file if needed */
.register-page-container {
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

.register-card {
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

/* Override Element Plus styles */
.register-card :deep(.el-card__header) {
  border-bottom: 1px solid rgba(255, 255, 255, 0.2);
}

.register-card :deep(.el-form-item__label) {
  color: #eee;
  font-weight: 500;
}

.register-card :deep(.el-input__wrapper) {
  background-color: rgba(0, 0, 0, 0.2);
  box-shadow: none;
}

.register-card :deep(.el-input__inner) {
  color: #fff;
}

.register-button {
  width: 100%;
  font-weight: 600;
}

.footer-actions {
  margin-top: 10px;
  text-align: center;
}
</style>