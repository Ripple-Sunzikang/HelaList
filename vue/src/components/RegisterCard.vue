<template>
  <div class="register-card">
    <form @submit.prevent="handleRegister" class="register-form">
      <div class="form-group">
        <label for="username">用户名:</label>
        <input v-model="username" type="text" id="username" placeholder="请输入用户名" required />
      </div>
      <div class="form-group">
        <label for="email">邮箱:</label>
        <input v-model="email" type="email" id="email" placeholder="请输入邮箱" required />
      </div>
      <div class="form-group">
        <label for="password">密码:</label>
        <input v-model="password" type="password" id="password" placeholder="请输入密码" required />
      </div>
      <div class="form-group">
        <label for="confirmPassword">确认密码:</label>
        <input v-model="confirmPassword" type="password" id="confirmPassword" placeholder="请再次输入密码" required />
      </div>
      <div class="button-group">
        <button type="submit" class="confirm-button">确认</button>
        <button type="button" @click="handleBack" class="back-button">返回</button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'

const router = useRouter()
const username = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')

const handleRegister = () => {
  if (password.value !== confirmPassword.value) {
    ElMessage.error('两次输入的密码不一致')
    return
  }
  // 这里可以添加注册逻辑，比如调用接口提交注册信息
  ElMessage.success('注册成功，可前往登录页登录')
  router.push('/login')
}

const handleBack = () => {
  router.push('/login')
}
</script>

<style scoped>
.register-card {
  background: rgba(255, 255, 255, 0.2);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  border-radius: 20px;
  padding: 40px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  max-width: 400px;
  width: 100%;
  margin: 20px;
  text-align: left;
}

.register-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group label {
  font-weight: 500;
  color: #fff;
  font-size: 14px;
}

.form-group input {
  padding: 12px 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-radius: 10px;
  font-size: 16px;
  transition: all 0.3s ease;
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}

.form-group input:focus {
  outline: none;
  border-color: #4f46e5;
  box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.1);
}

.button-group {
  display: flex;
  gap: 15px;
  margin-top: 10px;
}

.confirm-button,
.back-button {
  flex: 1;
  color: white;
  border: none;
  padding: 14px 20px;
  border-radius: 10px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  text-align: center;
}

.confirm-button {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.confirm-button:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 10px 20px rgba(102, 126, 234, 0.3);
}

.confirm-button:active:not(:disabled) {
  transform: translateY(0);
}

.back-button {
  background: linear-gradient(135deg, #1f4ce1 0%, #7942d1 100%);
}

.back-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 20px rgba(108, 117, 125, 0.3);
}

.back-button:active {
  transform: translateY(0);
}

@media (max-width: 480px) {
  .register-card {
    padding: 30px 20px;
    margin: 10px;
  }
  .button-group {
    flex-direction: column;
    gap: 10px;
  }
  .confirm-button,
  .back-button {
    width: 100%;
  }
}
</style>
